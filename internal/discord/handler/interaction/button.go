package interaction

import (
	"log/slog"
	"strings"

	"github.com/b1tray3r/isitstreamablebot/internal/discord/handler"
	"github.com/bwmarrin/discordgo"
)

// ButtonHandler is a handler for button interactions.
type ButtonHandler struct {
	handlers map[string]handler.InteractionHandler
}

// NewButtonHandler creates a new ButtonHandler with the given handlers.
func NewButtonHandler(handlers []handler.InteractionHandler) *ButtonHandler {
	handlersMap := make(map[string]handler.InteractionHandler)
	for _, handler := range handlers {
		handlersMap[handler.GetCustomID()] = handler
	}

	return &ButtonHandler{
		handlers: handlersMap,
	}
}

// Handle handles the slash command interaction.
func (h *ButtonHandler) Handle(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if i.Type != discordgo.InteractionMessageComponent {
		return
	}

	customID := i.MessageComponentData().CustomID
	for id, handler := range h.handlers {
		if strings.HasPrefix(customID, id) {
			slog.Debug("Handling button interaction", "customID", customID, "handler", id)
			handler.Handle(s, i)
			return
		}
	}
}

func (h *ButtonHandler) GetCustomID() string {
	return ""
}
