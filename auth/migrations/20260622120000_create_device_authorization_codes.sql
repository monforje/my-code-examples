-- +goose Up
create table device_authorization_codes (
    id uuid primary key,
    device_code_hash text unique not null,
    user_code text unique not null,
    identity_id uuid references identities(id) on delete set null,
    status text not null default 'pending',
    expires_at timestamptz not null,
    interval int not null default 3,
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now(),
    confirmed_at timestamptz,
    last_polled_at timestamptz
);

create index idx_device_authorization_codes_expires_at on device_authorization_codes(expires_at);

-- +goose Down
drop table if exists device_authorization_codes;
