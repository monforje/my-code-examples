package e2e_test

import (
	"net/http"
	"testing"
)

// TestE2E_ReportFlow_PassedSentToTasks - finished (passed) отчёт уходит в tasks с верным контрактом.
func TestE2E_ReportFlow_PassedSentToTasks(t *testing.T) {
	resetE2E(t)
	uid := "alice/golden-pizza-api"

	stdout := `{
  "run_id": "run-passed-1",
  "summary": {"status": "passed", "message": "Pipeline passed", "passed": 2, "failed": 0, "blocked": 0, "warnings": 0},
  "steps": [
    {"index": 1, "name": "build", "passed": true, "status": "passed"},
    {"index": 2, "name": "test", "passed": true, "status": "passed"}
  ],
  "warnings": [],
  "lint_errors": [],
  "raw_log_available": true
}`
	status, _ := postWebhook(t, map[string]any{
		"event":     "ci_finished",
		"uid":       uid,
		"commit":    "abc123",
		"exit_code": 0,
		"stdout":    stdout,
		"stage":     "run",
	})
	if status != http.StatusOK {
		t.Fatalf("ci_finished status = %d", status)
	}

	waitTasksReport(t, 1)

	report := tasksRecorder.last()
	if report["uid"] != uid {
		t.Fatalf("uid = %v, want %s", report["uid"], uid)
	}
	if report["commit"] != "abc123" {
		t.Fatalf("commit = %v", report["commit"])
	}
	if report["status"] != "passed" {
		t.Fatalf("status = %v, want passed", report["status"])
	}
	if report["run_id"] != "run-passed-1" {
		t.Fatalf("run_id = %v, want run-passed-1", report["run_id"])
	}

	summary, ok := report["summary"].(map[string]any)
	if !ok {
		t.Fatalf("summary = %v, want object", report["summary"])
	}
	if summary["status"] != "passed" {
		t.Fatalf("summary.status = %v, want passed", summary["status"])
	}
	if summary["passed"].(float64) != 2 {
		t.Fatalf("summary.passed = %v, want 2", summary["passed"])
	}
	if summary["failed"].(float64) != 0 {
		t.Fatalf("summary.failed = %v, want 0", summary["failed"])
	}

	steps, ok := report["steps"].([]any)
	if !ok || len(steps) != 2 {
		t.Fatalf("steps = %v, want 2 items", report["steps"])
	}
	if report["raw_log_available"] != true {
		t.Fatalf("raw_log_available = %v, want true", report["raw_log_available"])
	}
}

// TestE2E_ReportFlow_FailedSentToTasks - failed отчёт корректно сериализуется.
func TestE2E_ReportFlow_FailedSentToTasks(t *testing.T) {
	resetE2E(t)
	uid := "bob/golden-pizza-api"

	stdout := `{
  "run_id": "run-failed-1",
  "summary": {"status": "failed", "message": "Pipeline failed", "root_cause": "UNKNOWN_FAILURE", "passed": 1, "failed": 1, "blocked": 0, "warnings": 0},
  "steps": [
    {"index": 1, "name": "build", "status": "passed"},
    {"index": 2, "name": "test", "status": "failed", "code": "COMMAND_EXIT_NON_ZERO"}
  ],
  "warnings": [],
  "lint_errors": [],
  "raw_log_available": true
}`
	status, _ := postWebhook(t, map[string]any{
		"event":     "ci_finished",
		"uid":       uid,
		"commit":    "def456",
		"exit_code": 1,
		"stdout":    stdout,
		"stage":     "run",
	})
	if status != http.StatusOK {
		t.Fatalf("ci_finished status = %d", status)
	}

	waitTasksReport(t, 1)
	report := tasksRecorder.last()

	if report["status"] != "failed" {
		t.Fatalf("status = %v, want failed", report["status"])
	}
	if report["run_id"] != "run-failed-1" {
		t.Fatalf("run_id = %v, want run-failed-1", report["run_id"])
	}
	summary := report["summary"].(map[string]any)
	if summary["status"] != "failed" {
		t.Fatalf("summary.status = %v, want failed", summary["status"])
	}
	if summary["passed"].(float64) != 1 {
		t.Fatalf("summary.passed = %v, want 1", summary["passed"])
	}
	if summary["failed"].(float64) != 1 {
		t.Fatalf("summary.failed = %v, want 1", summary["failed"])
	}
	if summary["root_cause"] != "UNKNOWN_FAILURE" {
		t.Fatalf("summary.root_cause = %v, want UNKNOWN_FAILURE", summary["root_cause"])
	}
}

// TestE2E_ReportFlow_CrashedSentToTasks - crashed отчёт: root_cause + warning из stderr.
func TestE2E_ReportFlow_CrashedSentToTasks(t *testing.T) {
	resetE2E(t)
	uid := "cara/golden-pizza-api"

	status, _ := postWebhook(t, map[string]any{
		"event":     "ci_finished",
		"uid":       uid,
		"commit":    "ghi789",
		"exit_code": -1,
		"stderr":    "VM connection lost",
		"stage":     "vm_crash",
	})
	if status != http.StatusOK {
		t.Fatalf("ci_finished status = %d", status)
	}

	waitTasksReport(t, 1)
	report := tasksRecorder.last()

	if report["status"] != "failed" {
		t.Fatalf("status = %v, want failed", report["status"])
	}
	summary := report["summary"].(map[string]any)
	if summary["root_cause"] != "VM_CRASHED" {
		t.Fatalf("root_cause = %v, want VM_CRASHED", summary["root_cause"])
	}
	if summary["message"] != "CI crashed: VM connection lost" {
		t.Fatalf("summary.message = %v", summary["message"])
	}
	warnings, ok := report["warnings"].([]any)
	if !ok || len(warnings) != 1 || warnings[0] != "VM connection lost" {
		t.Fatalf("warnings = %v", report["warnings"])
	}
}

// TestE2E_ReportFlow_StartedSentToTasks - pending (ci_started) отчёт уходит в tasks как pending.
func TestE2E_ReportFlow_StartedSentToTasks(t *testing.T) {
	resetE2E(t)
	uid := "dave/golden-pizza-api"

	status, _ := postWebhook(t, map[string]any{
		"event":  "ci_started",
		"uid":    uid,
		"commit": "aaa111",
	})
	if status != http.StatusOK {
		t.Fatalf("ci_started status = %d", status)
	}

	waitTasksReport(t, 1)
	report := tasksRecorder.last()

	if report["status"] != "pending" {
		t.Fatalf("status = %v, want pending", report["status"])
	}
	if report["uid"] != uid {
		t.Fatalf("uid = %v, want %s", report["uid"], uid)
	}
	summary := report["summary"].(map[string]any)
	if summary["message"] != "CI pending" {
		t.Fatalf("summary.message = %v, want 'CI pending'", summary["message"])
	}
}
