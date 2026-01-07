-- name: CreateMatch :one
INSERT INTO matches (
    state,
    player1_id,
    player2_id,
    current_turn_number,
    current_actor_player_id
) VALUES (
    'PENDING',
    $1,                -- player1_id
    $2,                -- player2_id
    0,
    NULL
)
RETURNING
    id,
    state,
    created_at,
    started_at,
    completed_at,
    player1_id,
    player2_id,
    winner_player_id,
    current_turn_number,
    current_actor_player_id;

-- name: GetMatchByID :one
SELECT
    id,
    state,
    created_at,
    started_at,
    completed_at,
    player1_id,
    player2_id,
    winner_player_id,
    current_turn_number,
    current_actor_player_id
FROM matches
WHERE id = $1;

-- name: ListMatchesForPlayer :many
SELECT
    id,
    state,
    created_at,
    started_at,
    completed_at,
    player1_id,
    player2_id,
    winner_player_id,
    current_turn_number,
    current_actor_player_id
FROM matches
WHERE player1_id = $1 OR player2_id = $1
ORDER BY created_at DESC;

-- name: StartMatch :one
UPDATE matches
SET
  state = 'IN_PROGRESS',
  started_at = now(),
  current_turn_number = 1,
  current_actor_player_id = $2  -- initial actor
WHERE id = $1
RETURNING
  id, state, created_at, started_at, completed_at,
  player1_id, player2_id, winner_player_id,
  current_turn_number, current_actor_player_id;

-- name: CompleteMatch :one
UPDATE matches
SET
  state = 'COMPLETED',
  completed_at = now(),
  winner_player_id = $2,
  current_actor_player_id = NULL
WHERE id = $1
RETURNING
  id,
  state,
  created_at,
  started_at,
  completed_at,
  player1_id,
  player2_id,
  winner_player_id,
  current_turn_number,
  current_actor_player_id;

-- name: UpdateMatchTurnAndActor :one
UPDATE matches
SET
  current_turn_number = $2,
  current_actor_player_id = $3
WHERE id = $1
RETURNING 
  id,
  state,
  created_at,
  started_at,
  completed_at,
  player1_id,
  player2_id,
  winner_player_id,
  current_turn_number,
  current_actor_player_id;