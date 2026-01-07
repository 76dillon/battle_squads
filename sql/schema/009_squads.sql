-- +goose Up
CREATE TABLE squads (
    id         BIGSERIAL PRIMARY KEY,
    player_id  BIGINT      NOT NULL REFERENCES players(id),
    name       TEXT        NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- +goose Down
DROP TABLE IF EXISTS squads;