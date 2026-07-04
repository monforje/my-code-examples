package auth

import (
	"strings"
	"unicode"
)

// FormatUserCode нормализует user_code из ответа API и форматирует как XXXX-XXXX:
// оставляет только буквенно-цифровые символы, приводит к верхнему регистру,
// вставляет дефис после 4-го символа.
func FormatUserCode(raw string) string {
	var b strings.Builder
	b.Grow(len(raw))
	for _, r := range raw {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			b.WriteRune(unicode.ToUpper(r))
		}
	}
	s := b.String()
	if len(s) > 4 {
		return s[:4] + "-" + s[4:]
	}
	return s
}
