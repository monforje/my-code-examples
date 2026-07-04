-- +goose Up
create table email_change_requests (
    id uuid primary key,
    identity_id uuid not null references identities(id) on delete cascade,
    new_email text not null,
    status text not null default 'pending',
    expires_at timestamptz not null,
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now()
);

create index idx_email_change_requests_identity_id on email_change_requests(identity_id);

-- +goose Down
drop table if exists email_change_requests;
