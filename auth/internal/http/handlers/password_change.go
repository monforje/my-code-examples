package handlers

import (
	"net/http"

	httpserver "auth/internal/http/gen"
	"auth/internal/http/validation"
	authservice "auth/internal/services/auth"

	"github.com/labstack/echo/v4"
)

// AuthPasswordChange - обработчик запуска смены пароля авторизованным пользователем
/*
	1. BearerAuth middleware уже положил identity_id и session_id в context.
	2. Распарсить JSON body в httpserver.ChangePasswordRequest.
	3. Вызвать validation.ValidateChangePasswordRequest: проверить current_password.
	4. Собрать authservice.ChangePasswordInput.
	5. Вызвать AuthService.ChangePassword.
	6. Вернуть 200 OK с httpserver.MessageResponse.
	7. Ошибки bind/validation/service перевести в единый ErrorResponse.
*/
func (h *AuthHandlers) AuthPasswordChange(ctx echo.Context) error {
	// 1. BearerAuth middleware уже положил identity_id и session_id в context.

	// 2. Распарсить JSON body в httpserver.ChangePasswordRequest.
	var req httpserver.ChangePasswordRequest
	if err := ctx.Bind(&req); err != nil {
		return echo.ErrBadRequest
	}

	// 3. Вызвать validation.ValidateChangePasswordRequest: проверить current_password.
	currentPassword, err := validation.ValidateChangePasswordRequest(req.CurrentPassword)
	if err != nil {
		return validationError(ctx, err)
	}

	// 4. Собрать authservice.ChangePasswordInput.
	input := authservice.ChangePasswordInput{
		CurrentPassword: currentPassword,
	}

	// 5. Вызвать AuthService.ChangePassword.
	message, err := h.as.ChangePassword(ctx.Request().Context(), &input)
	if err != nil {
		// 7. Ошибки bind/validation/service перевести в единый ErrorResponse.
		return serviceError(ctx, err)
	}

	// 6. Вернуть 200 OK с httpserver.MessageResponse.
	return ctx.JSON(http.StatusOK, httpserver.MessageResponse{
		Message: message,
	})
}

// AuthPasswordChangeVerify - обработчик подтверждения кода смены пароля
/*
	1. BearerAuth middleware уже положил identity_id и session_id в context.
	2. Распарсить JSON body в httpserver.CodeVerifyRequest.
	3. Вызвать validation.ValidateCodeVerifyRequest: проверить code.
	4. Собрать authservice.ChangePasswordVerifyInput.
	5. Вызвать AuthService.ChangePasswordVerify.
	6. Смаппить сервисный ChangePasswordVerifyOutput в httpserver.ChangePasswordTokenResponse.
	7. Вернуть 200 OK с change_token и expires_in.
	8. Ошибки bind/validation/service перевести в единый ErrorResponse.
*/
func (h *AuthHandlers) AuthPasswordChangeVerify(ctx echo.Context) error {
	// 1. BearerAuth middleware уже положил identity_id и session_id в context.

	// 2. Распарсить JSON body в httpserver.CodeVerifyRequest.
	var req httpserver.CodeVerifyRequest
	if err := ctx.Bind(&req); err != nil {
		return echo.ErrBadRequest
	}

	// 3. Вызвать validation.ValidateCodeVerifyRequest: проверить code.
	code, err := validation.ValidateCodeVerifyRequest(req.Code)
	if err != nil {
		return validationError(ctx, err)
	}

	// 4. Собрать authservice.ChangePasswordVerifyInput.
	input := authservice.ChangePasswordVerifyInput{
		Code: code,
	}

	// 5. Вызвать AuthService.ChangePasswordVerify.
	output, err := h.as.ChangePasswordVerify(ctx.Request().Context(), &input)
	if err != nil {
		// 8. Ошибки bind/validation/service перевести в единый ErrorResponse.
		return serviceError(ctx, err)
	}

	// 6. Смаппить сервисный ChangePasswordVerifyOutput в httpserver.ChangePasswordTokenResponse.
	// 7. Вернуть 200 OK с change_token и expires_in.
	return ctx.JSON(http.StatusOK, httpserver.ChangePasswordTokenResponse{
		ChangeToken: output.ChangeToken,
		ExpiresIn:   output.ExpiresIn,
	})
}

// AuthPasswordChangeComplete - обработчик установки нового пароля по change token
/*
	1. BearerAuth middleware уже положил identity_id и session_id в context.
	2. Распарсить JSON body в httpserver.CompletePasswordChangeRequest.
	3. Вызвать validation.ValidateCompletePasswordChangeRequest: проверить change_token и new_password.
	4. Собрать authservice.CompletePasswordChangeInput.
	5. Вызвать AuthService.CompletePasswordChange.
	6. Вернуть 200 OK с httpserver.MessageResponse.
	7. Ошибки bind/validation/service перевести в единый ErrorResponse.
*/
func (h *AuthHandlers) AuthPasswordChangeComplete(ctx echo.Context) error {
	// 1. BearerAuth middleware уже положил identity_id и session_id в context.

	// 2. Распарсить JSON body в httpserver.CompletePasswordChangeRequest.
	var req httpserver.CompletePasswordChangeRequest
	if err := ctx.Bind(&req); err != nil {
		return echo.ErrBadRequest
	}

	// 3. Вызвать validation.ValidateCompletePasswordChangeRequest: проверить change_token и new_password.
	changeToken, newPassword, err := validation.ValidateCompletePasswordChangeRequest(req.ChangeToken, req.NewPassword)
	if err != nil {
		return validationError(ctx, err)
	}

	// 4. Собрать authservice.CompletePasswordChangeInput.
	input := authservice.CompletePasswordChangeInput{
		ChangeToken: changeToken,
		NewPassword: newPassword,
	}

	// 5. Вызвать AuthService.CompletePasswordChange.
	message, err := h.as.CompletePasswordChange(ctx.Request().Context(), &input)
	if err != nil {
		// 7. Ошибки bind/validation/service перевести в единый ErrorResponse.
		return serviceError(ctx, err)
	}

	// 6. Вернуть 200 OK с httpserver.MessageResponse.
	return ctx.JSON(http.StatusOK, httpserver.MessageResponse{
		Message: message,
	})
}

// AuthPasswordChangeCodeResend - обработчик повторной отправки кода смены пароля
/*
	1. BearerAuth middleware уже положил identity_id и session_id в context.
	2. Вызвать AuthService.ChangePasswordCodeResend.
	3. Вернуть 200 OK с httpserver.MessageResponse.
	4. Ошибки service перевести в единый ErrorResponse.
*/
func (h *AuthHandlers) AuthPasswordChangeCodeResend(ctx echo.Context) error {
	// 1. BearerAuth middleware уже положил identity_id и session_id в context.

	// 2. Вызвать AuthService.ChangePasswordCodeResend.
	message, err := h.as.ChangePasswordCodeResend(ctx.Request().Context())
	if err != nil {
		// 4. Ошибки service перевести в единый ErrorResponse.
		return serviceError(ctx, err)
	}

	// 3. Вернуть 200 OK с httpserver.MessageResponse.
	return ctx.JSON(http.StatusOK, httpserver.MessageResponse{
		Message: message,
	})
}
