package discord

import (
	"fmt"

	"github.com/b1tray3r/isitstreamablebot/internal/discord/handler"
	"github.com/bwmarrin/discordgo"
	"github.com/sagikazarmark/slog-shim"
)

// CommandHandler is an interface for handling Discord commands.
type Session struct {
	session  *discordgo.Session
	commands []*discordgo.ApplicationCommand
}

// CommandFuncHandler is an interface for handling command functions.
func (s *Session) Close() {
	if s.session != nil {
		s.session.Close()
	}
}

// NewSession creates a new Discord session with the given token and command handlers.
func NewSession(token string, commandHandler []handler.CommandHandler, interactionHandler []handler.InteractionHandler, messageHandler []handler.MessageHandler) (*Session, error) {
	ds, err := discordgo.New("Bot " + token)
	if err != nil {
		return nil, err
	}

	if err := ds.Open(); err != nil {
		return nil, err
	}

	for _, handler := range commandHandler {
		for _, command := range handler.GetCommands() {
			if command != nil {
				slog.Debug("Registering command", "command", command.Name, "description", command.Description)
				ds.ApplicationCommandCreate(ds.State.User.ID, "", command)
			}
		}
		slog.Debug("Registering handler", "handlerType", fmt.Sprintf("%T", handler))
		ds.AddHandler(handler.Handle)
	}

	for _, handler := range interactionHandler {
		slog.Debug("Registering interaction handler", "handlerType", fmt.Sprintf("%T", handler))
		ds.AddHandler(handler.Handle)
	}

	for _, handler := range messageHandler {
		slog.Debug("Registering message handler", "handlerType", fmt.Sprintf("%T", handler))
		ds.AddHandler(handler.Handle)
	}

	session := &Session{
		session:  ds,
		commands: make([]*discordgo.ApplicationCommand, 0),
	}

	return session, nil
}
