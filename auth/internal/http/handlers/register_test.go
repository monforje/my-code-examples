package handlers_test

import (
	"errors"
	"net/http"
	"strings"
	"testing"

	"go.uber.org/mock/gomock"

	httpserver "auth/internal/http/gen"
	authservice "auth/internal/services/auth"
)

// ──────────────────────────────────────────────
// AuthRegister
// ──────────────────────────────────────────────

func TestRegister_Success(t *testing.T) {
	svc := newMock(t)
	ts := setupServer(t, svc)

	svc.EXPECT().Register(gomock.Any(), &authservice.RegisterInput{
		Email:    "alice@example.com",
		Password: "pass1word",
	}).Return(&authservice.RegisterOutput{
		IdentityID: "550e8400-e29b-41d4-a716-446655440000",
		Email:      "alice@example.com",
		Status:     "pending_verification",
	}, nil)

	resp := postJSON(t, ts, "/api/v1/auth/register", map[string]string{
		"email":    "alice@example.com",
		"password": "pass1word",
	})
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusCreated)
	}

	var body httpserver.RegisterResponse
	decodeJSON(t, resp, &body)
	if body.IdentityId != "550e8400-e29b-41d4-a716-446655440000" {
		t.Errorf("identity_id = %q, want uuid", body.IdentityId)
	}
	if body.Email != "alice@example.com" {
		t.Errorf("email = %q, want alice@example.com", body.Email)
	}
	if body.Status != httpserver.RegisterResponseStatusPendingVerification {
		t.Errorf("status = %q, want pending_verification", body.Status)
	}
}

func TestRegister_InvalidJSON(t *testing.T) {
	svc := newMock(t)
	ts := setupServer(t, svc)

	resp := postJSON(t, ts, "/api/v1/auth/register", nil)
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusBadRequest)
	}
}

func TestRegister_EmptyEmail(t *testing.T) {
	svc := newMock(t)
	ts := setupServer(t, svc)

	resp := postJSON(t, ts, "/api/v1/auth/register", map[string]string{
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

func TestRegister_InvalidEmail(t *testing.T) {
	svc := newMock(t)
	ts := setupServer(t, svc)

	resp := postJSON(t, ts, "/api/v1/auth/register", map[string]string{
		"email":    "not-an-email",
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

func TestRegister_ShortPassword(t *testing.T) {
	svc := newMock(t)
	ts := setupServer(t, svc)

	resp := postJSON(t, ts, "/api/v1/auth/register", map[string]string{
		"email":    "alice@example.com",
		"password": "1",
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

func TestRegister_EmailAlreadyExists(t *testing.T) {
	svc := newMock(t)
	ts := setupServer(t, svc)

	svc.EXPECT().Register(gomock.Any(), gomock.Any()).
		Return(nil, authservice.ErrEmailAlreadyExists)

	resp := postJSON(t, ts, "/api/v1/auth/register", map[string]string{
		"email":    "alice@example.com",
		"password": "pass1word",
	})
	if resp.StatusCode != http.StatusConflict {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusConflict)
	}
	var body httpserver.ErrorResponse
	decodeJSON(t, resp, &body)
	if body.Code != httpserver.EMAILALREADYEXISTS {
		t.Errorf("code = %q, want %q", body.Code, httpserver.EMAILALREADYEXISTS)
	}
}

func TestRegister_InternalError(t *testing.T) {
	svc := newMock(t)
	ts := setupServer(t, svc)

	svc.EXPECT().Register(gomock.Any(), gomock.Any()).
		Return(nil, errors.New("db down"))

	resp := postJSON(t, ts, "/api/v1/auth/register", map[string]string{
		"email":    "alice@example.com",
		"password": "pass1word",
	})
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
// AuthRegisterVerify
// ──────────────────────────────────────────────

func TestRegisterVerify_Success(t *testing.T) {
	svc := newMock(t)
	ts := setupServer(t, svc)

	svc.EXPECT().RegisterVerify(gomock.Any(), &authservice.VerifyCodeInput{
		Email: "alice@example.com",
		Code:  "123456",
	}).Return("registration verified", nil)

	resp := postJSON(t, ts, "/api/v1/auth/register/verify", map[string]string{
		"email": "alice@example.com",
		"code":  "123456",
	})
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusOK)
	}
	var body httpserver.MessageResponse
	decodeJSON(t, resp, &body)
	if body.Message != "registration verified" {
		t.Errorf("message = %q, want %q", body.Message, "registration verified")
	}
}

func TestRegisterVerify_InvalidJSON(t *testing.T) {
	svc := newMock(t)
	ts := setupServer(t, svc)

	resp := postJSON(t, ts, "/api/v1/auth/register/verify", nil)
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusBadRequest)
	}
}

func TestRegisterVerify_EmptyCode(t *testing.T) {
	svc := newMock(t)
	ts := setupServer(t, svc)

	resp := postJSON(t, ts, "/api/v1/auth/register/verify", map[string]string{
		"email": "alice@example.com",
		"code":  "",
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

func TestRegisterVerify_InvalidCode(t *testing.T) {
	svc := newMock(t)
	ts := setupServer(t, svc)

	svc.EXPECT().RegisterVerify(gomock.Any(), gomock.Any()).
		Return("", authservice.ErrInvalidCode)

	resp := postJSON(t, ts, "/api/v1/auth/register/verify", map[string]string{
		"email": "alice@example.com",
		"code":  "000000",
	})
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusUnprocessableEntity)
	}
	var body httpserver.ErrorResponse
	decodeJSON(t, resp, &body)
	if body.Code != httpserver.INVALIDCODE {
		t.Errorf("code = %q, want %q", body.Code, httpserver.INVALIDCODE)
	}
}

func TestRegisterVerify_TooManyAttempts(t *testing.T) {
	svc := newMock(t)
	ts := setupServer(t, svc)

	svc.EXPECT().RegisterVerify(gomock.Any(), gomock.Any()).
		Return("", authservice.ErrTooManyAttempts)

	resp := postJSON(t, ts, "/api/v1/auth/register/verify", map[string]string{
		"email": "alice@example.com",
		"code":  "123456",
	})
	if resp.StatusCode != http.StatusTooManyRequests {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusTooManyRequests)
	}
	var body httpserver.ErrorResponse
	decodeJSON(t, resp, &body)
	if body.Code != httpserver.TOOMANYATTEMPTS {
		t.Errorf("code = %q, want %q", body.Code, httpserver.TOOMANYATTEMPTS)
	}
}

// ──────────────────────────────────────────────
// AuthRegisterCodeResend
// ──────────────────────────────────────────────

func TestCodeResend_Success(t *testing.T) {
	svc := newMock(t)
	ts := setupServer(t, svc)

	svc.EXPECT().ResendVerificationCode(gomock.Any(), &authservice.ResendCodeInput{
		Email: "alice@example.com",
	}).Return("verification code sent", nil)

	resp := postJSON(t, ts, "/api/v1/auth/register/code/resend", map[string]string{
		"email": "alice@example.com",
	})
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusOK)
	}
	var body httpserver.MessageResponse
	decodeJSON(t, resp, &body)
	if body.Message != "verification code sent" {
		t.Errorf("message = %q, want %q", body.Message, "verification code sent")
	}
}

func TestCodeResend_InvalidJSON(t *testing.T) {
	svc := newMock(t)
	ts := setupServer(t, svc)

	resp := postJSON(t, ts, "/api/v1/auth/register/code/resend", nil)
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusBadRequest)
	}
}

func TestCodeResend_InvalidEmail(t *testing.T) {
	svc := newMock(t)
	ts := setupServer(t, svc)

	resp := postJSON(t, ts, "/api/v1/auth/register/code/resend", map[string]string{
		"email": "bad",
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

func TestCodeResend_IdentityNotFound(t *testing.T) {
	svc := newMock(t)
	ts := setupServer(t, svc)

	svc.EXPECT().ResendVerificationCode(gomock.Any(), gomock.Any()).
		Return("", authservice.ErrIdentityNotFound)

	resp := postJSON(t, ts, "/api/v1/auth/register/code/resend", map[string]string{
		"email": "alice@example.com",
	})
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusNotFound)
	}
	var body httpserver.ErrorResponse
	decodeJSON(t, resp, &body)
	if body.Code != httpserver.NOTFOUND {
		t.Errorf("code = %q, want %q", body.Code, httpserver.NOTFOUND)
	}
}

func TestCodeResend_EmailAlreadyVerified(t *testing.T) {
	svc := newMock(t)
	ts := setupServer(t, svc)

	svc.EXPECT().ResendVerificationCode(gomock.Any(), gomock.Any()).
		Return("", authservice.ErrEmailAlreadyVerified)

	resp := postJSON(t, ts, "/api/v1/auth/register/code/resend", map[string]string{
		"email": "alice@example.com",
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

func TestCodeResend_EmptyBody(t *testing.T) {
	svc := newMock(t)
	ts := setupServer(t, svc)

	resp, err := http.Post(ts.URL+"/api/v1/auth/register/code/resend", "application/json", strings.NewReader(`{}`))
	if err != nil {
		t.Fatalf("http.Post: %v", err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusUnprocessableEntity)
	}
	var body httpserver.ErrorResponse
	decodeJSON(t, resp, &body)
	if body.Code != httpserver.VALIDATIONERROR {
		t.Errorf("code = %q, want %q", body.Code, httpserver.VALIDATIONERROR)
	}
}
