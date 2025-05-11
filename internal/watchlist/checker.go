package watchlist

import (
	"context"
	"log"
	"log/slog"
	"time"

	"github.com/b1tray3r/isitstreamablebot/internal/store"
	pkg "github.com/b1tray3r/isitstreamablebot/pkg/db"
)

type Checker struct {
	store     store.Storager
	sleepTime time.Duration
}

func NewChecker(store store.Storager, sleepTime time.Duration) *Checker {
	return &Checker{
		store: store,

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
	return nil
}
