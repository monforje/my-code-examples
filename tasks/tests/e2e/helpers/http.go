package e2e_test_helpers

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/cookiejar"
	"testing"
	"time"
)

// Client - HTTP-клиент для e2e-тестов с поддержкой JWT и cookies.
type Client struct {
	baseURL string
	http    *http.Client
}

// Response - обёртка над http.Response для удобства тестирования.
type Response struct {
	StatusCode int
	Header     http.Header
	Body       []byte
}

// TaskResponse - ответ API с данными задачи.
type TaskResponse struct {
	ID                  string             `json:"id"`
	TaskName            string             `json:"task_name"`
	Title               string             `json:"title"`
	Description         string             `json:"description"`
	SpecificationMdText string             `json:"specification_md_text"`
	TaskType            string             `json:"task_type"`
	Level               string             `json:"level"`
	Tags                []TagResponse      `json:"tags"`
	Languages           []LanguageResponse `json:"languages"`
	CreatedAt           time.Time          `json:"created_at"`
}

// TagResponse - тег в ответе.
type TagResponse struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// LanguageResponse - язык в ответе.
type LanguageResponse struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// TaskListResponse - ответ API со списком задач.
type TaskListResponse struct {
	Items    []TaskListItemResponse `json:"items"`
	PageInfo PageInfoResponse       `json:"page_info"`
}

// TaskListItemResponse - элемент списка задач.
type TaskListItemResponse struct {
	ID          string             `json:"id"`
	TaskName    string             `json:"task_name"`
	Title       string             `json:"title"`
	Description string             `json:"description"`
	TaskType    string             `json:"task_type"`
	Level       string             `json:"level"`
	Tags        []TagResponse      `json:"tags"`
	Languages   []LanguageResponse `json:"languages"`
	CreatedAt   string             `json:"created_at"`
}

type GitTaskResponse struct {
	TaskName string `json:"task_name"`
	Repo     string `json:"repo"`
	CloneURL string `json:"clone_url"`
}

// PageInfoResponse - пагинация.
type PageInfoResponse struct {
	HasNextPage bool    `json:"has_next_page"`
	NextCursor  *string `json:"next_cursor,omitempty"`
}

// ErrorResponse - ответ API с ошибкой.
type ErrorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// NewClient - конструктор HTTP-клиента с cookie jar.
func NewClient(baseURL string) *Client {
	jar, err := cookiejar.New(nil)
	if err != nil {
		panic(err)
	}
	return &Client{
		baseURL: baseURL,
		http:    &http.Client{Jar: jar, Timeout: 15 * time.Second},
	}
}

// PostJSON - POST-запрос с JSON-телом.
func (c *Client) PostJSON(t *testing.T, path string, body any) *Response {
	t.Helper()
	return c.doJSON(t, http.MethodPost, path, "", body)
}

// PostAuthJSON - POST-запрос с JWT и JSON-телом.
func (c *Client) PostAuthJSON(t *testing.T, token, path string, body any) *Response {
	t.Helper()
	return c.doJSON(t, http.MethodPost, path, token, body)
}

// Get - GET-запрос без авторизации.
func (c *Client) Get(t *testing.T, path string) *Response {
	t.Helper()
	return c.doJSON(t, http.MethodGet, path, "", nil)
}

// GetAuth - GET-запрос с JWT.
func (c *Client) GetAuth(t *testing.T, token, path string) *Response {
	t.Helper()
	return c.doJSON(t, http.MethodGet, path, token, nil)
}

// DeleteAuth - DELETE-запрос с JWT.
func (c *Client) DeleteAuth(t *testing.T, token, path string) *Response {
	t.Helper()
	return c.doJSON(t, http.MethodDelete, path, token, nil)
}

// PatchAuthJSON - PATCH-запрос с JWT и JSON-телом.
func (c *Client) PatchAuthJSON(t *testing.T, token, path string, body any) *Response {
	t.Helper()
	return c.doJSON(t, http.MethodPatch, path, token, body)
}

func (c *Client) doJSON(t *testing.T, method, path, token string, body any) *Response {
	t.Helper()

	var reader io.Reader
	if body != nil {
		payload, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("marshal request: %v", err)
		}
		reader = bytes.NewReader(payload)
	}

	req, err := http.NewRequest(method, c.baseURL+path, reader)
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		t.Fatalf("do request %s %s: %v", method, path, err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read response: %v", err)
	}

	return &Response{StatusCode: resp.StatusCode, Header: resp.Header.Clone(), Body: respBody}
}

// ExpectStatus - проверка HTTP-статуса ответа.
func ExpectStatus(t *testing.T, resp *Response, want int) {
	t.Helper()
	if resp.StatusCode != want {
		t.Fatalf("status = %d, want %d, body = %s", resp.StatusCode, want, string(resp.Body))
	}
}

// Decode - декодирование JSON-ответа в типизированную структуру.
func Decode[T any](t *testing.T, resp *Response) T {
	t.Helper()
	var out T
	if err := json.Unmarshal(resp.Body, &out); err != nil {
		t.Fatalf("decode response %q: %v", string(resp.Body), err)
	}
	return out
}
