package service

import (
	"context"
	"time"

	"users/internal/authctx"
	postgresrepo "users/internal/repository/postgres"
	apperrors "users/pkg/errors"
)

// UpdateSettings — обновление настроек профиля (display_name, bio).
/*
	1. Извлечь identity_id из контекста через authctx.FromContext.
	2. Получить текущий профиль из БД по identity_id.
	3. Если профиль не найден — вернуть ошибку.
	4. Если передан DisplayName — обновить его в профиле.
	5. Если передан Bio — обновить его в профиле.
	6. Установить updated_at = now.
	7. Сохранить изменения в БД через UserProfiles.Update.
	8. Смаппить records.UserProfile → ProfileOutput.
	9. Вернуть ProfileOutput.
*/
func (s *UsersService) UpdateSettings(ctx context.Context, input *UpdateSettingsInput) (*ProfileOutput, error) {
	const op = "UsersService.UpdateSettings"

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

	// 3. Обновляем переданные поля.
	if input.DisplayName != nil {
		profile.DisplayName = *input.DisplayName
	}
	if input.Bio != nil {
		profile.BIO = *input.Bio
	}
	profile.UpdatedAt = time.Now().UTC()

	// 4. Сохраняем изменения в БД.
	if err := s.userProfiles.Update(ctx, profile); err != nil {
		return nil, apperrors.New(op, err)
	}

	// 5. Смаппим record в сервисный ProfileOutput DTO.
	// 6. Возвращаем обновлённый профиль.
	return &ProfileOutput{
		ID:            profile.ID.String(),
		IdentityID:    profile.IdentityID.String(),
		Email:         profile.Email,
		DisplayName:   strPtr(profile.DisplayName),
		Bio:           profile.BIO,
		AvatarURL:     strPtr(profile.AvatarURL),
		Status:        profile.Status,
		EmailVerified: profile.EmailVerified,
		CreatedAt:     profile.CreatedAt,
		UpdatedAt:     profile.UpdatedAt,
	}, nil
}
