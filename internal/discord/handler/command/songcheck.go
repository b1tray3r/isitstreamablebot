package commandhandler

import (
	"strings"

	"github.com/b1tray3r/isitstreamablebot/internal/discord/handler"
	"github.com/b1tray3r/isitstreamablebot/internal/store"
	"github.com/b1tray3r/isitstreamablebot/internal/twitch"
	pkg "github.com/b1tray3r/isitstreamablebot/pkg/db"
	"github.com/bwmarrin/discordgo"
	"golang.org/x/exp/slog"
)

const (
	NO_DATA_FOUND = "No data found for the given song title."
)

// DJSongCheckCommandHandler is a handler for the DJSongCheck command.
type DJSongCheckCommandHandler struct {
	clientID    string
	storage     store.Storager
	bouncerFunc func(guildID, channelID string) bool
}

// NewDJSongCheckCommandHandler creates a new DJSongCheckCommandHandler with the given client ID.
func NewDJSongCheckCommandHandler(clientID string, storage store.Storager, bouncerFunc handler.BouncerFunc) handler.CommandHandler {
	return &DJSongCheckCommandHandler{
		clientID:    clientID,
		storage:     storage,
		bouncerFunc: bouncerFunc,
	}
}

func (h *DJSongCheckCommandHandler) extractOptions(options []*discordgo.ApplicationCommandInteractionDataOption) (string, string) {
	var requestedTitle, requestedArtist string
	if len(options) > 0 {
		requestedTitle = options[0].StringValue()
	}
	if len(options) > 1 {
		requestedArtist = options[1].StringValue()
	}
	return strings.ToLower(requestedTitle), strings.ToLower(requestedArtist)
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

	requestedTitle, requestedArtist := h.extractOptions(i.ApplicationCommandData().Options)

	slog.Info("Handling DJSongCheck command", "guildID", i.GuildID, "channelID", i.ChannelID)
	response, err := twitch.SendRequest(requestedTitle)
	if err != nil {
		slog.Error("Failed to send request to Twitch API", "err", err)
		return
	}

	songList := response.GetSongList()
	if songList == nil {
		err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: NO_DATA_FOUND,
			},
		})
		if err != nil {
			slog.Error("Failed to respond to interaction", "err", err)
		}
		return
	}

	var song *pkg.Song
	for _, songNode := range songList {
		if strings.ToLower(songNode.Title) != requestedTitle {
			continue
		}

		artistFound := false
		for _, artist := range songNode.Artists {
			if strings.ToLower(artist.Name) == requestedArtist {
				artistFound = true
				break
			}
		}

		if !artistFound {
			continue
		}

		artists := ""
		for _, artist := range songNode.Artists {
			artists += artist.Name + ", "
		}
		artists = strings.TrimSuffix(artists, ", ")

		streamable := int64(1)
		if songNode.IsBlockedTrack {
			streamable = int64(0)
		}

		song = &pkg.Song{
			ID:           songNode.ID,
			Title:        songNode.Title,
			Artists:      artists,
			Duration:     int64(songNode.Duration),
			IsStreamable: streamable,
			MessageID:    i.ID,
		}

		// add song to db

		break
	}

	slog.Debug("Found song", "songID", song.ID, "title", song.Title, "artist", song.Artists)

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
