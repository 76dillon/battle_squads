-- name: CreatePlayer :one
INSERT INTO players (username, password_hash)
VALUES ($1, $2)
RETURNING id, username, password_hash, created_at;

-- name: GetPlayerByUsername :one
SELECT id, username, password_hash, created_at
FROM players
WHERE username = $1;