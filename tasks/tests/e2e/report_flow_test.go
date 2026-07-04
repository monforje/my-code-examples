package e2e_test

import (
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"

	httpserver "tasks/internal/http/gen"
	e2ehelpers "tasks/tests/e2e/helpers"
)

func ptrStr(s string) *string { return &s }

func sampleReportRequest(uid, commit, runID string) httpserver.CreateReportRequest {
	return httpserver.CreateReportRequest{
		Uid:       uid,
		Commit:    commit,
		RunId:     ptrStr(runID),
		Status:    httpserver.ReportStatusFailed,
		CreatedAt: time.Date(2026, 6, 24, 12, 0, 0, 0, time.UTC),
		Summary: httpserver.ReportSummary{
			Status: httpserver.Failed, Message: "Application is not reachable on localhost:8080",
			RootCause: ptrStr("APP_UNREACHABLE"), Passed: 4, Failed: 3, Blocked: 5, Warnings: 1,
		},
		Steps:           []httpserver.ReportStep{},
		LintErrors:      []httpserver.ReportLintError{},
		Warnings:        []string{},
		RawLogAvailable: true,
	}
}

// pendingReportRequest - отчёт в статусе pending (имитация ci_started от notifications).
func pendingReportRequest(uid, commit, runID string) httpserver.CreateReportRequest {
	req := sampleReportRequest(uid, commit, runID)
	req.Status = httpserver.ReportStatusPending
	req.Summary = httpserver.ReportSummary{
		Status: httpserver.Failed, Message: "CI pending", Passed: 0, Failed: 0, Blocked: 0, Warnings: 0,
	}
	return req
}

// setupReportOwner - создаёт задачу + pulled_task через git-флоу, возвращает task_name и uid репо.
func setupReportOwner(t *testing.T, token, taskName string) string {
	t.Helper()

	resp := client.PostAuthJSON(t, token, "/tasks", httpserver.CreateTaskRequest{
		TaskName:            taskName,
		Title:               "Report Task",
		Description:         "for report flow",
		SpecificationMdText: "# spec",
		TaskType:            "backend",
		Level:               "middle",
		TagIds:              []string{},
		LanguageIds:         []string{},
	})
	e2ehelpers.ExpectStatus(t, resp, http.StatusCreated)

	resp = client.PostAuthJSON(t, token, "/tasks/"+taskName+"/git", httpserver.GitTaskRequest{TaskName: taskName})
	e2ehelpers.ExpectStatus(t, resp, http.StatusCreated)
	git := e2ehelpers.Decode[e2ehelpers.GitTaskResponse](t, resp)

	return git.Repo // == uid, напр. "alice/golden-pizza-api"
}

// TestE2E_Reports_CreateThenList - полный флоу: git-подготовка → POST отчёта → GET список.
func TestE2E_Reports_CreateThenList(t *testing.T) {
	resetE2E(t)

	identityID := uuid.New()
	token := getJWTForIdentity(t, identityID)

	uid := setupReportOwner(t, token, "pizza-api")

	// POST /reports сервисным токеном (имитация notifications).
	resp := client.PostAuthJSON(t, e2eReportsToken, "/reports", sampleReportRequest(uid, "abc123", uuid.New().String()))
	e2ehelpers.ExpectStatus(t, resp, http.StatusCreated)
	created := e2ehelpers.Decode[httpserver.Report](t, resp)
	if created.Uid != uid {
		t.Fatalf("uid = %s, want %s", created.Uid, uid)
	}
	if created.Summary.RootCause == nil || *created.Summary.RootCause != "APP_UNREACHABLE" {
		t.Fatalf("root_cause = %v", created.Summary.RootCause)
	}
	if created.Id == "" {
		t.Fatal("id is empty")
	}

	// GET /reports?task_name=pizza-api тем же JWT → список с отчётом.
	resp = client.GetAuth(t, token, "/reports?task_name=pizza-api&limit=20")
	e2ehelpers.ExpectStatus(t, resp, http.StatusOK)
	list := e2ehelpers.Decode[httpserver.ReportListResponse](t, resp)
	if len(list.Items) != 1 {
		t.Fatalf("items = %d, want 1", len(list.Items))
	}
	if list.Items[0].Uid != uid {
		t.Fatalf("item uid = %s", list.Items[0].Uid)
	}
	if list.Items[0].Commit != "abc123" {
		t.Fatalf("commit = %s", list.Items[0].Commit)
	}
	if list.PageInfo.HasNextPage {
		t.Fatal("has_next_page = true, want false")
	}
}

// TestE2E_Reports_CreateWrongServiceToken - неверный сервисный токен → 401.
func TestE2E_Reports_CreateWrongServiceToken(t *testing.T) {
	resetE2E(t)

	token := getJWTForIdentity(t, uuid.New())
	uid := setupReportOwner(t, token, "pizza-api")

	resp := client.PostAuthJSON(t, "wrong-token", "/reports", sampleReportRequest(uid, "abc", uuid.New().String()))
	e2ehelpers.ExpectStatus(t, resp, http.StatusUnauthorized)
}

// TestE2E_Reports_CreateOwnerNotFound - uid без pulled_task → 404.
func TestE2E_Reports_CreateOwnerNotFound(t *testing.T) {
	resetE2E(t)

	resp := client.PostAuthJSON(t, e2eReportsToken, "/reports", sampleReportRequest("nobody/unknown-task", "abc", uuid.New().String()))
	e2ehelpers.ExpectStatus(t, resp, http.StatusNotFound)
}

// TestE2E_Reports_ListWithoutAuth - GET без JWT → 401.
func TestE2E_Reports_ListWithoutAuth(t *testing.T) {
	resetE2E(t)

	resp := client.Get(t, "/reports?task_name=pizza-api")
	e2ehelpers.ExpectStatus(t, resp, http.StatusUnauthorized)
}

// TestE2E_Reports_ListIsolatedByIdentity - чужой identity не видит чужие отчёты.
func TestE2E_Reports_ListIsolatedByIdentity(t *testing.T) {
	resetE2E(t)

	ownerToken := getJWTForIdentity(t, uuid.New())
	uid := setupReportOwner(t, ownerToken, "pizza-api")

	resp := client.PostAuthJSON(t, e2eReportsToken, "/reports", sampleReportRequest(uid, "abc", uuid.New().String()))
	e2ehelpers.ExpectStatus(t, resp, http.StatusCreated)

	// Другой пользователь — пустой список.
	otherToken := getJWTForIdentity(t, uuid.New())
	resp = client.GetAuth(t, otherToken, "/reports?task_name=pizza-api&limit=20")
	e2ehelpers.ExpectStatus(t, resp, http.StatusOK)
	list := e2ehelpers.Decode[httpserver.ReportListResponse](t, resp)
	if len(list.Items) != 0 {
		t.Fatalf("items = %d, want 0 (изолировано по identity)", len(list.Items))
	}
}

// TestE2E_Reports_Pagination - курсорная пагинация.
func TestE2E_Reports_Pagination(t *testing.T) {
	resetE2E(t)

	identityID := uuid.New()
	token := getJWTForIdentity(t, identityID)
	uid := setupReportOwner(t, token, "pizza-api")

	// 3 отчёта с разными commit и run_id (run_id уникален — иначе UPSERT схлопнет).
	for i := 0; i < 3; i++ {
		resp := client.PostAuthJSON(t, e2eReportsToken, "/reports", sampleReportRequest(uid, "c"+string(rune('1'+i)), uuid.New().String()))
		e2ehelpers.ExpectStatus(t, resp, http.StatusCreated)
	}

	// limit=2 → первая страница из 2 элементов + has_next_page.
	resp := client.GetAuth(t, token, "/reports?task_name=pizza-api&limit=2")
	e2ehelpers.ExpectStatus(t, resp, http.StatusOK)
	page1 := e2ehelpers.Decode[httpserver.ReportListResponse](t, resp)
	if len(page1.Items) != 2 {
		t.Fatalf("page1 items = %d, want 2", len(page1.Items))
	}
	if !page1.PageInfo.HasNextPage {
		t.Fatal("has_next_page = false, want true")
	}
	if page1.PageInfo.NextCursor == nil {
		t.Fatal("next_cursor is nil")
	}

	// Вторая страница по курсору → 1 элемент, без has_next_page.
	resp = client.GetAuth(t, token, "/reports?task_name=pizza-api&limit=2&cursor="+*page1.PageInfo.NextCursor)
	e2ehelpers.ExpectStatus(t, resp, http.StatusOK)
	page2 := e2ehelpers.Decode[httpserver.ReportListResponse](t, resp)
	if len(page2.Items) != 1 {
		t.Fatalf("page2 items = %d, want 1", len(page2.Items))
	}
	if page2.PageInfo.HasNextPage {
		t.Fatal("page2 has_next_page = true, want false")
	}
}

// TestE2E_Reports_Idempotency - повторная отправка того же (uid, commit) не дублирует запись.
func TestE2E_Reports_Idempotency(t *testing.T) {
	resetE2E(t)

	token := getJWTForIdentity(t, uuid.New())
	uid := setupReportOwner(t, token, "pizza-api")

	resp := client.PostAuthJSON(t, e2eReportsToken, "/reports", sampleReportRequest(uid, "abc", uuid.New().String()))
	e2ehelpers.ExpectStatus(t, resp, http.StatusCreated)
	first := e2ehelpers.Decode[httpserver.Report](t, resp)

	// Тот же (uid, commit), другой run_id — UPSERT обновляет, не создаёт дубль.
	resp = client.PostAuthJSON(t, e2eReportsToken, "/reports", sampleReportRequest(uid, "abc", uuid.New().String()))
	e2ehelpers.ExpectStatus(t, resp, http.StatusCreated)
	second := e2ehelpers.Decode[httpserver.Report](t, resp)

	if first.Id != second.Id {
		t.Fatalf("id differs: %s vs %s (expected same (uid,commit) row)", first.Id, second.Id)
	}

	// В списке ровно одна запись.
	resp = client.GetAuth(t, token, "/reports?task_name=pizza-api&limit=20")
	e2ehelpers.ExpectStatus(t, resp, http.StatusOK)
	list := e2ehelpers.Decode[httpserver.ReportListResponse](t, resp)
	if len(list.Items) != 1 {
		t.Fatalf("items = %d, want 1 (idempotent)", len(list.Items))
	}
}

// TestE2E_Reports_PendingThenFinal - pending (ci_started) затем финал (ci_finished) = одна запись.
func TestE2E_Reports_PendingThenFinal(t *testing.T) {
	resetE2E(t)

	token := getJWTForIdentity(t, uuid.New())
	uid := setupReportOwner(t, token, "pizza-api")
	runID := uuid.New().String()

	// ci_started → pending.
	resp := client.PostAuthJSON(t, e2eReportsToken, "/reports", pendingReportRequest(uid, "abc", runID))
	e2ehelpers.ExpectStatus(t, resp, http.StatusCreated)
	pending := e2ehelpers.Decode[httpserver.Report](t, resp)
	if pending.Status != httpserver.ReportStatusPending {
		t.Fatalf("status = %s, want pending", pending.Status)
	}

	// ci_finished → финальный статус по тому же run_id.
	resp = client.PostAuthJSON(t, e2eReportsToken, "/reports", sampleReportRequest(uid, "abc", runID))
	e2ehelpers.ExpectStatus(t, resp, http.StatusCreated)
	final := e2ehelpers.Decode[httpserver.Report](t, resp)

	if pending.Id != final.Id {
		t.Fatalf("id differs: %s vs %s", pending.Id, final.Id)
	}
	if final.Status != httpserver.ReportStatusFailed {
		t.Fatalf("status = %s, want failed", final.Status)
	}

	// Одна запись, не две.
	resp = client.GetAuth(t, token, "/reports?task_name=pizza-api&limit=20")
	e2ehelpers.ExpectStatus(t, resp, http.StatusOK)
	list := e2ehelpers.Decode[httpserver.ReportListResponse](t, resp)
	if len(list.Items) != 1 {
		t.Fatalf("items = %d, want 1", len(list.Items))
	}
}

// TestE2E_Reports_GetByID - GET /reports/{id} одного отчёта.
func TestE2E_Reports_GetByID(t *testing.T) {
	resetE2E(t)

	token := getJWTForIdentity(t, uuid.New())
	uid := setupReportOwner(t, token, "pizza-api")

	resp := client.PostAuthJSON(t, e2eReportsToken, "/reports", sampleReportRequest(uid, "abc", uuid.New().String()))
	e2ehelpers.ExpectStatus(t, resp, http.StatusCreated)
	created := e2ehelpers.Decode[httpserver.Report](t, resp)

	// GET по id владельцем.
	resp = client.GetAuth(t, token, "/reports/"+created.Id)
	e2ehelpers.ExpectStatus(t, resp, http.StatusOK)
	got := e2ehelpers.Decode[httpserver.Report](t, resp)
	if got.Id != created.Id {
		t.Fatalf("id = %s, want %s", got.Id, created.Id)
	}
	if got.RunId == nil || created.RunId == nil || *got.RunId != *created.RunId {
		t.Fatalf("run_id = %v, want %v", got.RunId, created.RunId)
	}

	// Чужой пользователь — 404 (изоляция по владельцу).
	otherToken := getJWTForIdentity(t, uuid.New())
	resp = client.GetAuth(t, otherToken, "/reports/"+created.Id)
	e2ehelpers.ExpectStatus(t, resp, http.StatusNotFound)

	// Несуществующий id — 404.
	resp = client.GetAuth(t, token, "/reports/"+uuid.New().String())
	e2ehelpers.ExpectStatus(t, resp, http.StatusNotFound)
}

// TestE2E_Reports_StatusFilter - фильтр списка по статусу.
func TestE2E_Reports_StatusFilter(t *testing.T) {
	resetE2E(t)

	token := getJWTForIdentity(t, uuid.New())
	uid := setupReportOwner(t, token, "pizza-api")

	// Один pending + один failed.
	client.PostAuthJSON(t, e2eReportsToken, "/reports", pendingReportRequest(uid, "p1", uuid.New().String()))
	client.PostAuthJSON(t, e2eReportsToken, "/reports", sampleReportRequest(uid, "f1", uuid.New().String()))

	// Фильтр pending → только 1.
	resp := client.GetAuth(t, token, "/reports?task_name=pizza-api&status=pending&limit=20")
	e2ehelpers.ExpectStatus(t, resp, http.StatusOK)
	list := e2ehelpers.Decode[httpserver.ReportListResponse](t, resp)
	if len(list.Items) != 1 {
		t.Fatalf("pending items = %d, want 1", len(list.Items))
	}
	if list.Items[0].Status != httpserver.ReportStatusPending {
		t.Fatalf("status = %s, want pending", list.Items[0].Status)
	}
}
