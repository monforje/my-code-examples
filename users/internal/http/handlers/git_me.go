package handlers

import (
	"net/http"

	httpserver "users/internal/http/gen"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// UsersGitMeGet — получить информацию о текущем пользователе, связанном с git-аккаунтом.
/*
	1. Распарсить JSON body с identity_id.
	2. Вызвать UsersService.GetGitMe с identityID.
	3. Смаппить service.GitMeOutput в httpserver.GitMeResponse.
	4. Вернуть 200 OK с данными git-пользователя.
	5. Ошибки service перевести в единый ErrorResponse.
*/
func (h *UsersHandlers) UsersGitMeGet(ctx echo.Context) error {
	// 1. Bind JSON body.
	var req httpserver.GitMeRequest
	if err := ctx.Bind(&req); err != nil {
		return errorResponse(ctx, http.StatusBadRequest, httpserver.INVALIDJSON, err)
	}

	// 2. Парсим identity_id из строки в uuid.
	identityID, err := uuid.Parse(req.IdentityId)
	if err != nil {
		return errorResponse(ctx, http.StatusBadRequest, httpserver.VALIDATIONERROR, err)
	}

	// 3. Вызываем сервис.
	out, err := h.us.GetGitMe(ctx.Request().Context(), identityID)
	if err != nil {
		return serviceError(ctx, err)
	}

	// 4. Маппим сервисный DTO в HTTP-ответ.
	resp := httpserver.GitMeResponse{
		Username: out.Username,
		GitToken: out.GitToken,
		GitUrl:   out.GitURL,
	}

	return ctx.JSON(http.StatusOK, resp)
}
