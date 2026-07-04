-- +goose Up
create table sessions (
    id uuid primary key,
    identity_id uuid not null references identities(id) on delete cascade,
    refresh_token_hash text not null,
    user_agent text not null default '',
    ip_address inet,
    expires_at timestamptz not null,
    revoked_at timestamptz,
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now()
);

create index idx_sessions_identity_id on sessions(identity_id);
create index idx_sessions_expires_at on sessions(expires_at);

-- +goose Down
drop table if exists sessions;
