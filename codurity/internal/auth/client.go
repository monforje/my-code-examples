package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// APIError описывает ошибку, возвращённую auth-сервисом.
type APIError struct {
	StatusCode int
	Code       string
	Message    string
}

func (e *APIError) Error() string {
	if e.Code != "" {
		return fmt.Sprintf("auth api: %s (%d): %s", e.Code, e.StatusCode, e.Message)
	}
	return fmt.Sprintf("auth api: status %d: %s", e.StatusCode, e.Message)
}

// Client — HTTP-клиент к auth-сервису Codurity.
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

// StartDeviceAuth запускает device authorization flow.
func (c *Client) StartDeviceAuth(ctx context.Context) (*DeviceStartResponse, error) {
	resp, err := c.do(ctx, http.MethodPost, "/device/start", nil, "")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, decodeError(resp)
	}
	var out DeviceStartResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}
	return &out, nil
}

// PollDeviceToken опрашивает результат подтверждения CLI-входа.
// Возвращает *APIError со StatusCode 428, если подтверждение ещё ожидается.
func (c *Client) PollDeviceToken(ctx context.Context, deviceCode string) (*CliTokenResponse, error) {
	resp, err := c.do(ctx, http.MethodPost, "/device/token", DeviceTokenRequest{DeviceCode: deviceCode}, "")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, decodeError(resp)
	}
	var out CliTokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}
	return &out, nil
}

// Logout завершает сессию на сервере (POST /auth/logout, Bearer).
func (c *Client) Logout(ctx context.Context, accessToken string) error {
	resp, err := c.do(ctx, http.MethodPost, "/logout", nil, accessToken)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		return decodeError(resp)
	}
	return nil
}

// Refresh обновляет access token через refresh token.
func (c *Client) Refresh(ctx context.Context, refreshToken string) (*CliRefreshResponse, error) {
	resp, err := c.do(ctx, http.MethodPost, "/cli/refresh", CliRefreshRequest{RefreshToken: refreshToken}, "")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, decodeError(resp)
	}
	var out CliRefreshResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}
	return &out, nil
}
