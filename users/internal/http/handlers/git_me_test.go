package handlers_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"go.uber.org/mock/gomock"

	httpserver "users/internal/http/gen"
	postgresrepo "users/internal/repository/postgres"
	service "users/internal/services"
	apperrors "users/pkg/errors"
)

const gitMeOp = "users.git.me.get"

// serviceTokenGet sends a GET request with X-Service-Token header and JSON body.
func serviceTokenGet(t *testing.T, ts *httptest.Server, path string, body any) *http.Response {
	t.Helper()
	var buf bytes.Buffer
	if body != nil {
		if err := json.NewEncoder(&buf).Encode(body); err != nil {
			t.Fatalf("json.Encode: %v", err)
		}
	}
	req, err := http.NewRequest(http.MethodGet, ts.URL+path, &buf)
	if err != nil {
		t.Fatalf("http.NewRequest: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Service-Token", "internal-service-token")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("http.Do: %v", err)
	}
	return resp
}

// ──────────────────────────────────────────────
// UsersGitMeGet
// ──────────────────────────────────────────────

func TestGitMeGet_Success(t *testing.T) {
	svc := newMock(t)
	ts := setupServer(t, svc)

	svc.EXPECT().GetGitMe(gomock.Any(), gomock.Any()).
		Return(&service.GitMeOutput{
			Username: "alice",
			GitToken: "git-token-123",
			GitURL:   "http://gitea.local",
		}, nil)

	resp := serviceTokenGet(t, ts, "/git-user/me", map[string]string{
		"identity_id": uuid.New().String(),
	})
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusOK)
	}
	var body httpserver.GitMeResponse
	decodeJSON(t, resp, &body)
	if body.Username != "alice" {
		t.Errorf("username = %q, want alice", body.Username)
	}
	if body.GitToken != "git-token-123" {
		t.Errorf("git_token = %q, want git-token-123", body.GitToken)
	}
	if body.GitUrl != "http://gitea.local" {
		t.Errorf("git_url = %q, want http://gitea.local", body.GitUrl)
	}
}

func TestGitMeGet_InvalidJSON(t *testing.T) {
	svc := newMock(t)
	ts := setupServer(t, svc)

	var buf bytes.Buffer
	buf.WriteString("{bad")
	req, err := http.NewRequest(http.MethodGet, ts.URL+"/git-user/me", &buf)
	if err != nil {
		t.Fatalf("http.NewRequest: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Service-Token", "internal-service-token")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("http.Do: %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusBadRequest)
	}
	var body httpserver.ErrorResponse
	decodeJSON(t, resp, &body)
	if body.Code != httpserver.INVALIDJSON {
		t.Errorf("code = %q, want %q", body.Code, httpserver.INVALIDJSON)
	}
}

func TestGitMeGet_InvalidIdentityID(t *testing.T) {
	svc := newMock(t)
	ts := setupServer(t, svc)

	resp := serviceTokenGet(t, ts, "/git-user/me", map[string]string{
		"identity_id": "not-a-uuid",
	})
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusBadRequest)
	}
	var body httpserver.ErrorResponse
	decodeJSON(t, resp, &body)
	if body.Code != httpserver.VALIDATIONERROR {
		t.Errorf("code = %q, want %q", body.Code, httpserver.VALIDATIONERROR)
	}
}

func TestGitMeGet_ProfileNotFound(t *testing.T) {
	svc := newMock(t)
	ts := setupServer(t, svc)

	svc.EXPECT().GetGitMe(gomock.Any(), gomock.Any()).
		Return(nil, apperrors.New("GitAuthService.GetGitMe", postgresrepo.ErrUserProfileNotFound))

	resp := serviceTokenGet(t, ts, "/git-user/me", map[string]string{
		"identity_id": uuid.New().String(),
	})
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusNotFound)
	}
	var body httpserver.ErrorResponse
	decodeJSON(t, resp, &body)
	if body.Code != httpserver.NOTFOUND {
		t.Errorf("code = %q, want %q", body.Code, httpserver.NOTFOUND)
	}
}

func TestGitMeGet_GitUserNotFound(t *testing.T) {
	svc := newMock(t)
	ts := setupServer(t, svc)

	svc.EXPECT().GetGitMe(gomock.Any(), gomock.Any()).
		Return(nil, apperrors.New("GitAuthService.GetGitMe", postgresrepo.ErrGitUserNotFound))

	resp := serviceTokenGet(t, ts, "/git-user/me", map[string]string{
		"identity_id": uuid.New().String(),
	})
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusNotFound)
	}
	var body httpserver.ErrorResponse
	decodeJSON(t, resp, &body)
	if body.Code != httpserver.NOTFOUND {
		t.Errorf("code = %q, want %q", body.Code, httpserver.NOTFOUND)
	}
}

func TestGitMeGet_ServiceError(t *testing.T) {
	svc := newMock(t)
	ts := setupServer(t, svc)

	svc.EXPECT().GetGitMe(gomock.Any(), gomock.Any()).
		Return(nil, errors.New("db down"))

	resp := serviceTokenGet(t, ts, "/git-user/me", map[string]string{
		"identity_id": uuid.New().String(),
	})
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusInternalServerError)
	}
	var body httpserver.ErrorResponse
	decodeJSON(t, resp, &body)
	if body.Code != httpserver.INTERNALERROR {
		t.Errorf("code = %q, want %q", body.Code, httpserver.INTERNALERROR)
	}
}
