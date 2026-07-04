package handlers

import (
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
)

const (
	refreshCookieName = "refresh_token"
	refreshCookiePath = "/api/v1/auth"
)

func setRefreshCookie(ctx echo.Context, token string, maxAge time.Duration) {
	secure := isSecureCookie(ctx)
	cookie := &http.Cookie{
		Name:     refreshCookieName,
		Value:    token,
		Path:     refreshCookiePath,
		MaxAge:   int(maxAge.Seconds()),
		HttpOnly: true,
		Secure:   secure,
		SameSite: sameSiteMode(secure),
	}
	ctx.SetCookie(cookie)
}

func clearRefreshCookie(ctx echo.Context) {
	secure := isSecureCookie(ctx)
	cookie := &http.Cookie{
		Name:     refreshCookieName,
		Value:    "",
		Path:     refreshCookiePath,
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   secure,
		SameSite: sameSiteMode(secure),
	}
	ctx.SetCookie(cookie)
}

func sameSiteMode(secure bool) http.SameSite {
	if secure {
		return http.SameSiteNoneMode
	}
	return http.SameSiteLaxMode
}

// isSecureCookie определяет Secure флаг по схеме запроса.
// В production (HTTPS) — true, в dev/test (HTTP) — false.
func isSecureCookie(ctx echo.Context) bool {
	return ctx.Request().TLS != nil || ctx.Request().Header.Get("X-Forwarded-Proto") == "https"
}
