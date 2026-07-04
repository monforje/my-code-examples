// Package reportshandler
package reportshandler

import (
	"context"
	"net/http"

	"tasks/internal/authctx"
	httpserver "tasks/internal/http/gen"
	"tasks/internal/models/domain"
	reportsservice "tasks/internal/services/reports"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type reportsService interface {
	CreateReport(ctx context.Context, in *reportsservice.CreateReportInput) (*reportsservice.ReportOutput, error)
	ListReports(ctx context.Context, in *reportsservice.ListInput) (*reportsservice.ListOutput, error)
	GetReport(ctx context.Context, in *reportsservice.GetInput) (*reportsservice.ReportOutput, error)
}

// ReportsHandler - обработчики приёма и выдачи CI-отчётов.
type ReportsHandler struct {
	svc reportsService
}

// NewReportsHandler - конструктор.
func NewReportsHandler(svc reportsService) *ReportsHandler {
	return &ReportsHandler{svc: svc}
}

// ReportsCreate - принимает CI-отчёт от notifications (сервисный токен).
/*
   1. Распарсить JSON body в CreateReportRequest.
   2. Смаппить в CreateReportInput.
   3. Вызвать сервис CreateReport (резолвит identity+task_name по uid).
   4. Вернуть Report (201).
*/
func (h *ReportsHandler) ReportsCreate(ctx echo.Context) error {
	// 1.
	var req httpserver.CreateReportRequest
	if err := ctx.Bind(&req); err != nil {
		return errorResponse(ctx, http.StatusBadRequest, httpserver.INVALIDJSON, err)
	}

	// 2.
	input := mapCreateRequest(&req)

	// 3.
	out, err := h.svc.CreateReport(ctx.Request().Context(), input)
	if err != nil {
		return reportServiceError(ctx, err)
	}

	// 4.
	return ctx.JSON(http.StatusCreated, mapReportOutput(out))
}

// ReportsList - возвращает список CI-отчётов пользователя (user JWT).
/*
   1. Достать identity_id из JWT (authctx).
   2. Собрать ListInput из query-параметров.
   3. Вызвать сервис ListReports.
   4. Вернуть ReportListResponse.
*/
func (h *ReportsHandler) ReportsList(ctx echo.Context, params httpserver.ReportsListParams) error {
	// 1.
	identityID, _, err := authctx.FromContext(ctx.Request().Context())
	if err != nil {
		return errorResponse(ctx, http.StatusUnauthorized, httpserver.MISSINGAUTHTOKEN, err)
	}

	// 2.
	var taskName string
	if params.TaskName != nil {
		taskName = *params.TaskName
	}
	var status *domain.ReportStatus
	if params.Status != nil {
		s := domain.ReportStatus(*params.Status)
		status = &s
	}
	input := &reportsservice.ListInput{
		IdentityID: identityID,
		TaskName:   taskName,
		Status:     status,
		Limit:      params.Limit,
		Cursor:     params.Cursor,
	}

	// 3.
	out, err := h.svc.ListReports(ctx.Request().Context(), input)
	if err != nil {
		return reportServiceError(ctx, err)
	}

	// 4.
	items := make([]httpserver.Report, 0, len(out.Items))
	for i := range out.Items {
		items = append(items, mapReportOutput(&out.Items[i]))
	}

	return ctx.JSON(http.StatusOK, httpserver.ReportListResponse{
		Items: items,
		PageInfo: httpserver.PageInfo{
			HasNextPage: out.HasNextPage,
			NextCursor:  out.NextCursor,
		},
	})
}

// ReportsGet - возвращает один CI-отчёт по id (user JWT, проверка владельца).
/*
   1. Достать identity_id из JWT (authctx).
   2. Распарсить reportId из path.
   3. Вызвать сервис GetReport (резолвит username и проверяет владельца).
   4. Вернуть Report.
*/
func (h *ReportsHandler) ReportsGet(ctx echo.Context, reportId string) error {
	// 1.
	identityID, _, err := authctx.FromContext(ctx.Request().Context())
	if err != nil {
		return errorResponse(ctx, http.StatusUnauthorized, httpserver.MISSINGAUTHTOKEN, err)
	}

	// 2.
	id, err := uuid.Parse(reportId)
	if err != nil {
		return reportServiceError(ctx, reportsservice.ErrReportIDInvalid)
	}

	// 3.
	out, err := h.svc.GetReport(ctx.Request().Context(), &reportsservice.GetInput{
		IdentityID: identityID,
		ReportID:   id,
	})
	if err != nil {
		return reportServiceError(ctx, err)
	}

	// 4.
	return ctx.JSON(http.StatusOK, mapReportOutput(out))
}
