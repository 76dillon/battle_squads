-- name: ListUnits :many
SELECT
  id, name, type_id, base_hp, base_attack, base_speed
FROM units
ORDER BY id;

-- name: GetUnitByID :one
SELECT
  id, name, type_id, base_hp, base_attack, base_speed
FROM units
WHERE id = $1;

-- name: CreateUnit :one
INSERT INTO units (name, type_id, base_hp, base_attack, base_speed)
VALUES ($1, $2, $3, $4, $5)
RETURNING id, name, type_id, base_hp, base_attack, base_speed;