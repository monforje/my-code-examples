package reportsservice

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"tasks/internal/models/domain"
	"tasks/internal/models/records"
)

// CreateReport - парсит uid, валидирует задачу и сохраняет CI-отчёт (идемпотентно).
/*
   Алгоритм:
   1. Валидировать обязательные поля (uid, commit). run_id опционален (поле трассировки).
   2. Разобрать uid → username + task_name; run_id → UUID (или сгенерировать).
   3. Проверить, что задача существует (task_name из таблицы tasks).
   4. Сериализовать отчёт в JSON-payload.
   5. UPSERT записи по (uid, commit): pending (ci_started) → финал (ci_finished)
      для одного прогона, дедуп повторных/внеочерёдных доставок.
   6. Вернуть результат.
*/
func (s *ReportsService) CreateReport(ctx context.Context, in *CreateReportInput) (*ReportOutput, error) {
	// 1.
	if in == nil || in.UID == "" {
		return nil, ErrUIDEmpty
	}
	if in.Commit == "" {
		return nil, ErrCommitEmpty
	}

	// 2.
	username, taskName, err := parseUID(in.UID)
	if err != nil {
		return nil, err
	}

	// run_id опционален: если не передан (ci_started) — генерируем локально для трассировки.
	runID := uuid.New()
	if in.RunID != "" {
		runID, err = uuid.Parse(in.RunID)
		if err != nil {
			return nil, ErrRunIDInvalid
		}
	}

	// 3.
	if _, err := s.tasks.GetByTaskName(ctx, taskName); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrTaskNotFound
		}
		return nil, err
	}

	// 4.
	payloadBytes, err := json.Marshal(toPayload(in))
	if err != nil {
		return nil, err
	}

	// 5.
	createdAt := in.CreatedAt
	if createdAt.IsZero() {
		createdAt = time.Now().UTC()
	}

	rep := &records.CIReport{
		ID:        uuid.New(),
		RunID:     runID,
		Username:  username,
		TaskName:  taskName,
		UID:       in.UID,
		Commit:    in.Commit,
		Status:    string(in.Status),
		CreatedAt: createdAt,
		Payload:   payloadBytes,
	}
	if err := s.reports.Upsert(ctx, rep); err != nil {
		return nil, err
	}

	return &ReportOutput{
		ID: rep.ID,
		Report: domain.Report{
			UID:             in.UID,
			Commit:          in.Commit,
			RunID:           rep.RunID.String(),
			Status:          in.Status,
			CreatedAt:       createdAt,
			Summary:         in.Summary,
			Steps:           in.Steps,
			LintErrors:      in.LintErrors,
			Warnings:        in.Warnings,
			RawLogAvailable: in.RawLogAvailable,
		},
	}, nil
}
