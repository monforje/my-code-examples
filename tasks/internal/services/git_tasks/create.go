package gittasksservice

import (
	"context"
	"errors"
	"net/url"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"tasks/internal/authctx"
	"tasks/internal/models/records"
	clientsdto "tasks/pkg/http_clients/dto"
)

// CreateGitTask creates or returns a user's prepared git repository for a task.
func (s *GitTasksService) CreateGitTask(ctx context.Context, input *GitTaskCreateInput) (*GitTaskCreateOutput, error) {
	identityID, _, err := authctx.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	task, err := s.tasksRepo.GetByTaskName(ctx, input.TaskName)
	if err != nil {
		return nil, ErrTaskNotFound
	}

	existing, err := s.pulledRepo.GetByTaskIDAndIdentityID(ctx, task.ID, identityID)
	if err == nil {
		return &GitTaskCreateOutput{TaskName: task.TaskName, Repo: existing.Repo, CloneURL: existing.CloneURL}, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return nil, err
	}

	gitUser, err := s.users.GetGitUser(ctx, &clientsdto.GitUserRequest{IdentityID: identityID})
	if err != nil {
		return nil, err
	}

	created, err := s.gitTasks.GitTaskCreate(ctx, &clientsdto.GitTaskCreateRequest{
		Username: gitUser.Username,
		TaskID:   task.TaskName,
	})
	if err != nil {
		return nil, err
	}

	cloneURL, err := cloneURLWithCredentials(created.CloneURL, gitUser.Username, gitUser.GitToken)
	if err != nil {
		return nil, err
	}

	pulled := &records.PulledTask{
		ID:         uuid.New(),
		IdentityID: identityID,
		TaskID:     task.ID,
		Repo:       created.Repo,
		CloneURL:   cloneURL,
		CreatedAt:  nowFunc(),
	}
	if err := s.pulledRepo.Create(ctx, pulled); err != nil {
		return nil, err
	}

	return &GitTaskCreateOutput{TaskName: task.TaskName, Repo: created.Repo, CloneURL: cloneURL}, nil
}

func cloneURLWithCredentials(rawURL, username, token string) (string, error) {
	parsed, err := url.Parse(rawURL)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return "", errors.Join(ErrCloneURLInvalid, err)
	}
	parsed.User = url.UserPassword(username, token)
	return parsed.String(), nil
}
