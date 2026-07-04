import { useState, type FormEvent } from "react";
import { useNavigate, useLocation } from "react-router";
import { authRegisterVerify, authRegisterCodeResend } from "@shared/api";
import { useToast } from "@shared/lib/use-toast";
import { AuthForm, AuthFormActions } from "./AuthForm";
import { CodeInput, Button, Alert } from "@shared/ui";
import { ResendButton } from "./ResendButton";
import { getApiErrorMessage } from "@shared/lib/api-error";

export function VerifyForm() {
  const navigate = useNavigate();
  const location = useLocation();
  const email = (location.state as { email?: string })?.email ?? "";
  const { addToast } = useToast();
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
      const res = await authRegisterVerify({ email, code });
      if ("data" in res && res.status !== 200) {
        setError(getApiErrorMessage(res, "Не удалось подтвердить email."));
        return;
      }
      addToast("Email подтверждён!", "success");
      navigate("/login");
    } catch {
      setError("Ошибка верификации");
    } finally {
      setLoading(false);
    }
  };

  const handleResend = async () => {
    setError("");
    const res = await authRegisterCodeResend({ email });
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
          Подтвердить
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
        <button type="button" className={styles.link} onClick={() => navigate("/register")}>
          Изменить email
        </button>
      </div>
    </AuthForm>
  );
}

import styles from "./AuthForm.module.css";
