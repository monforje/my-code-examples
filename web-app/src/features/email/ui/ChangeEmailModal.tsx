import { useState, type FormEvent } from "react";
import { Modal, ModalFooter } from "@shared/ui/Modal";
import { Field, Input, PasswordInput, CodeInput, Button, Alert } from "@shared/ui";
import { useToast } from "@shared/lib/use-toast";
import {
  authMeEmailChange,
  authMeEmailChangeVerify,
  authMeEmailChangeConfirm,
  authMeEmailChangeComplete,
  authMeEmailChangeCodeResend,
} from "@shared/api";
import { useAuth } from "@features/auth";
import { ResendButton } from "@features/auth";
import { getApiErrorMessage } from "@shared/lib/api-error";

interface ChangeEmailModalProps {
  onClose: () => void;
}

export function ChangeEmailModal({ onClose }: ChangeEmailModalProps) {
  const { user, setUser } = useAuth();
  const { addToast } = useToast();
  const [step, setStep] = useState(1);
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(false);
  const [identityToken, setIdentityToken] = useState("");
  const [newEmail, setNewEmail] = useState("");
  const [currentCode, setCurrentCode] = useState("");
  const [newCode, setNewCode] = useState("");

  const handleStep1 = async (e: FormEvent) => {
    e.preventDefault();
    setError("");
    const form = e.target as HTMLFormElement;
    const password = (form.elements.namedItem("password") as HTMLInputElement).value;
    if (!password) return;

    setLoading(true);
    try {
      const res = await authMeEmailChange({ password });
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
    if (currentCode.length < 6) {
      setError("Введите полный код");
      return;
    }

    setLoading(true);
    try {
      const res = await authMeEmailChangeVerify({ code: currentCode });
      if ("data" in res && res.status === 200) {
        setIdentityToken(res.data.identity_token);
        setCurrentCode("");
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
    const email = (form.elements.namedItem("new_email") as HTMLInputElement).value;
    if (!email) return;

    setLoading(true);
    try {
      const res = await authMeEmailChangeConfirm({
        new_email: email,
        identity_token: identityToken,
      });
      if ("data" in res && res.status !== 200) {
        setError(getApiErrorMessage(res, "Не удалось подтвердить новый email."));
        return;
      }
      setNewEmail(email);
      setStep(4);
    } catch {
      setError("Ошибка");
    } finally {
      setLoading(false);
    }
  };

  const handleStep4 = async () => {
    setError("");
    if (newCode.length < 6) {
      setError("Введите полный код");
      return;
    }

    setLoading(true);
    try {
      const res = await authMeEmailChangeComplete({ code: newCode });
      if ("data" in res && res.status !== 200) {
        setError(getApiErrorMessage(res, "Не удалось завершить смену email."));
        return;
      }
      if (user) {
        setUser({ ...user, email: newEmail });
      }
      addToast("Email изменён", "success");
      addToast("Все остальные сессии были завершены", "info");
      onClose();
    } catch {
      setError("Ошибка");
    } finally {
      setLoading(false);
    }
  };

  const handleResendCurrent = async () => {
    setError("");
    const res = await authMeEmailChangeCodeResend({ step: "current" });
    if (res.status !== 200) {
      setError(getApiErrorMessage(res, "Не удалось отправить код повторно."));
      throw new Error("resend failed");
    }
  };

  const handleResendNew = async () => {
    setError("");
    const res = await authMeEmailChangeCodeResend({ step: "new" });
    if (res.status !== 200) {
      setError(getApiErrorMessage(res, "Не удалось отправить код повторно."));
      throw new Error("resend failed");
    }
  };

  return (
    <Modal title="Смена email" onClose={onClose} step={step}>
      {error && <Alert variant="danger">{error}</Alert>}

      {step === 1 && (
        <>
          <p style={{ fontSize: "var(--cd-text-sm)", color: "var(--cd-text-muted)" }}>
            Для смены email подтвердите, что это вы.
          </p>
          <form onSubmit={handleStep1}>
            <Field label="Пароль">
              <PasswordInput name="password" placeholder="Введите пароль" required />
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
        </>
      )}

      {step === 2 && (
        <>
          <p style={{ fontSize: "var(--cd-text-sm)", color: "var(--cd-text-muted)" }}>
            Введите код, отправленный на текущий email.
          </p>
          <CodeInput value={currentCode} onChange={setCurrentCode} />
          <ResendButton onResend={handleResendCurrent} />
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
          <Field label="Новый email">
            <Input name="new_email" type="email" placeholder="name@company.com" required />
          </Field>
          <ModalFooter>
            <Button variant="ghost" onClick={() => setStep(2)}>
              Назад
            </Button>
            <Button type="submit" variant="primary" loading={loading}>
              Отправить код на новый email
            </Button>
          </ModalFooter>
        </form>
      )}

      {step === 4 && (
        <>
          <p style={{ fontSize: "var(--cd-text-sm)", color: "var(--cd-text-muted)" }}>
            Введите код, отправленный на <strong>{newEmail}</strong>.
          </p>
          <CodeInput value={newCode} onChange={setNewCode} />
          <ResendButton onResend={handleResendNew} />
          <ModalFooter>
            <Button variant="ghost" onClick={onClose}>
              Отмена
            </Button>
            <Button variant="danger" onClick={handleStep4} loading={loading}>
              Завершить смену email
            </Button>
          </ModalFooter>
        </>
      )}
    </Modal>
  );
}
