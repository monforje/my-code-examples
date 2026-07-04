-- +goose Up
alter table email_change_requests
    add COLUMN token_hash text,
    add COLUMN consumed_at timestamptz;

-- +goose Down
alter table email_change_requests
    drop COLUMN if exists token_hash,
    drop COLUMN if exists consumed_at;
