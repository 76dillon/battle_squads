-- +goose Up
CREATE TABLE moves (
    id        BIGSERIAL PRIMARY KEY,
    name      TEXT        NOT NULL,
    power     INT         NOT NULL,
    accuracy  INT         NOT NULL,  -- 0–100, simple for now
    type_id   BIGINT      NOT NULL REFERENCES unit_types(id)
);

-- +goose Down
DROP TABLE IF EXISTS moves;