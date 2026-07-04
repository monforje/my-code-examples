import type { ReactNode } from "react";
import { Icon } from "@iconify/react";
import styles from "./AuthSuccess.module.css";

export function AuthSuccess({
  title,
  text,
  children,
}: {
  title: string;
  text?: string;
  children?: ReactNode;
}) {
  return (
    <div className={styles.success}>
      <div className={styles.icon}>
        <Icon icon="tabler:circle-check" width={32} />
      </div>
      <h2 className={styles.title}>{title}</h2>
      {text && <p className={styles.text}>{text}</p>}
      {children}
    </div>
  );
}

export function AuthCentered({
  icon,
  iconVariant = "info",
  title,
  text,
  children,
}: {
  icon: string;
  iconVariant?: "warning" | "danger" | "info";
  title: string;
  text?: string;
  children?: ReactNode;
}) {
  return (
    <div className={styles.centered}>
      <div className={`${styles.centeredIcon} ${styles[iconVariant]}`}>
        <Icon icon={icon} width={36} />
      </div>
      <h1 className={styles.centeredTitle}>{title}</h1>
      {text && <p className={styles.centeredText}>{text}</p>}
      {children}
    </div>
  );
}
