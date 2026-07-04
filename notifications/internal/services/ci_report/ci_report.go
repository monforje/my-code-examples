// Package cireportservice
package cireportservice

import (
	"context"
	"errors"
	"time"

	"notifications/internal/config"
	"notifications/internal/models/domain"
	"notifications/pkg/logger"
)

var (
	// ErrReportNotFound — отчёт не найден в кеше.
	ErrReportNotFound = errors.New("ci report not found")
	// ErrInvalidWebhook — некорректный вебхук.
	ErrInvalidWebhook = errors.New("invalid webhook payload")
)

// reportCache — интерфейс кеша CI-отчётов.
type reportCache interface {
	Save(ctx context.Context, report *domain.CIReport) error
	GetLatest(ctx context.Context, uid string) (*domain.CIReport, error)
}

type cache = reportCache

// reportForwarder — интерфейс сервиса, который сериализует отчёт и отправляет его в tasks.
type reportForwarder interface {
	SendAsync(ctx context.Context, report *domain.CIReport)
}

type forwarder = reportForwarder

// CIReportService — сервис обработки CI-отчётов.
type CIReportService struct {
	cache     cache
	forwarder forwarder
	logger    *logger.Logger
	ttl       time.Duration
}

// NewCIReportService - конструктор сервиса отчётов.
func NewCIReportService(c cache, f forwarder, log *logger.Logger, cfg config.FeaturesConfig) *CIReportService {
	return &CIReportService{
		cache:     c,
		forwarder: f,
		logger:    log,
		ttl:       cfg.CIReportCacheTTL,
	}
}
