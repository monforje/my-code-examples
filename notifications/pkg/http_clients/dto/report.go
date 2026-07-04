// Package clientsdto содержит DTO для обмена между сервисами по HTTP.
package clientsdto

import "time"

// CreateReportRequest - DTO CI-отчёта между notifications и tasks.
// Имена json-тегов соответствуют контракту ci-json-contract.md / tasks typespec.
type CreateReportRequest struct {
	UID             string            `json:"uid"`
	Commit          string            `json:"commit"`
	RunID           string            `json:"run_id,omitempty"`
	Status          string            `json:"status"`
	CreatedAt       time.Time         `json:"created_at"`
	Summary         ReportSummary     `json:"summary"`
	Steps           []ReportStep      `json:"steps"`
	LintErrors      []ReportLintError `json:"lint_errors"`
	Warnings        []string          `json:"warnings"`
	RawLogAvailable bool              `json:"raw_log_available"`
}

// ReportSummary - краткая сводка результата CI.
type ReportSummary struct {
	Status    string `json:"status"`
	Message   string `json:"message"`
	RootCause string `json:"root_cause,omitempty"`
	Passed    int    `json:"passed"`
	Failed    int    `json:"failed"`
	Blocked   int    `json:"blocked"`
	Warnings  int    `json:"warnings"`
}

// ReportStep - нормализованный шаг пайплайна.
type ReportStep struct {
	Index      int    `json:"index"`
	Name       string `json:"name"`
	Status     string `json:"status"`
	Code       string `json:"code,omitempty"`
	Failure    string `json:"failure,omitempty"`
	HTTPStatus string `json:"http_status,omitempty"`
}

// ReportLintError - нормализованная ошибка golangci-lint.
type ReportLintError struct {
	File    string `json:"file"`
	Line    int    `json:"line"`
	Col     int    `json:"col"`
	Rule    string `json:"rule"`
	Message string `json:"message"`
}
