
-- +migrate Up
CREATE TABLE songs (
    id TEXT NOT NULL PRIMARY KEY,
    title TEXT NOT NULL,
    artists TEXT NOT NULL,
    duration INT NOT NULL,
    request_time DATETIME NOT NULL DEFAULT (datetime(CURRENT_TIMESTAMP, 'localtime')),
    is_streamable INTEGER NOT NULL CHECK (is_streamable IN (0, 1)) DEFAULT 0
);

-- +migrate Down
DROP TABLE songs;
