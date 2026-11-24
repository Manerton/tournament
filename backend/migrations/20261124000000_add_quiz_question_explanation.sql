-- +goose Up
-- +goose StatementBegin
-- explanation: пояснение к ответу (показывается вместе с ответом).
-- criteria: критерии оценивания для жюри (тур 2: 4 / 2 / 0 баллов).
ALTER TABLE quiz_questions
    ADD COLUMN explanation TEXT NOT NULL DEFAULT '',
    ADD COLUMN criteria TEXT NOT NULL DEFAULT '';

-- Round 4 is the reserve question pool used by the tie-break mechanism.
ALTER TABLE quiz_questions DROP CONSTRAINT quiz_questions_round_check;
ALTER TABLE quiz_questions ADD CONSTRAINT quiz_questions_round_check CHECK (round IN (1, 2, 3, 4));
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DELETE FROM quiz_questions WHERE round = 4;
ALTER TABLE quiz_questions DROP CONSTRAINT quiz_questions_round_check;
ALTER TABLE quiz_questions ADD CONSTRAINT quiz_questions_round_check CHECK (round IN (1, 2, 3));
ALTER TABLE quiz_questions
    DROP COLUMN criteria,
    DROP COLUMN explanation;
-- +goose StatementEnd
