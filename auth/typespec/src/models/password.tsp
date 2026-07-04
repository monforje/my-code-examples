namespace Auth;

model ForgotPasswordRequest {
  email: string;
}

model ForgotVerifyRequest {
  email: string;
  code: string;
}

model ResetTokenResponse {
  reset_token: string;
  expires_in: int32;
}

model ResetPasswordRequest {
  reset_token: string;
  new_password: string;
}

model ChangePasswordRequest {
  current_password: string;
}

model ChangePasswordTokenResponse {
  change_token: string;
  expires_in: int32;
}

model CompletePasswordChangeRequest {
  change_token: string;
  new_password: string;
}

model CodeVerifyRequest {
  code: string;
}
