import type { ReactNode } from "react";
import { Icon } from "@iconify/react";
import styles from "./Alert.module.css";

type AlertVariant = "danger" | "warning" | "info" | "success";

interface AlertProps {
  variant?: AlertVariant;
  icon?: string;
  children: ReactNode;
}

const defaultIcons: Record<AlertVariant, string> = {
  danger: "tabler:alert-circle-filled",
  warning: "tabler:alert-circle",
  info: "tabler:info-circle",
  success: "tabler:circle-check",
};

export function Alert({ variant = "info", icon, children }: AlertProps) {
  return (
    <div className={`${styles.alert} ${styles[variant]}`}>
      <span className={styles.icon}>
        <Icon icon={icon ?? defaultIcons[variant]} width={20} />
      </span>
      <span className={styles.text}>{children}</span>
    </div>
  );
}
