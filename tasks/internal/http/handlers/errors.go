package handlers

import (
	"errors"
	"net/http"

	httpserver "tasks/internal/http/gen"
	"tasks/internal/http/validation"
	taskservice "tasks/internal/services"

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

// serviceError - маппинг ошибок сервиса в HTTP-ответы.
/*
	Переводит бизнес-ошибки TasksService в соответствующие HTTP-статусы и ErrorCode:
	- ErrTaskNotFound -> 404 NOT_FOUND
	- ErrTitleEmpty -> 422 VALIDATION_ERROR
	- ErrTaskIDInvalid -> 400 BAD_REQUEST
	- все остальные -> 500 INTERNAL_ERROR
*/
func serviceError(ctx echo.Context, err error) error {
	status := http.StatusInternalServerError
	code := httpserver.INTERNALERROR

	switch {
	case errors.Is(err, taskservice.ErrTaskNotFound):
		status = http.StatusNotFound
		code = httpserver.NOTFOUND
	case errors.Is(err, taskservice.ErrTitleEmpty):
		status = http.StatusUnprocessableEntity
		code = httpserver.VALIDATIONERROR
	case errors.Is(err, validation.ErrTaskIDInvalid):
		status = http.StatusBadRequest
		code = httpserver.VALIDATIONERROR
	}

	return errorResponse(ctx, status, code, err)
}
