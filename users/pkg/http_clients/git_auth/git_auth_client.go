// Package gitauthclient
package gitauthclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
	"users/internal/config"
	clientsdto "users/pkg/http_clients/dto"
)

// Все запросы с Authorization: Bearer d7cDSzPleedFECuJbksoeRNdiKAZZjFmm_6GDGv1H-E
/*
	TOKEN="d7cDSzPleedFECuJbksoeRNdiKAZZjFmm_6GDGv1H-E"
	API="http://codurity.ai/api/v1/task_runner"
*/

type GitAuthClient struct {
	token   string
	baseURL string
	http    *http.Client
}

func NewGitAuthClient(cfg config.GitAuthClientConfig) *GitAuthClient {
	return &GitAuthClient{
		token:   cfg.Token,
		baseURL: cfg.BaseURL,
		http: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// POST /register

func (c *GitAuthClient) RegisterGitUser(
	ctx context.Context, req *clientsdto.RegisterGitUserRequest,
) (*clientsdto.RegisterGitUserResponse, error) {
	/*
	   curl -X POST "$API/register" \
	     -H "Authorization: Bearer $TOKEN" \
	     -H "Content-Type: application/json" \
	     -d '{"username":"alice","email":"alice@example.com"}'
	   # → {"username":"alice","token":"<GIT_TOKEN>","git_url":"http://gitea.local"}
	*/
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal register git user request: %w", err)
	}

	endpoint, err := url.JoinPath(c.baseURL, "register")
	if err != nil {
		return nil, fmt.Errorf("build register git user url: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create register git user request: %w", err)
	}
	httpReq.Header.Set("Authorization", "Bearer "+c.token)
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("send register git user request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read register git user response: %w", err)
	}

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return nil, fmt.Errorf("register git user failed: status=%d body=%s", resp.StatusCode, string(respBody))
	}

	var out clientsdto.RegisterGitUserResponse
	if err := json.Unmarshal(respBody, &out); err != nil {
		return nil, fmt.Errorf("decode register git user response: %w", err)
	}

	return &out, nil
}
