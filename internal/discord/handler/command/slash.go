package commandhandler

import (
	"github.com/b1tray3r/isitstreamablebot/internal/discord/handler"
	"github.com/bwmarrin/discordgo"
	"golang.org/x/exp/slog"
)

// SlashCommandHandler is a handler for slash commands.
type SlashCommandHandler struct {
	handlers map[string]handler.CommandHandler
}

// NewSlashCommandHandler creates a new SlashCommandHandler with the given handlers.
func NewSlashCommandHandler(handlers []handler.CommandHandler) *SlashCommandHandler {
	handlersMap := make(map[string]handler.CommandHandler)
	for _, handler := range handlers {
		for _, command := range handler.GetCommands() {
			slog.Debug("Registering command", "command", command.Name, "description", command.Description)
			if command != nil {
				handlersMap[command.Name] = handler
			}
		}
	}

	return &SlashCommandHandler{
		handlers: handlersMap,
	}
}

// Handle handles the slash command interaction.
func (h *SlashCommandHandler) GetCommands() []*discordgo.ApplicationCommand {
	cmds := make([]*discordgo.ApplicationCommand, 0)
	for _, handler := range h.handlers {
		cmds = append(cmds, handler.GetCommands()...)
	}
	return cmds
}

// Handle handles the slash command interaction.
func (h *SlashCommandHandler) Handle(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if i.Type != discordgo.InteractionApplicationCommand {
		slog.Debug("Ignoring non-command interaction", "type", i.Type)
		return
	}

	slog.Info("cmds", "commandName", i.ApplicationCommandData().Name, "guildID", i.GuildID, "channelID", i.ChannelID, "handlers", h.handlers)

	commandName := i.ApplicationCommandData().Name
	handler, ok := h.handlers[commandName]
	if !ok {
		err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Unknown command!",
			},
		})
		if err != nil {
			slog.Error("Failed to respond to interaction", "err", err)
			return
		}
		return
	}
	handler.Handle(s, i)
}
