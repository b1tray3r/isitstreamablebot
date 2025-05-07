package commandhandler

import (
	"github.com/b1tray3r/isitstreamablebot/internal/discord/handler"
	"github.com/b1tray3r/isitstreamablebot/internal/watchlist"
	"github.com/bwmarrin/discordgo"
	"golang.org/x/exp/slog"
)

// ListWatchlistCommandHandler is a handler for the version command.
type ListWatchlistCommandHandler struct {
	watchlist *watchlist.Watchlist
}

// NewListWatchlistCommandHandler creates a new ListWatchlistCommandHandler with the given version.
func NewListWatchlistCommandHandler(watchlist *watchlist.Watchlist) handler.CommandHandler {
	return &ListWatchlistCommandHandler{
		watchlist: watchlist,
	}
}

// Handle handles the version command interaction.
func (h *ListWatchlistCommandHandler) Handle(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if i.Type != discordgo.InteractionApplicationCommand {
		slog.Debug("Ignoring non-command interaction", "type", i.Type)
		return
	}

	wl := ""
	for _, element := range h.watchlist.List() {
		wl += element + "\n"
	}

	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "Watchlist:\n" + wl,
		},
	})
	if err != nil {
		slog.Error("Failed to respond to interaction", "err", err)
		return
	}
}

// GetCommands returns the commands associated with the handler.
func (h *ListWatchlistCommandHandler) GetCommands() []*discordgo.ApplicationCommand {
	return []*discordgo.ApplicationCommand{
		{
			Name:        "watchlist",
			Description: "Get the watchlist",
			Options:     nil,
		},
	}
}
