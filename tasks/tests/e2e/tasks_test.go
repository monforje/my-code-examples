package e2e_test

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"

	"tasks/internal/config"
	httpserver "tasks/internal/http/gen"
	"tasks/internal/repository/security"
	e2ehelpers "tasks/tests/e2e/helpers"
)

func TestTasksCreate_ValidData(t *testing.T) {
	resetE2E(t)

	resp := client.PostAuthJSON(t, getJWT(t), "/tasks", httpserver.CreateTaskRequest{
		TaskName:            "e2e-task",
		Title:               "E2E Task",
		Description:         "E2E Description",
		SpecificationMdText: "# E2E Spec",
		TaskType:            "backend",
		Level:               "middle",
		TagIds:              []string{},
		LanguageIds:         []string{},
	})
	e2ehelpers.ExpectStatus(t, resp, http.StatusCreated)

	task := e2ehelpers.Decode[e2ehelpers.TaskResponse](t, resp)
	if task.Title != "E2E Task" {
		t.Fatalf("title = %q, want %q", task.Title, "E2E Task")
	}
	if task.Description != "E2E Description" {
		t.Fatalf("description = %q, want %q", task.Description, "E2E Description")
	}
	if task.SpecificationMdText != "# E2E Spec" {
		t.Fatalf("spec = %q, want %q", task.SpecificationMdText, "# E2E Spec")
	}
	if task.TaskType != "backend" {
		t.Fatalf("task_type = %q, want %q", task.TaskType, "backend")
	}
	if task.Level != "middle" {
		t.Fatalf("level = %q, want %q", task.Level, "middle")
	}
	if task.ID == "" {
		t.Fatal("id is empty")
	}

	id := uuid.MustParse(task.ID)
	dbTask := e2ehelpers.GetTaskByID(t, e2eEnv.PgPool(), id)
	if dbTask.Title != "E2E Task" {
		t.Fatalf("db title = %q, want %q", dbTask.Title, "E2E Task")
	}
	if dbTask.TaskType != "backend" {
		t.Fatalf("db task_type = %q, want %q", dbTask.TaskType, "backend")
	}
}

func TestTasksCreate_EmptyTitle(t *testing.T) {
	resetE2E(t)

	resp := client.PostAuthJSON(t, getJWT(t), "/tasks", httpserver.CreateTaskRequest{
		Title:               "",
		SpecificationMdText: "spec",
	})
	e2ehelpers.ExpectStatus(t, resp, http.StatusUnprocessableEntity)

	if e2ehelpers.CountTasks(t, e2eEnv.PgPool()) != 0 {
		t.Fatal("db should have 0 tasks after failed create")
	}
}

func TestTasksCreate_WithoutAuth(t *testing.T) {
	resetE2E(t)

	resp := client.PostJSON(t, "/tasks", httpserver.CreateTaskRequest{
		Title:               "No Auth",
		SpecificationMdText: "spec",
	})
	e2ehelpers.ExpectStatus(t, resp, http.StatusUnauthorized)
}

func TestTasksGet_ExistingTask(t *testing.T) {
	resetE2E(t)

	created := createTestTask(t, "Get Task", "get spec")

	resp := client.GetAuth(t, getJWT(t), "/tasks/"+created.ID)
	e2ehelpers.ExpectStatus(t, resp, http.StatusOK)

	task := e2ehelpers.Decode[e2ehelpers.TaskResponse](t, resp)
	if task.Title != "Get Task" {
		t.Fatalf("title = %q, want %q", task.Title, "Get Task")
	}
}

func TestTasksGet_NotFound(t *testing.T) {
	resetE2E(t)

	id := uuid.New()
	resp := client.GetAuth(t, getJWT(t), "/tasks/"+id.String())
	e2ehelpers.ExpectStatus(t, resp, http.StatusNotFound)
}

func TestTasksGet_InvalidUUID(t *testing.T) {
	resetE2E(t)

	resp := client.GetAuth(t, getJWT(t), "/tasks/not-a-uuid")
	e2ehelpers.ExpectStatus(t, resp, http.StatusBadRequest)
}

func TestTasksList_Empty(t *testing.T) {
	resetE2E(t)

	resp := client.GetAuth(t, getJWT(t), "/tasks?limit=20")
	e2ehelpers.ExpectStatus(t, resp, http.StatusOK)

	list := e2ehelpers.Decode[e2ehelpers.TaskListResponse](t, resp)
	if len(list.Items) != 0 {
		t.Fatalf("items count = %d, want 0", len(list.Items))
	}
	if list.PageInfo.HasNextPage {
		t.Fatal("has_next_page = true, want false")
	}
}

func TestTasksList_WithItems(t *testing.T) {
	resetE2E(t)

	for i := 0; i < 3; i++ {
		createTestTask(t, fmt.Sprintf("Task %d", i), "spec")
	}

	resp := client.GetAuth(t, getJWT(t), "/tasks?limit=2")
	e2ehelpers.ExpectStatus(t, resp, http.StatusOK)

	list := e2ehelpers.Decode[e2ehelpers.TaskListResponse](t, resp)
	if len(list.Items) != 2 {
		t.Fatalf("items count = %d, want 2", len(list.Items))
	}
	if !list.PageInfo.HasNextPage {
		t.Fatal("has_next_page = false, want true")
	}
	if list.PageInfo.NextCursor == nil {
		t.Fatal("next_cursor is nil")
	}
}

func TestTasksList_Pagination(t *testing.T) {
	resetE2E(t)

	for i := 0; i < 3; i++ {
		createTestTask(t, fmt.Sprintf("Task %d", i), "spec")
	}

	resp1 := client.GetAuth(t, getJWT(t), "/tasks?limit=2")
	e2ehelpers.ExpectStatus(t, resp1, http.StatusOK)
	page1 := e2ehelpers.Decode[e2ehelpers.TaskListResponse](t, resp1)
	if len(page1.Items) != 2 {
		t.Fatalf("page1 items = %d, want 2", len(page1.Items))
	}

	resp2 := client.GetAuth(t, getJWT(t), "/tasks?limit=2&cursor="+*page1.PageInfo.NextCursor)
	e2ehelpers.ExpectStatus(t, resp2, http.StatusOK)
	page2 := e2ehelpers.Decode[e2ehelpers.TaskListResponse](t, resp2)
	if len(page2.Items) != 1 {
		t.Fatalf("page2 items = %d, want 1", len(page2.Items))
	}
	if page2.PageInfo.HasNextPage {
		t.Fatal("page2 has_next_page = true, want false")
	}
}

func TestTasksList_Search(t *testing.T) {
	resetE2E(t)

	createTestTask(t, "Auth Service", "auth spec")
	createTestTask(t, "User Service", "user spec")

	resp := client.GetAuth(t, getJWT(t), "/tasks?limit=20&search=auth")
	e2ehelpers.ExpectStatus(t, resp, http.StatusOK)

	list := e2ehelpers.Decode[e2ehelpers.TaskListResponse](t, resp)
	if len(list.Items) != 1 {
		t.Fatalf("items count = %d, want 1", len(list.Items))
	}
	if list.Items[0].Title != "Auth Service" {
		t.Fatalf("title = %q, want %q", list.Items[0].Title, "Auth Service")
	}
}

func TestTasksList_FilterByTaskType(t *testing.T) {
	resetE2E(t)

	createTestTaskWithType(t, "Backend Task", "spec", "backend")
	createTestTaskWithType(t, "Frontend Task", "spec", "frontend")

	resp := client.GetAuth(t, getJWT(t), "/tasks?limit=20&task_type=backend")
	e2ehelpers.ExpectStatus(t, resp, http.StatusOK)

	list := e2ehelpers.Decode[e2ehelpers.TaskListResponse](t, resp)
	if len(list.Items) != 1 {
		t.Fatalf("items count = %d, want 1", len(list.Items))
	}
	if list.Items[0].TaskType != "backend" {
		t.Fatalf("task_type = %q, want %q", list.Items[0].TaskType, "backend")
	}
}

func TestTasksList_FilterByLevel(t *testing.T) {
	resetE2E(t)

	createTestTaskWithLevel(t, "Junior Task", "spec", "junior")
	createTestTaskWithLevel(t, "Senior Task", "spec", "senior")

	resp := client.GetAuth(t, getJWT(t), "/tasks?limit=20&level=junior")
	e2ehelpers.ExpectStatus(t, resp, http.StatusOK)

	list := e2ehelpers.Decode[e2ehelpers.TaskListResponse](t, resp)
	if len(list.Items) != 1 {
		t.Fatalf("items count = %d, want 1", len(list.Items))
	}
	if list.Items[0].Level != "junior" {
		t.Fatalf("level = %q, want %q", list.Items[0].Level, "junior")
	}
}

func TestTasksUpdate_Title(t *testing.T) {
	resetE2E(t)

	created := createTestTask(t, "Old Title", "old spec")

	newTitle := "New Title"
	resp := client.PatchAuthJSON(t, getJWT(t), "/tasks/"+created.ID, httpserver.UpdateTaskRequest{
		Title: &newTitle,
	})
	e2ehelpers.ExpectStatus(t, resp, http.StatusOK)

	updated := e2ehelpers.Decode[e2ehelpers.TaskResponse](t, resp)
	if updated.Title != "New Title" {
		t.Fatalf("title = %q, want %q", updated.Title, "New Title")
	}

	dbTask := e2ehelpers.GetTaskByID(t, e2eEnv.PgPool(), uuid.MustParse(created.ID))
	if dbTask.Title != "New Title" {
		t.Fatalf("db title = %q, want %q", dbTask.Title, "New Title")
	}
}

func TestTasksUpdate_Specification(t *testing.T) {
	resetE2E(t)

	created := createTestTask(t, "Task", "old spec")

	newSpec := "# New Spec"
	resp := client.PatchAuthJSON(t, getJWT(t), "/tasks/"+created.ID, httpserver.UpdateTaskRequest{
		SpecificationMdText: &newSpec,
	})
	e2ehelpers.ExpectStatus(t, resp, http.StatusOK)

	dbTask := e2ehelpers.GetTaskByID(t, e2eEnv.PgPool(), uuid.MustParse(created.ID))
	if dbTask.SpecificationMDText != "# New Spec" {
		t.Fatalf("db spec = %q, want %q", dbTask.SpecificationMDText, "# New Spec")
	}
}

func TestTasksUpdate_NotFound(t *testing.T) {
	resetE2E(t)

	id := uuid.New()
	newTitle := "Updated"
	resp := client.PatchAuthJSON(t, getJWT(t), "/tasks/"+id.String(), httpserver.UpdateTaskRequest{
		Title: &newTitle,
	})
	e2ehelpers.ExpectStatus(t, resp, http.StatusNotFound)
}

func TestTasksDelete_ExistingTask(t *testing.T) {
	resetE2E(t)

	created := createTestTask(t, "To Delete", "spec")

	resp := client.DeleteAuth(t, getJWT(t), "/tasks/"+created.ID)
	e2ehelpers.ExpectStatus(t, resp, http.StatusNoContent)

	id := uuid.MustParse(created.ID)
	e2ehelpers.RequireNoTask(t, e2eEnv.PgPool(), id)
}

func TestTasksDelete_NotFound(t *testing.T) {
	resetE2E(t)

	id := uuid.New()
	resp := client.DeleteAuth(t, getJWT(t), "/tasks/"+id.String())
	e2ehelpers.ExpectStatus(t, resp, http.StatusNotFound)
}

func TestTasksFullCRUD(t *testing.T) {
	resetE2E(t)

	created := createTestTask(t, "CRUD Task", "crud spec")
	if created.ID == "" {
		t.Fatal("created task id is empty")
	}

	resp := client.GetAuth(t, getJWT(t), "/tasks/"+created.ID)
	e2ehelpers.ExpectStatus(t, resp, http.StatusOK)
	got := e2ehelpers.Decode[e2ehelpers.TaskResponse](t, resp)
	if got.Title != "CRUD Task" {
		t.Fatalf("get title = %q, want %q", got.Title, "CRUD Task")
	}

	updatedTitle := "CRUD Updated"
	resp = client.PatchAuthJSON(t, getJWT(t), "/tasks/"+created.ID, httpserver.UpdateTaskRequest{
		Title: &updatedTitle,
	})
	e2ehelpers.ExpectStatus(t, resp, http.StatusOK)
	updated := e2ehelpers.Decode[e2ehelpers.TaskResponse](t, resp)
	if updated.Title != "CRUD Updated" {
		t.Fatalf("update title = %q, want %q", updated.Title, "CRUD Updated")
	}

	resp = client.DeleteAuth(t, getJWT(t), "/tasks/"+created.ID)
	e2ehelpers.ExpectStatus(t, resp, http.StatusNoContent)

	resp = client.GetAuth(t, getJWT(t), "/tasks/"+created.ID)
	e2ehelpers.ExpectStatus(t, resp, http.StatusNotFound)
}

func TestTasksGitCreate_FullFlow(t *testing.T) {
	resetE2E(t)

	identityID := uuid.New()
	token := getJWTForIdentity(t, identityID)

	resp := client.PostAuthJSON(t, token, "/tasks", httpserver.CreateTaskRequest{
		TaskName:            "pizza-api",
		Title:               "Pizza Ordering API",
		Description:         "Pizza API task",
		SpecificationMdText: "# Pizza",
		TaskType:            "backend",
		Level:               "middle",
		TagIds:              []string{},
		LanguageIds:         []string{},
	})
	e2ehelpers.ExpectStatus(t, resp, http.StatusCreated)
	created := e2ehelpers.Decode[e2ehelpers.TaskResponse](t, resp)

	resp = client.PostAuthJSON(t, token, "/tasks/pizza-api/git", httpserver.GitTaskRequest{TaskName: "pizza-api"})
	e2ehelpers.ExpectStatus(t, resp, http.StatusCreated)
	gitResp := e2ehelpers.Decode[e2ehelpers.GitTaskResponse](t, resp)

	if gitResp.TaskName != "pizza-api" {
		t.Fatalf("task_name = %q, want pizza-api", gitResp.TaskName)
	}
	if gitResp.Repo != "alice/golden-pizza-api" {
		t.Fatalf("repo = %q, want alice/golden-pizza-api", gitResp.Repo)
	}
	if gitResp.CloneURL != "http://alice:git-token@gitea.local/alice/golden-pizza-api.git" {
		t.Fatalf("clone_url = %q", gitResp.CloneURL)
	}

	external.mu.Lock()
	usersCalls := external.usersCalls
	gitCalls := external.gitCalls
	lastIdentityID := external.lastIdentityID
	lastUsername := external.lastUsername
	lastTaskID := external.lastTaskID
	external.mu.Unlock()

	if usersCalls != 1 || gitCalls != 1 {
		t.Fatalf("external calls users=%d git=%d, want 1/1", usersCalls, gitCalls)
	}
	if lastIdentityID != identityID.String() {
		t.Fatalf("identity_id = %q, want %q", lastIdentityID, identityID.String())
	}
	if lastUsername != "alice" || lastTaskID != "pizza-api" {
		t.Fatalf("git request username=%q task_id=%q", lastUsername, lastTaskID)
	}

	var repo, cloneURL string
	err := e2eEnv.PgPool().QueryRow(context.Background(), `
		select repo, clone_url
		from pulled_tasks
		where task_id = $1 and identity_id = $2
	`, uuid.MustParse(created.ID), identityID).Scan(&repo, &cloneURL)
	if err != nil {
		t.Fatalf("query pulled task: %v", err)
	}
	if repo != gitResp.Repo || cloneURL != gitResp.CloneURL {
		t.Fatalf("pulled task repo/clone_url = %q/%q", repo, cloneURL)
	}
}

func TestTagsList(t *testing.T) {
	resetE2E(t)

	resp := client.GetAuth(t, getJWT(t), "/tasks/tags")
	e2ehelpers.ExpectStatus(t, resp, http.StatusOK)
}

func TestLanguagesList(t *testing.T) {
	resetE2E(t)

	resp := client.GetAuth(t, getJWT(t), "/tasks/languages")
	e2ehelpers.ExpectStatus(t, resp, http.StatusOK)
}

func createTestTask(t *testing.T, title, spec string) e2ehelpers.TaskResponse {
	t.Helper()
	resp := client.PostAuthJSON(t, getJWT(t), "/tasks", httpserver.CreateTaskRequest{
		TaskName:            taskNameFromTitle(title),
		Title:               title,
		Description:         "Description for " + title,
		SpecificationMdText: spec,
		TaskType:            "backend",
		Level:               "middle",
		TagIds:              []string{},
		LanguageIds:         []string{},
	})
	e2ehelpers.ExpectStatus(t, resp, http.StatusCreated)
	return e2ehelpers.Decode[e2ehelpers.TaskResponse](t, resp)
}

func createTestTaskWithType(t *testing.T, title, spec, taskType string) e2ehelpers.TaskResponse {
	t.Helper()
	resp := client.PostAuthJSON(t, getJWT(t), "/tasks", httpserver.CreateTaskRequest{
		TaskName:            taskNameFromTitle(title),
		Title:               title,
		Description:         "Description for " + title,
		SpecificationMdText: spec,
		TaskType:            httpserver.TaskType(taskType),
		Level:               "middle",
		TagIds:              []string{},
		LanguageIds:         []string{},
	})
	e2ehelpers.ExpectStatus(t, resp, http.StatusCreated)
	return e2ehelpers.Decode[e2ehelpers.TaskResponse](t, resp)
}

func createTestTaskWithLevel(t *testing.T, title, spec, level string) e2ehelpers.TaskResponse {
	t.Helper()
	resp := client.PostAuthJSON(t, getJWT(t), "/tasks", httpserver.CreateTaskRequest{
		TaskName:            taskNameFromTitle(title),
		Title:               title,
		Description:         "Description for " + title,
		SpecificationMdText: spec,
		TaskType:            "backend",
		Level:               httpserver.Level(level),
		TagIds:              []string{},
		LanguageIds:         []string{},
	})
	e2ehelpers.ExpectStatus(t, resp, http.StatusCreated)
	return e2ehelpers.Decode[e2ehelpers.TaskResponse](t, resp)
}

func taskNameFromTitle(title string) string {
	return fmt.Sprintf("test-%x", uuid.NewSHA1(uuid.NameSpaceOID, []byte(title)))
}

func getJWT(t *testing.T) string {
	t.Helper()
	return getJWTForIdentity(t, uuid.New())
}

func getJWTForIdentity(t *testing.T, identityID uuid.UUID) string {
	t.Helper()
	m := security.NewManager("e2e-test-signing-key-which-is-long-enough", config.FeaturesConfig{
		AccessTokenTTL:  15 * time.Minute,
		RefreshTokenLen: 32,
	})
	sessionID := uuid.New()
	token, _, err := m.GenerateAccessToken(identityID, sessionID)
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}
	return token
}
