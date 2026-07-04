package validation

func ValidateForgotPasswordRequest(email string) (string, error) {
	normalizedEmail, err := Email(email).Validate()
	if err != nil {
		return "", err
	}
	return string(normalizedEmail), nil
}

func ValidateForgotVerifyRequest(email, code string) (string, string, error) {
	normalizedEmail, err := Email(email).Validate()
	if err != nil {
		return "", "", err
	}
	normalizedCode, err := Code(code).Validate()
	if err != nil {
		return "", "", err
	}
	return string(normalizedEmail), string(normalizedCode), nil
}

func ValidateResetPasswordRequest(resetToken, newPassword string) (string, string, error) {
	normalizedToken, err := Token(resetToken).Validate()
	if err != nil {
		return "", "", err
	}
	normalizedPassword, err := Password(newPassword).Validate()
	if err != nil {
		return "", "", err
	}
	return string(normalizedToken), string(normalizedPassword), nil
}
