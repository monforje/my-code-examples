package reportshandler

import (
	"errors"
	"net/http"

	httpserver "tasks/internal/http/gen"
	reportsservice "tasks/internal/services/reports"

	"github.com/labstack/echo/v4"
)

func errorResponse(ctx echo.Context, status int, code httpserver.ErrorCode, err error) error {
	return ctx.JSON(status, httpserver.ErrorResponse{
		Code:    code,
		Message: err.Error(),
	})
}

// reportServiceError - маппинг ошибок сервиса отчётов в HTTP-ответы.
func reportServiceError(ctx echo.Context, err error) error {
	status := http.StatusInternalServerError
	code := httpserver.INTERNALERROR

	switch {
	case errors.Is(err, reportsservice.ErrTaskNotFound):
		status = http.StatusNotFound
		code = httpserver.NOTFOUND
	case errors.Is(err, reportsservice.ErrReportNotFound):
		status = http.StatusNotFound
		code = httpserver.NOTFOUND
	case errors.Is(err, reportsservice.ErrUIDInvalid):
		status = http.StatusUnprocessableEntity
		code = httpserver.VALIDATIONERROR
	case errors.Is(err, reportsservice.ErrRunIDInvalid), errors.Is(err, reportsservice.ErrReportIDInvalid):
		status = http.StatusUnprocessableEntity
		code = httpserver.VALIDATIONERROR
	case errors.Is(err, reportsservice.ErrUsernameNotFound):
		status = http.StatusNotFound
		code = httpserver.NOTFOUND
	case errors.Is(err, reportsservice.ErrUIDEmpty), errors.Is(err, reportsservice.ErrCommitEmpty),
		errors.Is(err, reportsservice.ErrTaskNameEmpty):
		status = http.StatusUnprocessableEntity
		code = httpserver.VALIDATIONERROR
	}

	return errorResponse(ctx, status, code, err)
}
