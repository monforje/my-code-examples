import { useEffect, useRef, useState } from "react";
import { useAuth } from "@features/auth";
import {
  useDeleteAvatar,
  Avatar,
  useProfile,
  useUpdateAvatar,
  useUpdateProfileSettings,
} from "@entities/user";
import { AppHeader } from "@widgets/header/ui/AppHeader";
import { SettingsLayout } from "@widgets/settings-layout";
import { Modal } from "@shared/ui/Modal";
import { Button } from "@shared/ui/Button";
import { Field } from "@shared/ui/Field";
import { Input } from "@shared/ui/Input";
import { Alert } from "@shared/ui";
import { useToast } from "@shared/lib/use-toast";
import { Navigate } from "react-router";
import styles from "./SettingsEditProfile.module.css";

export function SettingsEditProfilePage() {
  const { user, isLoading } = useAuth();
  const { data: profile, error: profileError } = useProfile(!!user);
  const updateSettings = useUpdateProfileSettings();
  const updateAvatar = useUpdateAvatar();
  const deleteAvatar = useDeleteAvatar();
  const { addToast } = useToast();
  const fileInputRef = useRef<HTMLInputElement>(null);
  const [avatarModalOpen, setAvatarModalOpen] = useState(false);
  const [pendingAvatarFile, setPendingAvatarFile] = useState<File | null>(null);
  const [deleteAvatarConfirmOpen, setDeleteAvatarConfirmOpen] = useState(false);
  const [name, setName] = useState("");
  const [bio, setBio] = useState("");
  const [error, setError] = useState("");

  useEffect(() => {
    if (!profile) return;
    setName(profile.display_name ?? "");
    setBio(profile.bio ?? "");
  }, [profile]);

  if (isLoading) return null;
  if (!user) return <Navigate to="/login" replace />;

  const email = profile?.email ?? user.email;
  const avatarUrl = profile?.avatar_url;
  const initials = (name || email)[0]?.toUpperCase() ?? "U";

  const handleSave = async () => {
    setError("");
    try {
      await updateSettings.mutateAsync({
        display_name: name,
        bio,
      });
      addToast("Профиль сохранён", "success");
    } catch (err) {
      setError(err instanceof Error ? err.message : "Не удалось сохранить профиль.");
    }
  };

  const handleAvatarSelect = (file: File | undefined) => {
    if (!file) return;
    setPendingAvatarFile(file);
  };

  const handleUploadAvatar = async () => {
    if (!pendingAvatarFile) return;
    setError("");
    try {
      await updateAvatar.mutateAsync(pendingAvatarFile);
      addToast("Аватар обновлён", "success");
      setPendingAvatarFile(null);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Не удалось загрузить аватар.");
    } finally {
      if (fileInputRef.current) fileInputRef.current.value = "";
    }
  };

  const handleDeleteAvatar = async () => {
    setError("");
    try {
      await deleteAvatar.mutateAsync();
      addToast("Аватар удалён", "success");
      setDeleteAvatarConfirmOpen(false);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Не удалось удалить аватар.");
    }
  };

  return (
    <>
      <AppHeader />
      <SettingsLayout
        sidebar="settings"
        activeTab="profile"
        title="Настройки профиля"
        description="Управление аватаром, именем и информацией о себе."
      >
        {(error || profileError) && (
          <Alert variant="danger">{error || "Не удалось загрузить профиль."}</Alert>
        )}
        <div className={styles.card}>
          <div className={styles.header}>Аватар</div>
          <div className={styles.avatarRow}>
            <button
              className={styles.avatarBtn}
              onClick={() => setAvatarModalOpen(true)}
              type="button"
            >
              <Avatar className={styles.avatar} src={avatarUrl} initials={initials} />
            </button>
            <div className={styles.avatarActions}>
              <div className={styles.avatarText}>
                <div className={styles.avatarTitle}>Изображение профиля</div>
                <div className={styles.avatarHint}>PNG, JPEG или WebP до 5 МБ.</div>
              </div>
              <div className={styles.avatarButtons}>
                <input
                  ref={fileInputRef}
                  type="file"
                  accept="image/png,image/jpeg,image/webp"
                  className={styles.fileInput}
                  onChange={(event) => handleAvatarSelect(event.target.files?.[0])}
                />
                <Button
                  variant="soft"
                  size="sm"
                  className={styles.avatarUploadButton}
                  onClick={() => fileInputRef.current?.click()}
                  loading={updateAvatar.isPending}
                >
                  Загрузить
                </Button>
                {avatarUrl && (
                  <Button
                    variant="ghost"
                    size="sm"
                    className={styles.avatarDeleteButton}
                    onClick={() => setDeleteAvatarConfirmOpen(true)}
                    loading={deleteAvatar.isPending}
                  >
                    Удалить
                  </Button>
                )}
              </div>
            </div>
          </div>
        </div>

        <div className={styles.card}>
          <div className={styles.header}>Личные данные</div>
          <div className={styles.form}>
            <Field label="Имя">
              <Input
                value={name}
                onChange={(e) => setName(e.target.value)}
                placeholder="Ваше имя"
              />
            </Field>
            <Field label="О себе">
              <textarea
                className={styles.textarea}
                value={bio}
                onChange={(e) => setBio(e.target.value)}
                placeholder="Расскажите о себе"
                rows={3}
              />
            </Field>
            <div className={styles.actions}>
              <Button variant="primary" onClick={handleSave} loading={updateSettings.isPending}>
                Сохранить
              </Button>
            </div>
          </div>
        </div>
      </SettingsLayout>

      {avatarModalOpen && (
        <Modal title="Аватар" onClose={() => setAvatarModalOpen(false)}>
          <Avatar className={styles.avatarLarge} src={avatarUrl} initials={initials} alt="Аватар" />
        </Modal>
      )}

      {pendingAvatarFile && (
        <Modal
          title="Загрузить новый аватар?"
          onClose={() => {
            setPendingAvatarFile(null);
            if (fileInputRef.current) fileInputRef.current.value = "";
          }}
        >
          <div className={styles.confirmText}>
            Файл <strong>{pendingAvatarFile.name}</strong> заменит текущий аватар профиля.
          </div>
          <div className={styles.confirmActions}>
            <Button
              variant="ghost"
              onClick={() => {
                setPendingAvatarFile(null);
                if (fileInputRef.current) fileInputRef.current.value = "";
              }}
            >
              Отмена
            </Button>
            <Button variant="primary" onClick={handleUploadAvatar} loading={updateAvatar.isPending}>
              Загрузить
            </Button>
          </div>
        </Modal>
      )}

      {deleteAvatarConfirmOpen && (
        <Modal title="Удалить аватар?" onClose={() => setDeleteAvatarConfirmOpen(false)}>
          <div className={styles.confirmText}>
            Текущий аватар будет удалён. Вместо изображения будет показана инициальная буква.
          </div>
          <div className={styles.confirmActions}>
            <Button variant="ghost" onClick={() => setDeleteAvatarConfirmOpen(false)}>
              Отмена
            </Button>
            <Button variant="danger" onClick={handleDeleteAvatar} loading={deleteAvatar.isPending}>
              Удалить
            </Button>
          </div>
        </Modal>
      )}
    </>
  );
}
