-- +goose Up
create table password_change_tokens (
    id uuid primary key,
    identity_id uuid not null references identities(id) on delete cascade,
    token_hash text not null,
    expires_at timestamptz not null,
    consumed_at timestamptz,
    created_at timestamptz not null default now()
);

create index idx_password_change_tokens_identity_id on password_change_tokens(identity_id);
create unique index idx_password_change_tokens_token_hash on password_change_tokens(token_hash);

-- +goose Down
drop table if exists password_change_tokens;
