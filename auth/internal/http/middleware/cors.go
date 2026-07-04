package middleware

import (
	"auth/internal/config"

	"github.com/labstack/echo/v4"
	echomw "github.com/labstack/echo/v4/middleware"
)

func CORS(cfg config.CORSConfig) echo.MiddlewareFunc {
	return echomw.CORSWithConfig(echomw.CORSConfig{
		AllowHeaders:     cfg.Headers,
		AllowOrigins:     cfg.Origins,
		AllowMethods:     cfg.Methods,
		ExposeHeaders:    cfg.ExposeHeaders,
		AllowCredentials: true,
	})
}
