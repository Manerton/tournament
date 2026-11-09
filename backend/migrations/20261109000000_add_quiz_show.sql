-- +goose Up
-- +goose StatementBegin
CREATE TABLE quiz_sessions (
    event_id UUID PRIMARY KEY REFERENCES events(id)
        ON UPDATE CASCADE ON DELETE CASCADE,
    current_round INT NOT NULL DEFAULT 0 CHECK (current_round BETWEEN 0 AND 3),
    current_question INT NOT NULL DEFAULT 0,
    answer_revealed BOOLEAN NOT NULL DEFAULT FALSE,
    round_active BOOLEAN NOT NULL DEFAULT FALSE,
    leaderboard_shown BOOLEAN NOT NULL DEFAULT FALSE
);

CREATE TABLE quiz_questions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    event_id UUID NOT NULL REFERENCES events(id)
        ON UPDATE CASCADE ON DELETE CASCADE,
    round INT NOT NULL CHECK (round IN (1, 2, 3)),
    question_number INT NOT NULL,
    text TEXT NOT NULL,
    answer TEXT NOT NULL DEFAULT '',
    UNIQUE (event_id, round, question_number)
);

CREATE TABLE quiz_teams (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    event_id UUID NOT NULL REFERENCES events(id)
        ON UPDATE CASCADE ON DELETE CASCADE,
    team_id UUID NOT NULL REFERENCES teams(id)
        ON UPDATE CASCADE ON DELETE CASCADE,
    access_code VARCHAR(8) NOT NULL,
    round1_score INT NOT NULL DEFAULT 0 CHECK (round1_score >= 0),
    round2_score INT NOT NULL DEFAULT 0 CHECK (round2_score >= 0),
    round3_score INT NOT NULL DEFAULT 0 CHECK (round3_score >= 0),
    round3_order INT NOT NULL DEFAULT 0 CHECK (round3_order BETWEEN 0 AND 3),
    round3_wrong_count INT NOT NULL DEFAULT 0,
    round3_eliminated BOOLEAN NOT NULL DEFAULT FALSE,
    UNIQUE (event_id, team_id),
    UNIQUE (event_id, access_code)
);

CREATE TABLE quiz_buzzes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    event_id UUID NOT NULL REFERENCES events(id)
        ON UPDATE CASCADE ON DELETE CASCADE,
    round INT NOT NULL,
    question_number INT NOT NULL,
    team_id UUID NOT NULL REFERENCES teams(id)
        ON UPDATE CASCADE ON DELETE CASCADE,
    seq BIGINT NOT NULL DEFAULT 0,
    status INT NOT NULL DEFAULT 1 CHECK (status IN (1, 2, 3, 4)),
    buzzed_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(),
    UNIQUE (event_id, round, question_number, team_id)
);

CREATE TABLE quiz_answers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    event_id UUID NOT NULL REFERENCES events(id)
        ON UPDATE CASCADE ON DELETE CASCADE,
    round INT NOT NULL,
    question_number INT NOT NULL,
    team_id UUID NOT NULL REFERENCES teams(id)
        ON UPDATE CASCADE ON DELETE CASCADE,
    grade INT NOT NULL CHECK (grade IN (1, 2, 3)),
    points INT NOT NULL DEFAULT 0,
    UNIQUE (event_id, round, question_number, team_id)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS quiz_answers CASCADE;
DROP TABLE IF EXISTS quiz_buzzes CASCADE;
DROP TABLE IF EXISTS quiz_teams CASCADE;
DROP TABLE IF EXISTS quiz_questions CASCADE;
DROP TABLE IF EXISTS quiz_sessions CASCADE;
-- +goose StatementEnd
