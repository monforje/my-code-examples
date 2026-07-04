package e2e_test

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"sync"
	"testing"
	"time"

	"notifications/internal/app"
	"notifications/internal/app/closer"
	"notifications/internal/config"
	postgresrepo "notifications/internal/repository/postgres"
	service "notifications/internal/services"
	"notifications/internal/worker/consumer"
	"notifications/pkg/logger"
	e2ehelpers "notifications/tests/e2e/helpers"

	"github.com/nats-io/nats.go"
)

var (
	e2eEnv      *e2ehelpers.Environment
	natsConn    *nats.Conn
	sender      *recordingEmailSender
	consumerApp *consumer.Consumer

	httpApp    *app.App
	httpClient *http.Client
	baseURL    string

	// mock tasks service — принимает CI-отчёты от notifications.
	tasksRecorder *recordingTasksServer
	mockTasksSrv  *httptest.Server
)

// recordingTasksServer - записывает CI-отчёты, полученные mock tasks-сервисом.
type recordingTasksServer struct {
	mu      sync.Mutex
	reports []map[string]any
}

func (r *recordingTasksServer) add(rep map[string]any) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.reports = append(r.reports, rep)
}

func (r *recordingTasksServer) count() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return len(r.reports)
}

func (r *recordingTasksServer) last() map[string]any {
	r.mu.Lock()
	defer r.mu.Unlock()
	if len(r.reports) == 0 {
		return nil
	}
	return r.reports[len(r.reports)-1]
}

func (r *recordingTasksServer) reset() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.reports = nil
}

type sentEmail struct {
	template string
	params   service.SendCodeEmailParams
}

type recordingEmailSender struct {
	mu   sync.Mutex
	sent []sentEmail
}

func (s *recordingEmailSender) SendVerificationEmail(_ context.Context, params service.SendCodeEmailParams) error {
	s.add("verification", params)
	return nil
}

func (s *recordingEmailSender) SendPasswordResetEmail(_ context.Context, params service.SendCodeEmailParams) error {
	s.add("password_reset", params)
	return nil
}

func (s *recordingEmailSender) SendPasswordChangeEmail(_ context.Context, params service.SendCodeEmailParams) error {
	s.add("password_change", params)
	return nil
}

func (s *recordingEmailSender) SendEmailChangeEmail(_ context.Context, params service.SendCodeEmailParams) error {
	s.add("email_change", params)
	return nil
}

func (s *recordingEmailSender) SendDeleteAccountEmail(_ context.Context, params service.SendCodeEmailParams) error {
	s.add("delete_account", params)
	return nil
}

func (s *recordingEmailSender) add(template string, params service.SendCodeEmailParams) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.sent = append(s.sent, sentEmail{template: template, params: params})
}

func (s *recordingEmailSender) reset() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.sent = nil
}

func (s *recordingEmailSender) count() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.sent)
}

func (s *recordingEmailSender) last() sentEmail {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.sent[len(s.sent)-1]
}

func TestMain(m *testing.M) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	var err error
	e2eEnv, err = e2ehelpers.StartEnvironment(ctx)
	if err != nil {
		_, _ = os.Stderr.WriteString("start e2e environment: " + err.Error() + "\n")
		os.Exit(1)
	}

	natsConn, err = nats.Connect(e2eEnv.NATSConfig.URL())
	if err != nil {
		_, _ = os.Stderr.WriteString("connect nats: " + err.Error() + "\n")
		e2eEnv.Shutdown(context.Background())
		os.Exit(1)
	}

	log := logger.New(&config.LoggerConfig{Level: slog.LevelError, Format: config.FormatText, Output: io.Discard})

	// --- worker setup ---
	sender = &recordingEmailSender{}
	notiSvc := service.NewNotificationService(sender)
	repo := postgresrepo.New(e2eEnv.PgPool())
	consumerApp = consumer.NewConsumer(natsConn, log, notiSvc, postgresrepo.NewProcessedEventsRepo(repo))

	consumerErr := make(chan error, 1)
	go func() {
		consumerErr <- consumerApp.Run()
	}()
	if err := waitSubscriptions(context.Background(), natsConn); err != nil {
		_, _ = os.Stderr.WriteString("wait subscriptions: " + err.Error() + "\n")
		shutdown(consumerErr)
		e2eEnv.Shutdown(context.Background())
		os.Exit(1)
	}

	// --- HTTP server setup ---
	mockTasksSrv = startMockTasks()
	httpApp = startHTTPServer(log)

	code := m.Run()

	shutdown(consumerErr)
	shutdownHTTPServer()
	mockTasksSrv.Close()
	e2eEnv.Shutdown(context.Background())
	os.Exit(code)
}

// startMockTasks - поднимает mock tasks-сервиса, принимающий POST /reports.
func startMockTasks() *httptest.Server {
	tasksRecorder = &recordingTasksServer{}
	mux := http.NewServeMux()
	mux.HandleFunc("/reports", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		var body map[string]any
		_ = json.NewDecoder(r.Body).Decode(&body)
		tasksRecorder.add(body)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{}`))
	})
	return httptest.NewServer(mux)
}

func startHTTPServer(log *logger.Logger) *app.App {
	cfg := &config.Config{
		Server: config.ServerConfig{
			Port: "0",
			CORS: config.CORSConfig{
				Origins: []string{"*"},
				Methods: []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
				Headers: []string{"Authorization", "Content-Type"},
			},
		},
		PG:    e2eEnv.PostgresConfig,
		Redis: e2eEnv.RedisConfig,
		NATS:  config.NATSConfig{},
		Features: config.FeaturesConfig{
			CIReportCacheTTL: time.Hour,
		},
		HTTPClient: config.HTTPClientConfig{
			ReportsClient: config.ReportsClientConfig{
				Token:   "e2e-token",
				BaseURL: mockTasksSrv.URL,
			},
		},
	}

	application := app.New(context.Background(), log, cfg)
	application.InitHTTPServer()

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		_, _ = os.Stderr.WriteString("listen: " + err.Error() + "\n")
		os.Exit(1)
	}

	baseURL = "http://" + listener.Addr().String() + "/api/v1/notifications"

	go func() {
		if err := application.HTTPserver.RunOnListener(listener); err != nil {
			_ = err
		}
	}()

	if err := waitHTTP(context.Background(), baseURL); err != nil {
		_, _ = os.Stderr.WriteString("wait http: " + err.Error() + "\n")
		os.Exit(1)
	}

	httpClient = &http.Client{Timeout: 10 * time.Second}

	return application
}

func shutdownHTTPServer() {
	if httpApp == nil || httpApp.HTTPserver == nil {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = httpApp.HTTPserver.Shutdown(ctx)
	_ = closer.CloseAll(ctx)
}

func waitHTTP(ctx context.Context, base string) error {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	for {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, base+"/reports/probe", nil)
		if err != nil {
			return err
		}
		resp, err := http.DefaultClient.Do(req)
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

func waitSubscriptions(ctx context.Context, conn *nats.Conn) error {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	for {
		if err := conn.Flush(); err == nil {
			time.Sleep(100 * time.Millisecond)
			return nil
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(50 * time.Millisecond):
		}
	}
}

func shutdown(consumerErr <-chan error) {
	consumerApp.Stop()
	if natsConn != nil {
		natsConn.Close()
	}

	select {
	case <-consumerErr:
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
	sender.reset()
	if tasksRecorder != nil {
		tasksRecorder.reset()
	}
}

func waitEmailCount(t *testing.T, want int) {
	t.Helper()
	deadline := time.Now().Add(10 * time.Second)
	for time.Now().Before(deadline) {
		if sender.count() == want {
			return
		}
		time.Sleep(50 * time.Millisecond)
	}
	t.Fatalf("sent emails = %d, want %d", sender.count(), want)
}

// waitTasksReport - ждёт, пока mock tasks-сервис получит нужное число отчётов.
func waitTasksReport(t *testing.T, want int) {
	t.Helper()
	deadline := time.Now().Add(10 * time.Second)
	for time.Now().Before(deadline) {
		if tasksRecorder.count() >= want {
			return
		}
		time.Sleep(50 * time.Millisecond)
	}
	t.Fatalf("tasks reports = %d, want %d", tasksRecorder.count(), want)
}
