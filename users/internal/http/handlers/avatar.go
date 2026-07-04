package handlers

import (
	"errors"
	"mime"
	"net/http"
	"path/filepath"

	httpserver "users/internal/http/gen"
	"users/internal/http/validation"
	service "users/internal/services"

	"github.com/labstack/echo/v4"
)

// UsersProfileMeAvatarUpdate — обновление аватара текущего пользователя.
/*
	1. BearerAuth middleware уже положил identity_id в context.
	2. Распарсить multipart form, достать файл "avatar".
	3. Определить Content-Type файла.
	4. Вызвать validation.ValidateAvatarFile(contentType, fileSize).
	5. Если ошибка — вернуть 413/AVATAR_TOO_LARGE или 422/INVALID_AVATAR_FORMAT.
	6. Собрать service.UpdateAvatarInput: filename, file reader, file size.
	7. Вызвать UsersService.UpdateAvatar с контекстом запроса.
	8. Сервис сам извлечёт identity_id из контекста через authctx.
	9. Смаппить service.UpdateAvatarOutput в httpserver.UpdateAvatarResponse.
	10. Вернуть 200 OK с avatar_url и updated_at.
	11. Ошибки service перевести в единый ErrorResponse.
*/
func (h *UsersHandlers) UsersProfileMeAvatarUpdate(ctx echo.Context) error {
	// 1. Достаём файл из multipart form.
	fileHeader, err := ctx.FormFile("avatar")
	if err != nil {
		return errorResponse(ctx, http.StatusBadRequest, httpserver.INVALIDJSON, err)
	}

	// 2. Открываем файл для чтения.
	file, err := fileHeader.Open()
	if err != nil {
		return errorResponse(ctx, http.StatusInternalServerError, httpserver.INTERNALERROR, err)
	}
	defer file.Close()

	// 3. Определяем Content-Type по расширению файла.
	contentType := mime.TypeByExtension(filepath.Ext(fileHeader.Filename))
	if contentType == "" {
		contentType = fileHeader.Header.Get("Content-Type")
	}

	// 4. Валидация файла (размер ≤5MB, формат JPEG/PNG/WebP).
		if err := validation.ValidateAvatarFile(contentType, fileHeader.Size); err != nil {
		switch {
		case errors.Is(err, validation.ErrAvatarTooLarge):
			return errorResponse(ctx, http.StatusRequestEntityTooLarge, httpserver.AVATARTOOLARGE, err)
		case errors.Is(err, validation.ErrAvatarInvalidType):
			return errorResponse(ctx, http.StatusUnprocessableEntity, httpserver.INVALIDAVATARFORMAT, err)
		default:
			return validationError(ctx, err)
		}
	}

	// 5. Собираем входной DTO.
	input := &service.UpdateAvatarInput{
		Filename: fileHeader.Filename,
		File:     file,
		FileSize: fileHeader.Size,
	}

	// 6. Вызов сервиса.
	out, err := h.us.UpdateAvatar(ctx.Request().Context(), input)
	if err != nil {
		return serviceError(ctx, err)
	}

	// 7. Маппинг сервисного DTO в HTTP-ответ.
	resp := httpserver.UpdateAvatarResponse{
		AvatarUrl: out.AvatarURL,
		UpdatedAt: out.UpdatedAt,
	}

	return ctx.JSON(http.StatusOK, resp)
}

// UsersProfileMeAvatarDelete — удаление аватара текущего пользователя.
/*
	1. BearerAuth middleware уже положил identity_id в context.
	2. Вызвать UsersService.DeleteAvatar с контекстом запроса.
	3. Сервис сам извлечёт identity_id из контекста через authctx.
	4. Смаппить service.DeleteAvatarOutput в httpserver.DeleteAvatarResponse.
	5. Вернуть 200 OK с avatar_url (null) и updated_at.
	6. Ошибки service перевести в единый ErrorResponse.
*/
func (h *UsersHandlers) UsersProfileMeAvatarDelete(ctx echo.Context) error {
	// 1. Вызов сервиса — identity_id извлекается из контекста внутри сервиса.
	out, err := h.us.DeleteAvatar(ctx.Request().Context())
	if err != nil {
		return serviceError(ctx, err)
	}

	// 2. Маппинг сервисного DTO в HTTP-ответ.
	resp := httpserver.DeleteAvatarResponse{
		AvatarUrl: out.AvatarURL,
		UpdatedAt: out.UpdatedAt,
	}

	return ctx.JSON(http.StatusOK, resp)
}
