-- +goose Up
-- +goose StatementBegin
-- Move the last 3 questions of each main round of "Королевская битва" into the
-- reserve pool (round 4). Reserve numbering after the move (ascending order in
-- which tie-break picks questions):
--   1-3  — the questions that were already in the reserve
--   4-6  — former round 3 questions 8-10
--   7-9  — former round 2 questions 13-15
--   10-12 — former round 1 questions 18-20
-- Main rounds keep contiguous numbering (1-17, 1-12, 1-7). The event was
-- created via the UI, so it is looked up by name; without it the migration is
-- a no-op. Re-running updates nothing (rounds 1-3 no longer hold those numbers).
UPDATE quiz_questions
SET round = 4,
    question_number = CASE
        WHEN round = 3 THEN question_number - 4  -- 8,9,10   -> 4,5,6
        WHEN round = 2 THEN question_number - 6  -- 13,14,15 -> 7,8,9
        WHEN round = 1 THEN question_number - 8  -- 18,19,20 -> 10,11,12
    END
WHERE event_id IN (
        SELECT id FROM events
        WHERE name = 'Третий региональный математический турнир школьных команд «Королевская битва»'
    )
    AND (
        (round = 3 AND question_number BETWEEN 8 AND 10)
        OR (round = 2 AND question_number BETWEEN 13 AND 15)
        OR (round = 1 AND question_number BETWEEN 18 AND 20)
    );
-- +goose StatementEnd

-- +goose Down
