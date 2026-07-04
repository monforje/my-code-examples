-- +goose Up
-- run_id: идентификатор CI-прогона от ci-translator (поле трассировки, опционален,
-- т.к. на этапе ci_started CI ещё не запущен и run_id неизвестен).
-- Идемпотентность и связь pending (ci_started) → финал (ci_finished) обеспечивается
-- уникальностью (uid, commit): один коммит = один прогон (task_runner коалесит пуши).

ALTER TABLE ci_reports
    ADD COLUMN IF NOT EXISTS run_id UUID NOT NULL DEFAULT gen_random_uuid();

-- Один отчёт на (uid, commit): pending обновляется финальным статусом, дубли — no-op.
CREATE UNIQUE INDEX IF NOT EXISTS idx_ci_reports_uid_commit
    ON ci_reports (uid, commit);

-- +goose Down

DROP INDEX IF EXISTS idx_ci_reports_uid_commit;
ALTER TABLE ci_reports DROP COLUMN IF EXISTS run_id;
