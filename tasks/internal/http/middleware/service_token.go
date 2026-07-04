// Package middleware
package middleware

import (
	"crypto/subtle"
	"net/http"
	"strings"

	httpserver "tasks/internal/http/gen"

	"github.com/labstack/echo/v4"
)

// ServiceToken - middleware для проверки статического сервисного токена.
// Используется для внутренних service-to-service вызовов (например POST /reports от notifications).
func ServiceToken(expected string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			header := c.Request().Header.Get(echo.HeaderAuthorization)
			parts := strings.Fields(header)

			if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") ||
				subtle.ConstantTimeCompare([]byte(parts[1]), []byte(expected)) != 1 {
				return c.JSON(http.StatusUnauthorized, httpserver.ErrorResponse{
					Code:    httpserver.INVALIDAUTHTOKEN,
					Message: "invalid service token",
				})
			}

			return next(c)
		}
	}
}
