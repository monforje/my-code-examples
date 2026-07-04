package auth

// DeviceStartResponse — ответ POST /auth/device/start.
type DeviceStartResponse struct {
	DeviceCode      string `json:"device_code"`
	UserCode        string `json:"user_code"`
	VerificationURL string `json:"verification_url"`
	ExpiresIn       int    `json:"expires_in"`
	Interval        int    `json:"interval"`
}

// DeviceTokenRequest — тело POST /auth/device/token.
type DeviceTokenRequest struct {
	DeviceCode string `json:"device_code"`
}

// CliTokenResponse — успешный ответ POST /auth/device/token.
type CliTokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
	TokenType    string `json:"token_type"`
}

// CliRefreshRequest — тело POST /auth/cli/refresh.
type CliRefreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

// CliRefreshResponse — успешный ответ POST /auth/cli/refresh.
type CliRefreshResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
	TokenType    string `json:"token_type"`
}

// ErrorResponse соответствует контракту ErrorResponse из TypeSpec.
type ErrorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}
