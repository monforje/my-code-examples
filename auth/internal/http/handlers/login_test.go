package handlers_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"go.uber.org/mock/gomock"

	httpserver "auth/internal/http/gen"
	authservice "auth/internal/services/auth"
)

const (
	loginOp   = "auth.login"
	logoutOp  = "auth.logout"
	refreshOp = "auth.refresh"
)

// ──────────────────────────────────────────────
// AuthLogin
// ──────────────────────────────────────────────

func TestLogin_Success(t *testing.T) {
	svc := newMock(t)
	ts := setupServer(t, svc)

	svc.EXPECT().Login(gomock.Any(), gomock.Any()).
		Return(&authservice.TokenOutput{
			AccessToken:  "access-jwt",
			RefreshToken: "refresh-abc",
			ExpiresIn:    900,
		}, nil)

	resp := postJSON(t, ts, "/api/v1/auth/login", map[string]string{
		"email":    "alice@example.com",
		"password": "pass1word",
	})
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusOK)
	}
	var body httpserver.TokenResponse
	decodeJSON(t, resp, &body)
	if body.AccessToken != "access-jwt" {
		t.Errorf("access_token = %q, want access-jwt", body.AccessToken)
	}
	if body.ExpiresIn != 900 {
		t.Errorf("expires_in = %d, want 900", body.ExpiresIn)
	}
	refreshCookie := findCookie(resp.Cookies(), "refresh_token")
	if refreshCookie == nil {
		t.Fatal("refresh_token cookie not set")
	}
	if refreshCookie.Value != "refresh-abc" {
		t.Errorf("cookie value = %q, want refresh-abc", refreshCookie.Value)
	}
	if !refreshCookie.HttpOnly {
		t.Error("cookie should be HttpOnly")
	}
}

func TestLogin_InvalidJSON(t *testing.T) {
	svc := newMock(t)
	ts := setupServer(t, svc)

	resp := postJSON(t, ts, "/api/v1/auth/login", nil)
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusBadRequest)
	}
}

func TestLogin_EmptyEmail(t *testing.T) {
	svc := newMock(t)
	ts := setupServer(t, svc)

	resp := postJSON(t, ts, "/api/v1/auth/login", map[string]string{
		"email":    "",
		"password": "pass1word",
	})
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusUnprocessableEntity)
	}
	var body httpserver.ErrorResponse
	decodeJSON(t, resp, &body)
	if body.Code != httpserver.VALIDATIONERROR {
		t.Errorf("code = %q, want %q", body.Code, httpserver.VALIDATIONERROR)
	}
}

func TestLogin_InvalidCredentials(t *testing.T) {
	svc := newMock(t)
	ts := setupServer(t, svc)

	svc.EXPECT().Login(gomock.Any(), gomock.Any()).
		Return(nil, authservice.ErrInvalidCredentials)

	resp := postJSON(t, ts, "/api/v1/auth/login", map[string]string{
		"email":    "alice@example.com",
		"password": "wrongpass1",
	})
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusUnauthorized)
	}
	var body httpserver.ErrorResponse
	decodeJSON(t, resp, &body)
	if body.Code != httpserver.INVALIDCREDENTIALS {
		t.Errorf("code = %q, want %q", body.Code, httpserver.INVALIDCREDENTIALS)
	}
}

func TestLogin_EmailNotVerified(t *testing.T) {
	svc := newMock(t)
	ts := setupServer(t, svc)

	svc.EXPECT().Login(gomock.Any(), gomock.Any()).
		Return(nil, authservice.ErrEmailNotVerified)

	resp := postJSON(t, ts, "/api/v1/auth/login", map[string]string{
		"email":    "alice@example.com",
		"password": "pass1word",
	})
	if resp.StatusCode != http.StatusForbidden {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusForbidden)
	}
	var body httpserver.ErrorResponse
	decodeJSON(t, resp, &body)
	if body.Code != httpserver.EMAILNOTVERIFIED {
		t.Errorf("code = %q, want %q", body.Code, httpserver.EMAILNOTVERIFIED)
	}
}

// ──────────────────────────────────────────────
// AuthLogout
// ──────────────────────────────────────────────

func TestLogout_Success(t *testing.T) {
	svc := newMock(t)
	ts := setupAuthServer(t, svc, logoutOp)

	svc.EXPECT().Logout(gomock.Any()).Return(nil)

	resp := authPostJSON(t, ts, "/api/v1/auth/logout", nil)
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusNoContent)
	}
	refreshCookie := findCookie(resp.Cookies(), "refresh_token")
	if refreshCookie == nil {
		t.Fatal("refresh_token cookie not cleared")
	}
	if refreshCookie.MaxAge != -1 {
		t.Errorf("cookie MaxAge = %d, want -1", refreshCookie.MaxAge)
	}
}

func TestLogout_MissingAuthToken(t *testing.T) {
	svc := newMock(t)
	ts := setupAuthServer(t, svc, logoutOp)

	resp := postJSON(t, ts, "/api/v1/auth/logout", nil)
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusUnauthorized)
	}
}

func TestLogout_ServiceError(t *testing.T) {
	svc := newMock(t)
	ts := setupAuthServer(t, svc, logoutOp)

	svc.EXPECT().Logout(gomock.Any()).Return(errors.New("db down"))

	resp := authPostJSON(t, ts, "/api/v1/auth/logout", nil)
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusInternalServerError)
	}
	var body httpserver.ErrorResponse
	decodeJSON(t, resp, &body)
	if body.Code != httpserver.INTERNALERROR {
		t.Errorf("code = %q, want %q", body.Code, httpserver.INTERNALERROR)
	}
}

// ──────────────────────────────────────────────
// AuthRefresh
// ──────────────────────────────────────────────

func TestRefresh_Success(t *testing.T) {
	svc := newMock(t)
	ts := setupServer(t, svc)

	svc.EXPECT().Refresh(gomock.Any(), &authservice.RefreshInput{
		RefreshToken: "refresh-token-abc",
	}).Return(&authservice.TokenOutput{
		AccessToken:  "new-access-jwt",
		RefreshToken: "new-refresh-abc",
		ExpiresIn:    900,
	}, nil)

	resp := postWithCookie(t, ts, "/api/v1/auth/refresh", "refresh_token", "refresh-token-abc")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusOK)
	}
	var body httpserver.RefreshResponse
	decodeJSON(t, resp, &body)
	if body.AccessToken != "new-access-jwt" {
		t.Errorf("access_token = %q, want new-access-jwt", body.AccessToken)
	}
	if body.ExpiresIn != 900 {
		t.Errorf("expires_in = %d, want 900", body.ExpiresIn)
	}
	newCookie := findCookie(resp.Cookies(), "refresh_token")
	if newCookie == nil {
		t.Fatal("new refresh_token cookie not set")
	}
	if newCookie.Value != "new-refresh-abc" {
		t.Errorf("cookie value = %q, want new-refresh-abc", newCookie.Value)
	}
}

func TestRefresh_MissingCookie(t *testing.T) {
	svc := newMock(t)
	ts := setupServer(t, svc)

	resp := postJSON(t, ts, "/api/v1/auth/refresh", nil)
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusUnauthorized)
	}
	var body httpserver.ErrorResponse
	decodeJSON(t, resp, &body)
	if body.Code != httpserver.MISSINGAUTHTOKEN {
		t.Errorf("code = %q, want %q", body.Code, httpserver.MISSINGAUTHTOKEN)
	}
}

func TestRefresh_InvalidRefreshToken(t *testing.T) {
	svc := newMock(t)
	ts := setupServer(t, svc)

	svc.EXPECT().Refresh(gomock.Any(), gomock.Any()).
		Return(nil, authservice.ErrInvalidRefreshToken)

	resp := postWithCookie(t, ts, "/api/v1/auth/refresh", "refresh_token", "bad-token")
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusUnauthorized)
	}
	var body httpserver.ErrorResponse
	decodeJSON(t, resp, &body)
	if body.Code != httpserver.INVALIDREFRESHTOKEN {
		t.Errorf("code = %q, want %q", body.Code, httpserver.INVALIDREFRESHTOKEN)
	}
}

func TestRefresh_SessionRevoked(t *testing.T) {
	svc := newMock(t)
	ts := setupServer(t, svc)

	svc.EXPECT().Refresh(gomock.Any(), gomock.Any()).
		Return(nil, authservice.ErrSessionRevoked)

	resp := postWithCookie(t, ts, "/api/v1/auth/refresh", "refresh_token", "revoked-token")
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusUnauthorized)
	}
	var body httpserver.ErrorResponse
	decodeJSON(t, resp, &body)
	if body.Code != httpserver.INVALIDREFRESHTOKEN {
		t.Errorf("code = %q, want %q", body.Code, httpserver.INVALIDREFRESHTOKEN)
	}
}

func TestRefresh_SessionExpired(t *testing.T) {
	svc := newMock(t)
	ts := setupServer(t, svc)

	svc.EXPECT().Refresh(gomock.Any(), gomock.Any()).
		Return(nil, authservice.ErrSessionExpired)

	resp := postWithCookie(t, ts, "/api/v1/auth/refresh", "refresh_token", "expired-token")
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusUnauthorized)
	}
	var body httpserver.ErrorResponse
	decodeJSON(t, resp, &body)
	if body.Code != httpserver.EXPIREDREFRESHTOKEN {
		t.Errorf("code = %q, want %q", body.Code, httpserver.EXPIREDREFRESHTOKEN)
	}
}

// ──────────────────────────────────────────────
// helpers
// ──────────────────────────────────────────────

func findCookie(cookies []*http.Cookie, name string) *http.Cookie {
	for _, c := range cookies {
		if c.Name == name {
			return c
		}
	}
	return nil
}

func postWithCookie(t *testing.T, ts *httptest.Server, path, cookieName, cookieValue string) *http.Response {
	t.Helper()
	req, err := http.NewRequest(http.MethodPost, ts.URL+path, nil)
	if err != nil {
		t.Fatalf("http.NewRequest: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: cookieName, Value: cookieValue})
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("http.Do: %v", err)
	}
	return resp
}
