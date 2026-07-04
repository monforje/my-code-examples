import { createContext, useCallback, useContext, useEffect, useState, type ReactNode } from "react";
import type { User } from "@entities/user";
import {
  authMeGet,
  authLogin,
  authRegister,
  authLogout,
  ErrorCode,
  refreshAccessToken,
} from "@shared/api";
import { setAccessToken, clearAccessToken } from "@shared/api/token";
import { getApiErrorMessage } from "@shared/lib/api-error";

export class EmailNotVerifiedError extends Error {
  email: string;
  constructor(email: string) {
    super(ErrorCode.EMAIL_NOT_VERIFIED);
    this.name = "EmailNotVerifiedError";
    this.email = email;
  }
}

interface AuthContextValue {
  user: User | null;
  isLoading: boolean;
  isAuthenticated: boolean;
  login: (email: string, password: string) => Promise<void>;
  register: (email: string, password: string) => Promise<void>;
  logout: () => Promise<void>;
  setUser: (user: User | null) => void;
}

const AuthContext = createContext<AuthContextValue | null>(null);

interface AuthProviderProps {
  children: ReactNode;
}

export function AuthProvider({ children }: AuthProviderProps) {
  const [user, setUser] = useState<User | null>(null);
  const [isLoading, setIsLoading] = useState(true);

  useEffect(() => {
    let ignore = false;

    (async () => {
      try {
        let res = await authMeGet();

        if ((res as { status: number }).status === 401) {
          const refreshed = await refreshAccessToken();
          if (refreshed) {
            res = await authMeGet();
          }
        }

        if (!ignore && (res as { status: number }).status === 200) {
          setUser((res as { data: User }).data);
        }
      } catch {
        // not authenticated
      } finally {
        if (!ignore) setIsLoading(false);
      }
    })();

    return () => {
      ignore = true;
    };
  }, []);

  const login = useCallback(async (email: string, password: string) => {
    const res = await authLogin({ email, password });
    const status = (res as { status: number }).status;

    if (status === 200) {
      const data = (res as { data: { access_token: string } }).data;
      setAccessToken(data.access_token);

      const meRes = await authMeGet();
      if ((meRes as { status: number }).status === 200) {
        setUser((meRes as { data: User }).data);
      }
      return;
    }

    const errorData = (res as { data?: { code?: string; message?: string } }).data;

    if (status === 403 && errorData?.code === ErrorCode.EMAIL_NOT_VERIFIED) {
      throw new EmailNotVerifiedError(email);
    }

    if (status === 401) {
      throw new Error("Email или пароль указаны неверно.");
    }

    if (status === 429) {
      throw new Error("Слишком много попыток. Попробуйте позже.");
    }

    if (status === 500) {
      throw new Error("Не удалось войти. Попробуйте ещё раз.");
    }

    throw new Error(getApiErrorMessage(res, "Не удалось войти."));
  }, []);

  const register = useCallback(async (email: string, password: string) => {
    const res = await authRegister({ email, password });
    const status = (res as { status: number }).status;

    if (status === 201) return;

    if (status === 409) {
      throw new Error("Аккаунт с таким email уже существует.");
    }

    if (status === 422) {
      throw new Error("Проверьте email и пароль.");
    }

    if (status === 429) {
      throw new Error("Слишком много регистраций с этого устройства. Попробуйте позже.");
    }

    if (status === 500) {
      throw new Error("Не удалось создать аккаунт.");
    }

    throw new Error(getApiErrorMessage(res, "Не удалось создать аккаунт."));
  }, []);

  const logout = useCallback(async () => {
    try {
      await authLogout();
    } finally {
      clearAccessToken();
      setUser(null);
    }
  }, []);

  return (
    <AuthContext.Provider
      value={{
        user,
        isLoading,
        isAuthenticated: !!user,
        login,
        register,
        logout,
        setUser,
      }}
    >
      {children}
    </AuthContext.Provider>
  );
}

export function useAuth() {
  const ctx = useContext(AuthContext);
  if (!ctx) throw new Error("useAuth must be used within AuthProvider");
  return ctx;
}
