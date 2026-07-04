// Package reportservice сериализует domain.CIReport в domain.Report
// и асинхронно отправляет его в сервис tasks по HTTP с ретраями.
package reportservice

import (
	"context"
	"time"

	"notifications/internal/models/domain"
	clientsdto "notifications/pkg/http_clients/dto"
	"notifications/pkg/logger"
)

// reportSender - интерфейс HTTP-клиента для отправки отчёта в tasks.
type reportSender interface {
	SendReport(ctx context.Context, req *clientsdto.CreateReportRequest) error
}

type sender = reportSender

// Параметры ретраев доставки в tasks: покрывают кратковременную недоступность tasks.
const (
	sendMaxAttempts = 3
	sendBaseBackoff = 400 * time.Millisecond
)

// ReportService - сервис сериализации и отправки CI-отчёта.
type ReportService struct {
	sender sender
	logger *logger.Logger
}

// NewReportService - конструктор сервиса отчётов.
func NewReportService(s sender, log *logger.Logger) *ReportService {
	return &ReportService{
		sender: s,
		logger: log,
	}
}

// SendAsync - конвертирует ci_report в report и асинхронно отправляет в tasks.
// Запускается в горутине с отвязанным контекстом (переживает завершение вебхука)
// и ретраится до sendMaxAttempts с экспоненциальным backoff.
func (s *ReportService) SendAsync(ctx context.Context, r *domain.CIReport) {
	if r == nil {
		return
	}

	asyncCtx := context.WithoutCancel(ctx)
	go s.send(asyncCtx, r)
}

// send - конвертирует и отправляет отчёт с ретраями. Внутренний, вызывается в горутине.
/*
   Алгоритм:
   1. Конвертировать ci_report -> report -> dto.
   2. Повторять отправку до sendMaxAttempts с backoff, пока не успех или контекст отменён.
*/
func (s *ReportService) send(ctx context.Context, r *domain.CIReport) {
	const op = "ReportService.send"

	// 1.
	report := FromCIReport(r)
	req := toDTO(report)

	// 2.
	for attempt := 1; ; attempt++ {
		err := s.sender.SendReport(ctx, req)
		if err == nil {
			s.logger.Info(ctx, op, "report sent",
				"uid", r.UID,
				"commit", r.Commit,
				"status", string(report.Status),
				"attempt", attempt,
			)
			return
		}

		s.logger.Error(ctx, op, "failed to send report",
			"uid", r.UID,
			"commit", r.Commit,
			"err", err.Error(),
			"attempt", attempt,
		)

		if attempt >= sendMaxAttempts {
			return
		}

		select {
		case <-time.After(sendBaseBackoff * time.Duration(1<<(attempt-1))):
		case <-ctx.Done():
			return
		}
	}
}
