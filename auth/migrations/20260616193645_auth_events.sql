-- +goose Up
create table auth_events (
    id uuid primary key,
    identity_id uuid references identities(id) on delete set null,
    event_type text not null,
    ip_address inet,
    user_agent text not null default '',
    metadata jsonb,
    created_at timestamptz not null default now()
);

create index idx_auth_events_identity_id on auth_events(identity_id);
create index idx_auth_events_event_type on auth_events(event_type);

-- +goose Down
drop table if exists auth_events;
