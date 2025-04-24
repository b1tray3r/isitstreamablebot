package internal

import (
	"context"
	"fmt"
	"log/slog"
	"sort"
	"strings"

	"github.com/b1tray3r/isitstreamablebot/internal/twitch"
	"github.com/bwmarrin/discordgo"
	"github.com/zmb3/spotify/v2"
)

type DiscordBot struct {
	Session       *discordgo.Session
	SpotifyClient *spotify.Client
	Version       string

	Commands []*discordgo.ApplicationCommand

	ChannelFilter []string
}

func NewDiscordBot(token, version string, channelIDs []string) (*DiscordBot, error) {
	dg, err := discordgo.New("Bot " + token)
	if err != nil {
		return nil, fmt.Errorf("error creating Discord session: %w", err)
	}

	commands := []*discordgo.ApplicationCommand{
		{
			Name:        "version",
			Description: "Replies with the version of the bot.",
		},
		{
			Name:        "check",
			Description: "Check if a track is in the Twitch DJ catalog.",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "title",
					Description: "The title of the track to check.",
					Required:    true,
				},
			},
		},
	}

	if err := dg.Open(); err != nil {
		return nil, fmt.Errorf("error opening Discord session: %w", err)
	}

	bot := &DiscordBot{
		Session:       dg,
		SpotifyClient: nil,
		Version:       version,
		ChannelFilter: channelIDs,
		Commands:      commands,
	}

	for _, command := range commands {
		slog.Info("Registering command", "name", command.Name)
		_, err := dg.ApplicationCommandCreate(dg.State.User.ID, "", command)
		if err != nil {
			return nil, fmt.Errorf("error creating command %s: %w", command.Name, err)
		}
	}

	dg.AddHandler(bot.SlashCommandHandler)
	dg.AddHandler(bot.TwitchCatalogCheckHandler)

	return bot, nil
}

func (b *DiscordBot) SlashCommandHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if i.Type != discordgo.InteractionApplicationCommand {
		return
	}

	switch i.ApplicationCommandData().Name {
	case "version":
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Version: " + b.Version,
			},
		})
	case "check":
		if len(i.ApplicationCommandData().Options) == 0 {
			slog.Error("No options provided")
			return
		}
		if i.ApplicationCommandData().Options[0].Type != discordgo.ApplicationCommandOptionString {
			slog.Error("Invalid option type")
			return
		}
		if i.ApplicationCommandData().Options[0].Name != "title" {
			slog.Error("Invalid option name")
			return
		}
		if i.ApplicationCommandData().Options[0].StringValue() == "" {
			slog.Error("Empty track name")
			return
		}
		trackName := i.ApplicationCommandData().Options[0].StringValue()

		response, err := twitch.SendRequest(trackName)
		if err != nil {
			slog.Error("TwitchDJCatalog error", "err", err)
			return
		}

		slog.Debug("TwitchDJCatalog response", "response", response)
		if len(response) == 0 {
			slog.Error("No response from Twitch DJ Catalog")
			return
		}

		messages := []string{"Twitch DJ Catalog Results for \"" + trackName + "\":"}
		for _, r := range response {
			if len(r.Data.SearchDJCatalog.Edges) == 0 {
				slog.Error("No results found.")
				return
			}

			for _, edge := range r.Data.SearchDJCatalog.Edges {
				state := "✅"
				if edge.Node.IsBlockedTrack {
					state = "❌"
				}

				messages = append(messages, renderDiscordMessage(edge, state, ""))
			}
		}

		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: strings.Join(messages, "\n"),
			},
		})
	default:
		slog.Error("Unknown command", "command", i.ApplicationCommandData().Name)
	}
}

func (b *DiscordBot) TwitchCatalogCheckHandler(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID {
		return
	}

	isListeningChannel := false
	for _, channelID := range b.ChannelFilter {
		if m.ChannelID == strings.TrimSpace(channelID) {
			isListeningChannel = true
			break
		}
	}
	if !isListeningChannel {
		return
	}

	if HasSpotifyLink(m.Content) {
		links := ExtractSpotifyLinksFromText(m.Content)
		for _, link := range links {
			spotifyID := ExtractSpotifyTrackIDFromLink(link)
			slog.Info("Extracted", "SpotifyID", spotifyID)
			s.MessageReactionAdd(m.ChannelID, m.ID, "⏳")

			trackID := spotify.ID(spotifyID)
			track, err := b.SpotifyClient.GetTrack(context.Background(), trackID)
			if err != nil {
				slog.Error("Spotify error", "err", err, "trackID", spotifyID)
				return
			}
			if track == nil {
				slog.Error("Track not found on Spotify")
				return
			}

			response, err := twitch.SendRequest(track.Name)
			if err != nil {
				slog.Error("TwitchDJCatalog error", "err", err)
				return
			}

			for _, r := range response {
				if len(r.Data.SearchDJCatalog.Edges) == 0 {
					slog.Error("No results found.")
					return
				}

				messages := []string{"Twitch DJ Catalog Results for \"" + track.Name + "\":"}
				for _, edge := range r.Data.SearchDJCatalog.Edges {
					state := "✅"
					if edge.Node.IsBlockedTrack {
						state = "❌"
					}

					indent := ""
					if track.Name == edge.Node.Title {
						spotifyDuration := int(track.Duration / 1000)
						nodeDuration := int(edge.Node.Duration)
						if spotifyDuration == nodeDuration {
							s.MessageReactionRemove(m.ChannelID, m.ID, "⏳", s.State.User.ID)
							s.MessageReactionAdd(m.ChannelID, m.ID, state)
							indent = ">   "
						}
					}

					messages = append(messages, renderDiscordMessage(edge, state, indent))
				}

				s.ChannelMessageSendReply(m.ChannelID, strings.Join(messages, "\n"), m.Reference())
			}
		}
	}
}

func renderDiscordMessage(edge twitch.TwitchSongNode, state, indent string) string {
	artists := []string{}
	for _, artist := range edge.Node.Artists {
		artists = append(artists, artist.Name)
	}
	sort.Strings(artists)

	durationMinutes := int(edge.Node.Duration) / 60
	durationSeconds := int(edge.Node.Duration) % 60
	return fmt.Sprintf(
		"%s%s: %s - %s - %s - %dm %ds",
		indent,
		state,
		edge.Node.Title,
		strings.Join(artists, ", "),
		strings.Join(edge.Node.Genres, ", "),
		durationMinutes,
		durationSeconds,
	)
}
