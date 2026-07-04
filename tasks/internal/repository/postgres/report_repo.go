package postgresrepo

import (
	"context"
	"errors"
	"strconv"
	"strings"

	"github.com/google/uuid"

	"tasks/internal/models/records"
)

// ErrCIReportNotFound - CI-отчёт не найден.
var ErrCIReportNotFound = errors.New("ci report not found")

type ReportRepo struct {
	*Repo
}

func NewReportRepo(repo *Repo) *ReportRepo {
	return &ReportRepo{Repo: repo}
}

// reportColumns - общий список колонок для SELECT.
const reportColumns = `id, username, task_name, uid, commit, run_id, status, created_at, payload`

// scanReport - сканирует строку в records.CIReport.
func scanReport(scanner interface {
	Scan(dest ...any) error
}) (records.CIReport, error) {
	var rep records.CIReport
	err := scanner.Scan(
		&rep.ID, &rep.Username, &rep.TaskName, &rep.UID,
		&rep.Commit, &rep.RunID, &rep.Status, &rep.CreatedAt, &rep.Payload,
	)
	return rep, err
}

// Upsert - идемпотентное сохранение CI-отчёта по (uid, commit).
// Один коммит = один прогон: pending (ci_started) обновляется финальным статусом
// (ci_finished), повторные/внеочерёдные доставки — no-op/обновление. Стабильные
// id и created_at сохраняются (курсорная пагинация не дёргается). Возвращает итоговый id.
func (r *ReportRepo) Upsert(ctx context.Context, rep *records.CIReport) error {
	const query = `
		INSERT INTO ci_reports (id, run_id, username, task_name, uid, commit, status, created_at, payload)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT (uid, commit) DO UPDATE SET
			status  = EXCLUDED.status,
			run_id  = EXCLUDED.run_id,
			payload = EXCLUDED.payload
		RETURNING id`
	return r.QueryRow(ctx, query,
		rep.ID, rep.RunID, rep.Username, rep.TaskName, rep.UID,
		rep.Commit, rep.Status, rep.CreatedAt, rep.Payload,
	).Scan(&rep.ID)
}

// GetByIDAndUsername - возвращает отчёт по id, принадлежащий username.
// Отсутствие строки (ErrNoRows) трактуется сервисом как "не найден / недоступен".
func (r *ReportRepo) GetByIDAndUsername(
	ctx context.Context,
	id uuid.UUID,
	username string,
) (*records.CIReport, error) {
	query := `
		SELECT ` + reportColumns + `
		FROM ci_reports
		WHERE id = $1 AND username = $2`

	rep, err := scanReport(r.QueryRow(ctx, query, id, username))
	if err != nil {
		return nil, err
	}
	return &rep, nil
}

// appendStatusCond - добавляет опциональный фильтр по статусу в запрос.
// Возвращает SQL-фрагмент, доп. аргументы и следующий индекс аргумента.
func appendStatusCond(cond *strings.Builder, args []any, argIdx int, status *string) (int, []any) {
	if status == nil || *status == "" {
		return argIdx, args
	}
	cond.WriteString(` AND status = $`)
	cond.WriteString(strconv.Itoa(argIdx))
	args = append(args, *status)
	return argIdx + 1, args
}

// ListByUsernameAndTask - курсорная пагинация отчётов пользователя для задачи.
// Курсор — id последнего элемента предыдущей страницы (как у TaskRepo.List).
func (r *ReportRepo) ListByUsernameAndTask(
	ctx context.Context,
	username string,
	taskName string,
	status *string,
	limit int32,
	cursor *string,
) ([]records.CIReport, bool, error) {
	args := []any{username, taskName}
	argIdx := 3 // $1=username, $2=taskName

	var cond strings.Builder
	argIdx, args = appendStatusCond(&cond, args, argIdx, status)

	if cursor != nil {
		cond.WriteString(` AND (created_at, id) < (SELECT c.created_at, c.id FROM ci_reports c WHERE c.id = $`)
		cond.WriteString(strconv.Itoa(argIdx))
		cond.WriteString(`)`)
		args = append(args, *cursor)
		argIdx++
	}

	query := `
		SELECT ` + reportColumns + `
		FROM ci_reports
		WHERE username = $1 AND task_name = $2` + cond.String() + `
		ORDER BY created_at DESC, id DESC
		LIMIT $` + strconv.Itoa(argIdx)

	args = append(args, limit+1)

	rows, err := r.Query(ctx, query, args...)
	if err != nil {
		return nil, false, err
	}
	defer rows.Close()

	items := make([]records.CIReport, 0, limit)
	for rows.Next() {
		rep, err := scanReport(rows)
		if err != nil {
			return nil, false, err
		}
		items = append(items, rep)
	}
	if err := rows.Err(); err != nil {
		return nil, false, err
	}

	hasNextPage := int32(len(items)) > limit
	if hasNextPage {
		items = items[:int(limit)]
	}
	return items, hasNextPage, nil
}

// ListByUsername - курсорная пагинация всех отчётов пользователя (без фильтра по задаче).
func (r *ReportRepo) ListByUsername(
	ctx context.Context,
	username string,
	status *string,
	limit int32,
	cursor *string,
) ([]records.CIReport, bool, error) {
	args := []any{username}
	argIdx := 2 // $1=username

	var cond strings.Builder
	argIdx, args = appendStatusCond(&cond, args, argIdx, status)

	if cursor != nil {
		cond.WriteString(` AND (created_at, id) < (SELECT c.created_at, c.id FROM ci_reports c WHERE c.id = $`)
		cond.WriteString(strconv.Itoa(argIdx))
		cond.WriteString(`)`)
		args = append(args, *cursor)
		argIdx++
	}

	query := `
		SELECT ` + reportColumns + `
		FROM ci_reports
		WHERE username = $1` + cond.String() + `
		ORDER BY created_at DESC, id DESC
		LIMIT $` + strconv.Itoa(argIdx)

	args = append(args, limit+1)

	rows, err := r.Query(ctx, query, args...)
	if err != nil {
		return nil, false, err
	}
	defer rows.Close()

	items := make([]records.CIReport, 0, limit)
	for rows.Next() {
		rep, err := scanReport(rows)
		if err != nil {
			return nil, false, err
		}
		items = append(items, rep)
	}
	if err := rows.Err(); err != nil {
		return nil, false, err
	}

	hasNextPage := int32(len(items)) > limit
	if hasNextPage {
		items = items[:int(limit)]
	}
	return items, hasNextPage, nil
}
