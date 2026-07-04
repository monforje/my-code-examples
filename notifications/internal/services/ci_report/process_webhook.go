package cireportservice

import (
	"context"
	"errors"
	"time"

	apperrors "notifications/pkg/errors"

	"notifications/internal/models/domain"

	"github.com/redis/go-redis/v9"
)

// ProcessWebhook - обрабатывает входящий вебхук от task_runner.
/*
    1. Валидировать payload (event, uid, commit обязательны).
    2. Для ci_started — создать pending-отчёт, сохранить в кеш и форвардить в tasks.
    3. Для ci_finished — определить статус, сохранить отчёт в кеш и форвардить в tasks.
       (Парсинг stdout в шаги выполняется сервисом отчётов через канонический JSON ci-translator.)
*/
func (s *CIReportService) ProcessWebhook(ctx context.Context, payload *domain.WebhookPayload) error {
	const op = "CIReportService.ProcessWebhook"

	// 1.
	if payload == nil {
		return apperrors.New(op, ErrInvalidWebhook)
	}
	if payload.UID == "" || payload.Commit == "" || payload.Event == "" {
		return apperrors.New(op, ErrInvalidWebhook)
	}

	switch payload.Event {
	case domain.EventCIStarted:
		// 2.
		return s.processStarted(ctx, payload)
	case domain.EventCIFinished:
		// 3.
		return s.processFinished(ctx, payload)
	default:
		return apperrors.New(op, ErrInvalidWebhook)
	}
}

// processStarted - создаёт pending-отчёт, кеширует и форвардит в tasks.
func (s *CIReportService) processStarted(ctx context.Context, payload *domain.WebhookPayload) error {
	const op = "CIReportService.processStarted"

	report := &domain.CIReport{
		UID:       payload.UID,
		Commit:    payload.Commit,
		Status:    domain.StatusPending,
		ExitCode:  0,
		Stage:     "started",
		Stdout:    "",
		Stderr:    "",
		CreatedAt: time.Now().UTC(),
	}

	if err := s.cache.Save(ctx, report); err != nil {
		return apperrors.New(op, err)
	}

	// Форвард pending в tasks — tasks становится единым источником правды для UI
	// (включая live-статус pending → финал). Идемпотентность по (uid, commit) в tasks.
	if s.forwarder != nil {
		s.forwarder.SendAsync(ctx, report)
	}

	s.logger.Info(ctx, op, "ci started cached + forwarded",
		"uid", payload.UID,
		"commit", payload.Commit,
	)
	return nil
}

// processFinished - определяет статус, кеширует и форвардит отчёт в tasks.
func (s *CIReportService) processFinished(ctx context.Context, payload *domain.WebhookPayload) error {
	const op = "CIReportService.processFinished"

	// 1. Извлечь exit_code (default 0 если nil).
	var exitCode int32
	if payload.ExitCode != nil {
		exitCode = *payload.ExitCode
	}

	// 2. Определить статус прогона (passed/failed/crashed/timeout).
	stage := payload.Stage
	if stage == "" {
		stage = "run"
	}
	status := determineStatus(exitCode, stage)

	// 3. Собрать отчёт (stdout хранится как-is; шаги парсятся позже из JSON ci-translator).
	report := &domain.CIReport{
		UID:       payload.UID,
		Commit:    payload.Commit,
		Status:    status,
		ExitCode:  exitCode,
		Stage:     stage,
		Stdout:    payload.Stdout,
		Stderr:    payload.Stderr,
		CreatedAt: time.Now().UTC(),
	}

	// 4. Сохранить в кеш.
	if err := s.cache.Save(ctx, report); err != nil {
		return apperrors.New(op, err)
	}

	// 5. Асинхронно отправить нормализованный отчёт в tasks (после сохранения в redis).
	if s.forwarder != nil {
		s.forwarder.SendAsync(ctx, report)
	}

	s.logger.Info(ctx, op, "ci finished cached + forwarded",
		"uid", payload.UID,
		"commit", payload.Commit,
		"status", string(status),
		"exit_code", exitCode,
	)
	return nil
}

// GetReport - возвращает последний CI-отчёт по UID.
func (s *CIReportService) GetReport(ctx context.Context, uid string) (*domain.CIReport, error) {
	const op = "CIReportService.GetReport"

	report, err := s.cache.GetLatest(ctx, uid)
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, apperrors.New(op, ErrReportNotFound)
		}
		return nil, apperrors.New(op, err)
	}

	return report, nil
}

// determineStatus - определяет итоговый статус CI из exit_code и stage.
// Используется для кеша и для fallback root_cause при vm_crash/timeout.
func determineStatus(exitCode int32, stage string) domain.CIReportStatus {
	if stage == "vm_crash" {
		return domain.StatusCrashed
	}
	if stage == "timeout" {
		return domain.StatusTimeout
	}
	if exitCode == 0 {
		return domain.StatusPassed
	}
	return domain.StatusFailed
}
