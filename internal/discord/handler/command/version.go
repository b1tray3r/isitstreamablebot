package commandhandler

import (
	"github.com/b1tray3r/isitstreamablebot/internal/discord/handler"
	"github.com/bwmarrin/discordgo"
	"golang.org/x/exp/slog"
)

// VersionCommandHandler is a handler for the version command.
type VersionCommandHandler struct {
	version string
}

// NewVersionCommandHandler creates a new VersionCommandHandler with the given version.
func NewVersionCommandHandler(version string) handler.CommandHandler {
	return &VersionCommandHandler{
		version: version,
	}
}

// Handle handles the version command interaction.
func (h *VersionCommandHandler) Handle(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if i.Type != discordgo.InteractionApplicationCommand {
		slog.Debug("Ignoring non-command interaction", "type", i.Type)
		return
	}

	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "Version: " + h.version,
		},
	})
	if err != nil {
		slog.Error("Failed to respond to interaction", "err", err)
		return
	}
}

// GetCommands returns the commands associated with the handler.
func (h *VersionCommandHandler) GetCommands() []*discordgo.ApplicationCommand {
	return []*discordgo.ApplicationCommand{
		{
			Name:        "version",
			Description: "Get the version of the bot",
			Options:     nil,
		},
	}
}
