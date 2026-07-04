-- +goose Up
create table if not exists user_profiles (
    id uuid primary key,
    identity_id uuid not null,
    email text not null,
    display_name text,
    bio text,
    avatar_url text,
    avatar_object_key text,
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now(),
    deleted_at timestamptz
);

-- +goose Down
drop table if exists user_profiles;