package discord

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGuildBouncer(t *testing.T) {
	guildIDs := []string{"123", "456", "789"}
	bouncer := NewGuildBouncer(guildIDs)

	assert.True(t, bouncer.Check("123"), "Expected guild ID '123' to be allowed")
	assert.True(t, bouncer.Check("456"), "Expected guild ID '456' to be allowed")
	assert.False(t, bouncer.Check("000"), "Expected guild ID '000' to be disallowed")
}

func TestChannelBouncer(t *testing.T) {
	channelIDs := []string{"abc", "def", "ghi"}
	bouncer := NewChannelBouncer(channelIDs)

	assert.True(t, bouncer.Check("abc"), "Expected channel ID 'abc' to be allowed")
	assert.True(t, bouncer.Check("def"), "Expected channel ID 'def' to be allowed")
	assert.False(t, bouncer.Check("xyz"), "Expected channel ID 'xyz' to be disallowed")
}
