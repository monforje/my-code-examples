package handlers

import (
	"net/http"
	httpserver "users/internal/http/gen"

	"github.com/labstack/echo/v4"
)

// UsersProfileMeGet — получение информации о текущем пользователе.
/*
	1. BearerAuth middleware уже положил identity_id в context.
	2. Вызвать UsersService.GetProfile с контекстом запроса.
	3. Сервис сам извлечёт identity_id из контекста через authctx.
	4. Смаппить service.ProfileOutput в httpserver.ProfileResponse.
	5. Вернуть 200 OK с данными профиля.
	6. Ошибки service перевести в единый ErrorResponse.
*/
func (h *UsersHandlers) UsersProfileMeGet(ctx echo.Context) error {
	// 1. Вызываем сервис — identity_id извлекается из контекста внутри сервиса.
	out, err := h.us.GetProfile(ctx.Request().Context())
	if err != nil {
		return serviceError(ctx, err)
	}

	// 2. Маппим сервисный DTO в HTTP-ответ.
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
