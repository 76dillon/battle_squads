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