-- +goose Up
ALTER TABLE processed_events ALTER COLUMN event_id TYPE text;

-- +goose Down
ALTER TABLE processed_events ALTER COLUMN event_id TYPE uuid USING event_id::uuid;
