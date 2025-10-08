-- +goose Up
-- +goose StatementBegin

-- Функция для хеширования пароля (используем встроенные возможности PostgreSQL с pgcrypto)
CREATE OR REPLACE FUNCTION hash_password(password TEXT)
RETURNS TEXT AS $$
BEGIN
    -- Используем crypt из pgcrypto для bcrypt-хеширования
    RETURN crypt(password, gen_salt('bf', 14)); -- bf = blowfish (bcrypt), 14 - стоимость
END;
$$ LANGUAGE plpgsql;

-- Вставляем тестовых пользователей
INSERT INTO users (login, password) VALUES
    ('jury', hash_password('jury')),
    ('jury2', hash_password('jury2')),
    ('jury3', hash_password('jury3')),
    ('admin', hash_password('admin')),
    ('moderator', hash_password('moderator'));

-- Удаляем вспомогательную функцию
DROP FUNCTION hash_password(TEXT);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Удаляем тестовых пользователей
DELETE FROM users WHERE login IN ('jury', 'jury2', 'jury3', 'admin', 'moderator');

-- +goose StatementEnd