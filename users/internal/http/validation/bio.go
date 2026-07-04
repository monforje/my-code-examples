package validation

import "strings"

const maxBioLen = 500

// Bio — тип для валидации биографии пользователя.
type Bio string

// Validate — проверяет и нормализует биографию.
/*
	1. Убрать пробелы по краям.
	2. Если строка пустая — допустимо (bio опционально).
	3. Проверить, что длина не превышает 500 символов.
	4. Вернуть нормализованное значение.
*/
func (b Bio) Validate() (Bio, error) {
	s := strings.TrimSpace(string(b))
	if s == "" {
		return "", nil
	}
	if len(s) > maxBioLen {
		return "", ErrBioTooLong
	}
	return Bio(s), nil
}
