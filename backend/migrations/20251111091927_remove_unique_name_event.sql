-- +goose Up
-- +goose StatementBegin
-- Сначала найдём и удалим уникальный индекс, связанный с полем name
DO $$
DECLARE
    idx_name text;
BEGIN
    SELECT indexname INTO idx_name
    FROM pg_indexes
    WHERE tablename = 'events' AND indexdef ILIKE '%(name)%unique%';

    IF idx_name IS NOT NULL THEN
        EXECUTE format('DROP INDEX IF EXISTS %I;', idx_name);
    END IF;
END$$;

-- +goose StatementEnd
-- +goose Down
-- Добавляем обратно уникальное ограничение на поле name
ALTER TABLE events
ADD CONSTRAINT events_name_key UNIQUE (name);
