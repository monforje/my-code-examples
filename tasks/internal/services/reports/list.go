package reportsservice

import (
	"context"

	"tasks/internal/models/records"
	clientsdto "tasks/pkg/http_clients/dto"
)

// ListReports - возвращает постраничный список CI-отчётов пользователя.
// Если TaskName задан — фильтрует по задаче, иначе — все отчёты.
/*
   Алгоритм:
   1. Резолвить identity_id → username через users-сервис.
   2. Запросить страницу из репозитория (limit+1 для определения hasNextPage).
   3. Разобрать JSON-payload каждой записи в domain.Report.
   4. Сформировать next_cursor из id последнего элемента, если есть следующая страница.
*/
func (s *ReportsService) ListReports(ctx context.Context, in *ListInput) (*ListOutput, error) {
	limit := in.Limit
	if limit <= 0 {
		limit = 20
	}

	// 1.
	gitUser, err := s.users.GetGitUser(ctx, &clientsdto.GitUserRequest{IdentityID: in.IdentityID})
	if err != nil || gitUser == nil || gitUser.Username == "" {
		return nil, ErrUsernameNotFound
	}

	// Фильтр по статусу (опционально).
	var status *string
	if in.Status != nil {
		s := string(*in.Status)
		status = &s
	}

	// 2.
	var rows []records.CIReport
	var hasNextPage bool
	if in.TaskName != "" {
		rows, hasNextPage, err = s.reports.ListByUsernameAndTask(ctx, gitUser.Username, in.TaskName, status, limit, in.Cursor)
	} else {
		rows, hasNextPage, err = s.reports.ListByUsername(ctx, gitUser.Username, status, limit, in.Cursor)
	}
	if err != nil {
		return nil, err
	}

	// 3. + 4.
	items := make([]ReportOutput, 0, len(rows))
	var nextCursor *string
	for i, row := range rows {
		report, err := payloadToReport(row.UID, row.Commit, row.RunID.String(), row.CreatedAt, row.Payload)
		if err != nil {
			return nil, err
		}
		items = append(items, ReportOutput{ID: row.ID, Report: *report})
		if hasNextPage && i == len(rows)-1 {
			id := row.ID.String()
			nextCursor = &id
		}
	}

	return &ListOutput{
		Items:       items,
		HasNextPage: hasNextPage,
		NextCursor:  nextCursor,
	}, nil
}
