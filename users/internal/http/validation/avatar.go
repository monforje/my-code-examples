package validation

const maxAvatarSize = 5 * 1024 * 1024 // 5 МБ

var allowedContentTypes = map[string]bool{
	"image/jpeg": true,
	"image/png":  true,
	"image/webp": true,
}

// ValidateAvatarFile — валидация загружаемого файла аватара.
/*
	1. Проверить, что размер файла не превышает 5 МБ.
	2. Проверить, что Content-Type допустимый (image/jpeg, image/png, image/webp).
	3. Вернуть nil или ошибку.
*/
func ValidateAvatarFile(contentType string, size int64) error {
	if size > maxAvatarSize {
		return ErrAvatarTooLarge
	}
	if !allowedContentTypes[contentType] {
		return ErrAvatarInvalidType
	}
	return nil
}
