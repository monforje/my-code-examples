import { useState, type FormEvent } from "react";
import { useNavigate, useSearchParams } from "react-router";
import { useAuth, EmailNotVerifiedError } from "@features/auth";
import { useToast } from "@shared/lib/use-toast";
import { AuthForm, AuthFormActions } from "./AuthForm";
import { Field, Input, PasswordInput, Button, Alert } from "@shared/ui";

export function LoginForm() {
  const navigate = useNavigate();
  const [searchParams] = useSearchParams();
  const redirect = searchParams.get("redirect");
  const { login } = useAuth();
  const { addToast } = useToast();
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
      addToast("Добро пожаловать!", "success");
      navigate(redirect ?? "/profile");
    } catch (err: unknown) {
      if (err instanceof EmailNotVerifiedError) {
        navigate("/verify", { state: { email: err.email } });
        return;
      }
      const msg = err instanceof Error ? err.message : "Ошибка входа";
      setError(msg);
    } finally {
      setLoading(false);
    }
  };

  return (
    <AuthForm onSubmit={handleSubmit}>
      {error && <Alert variant="danger">{error}</Alert>}
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
        <Button type="button" variant="default" fullWidth onClick={() => navigate("/register")}>
          Создать аккаунт
        </Button>
      </AuthFormActions>
      <div style={{ display: "flex", justifyContent: "flex-end" }}>
        <button
          type="button"
          onClick={() => navigate("/forgot")}
          style={{
            fontSize: "var(--cd-text-sm)",
            color: "var(--cd-primary)",
            fontWeight: "var(--cd-weight-medium)",
            background: "none",
            border: "none",
            cursor: "pointer",
          }}
        >
          Забыли пароль?
        </button>
      </div>
    </AuthForm>
  );
}
