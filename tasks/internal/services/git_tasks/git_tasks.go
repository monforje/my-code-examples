// Package gittasksservice
package gittasksservice

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"

	"tasks/internal/models/records"
	clientsdto "tasks/pkg/http_clients/dto"
)

var (
	ErrTaskNotFound    = errors.New("task not found")
	ErrCloneURLInvalid = errors.New("clone url is invalid")
)

type TaskRepository interface {
	GetByTaskName(ctx context.Context, taskName string) (*records.Task, error)
}

type PulledTaskRepository interface {
	Create(ctx context.Context, task *records.PulledTask) error
	GetByTaskIDAndIdentityID(ctx context.Context, taskID, identityID uuid.UUID) (*records.PulledTask, error)
}

type UsersClient interface {
	GetGitUser(ctx context.Context, req *clientsdto.GitUserRequest) (*clientsdto.GitUserResponse, error)
}

type GitTasksClient interface {
	GitTaskCreate(ctx context.Context, req *clientsdto.GitTaskCreateRequest) (*clientsdto.GitTaskCreateResponse, error)
}

type GitTasksService struct {
	tasksRepo  TaskRepository
	pulledRepo PulledTaskRepository
	users      UsersClient
	gitTasks   GitTasksClient
}

func NewGitTasksService(
	tasksRepo TaskRepository,
	pulledRepo PulledTaskRepository,
	users UsersClient,
	gitTasks GitTasksClient,
) *GitTasksService {
	return &GitTasksService{
		tasksRepo:  tasksRepo,
		pulledRepo: pulledRepo,
		users:      users,
		gitTasks:   gitTasks,
	}
}

var nowFunc = func() time.Time { return time.Now() }

func SetNowFunc(fn func() time.Time) {
	if fn == nil {
		nowFunc = func() time.Time { return time.Now() }
		return
	}
	nowFunc = fn
}
