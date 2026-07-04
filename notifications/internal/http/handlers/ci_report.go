package handlers

import (
	"errors"
	"net/http"

	"notifications/internal/models/domain"
	cireportservice "notifications/internal/services/ci_report"

	httpserver "notifications/internal/http/gen"

	apperrors "notifications/pkg/errors"

	"github.com/labstack/echo/v4"
)

// WebhookReceive - принимает вебхук от task_runner.
/*
    1. Распарсить JSON body в WebhookRequest.
    2. Проверить обязательные поля (event, uid, commit).
    3. Смаппить в domain.WebhookPayload.
    4. Вызвать сервис ProcessWebhook.
    5. Вернуть 200 OK с сообщением.
*/
func (h *CiReportHandlers) WebhookReceive(ctx echo.Context) error {
	// 1.
	var req httpserver.WebhookRequest
	if err := ctx.Bind(&req); err != nil {
		return errorResponse(ctx, http.StatusBadRequest, httpserver.INVALIDJSON, err)
	}

	// 2.
	if req.Event == "" || req.Uid == "" || req.Commit == "" {
		return errorResponse(ctx, http.StatusUnprocessableEntity, httpserver.VALIDATIONERROR, errors.New("event, uid and commit are required"))
	}

	// 3.
	payload := &domain.WebhookPayload{
		Event:    domain.WebhookEvent(req.Event),
		UID:      req.Uid,
		Commit:   req.Commit,
		ExitCode: req.ExitCode,
		Stdout:   derefStr(req.Stdout),
		Stderr:   derefStr(req.Stderr),
		Stage:    derefStr(req.Stage),
	}
	if req.VmId != nil {
		payload.VMID = *req.VmId
	}

	// 4.
	if err := h.svc.ProcessWebhook(ctx.Request().Context(), payload); err != nil {
		return serviceError(ctx, err)
	}

	// 5.
	return ctx.JSON(http.StatusOK, httpserver.MessageResponse{
		Message: "ok",
	})
}

// ReportsGet - возвращает CI-отчёт по UID.
/*
    1. UID уже извлечён из path.
    2. Вызвать сервис GetReport.
    3. Смаппить домен в CIReportResponse.
    4. Вернуть JSON.
*/
func (h *CiReportHandlers) ReportsGet(ctx echo.Context, uid string) error {
	// 1.
	if uid == "" {
		return errorResponse(ctx, http.StatusBadRequest, httpserver.VALIDATIONERROR, errors.New("uid is required"))
	}

	// 2.
	report, err := h.svc.GetReport(ctx.Request().Context(), uid)
	if err != nil {
		return serviceError(ctx, err)
	}

	// 3.
	resp := mapReportToResponse(report)

	// 4.
	return ctx.JSON(http.StatusOK, resp)
}

func mapReportToResponse(r *domain.CIReport) httpserver.CIReportResponse {
	steps := make([]httpserver.CIReportStep, len(r.Steps))
	for i, s := range r.Steps {
		steps[i] = httpserver.CIReportStep{
			Name:       s.Name,
			Status:     httpserver.CIStepStatus(s.Status),
			DurationMs: s.DurationMs,
		}
	}

	return httpserver.CIReportResponse{
		Uid:       r.UID,
		Commit:    r.Commit,
		Status:    httpserver.CIReportStatus(r.Status),
		ExitCode:  r.ExitCode,
		Stage:     r.Stage,
		Steps:     steps,
		Stdout:    r.Stdout,
		Stderr:    r.Stderr,
		CreatedAt: r.CreatedAt,
	}
}

func derefStr(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}

func serviceError(ctx echo.Context, err error) error {
	status := http.StatusInternalServerError
	code := httpserver.INTERNALERROR

	if errors.Is(err, cireportservice.ErrReportNotFound) {
		status = http.StatusNotFound
		code = httpserver.NOTFOUND
	} else if errors.Is(err, cireportservice.ErrInvalidWebhook) {
		status = http.StatusUnprocessableEntity
		code = httpserver.VALIDATIONERROR
	} else {
		var ae *apperrors.Error
		if errors.As(err, &ae) {
			if errors.Is(ae.Err, cireportservice.ErrReportNotFound) {
				status = http.StatusNotFound
				code = httpserver.NOTFOUND
			} else if errors.Is(ae.Err, cireportservice.ErrInvalidWebhook) {
				status = http.StatusUnprocessableEntity
				code = httpserver.VALIDATIONERROR
			}
		}
	}

	return errorResponse(ctx, status, code, err)
}
