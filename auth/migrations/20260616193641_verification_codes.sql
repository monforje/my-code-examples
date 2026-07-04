-- +goose Up
create table verification_codes (
    id uuid primary key,
    identity_id uuid references identities(id) on delete set null,
    email text,
    purpose text not null,
    code_hash text not null,
    attempts_count int not null default 0,
    max_attempts int not null default 5,
    expires_at timestamptz not null,
    consumed_at timestamptz,
    created_at timestamptz not null default now()
);

create index idx_verification_codes_email on verification_codes(email);
create index idx_verification_codes_identity_id on verification_codes(identity_id);
create index idx_verification_codes_purpose on verification_codes(purpose);

-- +goose Down
drop table if exists verification_codes;
