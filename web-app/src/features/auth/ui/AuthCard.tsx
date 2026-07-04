import type { ReactNode } from "react";
import styles from "./AuthCard.module.css";

interface AuthCardProps {
  titleLg?: boolean;
  title: string;
  desc?: string;
  children: ReactNode;
}

export function AuthCard({ titleLg, title, desc, children }: AuthCardProps) {
  return (
    <div className={styles.card}>
      <div className={styles.header}>
        <h1 className={`${styles.title} ${titleLg ? styles.titleLg : ""}`}>{title}</h1>
        {desc && <p className={styles.desc}>{desc}</p>}
      </div>
      {children}
    </div>
  );
}
