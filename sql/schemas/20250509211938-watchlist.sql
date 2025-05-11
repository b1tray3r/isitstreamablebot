-- +migrate Up
CREATE TABLE watchlists (
    user_id TEXT NOT NULL,
    song_id TEXT NOT NULL,
    CONSTRAINT fk_song FOREIGN KEY (song_id) REFERENCES songs(id)
);

-- +migrate Down
DROP TABLE watchlists;
