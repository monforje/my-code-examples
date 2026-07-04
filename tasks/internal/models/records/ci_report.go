package records

import (
	"time"

	"github.com/google/uuid"
)

// CIReport - сохранённый CI-отчёт, полученный от сервиса notifications.
// Username и TaskName резолвятся на стороне tasks из uid (см. vmpool.go:179).
// RunID - идемпотентный идентификатор прогона (связывает pending и финальный статус).
// Payload - полный JSON-отчёт (summary, steps, lint_errors, warnings) для UI.
type CIReport struct {
	ID        uuid.UUID `db:"id"`
	Username  string    `db:"username"`
	TaskName  string    `db:"task_name"`
	UID       string    `db:"uid"`
	Commit    string    `db:"commit"`
	RunID     uuid.UUID `db:"run_id"`
	Status    string    `db:"status"`
	CreatedAt time.Time `db:"created_at"`
	Payload   []byte    `db:"payload"`
}
