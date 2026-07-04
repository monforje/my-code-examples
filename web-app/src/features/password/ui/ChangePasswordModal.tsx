import { useState, type FormEvent } from "react";
import { Modal, ModalFooter } from "@shared/ui/Modal";
import { Field, PasswordInput, CodeInput, Button, Alert, PasswordStrengthBar } from "@shared/ui";
import { useToast } from "@shared/lib/use-toast";
import {
  authPasswordChange,
  authPasswordChangeVerify,
  authPasswordChangeComplete,
  authPasswordChangeCodeResend,
} from "@shared/api";
import { ResendButton } from "@features/auth";
import { getApiErrorMessage } from "@shared/lib/api-error";

interface ChangePasswordModalProps {
  onClose: () => void;
}

export function ChangePasswordModal({ onClose }: ChangePasswordModalProps) {
  const [step, setStep] = useState(1);
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(false);
  const [changeToken, setChangeToken] = useState("");
  const [code, setCode] = useState("");
  const [newPassword, setNewPassword] = useState("");
  const { addToast } = useToast();

  const handleStep1 = async (e: FormEvent) => {
    e.preventDefault();
    setError("");
    const form = e.target as HTMLFormElement;
    const password = (form.elements.namedItem("current_password") as HTMLInputElement).value;
    if (!password) return;

    setLoading(true);
    try {
      const res = await authPasswordChange({ current_password: password });
      if ("data" in res && res.status !== 200) {
        setError(getApiErrorMessage(res, "Не удалось отправить код."));
        return;
      }
      addToast("Код отправлен", "success");
      setStep(2);
    } catch {
      setError("Ошибка");
    } finally {
      setLoading(false);
    }
  };

  const handleStep2 = async () => {
    setError("");
    if (code.length < 6) {
      setError("Введите полный код");
      return;
    }

    setLoading(true);
    try {
      const res = await authPasswordChangeVerify({ code });
      if ("data" in res && res.status === 200) {
        setChangeToken(res.data.change_token);
        setCode("");
        setStep(3);
        return;
      }
      setError(getApiErrorMessage(res, "Не удалось проверить код."));
    } catch {
      setError("Ошибка");
    } finally {
      setLoading(false);
    }
  };

  const handleStep3 = async (e: FormEvent) => {
    e.preventDefault();
    setError("");
    const form = e.target as HTMLFormElement;
    const p1 = (form.elements.namedItem("new_password") as HTMLInputElement).value;
    const p2 = (form.elements.namedItem("new_password2") as HTMLInputElement).value;
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
      const res = await authPasswordChangeComplete({ change_token: changeToken, new_password: p1 });
      if ("data" in res && res.status !== 200) {
        setError(getApiErrorMessage(res, "Не удалось сохранить пароль."));
        return;
      }
      addToast("Пароль изменён", "success");
      addToast("Все остальные сессии были завершены", "info");
      onClose();
    } catch {
      setError("Ошибка");
    } finally {
      setLoading(false);
    }
  };

  const handleResend = async () => {
    setError("");
    const res = await authPasswordChangeCodeResend();
    if (res.status !== 200) {
      setError(getApiErrorMessage(res, "Не удалось отправить код повторно."));
      throw new Error("resend failed");
    }
  };

  return (
    <Modal title="Смена пароля" onClose={onClose} step={step}>
      {error && <Alert variant="danger">{error}</Alert>}

      {step === 1 && (
        <form onSubmit={handleStep1}>
          <Field label="Текущий пароль">
            <PasswordInput name="current_password" placeholder="Введите пароль" required />
          </Field>
          <ModalFooter>
            <Button variant="ghost" onClick={onClose}>
              Отмена
            </Button>
            <Button type="submit" variant="primary" loading={loading}>
              Продолжить
            </Button>
          </ModalFooter>
        </form>
      )}

      {step === 2 && (
        <>
          <p style={{ fontSize: "var(--cd-text-sm)", color: "var(--cd-text-muted)" }}>
            Мы отправили код подтверждения на ваш email.
          </p>
          <CodeInput value={code} onChange={setCode} />
          <ResendButton onResend={handleResend} />
          <ModalFooter>
            <Button variant="ghost" onClick={onClose}>
              Отмена
            </Button>
            <Button variant="primary" onClick={handleStep2} loading={loading}>
              Подтвердить код
            </Button>
          </ModalFooter>
        </>
      )}

      {step === 3 && (
        <form onSubmit={handleStep3}>
          <Field label="Новый пароль" labelExtra={<PasswordStrengthBar password={newPassword} />}>
            <PasswordInput
              name="new_password"
              placeholder="Введите пароль"
              required
              autoComplete="new-password"
              value={newPassword}
              onChange={(e) => setNewPassword(e.target.value)}
            />
          </Field>
          <div style={{ marginTop: "var(--cd-space-4)" }}>
            <Field label="Повторите пароль">
              <PasswordInput
                name="new_password2"
                placeholder="Введите пароль"
                required
                autoComplete="new-password"
              />
            </Field>
          </div>
          <ModalFooter>
            <Button variant="ghost" onClick={onClose}>
              Отмена
            </Button>
            <Button type="submit" variant="primary" loading={loading}>
              Сохранить пароль
            </Button>
          </ModalFooter>
        </form>
      )}
    </Modal>
  );
}
