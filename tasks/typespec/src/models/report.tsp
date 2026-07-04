namespace Tasks;

// ─── Report Status ────────────────────────────────────────────────────────────

enum ReportStatus {
  pending,
  passed,
  failed,
}

// ─── Report Summary ───────────────────────────────────────────────────────────

model ReportSummary {
  status: "passed" | "failed";
  message: string;
  root_cause?: string;
  passed: integer;
  failed: integer;
  blocked: integer;
  warnings: integer;
}

// ─── Report Step ──────────────────────────────────────────────────────────────

model ReportStep {
  index: integer;
  name: string;
  status: "passed" | "failed" | "blocked" | "warning";
  code?: string;
  failure?: string;
  http_status?: string;
}

// ─── Report Lint Error ────────────────────────────────────────────────────────

model ReportLintError {
  file: string;
  line: integer;
  col: integer;
  rule: string;
  message: string;
}

// ─── Request / Response ───────────────────────────────────────────────────────

model CreateReportRequest {
  uid: string;
  commit: string;
  run_id?: string;
  status: ReportStatus;
  created_at: utcDateTime;
  summary: ReportSummary;
  steps: ReportStep[];
  lint_errors: ReportLintError[];
  warnings: string[];
  raw_log_available: boolean;
}

model Report {
  id: string;
  uid: string;
  commit: string;
  run_id?: string;
  status: ReportStatus;
  created_at: utcDateTime;
  summary: ReportSummary;
  steps: ReportStep[];
  lint_errors: ReportLintError[];
  warnings: string[];
  raw_log_available: boolean;
}

model ReportListResponse {
  items: Report[];
  page_info: PageInfo;
}
