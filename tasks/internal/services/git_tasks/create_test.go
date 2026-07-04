package gittasksservice_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"go.uber.org/mock/gomock"

	"tasks/internal/authctx"
	"tasks/internal/models/records"
	gittasksservice "tasks/internal/services/git_tasks"
	"tasks/internal/services/git_tasks/mocks"
	clientsdto "tasks/pkg/http_clients/dto"
)

func TestGitTasksService_CreateGitTask_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tasksRepo := mocks.NewMockTaskRepository(ctrl)
	pulledRepo := mocks.NewMockPulledTaskRepository(ctrl)
	users := mocks.NewMockUsersClient(ctrl)
	gitTasks := mocks.NewMockGitTasksClient(ctrl)
	svc := gittasksservice.NewGitTasksService(tasksRepo, pulledRepo, users, gitTasks)

	now := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
	gittasksservice.SetNowFunc(func() time.Time { return now })
	defer gittasksservice.SetNowFunc(nil)

	identityID := uuid.New()
	sessionID := uuid.New()
	taskID := uuid.New()
	ctx := authctx.WithAuth(context.Background(), identityID, sessionID)

	tasksRepo.EXPECT().GetByTaskName(gomock.Any(), "pizza-api").Return(&records.Task{ID: taskID, TaskName: "pizza-api"}, nil)
	pulledRepo.EXPECT().GetByTaskIDAndIdentityID(gomock.Any(), taskID, identityID).Return(nil, pgx.ErrNoRows)
	users.EXPECT().GetGitUser(gomock.Any(), &clientsdto.GitUserRequest{IdentityID: identityID}).Return(&clientsdto.GitUserResponse{
		Username: "alice",
		GitToken: "git-token",
	}, nil)
	gitTasks.EXPECT().GitTaskCreate(gomock.Any(), &clientsdto.GitTaskCreateRequest{
		Username: "alice",
		TaskID:   "pizza-api",
	}).Return(&clientsdto.GitTaskCreateResponse{
		TaskID:   "pizza-api",
		Repo:     "alice/golden-pizza-api",
		CloneURL: "http://gitea.local/alice/golden-pizza-api.git",
	}, nil)
	pulledRepo.EXPECT().Create(gomock.Any(), gomock.Any()).DoAndReturn(func(_ context.Context, pulled *records.PulledTask) error {
		if pulled.IdentityID != identityID {
			t.Fatalf("IdentityID = %s, want %s", pulled.IdentityID, identityID)
		}
		if pulled.TaskID != taskID {
			t.Fatalf("TaskID = %s, want %s", pulled.TaskID, taskID)
		}
		if pulled.Repo != "alice/golden-pizza-api" {
			t.Fatalf("Repo = %q", pulled.Repo)
		}
		if pulled.CloneURL != "http://alice:git-token@gitea.local/alice/golden-pizza-api.git" {
			t.Fatalf("CloneURL = %q", pulled.CloneURL)
		}
		if !pulled.CreatedAt.Equal(now) {
			t.Fatalf("CreatedAt = %s, want %s", pulled.CreatedAt, now)
		}
		return nil
	})

	output, err := svc.CreateGitTask(ctx, &gittasksservice.GitTaskCreateInput{TaskName: "pizza-api"})
	if err != nil {
		t.Fatalf("CreateGitTask() error = %v", err)
	}
	if output.TaskName != "pizza-api" {
		t.Fatalf("TaskName = %q", output.TaskName)
	}
	if output.CloneURL != "http://alice:git-token@gitea.local/alice/golden-pizza-api.git" {
		t.Fatalf("CloneURL = %q", output.CloneURL)
	}
}

func TestGitTasksService_CreateGitTask_ExistingPulledTask(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tasksRepo := mocks.NewMockTaskRepository(ctrl)
	pulledRepo := mocks.NewMockPulledTaskRepository(ctrl)
	svc := gittasksservice.NewGitTasksService(tasksRepo, pulledRepo, mocks.NewMockUsersClient(ctrl), mocks.NewMockGitTasksClient(ctrl))

	identityID := uuid.New()
	taskID := uuid.New()
	ctx := authctx.WithAuth(context.Background(), identityID, uuid.New())

	tasksRepo.EXPECT().GetByTaskName(gomock.Any(), "pizza-api").Return(&records.Task{ID: taskID, TaskName: "pizza-api"}, nil)
	pulledRepo.EXPECT().GetByTaskIDAndIdentityID(gomock.Any(), taskID, identityID).Return(&records.PulledTask{
		TaskID:   taskID,
		Repo:     "alice/golden-pizza-api",
		CloneURL: "http://alice:token@gitea.local/alice/golden-pizza-api.git",
	}, nil)

	output, err := svc.CreateGitTask(ctx, &gittasksservice.GitTaskCreateInput{TaskName: "pizza-api"})
	if err != nil {
		t.Fatalf("CreateGitTask() error = %v", err)
	}
	if output.Repo != "alice/golden-pizza-api" {
		t.Fatalf("Repo = %q", output.Repo)
	}
}

func TestGitTasksService_CreateGitTask_MissingAuthContext(t *testing.T) {
	svc := gittasksservice.NewGitTasksService(nil, nil, nil, nil)

	_, err := svc.CreateGitTask(context.Background(), &gittasksservice.GitTaskCreateInput{TaskName: "pizza-api"})
	if !errors.Is(err, authctx.ErrAuthContextMissing) {
		t.Fatalf("CreateGitTask() error = %v, want %v", err, authctx.ErrAuthContextMissing)
	}
}

func TestGitTasksService_CreateGitTask_TaskNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tasksRepo := mocks.NewMockTaskRepository(ctrl)
	svc := gittasksservice.NewGitTasksService(tasksRepo, nil, nil, nil)
	ctx := authctx.WithAuth(context.Background(), uuid.New(), uuid.New())

	tasksRepo.EXPECT().GetByTaskName(gomock.Any(), "missing").Return(nil, errors.New("not found"))

	_, err := svc.CreateGitTask(ctx, &gittasksservice.GitTaskCreateInput{TaskName: "missing"})
	if !errors.Is(err, gittasksservice.ErrTaskNotFound) {
		t.Fatalf("CreateGitTask() error = %v, want %v", err, gittasksservice.ErrTaskNotFound)
	}
}
