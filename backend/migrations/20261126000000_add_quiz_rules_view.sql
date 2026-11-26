-- +goose Up
-- +goose StatementBegin
-- rules_view: which rules slide the projector shows ("none", "overview", "1", "2", "3").
-- Mutually exclusive with leaderboard_shown; auto-reset when a question starts.
ALTER TABLE quiz_sessions
    ADD COLUMN rules_view VARCHAR(16) NOT NULL DEFAULT 'none';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE quiz_sessions DROP COLUMN rules_view;
-- +goose StatementEnd
