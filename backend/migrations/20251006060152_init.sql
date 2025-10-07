-- +goose Up
-- +goose StatementBegin

CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- Таблица пользователей
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    login VARCHAR(255) NOT NULL UNIQUE,
    password VARCHAR(255) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now()
);

-- Таблица пользовательских сессий
CREATE TABLE user_sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id)
        ON UPDATE CASCADE ON DELETE CASCADE,
    session_token VARCHAR(255) NOT NULL UNIQUE,
    csrf_token VARCHAR(255) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(),
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    
    -- Индексы
    INDEX idx_user_sessions_user_id (user_id),
    INDEX idx_user_sessions_session_token (session_token),
    INDEX idx_user_sessions_expires_at (expires_at)
);

-- Остальные существующие таблицы
CREATE TABLE countries (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL UNIQUE
);

CREATE TABLE regions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL UNIQUE,
    country_id UUID NOT NULL REFERENCES countries(id)
        ON UPDATE CASCADE ON DELETE CASCADE
);

CREATE TABLE districts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL UNIQUE,
    region_id UUID NOT NULL REFERENCES regions(id)
        ON UPDATE CASCADE ON DELETE CASCADE
);

CREATE TABLE schools (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL UNIQUE,
    district_id UUID NOT NULL REFERENCES districts(id)
        ON UPDATE CASCADE ON DELETE CASCADE
);

CREATE TABLE events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL UNIQUE,
    start_date DATE NOT NULL,
    end_date DATE NOT NULL,
    previous_event_id UUID REFERENCES events(id)
        ON UPDATE CASCADE ON DELETE SET NULL,
    type_draws INT NOT NULL,
    type INT NOT NULL,
    status INT NOT NULL DEFAULT 1,
    stage INT
);

CREATE TABLE participants (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    fio VARCHAR(255) NOT NULL,
    class SMALLINT NOT NULL,
    gender SMALLINT NOT NULL,
    school_id UUID NOT NULL REFERENCES schools(id)
        ON UPDATE CASCADE ON DELETE CASCADE
);

CREATE TABLE teams (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL UNIQUE,
    responsible_name VARCHAR(255) NOT NULL,
    responsible_phone VARCHAR(20) NOT NULL,
    event_id UUID NOT NULL REFERENCES events(id)
        ON UPDATE CASCADE ON DELETE CASCADE,
    school_id UUID NOT NULL REFERENCES schools(id)
        ON UPDATE CASCADE ON DELETE CASCADE
);

CREATE TABLE team_participants (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    team_id UUID REFERENCES teams(id)
        ON UPDATE CASCADE ON DELETE CASCADE,
    participant_id UUID REFERENCES participants(id)
        ON UPDATE CASCADE ON DELETE CASCADE
);

CREATE TABLE games (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    team_a UUID NOT NULL,
    team_b UUID NOT NULL,
    event_id UUID NOT NULL REFERENCES events(id)
        ON UPDATE CASCADE ON DELETE CASCADE,
    finished INT NOT NULL DEFAULT 0 CHECK (finished IN (0, 1, 2, 3))
);

CREATE TABLE game_participants (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    game_id UUID NOT NULL REFERENCES games(id)
        ON UPDATE CASCADE ON DELETE CASCADE,
    participant_id UUID NOT NULL REFERENCES participants(id)
        ON UPDATE CASCADE ON DELETE CASCADE,
    score INT NOT NULL DEFAULT 0 CHECK (score >= 0)
);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS game_participants CASCADE;
DROP TABLE IF EXISTS games CASCADE;
DROP TABLE IF EXISTS team_participants CASCADE;
DROP TABLE IF EXISTS teams CASCADE;
DROP TABLE IF EXISTS participants CASCADE;
DROP TABLE IF EXISTS events CASCADE;
DROP TABLE IF EXISTS schools CASCADE;
DROP TABLE IF EXISTS districts CASCADE;
DROP TABLE IF EXISTS regions CASCADE;
DROP TABLE IF EXISTS countries CASCADE;
DROP TABLE IF EXISTS user_sessions CASCADE;
DROP TABLE IF EXISTS users CASCADE;

-- +goose StatementEnd