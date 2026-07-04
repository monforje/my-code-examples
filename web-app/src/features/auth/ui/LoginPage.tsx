import { useState, type FormEvent } from "react";
import { useNavigate } from "react-router";
import { useAuth } from "@features/auth";
import { AuthLayout, AuthMain } from "./AuthLayout";
import { AuthCard } from "./AuthCard";
import { AuthForm, AuthFormActions } from "./AuthForm";
import { AuthFooter } from "./AuthFooter";
import { Field, Input, PasswordInput, Button, Alert } from "@shared/ui";

export function LoginPage() {
  const navigate = useNavigate();
  const { login } = useAuth();
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(false);

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault();
    setError("");
    const form = e.target as HTMLFormElement;
    const email = (form.elements.namedItem("email") as HTMLInputElement).value;
    const password = (form.elements.namedItem("password") as HTMLInputElement).value;

    if (!email || !password) return;

    setLoading(true);
    try {
      await login(email, password);
      navigate("/profile");
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : "Ошибка входа";
      setError(msg);
    } finally {
      setLoading(false);
    }
  };

  return (
    <AuthLayout>
      <PublicHeader navigate={navigate} />
      <AuthMain>
        <AuthCard title="Вход" titleLg>
          <p style={{ fontSize: "var(--cd-text-sm)", color: "var(--cd-text-muted)" }}>
            Используйте рабочий email, чтобы продолжить работу с платформой.
          </p>
          {error && <Alert variant="danger">{error}</Alert>}
          <AuthForm onSubmit={handleSubmit}>
            <Field label="Email">
              <Input
                name="email"
                type="email"
                placeholder="name@company.com"
                required
                autoComplete="email"
              />
            </Field>
            <Field label="Пароль">
              <PasswordInput
                name="password"
                placeholder="Введите пароль"
                required
                autoComplete="current-password"
              />
            </Field>
            <AuthFormActions>
              <Button type="submit" variant="primary" fullWidth loading={loading}>
                Войти
              </Button>
            </AuthFormActions>
            <div style={{ display: "flex", justifyContent: "flex-end" }}>
              <button type="button" className={styles.link} onClick={() => navigate("/forgot")}>
                Забыли пароль?
              </button>
            </div>
          </AuthForm>
        </AuthCard>
      </AuthMain>
      <AuthFooter />
    </AuthLayout>
  );
}

import styles from "./AuthForm.module.css";

function PublicHeader({ navigate }: { navigate: ReturnType<typeof useNavigate> }) {
  return (
    <header className={headerStyles.header}>
      <div className={headerStyles.left}>
        <a onClick={() => navigate("/")} className={headerStyles.logo}>
          <img src="/logos/logo.svg" alt="Codurity" height={32} />
        </a>
      </div>
      <div className={headerStyles.right}>
        <Button variant="ghost" onClick={() => navigate("/login")}>
          Войти
        </Button>
        <Button variant="primary" onClick={() => navigate("/register")}>
          Создать аккаунт
        </Button>
      </div>
    </header>
  );
}

import headerStyles from "../widgets/header/ui/Header.module.css";
