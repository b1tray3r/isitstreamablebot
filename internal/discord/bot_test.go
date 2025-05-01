package discord

import (
	"testing"

	"github.com/bwmarrin/discordgo"
	"github.com/stretchr/testify/assert"
)

func TestGuildBouncer(t *testing.T) {
	guildIDs := []string{"123", "456", "789"}
	bouncer := NewGuildBouncer(guildIDs)

	assert.True(t, bouncer.check("123"), "Expected guild ID '123' to be allowed")
	assert.True(t, bouncer.check("456"), "Expected guild ID '456' to be allowed")
	assert.False(t, bouncer.check("000"), "Expected guild ID '000' to be disallowed")
}

func TestChannelBouncer(t *testing.T) {
	channelIDs := []string{"abc", "def", "ghi"}
	bouncer := NewChannelBouncer(channelIDs)

	assert.True(t, bouncer.check("abc"), "Expected channel ID 'abc' to be allowed")
	assert.True(t, bouncer.check("def"), "Expected channel ID 'def' to be allowed")
	assert.False(t, bouncer.check("xyz"), "Expected channel ID 'xyz' to be disallowed")
}

func TestBot_IsGuildAllowed(t *testing.T) {
	guildIDs := []string{"123", "456"}
	guildBouncer := NewGuildBouncer(guildIDs)
	bot, _ := NewBot(guildBouncer, nil, nil)

	assert.True(t, bot.IsGuildAllowed("123"), "Expected guild ID '123' to be allowed")
	assert.False(t, bot.IsGuildAllowed("789"), "Expected guild ID '789' to be disallowed")
}

func TestBot_IsChannelAllowed(t *testing.T) {
	channelIDs := []string{"abc", "def"}
	channelBouncer := NewChannelBouncer(channelIDs)
	bot, _ := NewBot(nil, channelBouncer, nil)

	assert.True(t, bot.IsChannelAllowed("abc"), "Expected channel ID 'abc' to be allowed")
	assert.False(t, bot.IsChannelAllowed("xyz"), "Expected channel ID 'xyz' to be disallowed")
}

func TestBot_Shutdown(t *testing.T) {
	session := &discordgo.Session{}
	bot, _ := NewBot(nil, nil, session)

	bot.Shutdown()
	assert.Nil(t, bot.session, "Expected session to be nil after shutdown")
}
