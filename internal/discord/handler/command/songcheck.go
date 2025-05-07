package commandhandler

import (
	"strings"

	"github.com/b1tray3r/isitstreamablebot/internal/discord"
	"github.com/b1tray3r/isitstreamablebot/internal/discord/handler"
	"github.com/b1tray3r/isitstreamablebot/internal/twitch"
	"github.com/bwmarrin/discordgo"
	"golang.org/x/exp/slog"
)

// DJSongCheckCommandHandler is a handler for the DJSongCheck command.
type DJSongCheckCommandHandler struct {
	clientID    string
	bouncerFunc func(guildID, channelID string) bool
}

// NewDJSongCheckCommandHandler creates a new DJSongCheckCommandHandler with the given client ID.
func NewDJSongCheckCommandHandler(clientID string, bouncerFunc handler.BouncerFunc) handler.CommandHandler {
	return &DJSongCheckCommandHandler{
		clientID:    clientID,
		bouncerFunc: bouncerFunc,
	}
}

// Handle handles the DJSongCheck command interaction.
func (h *DJSongCheckCommandHandler) Handle(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if i.Type != discordgo.InteractionApplicationCommand {
		slog.Debug("Ignoring non-command interaction", "type", i.Type)
		return
	}

	if h.bouncerFunc != nil && !h.bouncerFunc(i.GuildID, i.ChannelID) {
		// Respond to the user indicating they are not allowed to use the command in this channel.
		err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "You are not allowed to use this command in this channel.",
			},
		})
		if err != nil {
			// Log the error if the response fails.
			slog.Error("Failed to respond to interaction", "err", err)
			return
		}
		return
	}

	requestedTitle := i.ApplicationCommandData().Options[0]
	songTitle := strings.ToLower(requestedTitle.StringValue())
	requestedArtist := ""
	if len(i.ApplicationCommandData().Options) > 1 {
		requestedArtist = i.ApplicationCommandData().Options[1].StringValue()
	}
	slog.Debug("Received DJSongCheck command", "title", requestedTitle.StringValue(), "artist", requestedArtist)
	if requestedTitle.Type != discordgo.ApplicationCommandOptionString {
		slog.Error("Invalid option type for title", "type", requestedTitle.Type)
		return
	}

	slog.Info("Handling DJSongCheck command", "guildID", i.GuildID, "channelID", i.ChannelID)
	response, err := twitch.SendRequest(songTitle)
	if err != nil {
		slog.Error("Failed to send request to Twitch API", "err", err)
		return
	}

	if len(response.Data.SearchDJCatalog.Edges) == 0 {
		err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "No data found for the given song title.",
			},
		})
		if err != nil {
			slog.Error("Failed to respond to interaction", "err", err)
			return
		}
		return
	}

	for _, edge := range response.Data.SearchDJCatalog.Edges {
		if songTitle != strings.ToLower(edge.Node.Title) {
			slog.Debug("Skipping song", "title", edge.Node.Title, "requestedTitle", requestedTitle.StringValue())
			continue
		}

		foundArtist := false
		for _, artist := range edge.Node.Artists {
			if requestedArtist != "" && artist.Name != requestedArtist {
				slog.Debug("Skipping artist", "artist", artist.Name, "requestedArtist", requestedArtist)
				continue
			}

			foundArtist = true
			break
		}

		if requestedArtist != "" && !foundArtist {
			err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "No data found for the given artist.",
				},
			})
			if err != nil {
				slog.Error("Failed to respond to interaction", "err", err)
				return
			}
			return
		}

		state := "✅"
		if edge.Node.IsBlockedTrack {
			state = "❌"
		}

		content := state + edge.Node.Title
		if requestedArtist != "" {
			content += " by " + requestedArtist
		}
		if edge.Node.IsBlockedTrack {
			content += " is ** not streamable! ** " + state
		}
		err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: content,
				Components: []discordgo.MessageComponent{
					discordgo.ActionsRow{
						Components: []discordgo.MessageComponent{
							discordgo.Button{
								Label:    "Watch this song",
								Style:    discordgo.SuccessButton,
								CustomID: discord.WATCHLIST_ADD,
							},
						},
					},
				},
			},
		})
		if err != nil {
			slog.Error("Failed to respond to interaction", "err", err)
			return
		}

		return
	}
}

// GetCommand returns the command associated with the handler.
func (h *DJSongCheckCommandHandler) GetCommands() []*discordgo.ApplicationCommand {
	return []*discordgo.ApplicationCommand{
		{
			Name:        "check",
			Description: "Check a song title against the Twitch DJ catalog",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "title",
					Description: "The title of the song to check",
					Required:    true,
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "artist",
					Description: "The artist of the song to check",
					Required:    false,
				},
			},
		},
	}
}
