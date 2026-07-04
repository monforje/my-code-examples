package authclihandlers

import (
	"net/http"

	httpserver "auth/internal/http/gen"
	authservice "auth/internal/services/auth"

	"github.com/labstack/echo/v4"
)

// AuthCliRefresh - обновление CLI access token
/*
	1. Распарсить JSON body в httpserver.CliRefreshRequest.
	2. Собрать authservice.CliRefreshInput.
	3. Вызвать AuthService.CliRefresh.
	4. Вернуть 200 OK с access_token, expires_in, token_type.
	5. Ошибки перевести в ErrorResponse.
*/
func (h *AuthCliHandlers) AuthCliRefresh(ctx echo.Context) error {
	// 1. Распарсить JSON body.
	var req httpserver.CliRefreshRequest
	if err := ctx.Bind(&req); err != nil {
		return echo.ErrBadRequest
	}

	// 2. Собрать input.
	input := authservice.CliRefreshInput{
		RefreshToken: req.RefreshToken,
	}

	// 3. Вызвать сервис.
	output, err := h.as.CliRefresh(ctx.Request().Context(), &input)
	if err != nil {
		return serviceError(ctx, err)
	}

	// 4. Вернуть ответ.
	return ctx.JSON(http.StatusOK, httpserver.CliRefreshResponse{
		AccessToken: output.AccessToken,
		RefreshToken: output.RefreshToken,
		ExpiresIn:   output.ExpiresIn,
		TokenType:   httpserver.CliRefreshResponseTokenTypeBearer,
	})
}
