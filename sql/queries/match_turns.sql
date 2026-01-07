-- name: CreateMatchTurn :one
INSERT INTO match_turns (
    match_id,
    turn_number,
    acting_player_id,
    acting_match_unit_id,
    move_id,
    target_match_unit_id,
    damage_done,
    target_hp_after,
    did_ko_target
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9
)
RETURNING
    id,
    match_id,
    turn_number,
    acting_player_id,
    acting_match_unit_id,
    move_id,
    target_match_unit_id,
    damage_done,
    target_hp_after,
    did_ko_target,
    created_at;

-- name: ListMatchTurns :many
SELECT
  id, match_id, turn_number, acting_player_id,
  acting_match_unit_id, move_id, target_match_unit_id,
  damage_done, target_hp_after, did_ko_target, created_at
FROM match_turns
WHERE match_id = $1
ORDER BY turn_number;