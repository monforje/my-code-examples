package e2e_test

import (
	"auth/internal/events"
	e2ehelpers "auth/tests/e2e/helpers"
	"net/http"
	"testing"
)

func TestDeviceAuth_FullFlow(t *testing.T) {
	resetE2E(t)

	capture := e2ehelpers.NewEventCapture(e2eEnv.NC())
	defer capture.Close()

	// 1. Register + verify user (to get a Bearer token for browser confirmation)
	user := e2ehelpers.RegisterAndVerify(t, e2eEnv, client)
	tokens := e2ehelpers.Login(t, client, user)
	browserToken := tokens.AccessToken

	// 2. CLI calls /auth/device/start
	startResp := client.PostJSON(t, "/auth/device/start", nil)
	e2ehelpers.ExpectStatus(t, startResp, http.StatusOK)
	start := e2ehelpers.Decode[e2ehelpers.DeviceStartResponse](t, startResp)

	if start.DeviceCode == "" {
		t.Fatal("device_code is empty")
	}
	if start.UserCode == "" {
		t.Fatal("user_code is empty")
	}
	if start.VerificationUrl == "" {
		t.Fatal("verification_url is empty")
	}
	if start.ExpiresIn <= 0 {
		t.Fatalf("expires_in = %d, want positive", start.ExpiresIn)
	}
	if start.Interval <= 0 {
		t.Fatalf("interval = %d, want positive", start.Interval)
	}

	// 3. Browser confirms with Bearer token + user_code
	confirmResp := client.PostAuthJSON(t, browserToken, "/auth/device/confirm", map[string]string{
		"user_code": start.UserCode,
	})
	e2ehelpers.ExpectStatus(t, confirmResp, http.StatusOK)
	confirm := e2ehelpers.Decode[e2ehelpers.DeviceConfirmResponse](t, confirmResp)

	if confirm.Status != "confirmed" {
		t.Fatalf("status = %q, want confirmed", confirm.Status)
	}

	// 4. CLI polls /auth/device/token with device_code
	tokenResp := client.PostJSON(t, "/auth/device/token", map[string]string{
		"device_code": start.DeviceCode,
	})
	e2ehelpers.ExpectStatus(t, tokenResp, http.StatusOK)
	cliTokens := e2ehelpers.Decode[e2ehelpers.CliTokenResponse](t, tokenResp)

	if cliTokens.AccessToken == "" {
		t.Fatal("access_token is empty")
	}
	if cliTokens.RefreshToken == "" {
		t.Fatal("refresh_token is empty")
	}
	if cliTokens.ExpiresIn <= 0 {
		t.Fatalf("expires_in = %d, want positive", cliTokens.ExpiresIn)
	}
	if cliTokens.TokenType != "Bearer" {
		t.Fatalf("token_type = %q, want Bearer", cliTokens.TokenType)
	}

	// 5. Verify session was created in DB (cli session with user_agent=cli)
	identity := e2ehelpers.GetIdentityByEmail(t, e2eEnv.PgPool(), user.Email)
	sessions := e2ehelpers.GetSessionsByIdentityID(t, e2eEnv.PgPool(), identity.ID)
	cliSessions := 0
	for _, s := range sessions {
		if s.UserAgent == "cli" {
			cliSessions++
		}
	}
	if cliSessions != 1 {
		t.Fatalf("db cli sessions count = %d, want 1", cliSessions)
	}

	// 6. Verify auth_event was created
	deviceEvents := e2ehelpers.GetAuthEventsByType(t, e2eEnv.PgPool(), identity.ID, "device_login")
	if len(deviceEvents) != 1 {
		t.Fatalf("db device_login events count = %d, want 1", len(deviceEvents))
	}

	// 7. CLI refreshes token
	refreshResp := client.PostJSON(t, "/auth/cli/refresh", map[string]string{
		"refresh_token": cliTokens.RefreshToken,
	})
	e2ehelpers.ExpectStatus(t, refreshResp, http.StatusOK)
	refreshed := e2ehelpers.Decode[e2ehelpers.CliRefreshResponse](t, refreshResp)

	if refreshed.AccessToken == "" {
		t.Fatal("refreshed access_token is empty")
	}
	if refreshed.RefreshToken == "" {
		t.Fatal("refreshed refresh_token is empty")
	}
	if refreshed.ExpiresIn <= 0 {
		t.Fatalf("refreshed expires_in = %d, want positive", refreshed.ExpiresIn)
	}

	capture.AssertPublished(t, events.EventIdentityLogin)
}

func TestDeviceAuth_PollBeforeConfirm(t *testing.T) {
	resetE2E(t)

	// 1. CLI starts device auth
	startResp := client.PostJSON(t, "/auth/device/start", nil)
	e2ehelpers.ExpectStatus(t, startResp, http.StatusOK)
	start := e2ehelpers.Decode[e2ehelpers.DeviceStartResponse](t, startResp)

	// 2. CLI polls BEFORE browser confirms → 428 Precondition Required
	tokenResp := client.PostJSON(t, "/auth/device/token", map[string]string{
		"device_code": start.DeviceCode,
	})
	e2ehelpers.ExpectStatus(t, tokenResp, http.StatusPreconditionRequired)
}

func TestDeviceAuth_ConfirmWithWrongCode(t *testing.T) {
	resetE2E(t)

	user := e2ehelpers.RegisterAndVerify(t, e2eEnv, client)
	tokens := e2ehelpers.Login(t, client, user)

	// 1. CLI starts device auth
	startResp := client.PostJSON(t, "/auth/device/start", nil)
	e2ehelpers.ExpectStatus(t, startResp, http.StatusOK)

	// 2. Browser tries to confirm with wrong user_code
	confirmResp := client.PostAuthJSON(t, tokens.AccessToken, "/auth/device/confirm", map[string]string{
		"user_code": "XXXX-XXXX",
	})
	e2ehelpers.ExpectStatus(t, confirmResp, http.StatusNotFound)
}

func TestDeviceAuth_ConfirmWithoutAuth(t *testing.T) {
	resetE2E(t)

	// 1. CLI starts device auth
	startResp := client.PostJSON(t, "/auth/device/start", nil)
	e2ehelpers.ExpectStatus(t, startResp, http.StatusOK)
	start := e2ehelpers.Decode[e2ehelpers.DeviceStartResponse](t, startResp)

	// 2. Browser tries to confirm without Bearer token → 401
	confirmResp := client.PostJSON(t, "/auth/device/confirm", map[string]string{
		"user_code": start.UserCode,
	})
	e2ehelpers.ExpectStatus(t, confirmResp, http.StatusUnauthorized)
}

func TestDeviceAuth_TokenWithInvalidDeviceCode(t *testing.T) {
	resetE2E(t)

	// CLI polls with non-existent device_code → 404
	tokenResp := client.PostJSON(t, "/auth/device/token", map[string]string{
		"device_code": "non-existent-device-code",
	})
	e2ehelpers.ExpectStatus(t, tokenResp, http.StatusNotFound)
}

func TestDeviceAuth_CliRefreshInvalidToken(t *testing.T) {
	resetE2E(t)

	// CLI refreshes with invalid refresh token → 401
	refreshResp := client.PostJSON(t, "/auth/cli/refresh", map[string]string{
		"refresh_token": "invalid-refresh-token",
	})
	e2ehelpers.ExpectStatus(t, refreshResp, http.StatusUnauthorized)
}
