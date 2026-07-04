package validation

// ValidateUpdateSettingsRequest — валидация запроса обновления настроек профиля.
/*
	1. Вызвать DisplayName(displayName).Validate() для проверки имени.
	2. Вызвать Bio(bio).Validate() для проверки биографии.
	3. Вернуть нормализованные строки или ошибку.
*/
func ValidateUpdateSettingsRequest(displayName, bio string) (string, string, error) {
	dn, err := DisplayName(displayName).Validate()
	if err != nil {
		return "", "", err
	}
	b, err := Bio(bio).Validate()
	if err != nil {
		return "", "", err
	}
	return string(dn), string(b), nil
}
