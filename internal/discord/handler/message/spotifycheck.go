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

		songList := r.GetSongList()
		if songList == nil {
			slog.Error("No songs found in TwitchDJCatalog")
			return
		}

		songs := []pkg.Song{}
		for _, songnode := range songList {
			streamable := int64(1)
			if songnode.IsBlockedTrack {
				streamable = int64(0)
			}

			artists := ""
			for _, artist := range songnode.Artists {
				artists += artist.Name + ", "
			}
			artists = strings.TrimSuffix(artists, ", ")

			songs = append(songs, pkg.Song{
				ID:           songnode.ID,
				Title:        songnode.Title,
				Artists:      artists,
				Duration:     int64(songnode.Duration),
				IsStreamable: streamable,
			})
		}

		var found *pkg.Song
		alternatives := []pkg.Song{}
		for _, song := range songs {
			spotifyDuration := int(track.Duration / 1000)
			nodeDuration := int(song.Duration)
			if spotifyDuration == nodeDuration {
				if song.Title != track.Name {
					continue
				}

				trackArtists := ""
				for _, artist := range track.Artists {
					trackArtists += artist.Name + ", "
				}
				trackArtists = strings.TrimSuffix(trackArtists, ", ")
				if song.Artists != trackArtists {
					continue
				}

				found = &song
			}

			alternatives = append(alternatives, song)
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
							MessageID:    m.ID,
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
			// SpotifyCheckHandler opens a thread for discussion and closes it after sending alternatives.
			thread, err := s.MessageThreadStart(m.ChannelID, m.ID, "Alternative song results", 60)
			if err != nil {
				slog.Error("Error creating thread", "err", err)
				return
			}

			// Send alternatives in the thread if any exist.
			if len(alternatives) > 0 {
				content := discord.RenderDiscordMessage("Other Twitch results", alternatives)
				if _, err := s.ChannelMessageSend(thread.ID, content); err != nil {
					slog.Error("Error sending alternatives to thread", "err", err)
				}
			}

			defer func() {
				// SpotifyCheckHandler closes the discussion thread after sending alternatives.
				_, err := s.ChannelEditComplex(thread.ID, &discordgo.ChannelEdit{
					Archived: func(b bool) *bool { return &b }(true),
				})
				if err != nil {
					slog.Error("Error closing thread", "err", err)
				}
			}()
		}
	}
}
