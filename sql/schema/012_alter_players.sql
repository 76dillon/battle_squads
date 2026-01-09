-- +goose Up
ALTER TABLE players
ADD COLUMN is_admin BOOLEAN NOT NULL DEFAULT FALSE;

-- +goose Down
ALTER TABLE players
DROP COLUMN IF EXISTS is_admin;