package auth

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"codurity/internal/build"

	"github.com/briandowns/spinner"
	"github.com/pkg/browser"
)

// Login выполняет device authorization flow:
// start → вывод кода/URL → открытие браузера → poll → сохранение токенов.
func Login(ctx context.Context, c *Client, out io.Writer) (*Auth, error) {
	start, err := c.StartDeviceAuth(ctx)
	if err != nil {
		return nil, fmt.Errorf("start device auth: %w", err)
	}

	loginURL := start.VerificationURL
	if loginURL == "" {
		loginURL = strings.TrimRight(build.FrontendURL, "/") + "/cli/login"
	}
	code := FormatUserCode(start.UserCode)

	fmt.Fprintln(out, "Откройте браузер:")
	fmt.Fprintln(out)
	fmt.Fprintln(out, loginURL)
	fmt.Fprintln(out)
	fmt.Fprintln(out, "Введите код:")
	fmt.Fprintln(out)
	fmt.Fprintln(out, code)
	fmt.Fprintln(out)

	// Открытие браузера — best-effort (в headless-окружении молча игнорируем).
	_ = browser.OpenURL(loginURL)

	interval := time.Duration(start.Interval) * time.Second
	if interval <= 0 {
		interval = 5 * time.Second
	}
	deadline := time.Now().Add(time.Duration(start.ExpiresIn) * time.Second)

	s := spinner.New(spinner.CharSets[11], 100*time.Millisecond)
	s.Suffix = " Ожидаем подтверждения..."
	s.Start()
	token, err := poll(ctx, c, start.DeviceCode, interval, deadline)
	s.Stop()
	fmt.Fprintln(out)

	if err != nil {
		return nil, err
	}

	a := &Auth{
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
		ExpiresAt:    time.Now().Add(time.Duration(token.ExpiresIn) * time.Second).UTC(),
	}
	if err := a.Save(); err != nil {
		return nil, fmt.Errorf("save auth: %w", err)
	}

	fmt.Fprintln(out, "Успешный вход выполнен.")
	return a, nil
}

func poll(ctx context.Context, c *Client, deviceCode string, interval time.Duration, deadline time.Time) (*CliTokenResponse, error) {
	for {
		if time.Now().After(deadline) {
			return nil, errors.New("время ожидания подтверждения истекло")
		}
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		tok, err := c.PollDeviceToken(ctx, deviceCode)
		if err == nil {
			return tok, nil
		}

		var ae *APIError
		if errors.As(err, &ae) {
			// 428 — подтверждение ещё ожидается, продолжаем опрос.
			if ae.StatusCode != http.StatusPreconditionRequired {
				// 404 (код не найден/истёк) и прочие ошибки — выходим.
				return nil, err
			}
		}
		// транзитная сетевая ошибка или pending — ждём и повторяем.

		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(interval):
		}
	}
}
