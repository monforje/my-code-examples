package authclihandlers

import (
	"net/http"

	httpserver "auth/internal/http/gen"

	"github.com/labstack/echo/v4"
)

// AuthDeviceStart - запуск авторизации CLI
/*
	1. Вызвать AuthService.DeviceStart.
	2. Вернуть device_code, user_code, verification_url, expires_in, interval.
	3. Ошибки service перевести в ErrorResponse.
*/
func (h *AuthCliHandlers) AuthDeviceStart(ctx echo.Context) error {
	// 1. Вызвать AuthService.DeviceStart.
	output, err := h.as.DeviceStart(ctx.Request().Context())
	if err != nil {
		return serviceError(ctx, err)
	}

	// 2. Вернуть ответ.
	return ctx.JSON(http.StatusOK, httpserver.DeviceStartResponse{
		DeviceCode:      output.DeviceCode,
		UserCode:        output.UserCode,
		VerificationUrl: output.VerificationURL,
		ExpiresIn:       output.ExpiresIn,
		Interval:        output.Interval,
	})
}
