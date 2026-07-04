-- +goose Up
-- Переход с identity_id на username: identity_id недоступен в CI-флоу
-- (task_runner знает только gitea-username), см. vmpool.go:179 uid = username + "/" + goldenRepoName.

ALTER TABLE ci_reports DROP COLUMN IF EXISTS identity_id;
ALTER TABLE ci_reports ADD COLUMN IF NOT EXISTS username TEXT NOT NULL DEFAULT '';

DROP INDEX IF EXISTS idx_ci_reports_identity_task;
CREATE INDEX IF NOT EXISTS idx_ci_reports_username_task
    ON ci_reports (username, task_name, created_at DESC);

-- +goose Down

DROP INDEX IF EXISTS idx_ci_reports_username_task;
ALTER TABLE ci_reports DROP COLUMN IF EXISTS username;
ALTER TABLE ci_reports ADD COLUMN IF NOT EXISTS identity_id UUID;
CREATE INDEX IF NOT EXISTS idx_ci_reports_identity_task
    ON ci_reports (identity_id, task_name, created_at DESC);
