package e2e_test

import (
	"context"
	"errors"
	"io"
	"net"
	"net/http"
	"os"
	"testing"
	"time"

	"users/internal/app"
	"users/internal/app/closer"
	"users/internal/config"
	"users/internal/repository/security"
	"users/pkg/logger"
	e2ehelpers "users/tests/e2e/helpers"
)

var (
	e2eEnv       *e2ehelpers.Environment
	client       *e2ehelpers.Client
	tokenManager *security.Manager
)

func TestMain(m *testing.M) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	var err error
	e2eEnv, err = e2ehelpers.StartEnvironment(ctx)
	if err != nil {
		_, _ = os.Stderr.WriteString("start e2e environment: " + err.Error() + "\n")
		os.Exit(1)
	}

	cfg := newE2EConfig(e2eEnv)
	log := logger.New(&config.LoggerConfig{Level: -4, Format: config.FormatText, Output: io.Discard})
	application := app.New(context.Background(), log, cfg, app.ModeServer)

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
	tokenManager = e2ehelpers.NewTokenManager()

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

func newE2EConfig(env *e2ehelpers.Environment) *config.Config {
	return &config.Config{
		Server: config.ServerConfig{
			Port:         "0",
			ServiceToken: "e2e-service-token",
			CORS: config.CORSConfig{
				Origins: []string{"*"},
				Methods: []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
				Headers: []string{"Authorization", "Content-Type"},
			},
		},
		JWT:   config.JWTConfig{Secret: e2ehelpers.TestJWTSigningKey},
		PG:    env.PostgresConfig,
		NATS:  config.NATSConfig{},
		Storage: config.StorageConfig{
			AvatarDir:    env.AvatarDir,
			AvatarPublic: env.AvatarDir,
		},
		Features: config.FeaturesConfig{
			AccessTokenTTL:  15 * time.Minute,
			RefreshTokenLen: 32,
		},
		HttpClient: config.HttpClientConfig{
			GitAuthClient: config.GitAuthClientConfig{
				Token:   "e2e-test-token",
				BaseURL: "http://localhost:0",
			},
		},
	}
}

func waitHTTP(ctx context.Context, baseURL string) error {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	for {
		request, err := http.NewRequestWithContext(ctx, http.MethodGet, baseURL+"/profile/me", nil)
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
}
