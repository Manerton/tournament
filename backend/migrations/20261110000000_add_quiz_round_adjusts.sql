-- +goose Up
ALTER TABLE quiz_teams
ADD COLUMN round1_adjust INT NOT NULL DEFAULT 0,
ADD COLUMN round2_adjust INT NOT NULL DEFAULT 0,
ADD COLUMN round3_adjust INT NOT NULL DEFAULT 0;

-- +goose Down
ALTER TABLE quiz_teams
DROP COLUMN round1_adjust,
DROP COLUMN round2_adjust,
DROP COLUMN round3_adjust;
