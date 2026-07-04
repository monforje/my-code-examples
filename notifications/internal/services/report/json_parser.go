package reportservice

import (
	"encoding/json"
	"strings"

	"notifications/internal/models/domain"
)

// ciTranslatorStep — шаг в формате ci-translator JSON (ci-json-contract.md).
type ciTranslatorStep struct {
	Index      int    `json:"index"`
	Name       string `json:"name"`
	Passed     bool   `json:"passed"`
	Status     string `json:"status"`
	Code       string `json:"code,omitempty"`
	Failure    string `json:"failure,omitempty"`
	HTTPStatus string `json:"http_status,omitempty"`
}

// ciTranslatorSummary — сводка в формате ci-translator.
type ciTranslatorSummary struct {
	Status    string `json:"status"`
	Message   string `json:"message"`
	RootCause string `json:"root_cause,omitempty"`
	Passed    int    `json:"passed"`
	Failed    int    `json:"failed"`
	Blocked   int    `json:"blocked"`
	Warnings  int    `json:"warnings"`
}

// ciTranslatorLintError — lint-ошибка в формате ci-translator.
type ciTranslatorLintError struct {
	File    string `json:"file"`
	Line    int    `json:"line"`
	Col     int    `json:"col"`
	Rule    string `json:"rule"`
	Message string `json:"message"`
}

// ciTranslatorPayload — полный JSON ci-translator.
type ciTranslatorPayload struct {
	RunID          string                `json:"run_id"`
	Passed         bool                  `json:"passed"`
	Summary        ciTranslatorSummary    `json:"summary"`
	Steps          []ciTranslatorStep     `json:"steps"`
	Warnings       []string               `json:"warnings"`
	LintErrors     []ciTranslatorLintError `json:"lint_errors"`
	RawLogAvailable bool                  `json:"raw_log_available"`
}

// parseCITranslatorJSON парсит stdout в формате ci-translator и возвращает domain.Report.
// Возвращает (report, true) при успехе, (nil, false) если stdout не JSON или не соответствует контракту.
func parseCITranslatorJSON(stdout string) (*domain.Report, bool) {
	trimmed := strings.TrimSpace(stdout)
	if trimmed == "" || trimmed[0] != '{' {
		return nil, false
	}

	var raw ciTranslatorPayload
	if err := json.Unmarshal([]byte(trimmed), &raw); err != nil {
		return nil, false
	}

	// Валидация: должен быть summary с непустым status.
	if raw.Summary.Status == "" {
		return nil, false
	}

	steps := make([]domain.ReportStep, 0, len(raw.Steps))
	for _, s := range raw.Steps {
		stepStatus := mapCITranslatorStepStatus(s.Status)
		steps = append(steps, domain.ReportStep{
			Index:      s.Index,
			Name:       s.Name,
			Status:     stepStatus,
			Code:       s.Code,
			Failure:    s.Failure,
			HTTPStatus: s.HTTPStatus,
		})
	}

	lintErrors := make([]domain.ReportLintError, 0, len(raw.LintErrors))
	for _, l := range raw.LintErrors {
		lintErrors = append(lintErrors, domain.ReportLintError{
			File:    l.File,
			Line:    l.Line,
			Col:     l.Col,
			Rule:    l.Rule,
			Message: l.Message,
		})
	}

	warnings := raw.Warnings
	if warnings == nil {
		warnings = []string{}
	}

	report := &domain.Report{
		RunID: raw.RunID,
		Summary: domain.ReportSummary{
			Status:    raw.Summary.Status,
			Message:   raw.Summary.Message,
			RootCause: raw.Summary.RootCause,
			Passed:    raw.Summary.Passed,
			Failed:    raw.Summary.Failed,
			Blocked:   raw.Summary.Blocked,
			Warnings:  raw.Summary.Warnings,
		},
		Steps:           steps,
		LintErrors:      lintErrors,
		Warnings:        warnings,
		RawLogAvailable: raw.RawLogAvailable,
	}

	return report, true
}

// mapCITranslatorStepStatus переводит статус шага из JSON в ReportStepStatus.
func mapCITranslatorStepStatus(s string) domain.ReportStepStatus {
	switch domain.ReportStepStatus(s) {
	case domain.ReportStepPassed:
		return domain.ReportStepPassed
	case domain.ReportStepBlocked:
		return domain.ReportStepBlocked
	case domain.ReportStepWarning:
		return domain.ReportStepWarning
	default:
		return domain.ReportStepFailed
	}
}
