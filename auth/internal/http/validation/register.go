package validation

func ValidateRegisterRequest(email, password string) (string, string, error) {
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

func ValidateVerifyCodeRequest(email, code string) (string, string, error) {
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

func ValidateResendCodeRequest(email string) (string, error) {
	normalizedEmail, err := Email(email).Validate()
	if err != nil {
		return "", err
	}

	return string(normalizedEmail), nil
}
