package validation

import (
	"strings"
)

type Code string

// Validate - проверяет код верификации
/*
	1. Убрать пробелы по краям.
	2. Проверить, что код не пустой.
*/
func (c Code) Validate() (Code, error) {
	code := strings.TrimSpace(string(c))

	if code == "" {
		return "", ErrCodeEmpty
	}

	return Code(code), nil
}
