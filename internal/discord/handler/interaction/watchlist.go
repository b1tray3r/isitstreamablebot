package interaction

import (
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
	id := strings.TrimPrefix(i.Message.ID, h.customID)

	slog.Info("Adding to watchlist", "id", id)
	h.watchlist.Add(id)
	slog.Debug("Watchlist after adding", "watchlist", h.watchlist.List())

	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "Added to watchlist: " + id,
		},
	})
	if err != nil {
		slog.Error("Failed to send interaction response", "error", err)
	}
}

func (h *WatchlistHandler) GetCustomID() string {
	return h.customID
}
