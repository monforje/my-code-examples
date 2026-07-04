package validation

import "strings"

type Title string

// Validate - проверяет заголовок задачи
/*
	1. Убрать пробелы по краям.
	2. Проверить, что заголовок не пустой.
	3. Проверить максимальную длину (255 символов).
*/
func (t Title) Validate() (Title, error) {
	title := strings.TrimSpace(string(t))

	if title == "" {
		return "", ErrTitleEmpty
	}

	if len(title) > 255 {
		return "", ErrTitleTooLong
	}

	return Title(title), nil
}
