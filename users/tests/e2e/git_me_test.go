package e2e_test

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"

	e2ehelpers "users/tests/e2e/helpers"
)

const e2eServiceToken = "e2e-service-token"

func TestGitMeGet_Success(t *testing.T) {
	resetE2E(t)

	identityID := uuid.New()
	profileID := uuid.New()
	email := "git-me-e2e@example.com"
	displayName := "git-user"
	gitToken := "test-git-token-abc"
	gitURL := "http://gitea.e2e.local"

	now := time.Now().UTC()
	_, err := e2eEnv.PgPool().Exec(context.Background(),
		`INSERT INTO user_profiles (id, identity_id, email, display_name, bio, avatar_url, avatar_object_key, status, email_verified, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, '', '', '', 'active', false, $5, $5)`,
		profileID, identityID, email, displayName, now,
	)
	if err != nil {
		t.Fatalf("insert profile: %v", err)
	}

	gitUserID := uuid.New()
	_, err = e2eEnv.PgPool().Exec(context.Background(),
		`INSERT INTO git_users (id, profile_id, git_token, git_url, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $5)`,
		gitUserID, profileID, gitToken, gitURL, now,
	)
	if err != nil {
		t.Fatalf("insert git user: %v", err)
	}

	resp := e2ehelpers.ServiceTokenGet(t, client, e2eServiceToken, "/git-user/me", map[string]string{
		"identity_id": identityID.String(),
	})
	e2ehelpers.ExpectStatus(t, resp, http.StatusOK)

	body := e2ehelpers.Decode[map[string]any](t, resp)
	if body["username"] != displayName {
		t.Fatalf("username = %v, want %v", body["username"], displayName)
	}
	if body["git_token"] != gitToken {
		t.Fatalf("git_token = %v, want %v", body["git_token"], gitToken)
	}
	if body["git_url"] != gitURL {
		t.Fatalf("git_url = %v, want %v", body["git_url"], gitURL)
	}
}

func TestGitMeGet_MissingServiceToken(t *testing.T) {
	resetE2E(t)

	resp := e2ehelpers.ServiceTokenGet(t, client, "", "/git-user/me", map[string]string{
		"identity_id": uuid.New().String(),
	})
	e2ehelpers.ExpectStatus(t, resp, http.StatusUnauthorized)

	body := e2ehelpers.Decode[map[string]any](t, resp)
	if body["code"] != "MISSING_AUTH_TOKEN" {
		t.Fatalf("code = %v, want MISSING_AUTH_TOKEN", body["code"])
	}
}

func TestGitMeGet_InvalidServiceToken(t *testing.T) {
	resetE2E(t)

	resp := e2ehelpers.ServiceTokenGet(t, client, "wrong-token", "/git-user/me", map[string]string{
		"identity_id": uuid.New().String(),
	})
	e2ehelpers.ExpectStatus(t, resp, http.StatusUnauthorized)

	body := e2ehelpers.Decode[map[string]any](t, resp)
	if body["code"] != "INVALID_AUTH_TOKEN" {
		t.Fatalf("code = %v, want INVALID_AUTH_TOKEN", body["code"])
	}
}

func TestGitMeGet_ProfileNotFound(t *testing.T) {
	resetE2E(t)

	resp := e2ehelpers.ServiceTokenGet(t, client, e2eServiceToken, "/git-user/me", map[string]string{
		"identity_id": uuid.New().String(),
	})
	e2ehelpers.ExpectStatus(t, resp, http.StatusNotFound)
}

func TestGitMeGet_GitUserNotFound(t *testing.T) {
	resetE2E(t)

	identityID := uuid.New()
	profileID := uuid.New()

	now := time.Now().UTC()
	_, err := e2eEnv.PgPool().Exec(context.Background(),
		`INSERT INTO user_profiles (id, identity_id, email, display_name, bio, avatar_url, avatar_object_key, status, email_verified, created_at, updated_at)
		 VALUES ($1, $2, 'no-git@example.com', 'no-git', '', '', '', 'active', false, $3, $3)`,
		profileID, identityID, now,
	)
	if err != nil {
		t.Fatalf("insert profile: %v", err)
	}

	resp := e2ehelpers.ServiceTokenGet(t, client, e2eServiceToken, "/git-user/me", map[string]string{
		"identity_id": identityID.String(),
	})
	e2ehelpers.ExpectStatus(t, resp, http.StatusNotFound)
}

func TestGitMeGet_InvalidIdentityID(t *testing.T) {
	resetE2E(t)

	resp := e2ehelpers.ServiceTokenGet(t, client, e2eServiceToken, "/git-user/me", map[string]string{
		"identity_id": "not-a-uuid",
	})
	e2ehelpers.ExpectStatus(t, resp, http.StatusBadRequest)
}

func TestGitMeGet_InvalidJSON(t *testing.T) {
	resetE2E(t)

	// Send malformed JSON
	resp := e2ehelpers.ServiceTokenGet(t, client, e2eServiceToken, "/git-user/me", nil)
	// Empty body with nil means no body at all — the handler should get a bind error or zero-value uuid parse error
	// Depending on implementation, this could be 400 or 422
	if resp.StatusCode != http.StatusBadRequest && resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 400 or 422", resp.StatusCode)
	}
}

func TestGitMeGet_FullFlow_WithProfileAndGitUser(t *testing.T) {
	resetE2E(t)

	identityID := uuid.New()
	profileID := uuid.New()
	email := "fullflow@example.com"
	displayName := "fullflow-user"
	gitToken := "flow-git-token"
	gitURL := "http://gitea.flow.local"

	now := time.Now().UTC()
	_, err := e2eEnv.PgPool().Exec(context.Background(),
		`INSERT INTO user_profiles (id, identity_id, email, display_name, bio, avatar_url, avatar_object_key, status, email_verified, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, 'bio text', '/avatar.png', 'avatars/1.png', 'active', true, $5, $5)`,
		profileID, identityID, email, displayName, now,
	)
	if err != nil {
		t.Fatalf("insert profile: %v", err)
	}

	gitUserID := uuid.New()
	_, err = e2eEnv.PgPool().Exec(context.Background(),
		`INSERT INTO git_users (id, profile_id, git_token, git_url, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $5)`,
		gitUserID, profileID, gitToken, gitURL, now,
	)
	if err != nil {
		t.Fatalf("insert git user: %v", err)
	}

	// Verify the data was inserted correctly
	profile := e2ehelpers.GetUserProfileByIdentityID(t, e2eEnv.PgPool(), identityID.String())
	if profile.DisplayName != displayName {
		t.Fatalf("profile display_name = %q, want %q", profile.DisplayName, displayName)
	}

	// Call the endpoint
	resp := e2ehelpers.ServiceTokenGet(t, client, e2eServiceToken, "/git-user/me", map[string]string{
		"identity_id": identityID.String(),
	})
	e2ehelpers.ExpectStatus(t, resp, http.StatusOK)

	body := e2ehelpers.Decode[map[string]any](t, resp)
	if body["username"] != displayName {
		t.Fatalf("username = %v, want %v", body["username"], displayName)
	}
	if body["git_token"] != gitToken {
		t.Fatalf("git_token = %v, want %v", body["git_token"], gitToken)
	}
	if body["git_url"] != gitURL {
		t.Fatalf("git_url = %v, want %v", body["git_url"], gitURL)
	}
}

// createTestGitUser inserts a profile and git_user directly into the database.
func createTestGitUser(t *testing.T, identityID uuid.UUID, email, displayName, gitToken, gitURL string) {
	t.Helper()
	profileID := uuid.New()
	now := time.Now().UTC()

	_, err := e2eEnv.PgPool().Exec(context.Background(),
		`INSERT INTO user_profiles (id, identity_id, email, display_name, bio, avatar_url, avatar_object_key, status, email_verified, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, '', '', '', 'active', false, $5, $5)`,
		profileID, identityID, email, displayName, now,
	)
	if err != nil {
		t.Fatalf("insert profile: %v", err)
	}

	if gitToken != "" {
		gitUserID := uuid.New()
		_, err = e2eEnv.PgPool().Exec(context.Background(),
			`INSERT INTO git_users (id, profile_id, git_token, git_url, created_at, updated_at)
			 VALUES ($1, $2, $3, $4, $5, $5)`,
			gitUserID, profileID, gitToken, gitURL, now,
		)
		if err != nil {
			t.Fatalf("insert git user: %v", err)
		}
	}
}
