-- +goose Up
ALTER TABLE leaderboards
ALTER COLUMN enemies_ids DROP NOT NULL,
ALTER COLUMN scores DROP NOT NULL;

-- +goose Down
ALTER TABLE leaderboards
ALTER COLUMN enemies_ids SET NOT NULL,
ALTER COLUMN scores SET NOT NULL;
