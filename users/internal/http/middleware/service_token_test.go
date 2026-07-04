package middleware_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"

	httpserver "users/internal/http/gen"
	"users/internal/http/middleware"
)

func TestServiceToken_Success(t *testing.T) {
	e := echo.New()
	mw := middleware.ServiceToken("valid-token")

	req := httptest.NewRequest(http.MethodGet, "/git-user/me", nil)
	req.Header.Set("X-Service-Token", "valid-token")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := mw(func(c echo.Context) error {
		return c.NoContent(http.StatusNoContent)
	})(c)
	if err != nil {
		t.Fatalf("ServiceToken() error = %v", err)
	}
	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusNoContent)
	}
}

func TestServiceToken_MissingHeader(t *testing.T) {
	e := echo.New()
	mw := middleware.ServiceToken("valid-token")
	rec := httptest.NewRecorder()
	c := e.NewContext(httptest.NewRequest(http.MethodGet, "/git-user/me", nil), rec)

	err := mw(func(echo.Context) error {
		t.Fatal("next handler must not be called")
		return nil
	})(c)
	assertServiceTokenError(t, err, rec, http.StatusUnauthorized, httpserver.MISSINGAUTHTOKEN)
}

func TestServiceToken_InvalidToken(t *testing.T) {
	e := echo.New()
	mw := middleware.ServiceToken("valid-token")
	req := httptest.NewRequest(http.MethodGet, "/git-user/me", nil)
	req.Header.Set("X-Service-Token", "wrong-token")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := mw(func(echo.Context) error {
		t.Fatal("next handler must not be called")
		return nil
	})(c)
	assertServiceTokenError(t, err, rec, http.StatusUnauthorized, httpserver.INVALIDAUTHTOKEN)
}

func assertServiceTokenError(t *testing.T, err error, rec *httptest.ResponseRecorder, wantStatus int, wantCode httpserver.ErrorCode) {
	t.Helper()
	if err != nil {
		t.Fatalf("error = %v, want nil", err)
	}
	if rec.Code != wantStatus {
		t.Fatalf("status = %d, want %d", rec.Code, wantStatus)
	}
	var body httpserver.ErrorResponse
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if body.Code != wantCode {
		t.Fatalf("code = %q, want %q", body.Code, wantCode)
	}
}
