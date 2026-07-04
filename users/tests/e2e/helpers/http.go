package e2e_test_helpers

import (
	"bytes"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/cookiejar"
	"testing"
	"time"
)

type Client struct {
	baseURL string
	http    *http.Client
}

type Response struct {
	StatusCode int
	Header     http.Header
	Body       []byte
}

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

func GetAuth(t *testing.T, c *Client, token, path string) *Response {
	t.Helper()
	return c.doRequest(t, http.MethodGet, path, token, "application/json", nil)
}

func PatchAuthJSON(t *testing.T, c *Client, token, path string, body any) *Response {
	t.Helper()
	return c.doRequest(t, http.MethodPatch, path, token, "application/json", body)
}

func DeleteAuth(t *testing.T, c *Client, token, path string) *Response {
	t.Helper()
	return c.doRequest(t, http.MethodDelete, path, token, "application/json", nil)
}

func (c *Client) doRequest(t *testing.T, method, path, token, contentType string, body any) *Response {
	t.Helper()

	var reader io.Reader
	if body != nil {
		switch v := body.(type) {
		case io.Reader:
			reader = v
		default:
			payload, err := json.Marshal(body)
			if err != nil {
				t.Fatalf("marshal request: %v", err)
			}
			reader = bytes.NewReader(payload)
		}
	}

	req, err := http.NewRequest(method, c.baseURL+path, reader)
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	if body != nil && contentType != "" {
		req.Header.Set("Content-Type", contentType)
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

func UploadAvatar(t *testing.T, c *Client, token, path, fieldName, filename string, fileContent []byte) *Response {
	t.Helper()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	if fileContent != nil {
		part, err := writer.CreateFormFile(fieldName, filename)
		if err != nil {
			t.Fatalf("create form file: %v", err)
		}
		if _, err := part.Write(fileContent); err != nil {
			t.Fatalf("write file content: %v", err)
		}
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("close multipart writer: %v", err)
	}

	return c.doRequest(t, http.MethodPut, path, token, writer.FormDataContentType(), bytes.NewReader(body.Bytes()))
}

func ServiceTokenGet(t *testing.T, c *Client, serviceToken, path string, body any) *Response {
	t.Helper()

	var reader io.Reader
	if body != nil {
		payload, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("marshal request: %v", err)
		}
		reader = bytes.NewReader(payload)
	}

	req, err := http.NewRequest(http.MethodGet, c.baseURL+path, reader)
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if serviceToken != "" {
		req.Header.Set("X-Service-Token", serviceToken)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		t.Fatalf("do request GET %s: %v", path, err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read response: %v", err)
	}

	return &Response{StatusCode: resp.StatusCode, Header: resp.Header.Clone(), Body: respBody}
}

func ExpectStatus(t *testing.T, resp *Response, want int) {
	t.Helper()
	if resp.StatusCode != want {
		t.Fatalf("status = %d, want %d, body = %s", resp.StatusCode, want, string(resp.Body))
	}
}

func Decode[T any](t *testing.T, resp *Response) T {
	t.Helper()
	var out T
	if err := json.Unmarshal(resp.Body, &out); err != nil {
		t.Fatalf("decode response %q: %v", string(resp.Body), err)
	}
	return out
}
