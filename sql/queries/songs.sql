-- name: InsertSong :one
INSERT INTO songs (id, title, artists, duration, request_time, is_streamable) VALUES (?, ?, ?, ?, ?, ?) RETURNING *;

-- name: GetSongById :one
SELECT * FROM songs WHERE id = ?;

-- name: RemoveSong :exec
DELETE FROM songs WHERE id = ?;