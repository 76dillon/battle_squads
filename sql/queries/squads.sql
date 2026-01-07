-- name: CreateSquad :one
INSERT INTO squads (
    player_id,
    name
) VALUES (
    $1, $2
)
RETURNING id, player_id, name, created_at;

-- name: GetSquadsForPlayer :many
SELECT id, player_id, name, created_at
FROM squads
WHERE player_id = $1
ORDER BY created_at DESC;