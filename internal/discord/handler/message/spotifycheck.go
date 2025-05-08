package message

import (
	"context"
	"fmt"
	"log/slog"
	"sort"
	"strings"

	"github.com/b1tray3r/isitstreamablebot/internal"
	"github.com/b1tray3r/isitstreamablebot/internal/twitch"
	"github.com/bwmarrin/discordgo"
	"github.com/zmb3/spotify/v2"
)

type SpotifyCheckHandler struct {
	client *spotify.Client
}

func NewSpotifyCheckHandler(client *spotify.Client) *SpotifyCheckHandler {
	return &SpotifyCheckHandler{
		client: client,
	}
}

func (h *SpotifyCheckHandler) Handle(s *discordgo.Session, m *discordgo.MessageCreate) {
	slog.Info("SpotifyCheckHandler", "messageID", m.ID, "channelID", m.ChannelID, "authorID", m.Author.ID)
	if m.Author.ID == s.State.User.ID {
		slog.Debug("Ignoring own message")
		return
	}

	slog.Debug("Checking for Spotify links", "messageID", m.ID, "channelID", m.ChannelID, "authorID", m.Author.ID)
	if internal.HasSpotifyLink(m.Content) {
		link := internal.ExtractSpotifyLinkFromText(m.Content)
		spotifyID := internal.ExtractSpotifyTrackIDFromLink(link)
		slog.Info("Extracted", "SpotifyID", spotifyID)
		s.MessageReactionAdd(m.ChannelID, m.ID, "⏳")

		trackID := spotify.ID(spotifyID)
		track, err := h.client.GetTrack(context.Background(), trackID)
		if err != nil {
			slog.Error("Spotify error", "err", err, "trackID", spotifyID)
			return
		}
		if track == nil {
			slog.Error("Track not found on Spotify")
			return
		}

		r, err := twitch.SendRequest(track.Name)
		if err != nil {
			slog.Error("TwitchDJCatalog error", "err", err)
			return
		}

		if len(r.Data.SearchDJCatalog.Edges) == 0 {
			slog.Error("No results found.")
			return
		}

		found := "❓"
		songs := []string{}
		for _, edge := range r.Data.SearchDJCatalog.Edges {
			state := "✅"
			if edge.Node.IsBlockedTrack {
				state = "❌"
			}

			if track.Name == edge.Node.Title {
				spotifyDuration := int(track.Duration / 1000)
				nodeDuration := int(edge.Node.Duration)
				if spotifyDuration == nodeDuration {
					found = state
					break
				}
			}

			songs = append(songs, renderDiscordMessage(edge, state, ""))
		}

		s.MessageReactionRemove(m.ChannelID, m.ID, "⏳", s.State.User.ID)
		if found != "❓" {
			s.MessageReactionAdd(m.ChannelID, m.ID, found)
		} else {
			s.MessageReactionAdd(m.ChannelID, m.ID, "❓")
			content := fmt.Sprintf("Twitch results:\n%s", strings.Join(songs, "\n"))
			if _, err := s.ChannelMessageSendReply(m.ChannelID, content, m.Reference()); err != nil {
				slog.Error("Error sending message", "err", err)
				return
			}
		}
	}
}

func renderDiscordMessage(edge twitch.TwitchSongNode, state, indent string) string {
	artists := []string{}
	for _, artist := range edge.Node.Artists {
		artists = append(artists, artist.Name)
	}
	sort.Strings(artists)

	durationMinutes := int(edge.Node.Duration) / 60
	durationSeconds := int(edge.Node.Duration) % 60
	return fmt.Sprintf(
		"%s%s: %s - %s - %s - %dm %ds",
		indent,
		state,
		edge.Node.Title,
		strings.Join(artists, ", "),
		strings.Join(edge.Node.Genres, ", "),
		durationMinutes,
		durationSeconds,
	)
}
