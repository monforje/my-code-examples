package tasks

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type APIError struct {
	StatusCode int
	Code       string
	Message    string
}

func (e *APIError) Error() string {
	if e.Code != "" {
		return fmt.Sprintf("tasks api: %s (%d): %s", e.Code, e.StatusCode, e.Message)
	}
	return fmt.Sprintf("tasks api: status %d: %s", e.StatusCode, e.Message)
}

type Client struct {
	HTTPClient *http.Client
	BaseURL    string
}

func NewClient(baseURL string) *Client {
	return &Client{
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
			Transport: &http.Transport{
				MaxIdleConns:        100,
				IdleConnTimeout:     90 * time.Second,
				TLSHandshakeTimeout: 10 * time.Second,
			},
		},
		BaseURL: baseURL,
	}
}

func (c *Client) do(ctx context.Context, method, path string, body any, bearer string) (*http.Response, error) {
	var reader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("marshal body: %w", err)
		}
		reader = bytes.NewReader(data)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.BaseURL+path, reader)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("Accept", "application/json")
	if bearer != "" {
		req.Header.Set("Authorization", "Bearer "+bearer)
	}
	return c.HTTPClient.Do(req)
}

func decodeError(resp *http.Response) error {
	var er ErrorResponse
	if err := json.NewDecoder(resp.Body).Decode(&er); err != nil {
		return &APIError{StatusCode: resp.StatusCode, Message: resp.Status}
	}
	return &APIError{StatusCode: resp.StatusCode, Code: er.Code, Message: er.Message}
}

func (c *Client) CreateGitTask(ctx context.Context, taskName string, accessToken string) (*GitTaskResponse, error) {
	resp, err := c.do(ctx, http.MethodPost, "/tasks/"+taskName+"/git", GitTaskRequest{TaskName: taskName}, accessToken)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("tasks api: unexpected status %d: %s", resp.StatusCode, string(body))
	}

	var out GitTaskResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}
	return &out, nil
}
