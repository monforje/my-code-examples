import styles from "./AuthFooter.module.css";

export function AuthFooter() {
  return (
    <footer className={styles.footer}>
      <a href="#">Terms</a>
      <a href="#">Privacy</a>
      <a href="#">Support</a>
      <span>© 2026 Codurity</span>
    </footer>
  );
}
