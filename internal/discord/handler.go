package discord

import (
	"github.com/bwmarrin/discordgo"
	"golang.org/x/exp/slog"
)

type CommandGetter interface {
	GetCommand() *discordgo.ApplicationCommand
}

type CommandFuncHandler interface {
	Handle(s *discordgo.Session, i *discordgo.InteractionCreate)
}

// CommandHandler is an interface for handling Discord commands.
type CommandHandler interface {
	CommandFuncHandler
	CommandGetter
}

// SlashCommandHandler is a handler for slash commands.
type SlashCommandHandler struct {
	handlers map[string]CommandHandler
}

// NewSlashCommandHandler creates a new SlashCommandHandler with the given handlers.
func NewSlashCommandHandler(handlers []CommandHandler) *SlashCommandHandler {
	handlersMap := make(map[string]CommandHandler)
	for _, handler := range handlers {
		command := handler.GetCommand()
		if command != nil {
			handlersMap[command.Name] = handler
		}
	}

	return &SlashCommandHandler{
		handlers: handlersMap,
	}
}

// Handle handles the slash command interaction.
func (h *SlashCommandHandler) GetCommand() *discordgo.ApplicationCommand {
	return nil
}

// Handle handles the slash command interaction.
func (h *SlashCommandHandler) Handle(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if i.Type != discordgo.InteractionApplicationCommand {
		slog.Debug("Ignoring non-command interaction", "type", i.Type)
		return
	}

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

// VersionCommandHandler is a handler for the version command.
type VersionCommandHandler struct {
	version string
}

// Handle handles the version command interaction.
func (h *VersionCommandHandler) Handle(s *discordgo.Session, i *discordgo.InteractionCreate) {
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

// GetCommand returns the command associated with the handler.
func (h *VersionCommandHandler) GetCommand() *discordgo.ApplicationCommand {
	return &discordgo.ApplicationCommand{
		Name:        "version",
		Description: "Get the version of the bot",
		Options:     nil,
	}
}

// NewVersionCommandHandler creates a new VersionCommandHandler with the given version.
func NewVersionCommandHandler(version string) CommandHandler {
	return &VersionCommandHandler{
		version: version,
	}
}

// DJSongCheckCommandHandler is a handler for the DJSongCheck command.
type DJSongCheckCommandHandler struct {
	clientID string
}

// Handle handles the DJSongCheck command interaction.
func (h *DJSongCheckCommandHandler) Handle(s *discordgo.Session, i *discordgo.InteractionCreate) {
	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "DJSongCheck command is not implemented yet.",
		},
	})
	if err != nil {
		slog.Error("Failed to respond to interaction", "err", err)
		return
	}
}

// GetCommand returns the command associated with the handler.
func (h *DJSongCheckCommandHandler) GetCommand() *discordgo.ApplicationCommand {
	return &discordgo.ApplicationCommand{
		Name:        "songcheck",
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
	}
}

// NewDJSongCheckCommandHandler creates a new DJSongCheckCommandHandler with the given client ID.
func NewDJSongCheckCommandHandler(clientID string) CommandHandler {
	return &DJSongCheckCommandHandler{
		clientID: clientID,
	}
}
