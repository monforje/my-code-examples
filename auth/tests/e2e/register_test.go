package e2e_test

import (
	"auth/internal/events"
	e2ehelpers "auth/tests/e2e/helpers"
	"net/http"
	"testing"
)

func TestRegister_ValidData(t *testing.T) {
	resetE2E(t)

	user := e2ehelpers.NewUser()
	registered := e2ehelpers.Register(t, client, user)

	if registered.Email != user.Email {
		t.Fatalf("email = %q, want %q", registered.Email, user.Email)
	}
	if registered.IdentityID == "" {
		t.Fatal("identity_id is empty")
	}
	if registered.Status != "pending_verification" {
		t.Fatalf("status = %q, want pending_verification", registered.Status)
	}

	identity := e2ehelpers.GetIdentityByEmail(t, e2eEnv.PgPool(), user.Email)
	if identity.Status != "pending_verification" {
		t.Fatalf("db status = %q, want pending_verification", identity.Status)
	}
	if identity.EmailVerified {
		t.Fatal("db email_verified = true, want false")
	}

	creds := e2ehelpers.GetCredentialsByIdentityID(t, e2eEnv.PgPool(), identity.ID)
	if creds.PasswordHash == "" {
		t.Fatal("db password_hash is empty")
	}

	codes := e2ehelpers.GetVerificationCodesByEmailAndPurpose(t, e2eEnv.PgPool(), user.Email, "register")
	if len(codes) != 1 {
		t.Fatalf("db verification_codes count = %d, want 1", len(codes))
	}
	if codes[0].ConsumedAt != nil {
		t.Fatal("db verification_code consumed_at should be nil before verify")
	}

	events := e2ehelpers.GetAuthEventsByIdentityID(t, e2eEnv.PgPool(), identity.ID)
	if len(events) != 1 {
		t.Fatalf("db auth_events count = %d, want 1", len(events))
	}
	if events[0].EventType != "register" {
		t.Fatalf("db auth_event type = %q, want register", events[0].EventType)
	}
}

func TestRegister_InvalidEmail(t *testing.T) {
	resetE2E(t)

	resp := client.PostJSON(t, "/auth/register", map[string]string{"email": "invalid", "password": "oldpass1word"})
	e2ehelpers.ExpectStatus(t, resp, http.StatusUnprocessableEntity)
}

func TestRegister_DuplicateEmail(t *testing.T) {
	resetE2E(t)

	user := e2ehelpers.NewUser()
	e2ehelpers.Register(t, client, user)

	resp := client.PostJSON(t, "/auth/register", map[string]string{"email": user.Email, "password": user.Password})
	e2ehelpers.ExpectStatus(t, resp, http.StatusConflict)

	identities, err := e2eEnv.PgPool().Query(t.Context(),
		`SELECT id FROM identities WHERE email = $1`, user.Email)
	if err != nil {
		t.Fatalf("query identities: %v", err)
	}
	defer identities.Close()
	count := 0
	for identities.Next() {
		count++
	}
	if count != 1 {
		t.Fatalf("db identity count = %d, want 1 (no duplicate)", count)
	}
}

func TestRegister_VerifyCode(t *testing.T) {
	resetE2E(t)

	capture := e2ehelpers.NewEventCapture(e2eEnv.NC())

	user := e2ehelpers.NewUser()
	code := e2ehelpers.WaitForCode(t, e2eEnv, e2ehelpers.SubjectRegisterCode, user.Email, e2ehelpers.PurposeRegister, func() {
		e2ehelpers.Register(t, client, user)
	})

	identity := e2ehelpers.GetIdentityByEmail(t, e2eEnv.PgPool(), user.Email)

	e2ehelpers.VerifyRegister(t, client, user.Email, code)

	identityAfter := e2ehelpers.GetIdentityByID(t, e2eEnv.PgPool(), identity.ID)
	if !identityAfter.EmailVerified {
		t.Fatal("db email_verified = false, want true after verify")
	}
	if identityAfter.Status != "active" {
		t.Fatalf("db status = %q, want active after verify", identityAfter.Status)
	}

	codes := e2ehelpers.GetVerificationCodesByIdentityID(t, e2eEnv.PgPool(), identity.ID)
	if len(codes) != 1 {
		t.Fatalf("db verification_codes count = %d, want 1", len(codes))
	}
	if codes[0].ConsumedAt == nil {
		t.Fatal("db verification_code consumed_at should NOT be nil after verify")
	}

	authEvents := e2ehelpers.GetAuthEventsByIdentityID(t, e2eEnv.PgPool(), identity.ID)
	if len(authEvents) != 2 {
		t.Fatalf("db auth_events count = %d, want 2", len(authEvents))
	}
	if authEvents[0].EventType != "register" {
		t.Fatalf("db auth_event[0] type = %q, want register", authEvents[0].EventType)
	}
	if authEvents[1].EventType != "register_verified" {
		t.Fatalf("db auth_event[1] type = %q, want register_verified", authEvents[1].EventType)
	}

	capture.AssertPublished(t, events.EventIdentityCreated)
	capture.AssertPublished(t, events.EventVerificationCodeSend)
	capture.AssertPublished(t, events.EventIdentityUpdated)
	capture.Close()
}

func TestRegister_VerifyWrongCode(t *testing.T) {
	resetE2E(t)

	user := e2ehelpers.NewUser()
	e2ehelpers.Register(t, client, user)

	resp := client.PostJSON(t, "/auth/register/verify", map[string]string{"email": user.Email, "code": "000000"})
	e2ehelpers.ExpectStatus(t, resp, http.StatusUnprocessableEntity)

	identity := e2ehelpers.GetIdentityByEmail(t, e2eEnv.PgPool(), user.Email)
	if identity.EmailVerified {
		t.Fatal("db email_verified should still be false after wrong code")
	}
	if identity.Status != "pending_verification" {
		t.Fatalf("db status = %q, want pending_verification after wrong code", identity.Status)
	}

	codes := e2ehelpers.GetVerificationCodesByIdentityID(t, e2eEnv.PgPool(), identity.ID)
	if len(codes) != 1 {
		t.Fatalf("db verification_codes count = %d, want 1", len(codes))
	}
	if codes[0].AttemptsCount != 1 {
		t.Fatalf("db attempts_count = %d, want 1 after wrong code", codes[0].AttemptsCount)
	}
	if codes[0].ConsumedAt != nil {
		t.Fatal("db verification_code consumed_at should be nil after wrong code")
	}
}
