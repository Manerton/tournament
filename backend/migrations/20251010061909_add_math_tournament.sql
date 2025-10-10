-- +goose Up
INSERT INTO events (
    name,
    start_date,
    end_date,
    previous_event_id,
    type_draws,
    type,
    status,
    stage
)
VALUES (
    'Математический турнир 2025',
    '2025-11-12 10:00:00+03',
    '2025-11-12 15:00:00+03',
    NULL,
    3,  -- Swiss
    1,  -- тип события (например, основной)
    1,  -- активен
    0
);

-- +goose Down
DELETE FROM events
WHERE name = 'Математический турнир 2025';
