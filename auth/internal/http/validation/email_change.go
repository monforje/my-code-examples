package validation

import "errors"

var ErrStepInvalid = errors.New("step must be 'current' or 'new'")

func ValidateEmailChangeRequest(password string) (string, error) {
	normalizedPassword, err := Password(password).Validate()
	if err != nil {
		return "", err
	}
	return string(normalizedPassword), nil
}

func ValidateEmailChangeVerifyRequest(code string) (string, error) {
	normalizedCode, err := Code(code).Validate()
	if err != nil {
		return "", err
	}
	return string(normalizedCode), nil
}

func ValidateConfirmEmailChangeRequest(newEmail, identityToken string) (string, string, error) {
	normalizedEmail, err := Email(newEmail).Validate()
	if err != nil {
		return "", "", err
	}
	normalizedToken, err := Token(identityToken).Validate()
	if err != nil {
		return "", "", err
	}
	return string(normalizedEmail), string(normalizedToken), nil
}

func ValidateCompleteEmailChangeRequest(code string) (string, error) {
	normalizedCode, err := Code(code).Validate()
	if err != nil {
		return "", err
	}
	return string(normalizedCode), nil
}

func ValidateEmailChangeCodeResendStep(step string) (string, error) {
	switch step {
	case "current", "new":
		return step, nil
	default:
		return "", ErrStepInvalid
	}
}
