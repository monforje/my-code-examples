import { getAccessToken, setAccessToken, clearAccessToken } from "./token";

const PUBLIC_AUTH_PATHS = [
  "/api/v1/auth/login",
  "/api/v1/auth/register",
  "/api/v1/auth/register/verify",
  "/api/v1/auth/register/code/resend",
  "/api/v1/auth/password/forgot",
  "/api/v1/auth/password/forgot/verify",
  "/api/v1/auth/password/forgot/code/resend",
  "/api/v1/auth/password/reset",
  "/api/v1/auth/refresh",
];

function isPublicAuthPath(url: string): boolean {
  return PUBLIC_AUTH_PATHS.some((p) => url.startsWith(p));
}

let refreshInProgress: Promise<boolean> | null = null;

export async function refreshAccessToken(): Promise<boolean> {
  if (refreshInProgress) return refreshInProgress;

  refreshInProgress = (async () => {
    try {
      const res = await fetch("/api/v1/auth/refresh", {
        method: "POST",
        credentials: "include",
      });
      if (res.ok) {
        const data = await res.json();
        setAccessToken(data.access_token);
        return true;
      }
      clearAccessToken();
      return false;
    } catch {
      clearAccessToken();
      return false;
    } finally {
      refreshInProgress = null;
    }
  })();

  return refreshInProgress;
}

export async function customFetch<T>(url: string, options: RequestInit = {}): Promise<T> {
  const token = getAccessToken();
  const headers: Record<string, string> = {
    ...(options.headers as Record<string, string>),
  };

  if (token && !headers["Authorization"]) {
    headers["Authorization"] = `Bearer ${token}`;
  }

  const doFetch = async () => {
    const res = await fetch(url, {
      ...options,
      headers,
      credentials: "include",
    });

    if (res.status === 204) return { data: undefined, status: 204 } as T;

    const text = await res.text();
    let body: unknown;
    if (text) {
      try {
        body = JSON.parse(text);
      } catch {
        body = { code: "INTERNAL_ERROR", message: "Ошибка сервера." };
      }
    }
    return { data: body, status: res.status } as T;
  };

  let result: T;
  try {
    result = await doFetch();
  } catch {
    return { data: { code: "INTERNAL_ERROR", message: "Ошибка сервера." }, status: 0 } as T;
  }

  if ((result as { status: number }).status === 401 && !isPublicAuthPath(url) && token) {
    const refreshed = await refreshAccessToken();
    if (refreshed) {
      const newToken = getAccessToken();
      headers["Authorization"] = `Bearer ${newToken}`;
      try {
        result = await doFetch();
      } catch {
        return { data: { code: "INTERNAL_ERROR", message: "Ошибка сервера." }, status: 0 } as T;
      }
    }
  }

  return result;
}
