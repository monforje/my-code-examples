import { useState, type FormEvent } from "react";
import { useNavigate } from "react-router";
import { Modal, ModalFooter } from "@shared/ui/Modal";
import { Field, PasswordInput, CodeInput, Button, Alert } from "@shared/ui";
import { useToast } from "@shared/lib/use-toast";
import { authMeDelete, authMeDeleteVerify, authMeDeleteCodeResend } from "@shared/api";
import { useAuth } from "@features/auth";
import { ResendButton } from "@features/auth";
import { getApiErrorMessage } from "@shared/lib/api-error";

interface DeleteAccountModalProps {
  onClose: () => void;
}

export function DeleteAccountModal({ onClose }: DeleteAccountModalProps) {
  const navigate = useNavigate();
  const { logout } = useAuth();
  const { addToast } = useToast();
  const [step, setStep] = useState(1);
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(false);
  const [code, setCode] = useState("");

  const handleStep2 = async (e: FormEvent) => {
    e.preventDefault();
    setError("");
    const form = e.target as HTMLFormElement;
    const password = (form.elements.namedItem("password") as HTMLInputElement).value;
    if (!password) return;

    setLoading(true);
    try {
      const res = await authMeDelete({ password });
      const status = (res as { status: number }).status;
      if (status !== 200) {
        setError(getApiErrorMessage(res, "Не удалось запросить удаление аккаунта."));
        return;
      }
      addToast("Код отправлен", "info");
      setStep(3);
    } catch {
      setError("Ошибка");
    } finally {
      setLoading(false);
    }
  };

  const handleStep3 = async () => {
    setError("");
    if (code.length < 6) {
      setError("Введите полный код");
      return;
    }

    setLoading(true);
    try {
      const res = await authMeDeleteVerify({ code });
      const status = (res as { status: number }).status;
      if (status !== 204) {
        setError(getApiErrorMessage(res, "Не удалось удалить аккаунт."));
        return;
      }
      addToast("Аккаунт удалён", "success");
      await logout();
      navigate("/account-deleted");
    } catch {
      setError("Ошибка");
    } finally {
      setLoading(false);
    }
  };

  const handleResend = async () => {
    setError("");
    const res = await authMeDeleteCodeResend();
    if (res.status !== 200) {
      setError(getApiErrorMessage(res, "Не удалось отправить код повторно."));
      throw new Error("resend failed");
    }
  };

  return (
    <Modal title="Удалить аккаунт?" onClose={onClose} step={step}>
      {error && <Alert variant="danger">{error}</Alert>}

      {step === 1 && (
        <>
          <p style={{ fontSize: "var(--cd-text-sm)", color: "var(--cd-text-muted)" }}>
            Это действие нельзя отменить. Доступ к аккаунту и данным будет закрыт.
          </p>
          <ul
            style={{
              listStyle: "none",
              padding: 0,
              margin: 0,
              display: "grid",
              gap: "var(--cd-space-2)",
            }}
          >
            {[
              "Профиль будет недоступен",
              "Вход в аккаунт будет невозможен",
              "Восстановление может быть недоступно",
            ].map((t) => (
              <li
                key={t}
                style={{
                  fontSize: "var(--cd-text-sm)",
                  color: "var(--cd-text-muted)",
                  display: "flex",
                  gap: "var(--cd-space-2)",
                }}
              >
                <span style={{ color: "var(--cd-text-subtle)" }}>—</span> {t}
              </li>
            ))}
          </ul>
          <ModalFooter>
            <Button variant="ghost" onClick={onClose}>
              Отмена
            </Button>
            <Button variant="danger" onClick={() => setStep(2)}>
              Продолжить удаление
            </Button>
          </ModalFooter>
        </>
      )}

      {step === 2 && (
        <form onSubmit={handleStep2}>
          <Field label="Пароль">
            <PasswordInput name="password" placeholder="Введите пароль" required />
          </Field>
          <ModalFooter>
            <Button variant="ghost" onClick={onClose}>
              Отмена
            </Button>
            <Button type="submit" variant="danger" loading={loading}>
              Запросить код удаления
            </Button>
          </ModalFooter>
        </form>
      )}

      {step === 3 && (
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
            <Button variant="danger" onClick={handleStep3} loading={loading}>
              Удалить аккаунт
            </Button>
          </ModalFooter>
        </>
      )}
    </Modal>
  );
}
