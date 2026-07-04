package reportshandler

import (
	httpserver "tasks/internal/http/gen"

	"tasks/internal/models/domain"
	reportsservice "tasks/internal/services/reports"
)

func strPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

// mapCreateRequest - httpserver.CreateReportRequest -> CreateReportInput.
func mapCreateRequest(req *httpserver.CreateReportRequest) *reportsservice.CreateReportInput {
	steps := make([]domain.ReportStep, 0, len(req.Steps))
	for _, s := range req.Steps {
		steps = append(steps, domain.ReportStep{
			Index:      s.Index,
			Name:       s.Name,
			Status:     domain.ReportStepStatus(s.Status),
			Code:       derefStr(s.Code),
			Failure:    derefStr(s.Failure),
			HTTPStatus: derefStr(s.HttpStatus),
		})
	}

	lint := make([]domain.ReportLintError, 0, len(req.LintErrors))
	for _, l := range req.LintErrors {
		lint = append(lint, domain.ReportLintError{
			File:    l.File,
			Line:    l.Line,
			Col:     l.Col,
			Rule:    l.Rule,
			Message: l.Message,
		})
	}

	return &reportsservice.CreateReportInput{
		UID:    req.Uid,
		Commit: req.Commit,
		RunID:  derefStr(req.RunId),
		Status: domain.ReportStatus(req.Status),
		Summary: domain.ReportSummary{
			Status:    string(req.Summary.Status),
			Message:   req.Summary.Message,
			RootCause: derefStr(req.Summary.RootCause),
			Passed:    req.Summary.Passed,
			Failed:    req.Summary.Failed,
			Blocked:   req.Summary.Blocked,
			Warnings:  req.Summary.Warnings,
		},
		CreatedAt:       req.CreatedAt,
		Steps:           steps,
		LintErrors:      lint,
		Warnings:        req.Warnings,
		RawLogAvailable: req.RawLogAvailable,
	}
}

// mapReportOutput - ReportOutput -> httpserver.Report.
func mapReportOutput(out *reportsservice.ReportOutput) httpserver.Report {
	steps := make([]httpserver.ReportStep, 0, len(out.Report.Steps))
	for _, s := range out.Report.Steps {
		steps = append(steps, httpserver.ReportStep{
			Index:      s.Index,
			Name:       s.Name,
			Status:     httpserver.ReportStepStatus(s.Status),
			Code:       strPtr(s.Code),
			Failure:    strPtr(s.Failure),
			HttpStatus: strPtr(s.HTTPStatus),
		})
	}

	lint := make([]httpserver.ReportLintError, 0, len(out.Report.LintErrors))
	for _, l := range out.Report.LintErrors {
		lint = append(lint, httpserver.ReportLintError{
			File:    l.File,
			Line:    l.Line,
			Col:     l.Col,
			Rule:    l.Rule,
			Message: l.Message,
		})
	}

	warnings := out.Report.Warnings
	if warnings == nil {
		warnings = []string{}
	}

	return httpserver.Report{
		Id:        out.ID.String(),
		Uid:       out.Report.UID,
		Commit:    out.Report.Commit,
		RunId:     strPtr(out.Report.RunID),
		Status:    httpserver.ReportStatus(out.Report.Status),
		CreatedAt: out.Report.CreatedAt,
		Summary: httpserver.ReportSummary{
			Status:    httpserver.ReportSummaryStatus(out.Report.Summary.Status),
			Message:   out.Report.Summary.Message,
			RootCause: strPtr(out.Report.Summary.RootCause),
			Passed:    out.Report.Summary.Passed,
			Failed:    out.Report.Summary.Failed,
			Blocked:   out.Report.Summary.Blocked,
			Warnings:  out.Report.Summary.Warnings,
		},
		Steps:           steps,
		LintErrors:      lint,
		Warnings:        warnings,
		RawLogAvailable: out.Report.RawLogAvailable,
	}
}

func derefStr(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}
