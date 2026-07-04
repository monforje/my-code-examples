import { Navigate } from "react-router";
import { useAuth, AuthLayout, AuthMain, AuthFooter, CliLoginForm } from "@features/auth";
import { PublicHeader, AppHeader } from "@widgets/header";
import { Spinner } from "@shared/ui";
import styles from "./CliLoginPage.module.css";

export function CliLoginPage() {
  const { isAuthenticated, isLoading } = useAuth();

  if (isLoading) {
    return (
      <AuthLayout>
        <PublicHeader />
        <main className={styles.loading}>
          <Spinner size={22} />
          Проверяем авторизацию…
        </main>
        <AuthFooter />
      </AuthLayout>
    );
  }

  if (!isAuthenticated) {
    return <Navigate to="/login?redirect=/cli/login" replace />;
  }

  return (
    <AuthLayout>
      <AppHeader />
      <AuthMain>
        <div className={styles.card}>
          <div className={styles.header}>
            <h1 className={styles.title}>Вход в codurity CLI</h1>
            <p className={styles.desc}>Введите код, который показан в консоли:</p>
          </div>
          <CliLoginForm />
        </div>
      </AuthMain>
      <AuthFooter />
    </AuthLayout>
  );
}
