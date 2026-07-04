package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"

	"tasks/internal/http/middleware"
)

func doReq(t *testing.T, authHeader string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, "/reports", nil)
	if authHeader != "" {
		req.Header.Set(echo.HeaderAuthorization, authHeader)
	}
	rec := httptest.NewRecorder()
	c := echo.New().NewContext(req, rec)

	h := middleware.ServiceToken("s3cret")(echo.HandlerFunc(func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	}))
	if err := h(c); err != nil {
		t.Fatalf("middleware error: %v", err)
	}
	return rec
}

func TestServiceToken_Valid(t *testing.T) {
	rec := doReq(t, "Bearer s3cret")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
}

func TestServiceToken_Wrong(t *testing.T) {
	rec := doReq(t, "Bearer nope")
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", rec.Code)
	}
}

func TestServiceToken_Missing(t *testing.T) {
	rec := doReq(t, "")
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", rec.Code)
	}
}

func TestServiceToken_MalformedHeader(t *testing.T) {
	rec := doReq(t, "s3cret")
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", rec.Code)
	}
}
