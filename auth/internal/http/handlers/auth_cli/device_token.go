package authclihandlers

import (
	"errors"
	"net/http"

	httpserver "auth/internal/http/gen"
	authservice "auth/internal/services/auth"

	"github.com/labstack/echo/v4"
)

// AuthDeviceToken - CLI опрашивает endpoint для получения токенов
/*
	1. Распарсить JSON body в httpserver.DeviceTokenRequest.
	2. Собрать authservice.DeviceTokenInput.
	3. Вызвать AuthService.DeviceToken.
	4. Вернуть 200 OK с access_token, refresh_token, expires_in, token_type.
	5. Ошибки перевести в ErrorResponse (428 для pending, 429 для pollTooFrequent).
*/
func (h *AuthCliHandlers) AuthDeviceToken(ctx echo.Context) error {
	// 1. Распарсить JSON body.
	var req httpserver.DeviceTokenRequest
	if err := ctx.Bind(&req); err != nil {
		return echo.ErrBadRequest
	}

	// 2. Собрать input.
	input := authservice.DeviceTokenInput{
		DeviceCode: req.DeviceCode,
	}

	// 3. Вызвать сервис.
	output, err := h.as.DeviceToken(ctx.Request().Context(), &input)
	if err != nil {
		// Специальная обработка для 428 Precondition Required.
		if errors.Is(err, authservice.ErrDeviceCodeNotConfirmed) {
			return errorResponse(ctx, http.StatusPreconditionRequired, httpserver.VALIDATIONERROR, err)
		}
		return serviceError(ctx, err)
	}

	// 4. Вернуть ответ.
	return ctx.JSON(http.StatusOK, httpserver.CliTokenResponse{
		AccessToken:  output.AccessToken,
		RefreshToken: output.RefreshToken,
		ExpiresIn:    output.ExpiresIn,
		TokenType:    httpserver.CliTokenResponseTokenTypeBearer,
	})
}
