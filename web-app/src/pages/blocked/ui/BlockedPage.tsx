import { AuthLayout, AuthMain, AuthCard, AuthCentered, AuthFooter } from "@features/auth";
import { PublicHeader } from "@widgets/header";

export function BlockedPage() {
  return (
    <AuthLayout>
      <PublicHeader />
      <AuthMain>
        <AuthCard title="">
          <AuthCentered
            icon="tabler:shield-lock"
            iconVariant="danger"
            title="Аккаунт заблокирован"
            text="Доступ к аккаунту ограничен. Обратитесь в поддержку."
          >
            <a className="cd-button cd-button-primary" href="#">
              Связаться с поддержкой
            </a>
          </AuthCentered>
        </AuthCard>
      </AuthMain>
      <AuthFooter />
    </AuthLayout>
  );
}
