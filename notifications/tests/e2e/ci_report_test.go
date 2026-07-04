package e2e_test

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"testing"
)

func postWebhook(t *testing.T, body map[string]any) (int, map[string]any) {
	t.Helper()

	data, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	resp, err := httpClient.Post(baseURL+"/webhook", "application/json", bytes.NewReader(data))
	if err != nil {
		t.Fatalf("POST /webhook: %v", err)
	}
	defer resp.Body.Close()

	raw, _ := io.ReadAll(resp.Body)
	var result map[string]any
	_ = json.Unmarshal(raw, &result)
	return resp.StatusCode, result
}

func getReport(t *testing.T, uid string) (int, map[string]any) {
	t.Helper()

	resp, err := httpClient.Get(baseURL + "/reports/" + uid)
	if err != nil {
		t.Fatalf("GET /reports/%s: %v", uid, err)
	}
	defer resp.Body.Close()

	raw, _ := io.ReadAll(resp.Body)
	var result map[string]any
	_ = json.Unmarshal(raw, &result)
	return resp.StatusCode, result
}

func TestE2E_CIReport_StartedThenFinished(t *testing.T) {
	resetE2E(t)
	uid := "alice/golden-pizza-api"

	// 1. ci_started
	status, body := postWebhook(t, map[string]any{
		"event":  "ci_started",
		"uid":    uid,
		"commit": "abc123",
		"vm_id":  "vm-1",
	})
	if status != http.StatusOK {
		t.Fatalf("ci_started status = %d, body = %+v", status, body)
	}

	// 2. Report should be pending
	status, report := getReport(t, uid)
	if status != http.StatusOK {
		t.Fatalf("get report status = %d", status)
	}
	if report["status"] != "pending" {
		t.Fatalf("status = %v, want pending", report["status"])
	}

	// 3. ci_finished (passed)
	stdout := "✓ build (5.3s)\n✓ test (2.1s)\n"
	status, body = postWebhook(t, map[string]any{
		"event":     "ci_finished",
		"uid":       uid,
		"commit":    "abc123",
		"exit_code": 0,
		"stdout":    stdout,
		"stderr":    "",
		"stage":     "run",
	})
	if status != http.StatusOK {
		t.Fatalf("ci_finished status = %d, body = %+v", status, body)
	}

	// 4. Report should be passed. Шаги не парсятся в notifications-кеше (источник
	//    шагов — канонический JSON ci-translator, доступный через сервис tasks).
	status, report = getReport(t, uid)
	if status != http.StatusOK {
		t.Fatalf("get report status = %d", status)
	}
	if report["status"] != "passed" {
		t.Fatalf("status = %v, want passed", report["status"])
	}
	steps, ok := report["steps"].([]any)
	if !ok || len(steps) != 0 {
		t.Fatalf("steps = %v, want 0 (parsing moved to ci-translator JSON path)", report["steps"])
	}
}

func TestE2E_CIReport_Failed(t *testing.T) {
	resetE2E(t)
	uid := "bob/golden-pizza-api"

	stdout := "✓ build (5.3s)\n✗ test-register (0.8s) - exit code 1\n"
	status, _ := postWebhook(t, map[string]any{
		"event":     "ci_finished",
		"uid":       uid,
		"commit":    "def456",
		"exit_code": 1,
		"stdout":    stdout,
		"stage":     "run",
	})
	if status != http.StatusOK {
		t.Fatalf("ci_finished status = %d", status)
	}

	status, report := getReport(t, uid)
	if status != http.StatusOK {
		t.Fatalf("get report status = %d", status)
	}
	if report["status"] != "failed" {
		t.Fatalf("status = %v, want failed", report["status"])
	}
	if report["exit_code"].(float64) != 1 {
		t.Fatalf("exit_code = %v, want 1", report["exit_code"])
	}
}

func TestE2E_CIReport_Crashed(t *testing.T) {
	resetE2E(t)
	uid := "charlie/golden-pizza-api"

	status, _ := postWebhook(t, map[string]any{
		"event":     "ci_finished",
		"uid":       uid,
		"commit":    "ghi789",
		"exit_code": -1,
		"stderr":    "VM connection lost",
		"stage":     "vm_crash",
	})
	if status != http.StatusOK {
		t.Fatalf("ci_finished status = %d", status)
	}

	status, report := getReport(t, uid)
	if status != http.StatusOK {
		t.Fatalf("get report status = %d", status)
	}
	if report["status"] != "crashed" {
		t.Fatalf("status = %v, want crashed", report["status"])
	}
}

func TestE2E_CIReport_NotFound(t *testing.T) {
	resetE2E(t)

	status, report := getReport(t, "nobody/golden-pizza-api")
	if status != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", status)
	}
	if report["code"] != "NOT_FOUND" {
		t.Fatalf("code = %v, want NOT_FOUND", report["code"])
	}
}

func TestE2E_CIReport_InvalidWebhook(t *testing.T) {
	resetE2E(t)

	status, _ := postWebhook(t, map[string]any{
		"event":  "ci_started",
		"commit": "abc123",
	})
	if status != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", status)
	}
}

func TestE2E_CIReport_LatestOverwrite(t *testing.T) {
	resetE2E(t)
	uid := "dave/golden-pizza-api"

	// First commit
	postWebhook(t, map[string]any{
		"event":     "ci_finished",
		"uid":       uid,
		"commit":    "aaa111",
		"exit_code": 0,
		"stdout":    "✓ build (1.0s)\n",
		"stage":     "run",
	})

	// Second commit overwrites latest
	postWebhook(t, map[string]any{
		"event":     "ci_finished",
		"uid":       uid,
		"commit":    "bbb222",
		"exit_code": 1,
		"stdout":    "✗ test (1.0s) - exit code 1\n",
		"stage":     "run",
	})

	status, report := getReport(t, uid)
	if status != http.StatusOK {
		t.Fatalf("get report status = %d", status)
	}
	if report["commit"] != "bbb222" {
		t.Fatalf("commit = %v, want bbb222", report["commit"])
	}
}
