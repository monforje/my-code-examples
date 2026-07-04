package e2e_test

import (
	"net/http"
	"strings"
	"testing"
	"time"

	"users/internal/models/records"
	e2ehelpers "users/tests/e2e/helpers"

	"github.com/google/uuid"
)

func createTestProfileWithEmail(t *testing.T, identityID uuid.UUID, email string) {
	t.Helper()
	now := time.Now().UTC()
	profile := &records.UserProfile{
		ID:         uuid.New(),
		IdentityID: identityID,
		Email:      email,
		Status:     "active",
		CreatedAt:  now,
		UpdatedAt:  now,
	}
	_, err := e2eEnv.PgPool().Exec(t.Context(), `
		INSERT INTO user_profiles (id, identity_id, email, display_name, bio, avatar_url, avatar_object_key, status, email_verified, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`, profile.ID, profile.IdentityID, profile.Email, profile.DisplayName, profile.BIO, profile.AvatarURL, profile.AvatarObjectKey, profile.Status, profile.EmailVerified, profile.CreatedAt, profile.UpdatedAt)
	if err != nil {
		t.Fatalf("create test profile: %v", err)
	}
}

func TestSettingsUpdate_Success(t *testing.T) {
	resetE2E(t)

	identityID := uuid.New()
	email := "settings-test@example.com"
	createTestProfileWithEmail(t, identityID, email)

	token := e2ehelpers.GenerateToken(t, tokenManager, identityID)
	resp := e2ehelpers.PatchAuthJSON(t, client, token, "/profile/me/settings", map[string]string{
		"display_name": "New Name",
		"bio":          "New bio",
	})
	e2ehelpers.ExpectStatus(t, resp, http.StatusOK)

	body := e2ehelpers.Decode[map[string]any](t, resp)
	if body["display_name"] != "New Name" {
		t.Fatalf("display_name = %v, want 'New Name'", body["display_name"])
	}
	if body["bio"] != "New bio" {
		t.Fatalf("bio = %v, want 'New bio'", body["bio"])
	}
	if body["email"] != email {
		t.Fatalf("email changed: %v", body["email"])
	}
}

func TestSettingsUpdate_PartialUpdate(t *testing.T) {
	resetE2E(t)

	identityID := uuid.New()
	email := "partial-test@example.com"
	createTestProfileWithEmail(t, identityID, email)

	token := e2ehelpers.GenerateToken(t, tokenManager, identityID)

	resp := e2ehelpers.PatchAuthJSON(t, client, token, "/profile/me/settings", map[string]string{
		"display_name": "Initial",
	})
	e2ehelpers.ExpectStatus(t, resp, http.StatusOK)

	resp = e2ehelpers.PatchAuthJSON(t, client, token, "/profile/me/settings", map[string]string{
		"bio": "Updated bio",
	})
	e2ehelpers.ExpectStatus(t, resp, http.StatusOK)

	body := e2ehelpers.Decode[map[string]any](t, resp)
	if body["display_name"] != "Initial" {
		t.Fatalf("display_name = %v, want 'Initial'", body["display_name"])
	}
	if body["bio"] != "Updated bio" {
		t.Fatalf("bio = %v, want 'Updated bio'", body["bio"])
	}
}

func TestSettingsUpdate_EmptyBody(t *testing.T) {
	resetE2E(t)

	identityID := uuid.New()
	createTestProfileWithEmail(t, identityID, "empty@example.com")

	token := e2ehelpers.GenerateToken(t, tokenManager, identityID)
	resp := e2ehelpers.PatchAuthJSON(t, client, token, "/profile/me/settings", map[string]string{})
	e2ehelpers.ExpectStatus(t, resp, http.StatusOK)
}

func TestSettingsUpdate_DisplayNameTooLong(t *testing.T) {
	resetE2E(t)

	identityID := uuid.New()
	createTestProfileWithEmail(t, identityID, "long-name@example.com")

	token := e2ehelpers.GenerateToken(t, tokenManager, identityID)
	resp := e2ehelpers.PatchAuthJSON(t, client, token, "/profile/me/settings", map[string]string{
		"display_name": strings.Repeat("A", 51),
	})
	e2ehelpers.ExpectStatus(t, resp, http.StatusUnprocessableEntity)

	body := e2ehelpers.Decode[map[string]any](t, resp)
	if body["code"] != "VALIDATION_ERROR" {
		t.Fatalf("code = %v, want VALIDATION_ERROR", body["code"])
	}
}

func TestSettingsUpdate_BioTooLong(t *testing.T) {
	resetE2E(t)

	identityID := uuid.New()
	createTestProfileWithEmail(t, identityID, "long-bio@example.com")

	token := e2ehelpers.GenerateToken(t, tokenManager, identityID)
	resp := e2ehelpers.PatchAuthJSON(t, client, token, "/profile/me/settings", map[string]string{
		"bio": strings.Repeat("A", 501),
	})
	e2ehelpers.ExpectStatus(t, resp, http.StatusUnprocessableEntity)

	body := e2ehelpers.Decode[map[string]any](t, resp)
	if body["code"] != "VALIDATION_ERROR" {
		t.Fatalf("code = %v, want VALIDATION_ERROR", body["code"])
	}
}

func TestSettingsUpdate_ProfileNotFound(t *testing.T) {
	resetE2E(t)

	identityID := uuid.New()
	token := e2ehelpers.GenerateToken(t, tokenManager, identityID)
	resp := e2ehelpers.PatchAuthJSON(t, client, token, "/profile/me/settings", map[string]string{
		"display_name": "test",
	})
	e2ehelpers.ExpectStatus(t, resp, http.StatusNotFound)
}

func TestSettingsUpdate_MissingToken(t *testing.T) {
	resetE2E(t)

	resp := e2ehelpers.PatchAuthJSON(t, client, "", "/profile/me/settings", map[string]string{
		"display_name": "test",
	})
	e2ehelpers.ExpectStatus(t, resp, http.StatusUnauthorized)

	body := e2ehelpers.Decode[map[string]any](t, resp)
	if body["code"] != "MISSING_AUTH_TOKEN" {
		t.Fatalf("code = %v, want MISSING_AUTH_TOKEN", body["code"])
	}
}
