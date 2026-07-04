import styles from "./TaskTypeDot.module.css";
import { TYPE_COLOR } from "../model/types";

export function TaskTypeDot({ type }: { type: string }) {
  const color = TYPE_COLOR[type.toLowerCase()] || "var(--cd-primary)";
  const display = type.charAt(0).toUpperCase() + type.slice(1);
  return (
    <span className={styles.wrapper}>
      <span className={styles.dot} style={{ background: color }} />
      <span className={styles.label}>{display}</span>
    </span>
  );
}
