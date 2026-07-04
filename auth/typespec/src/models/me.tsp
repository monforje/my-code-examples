namespace Auth;

// Identity — учётная запись в auth-сервисе (без profile-данных)
model Identity {
  id: string;
  email: string;
  email_verified: boolean;
  status: "pending_verification" | "active" | "blocked" | "deleted";
  created_at: utcDateTime;
}

model DeleteAccountRequest {
  password: string;
}

model EmailChangeRequest {
  password: string;
}

model IdentityTokenResponse {
  identity_token: string;
  expires_in: int32;
}

model ConfirmEmailChangeRequest {
  new_email: string;
  identity_token: string;
}

model CompleteEmailChangeRequest {
  code: string;
}

model ResendEmailChangeCodeRequest {
  step: "current" | "new";
}
