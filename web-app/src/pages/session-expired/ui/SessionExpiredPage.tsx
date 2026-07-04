import { useNavigate } from "react-router";
import { AuthLayout, AuthMain, AuthCard, AuthCentered, AuthFooter } from "@features/auth";
import { PublicHeader } from "@widgets/header";
import { Button } from "@shared/ui";

export function SessionExpiredPage() {
  const navigate = useNavigate();

  return (
    <AuthLayout>
      <PublicHeader />
      <AuthMain>
        <AuthCard title="">
          <AuthCentered
            icon="tabler:clock-alert"
            iconVariant="warning"
            title="Сессия истекла"
            text="Войдите снова, чтобы продолжить."
          >
            <Button variant="primary" onClick={() => navigate("/login")}>
              Войти
            </Button>
          </AuthCentered>
        </AuthCard>
      </AuthMain>
      <AuthFooter />
    </AuthLayout>
  );
}
