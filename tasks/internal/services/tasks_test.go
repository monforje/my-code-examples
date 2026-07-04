package services_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"go.uber.org/mock/gomock"

	"tasks/internal/models/records"
	"tasks/internal/services"
	"tasks/internal/services/mocks"
)

func TestTasksService_Create_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockTaskRepository(ctrl)
	svc := services.NewTasksService(repo)

	now := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	services.SetNowFunc(func() time.Time { return now })
	defer services.SetNowFunc(nil)

	repo.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil)
	repo.EXPECT().GetTagsByTaskID(gomock.Any(), gomock.Any()).Return(nil, nil)
	repo.EXPECT().GetLanguagesByTaskID(gomock.Any(), gomock.Any()).Return(nil, nil)

	output, err := svc.Create(context.Background(), &services.CreateInput{
		Title:               "Test Task",
		SpecificationMDText: "# Spec",
		TaskType:            "backend",
		Level:               "middle",
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if output.ID == uuid.Nil {
		t.Error("ID should not be nil")
	}
	if output.Title != "Test Task" {
		t.Errorf("Title = %q, want %q", output.Title, "Test Task")
	}
	if output.SpecificationMDText != "# Spec" {
		t.Errorf("SpecificationMDText = %q, want %q", output.SpecificationMDText, "# Spec")
	}
	if output.TaskType != "backend" {
		t.Errorf("TaskType = %q, want %q", output.TaskType, "backend")
	}
	if output.Level != "middle" {
		t.Errorf("Level = %q, want %q", output.Level, "middle")
	}
}

func TestTasksService_Create_EmptyTitle(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockTaskRepository(ctrl)
	svc := services.NewTasksService(repo)

	_, err := svc.Create(context.Background(), &services.CreateInput{
		Title: "",
	})
	if !errors.Is(err, services.ErrTitleEmpty) {
		t.Fatalf("Create() error = %v, want %v", err, services.ErrTitleEmpty)
	}
}

func TestTasksService_Create_RepoError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	wantErr := errors.New("db error")
	repo := mocks.NewMockTaskRepository(ctrl)
	svc := services.NewTasksService(repo)

	repo.EXPECT().Create(gomock.Any(), gomock.Any()).Return(wantErr)

	_, err := svc.Create(context.Background(), &services.CreateInput{
		Title: "Task",
	})
	if !errors.Is(err, wantErr) {
		t.Fatalf("Create() error = %v, want %v", err, wantErr)
	}
}

func TestTasksService_GetByID_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	id := uuid.New()
	now := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	repo := mocks.NewMockTaskRepository(ctrl)
	svc := services.NewTasksService(repo)

	repo.EXPECT().GetByID(gomock.Any(), id).Return(&records.Task{
		ID:                  id,
		Title:               "Found",
		SpecificationMDText: "spec",
		TaskType:            "backend",
		Level:               "middle",
		CreatedAt:           now,
	}, nil)
	repo.EXPECT().GetTagsByTaskID(gomock.Any(), id).Return(nil, nil)
	repo.EXPECT().GetLanguagesByTaskID(gomock.Any(), id).Return(nil, nil)

	output, err := svc.GetByID(context.Background(), id)
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}
	if output.Title != "Found" {
		t.Errorf("Title = %q, want %q", output.Title, "Found")
	}
}

func TestTasksService_GetByID_NotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockTaskRepository(ctrl)
	svc := services.NewTasksService(repo)

	repo.EXPECT().GetByID(gomock.Any(), gomock.Any()).Return(nil, errors.New("not found"))

	_, err := svc.GetByID(context.Background(), uuid.New())
	if !errors.Is(err, services.ErrTaskNotFound) {
		t.Fatalf("GetByID() error = %v, want %v", err, services.ErrTaskNotFound)
	}
}

func TestTasksService_List_DefaultLimit(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockTaskRepository(ctrl)
	svc := services.NewTasksService(repo)

	repo.EXPECT().List(gomock.Any(), int32(20), (*string)(nil), gomock.Any()).Return(nil, false, nil)

	_, err := svc.List(context.Background(), &services.ListInput{Limit: 0})
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
}

func TestTasksService_List_MaxLimit(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockTaskRepository(ctrl)
	svc := services.NewTasksService(repo)

	repo.EXPECT().List(gomock.Any(), int32(100), (*string)(nil), gomock.Any()).Return(nil, false, nil)

	_, err := svc.List(context.Background(), &services.ListInput{Limit: 200})
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
}

func TestTasksService_List_WithCursor(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockTaskRepository(ctrl)
	svc := services.NewTasksService(repo)

	cursor := "some-uuid"
	repo.EXPECT().List(gomock.Any(), int32(10), &cursor, gomock.Any()).Return(nil, false, nil)

	_, err := svc.List(context.Background(), &services.ListInput{Limit: 10, Cursor: &cursor})
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
}

func TestTasksService_Update_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	id := uuid.New()
	now := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	newTitle := "Updated"
	repo := mocks.NewMockTaskRepository(ctrl)
	svc := services.NewTasksService(repo)

	repo.EXPECT().GetByID(gomock.Any(), id).Return(&records.Task{
		ID:                  id,
		Title:               "Old",
		SpecificationMDText: "old spec",
		TaskType:            "backend",
		Level:               "middle",
		CreatedAt:           now,
	}, nil)
	repo.EXPECT().Update(gomock.Any(), gomock.Any()).Return(nil)
	repo.EXPECT().GetTagsByTaskID(gomock.Any(), id).Return(nil, nil)
	repo.EXPECT().GetLanguagesByTaskID(gomock.Any(), id).Return(nil, nil)

	output, err := svc.Update(context.Background(), &services.UpdateInput{
		ID:    id,
		Title: &newTitle,
	})
	if err != nil {
		t.Fatalf("Update() error = %v", err)
	}
	if output.Title != "Updated" {
		t.Errorf("Title = %q, want %q", output.Title, "Updated")
	}
}

func TestTasksService_Update_NotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockTaskRepository(ctrl)
	svc := services.NewTasksService(repo)

	repo.EXPECT().GetByID(gomock.Any(), gomock.Any()).Return(nil, errors.New("not found"))

	_, err := svc.Update(context.Background(), &services.UpdateInput{ID: uuid.New()})
	if !errors.Is(err, services.ErrTaskNotFound) {
		t.Fatalf("Update() error = %v, want %v", err, services.ErrTaskNotFound)
	}
}

func TestTasksService_Delete_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockTaskRepository(ctrl)
	svc := services.NewTasksService(repo)

	repo.EXPECT().Delete(gomock.Any(), gomock.Any()).Return(nil)

	err := svc.Delete(context.Background(), uuid.New())
	if err != nil {
		t.Fatalf("Delete() error = %v", err)
	}
}

func TestTasksService_Delete_NotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockTaskRepository(ctrl)
	svc := services.NewTasksService(repo)

	repo.EXPECT().Delete(gomock.Any(), gomock.Any()).Return(errors.New("not found"))

	err := svc.Delete(context.Background(), uuid.New())
	if !errors.Is(err, services.ErrTaskNotFound) {
		t.Fatalf("Delete() error = %v, want %v", err, services.ErrTaskNotFound)
	}
}
