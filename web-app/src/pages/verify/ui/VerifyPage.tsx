import { useLocation } from "react-router";
import { AuthLayout, AuthMain, AuthCard, VerifyForm, AuthFooter } from "@features/auth";
import { PublicHeader } from "@widgets/header";

export function VerifyPage() {
  const location = useLocation();
  const email = (location.state as { email?: string })?.email ?? "";

  return (
    <AuthLayout>
      <PublicHeader />
      <AuthMain>
        <AuthCard title="Подтвердите email">
          <p style={{ fontSize: "var(--cd-text-sm)", color: "var(--cd-text-muted)" }}>
            Мы отправили код на {email}. Введите его ниже.
          </p>
          <VerifyForm />
        </AuthCard>
      </AuthMain>
      <AuthFooter />
    </AuthLayout>
  );
}
