package middleware

import (
	"net/http"

	httpserver "users/internal/http/gen"

	"github.com/labstack/echo/v4"
)

func ServiceToken(expected string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			token := c.Request().Header.Get("X-Service-Token")
			if token == "" {
				return c.JSON(http.StatusUnauthorized, httpserver.ErrorResponse{
					Code:    httpserver.MISSINGAUTHTOKEN,
					Message: "missing service token",
				})
			}
			if token != expected {
				return c.JSON(http.StatusUnauthorized, httpserver.ErrorResponse{
					Code:    httpserver.INVALIDAUTHTOKEN,
					Message: "invalid service token",
				})
			}
			return next(c)
		}
	}
}
