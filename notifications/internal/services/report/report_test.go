package reportservice

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"

	"notifications/internal/config"
	"notifications/internal/models/domain"
	"notifications/internal/services/report/mocks"
	"notifications/pkg/logger"

	clientsdto "notifications/pkg/http_clients/dto"

	"go.uber.org/mock/gomock"
)

func newLogger() *logger.Logger {
	return logger.New(&config.LoggerConfig{Level: slog.LevelError, Format: config.FormatText, Output: io.Discard})
}

func TestSend_Success(t *testing.T) {
	ctrl := gomock.NewController(t)

	sender := mocks.NewMockReportSender(ctrl)
	svc := NewReportService(sender, newLogger())

	stdout := `{
  "run_id": "run-abc",
  "summary": {"status": "failed", "message": "bad", "passed": 1, "failed": 1, "blocked": 0, "warnings": 0},
  "steps": [
    {"index": 1, "name": "build", "status": "passed"},
    {"index": 2, "name": "test", "status": "failed"}
  ],
  "warnings": [],
  "lint_errors": [],
  "raw_log_available": false
}`

	ci := &domain.CIReport{
		UID:      "alice/golden-pizza-api",
		Commit:   "abc123",
		Status:   domain.StatusFailed,
		ExitCode: 1,
		Stdout:   stdout,
	}

	sender.EXPECT().SendReport(gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, req *clientsdto.CreateReportRequest) error {
			if req.UID != "alice/golden-pizza-api" {
				t.Errorf("uid = %s", req.UID)
			}
			if req.Status != "failed" {
				t.Errorf("status = %s, want failed", req.Status)
			}
			if req.RunID != "run-abc" {
				t.Errorf("run_id = %s, want run-abc", req.RunID)
			}
			if req.Summary.Passed != 1 || req.Summary.Failed != 1 {
				t.Errorf("counts = passed=%d failed=%d", req.Summary.Passed, req.Summary.Failed)
			}
			if len(req.Steps) != 2 {
				t.Errorf("steps len = %d, want 2", len(req.Steps))
			}
			return nil
		})

	svc.send(context.Background(), ci)
}

func TestSend_SendError(t *testing.T) {
	ctrl := gomock.NewController(t)

	sender := mocks.NewMockReportSender(ctrl)
	svc := NewReportService(sender, newLogger())

	// Ошибка отправки ретраится sendMaxAttempts раз и логируется внутри send.
	sender.EXPECT().SendReport(gomock.Any(), gomock.Any()).
		Return(errors.New("tasks unavailable")).
		Times(sendMaxAttempts)

	ci := &domain.CIReport{
		UID:    "alice/golden-pizza-api",
		Commit: "abc123",
		Status: domain.StatusPassed,
	}

	svc.send(context.Background(), ci)
}

func TestSend_RetriesUntilSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)

	sender := mocks.NewMockReportSender(ctrl)
	svc := NewReportService(sender, newLogger())

	// Первые 2 попытки падают, третья — успех.
	gomock.InOrder(
		sender.EXPECT().SendReport(gomock.Any(), gomock.Any()).Return(errors.New("transient")),
		sender.EXPECT().SendReport(gomock.Any(), gomock.Any()).Return(errors.New("transient")),
		sender.EXPECT().SendReport(gomock.Any(), gomock.Any()).Return(nil),
	)

	ci := &domain.CIReport{
		UID:    "alice/golden-pizza-api",
		Commit: "abc123",
		Status: domain.StatusPassed,
	}

	svc.send(context.Background(), ci)
}

func TestSendAsync_ForwardsAndDoesNotBlock(t *testing.T) {
	ctrl := gomock.NewController(t)

	sender := mocks.NewMockReportSender(ctrl)
	svc := NewReportService(sender, newLogger())

	done := make(chan struct{})
	sender.EXPECT().SendReport(gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, _ *clientsdto.CreateReportRequest) error {
			close(done)
			return nil
		})

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // имитируем уже отменённый запрос вебхука — отправка всё равно должна дойти

	svc.SendAsync(ctx, &domain.CIReport{
		UID:    "alice/golden-pizza-api",
		Commit: "abc123",
		Status: domain.StatusPassed,
	})

	<-done
}

func TestSendAsync_NilReport(t *testing.T) {
	ctrl := gomock.NewController(t)
	sender := mocks.NewMockReportSender(ctrl)
	svc := NewReportService(sender, newLogger())

	// nil-отчёт не должен порождать отправку.
	svc.SendAsync(context.Background(), nil)
}
