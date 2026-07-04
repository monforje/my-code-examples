-- +goose Up
create table credentials (
    identity_id uuid primary key references identities(id) on delete cascade,
    password_hash text not null,
    password_changed_at timestamptz not null default now(),
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now()
);

-- +goose Down
drop table if exists credentials;
