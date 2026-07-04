package postgresrepo_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"

	"tasks/internal/models/records"
	postgresrepo "tasks/internal/repository/postgres"
)

func TestPulledTaskRepo_Create(t *testing.T) {
	repo := newTestDB(t)
	taskRepo := postgresrepo.NewTaskRepo(repo)
	pulledRepo := postgresrepo.NewPulledTaskRepo(repo)
	cleanupTable(t, repo, "pulled_tasks")
	cleanupTable(t, repo, "tasks")

	task := &records.Task{
		ID:                  uuid.New(),
		TaskName:            "pulled-create",
		Title:               "Test Task",
		SpecificationMDText: "# Spec",
		TaskType:            "backend",
		Level:               "middle",
		CreatedAt:           time.Now(),
	}
	if err := taskRepo.Create(context.Background(), task); err != nil {
		t.Fatalf("TaskRepo.Create() error = %v", err)
	}

	pt := &records.PulledTask{
		ID:         uuid.New(),
		IdentityID: uuid.New(),
		TaskID:     task.ID,
		Repo:       "alice/golden-pizza-api",
		CloneURL:   "http://gitea.local/alice/golden-pizza-api.git",
		CreatedAt:  time.Now(),
	}

	err := pulledRepo.Create(context.Background(), pt)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	got, err := pulledRepo.GetByID(context.Background(), pt.ID)
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}

	if got.Repo != pt.Repo {
		t.Errorf("Repo = %v, want %v", got.Repo, pt.Repo)
	}
	if got.CloneURL != pt.CloneURL {
		t.Errorf("CloneURL = %v, want %v", got.CloneURL, pt.CloneURL)
	}
	if got.IdentityID != pt.IdentityID {
		t.Errorf("IdentityID = %v, want %v", got.IdentityID, pt.IdentityID)
	}
	if got.TaskID != pt.TaskID {
		t.Errorf("TaskID = %v, want %v", got.TaskID, pt.TaskID)
	}
}

func TestPulledTaskRepo_GetByID_NotFound(t *testing.T) {
	repo := newTestDB(t)
	pulledRepo := postgresrepo.NewPulledTaskRepo(repo)
	cleanupTable(t, repo, "pulled_tasks")

	_, err := pulledRepo.GetByID(context.Background(), uuid.New())
	if err == nil {
		t.Fatal("GetByID() error = nil, want not found error")
	}
}

func TestPulledTaskRepo_GetByTaskID(t *testing.T) {
	repo := newTestDB(t)
	taskRepo := postgresrepo.NewTaskRepo(repo)
	pulledRepo := postgresrepo.NewPulledTaskRepo(repo)
	cleanupTable(t, repo, "pulled_tasks")
	cleanupTable(t, repo, "tasks")

	task := &records.Task{
		ID:                  uuid.New(),
		TaskName:            "pulled-by-task",
		Title:               "Test Task",
		SpecificationMDText: "# Spec",
		TaskType:            "backend",
		Level:               "middle",
		CreatedAt:           time.Now(),
	}
	if err := taskRepo.Create(context.Background(), task); err != nil {
		t.Fatalf("TaskRepo.Create() error = %v", err)
	}

	pt := &records.PulledTask{
		ID:         uuid.New(),
		IdentityID: uuid.New(),
		TaskID:     task.ID,
		Repo:       "bob/repo",
		CloneURL:   "http://gitea.local/bob/repo.git",
		CreatedAt:  time.Now(),
	}
	if err := pulledRepo.Create(context.Background(), pt); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	got, err := pulledRepo.GetByTaskID(context.Background(), task.ID)
	if err != nil {
		t.Fatalf("GetByTaskID() error = %v", err)
	}

	if got.ID != pt.ID {
		t.Errorf("ID = %v, want %v", got.ID, pt.ID)
	}
}

func TestPulledTaskRepo_GetByTaskIDAndIdentityID(t *testing.T) {
	repo := newTestDB(t)
	taskRepo := postgresrepo.NewTaskRepo(repo)
	pulledRepo := postgresrepo.NewPulledTaskRepo(repo)
	cleanupTable(t, repo, "pulled_tasks")
	cleanupTable(t, repo, "tasks")

	task := &records.Task{
		ID:                  uuid.New(),
		TaskName:            "pulled-by-task-identity",
		Title:               "Test Task",
		SpecificationMDText: "# Spec",
		TaskType:            "backend",
		Level:               "middle",
		CreatedAt:           time.Now(),
	}
	if err := taskRepo.Create(context.Background(), task); err != nil {
		t.Fatalf("TaskRepo.Create() error = %v", err)
	}

	identityID := uuid.New()
	pt := &records.PulledTask{
		ID:         uuid.New(),
		IdentityID: identityID,
		TaskID:     task.ID,
		Repo:       "carol/repo",
		CloneURL:   "http://gitea.local/carol/repo.git",
		CreatedAt:  time.Now(),
	}
	if err := pulledRepo.Create(context.Background(), pt); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	got, err := pulledRepo.GetByTaskIDAndIdentityID(context.Background(), task.ID, identityID)
	if err != nil {
		t.Fatalf("GetByTaskIDAndIdentityID() error = %v", err)
	}

	if got.ID != pt.ID {
		t.Errorf("ID = %v, want %v", got.ID, pt.ID)
	}
}

func TestPulledTaskRepo_ListByIdentityID(t *testing.T) {
	repo := newTestDB(t)
	taskRepo := postgresrepo.NewTaskRepo(repo)
	pulledRepo := postgresrepo.NewPulledTaskRepo(repo)
	cleanupTable(t, repo, "pulled_tasks")
	cleanupTable(t, repo, "tasks")

	identityID := uuid.New()

	for i := 0; i < 3; i++ {
		task := &records.Task{
			ID:                  uuid.New(),
			TaskName:            fmt.Sprintf("pulled-list-%d", i),
			Title:               "Task",
			SpecificationMDText: "spec",
			TaskType:            "backend",
			Level:               "middle",
			CreatedAt:           time.Now(),
		}
		if err := taskRepo.Create(context.Background(), task); err != nil {
			t.Fatalf("TaskRepo.Create() error = %v", err)
		}

		pt := &records.PulledTask{
			ID:         uuid.New(),
			IdentityID: identityID,
			TaskID:     task.ID,
			Repo:       "user/repo",
			CloneURL:   "http://gitea.local/user/repo.git",
			CreatedAt:  time.Now().Add(time.Duration(i) * time.Second),
		}
		if err := pulledRepo.Create(context.Background(), pt); err != nil {
			t.Fatalf("Create() error = %v", err)
		}
	}

	items, err := pulledRepo.ListByIdentityID(context.Background(), identityID)
	if err != nil {
		t.Fatalf("ListByIdentityID() error = %v", err)
	}

	if len(items) != 3 {
		t.Errorf("ListByIdentityID() returned %d items, want 3", len(items))
	}
}

func TestPulledTaskRepo_Delete(t *testing.T) {
	repo := newTestDB(t)
	taskRepo := postgresrepo.NewTaskRepo(repo)
	pulledRepo := postgresrepo.NewPulledTaskRepo(repo)
	cleanupTable(t, repo, "pulled_tasks")
	cleanupTable(t, repo, "tasks")

	task := &records.Task{
		ID:                  uuid.New(),
		TaskName:            "pulled-delete",
		Title:               "To Delete",
		SpecificationMDText: "spec",
		TaskType:            "backend",
		Level:               "middle",
		CreatedAt:           time.Now(),
	}
	if err := taskRepo.Create(context.Background(), task); err != nil {
		t.Fatalf("TaskRepo.Create() error = %v", err)
	}

	pt := &records.PulledTask{
		ID:         uuid.New(),
		IdentityID: uuid.New(),
		TaskID:     task.ID,
		Repo:       "user/repo",
		CloneURL:   "http://gitea.local/user/repo.git",
		CreatedAt:  time.Now(),
	}
	if err := pulledRepo.Create(context.Background(), pt); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	if err := pulledRepo.Delete(context.Background(), pt.ID); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	_, err := pulledRepo.GetByID(context.Background(), pt.ID)
	if err == nil {
		t.Fatal("GetByID() error = nil, want not found error after delete")
	}
}

func TestPulledTaskRepo_Delete_NotFound(t *testing.T) {
	repo := newTestDB(t)
	pulledRepo := postgresrepo.NewPulledTaskRepo(repo)
	cleanupTable(t, repo, "pulled_tasks")

	err := pulledRepo.Delete(context.Background(), uuid.New())
	if err == nil {
		t.Fatal("Delete() error = nil, want not found error")
	}
}

func TestPulledTaskRepo_ExistsByTaskIDAndIdentityID(t *testing.T) {
	repo := newTestDB(t)
	taskRepo := postgresrepo.NewTaskRepo(repo)
	pulledRepo := postgresrepo.NewPulledTaskRepo(repo)
	cleanupTable(t, repo, "pulled_tasks")
	cleanupTable(t, repo, "tasks")

	task := &records.Task{
		ID:                  uuid.New(),
		TaskName:            "pulled-exists",
		Title:               "Test Task",
		SpecificationMDText: "spec",
		TaskType:            "backend",
		Level:               "middle",
		CreatedAt:           time.Now(),
	}
	if err := taskRepo.Create(context.Background(), task); err != nil {
		t.Fatalf("TaskRepo.Create() error = %v", err)
	}

	identityID := uuid.New()

	exists, err := pulledRepo.ExistsByTaskIDAndIdentityID(context.Background(), task.ID, identityID)
	if err != nil {
		t.Fatalf("ExistsByTaskIDAndIdentityID() error = %v", err)
	}
	if exists {
		t.Error("ExistsByTaskIDAndIdentityID() = true, want false")
	}

	pt := &records.PulledTask{
		ID:         uuid.New(),
		IdentityID: identityID,
		TaskID:     task.ID,
		Repo:       "user/repo",
		CloneURL:   "http://gitea.local/user/repo.git",
		CreatedAt:  time.Now(),
	}
	if err := pulledRepo.Create(context.Background(), pt); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	exists, err = pulledRepo.ExistsByTaskIDAndIdentityID(context.Background(), task.ID, identityID)
	if err != nil {
		t.Fatalf("ExistsByTaskIDAndIdentityID() error = %v", err)
	}
	if !exists {
		t.Error("ExistsByTaskIDAndIdentityID() = false, want true")
	}
}
