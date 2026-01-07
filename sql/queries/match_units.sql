-- name: CreateMatchUnit :one
INSERT INTO match_units (
    match_side_id,
    unit_id,
    position,
    current_hp
) VALUES (
    $1, $2, $3, $4
)
RETURNING id, match_side_id, unit_id, position, current_hp;

-- name: GetMatchUnitsBySideID :many
SELECT
    id, match_side_id, unit_id, position, current_hp
FROM match_units
WHERE match_side_id = $1
ORDER BY position;

-- name: GetActiveMatchUnitForSide :one
SELECT mu.id, mu.match_side_id, mu.unit_id, mu.position, mu.current_hp
FROM match_units mu
JOIN match_sides ms ON ms.id = mu.match_side_id
WHERE mu.match_side_id = $1
  AND mu.position = ms.active_index;

-- name: UpdateMatchUnitHP :one
UPDATE match_units
SET current_hp = $2
WHERE id = $1
RETURNING id, match_side_id, unit_id, position, current_hp;