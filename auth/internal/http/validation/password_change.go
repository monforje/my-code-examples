package validation

func ValidateChangePasswordRequest(currentPassword string) (string, error) {
	normalizedPassword, err := Password(currentPassword).Validate()
	if err != nil {
		return "", err
	}
	return string(normalizedPassword), nil
}

func ValidateCodeVerifyRequest(code string) (string, error) {
	normalizedCode, err := Code(code).Validate()
	if err != nil {
		return "", err
	}
	return string(normalizedCode), nil
}

func ValidateCompletePasswordChangeRequest(changeToken, newPassword string) (string, string, error) {
	normalizedToken, err := Token(changeToken).Validate()
	if err != nil {
		return "", "", err
	}
	normalizedPassword, err := Password(newPassword).Validate()
	if err != nil {
		return "", "", err
	}
	return string(normalizedToken), string(normalizedPassword), nil
}
