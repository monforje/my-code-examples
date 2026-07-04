package e2e_test

import (
	"auth/internal/events"
	e2ehelpers "auth/tests/e2e/helpers"
	"net/http"
	"testing"
	"time"
)

func TestEmailChange_FullFlow(t *testing.T) {
	resetE2E(t)

	capture := e2ehelpers.NewEventCapture(e2eEnv.NC())
	defer capture.Close()

	user, accessToken := e2ehelpers.LoginAs(t, e2eEnv, client)
	newEmail := "changed-" + time.Now().Format("20060102150405.000000000") + "@example.com"
	identity := e2ehelpers.GetIdentityByEmail(t, e2eEnv.PgPool(), user.Email)

	currentCode := e2ehelpers.WaitForCode(t, e2eEnv, e2ehelpers.SubjectEmailChangeCode, user.Email, e2ehelpers.PurposeEmailChangeCurrent, func() {
		resp := client.PostAuthJSON(t, accessToken, "/auth/me/email/change", map[string]string{"password": user.Password})
		e2ehelpers.ExpectStatus(t, resp, http.StatusOK)
	})

	ecr := e2ehelpers.GetEmailChangeRequestsByIdentityID(t, e2eEnv.PgPool(), identity.ID)
	if len(ecr) != 1 {
		t.Fatalf("db email_change_requests count = %d, want 1 after initiate", len(ecr))
	}
	if ecr[0].Status != "pending" {
		t.Fatalf("db email_change_request status = %q, want pending", ecr[0].Status)
	}

	vc := e2ehelpers.GetVerificationCodesByIdentityID(t, e2eEnv.PgPool(), identity.ID)
	var ecCurrentVC e2ehelpers.VerificationCodeRow
	for _, v := range vc {
		if v.Purpose == "email_change_current" {
			ecCurrentVC = v
			break
		}
	}
	if ecCurrentVC.ID == "" {
		t.Fatal("db verification_code with purpose=email_change_current not found after initiate")
	}

	ecStarted := e2ehelpers.GetAuthEventsByType(t, e2eEnv.PgPool(), identity.ID, "email_change_started")
	if len(ecStarted) != 1 {
		t.Fatalf("db email_change_started events = %d, want 1", len(ecStarted))
	}

	capture.AssertPublished(t, events.EventEmailChangeCodeSend)

	verifyResp := client.PostAuthJSON(t, accessToken, "/auth/me/email/change/verify", map[string]string{"code": currentCode})
	e2ehelpers.ExpectStatus(t, verifyResp, http.StatusOK)
	identityToken := e2ehelpers.Decode[e2ehelpers.IdentityTokenResponse](t, verifyResp)
	if identityToken.IdentityToken == "" {
		t.Fatal("identity_token is empty")
	}

	vcAfter := e2ehelpers.GetVerificationCodesByIdentityID(t, e2eEnv.PgPool(), identity.ID)
	for _, v := range vcAfter {
		if v.Purpose == "email_change_current" && v.ConsumedAt == nil {
			t.Fatal("db email_change_current verification_code consumed_at should NOT be nil after verify")
		}
	}

	ecrAfter := e2ehelpers.GetEmailChangeRequestsByIdentityID(t, e2eEnv.PgPool(), identity.ID)
	if len(ecrAfter) != 1 {
		t.Fatalf("db email_change_requests count = %d, want 1 after verify", len(ecrAfter))
	}
	if ecrAfter[0].Status != "verified" {
		t.Fatalf("db email_change_request status = %q, want verified after verify", ecrAfter[0].Status)
	}
	if ecrAfter[0].TokenHash == nil || *ecrAfter[0].TokenHash == "" {
		t.Fatal("db email_change_request token_hash should not be empty after verify")
	}

	ecVerified := e2ehelpers.GetAuthEventsByType(t, e2eEnv.PgPool(), identity.ID, "email_change_verified")
	if len(ecVerified) != 1 {
		t.Fatalf("db email_change_verified events = %d, want 1", len(ecVerified))
	}

	newCode := e2ehelpers.WaitForCode(t, e2eEnv, e2ehelpers.SubjectEmailChangeCode, newEmail, e2ehelpers.PurposeEmailChangeNew, func() {
		resp := client.PostAuthJSON(t, accessToken, "/auth/me/email/change/confirm", map[string]string{
			"new_email":      newEmail,
			"identity_token": identityToken.IdentityToken,
		})
		e2ehelpers.ExpectStatus(t, resp, http.StatusOK)
	})

	ecrConfirm := e2ehelpers.GetEmailChangeRequestsByIdentityID(t, e2eEnv.PgPool(), identity.ID)
	if len(ecrConfirm) != 1 {
		t.Fatalf("db email_change_requests count = %d, want 1 after confirm", len(ecrConfirm))
	}
	if ecrConfirm[0].Status != "confirming" {
		t.Fatalf("db email_change_request status = %q, want confirming after confirm", ecrConfirm[0].Status)
	}
	if ecrConfirm[0].NewEmail != newEmail {
		t.Fatalf("db email_change_request new_email = %q, want %q", ecrConfirm[0].NewEmail, newEmail)
	}

	vcAll := e2ehelpers.GetVerificationCodesByIdentityID(t, e2eEnv.PgPool(), identity.ID)
	var ecNewVC e2ehelpers.VerificationCodeRow
	for _, v := range vcAll {
		if v.Purpose == "email_change_new" {
			ecNewVC = v
			break
		}
	}
	if ecNewVC.ID == "" {
		t.Fatal("db verification_code with purpose=email_change_new not found after confirm")
	}
	if ecNewVC.ConsumedAt != nil {
		t.Fatal("db email_change_new verification_code consumed_at should be nil after confirm")
	}

	ecConfirmed := e2ehelpers.GetAuthEventsByType(t, e2eEnv.PgPool(), identity.ID, "email_change_confirmed")
	if len(ecConfirmed) != 1 {
		t.Fatalf("db email_change_confirmed events = %d, want 1", len(ecConfirmed))
	}

	completeResp := client.PostAuthJSON(t, accessToken, "/auth/me/email/change/complete", map[string]string{"code": newCode})
	e2ehelpers.ExpectStatus(t, completeResp, http.StatusOK)

	meAfter := e2ehelpers.GetIdentityByID(t, e2eEnv.PgPool(), identity.ID)
	if meAfter.Email != newEmail {
		t.Fatalf("db email = %q, want %q after complete", meAfter.Email, newEmail)
	}

	ecrFinal := e2ehelpers.GetEmailChangeRequestsByIdentityID(t, e2eEnv.PgPool(), identity.ID)
	if len(ecrFinal) != 1 {
		t.Fatalf("db email_change_requests count = %d, want 1 after complete", len(ecrFinal))
	}
	if ecrFinal[0].ConsumedAt == nil {
		t.Fatal("db email_change_request consumed_at should NOT be nil after complete")
	}

	vcFinal := e2ehelpers.GetVerificationCodesByIdentityID(t, e2eEnv.PgPool(), identity.ID)
	for _, v := range vcFinal {
		if (v.Purpose == "email_change_current" || v.Purpose == "email_change_new") && v.ConsumedAt == nil {
			t.Fatalf("db verification_code purpose=%s consumed_at should NOT be nil after complete", v.Purpose)
		}
	}

	sessions := e2ehelpers.GetSessionsByIdentityID(t, e2eEnv.PgPool(), identity.ID)
	for _, s := range sessions {
		if s.RevokedAt == nil {
			t.Fatalf("db session %s revoked_at should NOT be nil after email change", s.ID)
		}
	}

	ecChanged := e2ehelpers.GetAuthEventsByType(t, e2eEnv.PgPool(), identity.ID, "email_changed")
	if len(ecChanged) != 1 {
		t.Fatalf("db email_changed events = %d, want 1", len(ecChanged))
	}

	capture.AssertPublished(t, events.EventIdentityUpdated)

	oldLogin := client.PostJSON(t, "/auth/login", map[string]string{"email": user.Email, "password": user.Password})
	e2ehelpers.ExpectStatus(t, oldLogin, http.StatusUnauthorized)

	newLogin := client.PostJSON(t, "/auth/login", map[string]string{"email": newEmail, "password": user.Password})
	e2ehelpers.ExpectStatus(t, newLogin, http.StatusOK)
	newToken := e2ehelpers.Decode[e2ehelpers.TokenResponse](t, newLogin)

	meResp := client.GetAuth(t, newToken.AccessToken, "/auth/me")
	e2ehelpers.ExpectStatus(t, meResp, http.StatusOK)
	me := e2ehelpers.Decode[e2ehelpers.IdentityResponse](t, meResp)
	if me.Email != newEmail {
		t.Fatalf("email = %q, want %q", me.Email, newEmail)
	}
}
