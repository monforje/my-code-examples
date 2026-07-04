// Package handlers
package handlers

import (
	"context"

	"notifications/internal/models/domain"
)

type ciReportService interface {
	ProcessWebhook(ctx context.Context, payload *domain.WebhookPayload) error
	GetReport(ctx context.Context, uid string) (*domain.CIReport, error)
}

type CiReportHandlers struct {
	svc ciReportService
}

func NewCiReportHandlers(svc ciReportService) *CiReportHandlers {
	return &CiReportHandlers{svc: svc}
}
