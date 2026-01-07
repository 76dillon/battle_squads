-- +goose Up
CREATE TABLE unit_moves (
    unit_id BIGINT NOT NULL REFERENCES units(id),
    move_id BIGINT NOT NULL REFERENCES moves(id),
    PRIMARY KEY (unit_id, move_id)
);

-- +goose Down
DROP TABLE IF EXISTS unit_moves;