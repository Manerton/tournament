-- +goose Up
-- +goose StatementBegin
-- A team may press in several buzz windows within one question,
-- so the per-question uniqueness has to go; ordering is kept via seq/window_idx.
ALTER TABLE quiz_buzzes DROP CONSTRAINT IF EXISTS quiz_buzzes_event_id_round_question_number_team_id_key;
ALTER TABLE quiz_buzzes DROP CONSTRAINT IF EXISTS quiz_buzzes_status_check;
ALTER TABLE quiz_buzzes ADD CONSTRAINT quiz_buzzes_status_check CHECK (status IN (1, 2, 3, 4, 5));
ALTER TABLE quiz_buzzes ADD COLUMN offset_ms INT NOT NULL DEFAULT 0;
ALTER TABLE quiz_buzzes ADD COLUMN window_idx INT NOT NULL DEFAULT 0;

-- Physical ESP32 buttons: preconfigured mapping button id -> quiz team, nothing else.
CREATE TABLE quiz_buttons (
    id VARCHAR(64) PRIMARY KEY,
    quiz_team_id UUID NOT NULL REFERENCES quiz_teams(id)
        ON UPDATE CASCADE ON DELETE CASCADE
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS quiz_buttons CASCADE;
ALTER TABLE quiz_buzzes DROP COLUMN IF EXISTS window_idx;
ALTER TABLE quiz_buzzes DROP COLUMN IF EXISTS offset_ms;
ALTER TABLE quiz_buzzes DROP CONSTRAINT IF EXISTS quiz_buzzes_status_check;
ALTER TABLE quiz_buzzes ADD CONSTRAINT quiz_buzzes_status_check CHECK (status IN (1, 2, 3, 4));
ALTER TABLE quiz_buzzes ADD CONSTRAINT quiz_buzzes_event_id_round_question_number_team_id_key
    UNIQUE (event_id, round, question_number, team_id);
-- +goose StatementEnd
