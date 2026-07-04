import type { ProfileResponse } from "@shared/api";

export interface User {
  id: string;
  email: string;
  email_verified: boolean;
  status: "pending_verification" | "active" | "blocked" | "deleted";
  created_at: string;
}

export type Profile = ProfileResponse;
