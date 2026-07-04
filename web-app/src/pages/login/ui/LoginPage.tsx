import { AuthLayout, AuthMain, AuthCard, LoginForm, AuthFooter } from "@features/auth";
import { PublicHeader } from "@widgets/header";

export function LoginPage() {
  return (
    <AuthLayout>
      <PublicHeader />
      <AuthMain>
        <AuthCard title="Вход" titleLg>
          <p style={{ fontSize: "var(--cd-text-sm)", color: "var(--cd-text-muted)" }}>
            Используйте рабочий email, чтобы продолжить работу с платформой.
          </p>
          <LoginForm />
        </AuthCard>
      </AuthMain>
      <AuthFooter />
    </AuthLayout>
  );
}
