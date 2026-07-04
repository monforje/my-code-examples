import { useState, type FormEvent } from "react";
import { useNavigate, useLocation } from "react-router";
import { authPasswordForgotVerify, authPasswordForgotCodeResend } from "@shared/api";
import { AuthForm, AuthFormActions } from "./AuthForm";
import { CodeInput, Button, Alert } from "@shared/ui";
import { ResendButton } from "./ResendButton";
import { getApiErrorMessage } from "@shared/lib/api-error";

export function ForgotVerifyForm() {
  const navigate = useNavigate();
  const location = useLocation();
  const email = (location.state as { email?: string })?.email ?? "";
  const [code, setCode] = useState("");
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(false);

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault();
    setError("");
    if (code.length < 6) {
      setError("Введите полный код");
      return;
    }

    setLoading(true);
    try {
      const res = await authPasswordForgotVerify({ email, code });
      if ("data" in res && res.status === 200) {
        navigate("/reset", {
          state: { reset_token: res.data.reset_token },
        });
        return;
      }
      setError(getApiErrorMessage(res, "Не удалось проверить код."));
    } catch {
      setError("Ошибка верификации");
    } finally {
      setLoading(false);
    }
  };

  const handleResend = async () => {
    setError("");
    const res = await authPasswordForgotCodeResend({ email });
    if (res.status !== 200) {
      setError(getApiErrorMessage(res, "Не удалось отправить код повторно."));
      throw new Error("resend failed");
    }
  };

  return (
    <AuthForm onSubmit={handleSubmit}>
      {error && <Alert variant="danger">{error}</Alert>}
      <CodeInput value={code} onChange={setCode} />
      <AuthFormActions>
        <Button type="submit" variant="primary" fullWidth loading={loading}>
          Продолжить
        </Button>
      </AuthFormActions>
      <div
        style={{
          display: "flex",
          flexDirection: "column",
          alignItems: "center",
          gap: "var(--cd-space-2)",
        }}
      >
        <ResendButton onResend={handleResend} />
        <button type="button" className={styles.link} onClick={() => navigate("/forgot")}>
          Изменить email
        </button>
      </div>
    </AuthForm>
  );
}

import styles from "./AuthForm.module.css";
