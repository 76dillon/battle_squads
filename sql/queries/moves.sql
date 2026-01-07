-- name: ListMovesForUnit :many
SELECT
  m.id, m.name, m.power, m.accuracy, m.type_id
FROM moves m
JOIN unit_moves um ON um.move_id = m.id
WHERE um.unit_id = $1
ORDER BY m.id;