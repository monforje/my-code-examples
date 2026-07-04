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
	changeEmailOps         = "auth.me.email.change"
	changeEmailVerifyOps   = "auth.me.email.change.verify"
	changeEmailConfirmOps  = "auth.me.email.change.confirm"
	changeEmailCompleteOps = "auth.me.email.change.complete"
	changeEmailResendOps   = "auth.me.email.change.code.resend"
)

// ──────────────────────────────────────────────
// AuthMeEmailChange
// ──────────────────────────────────────────────

func TestEmailChange_Success(t *testing.T) {
	svc := newMock(t)
	ts := setupAuthServer(t, svc, changeEmailOps, changeEmailVerifyOps, changeEmailConfirmOps, changeEmailCompleteOps, changeEmailResendOps)

	svc.EXPECT().ChangeEmail(gomock.Any(), &authservice.ChangeEmailInput{
		Password: "oldpass1word",
	}).Return("verification code sent", nil)

	resp := authPostJSON(t, ts, "/api/v1/auth/me/email/change", map[string]string{
		"password": "oldpass1word",
	})
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusOK)
	}
	var body httpserver.MessageResponse
	decodeJSON(t, resp, &body)
	if body.Message != "verification code sent" {
		t.Errorf("message = %q, want verification code sent", body.Message)
	}
}

func TestEmailChange_MissingAuthToken(t *testing.T) {
	svc := newMock(t)
	ts := setupAuthServer(t, svc, changeEmailOps, changeEmailVerifyOps, changeEmailConfirmOps, changeEmailCompleteOps, changeEmailResendOps)

	resp := postJSON(t, ts, "/api/v1/auth/me/email/change", map[string]string{
		"password": "oldpass1word",
	})
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusUnauthorized)
	}
}

func TestEmailChange_InvalidJSON(t *testing.T) {
	svc := newMock(t)
	ts := setupAuthServer(t, svc, changeEmailOps, changeEmailVerifyOps, changeEmailConfirmOps, changeEmailCompleteOps, changeEmailResendOps)

	resp := authPostJSON(t, ts, "/api/v1/auth/me/email/change", nil)
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusBadRequest)
	}
}

func TestEmailChange_EmptyPassword(t *testing.T) {
	svc := newMock(t)
	ts := setupAuthServer(t, svc, changeEmailOps, changeEmailVerifyOps, changeEmailConfirmOps, changeEmailCompleteOps, changeEmailResendOps)

	resp := authPostJSON(t, ts, "/api/v1/auth/me/email/change", map[string]string{
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

func TestEmailChange_ShortPassword(t *testing.T) {
	svc := newMock(t)
	ts := setupAuthServer(t, svc, changeEmailOps, changeEmailVerifyOps, changeEmailConfirmOps, changeEmailCompleteOps, changeEmailResendOps)

	resp := authPostJSON(t, ts, "/api/v1/auth/me/email/change", map[string]string{
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

func TestEmailChange_IncorrectPassword(t *testing.T) {
	svc := newMock(t)
	ts := setupAuthServer(t, svc, changeEmailOps, changeEmailVerifyOps, changeEmailConfirmOps, changeEmailCompleteOps, changeEmailResendOps)

	svc.EXPECT().ChangeEmail(gomock.Any(), gomock.Any()).
		Return("", authservice.ErrCurrentPasswordIncorrect)

	resp := authPostJSON(t, ts, "/api/v1/auth/me/email/change", map[string]string{
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
// AuthMeEmailChangeVerify
// ──────────────────────────────────────────────

func TestEmailChangeVerify_Success(t *testing.T) {
	svc := newMock(t)
	ts := setupAuthServer(t, svc, changeEmailOps, changeEmailVerifyOps, changeEmailConfirmOps, changeEmailCompleteOps, changeEmailResendOps)

	svc.EXPECT().ChangeEmailVerify(gomock.Any(), &authservice.ChangeEmailVerifyInput{
		Code: "123456",
	}).Return(&authservice.ChangeEmailVerifyOutput{
		IdentityToken: "identity-token-abc",
		ExpiresIn:     1800,
	}, nil)

	resp := authPostJSON(t, ts, "/api/v1/auth/me/email/change/verify", map[string]string{
		"code": "123456",
	})
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusOK)
	}
	var body httpserver.IdentityTokenResponse
	decodeJSON(t, resp, &body)
	if body.IdentityToken != "identity-token-abc" {
		t.Errorf("identity_token = %q, want identity-token-abc", body.IdentityToken)
	}
	if body.ExpiresIn != 1800 {
		t.Errorf("expires_in = %d, want 1800", body.ExpiresIn)
	}
}

func TestEmailChangeVerify_MissingAuthToken(t *testing.T) {
	svc := newMock(t)
	ts := setupAuthServer(t, svc, changeEmailOps, changeEmailVerifyOps, changeEmailConfirmOps, changeEmailCompleteOps, changeEmailResendOps)

	resp := postJSON(t, ts, "/api/v1/auth/me/email/change/verify", map[string]string{
		"code": "123456",
	})
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusUnauthorized)
	}
}

func TestEmailChangeVerify_InvalidJSON(t *testing.T) {
	svc := newMock(t)
	ts := setupAuthServer(t, svc, changeEmailOps, changeEmailVerifyOps, changeEmailConfirmOps, changeEmailCompleteOps, changeEmailResendOps)

	resp := authPostJSON(t, ts, "/api/v1/auth/me/email/change/verify", nil)
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusBadRequest)
	}
}

func TestEmailChangeVerify_EmptyCode(t *testing.T) {
	svc := newMock(t)
	ts := setupAuthServer(t, svc, changeEmailOps, changeEmailVerifyOps, changeEmailConfirmOps, changeEmailCompleteOps, changeEmailResendOps)

	resp := authPostJSON(t, ts, "/api/v1/auth/me/email/change/verify", map[string]string{
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

func TestEmailChangeVerify_InvalidCode(t *testing.T) {
	svc := newMock(t)
	ts := setupAuthServer(t, svc, changeEmailOps, changeEmailVerifyOps, changeEmailConfirmOps, changeEmailCompleteOps, changeEmailResendOps)

	svc.EXPECT().ChangeEmailVerify(gomock.Any(), gomock.Any()).
		Return(nil, authservice.ErrInvalidCode)

	resp := authPostJSON(t, ts, "/api/v1/auth/me/email/change/verify", map[string]string{
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

func TestEmailChangeVerify_TooManyAttempts(t *testing.T) {
	svc := newMock(t)
	ts := setupAuthServer(t, svc, changeEmailOps, changeEmailVerifyOps, changeEmailConfirmOps, changeEmailCompleteOps, changeEmailResendOps)

	svc.EXPECT().ChangeEmailVerify(gomock.Any(), gomock.Any()).
		Return(nil, authservice.ErrTooManyAttempts)

	resp := authPostJSON(t, ts, "/api/v1/auth/me/email/change/verify", map[string]string{
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
// AuthMeEmailChangeConfirm
// ──────────────────────────────────────────────

func TestEmailChangeConfirm_Success(t *testing.T) {
	svc := newMock(t)
	ts := setupAuthServer(t, svc, changeEmailOps, changeEmailVerifyOps, changeEmailConfirmOps, changeEmailCompleteOps, changeEmailResendOps)

	svc.EXPECT().ChangeEmailConfirm(gomock.Any(), &authservice.ChangeEmailConfirmInput{
		NewEmail:      "new@example.com",
		IdentityToken: "identity-token-abc",
	}).Return("confirmation sent", nil)

	resp := authPostJSON(t, ts, "/api/v1/auth/me/email/change/confirm", map[string]string{
		"new_email":      "new@example.com",
		"identity_token": "identity-token-abc",
	})
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusOK)
	}
	var body httpserver.MessageResponse
	decodeJSON(t, resp, &body)
	if body.Message != "confirmation sent" {
		t.Errorf("message = %q, want confirmation sent", body.Message)
	}
}

func TestEmailChangeConfirm_MissingAuthToken(t *testing.T) {
	svc := newMock(t)
	ts := setupAuthServer(t, svc, changeEmailOps, changeEmailVerifyOps, changeEmailConfirmOps, changeEmailCompleteOps, changeEmailResendOps)

	resp := postJSON(t, ts, "/api/v1/auth/me/email/change/confirm", map[string]string{
		"new_email":      "new@example.com",
		"identity_token": "identity-token-abc",
	})
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusUnauthorized)
	}
}

func TestEmailChangeConfirm_InvalidJSON(t *testing.T) {
	svc := newMock(t)
	ts := setupAuthServer(t, svc, changeEmailOps, changeEmailVerifyOps, changeEmailConfirmOps, changeEmailCompleteOps, changeEmailResendOps)

	resp := authPostJSON(t, ts, "/api/v1/auth/me/email/change/confirm", nil)
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusBadRequest)
	}
}

func TestEmailChangeConfirm_EmptyEmail(t *testing.T) {
	svc := newMock(t)
	ts := setupAuthServer(t, svc, changeEmailOps, changeEmailVerifyOps, changeEmailConfirmOps, changeEmailCompleteOps, changeEmailResendOps)

	resp := authPostJSON(t, ts, "/api/v1/auth/me/email/change/confirm", map[string]string{
		"new_email":      "",
		"identity_token": "identity-token-abc",
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

func TestEmailChangeConfirm_InvalidEmail(t *testing.T) {
	svc := newMock(t)
	ts := setupAuthServer(t, svc, changeEmailOps, changeEmailVerifyOps, changeEmailConfirmOps, changeEmailCompleteOps, changeEmailResendOps)

	resp := authPostJSON(t, ts, "/api/v1/auth/me/email/change/confirm", map[string]string{
		"new_email":      "not-an-email",
		"identity_token": "identity-token-abc",
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

func TestEmailChangeConfirm_EmptyToken(t *testing.T) {
	svc := newMock(t)
	ts := setupAuthServer(t, svc, changeEmailOps, changeEmailVerifyOps, changeEmailConfirmOps, changeEmailCompleteOps, changeEmailResendOps)

	resp := authPostJSON(t, ts, "/api/v1/auth/me/email/change/confirm", map[string]string{
		"new_email":      "new@example.com",
		"identity_token": "",
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

func TestEmailChangeConfirm_EmailAlreadyExists(t *testing.T) {
	svc := newMock(t)
	ts := setupAuthServer(t, svc, changeEmailOps, changeEmailVerifyOps, changeEmailConfirmOps, changeEmailCompleteOps, changeEmailResendOps)

	svc.EXPECT().ChangeEmailConfirm(gomock.Any(), gomock.Any()).
		Return("", authservice.ErrEmailAlreadyExists)

	resp := authPostJSON(t, ts, "/api/v1/auth/me/email/change/confirm", map[string]string{
		"new_email":      "existing@example.com",
		"identity_token": "identity-token-abc",
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

func TestEmailChangeConfirm_InvalidToken(t *testing.T) {
	svc := newMock(t)
	ts := setupAuthServer(t, svc, changeEmailOps, changeEmailVerifyOps, changeEmailConfirmOps, changeEmailCompleteOps, changeEmailResendOps)

	svc.EXPECT().ChangeEmailConfirm(gomock.Any(), gomock.Any()).
		Return("", authservice.ErrInvalidEmailChangeToken)

	resp := authPostJSON(t, ts, "/api/v1/auth/me/email/change/confirm", map[string]string{
		"new_email":      "new@example.com",
		"identity_token": "bad-token",
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

func TestEmailChangeConfirm_TokenExpired(t *testing.T) {
	svc := newMock(t)
	ts := setupAuthServer(t, svc, changeEmailOps, changeEmailVerifyOps, changeEmailConfirmOps, changeEmailCompleteOps, changeEmailResendOps)

	svc.EXPECT().ChangeEmailConfirm(gomock.Any(), gomock.Any()).
		Return("", authservice.ErrEmailChangeTokenExpired)

	resp := authPostJSON(t, ts, "/api/v1/auth/me/email/change/confirm", map[string]string{
		"new_email":      "new@example.com",
		"identity_token": "expired-token",
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

// ──────────────────────────────────────────────
// AuthMeEmailChangeComplete
// ──────────────────────────────────────────────

func TestEmailChangeComplete_Success(t *testing.T) {
	svc := newMock(t)
	ts := setupAuthServer(t, svc, changeEmailOps, changeEmailVerifyOps, changeEmailConfirmOps, changeEmailCompleteOps, changeEmailResendOps)

	svc.EXPECT().ChangeEmailComplete(gomock.Any(), &authservice.ChangeEmailCompleteInput{
		Code: "654321",
	}).Return("email changed", nil)

	resp := authPostJSON(t, ts, "/api/v1/auth/me/email/change/complete", map[string]string{
		"code": "654321",
	})
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusOK)
	}
	var body httpserver.MessageResponse
	decodeJSON(t, resp, &body)
	if body.Message != "email changed" {
		t.Errorf("message = %q, want email changed", body.Message)
	}
}

func TestEmailChangeComplete_MissingAuthToken(t *testing.T) {
	svc := newMock(t)
	ts := setupAuthServer(t, svc, changeEmailOps, changeEmailVerifyOps, changeEmailConfirmOps, changeEmailCompleteOps, changeEmailResendOps)

	resp := postJSON(t, ts, "/api/v1/auth/me/email/change/complete", map[string]string{
		"code": "654321",
	})
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusUnauthorized)
	}
}

func TestEmailChangeComplete_InvalidJSON(t *testing.T) {
	svc := newMock(t)
	ts := setupAuthServer(t, svc, changeEmailOps, changeEmailVerifyOps, changeEmailConfirmOps, changeEmailCompleteOps, changeEmailResendOps)

	resp := authPostJSON(t, ts, "/api/v1/auth/me/email/change/complete", nil)
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusBadRequest)
	}
}

func TestEmailChangeComplete_EmptyCode(t *testing.T) {
	svc := newMock(t)
	ts := setupAuthServer(t, svc, changeEmailOps, changeEmailVerifyOps, changeEmailConfirmOps, changeEmailCompleteOps, changeEmailResendOps)

	resp := authPostJSON(t, ts, "/api/v1/auth/me/email/change/complete", map[string]string{
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

func TestEmailChangeComplete_InvalidCode(t *testing.T) {
	svc := newMock(t)
	ts := setupAuthServer(t, svc, changeEmailOps, changeEmailVerifyOps, changeEmailConfirmOps, changeEmailCompleteOps, changeEmailResendOps)

	svc.EXPECT().ChangeEmailComplete(gomock.Any(), gomock.Any()).
		Return("", authservice.ErrInvalidCode)

	resp := authPostJSON(t, ts, "/api/v1/auth/me/email/change/complete", map[string]string{
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

func TestEmailChangeComplete_TooManyAttempts(t *testing.T) {
	svc := newMock(t)
	ts := setupAuthServer(t, svc, changeEmailOps, changeEmailVerifyOps, changeEmailConfirmOps, changeEmailCompleteOps, changeEmailResendOps)

	svc.EXPECT().ChangeEmailComplete(gomock.Any(), gomock.Any()).
		Return("", authservice.ErrTooManyAttempts)

	resp := authPostJSON(t, ts, "/api/v1/auth/me/email/change/complete", map[string]string{
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
// AuthMeEmailChangeCodeResend
// ──────────────────────────────────────────────

func authPostWithQuery(t *testing.T, ts *httptest.Server, path string, query string) *http.Response {
	t.Helper()
	req, err := http.NewRequest(http.MethodPost, ts.URL+path+"?"+query, nil)
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

func TestEmailChangeCodeResend_Success_Current(t *testing.T) {
	svc := newMock(t)
	ts := setupAuthServer(t, svc, changeEmailOps, changeEmailVerifyOps, changeEmailConfirmOps, changeEmailCompleteOps, changeEmailResendOps)

	svc.EXPECT().ChangeEmailCodeResend(gomock.Any(), &authservice.ChangeEmailResendInput{
		Step: "current",
	}).Return("code resent", nil)

	resp := authPostWithQuery(t, ts, "/api/v1/auth/me/email/change/code/resend", "step=current")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusOK)
	}
	var body httpserver.MessageResponse
	decodeJSON(t, resp, &body)
	if body.Message != "code resent" {
		t.Errorf("message = %q, want code resent", body.Message)
	}
}

func TestEmailChangeCodeResend_Success_New(t *testing.T) {
	svc := newMock(t)
	ts := setupAuthServer(t, svc, changeEmailOps, changeEmailVerifyOps, changeEmailConfirmOps, changeEmailCompleteOps, changeEmailResendOps)

	svc.EXPECT().ChangeEmailCodeResend(gomock.Any(), &authservice.ChangeEmailResendInput{
		Step: "new",
	}).Return("code resent", nil)

	resp := authPostWithQuery(t, ts, "/api/v1/auth/me/email/change/code/resend", "step=new")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusOK)
	}
	var body httpserver.MessageResponse
	decodeJSON(t, resp, &body)
	if body.Message != "code resent" {
		t.Errorf("message = %q, want code resent", body.Message)
	}
}

func TestEmailChangeCodeResend_MissingAuthToken(t *testing.T) {
	svc := newMock(t)
	ts := setupAuthServer(t, svc, changeEmailOps, changeEmailVerifyOps, changeEmailConfirmOps, changeEmailCompleteOps, changeEmailResendOps)

	resp := postJSON(t, ts, "/api/v1/auth/me/email/change/code/resend?step=current", nil)
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusUnauthorized)
	}
}

func TestEmailChangeCodeResend_InvalidStep(t *testing.T) {
	svc := newMock(t)
	ts := setupAuthServer(t, svc, changeEmailOps, changeEmailVerifyOps, changeEmailConfirmOps, changeEmailCompleteOps, changeEmailResendOps)

	resp := authPostWithQuery(t, ts, "/api/v1/auth/me/email/change/code/resend", "step=invalid")
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusUnprocessableEntity)
	}
	var body httpserver.ErrorResponse
	decodeJSON(t, resp, &body)
	if body.Code != httpserver.VALIDATIONERROR {
		t.Errorf("code = %q, want %q", body.Code, httpserver.VALIDATIONERROR)
	}
}

func TestEmailChangeCodeResend_MissingStep(t *testing.T) {
	svc := newMock(t)
	ts := setupAuthServer(t, svc, changeEmailOps, changeEmailVerifyOps, changeEmailConfirmOps, changeEmailCompleteOps, changeEmailResendOps)

	resp := authPostWithQuery(t, ts, "/api/v1/auth/me/email/change/code/resend", "")
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusBadRequest)
	}
}

func TestEmailChangeCodeResend_ServiceError(t *testing.T) {
	svc := newMock(t)
	ts := setupAuthServer(t, svc, changeEmailOps, changeEmailVerifyOps, changeEmailConfirmOps, changeEmailCompleteOps, changeEmailResendOps)

	svc.EXPECT().ChangeEmailCodeResend(gomock.Any(), gomock.Any()).
		Return("", errors.New("db down"))

	resp := authPostWithQuery(t, ts, "/api/v1/auth/me/email/change/code/resend", "step=current")
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusInternalServerError)
	}
	var body httpserver.ErrorResponse
	decodeJSON(t, resp, &body)
	if body.Code != httpserver.INTERNALERROR {
		t.Errorf("code = %q, want %q", body.Code, httpserver.INTERNALERROR)
	}
}
