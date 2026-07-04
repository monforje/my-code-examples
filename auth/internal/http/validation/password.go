package validation

import (
	"strings"
	"unicode"
)

type Password string

// Validate - проверяет пароль и возвращает нормализованное значение
/*
	1. Убрать пробелы по краям.
	2. Проверить, что пароль не пустой.
	3. Проверить минимальную длину.
	4. Проверить максимальную длину.
	5. Проверить, что в пароле есть хотя бы одна буква и одна цифра.
	6. Вернуть пароль как Password.
*/
func (p Password) Validate() (Password, error) {
	password := strings.TrimSpace(string(p))

	if password == "" {
		return "", ErrPasswordEmpty
	}

	if len(password) < 8 {
		return "", ErrPasswordTooShort
	}

	if len(password) > 72 {
		return "", ErrPasswordTooLong
	}

	hasLetter := false
	hasDigit := false

	for _, ch := range password {
		if unicode.IsLetter(ch) {
			hasLetter = true
		}

		if unicode.IsDigit(ch) {
			hasDigit = true
		}
	}

	if !hasLetter {
		return "", ErrPasswordNoLetter
	}

	if !hasDigit {
		return "", ErrPasswordNoDigit
	}

	return Password(password), nil
}
