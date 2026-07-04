import type { ReactNode } from "react";
import styles from "./Field.module.css";

interface FieldProps {
  label?: string;
  labelExtra?: ReactNode;
  error?: string;
  children: ReactNode;
}

export function Field({ label, labelExtra, error, children }: FieldProps) {
  return (
    <div className={styles.field}>
      {(label || labelExtra) && (
        <div className={styles.labelRow}>
          {label && <label className={styles.label}>{label}</label>}
          {labelExtra && <span className={styles.labelExtra}>{labelExtra}</span>}
        </div>
      )}
      {children}
      {error && <div className={styles.error}>{error}</div>}
    </div>
  );
}
