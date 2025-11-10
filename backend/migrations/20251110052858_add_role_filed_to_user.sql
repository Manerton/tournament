-- +goose Up
-- +goose StatementBegin

-- Добавляем столбец role (по умолчанию 0)
ALTER TABLE users
    ADD COLUMN role INTEGER NOT NULL DEFAULT 0;

-- Проставляем роли существующим пользователям
UPDATE users SET role = 1 WHERE login = 'admin'; -- ADMIN
UPDATE users SET role = 2 WHERE login = 'jury';  -- JUDGE

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Удаляем столбец role при откате
ALTER TABLE users
    DROP COLUMN role;

-- +goose StatementEnd