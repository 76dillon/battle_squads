-- +goose Up
CREATE TABLE squad_units (
    id        BIGSERIAL PRIMARY KEY,
    squad_id  BIGINT NOT NULL REFERENCES squads(id),
    unit_id   BIGINT NOT NULL REFERENCES units(id),
    position  INT    NOT NULL,    -- 0,1,2... order within squad
    UNIQUE (squad_id, position)
);

-- +goose Down
DROP TABLE IF EXISTS squad_units;