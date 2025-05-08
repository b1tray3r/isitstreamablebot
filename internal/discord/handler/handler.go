package handler

import (
	"github.com/bwmarrin/discordgo"
)

// CommandGetter is an interface for getting commands.
type CommandGetter interface {
	GetCommands() []*discordgo.ApplicationCommand
}

// InteractionFunctionHandler is an interface for handling command functions.
type InteractionFunctionHandler interface {
	Handle(s *discordgo.Session, i *discordgo.InteractionCreate)
}

type MessageFunctionHandler interface {
	Handle(s *discordgo.Session, m *discordgo.MessageCreate)
}

// BouncerFunc is a function type that checks if a user is allowed to use a command in a specific channel.
type BouncerFunc func(guildID, channelID string) bool

// CommandHandler is an interface for handling Discord commands.
type CommandHandler interface {
	InteractionFunctionHandler
	CommandGetter
}

type CustomIDGetter interface {
	GetCustomID() string
}

type InteractionHandler interface {
	InteractionFunctionHandler
	CustomIDGetter
}

type MessageHandler interface {
	MessageFunctionHandler
}
