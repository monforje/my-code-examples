package handlers_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"go.uber.org/mock/gomock"

	httpserver "auth/internal/http/gen"
	"auth/internal/http/handlers"
	authclihandlers "auth/internal/http/handlers/auth_cli"
	hmocks "auth/internal/http/handlers/mocks"
	"auth/internal/http/middleware"
)

// testTokenValidator implements middleware.TokenValidator for tests.
type testTokenValidator func(string) (uuid.UUID, uuid.UUID, string, error)

func (f testTokenValidator) ValidateAccessToken(token string) (uuid.UUID, uuid.UUID, string, error) {
	return f(token)
}

var defaultTestValidator = testTokenValidator(func(token string) (uuid.UUID, uuid.UUID, string, error) {
	if token != "test-access-token" {
		return uuid.Nil, uuid.Nil, "", errors.New("invalid token")
	}
	return uuid.New(), uuid.New(), "jwt-id", nil
})

// testCombinedHandler wraps AuthHandlers and a stub CLI handler to satisfy ServerInterface.
type testCombinedHandler struct {
	*handlers.AuthHandlers
	cli *authclihandlers.AuthCliHandlers
}

func (c *testCombinedHandler) AuthDeviceStart(ctx echo.Context) error {
	return c.cli.AuthDeviceStart(ctx)
}

func (c *testCombinedHandler) AuthDeviceConfirm(ctx echo.Context) error {
	return c.cli.AuthDeviceConfirm(ctx)
}

func (c *testCombinedHandler) AuthDeviceToken(ctx echo.Context) error {
	return c.cli.AuthDeviceToken(ctx)
}

func (c *testCombinedHandler) AuthCliRefresh(ctx echo.Context) error {
	return c.cli.AuthCliRefresh(ctx)
}

// setupServer creates an httptest.Server with public (unprotected) routes.
func setupServer(t *testing.T, svc *hmocks.MockAuthService) *httptest.Server {
	t.Helper()
	e := echo.New()
	h := handlers.NewAuthHandlers(svc, 30*24*time.Hour)
	cliH := authclihandlers.NewAuthCliHandlers(svc, 30*24*time.Hour)
	combined := &testCombinedHandler{AuthHandlers: h, cli: cliH}
	httpserver.RegisterHandlersWithOptions(e, combined, httpserver.RegisterHandlersOptions{
		BaseURL: "/api/v1",
	})
	ts := httptest.NewServer(e)
	t.Cleanup(ts.Close)
	return ts
}

// setupAuthServer creates an httptest.Server with BearerAuth middleware on protected operations.
func setupAuthServer(t *testing.T, svc *hmocks.MockAuthService, protectedOps ...string) *httptest.Server {
	t.Helper()
	e := echo.New()
	h := handlers.NewAuthHandlers(svc, 30*24*time.Hour)
	cliH := authclihandlers.NewAuthCliHandlers(svc, 30*24*time.Hour)
	combined := &testCombinedHandler{AuthHandlers: h, cli: cliH}

	mwMap := make(map[string][]echo.MiddlewareFunc)
	for _, op := range protectedOps {
		mwMap[op] = []echo.MiddlewareFunc{middleware.BearerAuth(defaultTestValidator)}
	}

	httpserver.RegisterHandlersWithOptions(e, combined, httpserver.RegisterHandlersOptions{
		BaseURL:              "/api/v1",
		OperationMiddlewares: mwMap,
	})
	ts := httptest.NewServer(e)
	t.Cleanup(ts.Close)
	return ts
}

// postJSON sends a POST request without Authorization header.
// If body is nil, sends malformed JSON to trigger bind error.
func postJSON(t *testing.T, ts *httptest.Server, path string, body any) *http.Response {
	t.Helper()
	var buf bytes.Buffer
	if body != nil {
		if err := json.NewEncoder(&buf).Encode(body); err != nil {
			t.Fatalf("json.Encode: %v", err)
		}
	} else {
		buf.WriteString("{bad")
	}
	resp, err := http.Post(ts.URL+path, "application/json", &buf)
	if err != nil {
		t.Fatalf("http.Post: %v", err)
	}
	return resp
}

// authPostJSON sends a POST request with Bearer test-access-token.
// If body is nil, sends malformed JSON to trigger bind error.
func authPostJSON(t *testing.T, ts *httptest.Server, path string, body any) *http.Response {
	t.Helper()
	var buf bytes.Buffer
	if body != nil {
		if err := json.NewEncoder(&buf).Encode(body); err != nil {
			t.Fatalf("json.Encode: %v", err)
		}
	} else {
		buf.WriteString("{bad")
	}
	req, err := http.NewRequest(http.MethodPost, ts.URL+path, &buf)
	if err != nil {
		t.Fatalf("http.NewRequest: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer test-access-token")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("http.Do: %v", err)
	}
	return resp
}

// decodeJSON decodes the response body into v and closes the body.
func decodeJSON(t *testing.T, resp *http.Response, v any) {
	t.Helper()
	defer resp.Body.Close()
	if err := json.NewDecoder(resp.Body).Decode(v); err != nil {
		t.Fatalf("json.Decode: %v", err)
	}
}

// authGet sends a GET request with Bearer test-access-token.
func authGet(t *testing.T, ts *httptest.Server, path string) *http.Response {
	t.Helper()
	req, err := http.NewRequest(http.MethodGet, ts.URL+path, nil)
	if err != nil {
		t.Fatalf("http.NewRequest: %v", err)
	}
	req.Header.Set("Authorization", "Bearer test-access-token")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("http.Do: %v", err)
	}
	return resp
}

// getNoAuth sends a GET request without Authorization header.
func getNoAuth(t *testing.T, ts *httptest.Server, path string) *http.Response {
	t.Helper()
	resp, err := http.Get(ts.URL + path)
	if err != nil {
		t.Fatalf("http.Get: %v", err)
	}
	return resp
}

// deleteNoAuth sends a DELETE request without Authorization header.
// If body is nil, sends malformed JSON to trigger bind error.
func deleteNoAuth(t *testing.T, ts *httptest.Server, path string, body any) *http.Response {
	t.Helper()
	var buf bytes.Buffer
	if body != nil {
		if err := json.NewEncoder(&buf).Encode(body); err != nil {
			t.Fatalf("json.Encode: %v", err)
		}
	} else {
		buf.WriteString("{bad")
	}
	req, err := http.NewRequest(http.MethodDelete, ts.URL+path, &buf)
	if err != nil {
		t.Fatalf("http.NewRequest: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("http.Do: %v", err)
	}
	return resp
}

// authDelete sends a DELETE request with Bearer test-access-token and JSON body.
// If body is nil, sends malformed JSON to trigger bind error.
func authDeleteJSON(t *testing.T, ts *httptest.Server, path string, body any) *http.Response {
	t.Helper()
	var buf bytes.Buffer
	if body != nil {
		if err := json.NewEncoder(&buf).Encode(body); err != nil {
			t.Fatalf("json.Encode: %v", err)
		}
	} else {
		buf.WriteString("{bad")
	}
	req, err := http.NewRequest(http.MethodDelete, ts.URL+path, &buf)
	if err != nil {
		t.Fatalf("http.NewRequest: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer test-access-token")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("http.Do: %v", err)
	}
	return resp
}

// newMock creates a new MockAuthService with a gomock controller that is cleaned up with t.
func newMock(t *testing.T) *hmocks.MockAuthService {
	t.Helper()
	ctrl := gomock.NewController(t)
	return hmocks.NewMockAuthService(ctrl)
}
