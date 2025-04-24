package internal

import (
	"context"
	"fmt"
	"log/slog"
	"runtime/debug"
	"sort"
	"strings"

	"github.com/b1tray3r/isitstreamablebot/internal/twitch"
	"github.com/bwmarrin/discordgo"
	"github.com/zmb3/spotify/v2"
)

type DiscordBot struct {
	Session       *discordgo.Session
	SpotifyClient *spotify.Client

	Commands []*discordgo.ApplicationCommand

	ChannelFilter []string
}

func NewDiscordBot(token string, channelIDs []string) (*DiscordBot, error) {
	dg, err := discordgo.New("Bot " + token)
	if err != nil {
		return nil, fmt.Errorf("error creating Discord session: %w", err)
	}

	commands := []*discordgo.ApplicationCommand{
		{
			Name:        "version",
			Description: "Replies with the version of the bot.",
		},
	}

	if err := dg.Open(); err != nil {
		return nil, fmt.Errorf("error opening Discord session: %w", err)
	}

	bot := &DiscordBot{
		Session:       dg,
		SpotifyClient: nil,
		ChannelFilter: channelIDs,
		Commands:      commands,
	}

	for _, command := range commands {
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
		var gitCommit string
		if info, ok := debug.ReadBuildInfo(); ok {
			for _, setting := range info.Settings {
				if setting.Key == "vcs.revision" {
					gitCommit = setting.Value
				}
			}
		}
		if gitCommit == "" {
			gitCommit = "unknown"
		}
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Git commit: " + gitCommit,
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

					artists := []string{}
					for _, artist := range edge.Node.Artists {
						artists = append(artists, artist.Name)
					}
					sort.Strings(artists)

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

					durationMinutes := int(edge.Node.Duration) / 60
					durationSeconds := int(edge.Node.Duration) % 60
					messages = append(messages, fmt.Sprintf(
						"%s%s: %s - %s - %s - %dm %ds",
						indent,
						state,
						edge.Node.Title,
						strings.Join(artists, ", "),
						strings.Join(edge.Node.Genres, ", "),
						durationMinutes,
						durationSeconds,
					))
				}

				s.ChannelMessageSendReply(m.ChannelID, strings.Join(messages, "\n"), m.Reference())
			}
		}
	}
}
