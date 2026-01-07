-- +goose Up
CREATE TABLE matches (
    id                      BIGSERIAL PRIMARY KEY,
    state                   TEXT        NOT NULL,  -- 'PENDING', 'IN_PROGRESS', 'COMPLETED'
    created_at              TIMESTAMPTZ NOT NULL DEFAULT now(),
    started_at              TIMESTAMPTZ,
    completed_at            TIMESTAMPTZ,

    player1_id              BIGINT      NOT NULL REFERENCES players(id),
    player2_id              BIGINT      NOT NULL REFERENCES players(id),

    winner_player_id        BIGINT      REFERENCES players(id),

    current_turn_number     INT         NOT NULL DEFAULT 0,
    current_actor_player_id BIGINT      REFERENCES players(id)
);

-- +goose Down
DROP TABLE IF EXISTS matches;