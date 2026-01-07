-- +goose Up
CREATE TABLE match_sides (
    id           BIGSERIAL PRIMARY KEY,
    match_id     BIGINT NOT NULL REFERENCES matches(id),
    player_id    BIGINT NOT NULL REFERENCES players(id),
    squad_id     BIGINT NOT NULL, -- will reference squads later
    active_index INT    NOT NULL DEFAULT 0,
    UNIQUE (match_id, player_id)
);

-- +goose Down
DROP TABLE IF EXISTS match_sides;