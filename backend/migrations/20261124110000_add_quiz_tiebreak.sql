-- +goose Up
-- +goose StatementBegin
-- Tie-break (переигровка) state:
-- - quiz_sessions: whether a tie-break is running and which reserve question is shown
-- - quiz_questions.used: a shown reserve question is never shown again (shared hall/projector)
-- - quiz_teams.tiebreak_score: secondary leaderboard sort key, never touches round scores
-- - quiz_tiebreak_teams: the tied group selected by the host; won = secured a place, out of the group
ALTER TABLE quiz_sessions
    ADD COLUMN tiebreak_active BOOLEAN NOT NULL DEFAULT FALSE,
    ADD COLUMN tiebreak_question INT NOT NULL DEFAULT 0;

ALTER TABLE quiz_questions
    ADD COLUMN used BOOLEAN NOT NULL DEFAULT FALSE;

ALTER TABLE quiz_teams
    ADD COLUMN tiebreak_score INT NOT NULL DEFAULT 0;

CREATE TABLE quiz_tiebreak_teams (
    event_id UUID NOT NULL REFERENCES events(id)
        ON UPDATE CASCADE ON DELETE CASCADE,
    team_id UUID NOT NULL REFERENCES teams(id)
        ON UPDATE CASCADE ON DELETE CASCADE,
    won BOOLEAN NOT NULL DEFAULT FALSE,
    PRIMARY KEY (event_id, team_id)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS quiz_tiebreak_teams;
ALTER TABLE quiz_teams DROP COLUMN tiebreak_score;
ALTER TABLE quiz_questions DROP COLUMN used;
ALTER TABLE quiz_sessions
    DROP COLUMN tiebreak_question,
    DROP COLUMN tiebreak_active;
-- +goose StatementEnd
