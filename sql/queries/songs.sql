-- name: InsertSong :one
INSERT INTO songs (id, title, artists, duration, request_time, is_streamable, message_id) VALUES (?, ?, ?, ?, ?, ?, ?) RETURNING *;

-- name: GetSongById :one
SELECT * FROM songs WHERE id = ?;

-- name: RemoveSong :exec
DELETE FROM songs WHERE id = ?;

-- name: SetSongStreamable :exec
UPDATE songs SET is_streamable = 1 WHERE id = ?;