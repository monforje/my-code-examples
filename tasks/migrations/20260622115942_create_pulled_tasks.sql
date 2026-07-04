-- +goose Up

CREATE TABLE IF NOT EXISTS pulled_tasks (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    identity_id UUID NOT NULL,
    task_id     UUID NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
    repo        TEXT NOT NULL,
    clone_url   TEXT NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_pulled_tasks_identity_id ON pulled_tasks(identity_id);
CREATE UNIQUE INDEX idx_pulled_tasks_identity_task ON pulled_tasks(identity_id, task_id);

-- +goose Down

DROP INDEX IF EXISTS idx_pulled_tasks_identity_task;
DROP INDEX IF EXISTS idx_pulled_tasks_identity_id;
DROP TABLE IF EXISTS pulled_tasks;
