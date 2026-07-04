import { AuthLayout, AuthMain, AuthCard, ForgotForm, AuthFooter } from "@features/auth";
import { PublicHeader } from "@widgets/header";

export function ForgotPage() {
  return (
    <AuthLayout>
      <PublicHeader />
      <AuthMain>
        <AuthCard title="Восстановить пароль">
          <p style={{ fontSize: "var(--cd-text-sm)", color: "var(--cd-text-muted)" }}>
            Введите email, и мы отправим код для сброса пароля.
          </p>
          <ForgotForm />
        </AuthCard>
      </AuthMain>
      <AuthFooter />
    </AuthLayout>
  );
}
