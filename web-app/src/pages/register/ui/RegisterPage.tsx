import { AuthLayout, AuthMain, AuthCard, RegisterForm, AuthFooter } from "@features/auth";
import { PublicHeader } from "@widgets/header";

export function RegisterPage() {
  return (
    <AuthLayout>
      <PublicHeader />
      <AuthMain>
        <AuthCard title="Регистрация" titleLg>
          <p style={{ fontSize: "var(--cd-text-sm)", color: "var(--cd-text-muted)" }}>
            Зарегистрируйтесь, чтобы получить доступ к Codurity.
          </p>
          <RegisterForm />
        </AuthCard>
      </AuthMain>
      <AuthFooter />
    </AuthLayout>
  );
}
