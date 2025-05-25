
-- +migrate Up
ALTER TABLE songs ADD COLUMN message_id TEXT NOT NULL DEFAULT '';

-- +migrate Down
ALTER TABLE songs DROP COLUMN message_id;
