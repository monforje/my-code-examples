namespace Tasks;

// ─── Error Codes ──────────────────────────────────────────────────────────────

enum ErrorCode {
  VALIDATION_ERROR,
  INVALID_JSON,
  MISSING_AUTH_TOKEN,
  INVALID_AUTH_TOKEN,
  EXPIRED_AUTH_TOKEN,
  NOT_FOUND,
  INTERNAL_ERROR,
}

// ─── Enums ────────────────────────────────────────────────────────────────────

enum TaskType {
  backend,
  frontend,
}

enum Level {
  junior,
  middle,
  senior,
}

// ─── Error Models ─────────────────────────────────────────────────────────────

model ErrorResponse {
  code: ErrorCode;
  message: string;
  details?: Record<string>;
}

// ─── Common Models ────────────────────────────────────────────────────────────

model PageInfo {
  has_next_page: boolean;
  next_cursor?: string;
}

// ─── Reference Models ─────────────────────────────────────────────────────────

model Tag {
  id: string;
  name: string;
}

model Language {
  id: string;
  name: string;
}
