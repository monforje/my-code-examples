package handlers

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

func validationError(ctx echo.Context, err error) error {
	return errorResponse(ctx, http.StatusUnprocessableEntity, httpserver.VALIDATIONERROR, err)
}

func serviceError(ctx echo.Context, err error) error {
	status := http.StatusInternalServerError
	code := httpserver.INTERNALERROR

	switch {
	case errors.Is(err, authservice.ErrInvalidCredentials):
		status = http.StatusUnauthorized
		code = httpserver.INVALIDCREDENTIALS
	case errors.Is(err, authservice.ErrEmailNotVerified):
		status = http.StatusForbidden
		code = httpserver.EMAILNOTVERIFIED
	case errors.Is(err, authservice.ErrInvalidRefreshToken), errors.Is(err, authservice.ErrSessionRevoked):
		status = http.StatusUnauthorized
		code = httpserver.INVALIDREFRESHTOKEN
	case errors.Is(err, authservice.ErrSessionExpired):
		status = http.StatusUnauthorized
		code = httpserver.EXPIREDREFRESHTOKEN
	case errors.Is(err, authservice.ErrEmailAlreadyExists):
		status = http.StatusConflict
		code = httpserver.EMAILALREADYEXISTS
	case errors.Is(err, authservice.ErrEmailAlreadyVerified):
		status = http.StatusUnprocessableEntity
		code = httpserver.VALIDATIONERROR
	case errors.Is(err, authservice.ErrIdentityNotFound), errors.Is(err, authservice.ErrAccountDeleteNotFound), errors.Is(err, authservice.ErrPasswordChangeNotFound):
		status = http.StatusNotFound
		code = httpserver.NOTFOUND
	case errors.Is(err, authservice.ErrInvalidCode):
		status = http.StatusUnprocessableEntity
		code = httpserver.INVALIDCODE
	case errors.Is(err, authservice.ErrTooManyAttempts):
		status = http.StatusTooManyRequests
		code = httpserver.TOOMANYATTEMPTS
	case errors.Is(err, authservice.ErrCurrentPasswordIncorrect):
		status = http.StatusUnprocessableEntity
		code = httpserver.CURRENTPASSWORDINCORRECT
	case errors.Is(err, authservice.ErrInvalidResetToken):
		status = http.StatusUnprocessableEntity
		code = httpserver.RESETTOKENINVALID
	case errors.Is(err, authservice.ErrResetTokenExpired):
		status = http.StatusUnprocessableEntity
		code = httpserver.RESETTOKENEXPIRED
	case errors.Is(err, authservice.ErrInvalidChangeToken), errors.Is(err, authservice.ErrInvalidEmailChangeToken), errors.Is(err, authservice.ErrEmailChangeTokenExpired):
		status = http.StatusUnprocessableEntity
		code = httpserver.VALIDATIONERROR
	case errors.Is(err, authservice.ErrIdentityNotActive), errors.Is(err, authservice.ErrIdentityDeleted):
		status = http.StatusForbidden
		code = httpserver.INVALIDAUTHTOKEN
	}

	return errorResponse(ctx, status, code, err)
}
