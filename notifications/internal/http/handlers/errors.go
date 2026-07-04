package handlers

import (
	"net/http"

	httpserver "notifications/internal/http/gen"

	"github.com/labstack/echo/v4"
)

func errorResponse(ctx echo.Context, status int, code httpserver.ErrorCode, err error) error {
	return ctx.JSON(status, httpserver.ErrorResponse{
		Code:    code,
		Message: err.Error(),
	})
}

// validationError - возврат ошибки валидации (422).
func validationError(ctx echo.Context, err error) error {
	return errorResponse(ctx, http.StatusUnprocessableEntity, httpserver.VALIDATIONERROR, err)
}
