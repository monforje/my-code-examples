import type { ReactNode } from "react";
import { Link } from "react-router";
import styles from "./SettingsLayout.module.css";

interface SettingsLayoutProps {
  children: ReactNode;
  sidebar?: "profile" | "settings";
  activeTab?: "profile" | "security";
  title: string;
  description: string;
}

export function SettingsLayout({
  children,
  sidebar,
  activeTab,
  title,
  description,
}: SettingsLayoutProps) {
  return (
    <div className={styles.layout}>
      {sidebar && (
        <aside className={styles.sidebar}>
          <nav className={styles.nav}>
            {sidebar === "profile" && (
              <Link to="/profile" className={`${styles.item} ${styles.itemActive}`}>
                Профиль
              </Link>
            )}
            {sidebar === "settings" && (
              <>
                <Link
                  to="/settings/profile"
                  className={`${styles.item} ${activeTab === "profile" ? styles.itemActive : ""}`}
                >
                  Настройки профиля
                </Link>
                <Link
                  to="/settings/security"
                  className={`${styles.item} ${activeTab === "security" ? styles.itemActive : ""}`}
                >
                  Безопасность
                </Link>
              </>
            )}
          </nav>
        </aside>
      )}
      <div className={styles.content}>
        <div className={styles.contentHeader}>
          <h1 className={styles.title}>{title}</h1>
          <p className={styles.desc}>{description}</p>
        </div>
        {children}
      </div>
    </div>
  );
}
