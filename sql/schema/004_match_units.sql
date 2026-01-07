-- +goose Up
CREATE TABLE match_units (
    id            BIGSERIAL PRIMARY KEY,
    match_side_id BIGINT NOT NULL REFERENCES match_sides(id),
    unit_id       BIGINT NOT NULL, -- will reference units later
    position      INT    NOT NULL, -- 0,1,2... order in squad
    current_hp    INT    NOT NULL,
    UNIQUE (match_side_id, position)
);

-- +goose Down
DROP TABLE IF EXISTS match_units;