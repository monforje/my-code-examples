-- +goose Up

CREATE TABLE task_tags (
  task_id uuid not null REFERENCES tasks(id) ON DELETE CASCADE,
  tag_id  uuid not null REFERENCES tags(id) ON DELETE CASCADE,
  PRIMARY KEY (task_id, tag_id)
);

CREATE TABLE task_languages (
  task_id    uuid not null REFERENCES tasks(id) ON DELETE CASCADE,
  language_id uuid not NULL REFERENCES languages(id) ON DELETE CASCADE,
  PRIMARY KEY (task_id, language_id)
);

-- +goose Down

DROP TABLE IF EXISTS task_languages;
DROP TABLE IF EXISTS task_tags;
