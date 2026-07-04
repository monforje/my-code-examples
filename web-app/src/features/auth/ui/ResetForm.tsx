import { useState, type FormEvent } from "react";
import { useNavigate, useLocation } from "react-router";
import { authPasswordReset } from "@shared/api";
import { useToast } from "@shared/lib/use-toast";
import { AuthForm, AuthFormActions } from "./AuthForm";
import {
  Field,
  PasswordInput,
  Button,
  Alert,
  PasswordRequirements,
  PasswordStrengthBar,
} from "@shared/ui";
import { AuthSuccess } from "./AuthSuccess";
import { getApiErrorMessage } from "@shared/lib/api-error";

export function ResetForm() {
  const navigate = useNavigate();
  const location = useLocation();
  const resetToken = (location.state as { reset_token?: string })?.reset_token ?? "";
  const { addToast } = useToast();
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(false);
  const [success, setSuccess] = useState(false);
  const [password, setPassword] = useState("");

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault();
    setError("");
    const form = e.target as HTMLFormElement;
    const p1 = (form.elements.namedItem("password") as HTMLInputElement).value;
    const p2 = (form.elements.namedItem("password2") as HTMLInputElement).value;

    if (p1.length < 8) {
      setError("Пароль должен содержать минимум 8 символов");
      return;
    }
    if (p1 !== p2) {
      setError("Пароли не совпадают");
      return;
    }

    setLoading(true);
    try {
      const res = await authPasswordReset({ reset_token: resetToken, new_password: p1 });
      if (res.status !== 200) {
        setError(getApiErrorMessage(res, "Не удалось сбросить пароль."));
        return;
      }
      addToast("Все остальные сессии были завершены", "info");
      setSuccess(true);
    } catch {
      setError("Не удалось сбросить пароль");
    } finally {
      setLoading(false);
    }
  };

  if (success) {
    return (
      <AuthSuccess title="Пароль обновлён" text="Теперь вы можете войти с новым паролем.">
        <Button variant="primary" onClick={() => navigate("/login")}>
          Войти
        </Button>
      </AuthSuccess>
    );
  }

  return (
    <AuthForm onSubmit={handleSubmit}>
      {error && <Alert variant="danger">{error}</Alert>}
      <Field label="Новый пароль" labelExtra={<PasswordStrengthBar password={password} />}>
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
          Сохранить пароль
        </Button>
      </AuthFormActions>
    </AuthForm>
  );
}
