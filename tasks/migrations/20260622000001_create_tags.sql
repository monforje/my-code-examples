-- +goose Up

CREATE TABLE tags (
  id uuid primary key,
  name text not null unique
);

-- +goose Down

DROP TABLE IF EXISTS tags;
