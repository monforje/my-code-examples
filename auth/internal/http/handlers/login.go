package handlers

import (
	"net"
	"net/http"

	httpserver "auth/internal/http/gen"
	"auth/internal/http/validation"
	authservice "auth/internal/services/auth"

	"github.com/labstack/echo/v4"
)

// AuthLogin - обработчик входа пользователя в аккаунт
/*
	1. Распарсить JSON body в httpserver.LoginRequest.
	2. Вызвать validation.ValidateLoginRequest: проверить email и password.
	3. Собрать authservice.LoginInput: email, password, user_agent и ip_address из HTTP-запроса.
	4. Вызвать AuthService.Login.
	5. Установить refresh_token в httpOnly cookie.
	6. Вернуть 200 OK с access_token и expires_in.
	7. Ошибки bind/validation/service перевести в единый ErrorResponse.
*/
func (h *AuthHandlers) AuthLogin(ctx echo.Context) error {
	// 1. Распарсить JSON body в httpserver.LoginRequest.
	var req httpserver.LoginRequest
	if err := ctx.Bind(&req); err != nil {
		return echo.ErrBadRequest
	}

	// 2. Вызвать validation.ValidateLoginRequest: проверить email и password.
	email, password, err := validation.ValidateLoginRequest(req.Email, req.Password)
	if err != nil {
		return ctx.JSON(http.StatusUnprocessableEntity, httpserver.ErrorResponse{
			Code:    httpserver.VALIDATIONERROR,
			Message: err.Error(),
		})
	}

	// 3. Собрать authservice.LoginInput: email, password, user_agent и ip_address из HTTP-запроса.
	userAgent := ctx.Request().UserAgent()
	ipStr := ctx.RealIP()
	var ipAddress net.IP
	if ipStr != "" {
		ipAddress = net.ParseIP(ipStr)
	}

	input := authservice.LoginInput{
		Email:     email,
		Password:  password,
		UserAgent: userAgent,
		IPAddress: ipAddress,
	}

	// 4. Вызвать AuthService.Login.
	output, err := h.as.Login(ctx.Request().Context(), &input)
	if err != nil {
		// 7. Ошибки bind/validation/service перевести в единый ErrorResponse.
		return serviceError(ctx, err)
	}

	// 5. Установить refresh_token в httpOnly cookie.
	setRefreshCookie(ctx, output.RefreshToken, h.refreshSessionTTL)

	// 6. Вернуть 200 OK с access_token и expires_in.
	return ctx.JSON(http.StatusOK, httpserver.TokenResponse{
		AccessToken: output.AccessToken,
		ExpiresIn:   output.ExpiresIn,
	})
}

// AuthLogout - обработчик выхода пользователя из аккаунта
/*
	1. BearerAuth middleware уже положил identity_id и session_id в context.
	2. Вызвать AuthService.Logout.
	3. Если сервис вернул ошибку, перевести ее в ErrorResponse.
	4. Очистить refresh_token cookie.
	5. Вернуть 204 No Content.
*/
func (h *AuthHandlers) AuthLogout(ctx echo.Context) error {
	// 1. BearerAuth middleware уже положил identity_id и session_id в context.

	// 2. Вызвать AuthService.Logout.
	if err := h.as.Logout(ctx.Request().Context()); err != nil {
		// 3. Если сервис вернул ошибку, перевести ее в ErrorResponse.
		return serviceError(ctx, err)
	}

	// 4. Очистить refresh_token cookie.
	clearRefreshCookie(ctx)

	// 5. Вернуть 204 No Content.
	return ctx.NoContent(http.StatusNoContent)
}

// AuthRefresh - обработчик обновления access token по refresh token
/*
	1. Достать refresh_token из httpOnly cookie.
	2. Собрать authservice.RefreshInput.
	3. Вызвать AuthService.Refresh.
	4. Установить новый refresh_token в httpOnly cookie.
	5. Вернуть 200 OK с новым access_token и expires_in.
	6. Ошибки service перевести в единый ErrorResponse.
*/
func (h *AuthHandlers) AuthRefresh(ctx echo.Context) error {
	// 1. Достать refresh_token из httpOnly cookie.
	refreshToken, err := ctx.Cookie(refreshCookieName)
	if err != nil || refreshToken.Value == "" {
		return ctx.JSON(http.StatusUnauthorized, httpserver.ErrorResponse{
			Code:    httpserver.MISSINGAUTHTOKEN,
			Message: "refresh_token cookie is missing",
		})
	}

	// 2. Собрать authservice.RefreshInput.
	input := authservice.RefreshInput{
		RefreshToken: refreshToken.Value,
	}

	// 3. Вызвать AuthService.Refresh.
	output, err := h.as.Refresh(ctx.Request().Context(), &input)
	if err != nil {
		// 6. Ошибки service перевести в единый ErrorResponse.
		return serviceError(ctx, err)
	}

	// 4. Установить новый refresh_token в httpOnly cookie.
	setRefreshCookie(ctx, output.RefreshToken, h.refreshSessionTTL)

	// 5. Вернуть 200 OK с новым access_token и expires_in.
	return ctx.JSON(http.StatusOK, httpserver.RefreshResponse{
		AccessToken: output.AccessToken,
		ExpiresIn:   output.ExpiresIn,
	})
}
