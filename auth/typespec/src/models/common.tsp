namespace Auth;

// ─── Error Codes ──────────────────────────────────────────────────────────────

enum ErrorCode {
  VALIDATION_ERROR,
  INVALID_JSON,
  MISSING_AUTH_TOKEN,
  INVALID_AUTH_TOKEN,
  EXPIRED_AUTH_TOKEN,
  INVALID_REFRESH_TOKEN,
  EXPIRED_REFRESH_TOKEN,
  INVALID_CREDENTIALS,
  EMAIL_NOT_VERIFIED,
  EMAIL_ALREADY_EXISTS,
  INVALID_CODE,
  EXPIRED_CODE,
  TOO_MANY_ATTEMPTS,
  RATE_LIMITED,
  CURRENT_PASSWORD_INCORRECT,
  RESET_TOKEN_INVALID,
  RESET_TOKEN_EXPIRED,
  NOT_FOUND,
  INTERNAL_ERROR,
}

// ─── Error Models ─────────────────────────────────────────────────────────────

model ErrorResponse {
  code: ErrorCode;
  message: string;
}

// ─── Common Models ────────────────────────────────────────────────────────────

model MessageResponse {
  message: string;
}

model ResendCodeRequest {
  email: string;
}
