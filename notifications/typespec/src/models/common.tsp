namespace NotificationService;

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

model ErrorResponse {
  code: ErrorCode;
  message: string;
}

model MessageResponse {
  message: string;
}
