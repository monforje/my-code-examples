package gitauthservice

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"

	"users/internal/models/records"
	apperrors "users/pkg/errors"
	clientsdto "users/pkg/http_clients/dto"
)

// RegisterGitUser - регистрирует нового пользователя Git в системе.
/*
	1. Вызвать Git Auth HTTP client с username и email.
	2. Получить git token и git url из ответа.
	3. Создать records.GitUser с привязкой к profile ID.
	4. Сохранить Git пользователя через репозиторий.
	5. Вернуть ID созданной записи.
*/
func (s *GitAuthService) RegisterGitUser(ctx context.Context, input *RegisterGitUserInput) (uuid.UUID, error) {
	const op = "GitAuthService.RegisterGitUser"

	s.log.Info(ctx, op, "calling external git auth service",
		slog.String("profile_id", input.ProfileID.String()),
		slog.String("username", input.Username),
		slog.String("email", input.Email),
	)

	// 1. Регистрируем пользователя во внешнем Git Auth.
	resp, err := s.client.RegisterGitUser(ctx, &clientsdto.RegisterGitUserRequest{
		Username: input.Username,
		Email:    input.Email,
	})
	if err != nil {
		s.log.Error(ctx, op, "external git auth service call failed",
			slog.String("profile_id", input.ProfileID.String()),
			"error", err,
		)
		return uuid.Nil, apperrors.New(op, err)
	}
	if resp == nil || resp.Token == "" || resp.GitURL == "" {
		s.log.Error(ctx, op, "empty response from git auth service",
			slog.String("profile_id", input.ProfileID.String()),
		)
		return uuid.Nil, apperrors.New(op, fmt.Errorf("empty git auth response"))
	}

	s.log.Info(ctx, op, "external git auth service responded",
		slog.String("profile_id", input.ProfileID.String()),
		slog.String("git_url", resp.GitURL),
	)

	// 2. Сохраняем токен и Git URL в БД.
	now := time.Now().UTC()
	gitUser := &records.GitUser{
		ID:        uuid.New(),
		ProfileID: input.ProfileID,
		GitToken:  resp.Token,
		GitURL:    resp.GitURL,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := s.gitUsers.Create(ctx, gitUser); err != nil {
		s.log.Error(ctx, op, "save git user to database failed",
			slog.String("profile_id", input.ProfileID.String()),
			"error", err,
		)
		return uuid.Nil, apperrors.New(op, err)
	}

	s.log.Info(ctx, op, "git user saved to database",
		slog.String("profile_id", input.ProfileID.String()),
		slog.String("git_user_id", gitUser.ID.String()),
	)

	// 3. Возвращаем ID записи Git пользователя.
	return gitUser.ID, nil
}
