import { useNavigate } from "react-router";
import { AuthLayout, AuthMain, AuthCard, AuthCentered, AuthFooter } from "@features/auth";
import { PublicHeader } from "@widgets/header";
import { Button } from "@shared/ui";

export function NotFoundPage() {
  const navigate = useNavigate();

  return (
    <AuthLayout>
      <PublicHeader />
      <AuthMain>
        <AuthCard title="">
          <AuthCentered
            icon="tabler:question-mark-circle"
            iconVariant="info"
            title="Страница не найдена"
            text="Такой страницы не существует или она была перемещена."
          >
            <Button variant="primary" onClick={() => navigate("/")}>
              На главную
            </Button>
          </AuthCentered>
        </AuthCard>
      </AuthMain>
      <AuthFooter />
    </AuthLayout>
  );
}
