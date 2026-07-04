// Package reportsservice
package reportsservice

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"

	"tasks/internal/models/domain"
	"tasks/internal/models/records"
	clientsdto "tasks/pkg/http_clients/dto"
)

var (
	// ErrUIDEmpty - пустой uid отчёта.
	ErrUIDEmpty = errors.New("uid is empty")
	// ErrCommitEmpty - пустой commit отчёта.
	ErrCommitEmpty = errors.New("commit is empty")
	// ErrRunIDInvalid - run_id передан, но не является корректным UUID.
	ErrRunIDInvalid = errors.New("run_id is invalid")
	// ErrReportIDInvalid - id отчёта не является корректным UUID.
	ErrReportIDInvalid = errors.New("report id is invalid")
	// ErrTaskNameEmpty - пустой task_name в запросе списка.
	ErrTaskNameEmpty = errors.New("task_name is empty")
	// ErrUIDInvalid - uid не вида "{username}/golden-{task_name}".
	ErrUIDInvalid = errors.New("uid is invalid")
	// ErrTaskNotFound - задача для отчёта не найдена.
	ErrTaskNotFound = errors.New("task not found")
	// ErrReportNotFound - отчёт не найден (или недоступен пользователю).
	ErrReportNotFound = errors.New("report not found")
	// ErrUsernameNotFound - не удалось резолвить username пользователя (users-сервис).
	ErrUsernameNotFound = errors.New("user username not found")
)

// reportPayload - JSON-представление полного отчёта для хранения в JSONB.
// Имена тегов соответствуют контракту ci-json-contract.md.
type reportPayload struct {
	Status          string            `json:"status"`
	Summary         reportSummary     `json:"summary"`
	Steps           []reportStep      `json:"steps"`
	LintErrors      []reportLintError `json:"lint_errors"`
	Warnings        []string          `json:"warnings"`
	RawLogAvailable bool              `json:"raw_log_available"`
}

type reportSummary struct {
	Status    string `json:"status"`
	Message   string `json:"message"`
	RootCause string `json:"root_cause,omitempty"`
	Passed    int    `json:"passed"`
	Failed    int    `json:"failed"`
	Blocked   int    `json:"blocked"`
	Warnings  int    `json:"warnings"`
}

type reportStep struct {
	Index      int    `json:"index"`
	Name       string `json:"name"`
	Status     string `json:"status"`
	Code       string `json:"code,omitempty"`
	Failure    string `json:"failure,omitempty"`
	HTTPStatus string `json:"http_status,omitempty"`
}

type reportLintError struct {
	File    string `json:"file"`
	Line    int    `json:"line"`
	Col     int    `json:"col"`
	Rule    string `json:"rule"`
	Message string `json:"message"`
}

// reportRepository - интерфейс хранилища CI-отчётов.
type reportRepository interface {
	Upsert(ctx context.Context, rep *records.CIReport) error
	GetByIDAndUsername(
		ctx context.Context,
		id uuid.UUID,
		username string,
	) (*records.CIReport, error)
	ListByUsernameAndTask(
		ctx context.Context,
		username string,
		taskName string,
		status *string,
		limit int32,
		cursor *string,
	) ([]records.CIReport, bool, error)
	ListByUsername(
		ctx context.Context,
		username string,
		status *string,
		limit int32,
		cursor *string,
	) ([]records.CIReport, bool, error)
}

// taskRepository - интерфейс получения задачи по task_name (валидация FK).
type taskRepository interface {
	GetByTaskName(ctx context.Context, taskName string) (*records.Task, error)
}

// usersClient - интерфейс резолва identity_id → username (для GET).
type usersClient interface {
	GetGitUser(ctx context.Context, req *clientsdto.GitUserRequest) (*clientsdto.GitUserResponse, error)
}

// ReportsService - сервис приёма и выдачи CI-отчётов.
type ReportsService struct {
	reports reportRepository
	tasks   taskRepository
	users   usersClient
}

// NewReportsService - конструктор сервиса отчётов.
func NewReportsService(r reportRepository, t taskRepository, u usersClient) *ReportsService {
	return &ReportsService{reports: r, tasks: t, users: u}
}

// toPayload - переводит доменную модель в JSON-payload для хранения.
func toPayload(in *CreateReportInput) *reportPayload {
	steps := make([]reportStep, 0, len(in.Steps))
	for _, s := range in.Steps {
		steps = append(steps, reportStep{
			Index: s.Index, Name: s.Name, Status: string(s.Status),
			Code: s.Code, Failure: s.Failure, HTTPStatus: s.HTTPStatus,
		})
	}

	lint := make([]reportLintError, 0, len(in.LintErrors))
	for _, l := range in.LintErrors {
		lint = append(lint, reportLintError{
			File: l.File, Line: l.Line, Col: l.Col, Rule: l.Rule, Message: l.Message,
		})
	}

	warnings := in.Warnings
	if warnings == nil {
		warnings = []string{}
	}

	return &reportPayload{
		Status: string(in.Status),
		Summary: reportSummary{
			Status: in.Summary.Status, Message: in.Summary.Message, RootCause: in.Summary.RootCause,
			Passed: in.Summary.Passed, Failed: in.Summary.Failed, Blocked: in.Summary.Blocked, Warnings: in.Summary.Warnings,
		},
		Steps:           steps,
		LintErrors:      lint,
		Warnings:        warnings,
		RawLogAvailable: in.RawLogAvailable,
	}
}

// payloadToReport - разбирает JSON-payload обратно в доменную модель.
// uid/commit/run_id/created_at берутся из колонок записи (canonical source).
func payloadToReport(uid, commit, runID string, createdAt time.Time, raw []byte) (*domain.Report, error) {
	var p reportPayload
	if err := json.Unmarshal(raw, &p); err != nil {
		return nil, err
	}

	status := domain.ReportStatus(p.Status)
	if status == "" {
		status = domain.ReportStatusFailed
	}

	steps := make([]domain.ReportStep, 0, len(p.Steps))
	for _, s := range p.Steps {
		steps = append(steps, domain.ReportStep{
			Index: s.Index, Name: s.Name, Status: domain.ReportStepStatus(s.Status),
			Code: s.Code, Failure: s.Failure, HTTPStatus: s.HTTPStatus,
		})
	}

	lint := make([]domain.ReportLintError, 0, len(p.LintErrors))
	for _, l := range p.LintErrors {
		lint = append(lint, domain.ReportLintError{
			File: l.File, Line: l.Line, Col: l.Col, Rule: l.Rule, Message: l.Message,
		})
	}

	return &domain.Report{
		UID:    uid,
		Commit: commit,
		RunID:  runID,
		Status: status,
		Summary: domain.ReportSummary{
			Status: p.Summary.Status, Message: p.Summary.Message, RootCause: p.Summary.RootCause,
			Passed: p.Summary.Passed, Failed: p.Summary.Failed, Blocked: p.Summary.Blocked, Warnings: p.Summary.Warnings,
		},
		CreatedAt:       createdAt,
		Steps:           steps,
		LintErrors:      lint,
		Warnings:        p.Warnings,
		RawLogAvailable: p.RawLogAvailable,
	}, nil
}

// parseUID - разбирает uid вида "{username}/golden-{task_name}".
/*
   Алгоритм:
   1. Разделить uid по первому '/'.
   2. username = часть до '/', goldenRepoName = часть после.
   3. task_name = goldenRepoName без префикса "golden-".
*/
func parseUID(uid string) (username, taskName string, err error) {
	// 1.
	idx := strings.IndexByte(uid, '/')
	if idx < 0 {
		return "", "", ErrUIDInvalid
	}

	// 2.
	username = uid[:idx]
	goldenRepoName := uid[idx+1:]

	// 3.
	const goldenPrefix = "golden-"
	taskName = strings.TrimPrefix(goldenRepoName, goldenPrefix)
	if taskName == goldenRepoName || taskName == "" || username == "" {
		return "", "", ErrUIDInvalid
	}
	return username, taskName, nil
}
