namespace Users;

// ─── Profile ──────────────────────────────────────────────────────────────────

model ProfileResponse {
  id: string;
  identity_id: string;
  email: string;

  display_name?: string;
  bio?: string;
  avatar_url?: string;

  status: string;
  email_verified: boolean;

  created_at: utcDateTime;
  updated_at: utcDateTime;
}

// ─── Update Profile Settings ──────────────────────────────────────────────────

model UpdateProfileSettingsRequest {
  display_name?: string;
  bio?: string;
}

// ─── Avatar ───────────────────────────────────────────────────────────────────

model UpdateAvatarResponse {
  avatar_url: string;
  updated_at: utcDateTime;
}

model DeleteAvatarResponse {
  avatar_url?: string;
  updated_at: utcDateTime;
}
