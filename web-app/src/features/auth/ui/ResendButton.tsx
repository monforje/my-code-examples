import { useCallback, useEffect, useRef, useState } from "react";
import styles from "./AuthForm.module.css";

interface ResendButtonProps {
  onResend: () => Promise<unknown>;
  seconds?: number;
}

export function ResendButton({ onResend, seconds = 30 }: ResendButtonProps) {
  const [countdown, setCountdown] = useState(0);
  const [isSending, setIsSending] = useState(false);
  const timerRef = useRef<ReturnType<typeof setInterval> | null>(null);

  useEffect(() => {
    return () => {
      if (timerRef.current) clearInterval(timerRef.current);
    };
  }, []);

  const startCountdown = useCallback(() => {
    if (timerRef.current) clearInterval(timerRef.current);
    setCountdown(seconds);
    timerRef.current = setInterval(() => {
      setCountdown((prev) => {
        if (prev <= 1) {
          if (timerRef.current) clearInterval(timerRef.current);
          return 0;
        }
        return prev - 1;
      });
    }, 1000);
  }, [seconds]);

  const handleResend = useCallback(async () => {
    setIsSending(true);
    try {
      await onResend();
      startCountdown();
    } catch {
      // Parent form renders the user-facing error message.
    } finally {
      setIsSending(false);
    }
  }, [onResend, startCountdown]);

  const pad = (n: number) => String(n).padStart(2, "0");

  return (
    <button
      type="button"
      className={styles.link}
      disabled={countdown > 0 || isSending}
      onClick={handleResend}
    >
      {isSending
        ? "Отправляем..."
        : countdown > 0
          ? `Повторно через ${pad(Math.floor(countdown / 60))}:${pad(countdown % 60)}`
          : "Отправить код ещё раз"}
    </button>
  );
}
