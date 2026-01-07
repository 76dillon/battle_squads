-- +goose Up
CREATE TABLE units (
    id          BIGSERIAL PRIMARY KEY,
    name        TEXT        NOT NULL,
    type_id     BIGINT      NOT NULL REFERENCES unit_types(id),
    base_hp     INT         NOT NULL,
    base_attack INT         NOT NULL,
    base_speed  INT         NOT NULL
);

-- +goose Down
DROP TABLE IF EXISTS units;