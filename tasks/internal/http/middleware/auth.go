// Package middleware
package middleware

import (
	"errors"
	"net/http"
	"strings"

	httpserver "tasks/internal/http/gen"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"tasks/internal/authctx"
)

type TokenValidator interface {
	ValidateAccessToken(tokenString string) (uuid.UUID, uuid.UUID, string, error)
}

func BearerAuth(validator TokenValidator) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			header := c.Request().Header.Get(echo.HeaderAuthorization)
			if header == "" {
				return c.JSON(http.StatusUnauthorized, httpserver.ErrorResponse{Code: httpserver.MISSINGAUTHTOKEN, Message: "missing auth token"})
			}

			parts := strings.Fields(header)
			if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
				return c.JSON(http.StatusUnauthorized, httpserver.ErrorResponse{Code: httpserver.INVALIDAUTHTOKEN, Message: "invalid auth token"})
			}

			identityID, sessionID, _, err := validator.ValidateAccessToken(parts[1])
			if err != nil {
				code := httpserver.INVALIDAUTHTOKEN
				message := "invalid auth token"
				if errors.Is(err, jwt.ErrTokenExpired) {
					code = httpserver.EXPIREDAUTHTOKEN
					message = "expired auth token"
				}
				return c.JSON(http.StatusUnauthorized, httpserver.ErrorResponse{Code: code, Message: message})
			}

			ctx := authctx.WithAuth(c.Request().Context(), identityID, sessionID)
			c.SetRequest(c.Request().WithContext(ctx))
			return next(c)
		}
	}
}
