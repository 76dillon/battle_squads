-- +goose Up
CREATE TABLE match_turns (
    id                   BIGSERIAL PRIMARY KEY,
    match_id             BIGINT NOT NULL REFERENCES matches(id),
    turn_number          INT    NOT NULL,
    acting_player_id     BIGINT NOT NULL REFERENCES players(id),
    acting_match_unit_id BIGINT NOT NULL REFERENCES match_units(id),
    move_id              BIGINT NOT NULL REFERENCES moves(id),
    target_match_unit_id BIGINT NOT NULL REFERENCES match_units(id),
    damage_done          INT    NOT NULL,
    target_hp_after      INT    NOT NULL,
    did_ko_target        BOOLEAN NOT NULL DEFAULT FALSE,
    created_at           TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- +goose Down
DROP TABLE IF EXISTS match_turns;