package service

import (
	"context"
	"time"

	"users/internal/authctx"
	postgresrepo "users/internal/repository/postgres"
	apperrors "users/pkg/errors"
)

// UpdateAvatar — загрузка/замена аватара пользователя.
/*
	1. Извлечь identity_id из контекста через authctx.FromContext.
	2. Получить текущий профиль из БД по identity_id.
	3. Если профиль не найден — вернуть ошибку.
	4. Если у профиля уже есть аватар (avatar_object_key != "") — удалить старый файл с диска через AvatarStorage.Delete.
	5. Сохранить новый файл через AvatarStorage.Save.
	6. Обновить avatar_url и avatar_object_key в профиле.
	7. Установить updated_at = now.
	8. Сохранить изменения в БД.
	9. Вернуть UpdateAvatarOutput с avatar_url и updated_at.
*/
func (s *UsersService) UpdateAvatar(ctx context.Context, input *UpdateAvatarInput) (*UpdateAvatarOutput, error) {
	const op = "UsersService.UpdateAvatar"

	// 1. Извлекаем identity_id из auth-контекста.
	identityID, _, err := authctx.FromContext(ctx)
	if err != nil {
		return nil, apperrors.New(op, err)
	}

	// 2. Получаем текущий профиль из БД по identity_id.
	profile, err := s.userProfiles.GetByIdentityID(ctx, identityID)
	if err != nil {
		if err == postgresrepo.ErrUserProfileNotFound {
			return nil, apperrors.New(op, err)
		}
		return nil, apperrors.New(op, err)
	}

	// 3. Если у профиля уже есть аватар — удаляем старый файл.
	if profile.AvatarObjectKey != "" {
		if err := s.avatar.Delete(profile.AvatarObjectKey); err != nil {
			return nil, apperrors.New(op, err)
		}
	}

	// 4. Сохраняем новый файл.
	objectKey, url, err := s.avatar.Save(identityID, input.Filename, input.File)
	if err != nil {
		return nil, apperrors.New(op, err)
	}

	// 5. Обновляем профиль.
	profile.AvatarURL = url
	profile.AvatarObjectKey = objectKey
	profile.UpdatedAt = time.Now().UTC()

	// 6. Сохраняем изменения в БД.
	if err := s.userProfiles.Update(ctx, profile); err != nil {
		return nil, apperrors.New(op, err)
	}

	// 7. Возвращаем результат.
	return &UpdateAvatarOutput{
		AvatarURL: url,
		UpdatedAt: profile.UpdatedAt,
	}, nil
}

// DeleteAvatar — удаление аватара пользователя.
/*
	1. Извлечь identity_id из контекста через authctx.FromContext.
	2. Получить текущий профиль из БД по identity_id.
	3. Если профиль не найден — вернуть ошибку.
	4. Если аватара нет (avatar_object_key == "") — вернуть ошибку.
	5. Удалить файл с диска через AvatarStorage.Delete.
	6. Очистить avatar_url и avatar_object_key в профиле.
	7. Установить updated_at = now.
	8. Сохранить изменения в БД.
	9. Вернуть DeleteAvatarOutput с avatar_url (nil) и updated_at.
*/
func (s *UsersService) DeleteAvatar(ctx context.Context) (*DeleteAvatarOutput, error) {
	const op = "UsersService.DeleteAvatar"

	// 1. Извлекаем identity_id из auth-контекста.
	identityID, _, err := authctx.FromContext(ctx)
	if err != nil {
		return nil, apperrors.New(op, err)
	}

	// 2. Получаем текущий профиль из БД по identity_id.
	profile, err := s.userProfiles.GetByIdentityID(ctx, identityID)
	if err != nil {
		if err == postgresrepo.ErrUserProfileNotFound {
			return nil, apperrors.New(op, err)
		}
		return nil, apperrors.New(op, err)
	}

	// 3. Проверяем, что аватар существует.
	if profile.AvatarObjectKey == "" {
		return nil, apperrors.New(op, ErrAvatarNotFound)
	}

	// 4. Удаляем файл с диска.
	if err := s.avatar.Delete(profile.AvatarObjectKey); err != nil {
		return nil, apperrors.New(op, err)
	}

	// 5. Очищаем аватар в профиле.
	profile.AvatarURL = ""
	profile.AvatarObjectKey = ""
	profile.UpdatedAt = time.Now().UTC()

	// 6. Сохраняем изменения в БД.
	if err := s.userProfiles.Update(ctx, profile); err != nil {
		return nil, apperrors.New(op, err)
	}

	// 7. Возвращаем результат.
	return &DeleteAvatarOutput{
		AvatarURL: nil,
		UpdatedAt: profile.UpdatedAt,
	}, nil
}
