package authclihandlers

import (
	"errors"
	"net/http"

	httpserver "auth/internal/http/gen"
	authservice "auth/internal/services/auth"

	"github.com/labstack/echo/v4"
)

func errorResponse(ctx echo.Context, status int, code httpserver.ErrorCode, err error) error {
	return ctx.JSON(status, httpserver.ErrorResponse{
		Code:    code,
		Message: err.Error(),
	})
}

func serviceError(ctx echo.Context, err error) error {
	status := http.StatusInternalServerError
	code := httpserver.INTERNALERROR

	switch {
	case errors.Is(err, authservice.ErrDeviceCodeNotFound):
		status = http.StatusNotFound
		code = httpserver.NOTFOUND
	case errors.Is(err, authservice.ErrDeviceCodeExpired):
		status = http.StatusGone
		code = httpserver.EXPIREDCODE
	case errors.Is(err, authservice.ErrDeviceCodeAlreadyConfirmed):
		status = http.StatusConflict
		code = httpserver.VALIDATIONERROR
	case errors.Is(err, authservice.ErrDeviceCodeNotConfirmed):
		status = http.StatusPreconditionRequired
		code = httpserver.VALIDATIONERROR
	case errors.Is(err, authservice.ErrPollTooFrequent):
		status = http.StatusTooManyRequests
		code = httpserver.TOOMANYATTEMPTS
	case errors.Is(err, authservice.ErrInvalidRefreshToken), errors.Is(err, authservice.ErrSessionRevoked):
		status = http.StatusUnauthorized
		code = httpserver.INVALIDREFRESHTOKEN
	case errors.Is(err, authservice.ErrSessionExpired):
		status = http.StatusUnauthorized
		code = httpserver.EXPIREDREFRESHTOKEN
	case errors.Is(err, authservice.ErrIdentityNotActive):
		status = http.StatusForbidden
		code = httpserver.INVALIDAUTHTOKEN
	}

	return errorResponse(ctx, status, code, err)
}
