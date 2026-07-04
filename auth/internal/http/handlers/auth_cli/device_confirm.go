package authclihandlers

import (
	"net/http"

	httpserver "auth/internal/http/gen"
	authservice "auth/internal/services/auth"

	"github.com/labstack/echo/v4"
)

// AuthDeviceConfirm - подтверждение CLI-входа из браузера
/*
	1. Распарсить JSON body в httpserver.DeviceConfirmRequest.
	2. Собрать authservice.DeviceConfirmInput.
	3. Вызвать AuthService.DeviceConfirm.
	4. Вернуть 200 OK с status: "confirmed".
	5. Ошибки перевести в ErrorResponse.
*/
func (h *AuthCliHandlers) AuthDeviceConfirm(ctx echo.Context) error {
	// 1. Распарсить JSON body.
	var req httpserver.DeviceConfirmRequest
	if err := ctx.Bind(&req); err != nil {
		return echo.ErrBadRequest
	}

	// 2. Собрать input.
	input := authservice.DeviceConfirmInput{
		UserCode: req.UserCode,
	}

	// 3. Вызвать сервис.
	output, err := h.as.DeviceConfirm(ctx.Request().Context(), &input)
	if err != nil {
		return serviceError(ctx, err)
	}

	// 4. Вернуть ответ.
	return ctx.JSON(http.StatusOK, httpserver.DeviceConfirmResponse{
		Status: httpserver.DeviceConfirmResponseStatus(output.Status),
	})
}
