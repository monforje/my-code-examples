package e2e_test

import (
	"auth/internal/events"
	e2ehelpers "auth/tests/e2e/helpers"
	"net/http"
	"testing"
)

func TestPasswordForgot_FullFlow(t *testing.T) {
	resetE2E(t)

	capture := e2ehelpers.NewEventCapture(e2eEnv.NC())
	defer capture.Close()

	user := e2ehelpers.RegisterAndVerify(t, e2eEnv, client)
	newPassword := "resetpass1word"
	identity := e2ehelpers.GetIdentityByEmail(t, e2eEnv.PgPool(), user.Email)
	credBefore := e2ehelpers.GetCredentialsByIdentityID(t, e2eEnv.PgPool(), identity.ID)

	code := e2ehelpers.WaitForCode(t, e2eEnv, e2ehelpers.SubjectPasswordResetCode, user.Email, e2ehelpers.PurposePasswordForgot, func() {
		resp := client.PostJSON(t, "/auth/password/forgot", map[string]string{"email": user.Email})
		e2ehelpers.ExpectStatus(t, resp, http.StatusOK)
	})

	vc := e2ehelpers.GetVerificationCodesByEmailAndPurpose(t, e2eEnv.PgPool(), user.Email, "password_forgot")
	if len(vc) != 1 {
		t.Fatalf("db verification_codes count = %d, want 1 after forgot", len(vc))
	}
	if vc[0].ConsumedAt != nil {
		t.Fatal("db verification_code consumed_at should be nil after forgot")
	}

	pfEvents := e2ehelpers.GetAuthEventsByType(t, e2eEnv.PgPool(), identity.ID, "password_forgot")
	if len(pfEvents) != 1 {
		t.Fatalf("db password_forgot events = %d, want 1", len(pfEvents))
	}

	capture.AssertPublished(t, events.EventPasswordResetCodeSend)

	verifyResp := client.PostJSON(t, "/auth/password/forgot/verify", map[string]string{"email": user.Email, "code": code})
	e2ehelpers.ExpectStatus(t, verifyResp, http.StatusOK)
	resetToken := e2ehelpers.Decode[e2ehelpers.ResetTokenResponse](t, verifyResp)
	if resetToken.ResetToken == "" {
		t.Fatal("reset_token is empty")
	}

	vcAfter := e2ehelpers.GetVerificationCodesByEmailAndPurpose(t, e2eEnv.PgPool(), user.Email, "password_forgot")
	if len(vcAfter) != 1 {
		t.Fatalf("db verification_codes count = %d, want 1 after verify", len(vcAfter))
	}
	if vcAfter[0].ConsumedAt == nil {
		t.Fatal("db verification_code consumed_at should NOT be nil after verify")
	}

	prt := e2ehelpers.GetPasswordResetTokensByIdentityID(t, e2eEnv.PgPool(), identity.ID)
	if len(prt) != 1 {
		t.Fatalf("db password_reset_tokens count = %d, want 1", len(prt))
	}
	if prt[0].TokenHash == "" {
		t.Fatal("db password_reset_token token_hash is empty")
	}
	if prt[0].ConsumedAt != nil {
		t.Fatal("db password_reset_token consumed_at should be nil before reset")
	}

	pfVerified := e2ehelpers.GetAuthEventsByType(t, e2eEnv.PgPool(), identity.ID, "password_forgot_verified")
	if len(pfVerified) != 1 {
		t.Fatalf("db password_forgot_verified events = %d, want 1", len(pfVerified))
	}

	resetResp := client.PostJSON(t, "/auth/password/reset", map[string]string{
		"reset_token":  resetToken.ResetToken,
		"new_password": newPassword,
	})
	e2ehelpers.ExpectStatus(t, resetResp, http.StatusOK)

	credAfter := e2ehelpers.GetCredentialsByIdentityID(t, e2eEnv.PgPool(), identity.ID)
	if credAfter.PasswordHash == credBefore.PasswordHash {
		t.Fatal("db password_hash should have changed after reset")
	}
	if credAfter.PasswordChangedAt.IsZero() {
		t.Fatal("db password_changed_at should not be zero after reset")
	}

	prtAfter := e2ehelpers.GetPasswordResetTokensByIdentityID(t, e2eEnv.PgPool(), identity.ID)
	if len(prtAfter) != 1 {
		t.Fatalf("db password_reset_tokens count = %d, want 1 after reset", len(prtAfter))
	}
	if prtAfter[0].ConsumedAt == nil {
		t.Fatal("db password_reset_token consumed_at should NOT be nil after reset")
	}

	sessions := e2ehelpers.GetSessionsByIdentityID(t, e2eEnv.PgPool(), identity.ID)
	for _, s := range sessions {
		if s.RevokedAt == nil {
			t.Fatalf("db session %s revoked_at should NOT be nil after password reset", s.ID)
		}
	}

	pReset := e2ehelpers.GetAuthEventsByType(t, e2eEnv.PgPool(), identity.ID, "password_reset")
	if len(pReset) != 1 {
		t.Fatalf("db password_reset events = %d, want 1", len(pReset))
	}

	capture.AssertPublished(t, events.EventIdentityUpdated)

	oldLogin := client.PostJSON(t, "/auth/login", map[string]string{"email": user.Email, "password": user.Password})
	e2ehelpers.ExpectStatus(t, oldLogin, http.StatusUnauthorized)

	newLogin := client.PostJSON(t, "/auth/login", map[string]string{"email": user.Email, "password": newPassword})
	e2ehelpers.ExpectStatus(t, newLogin, http.StatusOK)
}
