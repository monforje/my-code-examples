import { useLocation } from "react-router";
import { AuthLayout, AuthMain, AuthCard, ForgotVerifyForm, AuthFooter } from "@features/auth";
import { PublicHeader } from "@widgets/header";

export function ForgotVerifyPage() {
  const location = useLocation();
  const email = (location.state as { email?: string })?.email ?? "";

  return (
    <AuthLayout>
      <PublicHeader />
      <AuthMain>
        <AuthCard title="Введите код восстановления">
          <p style={{ fontSize: "var(--cd-text-sm)", color: "var(--cd-text-muted)" }}>
            Код отправлен на {email}. Введите его ниже.
          </p>
          <ForgotVerifyForm />
        </AuthCard>
      </AuthMain>
      <AuthFooter />
    </AuthLayout>
  );
}
