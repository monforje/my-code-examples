package middleware_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"tasks/internal/authctx"
	httpserver "tasks/internal/http/gen"
	"tasks/internal/http/middleware"
)

type tokenValidatorFunc func(string) (uuid.UUID, uuid.UUID, string, error)

func (f tokenValidatorFunc) ValidateAccessToken(token string) (uuid.UUID, uuid.UUID, string, error) {
	return f(token)
}

func TestBearerAuth_Success(t *testing.T) {
	e := echo.New()
	identityID := uuid.New()
	sessionID := uuid.New()
	mw := middleware.BearerAuth(tokenValidatorFunc(func(token string) (uuid.UUID, uuid.UUID, string, error) {
		if token != "access-token" {
			t.Fatalf("token = %q, want access-token", token)
		}
		return identityID, sessionID, "jwt-id", nil
	}))

	req := httptest.NewRequest(http.MethodGet, "/auth/me", nil)
	req.Header.Set(echo.HeaderAuthorization, "Bearer access-token")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := mw(func(c echo.Context) error {
		gotIdentityID, gotSessionID, err := authctx.FromContext(c.Request().Context())
		if err != nil {
			t.Fatalf("FromContext() error = %v", err)
		}
		if gotIdentityID != identityID {
			t.Fatalf("identityID = %s, want %s", gotIdentityID, identityID)
		}
		if gotSessionID != sessionID {
			t.Fatalf("sessionID = %s, want %s", gotSessionID, sessionID)
		}
		return c.NoContent(http.StatusNoContent)
	})(c)
	if err != nil {
		t.Fatalf("BearerAuth() error = %v", err)
	}
	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusNoContent)
	}
}

func TestBearerAuth_MissingHeader(t *testing.T) {
	e := echo.New()
	mw := middleware.BearerAuth(tokenValidatorFunc(func(string) (uuid.UUID, uuid.UUID, string, error) {
		t.Fatal("ValidateAccessToken must not be called")
		return uuid.Nil, uuid.Nil, "", nil
	}))
	rec := httptest.NewRecorder()
	c := e.NewContext(httptest.NewRequest(http.MethodGet, "/auth/me", nil), rec)

	err := mw(func(echo.Context) error { return nil })(c)
	assertErrorResponse(t, err, rec, http.StatusUnauthorized, httpserver.MISSINGAUTHTOKEN)
}

func TestBearerAuth_InvalidScheme(t *testing.T) {
	e := echo.New()
	mw := middleware.BearerAuth(tokenValidatorFunc(func(string) (uuid.UUID, uuid.UUID, string, error) {
		t.Fatal("ValidateAccessToken must not be called")
		return uuid.Nil, uuid.Nil, "", nil
	}))
	req := httptest.NewRequest(http.MethodGet, "/auth/me", nil)
	req.Header.Set(echo.HeaderAuthorization, "Basic access-token")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := mw(func(echo.Context) error { return nil })(c)
	assertErrorResponse(t, err, rec, http.StatusUnauthorized, httpserver.INVALIDAUTHTOKEN)
}

func TestBearerAuth_InvalidToken(t *testing.T) {
	e := echo.New()
	mw := middleware.BearerAuth(tokenValidatorFunc(func(string) (uuid.UUID, uuid.UUID, string, error) {
		return uuid.Nil, uuid.Nil, "", errors.New("invalid token")
	}))
	req := httptest.NewRequest(http.MethodGet, "/auth/me", nil)
	req.Header.Set(echo.HeaderAuthorization, "Bearer access-token")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := mw(func(echo.Context) error {
		t.Fatal("next handler must not be called")
		return nil
	})(c)
	assertErrorResponse(t, err, rec, http.StatusUnauthorized, httpserver.INVALIDAUTHTOKEN)
}

func assertErrorResponse(t *testing.T, err error, rec *httptest.ResponseRecorder, wantStatus int, wantCode httpserver.ErrorCode) {
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
