-- +goose Up
-- +goose StatementBegin
-- round3_extra_lives: host-granted extra mistake allowance used to restore
-- an eliminated team in round 3 (added on top of the order's allowance).
ALTER TABLE quiz_teams
    ADD COLUMN round3_extra_lives INT NOT NULL DEFAULT 0;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE quiz_teams
    DROP COLUMN round3_extra_lives;
-- +goose StatementEnd
