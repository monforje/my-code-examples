import { useState, type FormEvent } from "react";
import { authDeviceConfirm } from "@shared/api";
import { getApiErrorMessage } from "@shared/lib/api-error";
import { CodeInput, Button, Alert } from "@shared/ui";
import { AuthForm, AuthFormActions } from "./AuthForm";
import { AuthSuccess } from "./AuthSuccess";

const CODE_LENGTH = 8;

function formatUserCode(raw: string): string {
  const clean = raw.slice(0, CODE_LENGTH);
  return `${clean.slice(0, 4)}-${clean.slice(4)}`;
}

export function CliLoginForm() {
  const [code, setCode] = useState("");
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(false);
  const [confirmed, setConfirmed] = useState(false);

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault();
    setError("");

    if (code.length < CODE_LENGTH) {
      setError("Введите полный код — 8 символов.");
      return;
    }

    setLoading(true);
    try {
      const res = await authDeviceConfirm({ user_code: formatUserCode(code) });
      if (res.status === 200) {
        setConfirmed(true);
        return;
      }
      setError(getApiErrorMessage(res, "Код недействителен или истёк."));
    } catch {
      setError("Код недействителен или истёк.");
    } finally {
      setLoading(false);
    }
  };

  if (confirmed) {
    return <AuthSuccess title="CLI-вход подтверждён" text="Вернитесь в терминал." />;
  }

  return (
    <AuthForm onSubmit={handleSubmit}>
      {error && <Alert variant="danger">{error}</Alert>}
      <CodeInput
        length={CODE_LENGTH}
        mode="alphanumeric"
        separatorAfter={4}
        autoFocus
        value={code}
        onChange={setCode}
      />
      <AuthFormActions>
        <Button
          type="submit"
          variant="primary"
          fullWidth
          loading={loading}
          disabled={code.length < CODE_LENGTH}
        >
          Подтвердить вход
        </Button>
      </AuthFormActions>
    </AuthForm>
  );
}
