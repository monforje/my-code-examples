package reportservice

import (
	"testing"
	"time"

	"notifications/internal/models/domain"
)

func mustTime(t *testing.T) time.Time {
	t.Helper()
	tt, err := time.Parse(time.RFC3339, "2026-06-24T12:00:00Z")
	if err != nil {
		t.Fatal(err)
	}
	return tt
}

func TestFromCIReport_FallbackPassed(t *testing.T) {
	ts := mustTime(t)
	// Не-JSON stdout (например, прямой woodpecker-вывод) → fallback без шагов.
	r := &domain.CIReport{
		UID:       "alice/golden-pizza-api",
		Commit:    "abc123",
		RunID:     "run-1",
		Status:    domain.StatusPassed,
		ExitCode:  0,
		Stage:     "run",
		Stdout:    "✓ build\n✓ test\n",
		CreatedAt: ts,
	}

	got := FromCIReport(r)
	if got.Status != domain.ReportStatusPassed {
		t.Fatalf("status = %q, want passed", got.Status)
	}
	if got.RunID != "run-1" {
		t.Fatalf("run_id = %q, want run-1 (propagated)", got.RunID)
	}
	// Fallback не парсит шаги (это делает ci-translator JSON-путь).
	if len(got.Steps) != 0 {
		t.Fatalf("steps len = %d, want 0 (fallback)", len(got.Steps))
	}
	if got.Summary.Message != "CI passed" {
		t.Fatalf("message = %q", got.Summary.Message)
	}
}

func TestFromCIReport_FallbackFailed(t *testing.T) {
	r := &domain.CIReport{
		UID:      "bob/golden-pizza-api",
		Commit:   "def456",
		RunID:    "run-2",
		Status:   domain.StatusFailed,
		ExitCode: 1,
		Stage:    "run",
	}

	got := FromCIReport(r)
	if got.Status != domain.ReportStatusFailed {
		t.Fatalf("status = %q, want failed", got.Status)
	}
	if got.RunID != "run-2" {
		t.Fatalf("run_id = %q, want run-2", got.RunID)
	}
	if len(got.Steps) != 0 {
		t.Fatalf("steps len = %d, want 0 (fallback)", len(got.Steps))
	}
}

func TestFromCIReport_Crashed(t *testing.T) {
	r := &domain.CIReport{
		UID:      "c/golden-pizza-api",
		Commit:   "x",
		Status:   domain.StatusCrashed,
		ExitCode: -1,
		Stage:    "vm_crash",
		Stderr:   "VM connection lost",
	}

	got := FromCIReport(r)
	if got.Status != domain.ReportStatusFailed {
		t.Fatalf("status = %q, want failed (crashed→failed)", got.Status)
	}
	if got.Summary.RootCause != "VM_CRASHED" {
		t.Fatalf("root_cause = %q, want VM_CRASHED", got.Summary.RootCause)
	}
	if got.Summary.Message != "CI crashed: VM connection lost" {
		t.Fatalf("message = %q", got.Summary.Message)
	}
	if len(got.Warnings) != 1 || got.Warnings[0] != "VM connection lost" {
		t.Fatalf("warnings = %v", got.Warnings)
	}
	if got.Summary.Warnings != 1 {
		t.Fatalf("summary.warnings = %d, want 1", got.Summary.Warnings)
	}
}

func TestFromCIReport_Timeout(t *testing.T) {
	r := &domain.CIReport{
		UID:    "c/golden-pizza-api",
		Commit: "x",
		Status: domain.StatusTimeout,
		Stage:  "timeout",
	}

	got := FromCIReport(r)
	if got.Status != domain.ReportStatusFailed {
		t.Fatalf("status = %q, want failed", got.Status)
	}
	if got.Summary.RootCause != "TIMEOUT" {
		t.Fatalf("root_cause = %q, want TIMEOUT", got.Summary.RootCause)
	}
	if got.Summary.Message != "CI timed out" {
		t.Fatalf("message = %q", got.Summary.Message)
	}
}

func TestFromCIReport_Pending(t *testing.T) {
	r := &domain.CIReport{
		UID:    "a/golden-pizza-api",
		Commit: "x",
		Status: domain.StatusPending,
		Stage:  "started",
	}

	got := FromCIReport(r)
	if got.Status != domain.ReportStatusPending {
		t.Fatalf("status = %q, want pending", got.Status)
	}
	if got.RawLogAvailable {
		t.Fatal("raw_log_available = true, want false (no stdout)")
	}
}

func TestFromCIReport_Nil(t *testing.T) {
	if got := FromCIReport(nil); got != nil {
		t.Fatalf("got = %+v, want nil", got)
	}
}

func TestToDTO_JSONTags(t *testing.T) {
	ts := mustTime(t)
	r := &domain.Report{
		UID:       "alice/golden-pizza-api",
		Commit:    "abc123",
		Status:    domain.ReportStatusFailed,
		CreatedAt: ts,
		Summary: domain.ReportSummary{
			Status:    "failed",
			Message:   "Application is not reachable on localhost:8080",
			RootCause: "APP_UNREACHABLE",
			Passed:    4,
			Failed:    3,
			Blocked:   5,
			Warnings:  1,
		},
		Steps: []domain.ReportStep{
			{Index: 1, Name: "Register", Status: domain.ReportStepFailed, Code: "CONNECTION_REFUSED", HTTPStatus: "000"},
		},
		Warnings:        []string{},
		RawLogAvailable: true,
	}

	dto := toDTO(r)
	if dto == nil {
		t.Fatal("dto = nil")
	}
	if dto.UID != r.UID || dto.Commit != r.Commit || dto.Status != "failed" {
		t.Fatalf("dto top-level = %+v", dto)
	}
	if dto.Summary.RootCause != "APP_UNREACHABLE" {
		t.Fatalf("root_cause = %s", dto.Summary.RootCause)
	}
	if dto.Steps[0].HTTPStatus != "000" {
		t.Fatalf("http_status = %s", dto.Steps[0].HTTPStatus)
	}
	if len(dto.LintErrors) != 0 {
		t.Fatalf("lint_errors len = %d, want 0", len(dto.LintErrors))
	}
	if dto.Warnings == nil {
		t.Fatal("warnings = nil, want non-nil empty slice")
	}
}

func TestFromCIReport_JSONStdout(t *testing.T) {
	ts := mustTime(t)
	stdout := `{
  "passed": false,
  "summary": {
    "status": "failed",
    "message": "Application is not reachable on localhost:8080",
    "root_cause": "APP_UNREACHABLE",
    "passed": 3,
    "failed": 5,
    "blocked": 4,
    "warnings": 0
  },
  "steps": [
    {"index": 1, "name": "Register", "passed": false, "status": "failed", "code": "CONNECTION_REFUSED", "failure": "curl could not connect to localhost:8080", "http_status": "000"},
    {"index": 2, "name": "Login", "passed": false, "status": "blocked", "failure": "app is unreachable"},
    {"index": 3, "name": "Benchmarks", "passed": true, "status": "passed"}
  ],
  "warnings": ["RPS: load test failed"],
  "lint_errors": [{"file": "main.go", "line": 10, "col": 5, "rule": "errcheck", "message": "unchecked error"}],
  "raw_log_available": true
}`

	r := &domain.CIReport{
		UID:       "alice/golden-pizza-api",
		Commit:    "abc123",
		Status:    domain.StatusFailed,
		ExitCode:  1,
		Stage:     "run",
		Stdout:    stdout,
		CreatedAt: ts,
		Steps:     nil,
	}

	got := FromCIReport(r)
	if got == nil {
		t.Fatal("FromCIReport returned nil")
	}
	if got.UID != "alice/golden-pizza-api" {
		t.Fatalf("uid = %q", got.UID)
	}
	if got.Commit != "abc123" {
		t.Fatalf("commit = %q", got.Commit)
	}
	if got.Status != domain.ReportStatusFailed {
		t.Fatalf("status = %q, want failed", got.Status)
	}
	if got.Summary.Message != "Application is not reachable on localhost:8080" {
		t.Fatalf("message = %q", got.Summary.Message)
	}
	if got.Summary.RootCause != "APP_UNREACHABLE" {
		t.Fatalf("root_cause = %q", got.Summary.RootCause)
	}
	if got.Summary.Passed != 3 || got.Summary.Failed != 5 || got.Summary.Blocked != 4 {
		t.Fatalf("counts = passed=%d failed=%d blocked=%d", got.Summary.Passed, got.Summary.Failed, got.Summary.Blocked)
	}
	if len(got.Steps) != 3 {
		t.Fatalf("steps len = %d, want 3", len(got.Steps))
	}
	if got.Steps[0].Name != "Register" || got.Steps[0].Status != domain.ReportStepFailed {
		t.Fatalf("step[0] = %+v", got.Steps[0])
	}
	if got.Steps[0].Code != "CONNECTION_REFUSED" {
		t.Fatalf("step[0].code = %q", got.Steps[0].Code)
	}
	if got.Steps[0].HTTPStatus != "000" {
		t.Fatalf("step[0].http_status = %q", got.Steps[0].HTTPStatus)
	}
	if got.Steps[1].Status != domain.ReportStepBlocked {
		t.Fatalf("step[1].status = %q, want blocked", got.Steps[1].Status)
	}
	if got.Steps[2].Status != domain.ReportStepPassed {
		t.Fatalf("step[2].status = %q, want passed", got.Steps[2].Status)
	}
	if len(got.LintErrors) != 1 {
		t.Fatalf("lint_errors len = %d, want 1", len(got.LintErrors))
	}
	if got.LintErrors[0].Rule != "errcheck" {
		t.Fatalf("lint_error.rule = %q", got.LintErrors[0].Rule)
	}
	if len(got.Warnings) != 1 {
		t.Fatalf("warnings len = %d, want 1", len(got.Warnings))
	}
	if !got.RawLogAvailable {
		t.Fatal("raw_log_available = false, want true")
	}
}

func TestFromCIReport_JSONStdout_Passed(t *testing.T) {
	stdout := `{
  "passed": true,
  "summary": {"status": "passed", "message": "CI passed", "passed": 5, "failed": 0, "blocked": 0, "warnings": 0},
  "steps": [{"index": 1, "name": "build", "status": "passed"}],
  "warnings": [],
  "lint_errors": [],
  "raw_log_available": false
}`

	r := &domain.CIReport{
		UID:    "bob/golden-api",
		Commit: "def",
		Status: domain.StatusPassed,
		Stdout: stdout,
	}

	got := FromCIReport(r)
	if got.Status != domain.ReportStatusPassed {
		t.Fatalf("status = %q, want passed", got.Status)
	}
	if got.Summary.Passed != 5 {
		t.Fatalf("passed = %d, want 5", got.Summary.Passed)
	}
	if len(got.Steps) != 1 || got.Steps[0].Name != "build" {
		t.Fatalf("steps = %+v", got.Steps)
	}
}

func TestParseCITranslatorJSON_Invalid(t *testing.T) {
	tests := []struct {
		name   string
		stdout string
	}{
		{"empty", ""},
		{"plain text", "some log output"},
		{"invalid json", "{not json}"},
		{"missing summary", `{"steps": []}`},
		{"array", `[1, 2, 3]`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := parseCITranslatorJSON(tt.stdout)
			if ok {
				t.Fatalf("expected false, got true: %+v", got)
			}
		})
	}
}
