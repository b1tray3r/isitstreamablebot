package discord

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/sagikazarmark/slog-shim"
)

type Session struct {
	session  *discordgo.Session
	commands []*discordgo.ApplicationCommand
}

func (s *Session) Close() {
	if s.session != nil {
		s.session.Close()
	}
}

func NewSession(token string, handlers []CommandHandler) (*Session, error) {
	ds, err := discordgo.New("Bot " + token)
	if err != nil {
		return nil, err
	}

	if err := ds.Open(); err != nil {
		return nil, err
	}

	for _, handler := range handlers {
		command := handler.GetCommand()
		if command != nil {
			slog.Debug("Registering command", "command", command.Name, "description", command.Description)
			if _, err := ds.ApplicationCommandCreate(ds.State.User.ID, "", command); err != nil {
				return nil, err
			}
		}
		slog.Debug("Registering handler", "handlerType", fmt.Sprintf("%T", handler))
		ds.AddHandler(handler.Handle)
	}

	session := &Session{
		session:  ds,
		commands: make([]*discordgo.ApplicationCommand, 0),
	}

	return session, nil
}
