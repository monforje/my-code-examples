package cireportservice_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"notifications/internal/config"
	"notifications/internal/models/domain"
	cireportservice "notifications/internal/services/ci_report"
	"notifications/internal/services/ci_report/mocks"
	"notifications/pkg/logger"

	"github.com/redis/go-redis/v9"
	"go.uber.org/mock/gomock"
)

func newService(t *testing.T) (*cireportservice.CIReportService, *mocks.MockReportCache) {
	t.Helper()
	ctrl := gomock.NewController(t)

	mockCache := mocks.NewMockReportCache(ctrl)
	log := logger.New(&config.LoggerConfig{Level: -8, Format: config.FormatText})
	cfg := config.FeaturesConfig{CIReportCacheTTL: time.Hour}

	svc := cireportservice.NewCIReportService(mockCache, noopForwarder{}, log, cfg)
	return svc, mockCache
}

// noopForwarder - заглушка reportForwarder для тестов, не проверяющих отправку.
type noopForwarder struct{}

func (noopForwarder) SendAsync(context.Context, *domain.CIReport) {}

func ptrInt32(v int32) *int32 { return &v }

func TestProcessWebhook_CIStarted(t *testing.T) {
	svc, mockCache := newService(t)

	mockCache.EXPECT().
		Save(gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, r *domain.CIReport) error {
			if r.Status != domain.StatusPending {
				t.Errorf("status = %q, want pending", r.Status)
			}
			if r.UID != "alice/golden-pizza-api" {
				t.Errorf("uid = %q", r.UID)
			}
			return nil
		})

	payload := &domain.WebhookPayload{
		Event:  domain.EventCIStarted,
		UID:    "alice/golden-pizza-api",
		Commit: "abc123",
	}

	if err := svc.ProcessWebhook(context.Background(), payload); err != nil {
		t.Fatalf("ProcessWebhook() error = %v", err)
	}
}

func TestProcessWebhook_CIFinished_Passed(t *testing.T) {
	svc, mockCache := newService(t)

	mockCache.EXPECT().
		Save(gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, r *domain.CIReport) error {
			if r.Status != domain.StatusPassed {
				t.Errorf("status = %q, want passed", r.Status)
			}
			if r.ExitCode != 0 {
				t.Errorf("exitCode = %d, want 0", r.ExitCode)
			}
			// Шаги не парсятся здесь (это делает ci-translator JSON-путь в сервисе отчётов).
			if len(r.Steps) != 0 {
				t.Errorf("steps len = %d, want 0 (not parsed here)", len(r.Steps))
			}
			return nil
		})

	stdout := `{"run_id":"r1","summary":{"status":"passed","message":"ok","passed":2,"failed":0,"blocked":0,"warnings":0},"steps":[{"index":1,"name":"build","status":"passed"}],"warnings":[],"lint_errors":[],"raw_log_available":false}`
	payload := &domain.WebhookPayload{
		Event:    domain.EventCIFinished,
		UID:      "alice/golden-pizza-api",
		Commit:   "abc123",
		ExitCode: ptrInt32(0),
		Stdout:   stdout,
		Stage:    "run",
	}

	if err := svc.ProcessWebhook(context.Background(), payload); err != nil {
		t.Fatalf("ProcessWebhook() error = %v", err)
	}
}

func TestProcessWebhook_CIFinished_Failed(t *testing.T) {
	svc, mockCache := newService(t)

	mockCache.EXPECT().
		Save(gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, r *domain.CIReport) error {
			if r.Status != domain.StatusFailed {
				t.Errorf("status = %q, want failed", r.Status)
			}
			return nil
		})

	payload := &domain.WebhookPayload{
		Event:    domain.EventCIFinished,
		UID:      "alice/golden-pizza-api",
		Commit:   "abc123",
		ExitCode: ptrInt32(1),
		Stdout:   "✗ test (0.5s) - exit code 1\n",
		Stage:    "run",
	}

	if err := svc.ProcessWebhook(context.Background(), payload); err != nil {
		t.Fatalf("ProcessWebhook() error = %v", err)
	}
}

func TestProcessWebhook_CIFinished_Crashed(t *testing.T) {
	svc, mockCache := newService(t)

	mockCache.EXPECT().
		Save(gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, r *domain.CIReport) error {
			if r.Status != domain.StatusCrashed {
				t.Errorf("status = %q, want crashed", r.Status)
			}
			return nil
		})

	payload := &domain.WebhookPayload{
		Event:    domain.EventCIFinished,
		UID:      "alice/golden-pizza-api",
		Commit:   "abc123",
		ExitCode: ptrInt32(-1),
		Stage:    "vm_crash",
		Stderr:   "VM connection lost",
	}

	if err := svc.ProcessWebhook(context.Background(), payload); err != nil {
		t.Fatalf("ProcessWebhook() error = %v", err)
	}
}

func TestProcessWebhook_InvalidPayload(t *testing.T) {
	svc, _ := newService(t)

	tests := []struct {
		name    string
		payload *domain.WebhookPayload
	}{
		{"nil", nil},
		{"empty_uid", &domain.WebhookPayload{Event: domain.EventCIStarted, Commit: "abc"}},
		{"empty_commit", &domain.WebhookPayload{Event: domain.EventCIStarted, UID: "alice"}},
		{"empty_event", &domain.WebhookPayload{UID: "alice", Commit: "abc"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := svc.ProcessWebhook(context.Background(), tt.payload)
			if !errors.Is(err, cireportservice.ErrInvalidWebhook) {
				t.Fatalf("error = %v, want ErrInvalidWebhook", err)
			}
		})
	}
}

func TestProcessWebhook_CacheError(t *testing.T) {
	svc, mockCache := newService(t)

	mockCache.EXPECT().
		Save(gomock.Any(), gomock.Any()).
		Return(errors.New("redis down"))

	payload := &domain.WebhookPayload{
		Event:  domain.EventCIStarted,
		UID:    "alice/golden-pizza-api",
		Commit: "abc123",
	}

	err := svc.ProcessWebhook(context.Background(), payload)
	if err == nil {
		t.Fatal("error = nil, want error")
	}
}

func TestGetReport_Success(t *testing.T) {
	svc, mockCache := newService(t)

	want := &domain.CIReport{
		UID:     "alice/golden-pizza-api",
		Commit:  "abc123",
		Status:  domain.StatusPassed,
		ExitCode: 0,
	}

	mockCache.EXPECT().
		GetLatest(gomock.Any(), "alice/golden-pizza-api").
		Return(want, nil)

	got, err := svc.GetReport(context.Background(), "alice/golden-pizza-api")
	if err != nil {
		t.Fatalf("GetReport() error = %v", err)
	}
	if got.UID != want.UID || got.Status != want.Status {
		t.Fatalf("got = %+v", got)
	}
}

func TestGetReport_NotFound(t *testing.T) {
	svc, mockCache := newService(t)

	mockCache.EXPECT().
		GetLatest(gomock.Any(), "nobody/golden-pizza-api").
		Return(nil, redis.Nil)

	_, err := svc.GetReport(context.Background(), "nobody/golden-pizza-api")
	if err == nil {
		t.Fatal("error = nil, want error")
	}
}

func newServiceWithForwarder(t *testing.T) (*cireportservice.CIReportService, *mocks.MockReportCache, *mocks.MockReportForwarder) {
	t.Helper()
	ctrl := gomock.NewController(t)

	mockCache := mocks.NewMockReportCache(ctrl)
	mockForwarder := mocks.NewMockReportForwarder(ctrl)
	log := logger.New(&config.LoggerConfig{Level: -8, Format: config.FormatText})
	cfg := config.FeaturesConfig{CIReportCacheTTL: time.Hour}

	svc := cireportservice.NewCIReportService(mockCache, mockForwarder, log, cfg)
	return svc, mockCache, mockForwarder
}

// TestProcessWebhook_ForwardsFinished - после сохранения finished-отчёта вызывается forwarder.
func TestProcessWebhook_ForwardsFinished(t *testing.T) {
	svc, mockCache, mockForwarder := newServiceWithForwarder(t)

	gomock.InOrder(
		mockCache.EXPECT().Save(gomock.Any(), gomock.Any()).Return(nil),
		mockForwarder.EXPECT().SendAsync(gomock.Any(), gomock.Any()).
			DoAndReturn(func(_ context.Context, r *domain.CIReport) {
				if r.UID != "alice/golden-pizza-api" {
					t.Errorf("forwarded uid = %q", r.UID)
				}
			}),
	)

	payload := &domain.WebhookPayload{
		Event:    domain.EventCIFinished,
		UID:      "alice/golden-pizza-api",
		Commit:   "abc123",
		ExitCode: ptrInt32(0),
		Stdout:   "✓ build (1.0s)\n",
		Stage:    "run",
	}

	if err := svc.ProcessWebhook(context.Background(), payload); err != nil {
		t.Fatalf("ProcessWebhook() error = %v", err)
	}
}

// TestProcessWebhook_NoForwardOnCacheError - при ошибке сохранения forwarder не вызывается.
func TestProcessWebhook_NoForwardOnCacheError(t *testing.T) {
	svc, mockCache, mockForwarder := newServiceWithForwarder(t)

	mockCache.EXPECT().Save(gomock.Any(), gomock.Any()).Return(errors.New("redis down"))
	mockForwarder.EXPECT().SendAsync(gomock.Any(), gomock.Any()).Times(0)

	payload := &domain.WebhookPayload{
		Event:    domain.EventCIFinished,
		UID:      "alice/golden-pizza-api",
		Commit:   "abc123",
		ExitCode: ptrInt32(0),
		Stage:    "run",
	}

	_ = svc.ProcessWebhook(context.Background(), payload)
}

// TestProcessWebhook_ForwardsStarted - pending-отчёт (ci_started) форвардится в tasks.
func TestProcessWebhook_ForwardsStarted(t *testing.T) {
	svc, mockCache, mockForwarder := newServiceWithForwarder(t)

	gomock.InOrder(
		mockCache.EXPECT().Save(gomock.Any(), gomock.Any()).Return(nil),
		mockForwarder.EXPECT().SendAsync(gomock.Any(), gomock.Any()).
			DoAndReturn(func(_ context.Context, r *domain.CIReport) {
				if r.Status != domain.StatusPending {
					t.Errorf("forwarded status = %q, want pending", r.Status)
				}
				if r.UID != "alice/golden-pizza-api" {
					t.Errorf("forwarded uid = %q", r.UID)
				}
			}),
	)

	payload := &domain.WebhookPayload{
		Event:  domain.EventCIStarted,
		UID:    "alice/golden-pizza-api",
		Commit: "abc123",
	}

	if err := svc.ProcessWebhook(context.Background(), payload); err != nil {
		t.Fatalf("ProcessWebhook() error = %v", err)
	}
}
