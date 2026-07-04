package handlers_test

import (
	"errors"
	"net/http"
	"testing"

	"go.uber.org/mock/gomock"

	httpserver "auth/internal/http/gen"
	authservice "auth/internal/services/auth"
)

// ──────────────────────────────────────────────
// AuthPasswordForgot
// ──────────────────────────────────────────────

func TestPasswordForgot_Success(t *testing.T) {
	svc := newMock(t)
	ts := setupServer(t, svc)

	svc.EXPECT().ForgotPassword(gomock.Any(), &authservice.ForgotPasswordInput{
		Email: "alice@example.com",
	}).Return("if an account exists, a code has been sent", nil)

	resp := postJSON(t, ts, "/api/v1/auth/password/forgot", map[string]string{
		"email": "alice@example.com",
	})
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusOK)
	}
	var body httpserver.MessageResponse
	decodeJSON(t, resp, &body)
	if body.Message != "if an account exists, a code has been sent" {
		t.Errorf("message = %q", body.Message)
	}
}

func TestPasswordForgot_InvalidJSON(t *testing.T) {
	svc := newMock(t)
	ts := setupServer(t, svc)

	resp := postJSON(t, ts, "/api/v1/auth/password/forgot", nil)
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusBadRequest)
	}
}

func TestPasswordForgot_InvalidEmail(t *testing.T) {
	svc := newMock(t)
	ts := setupServer(t, svc)

	resp := postJSON(t, ts, "/api/v1/auth/password/forgot", map[string]string{
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

func TestPasswordForgot_EmptyEmail(t *testing.T) {
	svc := newMock(t)
	ts := setupServer(t, svc)

	resp := postJSON(t, ts, "/api/v1/auth/password/forgot", map[string]string{
		"email": "",
	})
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusUnprocessableEntity)
	}
}

func TestPasswordForgot_ServiceError(t *testing.T) {
	svc := newMock(t)
	ts := setupServer(t, svc)

	svc.EXPECT().ForgotPassword(gomock.Any(), gomock.Any()).
		Return("", errors.New("db down"))

	resp := postJSON(t, ts, "/api/v1/auth/password/forgot", map[string]string{
		"email": "alice@example.com",
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
// AuthPasswordForgotVerify
// ──────────────────────────────────────────────

func TestPasswordForgotVerify_Success(t *testing.T) {
	svc := newMock(t)
	ts := setupServer(t, svc)

	svc.EXPECT().ForgotPasswordVerify(gomock.Any(), &authservice.ForgotPasswordVerifyInput{
		Email: "alice@example.com",
		Code:  "123456",
	}).Return(&authservice.ResetTokenOutput{
		ResetToken: "reset-token-abc",
		ExpiresIn:  3600,
	}, nil)

	resp := postJSON(t, ts, "/api/v1/auth/password/forgot/verify", map[string]string{
		"email": "alice@example.com",
		"code":  "123456",
	})
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusOK)
	}
	var body httpserver.ResetTokenResponse
	decodeJSON(t, resp, &body)
	if body.ResetToken != "reset-token-abc" {
		t.Errorf("reset_token = %q, want reset-token-abc", body.ResetToken)
	}
	if body.ExpiresIn != 3600 {
		t.Errorf("expires_in = %d, want 3600", body.ExpiresIn)
	}
}

func TestPasswordForgotVerify_InvalidJSON(t *testing.T) {
	svc := newMock(t)
	ts := setupServer(t, svc)

	resp := postJSON(t, ts, "/api/v1/auth/password/forgot/verify", nil)
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusBadRequest)
	}
}

func TestPasswordForgotVerify_EmptyCode(t *testing.T) {
	svc := newMock(t)
	ts := setupServer(t, svc)

	resp := postJSON(t, ts, "/api/v1/auth/password/forgot/verify", map[string]string{
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

func TestPasswordForgotVerify_InvalidCode(t *testing.T) {
	svc := newMock(t)
	ts := setupServer(t, svc)

	svc.EXPECT().ForgotPasswordVerify(gomock.Any(), gomock.Any()).
		Return(nil, authservice.ErrInvalidCode)

	resp := postJSON(t, ts, "/api/v1/auth/password/forgot/verify", map[string]string{
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

func TestPasswordForgotVerify_TooManyAttempts(t *testing.T) {
	svc := newMock(t)
	ts := setupServer(t, svc)

	svc.EXPECT().ForgotPasswordVerify(gomock.Any(), gomock.Any()).
		Return(nil, authservice.ErrTooManyAttempts)

	resp := postJSON(t, ts, "/api/v1/auth/password/forgot/verify", map[string]string{
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
// AuthPasswordForgotCodeResend
// ──────────────────────────────────────────────

func TestPasswordForgotCodeResend_Success(t *testing.T) {
	svc := newMock(t)
	ts := setupServer(t, svc)

	svc.EXPECT().ForgotPasswordCodeResend(gomock.Any(), &authservice.ResendCodeInput{
		Email: "alice@example.com",
	}).Return("if an account exists, a code has been sent", nil)

	resp := postJSON(t, ts, "/api/v1/auth/password/forgot/code/resend", map[string]string{
		"email": "alice@example.com",
	})
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusOK)
	}
	var body httpserver.MessageResponse
	decodeJSON(t, resp, &body)
	if body.Message != "if an account exists, a code has been sent" {
		t.Errorf("message = %q", body.Message)
	}
}

func TestPasswordForgotCodeResend_InvalidJSON(t *testing.T) {
	svc := newMock(t)
	ts := setupServer(t, svc)

	resp := postJSON(t, ts, "/api/v1/auth/password/forgot/code/resend", nil)
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusBadRequest)
	}
}

func TestPasswordForgotCodeResend_InvalidEmail(t *testing.T) {
	svc := newMock(t)
	ts := setupServer(t, svc)

	resp := postJSON(t, ts, "/api/v1/auth/password/forgot/code/resend", map[string]string{
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

func TestPasswordForgotCodeResend_ServiceError(t *testing.T) {
	svc := newMock(t)
	ts := setupServer(t, svc)

	svc.EXPECT().ForgotPasswordCodeResend(gomock.Any(), gomock.Any()).
		Return("", errors.New("db down"))

	resp := postJSON(t, ts, "/api/v1/auth/password/forgot/code/resend", map[string]string{
		"email": "alice@example.com",
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
// AuthPasswordReset
// ──────────────────────────────────────────────

func TestPasswordReset_Success(t *testing.T) {
	svc := newMock(t)
	ts := setupServer(t, svc)

	svc.EXPECT().ResetPassword(gomock.Any(), &authservice.ResetPasswordInput{
		ResetToken:  "reset-token-abc",
		NewPassword: "newpass1word",
	}).Return("password reset", nil)

	resp := postJSON(t, ts, "/api/v1/auth/password/reset", map[string]string{
		"reset_token":  "reset-token-abc",
		"new_password": "newpass1word",
	})
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusOK)
	}
	var body httpserver.MessageResponse
	decodeJSON(t, resp, &body)
	if body.Message != "password reset" {
		t.Errorf("message = %q, want password reset", body.Message)
	}
}

func TestPasswordReset_InvalidJSON(t *testing.T) {
	svc := newMock(t)
	ts := setupServer(t, svc)

	resp := postJSON(t, ts, "/api/v1/auth/password/reset", nil)
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusBadRequest)
	}
}

func TestPasswordReset_EmptyToken(t *testing.T) {
	svc := newMock(t)
	ts := setupServer(t, svc)

	resp := postJSON(t, ts, "/api/v1/auth/password/reset", map[string]string{
		"reset_token":  "",
		"new_password": "newpass1word",
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

func TestPasswordReset_ShortPassword(t *testing.T) {
	svc := newMock(t)
	ts := setupServer(t, svc)

	resp := postJSON(t, ts, "/api/v1/auth/password/reset", map[string]string{
		"reset_token":  "reset-token-abc",
		"new_password": "1",
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

func TestPasswordReset_InvalidResetToken(t *testing.T) {
	svc := newMock(t)
	ts := setupServer(t, svc)

	svc.EXPECT().ResetPassword(gomock.Any(), gomock.Any()).
		Return("", authservice.ErrInvalidResetToken)

	resp := postJSON(t, ts, "/api/v1/auth/password/reset", map[string]string{
		"reset_token":  "bad-token",
		"new_password": "newpass1word",
	})
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusUnprocessableEntity)
	}
	var body httpserver.ErrorResponse
	decodeJSON(t, resp, &body)
	if body.Code != httpserver.RESETTOKENINVALID {
		t.Errorf("code = %q, want %q", body.Code, httpserver.RESETTOKENINVALID)
	}
}

func TestPasswordReset_ResetTokenExpired(t *testing.T) {
	svc := newMock(t)
	ts := setupServer(t, svc)

	svc.EXPECT().ResetPassword(gomock.Any(), gomock.Any()).
		Return("", authservice.ErrResetTokenExpired)

	resp := postJSON(t, ts, "/api/v1/auth/password/reset", map[string]string{
		"reset_token":  "expired-token",
		"new_password": "newpass1word",
	})
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusUnprocessableEntity)
	}
	var body httpserver.ErrorResponse
	decodeJSON(t, resp, &body)
	if body.Code != httpserver.RESETTOKENEXPIRED {
		t.Errorf("code = %q, want %q", body.Code, httpserver.RESETTOKENEXPIRED)
	}
}
