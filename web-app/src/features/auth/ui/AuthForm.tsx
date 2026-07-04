import type { FormEvent, ReactNode } from "react";
import styles from "./AuthForm.module.css";

interface AuthFormProps {
  onSubmit: (e: FormEvent) => void;
  children: ReactNode;
}

export function AuthForm({ onSubmit, children }: AuthFormProps) {
  return (
    <form className={styles.form} onSubmit={onSubmit}>
      {children}
    </form>
  );
}

export function AuthFormActions({ children }: { children: ReactNode }) {
  return <div className={styles.actions}>{children}</div>;
}

export function AuthFormLinks({ children }: { children: ReactNode }) {
  return <div className={styles.links}>{children}</div>;
}
