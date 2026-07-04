-- +goose Up
create table if not exists git_users (
    id uuid primary key,
    profile_id uuid references user_profiles(id) on delete cascade,
    git_token text not null,
    git_url text not null,
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now()
);

-- +goose Down
drop table if exists git_users;