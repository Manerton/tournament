-- +goose Up
ALTER TABLE games
    RENAME COLUMN team_a TO team_a_id;

ALTER TABLE games
    RENAME COLUMN team_b TO team_b_id;

-- +goose Down
ALTER TABLE games
    RENAME COLUMN team_a_id TO team_a;

ALTER TABLE games
    RENAME COLUMN team_b_id TO team_b;
