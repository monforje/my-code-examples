import { useState, type FormEvent } from "react";
import { useNavigate } from "react-router";
import { useAuth } from "@features/auth";
import { AuthForm, AuthFormActions } from "./AuthForm";
import {
  Field,
  Input,
  PasswordInput,
  Button,
  Alert,
  PasswordRequirements,
  PasswordStrengthBar,
} from "@shared/ui";

export function RegisterForm() {
  const navigate = useNavigate();
  const { register } = useAuth();
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(false);
  const [password, setPassword] = useState("");

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault();
    setError("");
    const form = e.target as HTMLFormElement;
    const email = (form.elements.namedItem("email") as HTMLInputElement).value;
    const pass = (form.elements.namedItem("password") as HTMLInputElement).value;
    const pass2 = (form.elements.namedItem("password2") as HTMLInputElement).value;

    if (!email || !pass || !pass2) return;
    if (pass.length < 8) {
      setError("Пароль должен содержать минимум 8 символов");
      return;
    }
    if (pass !== pass2) {
      setError("Пароли не совпадают");
      return;
    }

    setLoading(true);
    try {
      await register(email, pass);
      navigate("/verify", { state: { email } });
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : "Ошибка регистрации";
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
      <Field label="Пароль" labelExtra={<PasswordStrengthBar password={password} />}>
        <PasswordInput
          name="password"
          placeholder="Введите пароль"
          required
          autoComplete="new-password"
          value={password}
          onChange={(e) => setPassword(e.target.value)}
        />
        <PasswordRequirements password={password} />
      </Field>
      <Field label="Повторите пароль">
        <PasswordInput
          name="password2"
          placeholder="Введите пароль"
          required
          autoComplete="new-password"
        />
      </Field>
      <AuthFormActions>
        <Button type="submit" variant="primary" fullWidth loading={loading}>
          Создать аккаунт
        </Button>
        <Button type="button" variant="default" fullWidth onClick={() => navigate("/login")}>
          Уже есть аккаунт? Войти
        </Button>
      </AuthFormActions>
    </AuthForm>
  );
}
