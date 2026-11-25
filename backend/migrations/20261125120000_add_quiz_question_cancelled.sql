-- +goose Up
-- +goose StatementBegin
-- cancelled: the question is annulled — all its results are cancelled for every team.
ALTER TABLE quiz_questions
    ADD COLUMN cancelled BOOLEAN NOT NULL DEFAULT FALSE;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE quiz_questions
    DROP COLUMN cancelled;
-- +goose StatementEnd
