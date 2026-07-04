import { Icon } from "@iconify/react";
import styles from "./Toast.module.css";

interface ToastProps {
  message: string;
  type?: "success" | "error" | "info";
  onClose: () => void;
}

const icons = {
  success: "tabler:circle-check",
  error: "tabler:alert-circle-filled",
  info: "tabler:info-circle",
} as const;

export function Toast({ message, type = "info", onClose }: ToastProps) {
  return (
    <div className={`${styles.toast} ${styles[type]}`}>
      <span className={styles.icon}>
        <Icon icon={icons[type]} width={20} />
      </span>
      <span className={styles.message}>{message}</span>
      <button className={styles.close} onClick={onClose}>
        <Icon icon="tabler:x" width={18} />
      </button>
    </div>
  );
}
