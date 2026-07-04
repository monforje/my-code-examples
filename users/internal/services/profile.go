package service

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"

	"users/internal/authctx"
	"users/internal/generate/avatar"
	"users/internal/generate/username"
	"users/internal/models/records"
	postgresrepo "users/internal/repository/postgres"
	gitauthservice "users/internal/services/git_auth"
	apperrors "users/pkg/errors"
)

// GetProfile — получение профиля текущего пользователя.
/*
	1. Извлечь identity_id из контекста через authctx.FromContext.
	2. Получить профиль из БД по identity_id через UserProfiles.GetByIdentityID.
	3. Если профиль не найден — вернуть ошибку.
	4. Смаппить records.UserProfile → ProfileOutput.
	5. Вернуть ProfileOutput.
*/
func (s *UsersService) GetProfile(ctx context.Context) (*ProfileOutput, error) {
	const op = "UsersService.GetProfile"

	// 1. Извлекаем identity_id из auth-контекста.
	identityID, _, err := authctx.FromContext(ctx)
	if err != nil {
		return nil, apperrors.New(op, err)
	}

	// 2. Получаем профиль из БД по identity_id.
	profile, err := s.userProfiles.GetByIdentityID(ctx, identityID)
	if err != nil {
		if err == postgresrepo.ErrUserProfileNotFound {
			return nil, apperrors.New(op, err)
		}
		return nil, apperrors.New(op, err)
	}

	// 3. Смаппим record в сервисный ProfileOutput DTO.
	// 4. Возвращаем данные профиля.
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

// HandleIdentityCreated — создание профиля при регистрации.
/*
	1. Распарсить identityID из строки в uuid.
	2. Проверить, существует ли уже профиль по identity_id (GetByIdentityID).
	3. Если профиль уже существует — вернуть nil (идемпотентность).
	4. Если GetByIdentityID вернул ошибку, отличную от ErrUserProfileNotFound — вернуть ошибку.
	5. Сгенерировать уникальный display_name.
	6. Сгенерировать SVG-аватар и сохранить на диск.
	7. Создать records.UserProfile со всеми полями.
	8. Сохранить через UserProfiles.Create.
	9. Асинхронно зарегистрировать Git пользователя и залогировать ошибку, если она возникла.
	10. Вернуть nil.
*/
func (s *UsersService) HandleIdentityCreated(ctx context.Context, input *HandleIdentityCreatedInput) error {
	const op = "UsersService.HandleIdentityCreated"

	// 1. Парсим identityID из строки.
	identityID, err := uuid.Parse(input.IdentityID)
	if err != nil {
		return apperrors.New(op, err)
	}

	// 2. Проверяем, существует ли уже профиль.
	existing, err := s.userProfiles.GetByIdentityID(ctx, identityID)
	if err == nil && existing != nil {
		// 3. Профиль уже существует — идемпотентность, выходим.
		return nil
	}
	// 4. Если ошибка не "не найдено" — возвращаем ошибку.
	if err != postgresrepo.ErrUserProfileNotFound {
		return apperrors.New(op, err)
	}

	// 5. Генерируем уникальный display_name.
	displayName, err := username.GenerateUniqueUsername(ctx, s.userProfiles.ExistsByDisplayName)
	if err != nil {
		return apperrors.New(op, fmt.Errorf("generate display name: %w", err))
	}

	// 6. Генерируем PNG-аватар и сохраняем на диск.
	var buf bytes.Buffer
	identicon := avatar.Generate(identityID.String())
	if err := identicon.Render(&buf, 420); err != nil {
		return apperrors.New(op, fmt.Errorf("render avatar: %w", err))
	}
	objectKey, avatarURL, err := s.avatar.Save(
		identityID,
		fmt.Sprintf("avatar-%s.png", identityID.String()),
		&buf,
	)
	if err != nil {
		return apperrors.New(op, fmt.Errorf("save avatar: %w", err))
	}

	// 7. Создаём новый профиль.
	now := time.Now().UTC()
	profile := &records.UserProfile{
		ID:              uuid.New(),
		IdentityID:      identityID,
		Email:           input.Email,
		DisplayName:     displayName,
		AvatarURL:       avatarURL,
		AvatarObjectKey: objectKey,
		Status:          "active",
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	// 8. Сохраняем в БД.
	if err := s.userProfiles.Create(ctx, profile); err != nil {
		return apperrors.New(op, err)
	}

	// 9. Регистрируем Git пользователя асинхронно, не блокируя обработку события.
	if s.gitAuth != nil {
		go func(profileID uuid.UUID, email, displayName string) {
			bgCtx := context.Background()
			s.log.Info(bgCtx, op, "registering git user",
				slog.String("profile_id", profileID.String()),
				slog.String("email", email),
			)
			gitUserID, err := s.gitAuth.RegisterGitUser(bgCtx, &gitauthservice.RegisterGitUserInput{
				ProfileID: profileID,
				Username:  displayName,
				Email:     email,
			})
			if err != nil {
				s.log.Error(bgCtx, op, "register git user",
					slog.String("profile_id", profileID.String()),
					"error", err,
				)
				return
			}
			s.log.Info(bgCtx, op, "git user registered",
				slog.String("profile_id", profileID.String()),
				slog.String("git_user_id", gitUserID.String()),
			)
		}(profile.ID, profile.Email, profile.DisplayName)
	} else {
		s.log.Warn(context.Background(), op, "git auth service not configured, skipping git user registration",
			slog.String("profile_id", profile.ID.String()),
		)
	}

	// 10. Готово.
	return nil
}

// HandleIdentityUpdated — обновление профиля при изменении данных в auth.
/*
	1. Распарсить identityID из строки в uuid.
	2. Получить профиль по identity_id (GetByIdentityID).
	3. Если профиль не найден — вернуть nil (профиль ещё не создан).
	4. Если GetByIdentityID вернул ошибку, отличную от ErrUserProfileNotFound — вернуть ошибку.
	5. Если передан Email — обновить.
	6. Если передан Status — обновить.
	7. Если передан EmailVerified — обновить.
	8. Установить updated_at = now.
	9. Сохранить через UserProfiles.Update.
	10. Вернуть nil.
*/
func (s *UsersService) HandleIdentityUpdated(ctx context.Context, input *HandleIdentityUpdatedInput) error {
	const op = "UsersService.HandleIdentityUpdated"

	// 1. Парсим identityID из строки.
	identityID, err := uuid.Parse(input.IdentityID)
	if err != nil {
		return apperrors.New(op, err)
	}

	// 2. Получаем текущий профиль.
	profile, err := s.userProfiles.GetByIdentityID(ctx, identityID)
	if err == postgresrepo.ErrUserProfileNotFound {
		// 3. Профиль не найден — ещё не создан, выходим.
		return nil
	}
	if err != nil {
		// 4. Другая ошибка — возвращаем.
		return apperrors.New(op, err)
	}

	// 5. Обновляем переданные поля.
	changed := false

	if input.Email != nil && *input.Email != "" {
		profile.Email = *input.Email
		changed = true
	}
	if input.Status != nil && *input.Status != "" {
		profile.Status = *input.Status
		changed = true
	}
	if input.EmailVerified != nil {
		profile.EmailVerified = *input.EmailVerified
		changed = true
	}

	if !changed {
		// Ничего не изменилось — выходим.
		return nil
	}

	// 8. Устанавливаем updated_at.
	profile.UpdatedAt = time.Now().UTC()

	// 9. Сохраняем изменения.
	if err := s.userProfiles.Update(ctx, profile); err != nil {
		return apperrors.New(op, err)
	}

	// 10. Готово.
	return nil
}

// HandleIdentityDeleted — soft-delete профиля при удалении аккаунта.
/*
	1. Распарсить identityID из строки в uuid.
	2. Получить профиль по identity_id (GetByIdentityID).
	3. Если профиль не найден — вернуть nil (идемпотентность).
	4. Если GetByIdentityID вернул ошибку, отличную от ErrUserProfileNotFound — вернуть ошибку.
	5. Если аватар был — удалить файл через AvatarStorage.Delete.
	6. Выполнить UserProfiles.Delete (soft delete: SET deleted_at = NOW()).
	7. Вернуть nil.
*/
func (s *UsersService) HandleIdentityDeleted(ctx context.Context, input *HandleIdentityDeletedInput) error {
	const op = "UsersService.HandleIdentityDeleted"

	// 1. Парсим identityID из строки.
	identityID, err := uuid.Parse(input.IdentityID)
	if err != nil {
		return apperrors.New(op, err)
	}

	// 2. Получаем текущий профиль.
	profile, err := s.userProfiles.GetByIdentityID(ctx, identityID)
	if err == postgresrepo.ErrUserProfileNotFound {
		// 3. Профиль не найден — идемпотентность, выходим.
		return nil
	}
	if err != nil {
		// 4. Другая ошибка — возвращаем.
		return apperrors.New(op, err)
	}

	// 5. Удаляем файл аватара, если был.
	if profile.AvatarObjectKey != "" {
		if err := s.avatar.Delete(profile.AvatarObjectKey); err != nil {
			return apperrors.New(op, err)
		}
	}

	// 6. Soft delete профиля.
	if err := s.userProfiles.Delete(ctx, profile.ID); err != nil {
		return apperrors.New(op, err)
	}

	// 7. Готово.
	return nil
}

func strPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

// GetGitMe — получить информацию о git-пользователе по identity_id.
/*
	1. Вызвать GitAuthService.GetGitMe с identityID.
	2. Смаппить gitauthservice.GitMeResponse → GitMeOutput.
	3. Вернуть GitMeOutput.
*/
func (s *UsersService) GetGitMe(ctx context.Context, identityID uuid.UUID) (*GitMeOutput, error) {
	const op = "UsersService.GetGitMe"

	resp, err := s.gitAuth.GetGitMe(ctx, identityID)
	if err != nil {
		return nil, apperrors.New(op, err)
	}

	return &GitMeOutput{
		Username: resp.Username,
		GitToken: resp.GitToken,
		GitURL:   resp.GitURL,
	}, nil
}
