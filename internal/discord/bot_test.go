package discord

import (
	"testing"

// Removed the import of github.com/stretchr/testify/assert as per project guidelines.
)

func TestGuildBouncer(t *testing.T) {
	guildIDs := []string{"123", "456", "789"}
	bouncer := NewGuildBouncer(guildIDs)

	if !bouncer.Check("123") {
		t.Errorf("Expected guild ID '123' to be allowed")
	}
	if !bouncer.Check("456") {
		t.Errorf("Expected guild ID '456' to be allowed")
	}
	if bouncer.Check("000") {
		t.Errorf("Expected guild ID '000' to be disallowed")
	}
}

func TestChannelBouncer(t *testing.T) {
	channelIDs := []string{"abc", "def", "ghi"}
	bouncer := NewChannelBouncer(channelIDs)

	if !bouncer.Check("abc") {
		t.Errorf("Expected channel ID 'abc' to be allowed")
	}
	if !bouncer.Check("def") {
		t.Errorf("Expected channel ID 'def' to be allowed")
	}
	if bouncer.Check("xyz") {
		t.Errorf("Expected channel ID 'xyz' to be disallowed")
	}
}
