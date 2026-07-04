namespace NotificationService;

// ─── Webhook (входящий от task_runner) ───────────────────────────────────────

model WebhookRequest {
  event: WebhookEvent;
  uid: string;
  commit: string;
  vm_id?: string;
  exit_code?: int32;
  stdout?: string;
  stderr?: string;
  stage?: string;
}

enum WebhookEvent {
  CI_STARTED: "ci_started",
  CI_FINISHED: "ci_finished",
}

// ─── CI Report ───────────────────────────────────────────────────────────────

model CIReportResponse {
  uid: string;
  commit: string;
  status: CIReportStatus;
  exit_code: int32;
  stage: string;
  steps: CIReportStep[];
  stdout: string;
  stderr: string;
  created_at: utcDateTime;
}

enum CIReportStatus {
  PENDING: "pending",
  PASSED: "passed",
  FAILED: "failed",
  CRASHED: "crashed",
  TIMEOUT: "timeout",
}

model CIReportStep {
  name: string;
  status: CIStepStatus;
  duration_ms: int32;
}

enum CIStepStatus {
  SUCCESS: "success",
  FAILURE: "failure",
  SKIPPED: "skipped",
}
