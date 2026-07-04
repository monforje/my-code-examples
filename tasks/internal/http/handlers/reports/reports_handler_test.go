package reportshandler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"go.uber.org/mock/gomock"

	"tasks/internal/authctx"
	httpserver "tasks/internal/http/gen"
	reportshandler "tasks/internal/http/handlers/reports"
	"tasks/internal/http/handlers/reports/mocks"
	"tasks/internal/models/domain"
	reportsservice "tasks/internal/services/reports"
)

func reportToVerify() domain.Report {
	return domain.Report{
		UID:    "alice/golden-pizza-api",
		Commit: "abc123",
		Status: domain.ReportStatusFailed,
		Summary: domain.ReportSummary{
			Status: "failed", Message: "app unreachable", RootCause: "APP_UNREACHABLE",
			Passed: 4, Failed: 3, Blocked: 5, Warnings: 1,
		},
		RawLogAvailable: true,
	}
}

func sampleHTTPRequest() httpserver.CreateReportRequest {
	return httpserver.CreateReportRequest{
		Uid:       "alice/golden-pizza-api",
		Commit:    "abc123",
		Status:    httpserver.ReportStatusFailed,
		CreatedAt: time.Date(2026, 6, 24, 12, 0, 0, 0, time.UTC),
		Summary: httpserver.ReportSummary{
			Status: httpserver.Failed, Message: "app unreachable",
			Passed: 4, Failed: 3, Blocked: 5, Warnings: 1,
		},
		Steps:           []httpserver.ReportStep{},
		LintErrors:      []httpserver.ReportLintError{},
		Warnings:        []string{},
		RawLogAvailable: true,
	}
}

func TestReportsCreate_Success(t *testing.T) {
	ctrl := gomock.NewController(t)

	svc := mocks.NewMockReportsService(ctrl)
	h := reportshandler.NewReportsHandler(svc)

	svc.EXPECT().CreateReport(gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, in *reportsservice.CreateReportInput) (*reportsservice.ReportOutput, error) {
			if in.UID != "alice/golden-pizza-api" {
				t.Errorf("uid = %s", in.UID)
			}
			return &reportsservice.ReportOutput{ID: uuid.New(), Report: reportToVerify()}, nil
		})

	body, _ := json.Marshal(sampleHTTPRequest())
	req := httptest.NewRequest(http.MethodPost, "/reports", bytes.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	ctx := echo.New().NewContext(req, rec)

	if err := h.ReportsCreate(ctx); err != nil {
		t.Fatalf("ReportsCreate() error = %v", err)
	}
	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want 201, body = %s", rec.Code, rec.Body.String())
	}

	var resp httpserver.Report
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp.Uid != "alice/golden-pizza-api" {
		t.Fatalf("uid = %s", resp.Uid)
	}
	if resp.Id == "" {
		t.Fatal("id is empty")
	}
}

func TestReportsCreate_OwnerNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	svc := mocks.NewMockReportsService(ctrl)
	h := reportshandler.NewReportsHandler(svc)

	svc.EXPECT().CreateReport(gomock.Any(), gomock.Any()).
		Return(nil, reportsservice.ErrTaskNotFound)

	body, _ := json.Marshal(sampleHTTPRequest())
	req := httptest.NewRequest(http.MethodPost, "/reports", bytes.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()

	if err := h.ReportsCreate(echo.New().NewContext(req, rec)); err != nil {
		t.Fatalf("error = %v", err)
	}
	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", rec.Code)
	}
}

func TestReportsCreate_InvalidJSON(t *testing.T) {
	ctrl := gomock.NewController(t)
	svc := mocks.NewMockReportsService(ctrl)
	h := reportshandler.NewReportsHandler(svc)

	req := httptest.NewRequest(http.MethodPost, "/reports", bytes.NewReader([]byte(`{bad`)))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()

	if err := h.ReportsCreate(echo.New().NewContext(req, rec)); err != nil {
		t.Fatalf("error = %v", err)
	}
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", rec.Code)
	}
}

func TestReportsCreate_InternalError(t *testing.T) {
	ctrl := gomock.NewController(t)
	svc := mocks.NewMockReportsService(ctrl)
	h := reportshandler.NewReportsHandler(svc)

	svc.EXPECT().CreateReport(gomock.Any(), gomock.Any()).Return(nil, errors.New("boom"))

	body, _ := json.Marshal(sampleHTTPRequest())
	req := httptest.NewRequest(http.MethodPost, "/reports", bytes.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()

	if err := h.ReportsCreate(echo.New().NewContext(req, rec)); err != nil {
		t.Fatalf("error = %v", err)
	}
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", rec.Code)
	}
}

func TestReportsList_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	svc := mocks.NewMockReportsService(ctrl)
	h := reportshandler.NewReportsHandler(svc)

	identityID := uuid.New()
	sessionID := uuid.New()

	svc.EXPECT().ListReports(gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, in *reportsservice.ListInput) (*reportsservice.ListOutput, error) {
			if in.IdentityID != identityID {
				t.Errorf("identity_id = %s", in.IdentityID)
			}
			if in.TaskName != "golden-pizza-api" {
				t.Errorf("task_name = %s", in.TaskName)
			}
			return &reportsservice.ListOutput{Items: []reportsservice.ReportOutput{}, HasNextPage: false}, nil
		})

	req := httptest.NewRequest(http.MethodGet, "/reports?task_name=golden-pizza-api&limit=20", nil)
	rec := httptest.NewRecorder()
	ctx := echo.New().NewContext(req, rec)
	ctx.SetRequest(req.WithContext(authctx.WithAuth(req.Context(), identityID, sessionID)))

	taskName := "golden-pizza-api"
	if err := h.ReportsList(ctx, httpserver.ReportsListParams{
		TaskName: &taskName,
		Limit:    20,
	}); err != nil {
		t.Fatalf("ReportsList() error = %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
}

func TestReportsList_Unauthorized(t *testing.T) {
	ctrl := gomock.NewController(t)
	svc := mocks.NewMockReportsService(ctrl)
	h := reportshandler.NewReportsHandler(svc)

	svc.EXPECT().ListReports(gomock.Any(), gomock.Any()).Times(0)

	req := httptest.NewRequest(http.MethodGet, "/reports?task_name=golden-pizza-api", nil)
	rec := httptest.NewRecorder()

	taskName := "golden-pizza-api"
	if err := h.ReportsList(echo.New().NewContext(req, rec), httpserver.ReportsListParams{
		TaskName: &taskName, Limit: 20,
	}); err != nil {
		t.Fatalf("error = %v", err)
	}
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", rec.Code)
	}
}
