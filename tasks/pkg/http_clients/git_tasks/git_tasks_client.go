// Package gittasksclient - HTTP клиент для взаимодействия с Git Tasks API (taskrunner).
package gittasksclient

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

// Все запросы с Authorization: Bearer d7cDSzPleedFECuJbksoeRNdiKAZZjFmm_6GDGv1H-E
/*
	TOKEN="d7cDSzPleedFECuJbksoeRNdiKAZZjFmm_6GDGv1H-E"
	API="http://taskrunner:8000/api/v1/task_runner"
*/

type GitAuthClient struct {
	token   string
	baseURL string
	http    *http.Client
}

func NewGitAuthClient(cfg config.GitTasksClientConfig) *GitAuthClient {
	return &GitAuthClient{
		token:   cfg.Token,
		baseURL: cfg.BaseURL,
		http: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// GitTaskCreate - Создаёт/подготавливает task repo и возвращает clone_url.
func (c *GitAuthClient) GitTaskCreate(
	ctx context.Context, req *clientsdto.GitTaskCreateRequest,
) (*clientsdto.GitTaskCreateResponse, error) {
	/*
	   curl -X POST "$API/tasks" \
	     -H "Authorization: Bearer $TOKEN" \
	     -H "Content-Type: application/json" \
	     -d '{"username":"alice","task_id":"pizza-api"}'
	   # → {"task_id":"pizza-api","repo":"alice/golden-pizza-api","clone_url":"http://gitea.local/alice/golden-pizza-api.git"}
	*/
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/tasks", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+c.token)

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

	var result clientsdto.GitTaskCreateResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("unmarshal response: %w", err)
	}

	return &result, nil
}
