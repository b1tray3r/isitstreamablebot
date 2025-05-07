package handler

import (
	"github.com/bwmarrin/discordgo"
)

// CommandGetter is an interface for getting commands.
type CommandGetter interface {
	GetCommands() []*discordgo.ApplicationCommand
}

// HandleFuncHandler is an interface for handling command functions.
type HandleFuncHandler interface {
	Handle(s *discordgo.Session, i *discordgo.InteractionCreate)
}

// BouncerFunc is a function type that checks if a user is allowed to use a command in a specific channel.
type BouncerFunc func(guildID, channelID string) bool

// CommandHandler is an interface for handling Discord commands.
type CommandHandler interface {
	HandleFuncHandler
	CommandGetter
}

type CustomIDGetter interface {
	GetCustomID() string
}

type MessageHandler interface {
	HandleFuncHandler
	CustomIDGetter
}
