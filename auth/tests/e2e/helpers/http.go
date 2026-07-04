package e2e_test_helpers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"sync/atomic"
	"testing"
	"time"
)

const (
	SubjectRegisterCode       = "notification.email.verification_code.send"
	SubjectPasswordResetCode  = "notification.email.password_reset_code.send"
	SubjectPasswordChangeCode = "notification.email.password_change_code.send"
	SubjectEmailChangeCode    = "notification.email.email_change_code.send"

	PurposeRegister           = "register"
	PurposePasswordForgot     = "password_forgot"
	PurposePasswordChange     = "password_change"
	PurposeEmailChangeCurrent = "email_change_current"
	PurposeEmailChangeNew     = "email_change_new"
)

var emailSeq uint64

type Client struct {
	baseURL string
	http    *http.Client
}

type Response struct {
	StatusCode int
	Header     http.Header
	Body       []byte
}

type TokenResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int32  `json:"expires_in"`
}

type RegisterResponse struct {
	IdentityID string `json:"identity_id"`
	Email      string `json:"email"`
	Status     string `json:"status"`
}

type MessageResponse struct {
	Message string `json:"message"`
}

type IdentityResponse struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	EmailVerified bool   `json:"email_verified"`
	Status        string `json:"status"`
}

type ResetTokenResponse struct {
	ResetToken string `json:"reset_token"`
	ExpiresIn  int32  `json:"expires_in"`
}

type ChangePasswordTokenResponse struct {
	ChangeToken string `json:"change_token"`
	ExpiresIn   int32  `json:"expires_in"`
}

type IdentityTokenResponse struct {
	IdentityToken string `json:"identity_token"`
	ExpiresIn     int32  `json:"expires_in"`
}

type DeviceStartResponse struct {
	DeviceCode      string `json:"device_code"`
	UserCode        string `json:"user_code"`
	VerificationUrl string `json:"verification_url"`
	ExpiresIn       int32  `json:"expires_in"`
	Interval        int32  `json:"interval"`
}

type DeviceConfirmResponse struct {
	Status string `json:"status"`
}

type CliTokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int32  `json:"expires_in"`
	TokenType    string `json:"token_type"`
}

type CliRefreshResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int32  `json:"expires_in"`
	TokenType    string `json:"token_type"`
}

type User struct {
	Email    string
	Password string
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

func NewUser() User {
	seq := atomic.AddUint64(&emailSeq, 1)
	return User{
		Email:    fmt.Sprintf("e2e-%d-%d@example.com", time.Now().UnixNano(), seq),
		Password: "oldpass1word",
	}
}

func (c *Client) PostJSON(t *testing.T, path string, body any) *Response {
	t.Helper()
	return c.doJSON(t, http.MethodPost, path, "", body)
}

func (c *Client) PostAuthJSON(t *testing.T, token, path string, body any) *Response {
	t.Helper()
	return c.doJSON(t, http.MethodPost, path, token, body)
}

func (c *Client) Get(t *testing.T, path string) *Response {
	t.Helper()
	return c.doJSON(t, http.MethodGet, path, "", nil)
}

func (c *Client) GetAuth(t *testing.T, token, path string) *Response {
	t.Helper()
	return c.doJSON(t, http.MethodGet, path, token, nil)
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

func WaitForCode(t *testing.T, env *Environment, subject, email, purpose string, trigger func()) string {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	code, err := env.WaitCode(ctx, subject, email, purpose, func() error {
		trigger()
		return nil
	})
	if err != nil {
		t.Fatalf("wait code subject=%s email=%s purpose=%s: %v", subject, email, purpose, err)
	}
	return code
}

func Register(t *testing.T, c *Client, user User) RegisterResponse {
	t.Helper()
	resp := c.PostJSON(t, "/auth/register", map[string]string{"email": user.Email, "password": user.Password})
	ExpectStatus(t, resp, http.StatusCreated)
	return Decode[RegisterResponse](t, resp)
}

func VerifyRegister(t *testing.T, c *Client, email, code string) {
	t.Helper()
	resp := c.PostJSON(t, "/auth/register/verify", map[string]string{"email": email, "code": code})
	ExpectStatus(t, resp, http.StatusOK)
}

func RegisterAndVerify(t *testing.T, env *Environment, c *Client) User {
	t.Helper()
	user := NewUser()
	code := WaitForCode(t, env, SubjectRegisterCode, user.Email, PurposeRegister, func() {
		Register(t, c, user)
	})
	VerifyRegister(t, c, user.Email, code)
	return user
}

func Login(t *testing.T, c *Client, user User) TokenResponse {
	t.Helper()
	resp := c.PostJSON(t, "/auth/login", map[string]string{"email": user.Email, "password": user.Password})
	ExpectStatus(t, resp, http.StatusOK)
	return Decode[TokenResponse](t, resp)
}

func LoginAs(t *testing.T, env *Environment, c *Client) (User, string) {
	t.Helper()
	user := RegisterAndVerify(t, env, c)
	token := Login(t, c, user)
	if token.AccessToken == "" {
		t.Fatal("access token is empty")
	}
	return user, token.AccessToken
}
