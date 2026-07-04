import { AuthLayout, AuthMain, AuthCard, ResetForm, AuthFooter } from "@features/auth";
import { PublicHeader } from "@widgets/header";

export function ResetPage() {
  return (
    <AuthLayout>
      <PublicHeader />
      <AuthMain>
        <AuthCard title="Создайте новый пароль">
          <ResetForm />
        </AuthCard>
      </AuthMain>
      <AuthFooter />
    </AuthLayout>
  );
}
