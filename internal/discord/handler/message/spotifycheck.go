package message

import (
	"context"
	"database/sql"
	"log/slog"
	"strings"
	"time"

	"github.com/b1tray3r/isitstreamablebot/internal"
	"github.com/b1tray3r/isitstreamablebot/internal/discord"
	"github.com/b1tray3r/isitstreamablebot/internal/store"
	"github.com/b1tray3r/isitstreamablebot/internal/twitch"
	pkg "github.com/b1tray3r/isitstreamablebot/pkg/db"
	"github.com/bwmarrin/discordgo"
	"github.com/zmb3/spotify/v2"
)

type SpotifyCheckHandler struct {
	client  *spotify.Client
	storage store.Storager
}

func NewSpotifyCheckHandler(client *spotify.Client, storage store.Storager) *SpotifyCheckHandler {
	return &SpotifyCheckHandler{
		client:  client,
		storage: storage,
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

		songList := []pkg.Song{}
		for _, edge := range r.Data.SearchDJCatalog.Edges {
			streamable := int64(1)
			if edge.Node.IsBlockedTrack {
				streamable = int64(0)
			}

			artists := ""
			for _, artist := range track.Artists {
				artists += artist.Name + ", "
			}
			artists = strings.TrimSuffix(artists, ", ")

			songList = append(songList, pkg.Song{
				ID:           edge.Node.ID,
				Title:        edge.Node.Title,
				Artists:      artists,
				Duration:     int64(edge.Node.Duration),
				IsStreamable: streamable,
			})
		}

		var found *pkg.Song
		alternatives := []pkg.Song{}
		for _, song := range songList {
			if song.Title == track.Name {
				spotifyDuration := int(track.Duration / 1000)
				nodeDuration := int(song.Duration)
				if spotifyDuration == nodeDuration {
					found = &song
					break
				}
			} else {
				alternatives = append(alternatives, song)
			}
		}

		s.MessageReactionRemove(m.ChannelID, m.ID, "⏳", s.State.User.ID)
		if found != nil {
			if found.IsStreamable == 0 {
				s.MessageReactionAdd(m.ChannelID, m.ID, "❌")
				song, err := h.storage.GetQueries().GetSongById(context.Background(), found.ID)
				if err != nil {
					if err != sql.ErrNoRows {
						slog.Error("Error getting song from database", "err", err)
						return
					}
				}

				if song.ID == "" { // not yet in the database
					song, err = h.storage.GetQueries().InsertSong(context.Background(),
						pkg.InsertSongParams{
							ID:           found.ID,
							Title:        found.Title,
							Artists:      found.Artists,
							Duration:     int64(found.Duration),
							IsStreamable: found.IsStreamable,
							RequestTime:  time.Now(),
						},
					)
					if err != nil {
						slog.Error("Error inserting song into database", "err", err)
						return
					}
				}

				// Create a message with the button
				message := discordgo.MessageSend{
					Content: "This song is blocked. You can add it to your watchlist and I'll inform you when it becomes available.",
					Components: []discordgo.MessageComponent{
						discordgo.ActionsRow{
							Components: []discordgo.MessageComponent{
								discordgo.Button{
									Label:    "Watch this song",
									Style:    discordgo.SuccessButton,
									CustomID: discord.WATCHLIST_ADD + song.ID,
								},
							},
						},
					},
				}

				// Send the message with the button
				if _, err := s.ChannelMessageSendComplex(m.ChannelID, &message); err != nil {
					slog.Error("Error sending message with button", "err", err)
					return
				}

				return // Early return after sending the message
			}

			s.MessageReactionAdd(m.ChannelID, m.ID, "✅") // Song is streamable

		} else {
			s.MessageReactionAdd(m.ChannelID, m.ID, "❓")
			content := discord.RenderDiscordMessage("Other Twitch results", alternatives)
			if _, err := s.ChannelMessageSendReply(m.ChannelID, content, m.Reference()); err != nil {
				slog.Error("Error sending message", "err", err)
				return
			}
		}
	}
}
