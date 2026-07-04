package postgresrepo_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"

	"tasks/internal/models/records"
	postgresrepo "tasks/internal/repository/postgres"
	"tasks/internal/services"
)

func TestTaskRepo_Create(t *testing.T) {
	repo := newTestDB(t)
	taskRepo := postgresrepo.NewTaskRepo(repo)
	cleanupTable(t, repo, "tasks")

	task := &records.Task{
		ID:                  uuid.New(),
		TaskName:            "test-task",
		Title:               "Test Task",
		SpecificationMDText: "# Spec\n\nSome spec.",
		TaskType:            "backend",
		Level:               "middle",
		CreatedAt:           time.Now(),
	}

	err := taskRepo.Create(context.Background(), task)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	got, err := taskRepo.GetByID(context.Background(), task.ID)
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}

	if got.Title != task.Title {
		t.Errorf("Title = %v, want %v", got.Title, task.Title)
	}
	if got.SpecificationMDText != task.SpecificationMDText {
		t.Errorf("SpecificationMDText = %v, want %v", got.SpecificationMDText, task.SpecificationMDText)
	}
}

func TestTaskRepo_GetByID_NotFound(t *testing.T) {
	repo := newTestDB(t)
	taskRepo := postgresrepo.NewTaskRepo(repo)
	cleanupTable(t, repo, "tasks")

	_, err := taskRepo.GetByID(context.Background(), uuid.New())
	if err == nil {
		t.Fatal("GetByID() error = nil, want not found error")
	}
}

func TestTaskRepo_Update(t *testing.T) {
	repo := newTestDB(t)
	taskRepo := postgresrepo.NewTaskRepo(repo)
	cleanupTable(t, repo, "tasks")

	task := &records.Task{
		ID:                  uuid.New(),
		TaskName:            "old-title",
		Title:               "Old Title",
		SpecificationMDText: "old spec",
		TaskType:            "backend",
		Level:               "middle",
		CreatedAt:           time.Now(),
	}

	if err := taskRepo.Create(context.Background(), task); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	task.Title = "New Title"
	task.SpecificationMDText = "new spec"

	if err := taskRepo.Update(context.Background(), task); err != nil {
		t.Fatalf("Update() error = %v", err)
	}

	got, err := taskRepo.GetByID(context.Background(), task.ID)
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}

	if got.Title != "New Title" {
		t.Errorf("Title = %v, want New Title", got.Title)
	}
	if got.SpecificationMDText != "new spec" {
		t.Errorf("SpecificationMDText = %v, want new spec", got.SpecificationMDText)
	}
}

func TestTaskRepo_Delete(t *testing.T) {
	repo := newTestDB(t)
	taskRepo := postgresrepo.NewTaskRepo(repo)
	cleanupTable(t, repo, "tasks")

	task := &records.Task{
		ID:                  uuid.New(),
		TaskName:            "to-delete",
		Title:               "To Delete",
		SpecificationMDText: "spec",
		TaskType:            "backend",
		Level:               "middle",
		CreatedAt:           time.Now(),
	}

	if err := taskRepo.Create(context.Background(), task); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	if err := taskRepo.Delete(context.Background(), task.ID); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	_, err := taskRepo.GetByID(context.Background(), task.ID)
	if err == nil {
		t.Fatal("GetByID() error = nil, want not found error after delete")
	}
}

func TestTaskRepo_List_Empty(t *testing.T) {
	repo := newTestDB(t)
	taskRepo := postgresrepo.NewTaskRepo(repo)
	cleanupTable(t, repo, "tasks")

	items, hasNext, err := taskRepo.List(context.Background(), 20, nil, services.ListFilters{})
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(items) != 0 {
		t.Errorf("List() returned %d items, want 0", len(items))
	}
	if hasNext {
		t.Error("List() hasNextPage = true, want false")
	}
}

func TestTaskRepo_List_WithItems(t *testing.T) {
	repo := newTestDB(t)
	taskRepo := postgresrepo.NewTaskRepo(repo)
	cleanupTable(t, repo, "tasks")

	for i := 0; i < 3; i++ {
		if err := taskRepo.Create(context.Background(), &records.Task{
			ID:                  uuid.New(),
			TaskName:            fmt.Sprintf("task-%d", i),
			Title:               "Task",
			SpecificationMDText: "spec",
			TaskType:            "backend",
			Level:               "middle",
			CreatedAt:           time.Now().Add(time.Duration(i) * time.Second),
		}); err != nil {
			t.Fatalf("Create() error = %v", err)
		}
	}

	items, hasNext, err := taskRepo.List(context.Background(), 2, nil, services.ListFilters{})
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(items) != 2 {
		t.Errorf("List() returned %d items, want 2", len(items))
	}
	if !hasNext {
		t.Error("List() hasNextPage = false, want true")
	}
}

func TestTaskRepo_List_Pagination(t *testing.T) {
	repo := newTestDB(t)
	taskRepo := postgresrepo.NewTaskRepo(repo)
	cleanupTable(t, repo, "tasks")

	ids := make([]uuid.UUID, 3)
	for i := 0; i < 3; i++ {
		ids[i] = uuid.New()
		if err := taskRepo.Create(context.Background(), &records.Task{
			ID:                  ids[i],
			TaskName:            fmt.Sprintf("page-task-%d", i),
			Title:               "Task",
			SpecificationMDText: "spec",
			TaskType:            "backend",
			Level:               "middle",
			CreatedAt:           time.Now().Add(time.Duration(i) * time.Second),
		}); err != nil {
			t.Fatalf("Create() error = %v", err)
		}
	}

	page1, hasNext, err := taskRepo.List(context.Background(), 2, nil, services.ListFilters{})
	if err != nil {
		t.Fatalf("List() page1 error = %v", err)
	}
	if len(page1) != 2 {
		t.Fatalf("List() page1 returned %d items, want 2", len(page1))
	}
	if !hasNext {
		t.Fatal("List() page1 hasNextPage = false, want true")
	}

	cursor := page1[len(page1)-1].ID.String()
	page2, hasNext, err := taskRepo.List(context.Background(), 2, &cursor, services.ListFilters{})
	if err != nil {
		t.Fatalf("List() page2 error = %v", err)
	}
	if len(page2) != 1 {
		t.Fatalf("List() page2 returned %d items, want 1", len(page2))
	}
	if hasNext {
		t.Error("List() page2 hasNextPage = true, want false")
	}
}

func TestRepo_WithTx(t *testing.T) {
	repo := newTestDB(t)
	cleanupTable(t, repo, "tasks")

	ctx := context.Background()
	taskRepo := postgresrepo.NewTaskRepo(repo)

	committedID := uuid.New()
	err := repo.WithTx(ctx, func(txRepo *postgresrepo.Repo) error {
		return postgresrepo.NewTaskRepo(txRepo).Create(ctx, &records.Task{
			ID:                  committedID,
			TaskName:            "committed",
			Title:               "committed",
			SpecificationMDText: "spec",
			TaskType:            "backend",
			Level:               "middle",
			CreatedAt:           time.Now(),
		})
	})
	if err != nil {
		t.Fatalf("WithTx() commit error = %v", err)
	}

	if _, err := taskRepo.GetByID(ctx, committedID); err != nil {
		t.Fatalf("GetByID() committed task error = %v", err)
	}

	rolledBackID := uuid.New()
	wantErr := context.Canceled
	err = repo.WithTx(ctx, func(txRepo *postgresrepo.Repo) error {
		if err := postgresrepo.NewTaskRepo(txRepo).Create(ctx, &records.Task{
			ID:                  rolledBackID,
			TaskName:            "rolled-back",
			Title:               "rolled back",
			SpecificationMDText: "spec",
			TaskType:            "backend",
			Level:               "middle",
			CreatedAt:           time.Now(),
		}); err != nil {
			return err
		}
		return wantErr
	})
	if err != wantErr {
		t.Fatalf("WithTx() rollback error = %v, want %v", err, wantErr)
	}

	if _, err := taskRepo.GetByID(ctx, rolledBackID); err == nil {
		t.Fatal("GetByID() rolled back task error = nil")
	}
}

func TestStore_WithTx(t *testing.T) {
	repo := newTestDB(t)
	cleanupTable(t, repo, "tasks")
	store := postgresrepo.NewStore(repo)

	ctx := context.Background()
	id := uuid.New()

	err := store.WithTx(ctx, func(s *postgresrepo.Store) error {
		return s.Tasks().Create(ctx, &records.Task{
			ID:                  id,
			TaskName:            "store-tx",
			Title:               "store tx",
			SpecificationMDText: "spec",
			TaskType:            "backend",
			Level:               "middle",
			CreatedAt:           time.Now(),
		})
	})
	if err != nil {
		t.Fatalf("Store.WithTx() error = %v", err)
	}

	if _, err := store.Tasks().GetByID(ctx, id); err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}
}
