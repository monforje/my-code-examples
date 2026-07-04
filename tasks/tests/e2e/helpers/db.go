package e2e_test_helpers

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// TaskRow - строка таблицы tasks для проверки в БД.
type TaskRow struct {
	ID                  uuid.UUID
	Title               string
	Description         string
	SpecificationMDText string
	TaskType            string
	Level               string
}

// GetTaskByID - получение задачи по ID из БД.
func GetTaskByID(t *testing.T, pool *pgxpool.Pool, id uuid.UUID) TaskRow {
	t.Helper()
	var row TaskRow
	err := pool.QueryRow(context.Background(),
		`SELECT id, title, description, specification_md_text, task_type, level FROM tasks WHERE id = $1`, id,
	).Scan(&row.ID, &row.Title, &row.Description, &row.SpecificationMDText, &row.TaskType, &row.Level)
	if err != nil {
		t.Fatalf("GetTaskByID(%s): %v", id, err)
	}
	return row
}

// CountTasks - подсчёт количества задач в БД.
func CountTasks(t *testing.T, pool *pgxpool.Pool) int {
	t.Helper()
	var count int
	err := pool.QueryRow(context.Background(), `SELECT COUNT(*) FROM tasks`).Scan(&count)
	if err != nil {
		t.Fatalf("CountTasks: %v", err)
	}
	return count
}

// RequireNoTask - проверка отсутствия задачи в БД.
func RequireNoTask(t *testing.T, pool *pgxpool.Pool, id uuid.UUID) {
	t.Helper()
	var discard int
	err := pool.QueryRow(context.Background(),
		`SELECT 1 FROM tasks WHERE id = $1`, id,
	).Scan(&discard)
	if err == nil {
		t.Fatalf("task %s should not exist", id)
	}
}
