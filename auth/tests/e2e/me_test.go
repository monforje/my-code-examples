package e2e_test

import (
	e2ehelpers "auth/tests/e2e/helpers"
	"net/http"
	"testing"
	"time"
)

func TestMe_ValidJWT(t *testing.T) {
	resetE2E(t)

	user, token := e2ehelpers.LoginAs(t, e2eEnv, client)
	resp := client.GetAuth(t, token, "/auth/me")
	e2ehelpers.ExpectStatus(t, resp, http.StatusOK)

	me := e2ehelpers.Decode[e2ehelpers.IdentityResponse](t, resp)
	if me.Email != user.Email {
		t.Fatalf("email = %q, want %q", me.Email, user.Email)
	}
	if !me.EmailVerified {
		t.Fatal("email_verified = false, want true")
	}
	if me.Status != "active" {
		t.Fatalf("status = %q, want active", me.Status)
	}

	dbIdentity := e2ehelpers.GetIdentityByEmail(t, e2eEnv.PgPool(), user.Email)
	if me.ID != dbIdentity.ID {
		t.Fatalf("me id = %q, db id = %q", me.ID, dbIdentity.ID)
	}
	if dbIdentity.CreatedAt.IsZero() || dbIdentity.CreatedAt.Before(time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)) {
		t.Fatalf("db created_at = %v, expected recent timestamp", dbIdentity.CreatedAt)
	}
}

func TestMe_WithoutToken(t *testing.T) {
	resetE2E(t)

	resp := client.Get(t, "/auth/me")
	e2ehelpers.ExpectStatus(t, resp, http.StatusUnauthorized)
}

func TestMe_InvalidToken(t *testing.T) {
	resetE2E(t)

	resp := client.GetAuth(t, "not-a-jwt", "/auth/me")
	e2ehelpers.ExpectStatus(t, resp, http.StatusUnauthorized)
}
