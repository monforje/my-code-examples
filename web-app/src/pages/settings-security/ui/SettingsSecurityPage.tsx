import { useState } from "react";
import { Navigate, useNavigate } from "react-router";
import { useAuth } from "@features/auth";
import { useToast } from "@shared/lib/use-toast";
import { AppHeader } from "@widgets/header/ui/AppHeader";
import { SettingsLayout } from "@widgets/settings-layout";
import { Button } from "@shared/ui";
import { ChangePasswordModal } from "@features/password";
import { ChangeEmailModal } from "@features/email";
import { DeleteAccountModal } from "@features/delete-account";
import styles from "./SecurityPage.module.css";

export function SettingsSecurityPage() {
  const navigate = useNavigate();
  const { user, isLoading, logout } = useAuth();
  const { addToast } = useToast();
  const [changePasswordOpen, setChangePasswordOpen] = useState(false);
  const [changeEmailOpen, setChangeEmailOpen] = useState(false);
  const [deleteAccountOpen, setDeleteAccountOpen] = useState(false);

  if (isLoading) return null;
  if (!user) return <Navigate to="/login" replace />;

  const handleLogout = async () => {
    await logout();
    addToast("Вы вышли из аккаунта", "info");
    navigate("/login");
  };

  return (
    <>
      <AppHeader />
      <SettingsLayout
        sidebar="settings"
        activeTab="security"
        title="Безопасность"
        description="Управление паролем, email и сессиями."
      >
        <div className={styles.securityStack}>
          <div className={styles.securityCard}>
            <div className={styles.securityInfo}>
              <div className={styles.securityTitle}>Пароль</div>
              <div className={styles.securityDesc}>Обновите пароль для защиты аккаунта.</div>
            </div>
            <Button variant="soft" onClick={() => setChangePasswordOpen(true)}>
              Сменить пароль
            </Button>
          </div>

          <div className={styles.securityCard}>
            <div className={styles.securityInfo}>
              <div className={styles.securityTitle}>Email</div>
              <div className={styles.securityDesc}>
                Email используется для входа и security-кодов.
              </div>
            </div>
            <Button variant="soft" onClick={() => setChangeEmailOpen(true)}>
              Сменить email
            </Button>
          </div>

          <div className={styles.securityCard}>
            <div className={styles.securityInfo}>
              <div className={styles.securityTitle}>Сессия</div>
              <div className={styles.securityDesc}>
                Завершите текущую сессию на этом устройстве.
              </div>
            </div>
            <Button onClick={handleLogout}>Выйти</Button>
          </div>

          <div className={styles.dangerZone}>
            <div className={styles.dangerTitle}>Удаление аккаунта</div>
            <div className={styles.dangerDesc}>Удаление аккаунта необратимо.</div>
            <div>
              <Button variant="danger" onClick={() => setDeleteAccountOpen(true)}>
                Удалить аккаунт
              </Button>
            </div>
          </div>
        </div>

        {changePasswordOpen && <ChangePasswordModal onClose={() => setChangePasswordOpen(false)} />}
        {changeEmailOpen && <ChangeEmailModal onClose={() => setChangeEmailOpen(false)} />}
        {deleteAccountOpen && <DeleteAccountModal onClose={() => setDeleteAccountOpen(false)} />}
      </SettingsLayout>
    </>
  );
}
