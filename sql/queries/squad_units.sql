-- name: CreateSquadUnit :one
INSERT INTO squad_units (
    squad_id,
    unit_id,
    position
) VALUES (
    $1, $2, $3
)
RETURNING id, squad_id, unit_id, position;

-- name: GetSquadUnits :many
SELECT id, squad_id, unit_id, position
FROM squad_units
WHERE squad_id = $1
ORDER BY position;