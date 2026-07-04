-- +goose Up
create table if not exists processed_events (
    event_id text primary key,
    event_type text not null,
    aggregate_id uuid not null,
    processed_at timestamptz not null default now()
);

-- +goose Down
drop table if exists processed_events;
