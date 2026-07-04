package gittasksclient

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"tasks/internal/config"
	clientsdto "tasks/pkg/http_clients/dto"
)

func TestGitTaskCreate_Success(t *testing.T) {
	want := &clientsdto.GitTaskCreateResponse{
		TaskID:   "pizza-api",
		Repo:     "alice/golden-pizza-api",
		CloneURL: "http://gitea.local/alice/golden-pizza-api.git",
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %s, want POST", r.Method)
		}
		if r.URL.Path != "/tasks" {
			t.Errorf("path = %s, want /tasks", r.URL.Path)
		}
		if r.Header.Get("Authorization") != "Bearer test-token" {
			t.Errorf("Authorization = %s, want Bearer test-token", r.Header.Get("Authorization"))
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("Content-Type = %s, want application/json", r.Header.Get("Content-Type"))
		}

		var req clientsdto.GitTaskCreateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		if req.Username != "alice" {
			t.Errorf("username = %s, want alice", req.Username)
		}
		if req.TaskID != "pizza-api" {
			t.Errorf("task_id = %s, want pizza-api", req.TaskID)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(want)
	}))
	defer srv.Close()

	client := NewGitAuthClient(config.GitTasksClientConfig{
		Token:   "test-token",
		BaseURL: srv.URL,
	})

	resp, err := client.GitTaskCreate(context.Background(), &clientsdto.GitTaskCreateRequest{
		Username: "alice",
		TaskID:   "pizza-api",
	})
	if err != nil {
		t.Fatalf("GitTaskCreate() error = %v", err)
	}

	if resp.TaskID != want.TaskID {
		t.Errorf("TaskID = %s, want %s", resp.TaskID, want.TaskID)
	}
	if resp.Repo != want.Repo {
		t.Errorf("Repo = %s, want %s", resp.Repo, want.Repo)
	}
	if resp.CloneURL != want.CloneURL {
		t.Errorf("CloneURL = %s, want %s", resp.CloneURL, want.CloneURL)
	}
}

func TestGitTaskCreate_Unauthorized(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"error":"unauthorized"}`))
	}))
	defer srv.Close()

	client := NewGitAuthClient(config.GitTasksClientConfig{
		Token:   "wrong-token",
		BaseURL: srv.URL,
	})

	_, err := client.GitTaskCreate(context.Background(), &clientsdto.GitTaskCreateRequest{
		Username: "alice",
		TaskID:   "pizza-api",
	})
	if err == nil {
		t.Fatal("GitTaskCreate() error = nil, want error")
	}
}

func TestGitTaskCreate_Conflict(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusConflict)
		w.Write([]byte(`{"error":"repo already exists"}`))
	}))
	defer srv.Close()

	client := NewGitAuthClient(config.GitTasksClientConfig{
		Token:   "test-token",
		BaseURL: srv.URL,
	})

	_, err := client.GitTaskCreate(context.Background(), &clientsdto.GitTaskCreateRequest{
		Username: "alice",
		TaskID:   "pizza-api",
	})
	if err == nil {
		t.Fatal("GitTaskCreate() error = nil, want error")
	}
}

func TestGitTaskCreate_InvalidJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{invalid json`))
	}))
	defer srv.Close()

	client := NewGitAuthClient(config.GitTasksClientConfig{
		Token:   "test-token",
		BaseURL: srv.URL,
	})

	_, err := client.GitTaskCreate(context.Background(), &clientsdto.GitTaskCreateRequest{
		Username: "alice",
		TaskID:   "pizza-api",
	})
	if err == nil {
		t.Fatal("GitTaskCreate() error = nil, want unmarshal error")
	}
}

func TestGitTaskCreate_ServerDown(t *testing.T) {
	client := NewGitAuthClient(config.GitTasksClientConfig{
		Token:   "test-token",
		BaseURL: "http://localhost:1",
	})

	_, err := client.GitTaskCreate(context.Background(), &clientsdto.GitTaskCreateRequest{
		Username: "alice",
		TaskID:   "pizza-api",
	})
	if err == nil {
		t.Fatal("GitTaskCreate() error = nil, want connection error")
	}
}

func TestGitTaskCreate_InternalError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":"internal"}`))
	}))
	defer srv.Close()

	client := NewGitAuthClient(config.GitTasksClientConfig{
		Token:   "test-token",
		BaseURL: srv.URL,
	})

	_, err := client.GitTaskCreate(context.Background(), &clientsdto.GitTaskCreateRequest{
		Username: "alice",
		TaskID:   "pizza-api",
	})
	if err == nil {
		t.Fatal("GitTaskCreate() error = nil, want error")
	}
}
