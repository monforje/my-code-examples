namespace Users;

// --- Git User Info ---
model GitMeRequest {
  identity_id: string;
}

model GitMeResponse {
  username: string;
  git_token: string;
  git_url: string;
}
