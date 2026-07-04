package handlers

import (
	"net/http"

	httpserver "auth/internal/http/gen"
	"auth/internal/http/validation"
	authservice "auth/internal/services/auth"

	"github.com/labstack/echo/v4"
)

// AuthRegister - обработчик регистрации пользователя
/*
	1. Распарсить JSON body в httpserver.RegisterRequest.
	2. Вызвать validation.ValidateRegisterRequest: проверить email и password.
	3. Собрать authservice.RegisterInput.
	4. Вызвать AuthService.Register.
	5. Смаппить сервисный RegisterOutput в httpserver.RegisterResponse.
	6. Вернуть 201 Created с identity_id, email и status=pending_verification.
	7. Ошибки bind/validation/service перевести в единый ErrorResponse.
*/
func (h *AuthHandlers) AuthRegister(ctx echo.Context) error {
	// 1. Распарсить JSON body в httpserver.RegisterRequest.
	var req httpserver.RegisterRequest
	if err := ctx.Bind(&req); err != nil {
		return echo.ErrBadRequest
	}

	// 2. Вызвать validation.ValidateRegisterRequest: проверить email и password.
	email, password, err := validation.ValidateRegisterRequest(req.Email, req.Password)
	if err != nil {
		return validationError(ctx, err)
	}

	// 3. Собрать authservice.RegisterInput.
	input := authservice.RegisterInput{
		Email:    email,
		Password: password,
	}

	// 4. Вызвать AuthService.Register.
	output, err := h.as.Register(ctx.Request().Context(), &input)
	if err != nil {
		return serviceError(ctx, err)
	}

	// 5. Смаппить сервисный RegisterOutput в httpserver.RegisterResponse.
	return ctx.JSON(http.StatusCreated, httpserver.RegisterResponse{
		IdentityId: output.IdentityID,
		Email:      output.Email,
		Status:     httpserver.RegisterResponseStatusPendingVerification,
	})
}

// AuthRegisterVerify - обработчик подтверждения регистрации кодом
/*
	1. Распарсить JSON body в httpserver.VerifyCodeRequest.
	2. Вызвать validation.ValidateVerifyCodeRequest: проверить email и code.
	3. Собрать authservice.VerifyCodeInput.
	4. Вызвать AuthService.RegisterVerify.
	5. Вернуть 200 OK с httpserver.MessageResponse.
	6. Ошибки bind/validation/service перевести в единый ErrorResponse.
*/
func (h *AuthHandlers) AuthRegisterVerify(ctx echo.Context) error {
	// 1. Распарсить JSON body в httpserver.VerifyCodeRequest.
	var req httpserver.VerifyCodeRequest
	if err := ctx.Bind(&req); err != nil {
		return echo.ErrBadRequest
	}

	// 2. Вызвать validation.ValidateVerifyCodeRequest: проверить email и code.
	email, code, err := validation.ValidateVerifyCodeRequest(req.Email, req.Code)
	if err != nil {
		return validationError(ctx, err)
	}

	// 3. Собрать authservice.VerifyCodeInput.
	input := authservice.VerifyCodeInput{
		Email: email,
		Code:  code,
	}

	// 4. Вызвать AuthService.RegisterVerify.
	message, err := h.as.RegisterVerify(ctx.Request().Context(), &input)
	if err != nil {
		return serviceError(ctx, err)
	}

	// 5. Вернуть 200 OK с httpserver.MessageResponse.
	return ctx.JSON(http.StatusOK, httpserver.MessageResponse{
		Message: message,
	})
}

// AuthRegisterCodeResend - обработчик повторной отправки кода регистрации
/*
	1. Распарсить JSON body в httpserver.ResendCodeRequest.
	2. Вызвать validation.ValidateResendCodeRequest: проверить email.
	3. Собрать authservice.ResendCodeInput.
	4. Вызвать AuthService.ResendVerificationCode.
	5. Вернуть 200 OK с httpserver.MessageResponse.
	6. Ошибки bind/validation/service перевести в единый ErrorResponse.
*/
func (h *AuthHandlers) AuthRegisterCodeResend(ctx echo.Context) error {
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

	// 4. Вызвать AuthService.ResendVerificationCode.
	message, err := h.as.ResendVerificationCode(ctx.Request().Context(), &input)
	if err != nil {
		return serviceError(ctx, err)
	}

	// 5. Вернуть 200 OK с httpserver.MessageResponse.
	return ctx.JSON(http.StatusOK, httpserver.MessageResponse{
		Message: message,
	})
}
