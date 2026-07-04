-- +goose Up
create table password_reset_tokens (
    id uuid primary key,
    identity_id uuid not null references identities(id) on delete cascade,
    token_hash text not null,
    expires_at timestamptz not null,
    consumed_at timestamptz,
    created_at timestamptz not null default now()
);

create index idx_password_reset_tokens_identity_id on password_reset_tokens(identity_id);

-- +goose Down
drop table if exists password_reset_tokens;
