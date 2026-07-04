package validation

func ValidateLoginRequest(email, password string) (string, string, error) {
	normalizedEmail, err := Email(email).Validate()
	if err != nil {
		return "", "", err
	}
	normalizedPassword, err := Password(password).Validate()
	if err != nil {
		return "", "", err
	}
	return string(normalizedEmail), string(normalizedPassword), nil
}
