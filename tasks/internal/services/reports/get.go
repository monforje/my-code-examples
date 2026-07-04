package reportsservice

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"

	clientsdto "tasks/pkg/http_clients/dto"
)

// GetReport - возвращает один CI-отчёт по id с проверкой владельца.
/*
   Алгоритм:
   1. Резолвить identity_id → username через users-сервис.
   2. Запросить запись по (id, username) — фильтр по username гарантирует,
      что пользователь не может прочитать чужой отчёт.
   3. Разобрать JSON-payload в domain.Report.
*/
func (s *ReportsService) GetReport(ctx context.Context, in *GetInput) (*ReportOutput, error) {
	// 1.
	gitUser, err := s.users.GetGitUser(ctx, &clientsdto.GitUserRequest{IdentityID: in.IdentityID})
	if err != nil || gitUser == nil || gitUser.Username == "" {
		return nil, ErrUsernameNotFound
	}

	// 2.
	row, err := s.reports.GetByIDAndUsername(ctx, in.ReportID, gitUser.Username)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrReportNotFound
		}
		return nil, err
	}

	// 3.
	report, err := payloadToReport(row.UID, row.Commit, row.RunID.String(), row.CreatedAt, row.Payload)
	if err != nil {
		return nil, err
	}

	return &ReportOutput{ID: row.ID, Report: *report}, nil
}
