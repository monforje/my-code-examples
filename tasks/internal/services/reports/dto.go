// Package reportsservice - приём и выдача CI-отчётов.
package reportsservice

import (
	"time"

	"github.com/google/uuid"

	"tasks/internal/models/domain"
)

// CreateReportInput - входные данные CI-отчёта от notifications.
type CreateReportInput struct {
	UID             string
	Commit          string
	RunID           string
	Status          domain.ReportStatus
	CreatedAt       time.Time
	Summary         domain.ReportSummary
	Steps           []domain.ReportStep
	LintErrors      []domain.ReportLintError
	Warnings        []string
	RawLogAvailable bool
}

// ListInput - параметры запроса списка отчётов.
type ListInput struct {
	IdentityID uuid.UUID
	TaskName   string
	Status     *domain.ReportStatus
	Limit      int32
	Cursor     *string
}

// GetInput - параметры запроса одного отчёта.
type GetInput struct {
	IdentityID uuid.UUID
	ReportID   uuid.UUID
}

// ReportOutput - сохранённый CI-отчёт для отображения в UI.
type ReportOutput struct {
	ID uuid.UUID
	domain.Report
}

// ListOutput - результат запроса списка отчётов.
type ListOutput struct {
	Items       []ReportOutput
	HasNextPage bool
	NextCursor  *string
}
