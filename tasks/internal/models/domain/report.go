// Package domain
package domain

import "time"

// ReportStatus - итоговый статус CI-отчёта.
type ReportStatus string

const (
	ReportStatusPending ReportStatus = "pending"
	ReportStatusPassed  ReportStatus = "passed"
	ReportStatusFailed  ReportStatus = "failed"
)

// ReportSummary - краткая сводка результата для верхней карты UI.
type ReportSummary struct {
	Status    string
	Message   string
	RootCause string
	Passed    int
	Failed    int
	Blocked   int
	Warnings  int
}

// ReportStepStatus - статус отдельного шага.
type ReportStepStatus string

const (
	ReportStepPassed  ReportStepStatus = "passed"
	ReportStepFailed  ReportStepStatus = "failed"
	ReportStepBlocked ReportStepStatus = "blocked"
	ReportStepWarning ReportStepStatus = "warning"
)

// ReportStep - нормализованный шаг пайплайна.
type ReportStep struct {
	Index      int
	Name       string
	Status     ReportStepStatus
	Code       string
	Failure    string
	HTTPStatus string
}

// ReportLintError - нормализованная ошибка golangci-lint.
type ReportLintError struct {
	File    string
	Line    int
	Col     int
	Rule    string
	Message string
}

// Report - доменная модель CI-отчёта, получаемого от notifications.
type Report struct {
	UID             string
	Commit          string
	RunID           string
	Status          ReportStatus
	CreatedAt       time.Time
	Summary         ReportSummary
	Steps           []ReportStep
	LintErrors      []ReportLintError
	Warnings        []string
	RawLogAvailable bool
}
