-- name: ListMovesForUnit :many
SELECT
  m.id, m.name, m.power, m.accuracy, m.type_id
FROM moves m
JOIN unit_moves um ON um.move_id = m.id
WHERE um.unit_id = $1
ORDER BY m.id;

-- name: CreateMove :one
INSERT INTO moves (name, power, accuracy, type_id)
VALUES ($1, $2, $3, $4)
RETURNING id, name, power, accuracy, type_id;

-- name: CreateUnitMove :one
INSERT INTO unit_moves (unit_id, move_id)
VALUES ($1, $2)
RETURNING unit_id, move_id;