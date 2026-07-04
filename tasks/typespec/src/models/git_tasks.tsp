namespace Tasks;

// ─── Request Models ───────────────────────────────────────────────────────────

model GitTaskRequest {
  task_name: string;
}

// ─── Response Models ──────────────────────────────────────────────────────────

model GitTaskResponse {
  task_name: string;
  repo: string;
  clone_url: string;
}
