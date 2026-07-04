package handlers

import (
	"net/http"

	httpserver "auth/internal/http/gen"
	"auth/internal/http/validation"
	authservice "auth/internal/services/auth"

	"github.com/labstack/echo/v4"
)

// AuthMeGet - обработчик получения текущей учетной записи
/*
	1. BearerAuth middleware уже положил identity_id и session_id в context.
	2. Вызвать AuthService.GetMe.
	3. Смаппить сервисный Identity в httpserver.Identity.
	4. Вернуть 200 OK с данными текущей учетной записи.
	5. Ошибки service перевести в единый ErrorResponse.
*/
func (h *AuthHandlers) AuthMeGet(ctx echo.Context) error {
	// 1. BearerAuth middleware уже положил identity_id и session_id в context.

	// 2. Вызвать AuthService.GetMe.
	identity, err := h.as.GetMe(ctx.Request().Context())
	if err != nil {
		// 5. Ошибки service перевести в единый ErrorResponse.
		return serviceError(ctx, err)
	}

	// 3. Смаппить сервисный Identity в httpserver.Identity.
	// 4. Вернуть 200 OK с данными текущей учетной записи.
	return ctx.JSON(http.StatusOK, httpserver.Identity{
		Id:            identity.ID,
		Email:         identity.Email,
		EmailVerified: identity.EmailVerified,
		Status:        httpserver.IdentityStatus(identity.Status),
		CreatedAt:     identity.CreatedAt,
	})
}

// AuthMeDelete - обработчик запуска удаления аккаунта
/*
	1. BearerAuth middleware уже положил identity_id и session_id в context.
	2. Распарсить JSON body в httpserver.DeleteAccountRequest.
	3. Вызвать validation.ValidateDeleteAccountRequest: проверить password.
	4. Собрать authservice.DeleteAccountInput.
	5. Вызвать AuthService.DeleteAccount.
	6. Вернуть 200 OK с httpserver.MessageResponse.
	7. Ошибки bind/validation/service перевести в единый ErrorResponse.
*/
func (h *AuthHandlers) AuthMeDelete(ctx echo.Context) error {
	// 1. BearerAuth middleware уже положил identity_id и session_id в context.

	// 2. Распарсить JSON body в httpserver.DeleteAccountRequest.
	var req httpserver.DeleteAccountRequest
	if err := ctx.Bind(&req); err != nil {
		return echo.ErrBadRequest
	}

	// 3. Вызвать validation.ValidateDeleteAccountRequest: проверить password.
	password, err := validation.ValidateDeleteAccountRequest(req.Password)
	if err != nil {
		return validationError(ctx, err)
	}

	// 4. Собрать authservice.DeleteAccountInput.
	input := authservice.DeleteAccountInput{
		Password: password,
	}

	// 5. Вызвать AuthService.DeleteAccount.
	message, err := h.as.DeleteAccount(ctx.Request().Context(), &input)
	if err != nil {
		// 7. Ошибки bind/validation/service перевести в единый ErrorResponse.
		return serviceError(ctx, err)
	}

	// 6. Вернуть 200 OK с httpserver.MessageResponse.
	return ctx.JSON(http.StatusOK, httpserver.MessageResponse{
		Message: message,
	})
}

// AuthMeDeleteVerify - обработчик подтверждения удаления аккаунта кодом
/*
	1. BearerAuth middleware уже положил identity_id и session_id в context.
	2. Распарсить JSON body в httpserver.CodeVerifyRequest.
	3. Вызвать validation.ValidateCodeVerifyRequest: проверить code.
	4. Собрать authservice.DeleteAccountVerifyInput.
	5. Вызвать AuthService.DeleteAccountVerify.
	6. Вернуть 204 No Content.
	7. Ошибки bind/validation/service перевести в единый ErrorResponse.
*/
func (h *AuthHandlers) AuthMeDeleteVerify(ctx echo.Context) error {
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

	// 4. Собрать authservice.DeleteAccountVerifyInput.
	input := authservice.DeleteAccountVerifyInput{
		Code: code,
	}

	// 5. Вызвать AuthService.DeleteAccountVerify.
	if err := h.as.DeleteAccountVerify(ctx.Request().Context(), &input); err != nil {
		// 7. Ошибки bind/validation/service перевести в единый ErrorResponse.
		return serviceError(ctx, err)
	}

	// 6. Вернуть 204 No Content.
	return ctx.NoContent(http.StatusNoContent)
}

// AuthMeDeleteCodeResend - обработчик повторной отправки кода удаления аккаунта
/*
	1. BearerAuth middleware уже положил identity_id и session_id в context.
	2. Вызвать AuthService.DeleteAccountCodeResend.
	3. Вернуть 200 OK с httpserver.MessageResponse.
	4. Ошибки service перевести в единый ErrorResponse.
*/
func (h *AuthHandlers) AuthMeDeleteCodeResend(ctx echo.Context) error {
	// 1. BearerAuth middleware уже положил identity_id и session_id в context.

	// 2. Вызвать AuthService.DeleteAccountCodeResend.
	message, err := h.as.DeleteAccountCodeResend(ctx.Request().Context())
	if err != nil {
		// 4. Ошибки service перевести в единый ErrorResponse.
		return serviceError(ctx, err)
	}

	// 3. Вернуть 200 OK с httpserver.MessageResponse.
	return ctx.JSON(http.StatusOK, httpserver.MessageResponse{
		Message: message,
	})
}
