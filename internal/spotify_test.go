package internal

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestExtractSpotifyLinkFromText(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		expected string
	}{
		{
			name:     "Single Spotify link",
			text:     "https://open.spotify.com/track/0WVgnZcUUvQVoaa2gEnv3Q?si=5e37c0a62f894f84 what a banger",
			expected: "https://open.spotify.com/track/0WVgnZcUUvQVoaa2gEnv3Q?si=5e37c0a62f894f84",
		},
		{
			name:     "Multiple Spotify links",
			text:     "Here are two songs: https://open.spotify.com/track/0WVgnZcUUvQVoaa2gEnv3Q?si=5e37c0a62f894f84 and https://open.spotify.com/track/1a2b3c4d5e6f7g8h9i0jklmnopqrstuvwx?si=1234567890abcdef",
			expected: "",
		},
		{
			name:     "No Spotify links",
			text:     "This text contains no Spotify links.",
			expected: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ExtractSpotifyLinkFromText(tt.text)
			if !cmp.Equal(result, tt.expected) {
				t.Errorf("ExtractSpotifyLinksFromText() mismatch (-want +got):\n%s", cmp.Diff(tt.expected, result))
			}
		})
	}
}

func TestExtractSpotifyTrackIDFromLink(t *testing.T) {
	tests := []struct {
		name     string
		link     string
		expected string
	}{
		{
			name:     "Valid Spotify link with query string",
			link:     "https://open.spotify.com/track/0WVgnZcUUvQVoaa2gEnv3Q?si=5e37c0a62f894f84",
			expected: "0WVgnZcUUvQVoaa2gEnv3Q",
		},
		{
			name:     "Valid Spotify link without query string",
			link:     "https://open.spotify.com/track/0WVgnZcUUvQVoaa2gEnv3Q",
			expected: "0WVgnZcUUvQVoaa2gEnv3Q",
		},
		{
			name:     "Invalid Spotify link",
			link:     "https://example.com/track/invalid",
			expected: "",
		},
		{
			name:     "Empty string",
			link:     "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ExtractSpotifyTrackIDFromLink(tt.link)
			if result != tt.expected {
				t.Errorf("ExtractSpotifyTrackIDFromLink() = %v, want %v", result, tt.expected)
			}
		})
	}
}
