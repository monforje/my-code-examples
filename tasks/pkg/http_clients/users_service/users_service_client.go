// Package usersserviceclient
package usersserviceclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"tasks/internal/config"
	clientsdto "tasks/pkg/http_clients/dto"
	"time"
)

// Все запросы с X-Service-Token: DIa47jw8qFVregS52SeDcdkWxWUFCERVZu9kp3ZSWDljmIXx8ofEUU8reEk3sUJ9
/*
	TOKEN="DIa47jw8qFVregS52SeDcdkWxWUFCERVZu9kp3ZSWDljmIXx8ofEUU8reEk3sUJ9"
	API="http://users_http:8080/api/v1/git-user/me"
*/

type UsersClient struct {
	token   string
	baseURL string
	http    *http.Client
}

func NewUsersClient(cfg config.UsersClientConfig) *UsersClient {
	return &UsersClient{
		token:   cfg.Token,
		baseURL: cfg.BaseURL,
		http: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (c *UsersClient) GetGitUser(ctx context.Context, req *clientsdto.GitUserRequest) (*clientsdto.GitUserResponse, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/api/v1/git-user/me", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("X-Service-Token", c.token)

	resp, err := c.http.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(respBody))
	}

	var result clientsdto.GitUserResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("unmarshal response: %w", err)
	}

	return &result, nil
}
