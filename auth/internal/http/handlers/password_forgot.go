package handlers

import (
	"net/http"

	httpserver "auth/internal/http/gen"
	"auth/internal/http/validation"
	authservice "auth/internal/services/auth"

	"github.com/labstack/echo/v4"
)

// AuthPasswordForgot - обработчик запуска восстановления пароля
/*
	1. Распарсить JSON body в httpserver.ForgotPasswordRequest.
	2. Вызвать validation.ValidateForgotPasswordRequest: проверить email.
	3. Собрать authservice.ForgotPasswordInput.
	4. Вызвать AuthService.ForgotPassword.
	5. Вернуть 200 OK с httpserver.MessageResponse.
	6. Ошибки bind/validation/service перевести в единый ErrorResponse.
	7. Не раскрывать существование email: нейтральное сообщение формирует сервис.
*/
func (h *AuthHandlers) AuthPasswordForgot(ctx echo.Context) error {
	// 1. Распарсить JSON body в httpserver.ForgotPasswordRequest.
	var req httpserver.ForgotPasswordRequest
	if err := ctx.Bind(&req); err != nil {
		return echo.ErrBadRequest
	}

	// 2. Вызвать validation.ValidateForgotPasswordRequest: проверить email.
	email, err := validation.ValidateForgotPasswordRequest(req.Email)
	if err != nil {
		return validationError(ctx, err)
	}

	// 3. Собрать authservice.ForgotPasswordInput.
	input := authservice.ForgotPasswordInput{
		Email: email,
	}

	// 4. Вызвать AuthService.ForgotPassword.
	message, err := h.as.ForgotPassword(ctx.Request().Context(), &input)
	if err != nil {
		return serviceError(ctx, err)
	}

	// 5. Вернуть 200 OK с httpserver.MessageResponse.
	return ctx.JSON(http.StatusOK, httpserver.MessageResponse{
		Message: message,
	})
}

// AuthPasswordForgotVerify - обработчик проверки кода восстановления пароля
/*
	1. Распарсить JSON body в httpserver.ForgotVerifyRequest.
	2. Вызвать validation.ValidateForgotVerifyRequest: проверить email и code.
	3. Собрать authservice.ForgotPasswordVerifyInput.
	4. Вызвать AuthService.ForgotPasswordVerify.
	5. Смаппить сервисный ResetTokenOutput в httpserver.ResetTokenResponse.
	6. Вернуть 200 OK с reset_token и expires_in.
	7. Ошибки bind/validation/service перевести в единый ErrorResponse.
*/
func (h *AuthHandlers) AuthPasswordForgotVerify(ctx echo.Context) error {
	// 1. Распарсить JSON body в httpserver.ForgotVerifyRequest.
	var req httpserver.ForgotVerifyRequest
	if err := ctx.Bind(&req); err != nil {
		return echo.ErrBadRequest
	}

	// 2. Вызвать validation.ValidateForgotVerifyRequest: проверить email и code.
	email, code, err := validation.ValidateForgotVerifyRequest(req.Email, req.Code)
	if err != nil {
		return validationError(ctx, err)
	}

	// 3. Собрать authservice.ForgotPasswordVerifyInput.
	input := authservice.ForgotPasswordVerifyInput{
		Email: email,
		Code:  code,
	}

	// 4. Вызвать AuthService.ForgotPasswordVerify.
	output, err := h.as.ForgotPasswordVerify(ctx.Request().Context(), &input)
	if err != nil {
		// 7. Ошибки bind/validation/service перевести в единый ErrorResponse.
		return serviceError(ctx, err)
	}

	// 5. Смаппить сервисный ResetTokenOutput в httpserver.ResetTokenResponse.
	// 6. Вернуть 200 OK с reset_token и expires_in.
	return ctx.JSON(http.StatusOK, httpserver.ResetTokenResponse{
		ResetToken: output.ResetToken,
		ExpiresIn:  output.ExpiresIn,
	})
}

// AuthPasswordForgotCodeResend - обработчик повторной отправки кода восстановления пароля
/*
	1. Распарсить JSON body в httpserver.ResendCodeRequest.
	2. Вызвать validation.ValidateResendCodeRequest: проверить email.
	3. Собрать authservice.ResendCodeInput.
	4. Вызвать AuthService.ForgotPasswordCodeResend.
	5. Вернуть 200 OK с httpserver.MessageResponse.
	6. Ошибки bind/validation/service перевести в единый ErrorResponse.
	7. Не раскрывать существование email: нейтральное сообщение формирует сервис.
*/
func (h *AuthHandlers) AuthPasswordForgotCodeResend(ctx echo.Context) error {
	// 1. Распарсить JSON body в httpserver.ResendCodeRequest.
	var req httpserver.ResendCodeRequest
	if err := ctx.Bind(&req); err != nil {
		return echo.ErrBadRequest
	}

	// 2. Вызвать validation.ValidateResendCodeRequest: проверить email.
	email, err := validation.ValidateResendCodeRequest(req.Email)
	if err != nil {
		return validationError(ctx, err)
	}

	// 3. Собрать authservice.ResendCodeInput.
	input := authservice.ResendCodeInput{
		Email: email,
	}

	// 4. Вызвать AuthService.ForgotPasswordCodeResend.
	message, err := h.as.ForgotPasswordCodeResend(ctx.Request().Context(), &input)
	if err != nil {
		// 6. Ошибки bind/validation/service перевести в единый ErrorResponse.
		return serviceError(ctx, err)
	}

	// 5. Вернуть 200 OK с httpserver.MessageResponse.
	return ctx.JSON(http.StatusOK, httpserver.MessageResponse{
		Message: message,
	})
}

// AuthPasswordReset - обработчик установки нового пароля по reset token
/*
	1. Распарсить JSON body в httpserver.ResetPasswordRequest.
	2. Вызвать validation.ValidateResetPasswordRequest: проверить reset_token и new_password.
	3. Собрать authservice.ResetPasswordInput.
	4. Вызвать AuthService.ResetPassword.
	5. Вернуть 200 OK с httpserver.MessageResponse.
	6. Ошибки bind/validation/service перевести в единый ErrorResponse.
*/
func (h *AuthHandlers) AuthPasswordReset(ctx echo.Context) error {
	// 1. Распарсить JSON body в httpserver.ResetPasswordRequest.
	var req httpserver.ResetPasswordRequest
	if err := ctx.Bind(&req); err != nil {
		return echo.ErrBadRequest
	}

	// 2. Вызвать validation.ValidateResetPasswordRequest: проверить reset_token и new_password.
	resetToken, newPassword, err := validation.ValidateResetPasswordRequest(req.ResetToken, req.NewPassword)
	if err != nil {
		return validationError(ctx, err)
	}

	// 3. Собрать authservice.ResetPasswordInput.
	input := authservice.ResetPasswordInput{
		ResetToken:  resetToken,
		NewPassword: newPassword,
	}

	// 4. Вызвать AuthService.ResetPassword.
	message, err := h.as.ResetPassword(ctx.Request().Context(), &input)
	if err != nil {
		// 6. Ошибки bind/validation/service перевести в единый ErrorResponse.
		return serviceError(ctx, err)
	}

	// 5. Вернуть 200 OK с httpserver.MessageResponse.
	return ctx.JSON(http.StatusOK, httpserver.MessageResponse{
		Message: message,
	})
}
