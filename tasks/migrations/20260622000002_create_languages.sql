-- +goose Up

CREATE TABLE languages (
  id uuid primary key,
  name text not null unique
);

-- +goose Down

DROP TABLE IF EXISTS languages;
