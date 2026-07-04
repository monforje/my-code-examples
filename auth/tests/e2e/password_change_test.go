package e2e_test

import (
	"auth/internal/events"
	e2ehelpers "auth/tests/e2e/helpers"
	"net/http"
	"testing"
)

func TestPasswordChange_FullFlow(t *testing.T) {
	resetE2E(t)

	capture := e2ehelpers.NewEventCapture(e2eEnv.NC())
	defer capture.Close()

	user, accessToken := e2ehelpers.LoginAs(t, e2eEnv, client)
	newPassword := "newpass1word"
	identity := e2ehelpers.GetIdentityByEmail(t, e2eEnv.PgPool(), user.Email)
	credBefore := e2ehelpers.GetCredentialsByIdentityID(t, e2eEnv.PgPool(), identity.ID)

	code := e2ehelpers.WaitForCode(t, e2eEnv, e2ehelpers.SubjectPasswordChangeCode, user.Email, e2ehelpers.PurposePasswordChange, func() {
		resp := client.PostAuthJSON(t, accessToken, "/auth/password/change", map[string]string{"current_password": user.Password})
		e2ehelpers.ExpectStatus(t, resp, http.StatusOK)
	})

	vc := e2ehelpers.GetVerificationCodesByIdentityID(t, e2eEnv.PgPool(), identity.ID)
	var pcVC e2ehelpers.VerificationCodeRow
	for _, v := range vc {
		if v.Purpose == "password_change" {
			pcVC = v
			break
		}
	}
	if pcVC.ID == "" {
		t.Fatal("db verification_code with purpose=password_change not found after initiate")
	}
	if pcVC.ConsumedAt != nil {
		t.Fatal("db verification_code consumed_at should be nil after initiate")
	}

	pcEvents := e2ehelpers.GetAuthEventsByType(t, e2eEnv.PgPool(), identity.ID, "password_change_started")
	if len(pcEvents) != 1 {
		t.Fatalf("db password_change_started events = %d, want 1", len(pcEvents))
	}

	capture.AssertPublished(t, events.EventPasswordChangeCodeSend)

	verifyResp := client.PostAuthJSON(t, accessToken, "/auth/password/change/verify", map[string]string{"code": code})
	e2ehelpers.ExpectStatus(t, verifyResp, http.StatusOK)
	changeToken := e2ehelpers.Decode[e2ehelpers.ChangePasswordTokenResponse](t, verifyResp)
	if changeToken.ChangeToken == "" {
		t.Fatal("change_token is empty")
	}

	vcAfter := e2ehelpers.GetVerificationCodesByIdentityID(t, e2eEnv.PgPool(), identity.ID)
	for _, v := range vcAfter {
		if v.Purpose == "password_change" && v.ConsumedAt == nil {
			t.Fatal("db password_change verification_code consumed_at should NOT be nil after verify")
		}
	}

	pct := e2ehelpers.GetPasswordChangeTokensByIdentityID(t, e2eEnv.PgPool(), identity.ID)
	if len(pct) != 1 {
		t.Fatalf("db password_change_tokens count = %d, want 1", len(pct))
	}
	if pct[0].TokenHash == "" {
		t.Fatal("db password_change_token token_hash is empty")
	}
	if pct[0].ConsumedAt != nil {
		t.Fatal("db password_change_token consumed_at should be nil before complete")
	}

	pcVerified := e2ehelpers.GetAuthEventsByType(t, e2eEnv.PgPool(), identity.ID, "password_change_verified")
	if len(pcVerified) != 1 {
		t.Fatalf("db password_change_verified events = %d, want 1", len(pcVerified))
	}

	completeResp := client.PostAuthJSON(t, accessToken, "/auth/password/change/complete", map[string]string{
		"change_token": changeToken.ChangeToken,
		"new_password": newPassword,
	})
	e2ehelpers.ExpectStatus(t, completeResp, http.StatusOK)

	credAfter := e2ehelpers.GetCredentialsByIdentityID(t, e2eEnv.PgPool(), identity.ID)
	if credAfter.PasswordHash == credBefore.PasswordHash {
		t.Fatal("db password_hash should have changed after complete")
	}
	if credAfter.PasswordChangedAt.IsZero() {
		t.Fatal("db password_changed_at should not be zero after complete")
	}

	pctAfter := e2ehelpers.GetPasswordChangeTokensByIdentityID(t, e2eEnv.PgPool(), identity.ID)
	if len(pctAfter) != 1 {
		t.Fatalf("db password_change_tokens count = %d, want 1 after complete", len(pctAfter))
	}
	if pctAfter[0].ConsumedAt == nil {
		t.Fatal("db password_change_token consumed_at should NOT be nil after complete")
	}

	sessions := e2ehelpers.GetSessionsByIdentityID(t, e2eEnv.PgPool(), identity.ID)
	for _, s := range sessions {
		if s.RevokedAt == nil {
			t.Fatalf("db session %s revoked_at should NOT be nil after password change", s.ID)
		}
	}

	pcChanged := e2ehelpers.GetAuthEventsByType(t, e2eEnv.PgPool(), identity.ID, "password_changed")
	if len(pcChanged) != 1 {
		t.Fatalf("db password_changed events = %d, want 1", len(pcChanged))
	}

	capture.AssertPublished(t, events.EventIdentityUpdated)

	oldLogin := client.PostJSON(t, "/auth/login", map[string]string{"email": user.Email, "password": user.Password})
	e2ehelpers.ExpectStatus(t, oldLogin, http.StatusUnauthorized)

	newLogin := client.PostJSON(t, "/auth/login", map[string]string{"email": user.Email, "password": newPassword})
	e2ehelpers.ExpectStatus(t, newLogin, http.StatusOK)
}
