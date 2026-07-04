namespace Auth;

model DeviceStartResponse {
  device_code: string;
  user_code: string;
  verification_url: string;
  expires_in: int32;
  interval: int32;
}

model DeviceConfirmRequest {
  user_code: string;
}

model DeviceConfirmResponse {
  status: "confirmed";
}

model DeviceTokenRequest {
  device_code: string;
}

model CliTokenResponse {
  access_token: string;
  refresh_token: string;
  expires_in: int32;
  token_type: "Bearer";
}

model CliRefreshRequest {
  refresh_token: string;
}

model CliRefreshResponse {
  access_token: string;
  refresh_token: string;
  expires_in: int32;
  token_type: "Bearer";
}
