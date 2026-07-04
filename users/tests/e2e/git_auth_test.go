package e2e_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"users/internal/config"
	postgres "users/internal/repository/postgres"
	"users/internal/repository/storage"
	service "users/internal/services"
	gitauthservice "users/internal/services/git_auth"
	clientsdto "users/pkg/http_clients/dto"
	gitauthclient "users/pkg/http_clients/git_auth"
	"users/pkg/logger"
	e2ehelpers "users/tests/e2e/helpers"
)

func TestHandleIdentityCreated_RegistersGitUser(t *testing.T) {
	resetE2E(t)

	const gitAuthToken = "test-git-auth-token"
	const gitToken = "registered-git-token"
	const gitURL = "http://gitea.e2e.local"

	requests := make(chan clientsdto.RegisterGitUserRequest, 1)
	gitAuthServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/register" {
			http.NotFound(w, r)
			return
		}
		if r.Method != http.MethodPost {
			t.Fatalf("method = %s, want POST", r.Method)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer "+gitAuthToken {
			t.Fatalf("Authorization = %q, want bearer token", got)
		}

		var req clientsdto.RegisterGitUserRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("decode git auth request: %v", err)
		}
		requests <- req

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(clientsdto.RegisterGitUserResponse{
			Username: req.Username,
			Token:    gitToken,
			GitURL:   gitURL,
		})
	}))
	defer gitAuthServer.Close()

	store := postgres.NewStore(postgres.New(e2eEnv.PgPool()))
	log := logger.New(&config.LoggerConfig{Level: -4, Format: config.FormatText})
	gitAuthSvc := gitauthservice.NewGitAuthService(
		log,
		gitauthclient.NewGitAuthClient(config.GitAuthClientConfig{
			Token:   gitAuthToken,
			BaseURL: gitAuthServer.URL,
		}),
		store.GitUsers(),
		store.UserProfiles(),
	)
	usersSvc := service.NewUsersService(
		log,
		store.UserProfiles(),
		storage.NewLocalAvatarStorage(e2eEnv.AvatarDir, e2eEnv.AvatarDir),
		nil,
		gitAuthSvc,
	)

	identityID := uuid.New()
	email := "git-e2e@example.com"
	if err := usersSvc.HandleIdentityCreated(context.Background(), &service.HandleIdentityCreatedInput{
		IdentityID: identityID.String(),
		Email:      email,
	}); err != nil {
		t.Fatalf("HandleIdentityCreated() error = %v", err)
	}

	profile := e2ehelpers.GetUserProfileByIdentityID(t, e2eEnv.PgPool(), identityID.String())
	t.Logf("profile created: id=%s identity_id=%s display_name=%s", profile.ID, profile.IdentityID, profile.DisplayName)

	select {
	case req := <-requests:
		t.Logf("git auth request received: username=%s email=%s", req.Username, req.Email)
		if req.Email != email {
			t.Fatalf("git auth email = %q, want %q", req.Email, email)
		}
		if req.Username != profile.DisplayName {
			t.Fatalf("git auth username = %q, want %q", req.Username, profile.DisplayName)
		}
	case <-time.After(2 * time.Second):
		diagnoseGitUserMissing(t, e2eEnv.PgPool(), profile.ID)
		t.Fatal("git auth request was not sent")
	}

	gitUser := waitGitUserByIdentityID(t, profile.ID)
	if gitUser.ProfileID != profile.ID {
		t.Fatalf("git user profile_id = %q, want profile id %q", gitUser.ProfileID, profile.ID)
	}
	if gitUser.GitToken != gitToken {
		t.Fatalf("git token = %q, want %q", gitUser.GitToken, gitToken)
	}
	if gitUser.GitURL != gitURL {
		t.Fatalf("git url = %q, want %q", gitUser.GitURL, gitURL)
	}
}

func waitGitUserByIdentityID(t *testing.T, profileID string) e2ehelpers.GitUserRow {
	t.Helper()

	deadline := time.Now().Add(2 * time.Second)
	for {
		var row e2ehelpers.GitUserRow
		err := e2eEnv.PgPool().QueryRow(context.Background(),
			`SELECT id, profile_id, git_token, git_url, created_at, updated_at
			 FROM git_users WHERE profile_id = $1`, profileID,
		).Scan(&row.ID, &row.ProfileID, &row.GitToken, &row.GitURL, &row.CreatedAt, &row.UpdatedAt)
		if err == nil {
			return row
		}
		if err != pgx.ErrNoRows {
			t.Fatalf("query git user: %v", err)
		}
		if time.Now().After(deadline) {
			diagnoseGitUserMissing(t, e2eEnv.PgPool(), profileID)
			t.Fatalf("git user for profile_id %s was not created", profileID)
		}
		time.Sleep(25 * time.Millisecond)
	}
}

func diagnoseGitUserMissing(t *testing.T, pool interface{ QueryRow(context.Context, string, ...any) pgx.Row }, profileID string) {
	t.Helper()

	var count int
	err := pool.QueryRow(context.Background(),
		`SELECT COUNT(*) FROM git_users WHERE profile_id = $1`, profileID,
	).Scan(&count)
	if err != nil {
		t.Logf("diagnose: failed to count git_users: %v", err)
	} else {
		t.Logf("diagnose: git_users rows for profile_id=%s: %d", profileID, count)
	}

	var total int
	err = pool.QueryRow(context.Background(), `SELECT COUNT(*) FROM git_users`).Scan(&total)
	if err != nil {
		t.Logf("diagnose: failed to count total git_users: %v", err)
	} else {
		t.Logf("diagnose: total git_users rows: %d", total)
	}
}
