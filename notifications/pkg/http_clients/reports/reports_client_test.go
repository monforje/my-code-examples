package reportsclient

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"notifications/internal/config"
	clientsdto "notifications/pkg/http_clients/dto"
)

func mustTime() time.Time {
	t, err := time.Parse(time.RFC3339, "2026-06-24T12:00:00Z")
	if err != nil {
		panic(err)
	}
	return t
}

func sampleRequest() *clientsdto.CreateReportRequest {
	return &clientsdto.CreateReportRequest{
		UID:       "alice/golden-pizza-api",
		Commit:    "abc123",
		Status:    "failed",
		CreatedAt: mustTime(),
		Summary: clientsdto.ReportSummary{
			Status:    "failed",
			Message:   "Application is not reachable on localhost:8080",
			RootCause: "APP_UNREACHABLE",
			Passed:    4,
			Failed:    3,
			Blocked:   5,
			Warnings:  1,
		},
		Steps:           []clientsdto.ReportStep{},
		LintErrors:      []clientsdto.ReportLintError{},
		Warnings:        []string{},
		RawLogAvailable: true,
	}
}

func TestSendReport_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %s, want POST", r.Method)
		}
		if r.URL.Path != "/reports" {
			t.Errorf("path = %s, want /reports", r.URL.Path)
		}
		if r.Header.Get("Authorization") != "Bearer test-token" {
			t.Errorf("Authorization = %s, want Bearer test-token", r.Header.Get("Authorization"))
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("Content-Type = %s, want application/json", r.Header.Get("Content-Type"))
		}

		var req clientsdto.CreateReportRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		if req.UID != "alice/golden-pizza-api" {
			t.Errorf("uid = %s", req.UID)
		}
		if req.Summary.RootCause != "APP_UNREACHABLE" {
			t.Errorf("root_cause = %s", req.Summary.RootCause)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
	}))
	defer srv.Close()

	client := NewReportsClient(config.ReportsClientConfig{
		Token:   "test-token",
		BaseURL: srv.URL,
	})

	if err := client.SendReport(context.Background(), sampleRequest()); err != nil {
		t.Fatalf("SendReport() error = %v", err)
	}
}

func TestSendReport_NoToken(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if auth := r.Header.Get("Authorization"); auth != "" {
			t.Errorf("Authorization = %s, want empty", auth)
		}
		w.WriteHeader(http.StatusCreated)
	}))
	defer srv.Close()

	client := NewReportsClient(config.ReportsClientConfig{
		Token:   "",
		BaseURL: srv.URL,
	})

	if err := client.SendReport(context.Background(), sampleRequest()); err != nil {
		t.Fatalf("SendReport() error = %v", err)
	}
}

func TestSendReport_BadRequest(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":"bad request"}`))
	}))
	defer srv.Close()

	client := NewReportsClient(config.ReportsClientConfig{
		Token:   "test-token",
		BaseURL: srv.URL,
	})

	err := client.SendReport(context.Background(), sampleRequest())
	if err == nil {
		t.Fatal("SendReport() error = nil, want error")
	}
}

func TestSendReport_InternalError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":"internal"}`))
	}))
	defer srv.Close()

	client := NewReportsClient(config.ReportsClientConfig{
		Token:   "test-token",
		BaseURL: srv.URL,
	})

	err := client.SendReport(context.Background(), sampleRequest())
	if err == nil {
		t.Fatal("SendReport() error = nil, want error")
	}
}

func TestSendReport_ServerDown(t *testing.T) {
	client := NewReportsClient(config.ReportsClientConfig{
		Token:   "test-token",
		BaseURL: "http://localhost:1",
	})

	err := client.SendReport(context.Background(), sampleRequest())
	if err == nil {
		t.Fatal("SendReport() error = nil, want connection error")
	}
}
