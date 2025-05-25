package watchlist

import (
	"context"
	"log"
	"log/slog"
	"time"

	"github.com/b1tray3r/isitstreamablebot/internal/discord"
	"github.com/b1tray3r/isitstreamablebot/internal/store"
	"github.com/b1tray3r/isitstreamablebot/internal/twitch"
	pkg "github.com/b1tray3r/isitstreamablebot/pkg/db"
)

type Checker struct {
	store     store.Storager
	session   *discord.Session
	sleepTime time.Duration
}

func NewChecker(store store.Storager, session *discord.Session, sleepTime time.Duration) *Checker {
	return &Checker{
		store:   store,
		session: session,

		sleepTime: sleepTime,
	}
}

// StartWatchlistChecker starts a goroutine that runs every 2 days to check songs on the watchlist.
func (c *Checker) StartWatchlistChecker(ctx context.Context) {
	go func() {
		ticker := time.NewTicker(c.sleepTime) // 2 days interval
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				// Context canceled, stop the goroutine.
				log.Println("Watchlist checker stopped")
				return
			case <-ticker.C:
				slog.Info("Checking watchlist", "interval", c.sleepTime)
				// Perform the check for songs on the watchlist.
				c.checkWatchlist(ctx)
			}
		}
	}()
}

// checkWatchlist checks every song on the watchlist.
func (c *Checker) checkWatchlist(ctx context.Context) {
	// Retrieve the list of songs from the store.
	songs, err := c.store.GetQueries().GetUniqueSongsInWatchlist(ctx)
	if err != nil {
		log.Printf("Failed to retrieve watchlist: %v", err)
		return
	}

	// Iterate over each song and perform the necessary checks.
	for _, song := range songs {
		// Example: Check if the song is streamable (implementation depends on your requirements).
		err := c.checkSong(ctx, song)
		if err != nil {
			log.Printf("Failed to check song %v: %v", song, err)
		}
	}
}

// checkSong performs the necessary checks for a single song.
func (c *Checker) checkSong(ctx context.Context, song pkg.Song) error {
	if time.Since(song.RequestTime) < 3*24*time.Hour {
		return nil
	}

	response, err := twitch.SendRequest(song.Title)
	if err != nil {
		log.Printf("Failed to send request for song %v: %v", song, err)
		return err
	}

	if response != nil {
		for _, edge := range response.Data.SearchDJCatalog.Edges {
			if edge.Node.ID != song.ID {
				continue
			}

			if !edge.Node.IsBlockedTrack {
				// Song is no longer blocked, update the database.
				if err := c.store.GetQueries().SetSongStreamable(ctx, song.ID); err != nil {
					log.Printf("Failed to update song %v as streamable: %v", song.ID, err)
					return err
				}

			}
		}
	}
	return nil
}
