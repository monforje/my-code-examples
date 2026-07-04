package reportsservice_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"go.uber.org/mock/gomock"

	"tasks/internal/models/domain"
	"tasks/internal/models/records"
	reportsservice "tasks/internal/services/reports"
	"tasks/internal/services/reports/mocks"
	clientsdto "tasks/pkg/http_clients/dto"
)

func sampleInput(uid string) *reportsservice.CreateReportInput {
	return &reportsservice.CreateReportInput{
		UID:       uid,
		Commit:    "abc123",
		RunID:     uuid.New().String(),
		Status:    domain.ReportStatusFailed,
		CreatedAt: time.Date(2026, 6, 24, 12, 0, 0, 0, time.UTC),
		Summary: domain.ReportSummary{
			Status: "failed", Message: "Application is not reachable on localhost:8080",
			RootCause: "APP_UNREACHABLE", Passed: 4, Failed: 3, Blocked: 5, Warnings: 1,
		},
		Steps: []domain.ReportStep{
			{Index: 1, Name: "Register", Status: domain.ReportStepFailed, Code: "CONNECTION_REFUSED", HTTPStatus: "000"},
		},
		LintErrors:      []domain.ReportLintError{},
		Warnings:        []string{},
		RawLogAvailable: true,
	}
}

func samplePayload() []byte {
	b, _ := json.Marshal(map[string]any{
		"status":           "failed",
		"summary":          map[string]any{"status": "failed", "message": "bad", "passed": 0, "failed": 0, "blocked": 0, "warnings": 0},
		"steps":            []any{},
		"lint_errors":      []any{},
		"warnings":         []any{},
		"raw_log_available": false,
	})
	return b
}

func TestCreateReport_Success(t *testing.T) {
	ctrl := gomock.NewController(t)

	repo := mocks.NewMockReportRepository(ctrl)
	tasks := mocks.NewMockTaskRepository(ctrl)
	users := mocks.NewMockUsersClient(ctrl)
	svc := reportsservice.NewReportsService(repo, tasks, users)

	tasks.EXPECT().GetByTaskName(gomock.Any(), "pizza-api").
		Return(&records.Task{ID: uuid.New(), TaskName: "pizza-api"}, nil)
	repo.EXPECT().Upsert(gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, r *records.CIReport) error {
			if r.Username != "alice" {
				t.Errorf("username = %s", r.Username)
			}
			if r.TaskName != "pizza-api" {
				t.Errorf("task_name = %s", r.TaskName)
			}
			if r.Status != "failed" {
				t.Errorf("status = %s", r.Status)
			}
			if r.RunID == uuid.Nil {
				t.Error("run_id is nil")
			}
			if len(r.Payload) == 0 {
				t.Fatal("payload is empty")
			}
			return nil
		})

	out, err := svc.CreateReport(context.Background(), sampleInput("alice/golden-pizza-api"))
	if err != nil {
		t.Fatalf("CreateReport() error = %v", err)
	}
	if out.Report.UID != "alice/golden-pizza-api" {
		t.Fatalf("uid = %s", out.Report.UID)
	}
	if out.Report.RunID == "" {
		t.Fatal("run_id is empty in output")
	}
	if out.Report.Summary.RootCause != "APP_UNREACHABLE" {
		t.Fatalf("root_cause = %s", out.Report.Summary.RootCause)
	}
	if out.ID == uuid.Nil {
		t.Fatal("id is nil")
	}
}

func TestCreateReport_TaskNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)

	repo := mocks.NewMockReportRepository(ctrl)
	tasks := mocks.NewMockTaskRepository(ctrl)
	users := mocks.NewMockUsersClient(ctrl)
	svc := reportsservice.NewReportsService(repo, tasks, users)

	tasks.EXPECT().GetByTaskName(gomock.Any(), "pizza-api").Return(nil, pgx.ErrNoRows)
	repo.EXPECT().Upsert(gomock.Any(), gomock.Any()).Times(0)

	_, err := svc.CreateReport(context.Background(), sampleInput("alice/golden-pizza-api"))
	if !errors.Is(err, reportsservice.ErrTaskNotFound) {
		t.Fatalf("error = %v, want ErrTaskNotFound", err)
	}
}

func TestCreateReport_EmptyUID(t *testing.T) {
	ctrl := gomock.NewController(t)
	svc := reportsservice.NewReportsService(
		mocks.NewMockReportRepository(ctrl),
		mocks.NewMockTaskRepository(ctrl),
		mocks.NewMockUsersClient(ctrl),
	)

	_, err := svc.CreateReport(context.Background(), sampleInput(""))
	if !errors.Is(err, reportsservice.ErrUIDEmpty) {
		t.Fatalf("error = %v, want ErrUIDEmpty", err)
	}
}

func TestCreateReport_EmptyCommit(t *testing.T) {
	ctrl := gomock.NewController(t)
	svc := reportsservice.NewReportsService(
		mocks.NewMockReportRepository(ctrl),
		mocks.NewMockTaskRepository(ctrl),
		mocks.NewMockUsersClient(ctrl),
	)

	in := sampleInput("alice/golden-pizza-api")
	in.Commit = ""
	_, err := svc.CreateReport(context.Background(), in)
	if !errors.Is(err, reportsservice.ErrCommitEmpty) {
		t.Fatalf("error = %v, want ErrCommitEmpty", err)
	}
}

func TestCreateReport_EmptyRunID_Generates(t *testing.T) {
	ctrl := gomock.NewController(t)

	repo := mocks.NewMockReportRepository(ctrl)
	tasks := mocks.NewMockTaskRepository(ctrl)
	users := mocks.NewMockUsersClient(ctrl)
	svc := reportsservice.NewReportsService(repo, tasks, users)

	tasks.EXPECT().GetByTaskName(gomock.Any(), "pizza-api").
		Return(&records.Task{ID: uuid.New(), TaskName: "pizza-api"}, nil)
	repo.EXPECT().Upsert(gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, r *records.CIReport) error {
			if r.RunID == uuid.Nil {
				t.Error("run_id is nil; expected generated uuid")
			}
			return nil
		})

	in := sampleInput("alice/golden-pizza-api")
	in.RunID = ""
	out, err := svc.CreateReport(context.Background(), in)
	if err != nil {
		t.Fatalf("error = %v", err)
	}
	if out.Report.RunID == "" {
		t.Fatal("run_id empty in output")
	}
}

func TestCreateReport_InvalidRunID(t *testing.T) {
	ctrl := gomock.NewController(t)
	svc := reportsservice.NewReportsService(
		mocks.NewMockReportRepository(ctrl),
		mocks.NewMockTaskRepository(ctrl),
		mocks.NewMockUsersClient(ctrl),
	)

	in := sampleInput("alice/golden-pizza-api")
	in.RunID = "not-a-uuid"
	_, err := svc.CreateReport(context.Background(), in)
	if !errors.Is(err, reportsservice.ErrRunIDInvalid) {
		t.Fatalf("error = %v, want ErrRunIDInvalid", err)
	}
}

func TestListReports_WithTaskName(t *testing.T) {
	ctrl := gomock.NewController(t)

	repo := mocks.NewMockReportRepository(ctrl)
	tasks := mocks.NewMockTaskRepository(ctrl)
	users := mocks.NewMockUsersClient(ctrl)
	svc := reportsservice.NewReportsService(repo, tasks, users)

	identityID := uuid.New()
	rows := []records.CIReport{
		{ID: uuid.New(), Username: "alice", TaskName: "pizza-api", UID: "alice/golden-pizza-api", Commit: "c1", Status: "passed", CreatedAt: time.Now(), Payload: samplePayload()},
		{ID: uuid.New(), Username: "alice", TaskName: "pizza-api", UID: "alice/golden-pizza-api", Commit: "c2", Status: "failed", CreatedAt: time.Now(), Payload: samplePayload()},
	}

	users.EXPECT().GetGitUser(gomock.Any(), &clientsdto.GitUserRequest{IdentityID: identityID}).
		Return(&clientsdto.GitUserResponse{Username: "alice"}, nil)
	repo.EXPECT().ListByUsernameAndTask(gomock.Any(), "alice", "pizza-api", (*string)(nil), int32(20), (*string)(nil)).
		Return(rows, false, nil)

	out, err := svc.ListReports(context.Background(), &reportsservice.ListInput{
		IdentityID: identityID, TaskName: "pizza-api", Limit: 20,
	})
	if err != nil {
		t.Fatalf("ListReports() error = %v", err)
	}
	if len(out.Items) != 2 {
		t.Fatalf("items = %d, want 2", len(out.Items))
	}
	if out.HasNextPage {
		t.Fatal("has_next_page = true, want false")
	}
}

func TestListReports_AllReports(t *testing.T) {
	ctrl := gomock.NewController(t)

	repo := mocks.NewMockReportRepository(ctrl)
	tasks := mocks.NewMockTaskRepository(ctrl)
	users := mocks.NewMockUsersClient(ctrl)
	svc := reportsservice.NewReportsService(repo, tasks, users)

	identityID := uuid.New()
	rows := []records.CIReport{
		{ID: uuid.New(), Username: "alice", TaskName: "pizza-api", UID: "alice/golden-pizza-api", Commit: "c1", Status: "passed", CreatedAt: time.Now(), Payload: samplePayload()},
	}

	users.EXPECT().GetGitUser(gomock.Any(), &clientsdto.GitUserRequest{IdentityID: identityID}).
		Return(&clientsdto.GitUserResponse{Username: "alice"}, nil)
	repo.EXPECT().ListByUsername(gomock.Any(), "alice", (*string)(nil), int32(20), (*string)(nil)).
		Return(rows, false, nil)

	out, err := svc.ListReports(context.Background(), &reportsservice.ListInput{
		IdentityID: identityID, Limit: 20,
	})
	if err != nil {
		t.Fatalf("ListReports() error = %v", err)
	}
	if len(out.Items) != 1 {
		t.Fatalf("items = %d, want 1", len(out.Items))
	}
}

func TestListReports_UsernameNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)

	repo := mocks.NewMockReportRepository(ctrl)
	tasks := mocks.NewMockTaskRepository(ctrl)
	users := mocks.NewMockUsersClient(ctrl)
	svc := reportsservice.NewReportsService(repo, tasks, users)

	identityID := uuid.New()
	users.EXPECT().GetGitUser(gomock.Any(), gomock.Any()).Return(nil, errors.New("not found"))

	_, err := svc.ListReports(context.Background(), &reportsservice.ListInput{
		IdentityID: identityID, Limit: 20,
	})
	if !errors.Is(err, reportsservice.ErrUsernameNotFound) {
		t.Fatalf("error = %v, want ErrUsernameNotFound", err)
	}
}

func TestListReports_Empty(t *testing.T) {
	ctrl := gomock.NewController(t)

	repo := mocks.NewMockReportRepository(ctrl)
	tasks := mocks.NewMockTaskRepository(ctrl)
	users := mocks.NewMockUsersClient(ctrl)
	svc := reportsservice.NewReportsService(repo, tasks, users)

	identityID := uuid.New()
	users.EXPECT().GetGitUser(gomock.Any(), gomock.Any()).
		Return(&clientsdto.GitUserResponse{Username: "alice"}, nil)
	repo.EXPECT().ListByUsernameAndTask(gomock.Any(), "alice", "pizza-api", (*string)(nil), int32(20), (*string)(nil)).
		Return(nil, false, nil)

	out, err := svc.ListReports(context.Background(), &reportsservice.ListInput{
		IdentityID: identityID, TaskName: "pizza-api", Limit: 20,
	})
	if err != nil {
		t.Fatalf("error = %v", err)
	}
	if len(out.Items) != 0 {
		t.Fatalf("items = %d, want 0", len(out.Items))
	}
}

func TestGetReport_Success(t *testing.T) {
	ctrl := gomock.NewController(t)

	repo := mocks.NewMockReportRepository(ctrl)
	tasks := mocks.NewMockTaskRepository(ctrl)
	users := mocks.NewMockUsersClient(ctrl)
	svc := reportsservice.NewReportsService(repo, tasks, users)

	identityID := uuid.New()
	reportID := uuid.New()
	runID := uuid.New()
	row := &records.CIReport{
		ID: reportID, Username: "alice", TaskName: "pizza-api",
		UID: "alice/golden-pizza-api", Commit: "c1", RunID: runID,
		Status: "passed", CreatedAt: time.Now(), Payload: samplePayload(),
	}

	users.EXPECT().GetGitUser(gomock.Any(), &clientsdto.GitUserRequest{IdentityID: identityID}).
		Return(&clientsdto.GitUserResponse{Username: "alice"}, nil)
	repo.EXPECT().GetByIDAndUsername(gomock.Any(), reportID, "alice").Return(row, nil)

	out, err := svc.GetReport(context.Background(), &reportsservice.GetInput{
		IdentityID: identityID, ReportID: reportID,
	})
	if err != nil {
		t.Fatalf("GetReport() error = %v", err)
	}
	if out.ID != reportID {
		t.Fatalf("id = %s, want %s", out.ID, reportID)
	}
	if out.Report.RunID != runID.String() {
		t.Fatalf("run_id = %s, want %s", out.Report.RunID, runID)
	}
}

func TestGetReport_NotFound(t *testing.T) {
	ctrl := gomock.NewController(t)

	repo := mocks.NewMockReportRepository(ctrl)
	tasks := mocks.NewMockTaskRepository(ctrl)
	users := mocks.NewMockUsersClient(ctrl)
	svc := reportsservice.NewReportsService(repo, tasks, users)

	identityID := uuid.New()
	reportID := uuid.New()

	users.EXPECT().GetGitUser(gomock.Any(), gomock.Any()).
		Return(&clientsdto.GitUserResponse{Username: "alice"}, nil)
	repo.EXPECT().GetByIDAndUsername(gomock.Any(), reportID, "alice").Return(nil, pgx.ErrNoRows)

	_, err := svc.GetReport(context.Background(), &reportsservice.GetInput{
		IdentityID: identityID, ReportID: reportID,
	})
	if !errors.Is(err, reportsservice.ErrReportNotFound) {
		t.Fatalf("error = %v, want ErrReportNotFound", err)
	}
}

func TestListReports_StatusFilter(t *testing.T) {
	ctrl := gomock.NewController(t)

	repo := mocks.NewMockReportRepository(ctrl)
	tasks := mocks.NewMockTaskRepository(ctrl)
	users := mocks.NewMockUsersClient(ctrl)
	svc := reportsservice.NewReportsService(repo, tasks, users)

	identityID := uuid.New()
	pending := domain.ReportStatusPending

	users.EXPECT().GetGitUser(gomock.Any(), gomock.Any()).
		Return(&clientsdto.GitUserResponse{Username: "alice"}, nil)
	repo.EXPECT().ListByUsername(gomock.Any(), "alice", gomock.Any(), int32(20), (*string)(nil)).
		DoAndReturn(func(_ context.Context, _ string, status *string, _ int32, _ *string) ([]records.CIReport, bool, error) {
			if status == nil || *status != "pending" {
				t.Fatalf("status filter = %v, want pending", status)
			}
			return nil, false, nil
		})

	out, err := svc.ListReports(context.Background(), &reportsservice.ListInput{
		IdentityID: identityID, Status: &pending, Limit: 20,
	})
	if err != nil {
		t.Fatalf("error = %v", err)
	}
	if len(out.Items) != 0 {
		t.Fatalf("items = %d, want 0", len(out.Items))
	}
}
