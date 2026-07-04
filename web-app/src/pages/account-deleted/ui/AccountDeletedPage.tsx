import { useNavigate } from "react-router";
import { AuthLayout, AuthMain, AuthCard, AuthSuccess, AuthFooter } from "@features/auth";
import { PublicHeader } from "@widgets/header";
import { Button } from "@shared/ui";

export function AccountDeletedPage() {
  const navigate = useNavigate();

  return (
    <AuthLayout>
      <PublicHeader />
      <AuthMain>
        <AuthCard title="">
          <AuthSuccess title="Аккаунт удалён" text="Ваш аккаунт был удалён.">
            <Button variant="primary" onClick={() => navigate("/login")}>
              На главную
            </Button>
          </AuthSuccess>
        </AuthCard>
      </AuthMain>
      <AuthFooter />
    </AuthLayout>
  );
}
