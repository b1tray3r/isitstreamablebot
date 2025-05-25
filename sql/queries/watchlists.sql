-- name: GetWatchListForUser :many
SELECT songs.* FROM watchlists JOIN songs ON watchlists.song_id = songs.id WHERE watchlists.user_id = ?;

-- name: WatchSong :one
INSERT INTO watchlists (user_id, song_id) VALUES (?, ?) RETURNING *;

-- name: RemoveSongFromWatchlist :exec
DELETE FROM watchlists WHERE song_id = ? AND user_id = ?;

-- name: ClearWatchlist :exec
DELETE FROM watchlists WHERE user_id = ?;

-- name: GetUniqueSongsInWatchlist :many
SELECT DISTINCT songs.* FROM watchlists JOIN songs ON watchlists.song_id = songs.id;

-- name: GetUsersForSongID :many
SELECT user_id FROM watchlists WHERE song_id = ?;