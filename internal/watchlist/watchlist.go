package watchlist

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/b1tray3r/isitstreamablebot/internal/store"
	pkg "github.com/b1tray3r/isitstreamablebot/pkg/db"
)

// Watchlist is a struct that represents a user's watchlist.
type Watchlist struct {
	store store.Storager
}

// NewWatchlist creates a new Watchlist instance with the given store.
func NewWatchlist(store store.Storager) *Watchlist {
	return &Watchlist{
		store: store,
	}
}

// Add adds a song to the watchlist for the given user ID and song ID.
func (w *Watchlist) Add(userID string, songID string) (*pkg.Song, error) {
	song, err := w.store.GetQueries().GetSongById(context.Background(), songID)
	if err != nil {
		return nil, fmt.Errorf("failed to get song: %w", err)
	}

	if song.ID == "" {
		return nil, fmt.Errorf("song was never requested with ID. %s not found", songID)
	}

	watchlist, err := w.store.GetQueries().GetWatchListForUser(context.Background(), userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get watchlist for user: %w", err)
	}

	found := false
	for _, s := range watchlist {
		if s.ID == song.ID {
			found = true
			break
		}
	}

	if found {
		slog.Debug("Song already in watchlist", "songID", song.ID, "userID", userID)
		return &song, nil
	}

	if _, err := w.store.GetQueries().WatchSong(context.Background(), pkg.WatchSongParams{
		UserID: userID,
		SongID: song.ID,
	}); err != nil {
		return nil, fmt.Errorf("failed to add song to watchlist: %w", err)
	}

	return &song, nil
}

// Remove removes a song from the watchlist for the given user ID and song ID.
func (w *Watchlist) Remove(songID string) error {
	return w.store.GetQueries().RemoveSong(context.Background(), songID)
}

// List returns the watchlist for the given user ID.
func (w *Watchlist) List(userID string) (*[]pkg.Song, error) {
	songs, err := w.store.GetQueries().GetWatchListForUser(context.Background(), userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get watchlist: %w", err)
	}

	return &songs, nil
}

// Clear clears the watchlist for the given user ID.
func (w *Watchlist) Clear(userID string) error {
	return w.store.GetQueries().ClearWatchlist(context.Background(), userID)
}
