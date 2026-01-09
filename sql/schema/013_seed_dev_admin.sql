-- +goose Up
INSERT INTO players (id, username, password_hash, is_admin)
VALUES (1, 'devadmin', 'devpassword', TRUE)
ON CONFLICT (id) DO NOTHING;

-- +goose Down
DELETE FROM players WHERE id = 1 AND username = 'devadmin';