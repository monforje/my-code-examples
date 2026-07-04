package handlers_test

import (
	"errors"
	"net/http"
	"testing"

	"go.uber.org/mock/gomock"

	httpserver "auth/internal/http/gen"
	authservice "auth/internal/services/auth"
)

const (
	changePasswordOps = "auth.password.change"
	changeVerifyOps   = "auth.password.change.verify"
	changeCompleteOps = "auth.password.change.complete"
	changeResendOps   = "auth.password.change.code.resend"
)

// ──────────────────────────────────────────────
// AuthPasswordChange
// ──────────────────────────────────────────────

func TestPasswordChange_Success(t *testing.T) {
	svc := newMock(t)
	ts := setupAuthServer(t, svc, changePasswordOps, changeVerifyOps, changeCompleteOps, changeResendOps)

	svc.EXPECT().ChangePassword(gomock.Any(), &authservice.ChangePasswordInput{
		CurrentPassword: "oldpass1word",
	}).Return("verification code sent", nil)

	resp := authPostJSON(t, ts, "/api/v1/auth/password/change", map[string]string{
		"current_password": "oldpass1word",
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

func TestPasswordChange_MissingAuthToken(t *testing.T) {
	svc := newMock(t)
	ts := setupAuthServer(t, svc, changePasswordOps, changeVerifyOps, changeCompleteOps, changeResendOps)

	resp := postJSON(t, ts, "/api/v1/auth/password/change", map[string]string{
		"current_password": "oldpass1word",
	})
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusUnauthorized)
	}
}

func TestPasswordChange_InvalidJSON(t *testing.T) {
	svc := newMock(t)
	ts := setupAuthServer(t, svc, changePasswordOps, changeVerifyOps, changeCompleteOps, changeResendOps)

	resp := authPostJSON(t, ts, "/api/v1/auth/password/change", nil)
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusBadRequest)
	}
}

func TestPasswordChange_EmptyPassword(t *testing.T) {
	svc := newMock(t)
	ts := setupAuthServer(t, svc, changePasswordOps, changeVerifyOps, changeCompleteOps, changeResendOps)

	resp := authPostJSON(t, ts, "/api/v1/auth/password/change", map[string]string{
		"current_password": "",
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

func TestPasswordChange_ShortPassword(t *testing.T) {
	svc := newMock(t)
	ts := setupAuthServer(t, svc, changePasswordOps, changeVerifyOps, changeCompleteOps, changeResendOps)

	resp := authPostJSON(t, ts, "/api/v1/auth/password/change", map[string]string{
		"current_password": "1",
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

func TestPasswordChange_IncorrectPassword(t *testing.T) {
	svc := newMock(t)
	ts := setupAuthServer(t, svc, changePasswordOps, changeVerifyOps, changeCompleteOps, changeResendOps)

	svc.EXPECT().ChangePassword(gomock.Any(), gomock.Any()).
		Return("", authservice.ErrCurrentPasswordIncorrect)

	resp := authPostJSON(t, ts, "/api/v1/auth/password/change", map[string]string{
		"current_password": "wrongpass1",
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
// AuthPasswordChangeVerify
// ──────────────────────────────────────────────

func TestPasswordChangeVerify_Success(t *testing.T) {
	svc := newMock(t)
	ts := setupAuthServer(t, svc, changePasswordOps, changeVerifyOps, changeCompleteOps, changeResendOps)

	svc.EXPECT().ChangePasswordVerify(gomock.Any(), &authservice.ChangePasswordVerifyInput{
		Code: "123456",
	}).Return(&authservice.ChangePasswordVerifyOutput{
		ChangeToken: "change-token-abc",
		ExpiresIn:   1800,
	}, nil)

	resp := authPostJSON(t, ts, "/api/v1/auth/password/change/verify", map[string]string{
		"code": "123456",
	})
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusOK)
	}
	var body httpserver.ChangePasswordTokenResponse
	decodeJSON(t, resp, &body)
	if body.ChangeToken != "change-token-abc" {
		t.Errorf("change_token = %q, want change-token-abc", body.ChangeToken)
	}
	if body.ExpiresIn != 1800 {
		t.Errorf("expires_in = %d, want 1800", body.ExpiresIn)
	}
}

func TestPasswordChangeVerify_MissingAuthToken(t *testing.T) {
	svc := newMock(t)
	ts := setupAuthServer(t, svc, changePasswordOps, changeVerifyOps, changeCompleteOps, changeResendOps)

	resp := postJSON(t, ts, "/api/v1/auth/password/change/verify", map[string]string{
		"code": "123456",
	})
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusUnauthorized)
	}
}

func TestPasswordChangeVerify_InvalidJSON(t *testing.T) {
	svc := newMock(t)
	ts := setupAuthServer(t, svc, changePasswordOps, changeVerifyOps, changeCompleteOps, changeResendOps)

	resp := authPostJSON(t, ts, "/api/v1/auth/password/change/verify", nil)
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusBadRequest)
	}
}

func TestPasswordChangeVerify_EmptyCode(t *testing.T) {
	svc := newMock(t)
	ts := setupAuthServer(t, svc, changePasswordOps, changeVerifyOps, changeCompleteOps, changeResendOps)

	resp := authPostJSON(t, ts, "/api/v1/auth/password/change/verify", map[string]string{
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

func TestPasswordChangeVerify_InvalidCode(t *testing.T) {
	svc := newMock(t)
	ts := setupAuthServer(t, svc, changePasswordOps, changeVerifyOps, changeCompleteOps, changeResendOps)

	svc.EXPECT().ChangePasswordVerify(gomock.Any(), gomock.Any()).
		Return(nil, authservice.ErrInvalidCode)

	resp := authPostJSON(t, ts, "/api/v1/auth/password/change/verify", map[string]string{
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

func TestPasswordChangeVerify_TooManyAttempts(t *testing.T) {
	svc := newMock(t)
	ts := setupAuthServer(t, svc, changePasswordOps, changeVerifyOps, changeCompleteOps, changeResendOps)

	svc.EXPECT().ChangePasswordVerify(gomock.Any(), gomock.Any()).
		Return(nil, authservice.ErrTooManyAttempts)

	resp := authPostJSON(t, ts, "/api/v1/auth/password/change/verify", map[string]string{
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
// AuthPasswordChangeComplete
// ──────────────────────────────────────────────

func TestPasswordChangeComplete_Success(t *testing.T) {
	svc := newMock(t)
	ts := setupAuthServer(t, svc, changePasswordOps, changeVerifyOps, changeCompleteOps, changeResendOps)

	svc.EXPECT().CompletePasswordChange(gomock.Any(), &authservice.CompletePasswordChangeInput{
		ChangeToken: "change-token-abc",
		NewPassword: "newpass1word",
	}).Return("password changed", nil)

	resp := authPostJSON(t, ts, "/api/v1/auth/password/change/complete", map[string]string{
		"change_token": "change-token-abc",
		"new_password": "newpass1word",
	})
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusOK)
	}
	var body httpserver.MessageResponse
	decodeJSON(t, resp, &body)
	if body.Message != "password changed" {
		t.Errorf("message = %q, want password changed", body.Message)
	}
}

func TestPasswordChangeComplete_MissingAuthToken(t *testing.T) {
	svc := newMock(t)
	ts := setupAuthServer(t, svc, changePasswordOps, changeVerifyOps, changeCompleteOps, changeResendOps)

	resp := postJSON(t, ts, "/api/v1/auth/password/change/complete", map[string]string{
		"change_token": "change-token-abc",
		"new_password": "newpass1word",
	})
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusUnauthorized)
	}
}

func TestPasswordChangeComplete_InvalidJSON(t *testing.T) {
	svc := newMock(t)
	ts := setupAuthServer(t, svc, changePasswordOps, changeVerifyOps, changeCompleteOps, changeResendOps)

	resp := authPostJSON(t, ts, "/api/v1/auth/password/change/complete", nil)
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusBadRequest)
	}
}

func TestPasswordChangeComplete_EmptyToken(t *testing.T) {
	svc := newMock(t)
	ts := setupAuthServer(t, svc, changePasswordOps, changeVerifyOps, changeCompleteOps, changeResendOps)

	resp := authPostJSON(t, ts, "/api/v1/auth/password/change/complete", map[string]string{
		"change_token": "",
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

func TestPasswordChangeComplete_ShortPassword(t *testing.T) {
	svc := newMock(t)
	ts := setupAuthServer(t, svc, changePasswordOps, changeVerifyOps, changeCompleteOps, changeResendOps)

	resp := authPostJSON(t, ts, "/api/v1/auth/password/change/complete", map[string]string{
		"change_token": "change-token-abc",
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

func TestPasswordChangeComplete_InvalidChangeToken(t *testing.T) {
	svc := newMock(t)
	ts := setupAuthServer(t, svc, changePasswordOps, changeVerifyOps, changeCompleteOps, changeResendOps)

	svc.EXPECT().CompletePasswordChange(gomock.Any(), gomock.Any()).
		Return("", authservice.ErrInvalidChangeToken)

	resp := authPostJSON(t, ts, "/api/v1/auth/password/change/complete", map[string]string{
		"change_token": "bad-token",
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

// ──────────────────────────────────────────────
// AuthPasswordChangeCodeResend
// ──────────────────────────────────────────────

func TestPasswordChangeCodeResend_Success(t *testing.T) {
	svc := newMock(t)
	ts := setupAuthServer(t, svc, changePasswordOps, changeVerifyOps, changeCompleteOps, changeResendOps)

	svc.EXPECT().ChangePasswordCodeResend(gomock.Any()).
		Return("verification code sent", nil)

	resp := authPostJSON(t, ts, "/api/v1/auth/password/change/code/resend", nil)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusOK)
	}
	var body httpserver.MessageResponse
	decodeJSON(t, resp, &body)
	if body.Message != "verification code sent" {
		t.Errorf("message = %q, want verification code sent", body.Message)
	}
}

func TestPasswordChangeCodeResend_MissingAuthToken(t *testing.T) {
	svc := newMock(t)
	ts := setupAuthServer(t, svc, changePasswordOps, changeVerifyOps, changeCompleteOps, changeResendOps)

	resp := postJSON(t, ts, "/api/v1/auth/password/change/code/resend", nil)
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusUnauthorized)
	}
}

func TestPasswordChangeCodeResend_ServiceError(t *testing.T) {
	svc := newMock(t)
	ts := setupAuthServer(t, svc, changePasswordOps, changeVerifyOps, changeCompleteOps, changeResendOps)

	svc.EXPECT().ChangePasswordCodeResend(gomock.Any()).
		Return("", errors.New("db down"))

	resp := authPostJSON(t, ts, "/api/v1/auth/password/change/code/resend", nil)
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusInternalServerError)
	}
	var body httpserver.ErrorResponse
	decodeJSON(t, resp, &body)
	if body.Code != httpserver.INTERNALERROR {
		t.Errorf("code = %q, want %q", body.Code, httpserver.INTERNALERROR)
	}
}
