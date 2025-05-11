package store

import (
	"context"
	"database/sql"

	pkg "github.com/b1tray3r/isitstreamablebot/pkg/db"
)

type Storager interface {
	GetQueries() *pkg.Queries
	GetSongById(id string) (pkg.Song, error)
}

type Store struct {
	db      *sql.DB
	queries *pkg.Queries
}

func NewStore(db *sql.DB) *Store {
	queries := pkg.New(db)
	return &Store{
		db:      db,
		queries: queries,
	}
}

func (s *Store) GetSongById(id string) (pkg.Song, error) {
	song, err := s.queries.GetSongById(context.Background(), id)
	if err != nil {
		return pkg.Song{}, err
	}
	return song, nil
}

func (s *Store) GetQueries() *pkg.Queries {
	return s.queries
}
