package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func (c *Client) GetRepo(ctx context.Context, owner, name string) (*RepoResponse, error) {
	url := fmt.Sprintf("%s/repos/%s/%s", c.BaseURL, owner, name)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Accept", "application/vnd.github.v3+json")
	if c.Token != "" {
		req.Header.Set("Authorization", "Bearer "+c.Token)
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errResp ErrorResponse
		if err := json.NewDecoder(resp.Body).Decode(&errResp); err != nil {
			return nil, fmt.Errorf("unexpected status %d", resp.StatusCode)
		}
		return nil, fmt.Errorf("github api: %s", errResp.Message)
	}

	var repo RepoResponse
	if err := json.NewDecoder(resp.Body).Decode(&repo); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}
	return &repo, nil
}
