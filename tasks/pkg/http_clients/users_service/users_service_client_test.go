package usersserviceclient

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"tasks/internal/config"
	clientsdto "tasks/pkg/http_clients/dto"

	"github.com/google/uuid"
)

func TestGetGitUser_Success(t *testing.T) {
	want := &clientsdto.GitUserResponse{
		Username: "alice",
		GitToken: "ghp_xxxxxxxxxxxx",
		GitURL:   "http://gitea.local",
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("method = %s, want GET", r.Method)
		}
		if r.URL.Path != "/api/v1/git-user/me" {
			t.Errorf("path = %s, want /api/v1/git-user/me", r.URL.Path)
		}
		if r.Header.Get("X-Service-Token") != "test-token" {
			t.Errorf("X-Service-Token = %s, want test-token", r.Header.Get("X-Service-Token"))
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("Content-Type = %s, want application/json", r.Header.Get("Content-Type"))
		}

		var req clientsdto.GitUserRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		if req.IdentityID == uuid.Nil {
			t.Error("identity_id is nil")
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(want)
	}))
	defer srv.Close()

	client := NewUsersClient(config.UsersClientConfig{
		Token:   "test-token",
		BaseURL: srv.URL,
	})

	resp, err := client.GetGitUser(context.Background(), &clientsdto.GitUserRequest{
		IdentityID: uuid.New(),
	})
	if err != nil {
		t.Fatalf("GetGitUser() error = %v", err)
	}

	if resp.Username != want.Username {
		t.Errorf("Username = %s, want %s", resp.Username, want.Username)
	}
	if resp.GitToken != want.GitToken {
		t.Errorf("GitToken = %s, want %s", resp.GitToken, want.GitToken)
	}
	if resp.GitURL != want.GitURL {
		t.Errorf("GitURL = %s, want %s", resp.GitURL, want.GitURL)
	}
}

func TestGetGitUser_Unauthorized(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"code":"MISSING_AUTH_TOKEN","message":"unauthorized"}`))
	}))
	defer srv.Close()

	client := NewUsersClient(config.UsersClientConfig{
		Token:   "wrong-token",
		BaseURL: srv.URL,
	})

	_, err := client.GetGitUser(context.Background(), &clientsdto.GitUserRequest{
		IdentityID: uuid.New(),
	})
	if err == nil {
		t.Fatal("GetGitUser() error = nil, want error")
	}
}

func TestGetGitUser_NotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"code":"NOT_FOUND","message":"user not found"}`))
	}))
	defer srv.Close()

	client := NewUsersClient(config.UsersClientConfig{
		Token:   "test-token",
		BaseURL: srv.URL,
	})

	_, err := client.GetGitUser(context.Background(), &clientsdto.GitUserRequest{
		IdentityID: uuid.New(),
	})
	if err == nil {
		t.Fatal("GetGitUser() error = nil, want error")
	}
}

func TestGetGitUser_BadRequest(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"code":"VALIDATION_ERROR","message":"invalid identity_id"}`))
	}))
	defer srv.Close()

	client := NewUsersClient(config.UsersClientConfig{
		Token:   "test-token",
		BaseURL: srv.URL,
	})

	_, err := client.GetGitUser(context.Background(), &clientsdto.GitUserRequest{
		IdentityID: uuid.Nil,
	})
	if err == nil {
		t.Fatal("GetGitUser() error = nil, want error")
	}
}

func TestGetGitUser_InvalidJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{invalid json`))
	}))
	defer srv.Close()

	client := NewUsersClient(config.UsersClientConfig{
		Token:   "test-token",
		BaseURL: srv.URL,
	})

	_, err := client.GetGitUser(context.Background(), &clientsdto.GitUserRequest{
		IdentityID: uuid.New(),
	})
	if err == nil {
		t.Fatal("GetGitUser() error = nil, want unmarshal error")
	}
}

func TestGetGitUser_ServerDown(t *testing.T) {
	client := NewUsersClient(config.UsersClientConfig{
		Token:   "test-token",
		BaseURL: "http://localhost:1",
	})

	_, err := client.GetGitUser(context.Background(), &clientsdto.GitUserRequest{
		IdentityID: uuid.New(),
	})
	if err == nil {
		t.Fatal("GetGitUser() error = nil, want connection error")
	}
}
