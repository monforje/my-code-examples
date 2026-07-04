namespace Users;

// ─── Error Codes ──────────────────────────────────────────────────────────────

enum ErrorCode {
  VALIDATION_ERROR,
  INVALID_JSON,
  MISSING_AUTH_TOKEN,
  INVALID_AUTH_TOKEN,
  EXPIRED_AUTH_TOKEN,
  NOT_FOUND,
  INTERNAL_ERROR,
  AVATAR_TOO_LARGE,
  INVALID_AVATAR_FORMAT,
}

// ─── Error Models ─────────────────────────────────────────────────────────────

model ErrorResponse {
  code: ErrorCode;
  message: string;
}
