import type { ReactNode } from "react";
import styles from "./AuthLayout.module.css";

export function AuthLayout({ children }: { children: ReactNode }) {
  return <div className={styles.page}>{children}</div>;
}

export function AuthMain({ children }: { children: ReactNode }) {
  return <main className={styles.main}>{children}</main>;
}
