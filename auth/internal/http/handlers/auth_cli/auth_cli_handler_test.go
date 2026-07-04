package authclihandlers_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	"go.uber.org/mock/gomock"

	httpserver "auth/internal/http/gen"
	authclihandlers "auth/internal/http/handlers/auth_cli"
	"auth/internal/http/handlers/auth_cli/mocks"
	authservice "auth/internal/services/auth"
)

func newMock(t *testing.T) *mocks.MockAuthService {
	t.Helper()
	ctrl := gomock.NewController(t)
	return mocks.NewMockAuthService(ctrl)
}

func setupCliServer(t *testing.T, svc *mocks.MockAuthService) *httptest.Server {
	t.Helper()
	e := echo.New()
	h := authclihandlers.NewAuthCliHandlers(svc, 30*24*time.Hour)

	e.POST("/api/v1/auth/device/start", h.AuthDeviceStart)
	e.POST("/api/v1/auth/device/confirm", h.AuthDeviceConfirm)
	e.POST("/api/v1/auth/device/token", h.AuthDeviceToken)
	e.POST("/api/v1/auth/cli/refresh", h.AuthCliRefresh)

	ts := httptest.NewServer(e)
	t.Cleanup(ts.Close)
	return ts
}

func postJSON(t *testing.T, ts *httptest.Server, path string, body any) *http.Response {
	t.Helper()
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(body); err != nil {
		t.Fatalf("json.Encode: %v", err)
	}
	resp, err := http.Post(ts.URL+path, "application/json", &buf)
	if err != nil {
		t.Fatalf("http.Post: %v", err)
	}
	return resp
}

func decodeJSON(t *testing.T, resp *http.Response, v any) {
	t.Helper()
	defer resp.Body.Close()
	if err := json.NewDecoder(resp.Body).Decode(v); err != nil {
		t.Fatalf("json.Decode: %v", err)
	}
}

func TestAuthDeviceStart_Success(t *testing.T) {
	svc := newMock(t)
	ts := setupCliServer(t, svc)

	svc.EXPECT().DeviceStart(gomock.Any()).Return(&authservice.DeviceStartOutput{
		DeviceCode:      "test-device-code",
		UserCode:        "ABCD-EFGH",
		VerificationURL: "https://codurity.dev/cli/login",
		ExpiresIn:       600,
		Interval:        3,
	}, nil)

	resp := postJSON(t, ts, "/api/v1/auth/device/start", nil)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusOK)
	}

	var body httpserver.DeviceStartResponse
	decodeJSON(t, resp, &body)

	if body.DeviceCode != "test-device-code" {
		t.Errorf("DeviceCode = %v, want test-device-code", body.DeviceCode)
	}
	if body.UserCode != "ABCD-EFGH" {
		t.Errorf("UserCode = %v, want ABCD-EFGH", body.UserCode)
	}
}

func TestAuthDeviceStart_Error(t *testing.T) {
	svc := newMock(t)
	ts := setupCliServer(t, svc)

	svc.EXPECT().DeviceStart(gomock.Any()).Return(nil, errors.New("internal error"))

	resp := postJSON(t, ts, "/api/v1/auth/device/start", nil)
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusInternalServerError)
	}
}

func TestAuthDeviceConfirm_Success(t *testing.T) {
	svc := newMock(t)
	ts := setupCliServer(t, svc)

	svc.EXPECT().DeviceConfirm(gomock.Any(), &authservice.DeviceConfirmInput{
		UserCode: "ABCD-EFGH",
	}).Return(&authservice.DeviceConfirmOutput{
		Status: "confirmed",
	}, nil)

	resp := postJSON(t, ts, "/api/v1/auth/device/confirm", httpserver.DeviceConfirmRequest{
		UserCode: "ABCD-EFGH",
	})
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusOK)
	}

	var body httpserver.DeviceConfirmResponse
	decodeJSON(t, resp, &body)

	if body.Status != httpserver.Confirmed {
		t.Errorf("Status = %v, want confirmed", body.Status)
	}
}

func TestAuthDeviceConfirm_NotFound(t *testing.T) {
	svc := newMock(t)
	ts := setupCliServer(t, svc)

	svc.EXPECT().DeviceConfirm(gomock.Any(), gomock.Any()).Return(nil, authservice.ErrDeviceCodeNotFound)

	resp := postJSON(t, ts, "/api/v1/auth/device/confirm", httpserver.DeviceConfirmRequest{
		UserCode: "XXXX-XXXX",
	})
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusNotFound)
	}
}

func TestAuthDeviceConfirm_AlreadyConfirmed(t *testing.T) {
	svc := newMock(t)
	ts := setupCliServer(t, svc)

	svc.EXPECT().DeviceConfirm(gomock.Any(), gomock.Any()).Return(nil, authservice.ErrDeviceCodeAlreadyConfirmed)

	resp := postJSON(t, ts, "/api/v1/auth/device/confirm", httpserver.DeviceConfirmRequest{
		UserCode: "ABCD-EFGH",
	})
	if resp.StatusCode != http.StatusConflict {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusConflict)
	}
}

func TestAuthDeviceToken_Success(t *testing.T) {
	svc := newMock(t)
	ts := setupCliServer(t, svc)

	svc.EXPECT().DeviceToken(gomock.Any(), &authservice.DeviceTokenInput{
		DeviceCode: "test-device-code",
	}).Return(&authservice.DeviceTokenOutput{
		AccessToken:  "access-token",
		RefreshToken: "refresh-token",
		ExpiresIn:    900,
	}, nil)

	resp := postJSON(t, ts, "/api/v1/auth/device/token", httpserver.DeviceTokenRequest{
		DeviceCode: "test-device-code",
	})
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusOK)
	}

	var body httpserver.CliTokenResponse
	decodeJSON(t, resp, &body)

	if body.AccessToken != "access-token" {
		t.Errorf("AccessToken = %v, want access-token", body.AccessToken)
	}
	if body.RefreshToken != "refresh-token" {
		t.Errorf("RefreshToken = %v, want refresh-token", body.RefreshToken)
	}
	if body.TokenType != httpserver.CliTokenResponseTokenTypeBearer {
		t.Errorf("TokenType = %v, want Bearer", body.TokenType)
	}
}

func TestAuthDeviceToken_NotConfirmed(t *testing.T) {
	svc := newMock(t)
	ts := setupCliServer(t, svc)

	svc.EXPECT().DeviceToken(gomock.Any(), gomock.Any()).Return(nil, authservice.ErrDeviceCodeNotConfirmed)

	resp := postJSON(t, ts, "/api/v1/auth/device/token", httpserver.DeviceTokenRequest{
		DeviceCode: "test-device-code",
	})
	if resp.StatusCode != http.StatusPreconditionRequired {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusPreconditionRequired)
	}
}

func TestAuthDeviceToken_PollTooFrequent(t *testing.T) {
	svc := newMock(t)
	ts := setupCliServer(t, svc)

	svc.EXPECT().DeviceToken(gomock.Any(), gomock.Any()).Return(nil, authservice.ErrPollTooFrequent)

	resp := postJSON(t, ts, "/api/v1/auth/device/token", httpserver.DeviceTokenRequest{
		DeviceCode: "test-device-code",
	})
	if resp.StatusCode != http.StatusTooManyRequests {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusTooManyRequests)
	}
}

func TestAuthCliRefresh_Success(t *testing.T) {
	svc := newMock(t)
	ts := setupCliServer(t, svc)

	svc.EXPECT().CliRefresh(gomock.Any(), &authservice.CliRefreshInput{
		RefreshToken: "refresh-token",
	}).Return(&authservice.CliRefreshOutput{
		AccessToken: "new-access-token",
		ExpiresIn:   900,
	}, nil)

	resp := postJSON(t, ts, "/api/v1/auth/cli/refresh", httpserver.CliRefreshRequest{
		RefreshToken: "refresh-token",
	})
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusOK)
	}

	var body httpserver.CliRefreshResponse
	decodeJSON(t, resp, &body)

	if body.AccessToken != "new-access-token" {
		t.Errorf("AccessToken = %v, want new-access-token", body.AccessToken)
	}
	if body.TokenType != httpserver.CliRefreshResponseTokenTypeBearer {
		t.Errorf("TokenType = %v, want Bearer", body.TokenType)
	}
}

func TestAuthCliRefresh_InvalidRefreshToken(t *testing.T) {
	svc := newMock(t)
	ts := setupCliServer(t, svc)

	svc.EXPECT().CliRefresh(gomock.Any(), gomock.Any()).Return(nil, authservice.ErrInvalidRefreshToken)

	resp := postJSON(t, ts, "/api/v1/auth/cli/refresh", httpserver.CliRefreshRequest{
		RefreshToken: "invalid-token",
	})
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusUnauthorized)
	}
}
