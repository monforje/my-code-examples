import { useState } from "react";
import { useAuth } from "@features/auth";
import { Avatar, useProfile } from "@entities/user";
import { AppHeader } from "@widgets/header/ui/AppHeader";
import { SettingsLayout } from "@widgets/settings-layout";
import { Modal } from "@shared/ui/Modal";
import { Alert } from "@shared/ui";
import { formatDate } from "@shared/lib/utils";
import { Navigate } from "react-router";
import styles from "./ProfilePage.module.css";

export function SettingsProfilePage() {
  const { user, isLoading } = useAuth();
  const { data: profile, error } = useProfile(!!user);
  const [avatarModalOpen, setAvatarModalOpen] = useState(false);

  if (isLoading) return null;
  if (!user) return <Navigate to="/login" replace />;

  const displayName = profile?.display_name || "Не указано";
  const bio = profile?.bio || "Не указано";
  const email = profile?.email ?? user.email;
  const avatarUrl = profile?.avatar_url;
  const initials = (profile?.display_name || email)[0]?.toUpperCase() ?? "U";

  return (
    <>
      <AppHeader />
      <SettingsLayout
        sidebar="profile"
        title="Профиль"
        description="Основные данные вашего аккаунта Codurity."
      >
        {error && <Alert variant="danger">Не удалось загрузить профиль.</Alert>}
        <div className={styles.accountCard}>
          <div className={styles.header}>Данные аккаунта</div>
          <div className={styles.rowAvatar}>
            <span className={styles.label}>Аватар</span>
            <button
              className={styles.avatarBtn}
              onClick={() => setAvatarModalOpen(true)}
              type="button"
            >
              <Avatar className={styles.avatar} src={avatarUrl} initials={initials} />
            </button>
          </div>
          <div className={styles.row}>
            <span className={styles.label}>Имя</span>
            <span className={styles.value}>{displayName}</span>
          </div>
          <div className={styles.row}>
            <span className={styles.label}>О себе</span>
            <span className={styles.value}>{bio}</span>
          </div>
          <div className={styles.row}>
            <span className={styles.label}>Email</span>
            <span className={styles.value}>{email}</span>
          </div>
          <div className={styles.row}>
            <span className={styles.label}>Дата создания</span>
            <span className={styles.value}>
              {formatDate(profile?.created_at ?? user.created_at)}
            </span>
          </div>
        </div>
      </SettingsLayout>

      {avatarModalOpen && (
        <Modal title="Аватар" onClose={() => setAvatarModalOpen(false)}>
          <Avatar className={styles.avatarLarge} src={avatarUrl} initials={initials} alt="Аватар" />
        </Modal>
      )}
    </>
  );
}
