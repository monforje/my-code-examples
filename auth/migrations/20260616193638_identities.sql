-- +goose Up
create table identities (
    id uuid primary key,
    email text not null unique,
    email_verified boolean not null default false,
    status text not null default 'pending_verification',
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now(),
    deleted_at timestamptz
);

-- +goose Down
drop table if exists identities;
