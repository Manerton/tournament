-- +goose Up
ALTER TABLE events
ADD COLUMN max_stage INT NOT NULL DEFAULT 0;

-- +goose Down
ALTER TABLE events
DROP COLUMN IF EXISTS max_stage;
