namespace Auth;

model RegisterRequest {
  email: string;
  password: string;
}

model RegisterResponse {
  identity_id: string;
  email: string;
  status: "pending_verification";
}

model VerifyCodeRequest {
  email: string;
  code: string;
}

model TokenResponse {
  access_token: string;
  expires_in: int32;
}

model LoginRequest {
  email: string;
  password: string;
}

model RefreshResponse {
  access_token: string;
  expires_in: int32;
}
