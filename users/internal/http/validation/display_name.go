package validation

import "strings"

const maxDisplayNameLen = 50

// DisplayName — тип для валидации имени пользователя.
type DisplayName string

// Validate — проверяет и нормализует имя пользователя.
/*
	1. Убрать пробелы по краям.
	2. Проверить, что строка не пустая.
	3. Проверить, что длина не превышает 50 символов.
	4. Вернуть нормализованное значение.
*/
func (d DisplayName) Validate() (DisplayName, error) {
	s := strings.TrimSpace(string(d))
	if s == "" {
		return "", ErrDisplayNameEmpty
	}
	if len(s) > maxDisplayNameLen {
		return "", ErrDisplayNameTooLong
	}
	return DisplayName(s), nil
}
