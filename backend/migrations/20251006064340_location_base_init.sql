-- +goose Up
-- +goose StatementBegin

-- Добавляем страну
INSERT INTO countries (id, name)
VALUES (gen_random_uuid(), 'Россия')
ON CONFLICT (name) DO NOTHING;

-- Добавляем регион, связанный с Россией
INSERT INTO regions (id, name, country_id)
SELECT gen_random_uuid(), 'Астраханская область', c.id
FROM countries c
WHERE c.name = 'Россия'
ON CONFLICT (name) DO NOTHING;

-- +goose StatementEnd


-- +goose Down
-- +goose StatementBegin

DELETE FROM regions WHERE name = 'Астраханская область';
DELETE FROM countries WHERE name = 'Россия';

-- +goose StatementEnd
