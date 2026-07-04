import { useEffect, type ReactNode } from "react";
import styles from "./Modal.module.css";

interface ModalProps {
  title: string;
  onClose: () => void;
  step?: number;
  totalSteps?: number;
  children: ReactNode;
}

export function ModalFooter({ children }: { children: ReactNode }) {
  return <div className={styles.footer}>{children}</div>;
}

export function Modal({ title, onClose, step, totalSteps = 4, children }: ModalProps) {
  useEffect(() => {
    const handleEsc = (e: KeyboardEvent) => {
      if (e.key === "Escape") onClose();
    };
    document.addEventListener("keydown", handleEsc);
    return () => document.removeEventListener("keydown", handleEsc);
  }, [onClose]);

  return (
    <div className={styles.overlay} onClick={onClose}>
      <div className={styles.modal} onClick={(e) => e.stopPropagation()}>
        <div className={styles.header}>
          <h2 className={styles.title}>{title}</h2>
          <button type="button" className={styles.close} onClick={onClose}>
            ×
          </button>
        </div>

        {step !== undefined && (
          <div className={styles.steps}>
            {Array.from({ length: totalSteps }, (_, i) => (
              <div
                key={i}
                className={`${styles.step} ${i + 1 < step ? styles.stepDone : i + 1 === step ? styles.stepActive : ""}`}
              />
            ))}
          </div>
        )}

        <div className={styles.body}>{children}</div>
      </div>
    </div>
  );
}
