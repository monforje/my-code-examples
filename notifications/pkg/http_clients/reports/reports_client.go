// Package reportsclient - HTTP-клиент для отправки CI-отчётов в сервис tasks.
package reportsclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"notifications/internal/config"
	clientsdto "notifications/pkg/http_clients/dto"
)

type ReportsClient struct {
	token   string
	baseURL string
	http    *http.Client
}

func NewReportsClient(cfg config.ReportsClientConfig) *ReportsClient {
	return &ReportsClient{
		token:   cfg.Token,
		baseURL: cfg.BaseURL,
		http: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// SendReport - отправляет CI-отчёт в сервис tasks (POST /reports).
/*
   curl -X POST "$BASE_URL/reports" \
     -H "Authorization: Bearer $TOKEN" \
     -H "Content-Type: application/json" \
     -d '{...}'
*/
func (c *ReportsClient) SendReport(ctx context.Context, req *clientsdto.CreateReportRequest) error {
	body, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("marshal report: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/reports", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	if c.token != "" {
		httpReq.Header.Set("Authorization", "Bearer "+c.token)
	}

	resp, err := c.http.Do(httpReq)
	if err != nil {
		return fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(respBody))
	}

	return nil
}
