package handlers

import (
	"errors"
	"net/http"

	httpserver "users/internal/http/gen"
	"users/internal/authctx"
	postgresrepo "users/internal/repository/postgres"
	service "users/internal/services"
	"users/internal/http/validation"

	"github.com/labstack/echo/v4"
)

func errorResponse(ctx echo.Context, status int, code httpserver.ErrorCode, err error) error {
	return ctx.JSON(status, httpserver.ErrorResponse{
		Code:    code,
		Message: err.Error(),
	})
}

func validationError(ctx echo.Context, err error) error {
	return errorResponse(ctx, http.StatusUnprocessableEntity, httpserver.VALIDATIONERROR, err)
}

func serviceError(ctx echo.Context, err error) error {
	status := http.StatusInternalServerError
	code := httpserver.INTERNALERROR

	switch {
	case errors.Is(err, authctx.ErrAuthContextMissing):
		status = http.StatusUnauthorized
		code = httpserver.MISSINGAUTHTOKEN
	case errors.Is(err, postgresrepo.ErrUserProfileNotFound):
		status = http.StatusNotFound
		code = httpserver.NOTFOUND
	case errors.Is(err, postgresrepo.ErrGitUserNotFound):
		status = http.StatusNotFound
		code = httpserver.NOTFOUND
	case errors.Is(err, service.ErrAvatarNotFound):
		status = http.StatusNotFound
		code = httpserver.NOTFOUND
	case errors.Is(err, validation.ErrDisplayNameEmpty),
		errors.Is(err, validation.ErrDisplayNameTooLong),
		errors.Is(err, validation.ErrBioTooLong):
		status = http.StatusUnprocessableEntity
		code = httpserver.VALIDATIONERROR
	}

	return errorResponse(ctx, status, code, err)
}
