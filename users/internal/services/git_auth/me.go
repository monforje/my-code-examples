package gitauthservice

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/google/uuid"

	postgresrepo "users/internal/repository/postgres"
	apperrors "users/pkg/errors"
)

// GetGitMe - получить информацию о текущем пользователе, связанном с git-аккаунтом
/*
	1. Получить профиль по identityID.
	2. Получить git-аккаунты по profileID из профиля.
	3. Если git-аккаунтов нет — вернуть ошибку.
	4. Вернуть GitMeResponse с данными из профиля и git-аккаунта.
*/
func (s *GitAuthService) GetGitMe(ctx context.Context, identityID uuid.UUID) (*GitMeResponse, error) {
	const op = "GitAuthService.GetGitMe"

	// 1. Получаем профиль по identityID.
	profile, err := s.userProfiles.GetByIdentityID(ctx, identityID)
	if err != nil {
		s.log.Error(ctx, op, "get profile by identity_id failed",
			slog.String("identity_id", identityID.String()),
			"error", err,
		)
		return nil, apperrors.New(op, fmt.Errorf("get profile: %w", err))
	}

	// 2. Получаем git-аккаунты по profileID.
	gitUsers, err := s.gitUsers.GetByProfileID(ctx, profile.ID)
	if err != nil {
		s.log.Error(ctx, op, "get git users by profile_id failed",
			slog.String("profile_id", profile.ID.String()),
			"error", err,
		)
		return nil, apperrors.New(op, fmt.Errorf("get git users: %w", err))
	}

	if len(gitUsers) == 0 {
		s.log.Error(ctx, op, "no git users found for profile",
			slog.String("profile_id", profile.ID.String()),
		)
		return nil, apperrors.New(op, postgresrepo.ErrGitUserNotFound)
	}

	// 3. Берём первый git-аккаунт.
	gitUser := gitUsers[0]

	s.log.Info(ctx, op, "git me resolved",
		slog.String("identity_id", identityID.String()),
		slog.String("profile_id", profile.ID.String()),
		slog.String("git_user_id", gitUser.ID.String()),
	)

	// 4. Возвращаем ответ.
	username := profile.DisplayName
	if username == "" {
		username = profile.Email
	}

	return &GitMeResponse{
		Username: username,
		GitToken: gitUser.GitToken,
		GitURL:   gitUser.GitURL,
	}, nil
}
