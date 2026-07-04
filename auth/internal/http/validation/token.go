package validation

import "strings"

type Token string

func (t Token) Validate() (Token, error) {
	token := strings.TrimSpace(string(t))
	if token == "" {
		return "", ErrTokenEmpty
	}
	return Token(token), nil
}
