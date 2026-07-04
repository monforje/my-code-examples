package e2e_test

import (
	"auth/internal/events"
	e2ehelpers "auth/tests/e2e/helpers"
	"net/http"
	"testing"
)

func TestLogin_ValidPassword(t *testing.T) {
	resetE2E(t)

	capture := e2ehelpers.NewEventCapture(e2eEnv.NC())
	defer capture.Close()

	user := e2ehelpers.RegisterAndVerify(t, e2eEnv, client)
	identity := e2ehelpers.GetIdentityByEmail(t, e2eEnv.PgPool(), user.Email)

	token := e2ehelpers.Login(t, client, user)

	if token.AccessToken == "" {
		t.Fatal("access_token is empty")
	}
	if token.ExpiresIn <= 0 {
		t.Fatalf("expires_in = %d, want positive", token.ExpiresIn)
	}

	sessions := e2ehelpers.GetSessionsByIdentityID(t, e2eEnv.PgPool(), identity.ID)
	if len(sessions) != 1 {
		t.Fatalf("db sessions count = %d, want 1", len(sessions))
	}
	if sessions[0].RefreshTokenHash == "" {
		t.Fatal("db refresh_token_hash is empty")
	}
	if sessions[0].RevokedAt != nil {
		t.Fatal("db session revoked_at should be nil for active session")
	}

	loginEvents := e2ehelpers.GetAuthEventsByType(t, e2eEnv.PgPool(), identity.ID, "login")
	if len(loginEvents) != 1 {
		t.Fatalf("db login events count = %d, want 1", len(loginEvents))
	}

	capture.AssertPublished(t, events.EventIdentityLogin)
}

func TestLogin_WrongPassword(t *testing.T) {
	resetE2E(t)

	user := e2ehelpers.RegisterAndVerify(t, e2eEnv, client)
	identity := e2ehelpers.GetIdentityByEmail(t, e2eEnv.PgPool(), user.Email)

	resp := client.PostJSON(t, "/auth/login", map[string]string{"email": user.Email, "password": "wrongpass1"})
	e2ehelpers.ExpectStatus(t, resp, http.StatusUnauthorized)

	sessions := e2ehelpers.GetSessionsByIdentityID(t, e2eEnv.PgPool(), identity.ID)
	if len(sessions) != 0 {
		t.Fatalf("db sessions count = %d, want 0 after failed login", len(sessions))
	}

	loginEvents := e2ehelpers.GetAuthEventsByType(t, e2eEnv.PgPool(), identity.ID, "login")
	if len(loginEvents) != 0 {
		t.Fatalf("db login events count = %d, want 0 after failed login", len(loginEvents))
	}
}

func TestLogin_UnverifiedUser(t *testing.T) {
	resetE2E(t)

	user := e2ehelpers.NewUser()
	e2ehelpers.Register(t, client, user)
	identity := e2ehelpers.GetIdentityByEmail(t, e2eEnv.PgPool(), user.Email)

	resp := client.PostJSON(t, "/auth/login", map[string]string{"email": user.Email, "password": user.Password})
	e2ehelpers.ExpectStatus(t, resp, http.StatusForbidden)

	sessions := e2ehelpers.GetSessionsByIdentityID(t, e2eEnv.PgPool(), identity.ID)
	if len(sessions) != 0 {
		t.Fatalf("db sessions count = %d, want 0 for unverified user", len(sessions))
	}
}
