package e2e_test

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"sync"
	"testing"
	"time"

	"tasks/internal/app"
	"tasks/internal/app/closer"
	"tasks/internal/config"
	clientsdto "tasks/pkg/http_clients/dto"
	"tasks/pkg/logger"
	e2ehelpers "tasks/tests/e2e/helpers"
)

var (
	e2eEnv   *e2ehelpers.Environment
	client   *e2ehelpers.Client
	external *externalServices
)

// e2eReportsToken - сервисный токен для POST /reports (имитирует notifications).
const e2eReportsToken = "e2e-reports-token"

type externalServices struct {
	users *httptest.Server
	git   *httptest.Server

	mu             sync.Mutex
	usersCalls     int
	gitCalls       int
	lastIdentityID string
	lastUsername   string
	lastTaskID     string
}

// TestMain - точка входа для e2e-тестов.
// Запускает PostgreSQL контейнер, миграции, HTTP-сервер и выполняет тесты.
func TestMain(m *testing.M) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	var err error
	e2eEnv, err = e2ehelpers.StartEnvironment(ctx)
	if err != nil {
		_, _ = os.Stderr.WriteString("start e2e environment: " + err.Error() + "\n")
		os.Exit(1)
	}

	external = startExternalServices()
	defer external.Close()

	cfg := newE2EConfig(e2eEnv, external)
	log := logger.New(&config.LoggerConfig{Level: -4, Format: config.FormatText, Output: io.Discard})
	application := app.New(context.Background(), log, cfg)

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		_, _ = os.Stderr.WriteString("listen: " + err.Error() + "\n")
		e2eEnv.Shutdown(context.Background())
		os.Exit(1)
	}

	serverErr := make(chan error, 1)
	go func() {
		serverErr <- application.HTTPserver.RunOnListener(listener)
	}()

	baseURL := "http://" + listener.Addr().String() + "/api/v1"
	client = e2ehelpers.NewClient(baseURL)

	if err := waitHTTP(context.Background(), baseURL); err != nil {
		_, _ = os.Stderr.WriteString("wait http: " + err.Error() + "\n")
		shutdown(application, serverErr)
		e2eEnv.Shutdown(context.Background())
		os.Exit(1)
	}

	code := m.Run()

	shutdown(application, serverErr)
	e2eEnv.Shutdown(context.Background())
	os.Exit(code)
}

func newE2EConfig(env *e2ehelpers.Environment, ext *externalServices) *config.Config {
	return &config.Config{
		Server: config.ServerConfig{
			Port: "0",
			CORS: config.CORSConfig{
				Origins: []string{"*"},
				Methods: []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
				Headers: []string{"Authorization", "Content-Type"},
			},
		},
		JWT: config.JWTConfig{Secret: "e2e-test-signing-key-which-is-long-enough"},
		PG:  env.PostgresConfig,
		Reports: config.ReportsConfig{
			ServiceToken: e2eReportsToken,
		},
		Features: config.FeaturesConfig{
			AccessTokenTTL:  15 * time.Minute,
			RefreshTokenLen: 32,
		},
		HTTPClient: config.HTTPClientConfig{
			UsersClient: config.UsersClientConfig{
				Token:   "users-service-token",
				BaseURL: ext.users.URL,
			},
			GitTasksClient: config.GitTasksClientConfig{
				Token:   "taskrunner-token",
				BaseURL: ext.git.URL,
			},
		},
	}
}

func startExternalServices() *externalServices {
	ext := &externalServices{}
	ext.users = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/api/v1/git-user/me" || r.Header.Get("X-Service-Token") != "users-service-token" {
			http.Error(w, "bad users request", http.StatusBadRequest)
			return
		}

		var req clientsdto.GitUserRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		ext.mu.Lock()
		ext.usersCalls++
		ext.lastIdentityID = req.IdentityID.String()
		ext.mu.Unlock()

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(clientsdto.GitUserResponse{
			Username: "alice",
			GitToken: "git-token",
			GitURL:   "http://gitea.local",
		})
	}))

	ext.git = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/tasks" || r.Header.Get("Authorization") != "Bearer taskrunner-token" {
			http.Error(w, "bad git request", http.StatusBadRequest)
			return
		}

		var req clientsdto.GitTaskCreateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		ext.mu.Lock()
		ext.gitCalls++
		ext.lastUsername = req.Username
		ext.lastTaskID = req.TaskID
		ext.mu.Unlock()

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(clientsdto.GitTaskCreateResponse{
			TaskID:   req.TaskID,
			Repo:     req.Username + "/golden-" + req.TaskID,
			CloneURL: "http://gitea.local/" + req.Username + "/golden-" + req.TaskID + ".git",
		})
	}))

	return ext
}

func (e *externalServices) Reset() {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.usersCalls = 0
	e.gitCalls = 0
	e.lastIdentityID = ""
	e.lastUsername = ""
	e.lastTaskID = ""
}

func (e *externalServices) Close() {
	if e.users != nil {
		e.users.Close()
	}
	if e.git != nil {
		e.git.Close()
	}
}

func waitHTTP(ctx context.Context, baseURL string) error {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	for {
		request, err := http.NewRequestWithContext(ctx, http.MethodGet, baseURL+"/tasks?limit=1", nil)
		if err != nil {
			return err
		}
		resp, err := http.DefaultClient.Do(request)
		if err == nil {
			_ = resp.Body.Close()
			return nil
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(50 * time.Millisecond):
		}
	}
}

func shutdown(application *app.App, serverErr <-chan error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	_ = application.HTTPserver.Shutdown(ctx)
	_ = closer.CloseAll(ctx)

	select {
	case err := <-serverErr:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			_, _ = os.Stderr.WriteString("http server error: " + err.Error() + "\n")
		}
	case <-time.After(time.Second):
	}
}

func resetE2E(t *testing.T) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := e2eEnv.Reset(ctx); err != nil {
		t.Fatalf("reset e2e environment: %v", err)
	}
	external.Reset()
}
