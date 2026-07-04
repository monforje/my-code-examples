package validation

import (
	"net"
	"net/mail"
	"strings"
)

type Email string

// Validate - проверяет и валидирует email
/*
	1. Убрать пробелы по краям.
	2. Проверить, что строка не пустая.
	3. Проверить базовый формат email через mail.ParseAddress.
	4. Убедиться, что ParseAddress не принял строку вида "Name <a@b.com>".
	5. Достать домен после @.
	6. Проверить, что у домена есть MX-записи.
*/
func (e Email) Validate() (Email, error) {
	email := strings.TrimSpace(string(e))
	email = strings.ToLower(email)

	if email == "" {
		return "", ErrEmailEmpty
	}

	addr, err := mail.ParseAddress(email)
	if err != nil {
		return "", ErrEmailInvalidFormat
	}

	if addr.Address != email {
		return "", ErrEmailInvalidFormat
	}

	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return "", ErrEmailInvalidFormat
	}

	domain := parts[1]

	mx, err := net.LookupMX(domain)
	if err != nil || len(mx) == 0 {
		return "", ErrEmailDomainNoMX
	}

	return Email(email), nil
}
