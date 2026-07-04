-- +goose Up

CREATE TABLE IF NOT EXISTS ci_reports (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    username    TEXT NOT NULL,
    task_name   TEXT NOT NULL REFERENCES tasks(task_name) ON DELETE CASCADE,
    uid         TEXT NOT NULL,
    commit      TEXT NOT NULL,
    status      TEXT NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL,
    payload     JSONB NOT NULL
);

CREATE INDEX idx_ci_reports_username_task
    ON ci_reports (username, task_name, created_at DESC);

-- +goose Down

DROP INDEX IF EXISTS idx_ci_reports_username_task;
DROP TABLE IF EXISTS ci_reports;
