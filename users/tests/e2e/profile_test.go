package e2e_test

import (
	"net/http"
	"testing"

	e2ehelpers "users/tests/e2e/helpers"

	"github.com/google/uuid"
)

func TestProfileGet_Success(t *testing.T) {
	resetE2E(t)

	identityID := uuid.New()
	email := "profile-test@example.com"
	createTestProfileWithEmail(t, identityID, email)

	token := e2ehelpers.GenerateToken(t, tokenManager, identityID)
	resp := e2ehelpers.GetAuth(t, client, token, "/profile/me")
	e2ehelpers.ExpectStatus(t, resp, http.StatusOK)

	body := e2ehelpers.Decode[map[string]any](t, resp)
	if body["email"] != email {
		t.Fatalf("email = %v, want %v", body["email"], email)
	}
	if body["identity_id"] != identityID.String() {
		t.Fatalf("identity_id = %v, want %v", body["identity_id"], identityID.String())
	}
	if body["display_name"] != nil {
		t.Fatalf("display_name = %v, want nil", body["display_name"])
	}
	if body["bio"] != "" {
		t.Fatalf("bio = %v, want empty", body["bio"])
	}
	if body["avatar_url"] != nil {
		t.Fatalf("avatar_url = %v, want nil", body["avatar_url"])
	}
	if body["status"] != "active" {
		t.Fatalf("status = %v, want active", body["status"])
	}
	if body["email_verified"] != false {
		t.Fatalf("email_verified = %v, want false", body["email_verified"])
	}
}

func TestProfileGet_MissingToken(t *testing.T) {
	resetE2E(t)

	resp := e2ehelpers.GetAuth(t, client, "", "/profile/me")
	e2ehelpers.ExpectStatus(t, resp, http.StatusUnauthorized)

	body := e2ehelpers.Decode[map[string]any](t, resp)
	if body["code"] != "MISSING_AUTH_TOKEN" {
		t.Fatalf("code = %v, want MISSING_AUTH_TOKEN", body["code"])
	}
}

func TestProfileGet_InvalidToken(t *testing.T) {
	resetE2E(t)

	resp := e2ehelpers.GetAuth(t, client, "not-a-jwt", "/profile/me")
	e2ehelpers.ExpectStatus(t, resp, http.StatusUnauthorized)

	body := e2ehelpers.Decode[map[string]any](t, resp)
	if body["code"] != "INVALID_AUTH_TOKEN" {
		t.Fatalf("code = %v, want INVALID_AUTH_TOKEN", body["code"])
	}
}

func TestProfileGet_ProfileNotFound(t *testing.T) {
	resetE2E(t)

	identityID := uuid.New()
	token := e2ehelpers.GenerateToken(t, tokenManager, identityID)
	resp := e2ehelpers.GetAuth(t, client, token, "/profile/me")
	e2ehelpers.ExpectStatus(t, resp, http.StatusNotFound)
}
