import { useEffect, useRef, useState, type FormEvent } from "react";
import { useNavigate } from "react-router";
import { authPasswordForgot } from "@shared/api";
import { AuthForm, AuthFormActions } from "./AuthForm";
import { Field, Input, Button, Alert } from "@shared/ui";
import { getApiErrorMessage } from "@shared/lib/api-error";
import styles from "./AuthForm.module.css";

export function ForgotForm() {
  const navigate = useNavigate();
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(false);
  const [sent, setSent] = useState(false);
  const navigateTimerRef = useRef<number | null>(null);

  useEffect(() => {
    return () => {
      if (navigateTimerRef.current) window.clearTimeout(navigateTimerRef.current);
    };
  }, []);

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault();
    setError("");
    const form = e.target as HTMLFormElement;
    const email = (form.elements.namedItem("email") as HTMLInputElement).value;

    if (!email) return;

    setLoading(true);
    try {
      const res = await authPasswordForgot({ email });
      if (res.status !== 200) {
        setError(getApiErrorMessage(res, "Не удалось отправить код."));
        return;
      }
      setSent(true);
      if (navigateTimerRef.current) window.clearTimeout(navigateTimerRef.current);
      navigateTimerRef.current = window.setTimeout(() => {
        navigate("/forgot/verify", { state: { email } });
      }, 1200);
    } catch {
      setError("Не удалось отправить код");
    } finally {
      setLoading(false);
    }
  };

  return (
    <AuthForm onSubmit={handleSubmit}>
      {error && <Alert variant="danger">{error}</Alert>}
      {sent && (
        <Alert variant="info">
          Если аккаунт с таким email существует, мы отправили код восстановления.
        </Alert>
      )}
      <Field label="Email">
        <Input
          name="email"
          type="email"
          placeholder="name@company.com"
          required
          autoComplete="email"
        />
      </Field>
      <AuthFormActions>
        <Button type="submit" variant="primary" fullWidth loading={loading}>
          Отправить код
        </Button>
      </AuthFormActions>
      <div style={{ display: "flex", justifyContent: "center" }}>
        <button type="button" className={styles.link} onClick={() => navigate("/login")}>
          Вернуться ко входу
        </button>
      </div>
    </AuthForm>
  );
}
