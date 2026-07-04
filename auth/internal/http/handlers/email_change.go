package handlers

import (
	"net/http"

	httpserver "auth/internal/http/gen"
	"auth/internal/http/validation"
	authservice "auth/internal/services/auth"

	"github.com/labstack/echo/v4"
)

// AuthMeEmailChange - обработчик запуска смены email (шаг 1)
/*
	1. BearerAuth middleware уже положил identity_id и session_id в context.
	2. Распарсить JSON body в httpserver.EmailChangeRequest.
	3. Вызвать validation.ValidateEmailChangeRequest: проверить password.
	4. Собрать authservice.ChangeEmailInput.
	5. Вызвать AuthService.ChangeEmail.
	6. Вернуть 200 OK с httpserver.MessageResponse.
	7. Ошибки bind/validation/service перевести в единый ErrorResponse.
*/
func (h *AuthHandlers) AuthMeEmailChange(ctx echo.Context) error {
	// 1. BearerAuth middleware уже положил identity_id и session_id в context.

	// 2. Распарсить JSON body в httpserver.EmailChangeRequest.
	var req httpserver.EmailChangeRequest
	if err := ctx.Bind(&req); err != nil {
		return echo.ErrBadRequest
	}

	// 3. Вызвать validation.ValidateEmailChangeRequest: проверить password.
	password, err := validation.ValidateEmailChangeRequest(req.Password)
	if err != nil {
		return validationError(ctx, err)
	}

	// 4. Собрать authservice.ChangeEmailInput.
	input := authservice.ChangeEmailInput{
		Password: password,
	}

	// 5. Вызвать AuthService.ChangeEmail.
	message, err := h.as.ChangeEmail(ctx.Request().Context(), &input)
	if err != nil {
		// 7. Ошибки bind/validation/service перевести в единый ErrorResponse.
		return serviceError(ctx, err)
	}

	// 6. Вернуть 200 OK с httpserver.MessageResponse.
	return ctx.JSON(http.StatusOK, httpserver.MessageResponse{
		Message: message,
	})
}

// AuthMeEmailChangeVerify - обработчик подтверждения текущего email кодом (шаг 2)
/*
	1. BearerAuth middleware уже положил identity_id и session_id в context.
	2. Распарсить JSON body в httpserver.CodeVerifyRequest.
	3. Вызвать validation.ValidateEmailChangeVerifyRequest: проверить code.
	4. Собрать authservice.ChangeEmailVerifyInput.
	5. Вызвать AuthService.ChangeEmailVerify.
	6. Смаппить сервисный ChangeEmailVerifyOutput в httpserver.IdentityTokenResponse.
	7. Вернуть 200 OK с identity_token и expires_in.
	8. Ошибки bind/validation/service перевести в единый ErrorResponse.
*/
func (h *AuthHandlers) AuthMeEmailChangeVerify(ctx echo.Context) error {
	// 1. BearerAuth middleware уже положил identity_id и session_id в context.

	// 2. Распарсить JSON body в httpserver.CodeVerifyRequest.
	var req httpserver.CodeVerifyRequest
	if err := ctx.Bind(&req); err != nil {
		return echo.ErrBadRequest
	}

	// 3. Вызвать validation.ValidateEmailChangeVerifyRequest: проверить code.
	code, err := validation.ValidateEmailChangeVerifyRequest(req.Code)
	if err != nil {
		return validationError(ctx, err)
	}

	// 4. Собрать authservice.ChangeEmailVerifyInput.
	input := authservice.ChangeEmailVerifyInput{
		Code: code,
	}

	// 5. Вызвать AuthService.ChangeEmailVerify.
	output, err := h.as.ChangeEmailVerify(ctx.Request().Context(), &input)
	if err != nil {
		// 8. Ошибки bind/validation/service перевести в единый ErrorResponse.
		return serviceError(ctx, err)
	}

	// 6. Смаппить сервисный ChangeEmailVerifyOutput в httpserver.IdentityTokenResponse.
	// 7. Вернуть 200 OK с identity_token и expires_in.
	return ctx.JSON(http.StatusOK, httpserver.IdentityTokenResponse{
		IdentityToken: output.IdentityToken,
		ExpiresIn:     output.ExpiresIn,
	})
}

// AuthMeEmailChangeConfirm - обработчик подтверждения нового email (шаг 3)
/*
	1. BearerAuth middleware уже положил identity_id и session_id в context.
	2. Распарсить JSON body в httpserver.ConfirmEmailChangeRequest.
	3. Вызвать validation.ValidateConfirmEmailChangeRequest: проверить new_email и identity_token.
	4. Собрать authservice.ChangeEmailConfirmInput.
	5. Вызвать AuthService.ChangeEmailConfirm.
	6. Вернуть 200 OK с httpserver.MessageResponse.
	7. Ошибки bind/validation/service перевести в единый ErrorResponse.
*/
func (h *AuthHandlers) AuthMeEmailChangeConfirm(ctx echo.Context) error {
	// 1. BearerAuth middleware уже положил identity_id и session_id в context.

	// 2. Распарсить JSON body в httpserver.ConfirmEmailChangeRequest.
	var req httpserver.ConfirmEmailChangeRequest
	if err := ctx.Bind(&req); err != nil {
		return echo.ErrBadRequest
	}

	// 3. Вызвать validation.ValidateConfirmEmailChangeRequest: проверить new_email и identity_token.
	newEmail, identityToken, err := validation.ValidateConfirmEmailChangeRequest(req.NewEmail, req.IdentityToken)
	if err != nil {
		return validationError(ctx, err)
	}

	// 4. Собрать authservice.ChangeEmailConfirmInput.
	input := authservice.ChangeEmailConfirmInput{
		NewEmail:      newEmail,
		IdentityToken: identityToken,
	}

	// 5. Вызвать AuthService.ChangeEmailConfirm.
	message, err := h.as.ChangeEmailConfirm(ctx.Request().Context(), &input)
	if err != nil {
		// 7. Ошибки bind/validation/service перевести в единый ErrorResponse.
		return serviceError(ctx, err)
	}

	// 6. Вернуть 200 OK с httpserver.MessageResponse.
	return ctx.JSON(http.StatusOK, httpserver.MessageResponse{
		Message: message,
	})
}

// AuthMeEmailChangeComplete - обработчик завершения смены email по коду с нового email (шаг 4)
/*
	1. BearerAuth middleware уже положил identity_id и session_id в context.
	2. Распарсить JSON body в httpserver.CompleteEmailChangeRequest.
	3. Вызвать validation.ValidateCompleteEmailChangeRequest: проверить code.
	4. Собрать authservice.ChangeEmailCompleteInput.
	5. Вызвать AuthService.ChangeEmailComplete.
	6. Вернуть 200 OK с httpserver.MessageResponse.
	7. Ошибки bind/validation/service перевести в единый ErrorResponse.
*/
func (h *AuthHandlers) AuthMeEmailChangeComplete(ctx echo.Context) error {
	// 1. BearerAuth middleware уже положил identity_id и session_id в context.

	// 2. Распарсить JSON body в httpserver.CompleteEmailChangeRequest.
	var req httpserver.CompleteEmailChangeRequest
	if err := ctx.Bind(&req); err != nil {
		return echo.ErrBadRequest
	}

	// 3. Вызвать validation.ValidateCompleteEmailChangeRequest: проверить code.
	code, err := validation.ValidateCompleteEmailChangeRequest(req.Code)
	if err != nil {
		return validationError(ctx, err)
	}

	// 4. Собрать authservice.ChangeEmailCompleteInput.
	input := authservice.ChangeEmailCompleteInput{
		Code: code,
	}

	// 5. Вызвать AuthService.ChangeEmailComplete.
	message, err := h.as.ChangeEmailComplete(ctx.Request().Context(), &input)
	if err != nil {
		// 7. Ошибки bind/validation/service перевести в единый ErrorResponse.
		return serviceError(ctx, err)
	}

	// 6. Вернуть 200 OK с httpserver.MessageResponse.
	return ctx.JSON(http.StatusOK, httpserver.MessageResponse{
		Message: message,
	})
}

// AuthMeEmailChangeCodeResend - обработчик повторной отправки кода смены email
/*
	1. BearerAuth middleware уже положил identity_id и session_id в context.
	2. Получить params.Step из query-параметра step, который распарсил generated wrapper.
	3. Вызвать validation.ValidateEmailChangeCodeResendStep: проверить step=current|new.
	4. Собрать authservice.ChangeEmailResendInput.
	5. Вызвать AuthService.ChangeEmailCodeResend.
	6. Вернуть 200 OK с httpserver.MessageResponse.
	7. Ошибки validation/service перевести в единый ErrorResponse.
*/
func (h *AuthHandlers) AuthMeEmailChangeCodeResend(ctx echo.Context, params httpserver.AuthMeEmailChangeCodeResendParams) error {
	// 1. BearerAuth middleware уже положил identity_id и session_id в context.

	// 2. Получить params.Step из query-параметра step.
	// 3. Вызвать validation.ValidateEmailChangeCodeResendStep: проверить step=current|new.
	step, err := validation.ValidateEmailChangeCodeResendStep(string(params.Step))
	if err != nil {
		return validationError(ctx, err)
	}

	// 4. Собрать authservice.ChangeEmailResendInput.
	input := authservice.ChangeEmailResendInput{
		Step: step,
	}

	// 5. Вызвать AuthService.ChangeEmailCodeResend.
	message, err := h.as.ChangeEmailCodeResend(ctx.Request().Context(), &input)
	if err != nil {
		// 7. Ошибки validation/service перевести в единый ErrorResponse.
		return serviceError(ctx, err)
	}

	// 6. Вернуть 200 OK с httpserver.MessageResponse.
	return ctx.JSON(http.StatusOK, httpserver.MessageResponse{
		Message: message,
	})
}
