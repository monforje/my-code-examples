package handlers

import (
	"net/http"
	httpserver "users/internal/http/gen"
	"users/internal/http/validation"
	service "users/internal/services"

	"github.com/labstack/echo/v4"
)

// UsersProfileMeSettingsUpdate — обновление настроек текущего пользователя.
/*
	1. BearerAuth middleware уже положил identity_id в context.
	2. Распарсить JSON body в httpserver.UpdateProfileSettingsRequest.
	3. Извлечь строковые значения из указателей (nil → "").
	4. Вызвать validation.ValidateUpdateSettingsRequest(dn, bio).
	5. Если ошибка валидации — вернуть 422 + VALIDATIONERROR.
	6. Собрать service.UpdateSettingsInput из указателей запроса.
	7. Вызвать UsersService.UpdateSettings с контекстом запроса.
	8. Сервис сам извлечёт identity_id из контекста через authctx.
	9. Смаппить service.ProfileOutput в httpserver.ProfileResponse.
	10. Вернуть 200 OK с обновлёнными данными профиля.
	11. Ошибки bind/validation/service перевести в единый ErrorResponse.
*/
func (h *UsersHandlers) UsersProfileMeSettingsUpdate(ctx echo.Context) error {
	// 1. Bind JSON body.
	var req httpserver.UpdateProfileSettingsRequest
	if err := ctx.Bind(&req); err != nil {
		return errorResponse(ctx, http.StatusBadRequest, httpserver.INVALIDJSON, err)
	}

	// 2. Извлечь строковые значения из указателей для валидации.
	//    При частичном обновлении валидируем только переданные поля.
	if req.DisplayName != nil {
		if _, err := validation.DisplayName(ptrValue(req.DisplayName)).Validate(); err != nil {
			return validationError(ctx, err)
		}
	}
	if req.Bio != nil {
		if _, err := validation.Bio(ptrValue(req.Bio)).Validate(); err != nil {
			return validationError(ctx, err)
		}
	}

	// 3. Собрать входной DTO из указателей запроса.
	input := &service.UpdateSettingsInput{
		DisplayName: req.DisplayName,
		Bio:         req.Bio,
	}

	// 5. Вызов сервиса.
	out, err := h.us.UpdateSettings(ctx.Request().Context(), input)
	if err != nil {
		return serviceError(ctx, err)
	}

	// 6. Маппинг сервисного DTO в HTTP-ответ.
	resp := httpserver.ProfileResponse{
		Id:            out.ID,
		IdentityId:    out.IdentityID,
		Email:         out.Email,
		DisplayName:   out.DisplayName,
		Bio:           &out.Bio,
		AvatarUrl:     out.AvatarURL,
		Status:        out.Status,
		EmailVerified: out.EmailVerified,
		CreatedAt:     out.CreatedAt,
		UpdatedAt:     out.UpdatedAt,
	}

	return ctx.JSON(http.StatusOK, resp)
}

// ptrValue — возвращает значение указателя или "" если указатель nil.
func ptrValue(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
