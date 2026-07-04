import { Link } from "react-router";
import logo from "@shared/assets/logo.svg";
import { ThemeToggle } from "@features/theme-toggle";
import styles from "./Header.module.css";

export function PublicHeader() {
  return (
    <header className={styles.header}>
      <div className={styles.left}>
        <Link to="/" className={styles.logo}>
          <img src={logo} alt="Codurity" height={32} />
        </Link>
      </div>
      <div className={styles.right}>
        <ThemeToggle />
        <Link to="/login" className="cd-button cd-button-ghost cd-button-sm">
          Войти
        </Link>
        <Link to="/register" className="cd-button cd-button-primary cd-button-sm">
          Создать аккаунт
        </Link>
      </div>
    </header>
  );
}
