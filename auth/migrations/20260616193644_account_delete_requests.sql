-- +goose Up
create table account_delete_requests (
    id uuid primary key,
    identity_id uuid not null references identities(id) on delete cascade,
    status text not null default 'pending',
    expires_at timestamptz not null,
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now()
);

create index idx_account_delete_requests_identity_id on account_delete_requests(identity_id);

-- +goose Down
drop table if exists account_delete_requests;
