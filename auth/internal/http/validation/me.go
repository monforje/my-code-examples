package validation

func ValidateDeleteAccountRequest(password string) (string, error) {
	normalizedPassword, err := Password(password).Validate()
	if err != nil {
		return "", err
	}
	return string(normalizedPassword), nil
}
