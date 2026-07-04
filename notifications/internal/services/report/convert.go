package reportservice

import (
	"fmt"
	"strings"

	"notifications/internal/models/domain"
	clientsdto "notifications/pkg/http_clients/dto"
)

// FromCIReport - конвертирует domain.CIReport в нормализованный domain.Report.
/*
   Алгоритм:
   1. Если stdout содержит канонический JSON ci-translator — взять отчёт из него как-is
      (ci-translator = единственный продюсер шагов/сводки/lint). Прокинуть run_id/uid/commit.
   2. Иначе (stdout пустой/не-JSON: vm_crash/timeout/pending) — синтез минимального отчёта
      (только status/message/root_cause). Парсинг woodpecker-вывода удалён как мёртвый код.
*/
func FromCIReport(r *domain.CIReport) *domain.Report {
	if r == nil {
		return nil
	}

	// 1. Главный путь: готовый отчёт от ci-translator.
	if jsonReport, ok := parseCITranslatorJSON(r.Stdout); ok {
		jsonReport.UID = r.UID
		jsonReport.Commit = r.Commit
		// run_id берётся из канонического JSON ci-translator (не перетираем).
		jsonReport.CreatedAt = r.CreatedAt
		jsonReport.Status = mapStatus(r.Status)
		return jsonReport
	}

	// 2. Fallback: минимальный отчёт для стадий без ci-translator.
	status := mapStatus(r.Status)

	var warnings []string
	if strings.TrimSpace(r.Stderr) != "" {
		warnings = append(warnings, r.Stderr)
	}
	if warnings == nil {
		warnings = []string{}
	}

	return &domain.Report{
		UID:       r.UID,
		Commit:    r.Commit,
		RunID:     r.RunID,
		Status:    status,
		CreatedAt: r.CreatedAt,
		Summary: domain.ReportSummary{
			Status:    summaryStatus(status),
			Message:   buildMessage(r),
			RootCause: buildRootCause(r),
			Warnings:  len(warnings),
		},
		Steps:           []domain.ReportStep{},
		Warnings:        warnings,
		RawLogAvailable: strings.TrimSpace(r.Stdout) != "",
	}
}

// toDTO - конвертирует domain.Report в DTO для отправки по HTTP.
func toDTO(r *domain.Report) *clientsdto.CreateReportRequest {
	if r == nil {
		return nil
	}

	steps := make([]clientsdto.ReportStep, 0, len(r.Steps))
	for _, s := range r.Steps {
		steps = append(steps, clientsdto.ReportStep{
			Index:      s.Index,
			Name:       s.Name,
			Status:     string(s.Status),
			Code:       s.Code,
			Failure:    s.Failure,
			HTTPStatus: s.HTTPStatus,
		})
	}

	lintErrors := make([]clientsdto.ReportLintError, 0, len(r.LintErrors))
	for _, l := range r.LintErrors {
		lintErrors = append(lintErrors, clientsdto.ReportLintError{
			File:    l.File,
			Line:    l.Line,
			Col:     l.Col,
			Rule:    l.Rule,
			Message: l.Message,
		})
	}

	warnings := r.Warnings
	if warnings == nil {
		warnings = []string{}
	}

	return &clientsdto.CreateReportRequest{
		UID:       r.UID,
		Commit:    r.Commit,
		RunID:     r.RunID,
		Status:    string(r.Status),
		CreatedAt: r.CreatedAt,
		Summary: clientsdto.ReportSummary{
			Status:    r.Summary.Status,
			Message:   r.Summary.Message,
			RootCause: r.Summary.RootCause,
			Passed:    r.Summary.Passed,
			Failed:    r.Summary.Failed,
			Blocked:   r.Summary.Blocked,
			Warnings:  r.Summary.Warnings,
		},
		Steps:           steps,
		LintErrors:      lintErrors,
		Warnings:        warnings,
		RawLogAvailable: r.RawLogAvailable,
	}
}

// mapStatus - переводит статус CIReport в статус Report.
func mapStatus(s domain.CIReportStatus) domain.ReportStatus {
	switch s {
	case domain.StatusPending:
		return domain.ReportStatusPending
	case domain.StatusPassed:
		return domain.ReportStatusPassed
	default:
		// crashed, timeout и failed — семантически провал.
		return domain.ReportStatusFailed
	}
}

// summaryStatus - статус сводки: passed только при общем успехе.
func summaryStatus(s domain.ReportStatus) string {
	if s == domain.ReportStatusPassed {
		return "passed"
	}
	return "failed"
}

// buildMessage - собирает человекочитаемое сообщение для fallback-отчёта.
func buildMessage(r *domain.CIReport) string {
	switch r.Status {
	case domain.StatusPassed:
		return "CI passed"
	case domain.StatusCrashed:
		if msg := strings.TrimSpace(r.Stderr); msg != "" {
			return fmt.Sprintf("CI crashed: %s", msg)
		}
		return "CI crashed"
	case domain.StatusTimeout:
		return "CI timed out"
	case domain.StatusPending:
		return "CI pending"
	default:
		if r.ExitCode != 0 {
			return fmt.Sprintf("CI failed (exit code %d)", r.ExitCode)
		}
		return "CI failed"
	}
}

// buildRootCause - стабильный машинно-читаемый код первичной причины для fallback-отчёта.
func buildRootCause(r *domain.CIReport) string {
	switch r.Status {
	case domain.StatusCrashed:
		return "VM_CRASHED"
	case domain.StatusTimeout:
		return "TIMEOUT"
	default:
		return ""
	}
}
