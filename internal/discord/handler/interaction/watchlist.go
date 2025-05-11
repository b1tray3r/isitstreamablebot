package interaction

import (
	"fmt"
	"log/slog"
	"strings"

	"github.com/b1tray3r/isitstreamablebot/internal/watchlist"
	"github.com/bwmarrin/discordgo"
)

type WatchlistHandler struct {
	customID  string
	watchlist *watchlist.Watchlist
}

func NewWatchlistHandler(watchlist *watchlist.Watchlist, customID string) *WatchlistHandler {
	return &WatchlistHandler{
		customID:  customID,
		watchlist: watchlist,
	}
}

func (h *WatchlistHandler) Handle(s *discordgo.Session, i *discordgo.InteractionCreate) {
	id := strings.TrimPrefix(i.MessageComponentData().CustomID, h.customID)

	slog.Info("Adding to watchlist", "id", id)
	song, err := h.watchlist.Add(i.Member.User.ID, id)
	if err != nil {
		slog.Error("Failed to add to watchlist", "error", err)
		if err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Failed to add to watchlist: " + err.Error(),
			},
		}); err != nil {
			slog.Error("Failed to send interaction response", "error", err)
		}
		return
	}

	songs, err := h.watchlist.List(i.Member.User.ID)
	if err != nil {
		slog.Error("Failed to get watchlist", "error", err)
		return
	}
	slog.Debug("Watchlist after adding", "watchlist", songs)

	if err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "Added to watchlist: " + fmt.Sprintf("%s - %s (%ds)", song.Title, song.Artists, song.Duration),
		},
	}); err != nil {
		slog.Error("Failed to send interaction response", "error", err)
	}
}

func (h *WatchlistHandler) GetCustomID() string {
	return h.customID
}
