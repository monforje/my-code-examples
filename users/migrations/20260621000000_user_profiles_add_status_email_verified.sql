-- +goose Up
alter table user_profiles add column if not exists status text not null default 'active';
alter table user_profiles add column if not exists email_verified boolean not null default false;

-- +goose Down
alter table user_profiles drop column if exists email_verified;
alter table user_profiles drop column if exists status;
