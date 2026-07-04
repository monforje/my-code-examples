package handlers_test

import (
	"errors"
	"net/http"
	"testing"
	"time"

	"go.uber.org/mock/gomock"

	httpserver "auth/internal/http/gen"
	authservice "auth/internal/services/auth"
)

const (
	meGetOps          = "auth.me.get"
	meDeleteOps       = "auth.me.delete"
	meDeleteVerifyOps = "auth.me.delete.verify"
	meDeleteResendOps = "auth.me.delete.code.resend"
)

// ──────────────────────────────────────────────
// AuthMeGet
// ──────────────────────────────────────────────

func TestMeGet_Success(t *testing.T) {
	svc := newMock(t)
	ts := setupAuthServer(t, svc, meGetOps, meDeleteOps, meDeleteVerifyOps, meDeleteResendOps)

	svc.EXPECT().GetMe(gomock.Any()).Return(&authservice.Identity{
		ID:            "550e8400-e29b-41d4-a716-446655440000",
		Email:         "user@example.com",
		EmailVerified: true,
		Status:        "active",
		CreatedAt:     time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC),
	}, nil)

	resp := authGet(t, ts, "/api/v1/auth/me")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusOK)
	}
	var body httpserver.Identity
	decodeJSON(t, resp, &body)
	if body.Email != "user@example.com" {
		t.Errorf("email = %q, want user@example.com", body.Email)
	}
	if body.EmailVerified != true {
		t.Errorf("email_verified = %v, want true", body.EmailVerified)
	}
	if body.Status != httpserver.IdentityStatusActive {
		t.Errorf("status = %q, want %q", body.Status, httpserver.IdentityStatusActive)
	}
}

func TestMeGet_MissingAuthToken(t *testing.T) {
	svc := newMock(t)
	ts := setupAuthServer(t, svc, meGetOps, meDeleteOps, meDeleteVerifyOps, meDeleteResendOps)

	resp := getNoAuth(t, ts, "/api/v1/auth/me")
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusUnauthorized)
	}
}

func TestMeGet_ServiceError(t *testing.T) {
	svc := newMock(t)
	ts := setupAuthServer(t, svc, meGetOps, meDeleteOps, meDeleteVerifyOps, meDeleteResendOps)

	svc.EXPECT().GetMe(gomock.Any()).Return(nil, errors.New("db down"))

	resp := authGet(t, ts, "/api/v1/auth/me")
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
// AuthMeDelete
// ──────────────────────────────────────────────

func TestMeDelete_Success(t *testing.T) {
	svc := newMock(t)
	ts := setupAuthServer(t, svc, meGetOps, meDeleteOps, meDeleteVerifyOps, meDeleteResendOps)

	svc.EXPECT().DeleteAccount(gomock.Any(), &authservice.DeleteAccountInput{
		Password: "oldpass1word",
	}).Return("account delete code sent", nil)

	resp := authDeleteJSON(t, ts, "/api/v1/auth/me", map[string]string{
		"password": "oldpass1word",
	})
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusOK)
	}
	var body httpserver.MessageResponse
	decodeJSON(t, resp, &body)
	if body.Message != "account delete code sent" {
		t.Errorf("message = %q, want account delete code sent", body.Message)
	}
}

func TestMeDelete_MissingAuthToken(t *testing.T) {
	svc := newMock(t)
	ts := setupAuthServer(t, svc, meGetOps, meDeleteOps, meDeleteVerifyOps, meDeleteResendOps)

	resp := deleteNoAuth(t, ts, "/api/v1/auth/me", map[string]string{
		"password": "oldpass1word",
	})
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusUnauthorized)
	}
}

func TestMeDelete_InvalidJSON(t *testing.T) {
	svc := newMock(t)
	ts := setupAuthServer(t, svc, meGetOps, meDeleteOps, meDeleteVerifyOps, meDeleteResendOps)

	resp := authDeleteJSON(t, ts, "/api/v1/auth/me", nil)
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusBadRequest)
	}
}

func TestMeDelete_EmptyPassword(t *testing.T) {
	svc := newMock(t)
	ts := setupAuthServer(t, svc, meGetOps, meDeleteOps, meDeleteVerifyOps, meDeleteResendOps)

	resp := authDeleteJSON(t, ts, "/api/v1/auth/me", map[string]string{
		"password": "",
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

func TestMeDelete_ShortPassword(t *testing.T) {
	svc := newMock(t)
	ts := setupAuthServer(t, svc, meGetOps, meDeleteOps, meDeleteVerifyOps, meDeleteResendOps)

	resp := authDeleteJSON(t, ts, "/api/v1/auth/me", map[string]string{
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

func TestMeDelete_IncorrectPassword(t *testing.T) {
	svc := newMock(t)
	ts := setupAuthServer(t, svc, meGetOps, meDeleteOps, meDeleteVerifyOps, meDeleteResendOps)

	svc.EXPECT().DeleteAccount(gomock.Any(), gomock.Any()).
		Return("", authservice.ErrCurrentPasswordIncorrect)

	resp := authDeleteJSON(t, ts, "/api/v1/auth/me", map[string]string{
		"password": "wrongpass1",
	})
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusUnprocessableEntity)
	}
	var body httpserver.ErrorResponse
	decodeJSON(t, resp, &body)
	if body.Code != httpserver.CURRENTPASSWORDINCORRECT {
		t.Errorf("code = %q, want %q", body.Code, httpserver.CURRENTPASSWORDINCORRECT)
	}
}

// ──────────────────────────────────────────────
// AuthMeDeleteVerify
// ──────────────────────────────────────────────

func TestMeDeleteVerify_Success(t *testing.T) {
	svc := newMock(t)
	ts := setupAuthServer(t, svc, meGetOps, meDeleteOps, meDeleteVerifyOps, meDeleteResendOps)

	svc.EXPECT().DeleteAccountVerify(gomock.Any(), &authservice.DeleteAccountVerifyInput{
		Code: "123456",
	}).Return(nil)

	resp := authPostJSON(t, ts, "/api/v1/auth/me/delete/verify", map[string]string{
		"code": "123456",
	})
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusNoContent)
	}
}

func TestMeDeleteVerify_MissingAuthToken(t *testing.T) {
	svc := newMock(t)
	ts := setupAuthServer(t, svc, meGetOps, meDeleteOps, meDeleteVerifyOps, meDeleteResendOps)

	resp := postJSON(t, ts, "/api/v1/auth/me/delete/verify", map[string]string{
		"code": "123456",
	})
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusUnauthorized)
	}
}

func TestMeDeleteVerify_InvalidJSON(t *testing.T) {
	svc := newMock(t)
	ts := setupAuthServer(t, svc, meGetOps, meDeleteOps, meDeleteVerifyOps, meDeleteResendOps)

	resp := authPostJSON(t, ts, "/api/v1/auth/me/delete/verify", nil)
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusBadRequest)
	}
}

func TestMeDeleteVerify_EmptyCode(t *testing.T) {
	svc := newMock(t)
	ts := setupAuthServer(t, svc, meGetOps, meDeleteOps, meDeleteVerifyOps, meDeleteResendOps)

	resp := authPostJSON(t, ts, "/api/v1/auth/me/delete/verify", map[string]string{
		"code": "",
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

func TestMeDeleteVerify_InvalidCode(t *testing.T) {
	svc := newMock(t)
	ts := setupAuthServer(t, svc, meGetOps, meDeleteOps, meDeleteVerifyOps, meDeleteResendOps)

	svc.EXPECT().DeleteAccountVerify(gomock.Any(), gomock.Any()).
		Return(authservice.ErrInvalidCode)

	resp := authPostJSON(t, ts, "/api/v1/auth/me/delete/verify", map[string]string{
		"code": "000000",
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

func TestMeDeleteVerify_TooManyAttempts(t *testing.T) {
	svc := newMock(t)
	ts := setupAuthServer(t, svc, meGetOps, meDeleteOps, meDeleteVerifyOps, meDeleteResendOps)

	svc.EXPECT().DeleteAccountVerify(gomock.Any(), gomock.Any()).
		Return(authservice.ErrTooManyAttempts)

	resp := authPostJSON(t, ts, "/api/v1/auth/me/delete/verify", map[string]string{
		"code": "123456",
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
// AuthMeDeleteCodeResend
// ──────────────────────────────────────────────

func TestMeDeleteCodeResend_Success(t *testing.T) {
	svc := newMock(t)
	ts := setupAuthServer(t, svc, meGetOps, meDeleteOps, meDeleteVerifyOps, meDeleteResendOps)

	svc.EXPECT().DeleteAccountCodeResend(gomock.Any()).
		Return("verification code resent", nil)

	resp := authPostJSON(t, ts, "/api/v1/auth/me/delete/code/resend", nil)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusOK)
	}
	var body httpserver.MessageResponse
	decodeJSON(t, resp, &body)
	if body.Message != "verification code resent" {
		t.Errorf("message = %q, want verification code resent", body.Message)
	}
}

func TestMeDeleteCodeResend_MissingAuthToken(t *testing.T) {
	svc := newMock(t)
	ts := setupAuthServer(t, svc, meGetOps, meDeleteOps, meDeleteVerifyOps, meDeleteResendOps)

	resp := postJSON(t, ts, "/api/v1/auth/me/delete/code/resend", nil)
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusUnauthorized)
	}
}

func TestMeDeleteCodeResend_ServiceError(t *testing.T) {
	svc := newMock(t)
	ts := setupAuthServer(t, svc, meGetOps, meDeleteOps, meDeleteVerifyOps, meDeleteResendOps)

	svc.EXPECT().DeleteAccountCodeResend(gomock.Any()).
		Return("", errors.New("db down"))

	resp := authPostJSON(t, ts, "/api/v1/auth/me/delete/code/resend", nil)
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusInternalServerError)
	}
	var body httpserver.ErrorResponse
	decodeJSON(t, resp, &body)
	if body.Code != httpserver.INTERNALERROR {
		t.Errorf("code = %q, want %q", body.Code, httpserver.INTERNALERROR)
	}
}
