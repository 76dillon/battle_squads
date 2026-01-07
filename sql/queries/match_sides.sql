-- name: CreateMatchSide :one
INSERT INTO match_sides (
    match_id,
    player_id,
    squad_id,
    active_index
) VALUES (
    $1, $2, $3, 0
)
RETURNING id, match_id, player_id, squad_id, active_index;

-- name: GetMatchSidesByMatchID :many
SELECT
    id, match_id, player_id, squad_id, active_index
FROM match_sides
WHERE match_id = $1;

-- name: UpdateMatchSideActiveIndex :one
UPDATE match_sides
SET active_index = $2
WHERE id = $1
RETURNING id, match_id, player_id, squad_id, active_index;