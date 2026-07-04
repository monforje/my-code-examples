package logger

import (
	"log/slog"

	"github.com/labstack/echo/v4"
	echomw "github.com/labstack/echo/v4/middleware"
)

func RequestLogMiddleware(log *Logger) echo.MiddlewareFunc {
	return echomw.RequestLoggerWithConfig(echomw.RequestLoggerConfig{
		LogLatency:   true,
		LogRemoteIP:  true,
		LogMethod:    true,
		LogURI:       true,
		LogRoutePath: true,
		LogRequestID: true,
		LogUserAgent: true,
		LogStatus:    true,
		LogHost:      true,
		HandleError:  true,
		LogValuesFunc: func(c echo.Context, v echomw.RequestLoggerValues) error {
			level := slog.LevelInfo
			msg := "http request"
			if v.Status >= 500 {
				level = slog.LevelError
				msg = "http request failed"
			} else if v.Status >= 400 {
				level = slog.LevelWarn
				msg = "http request rejected"
			}
			log.Logger.LogAttrs(c.Request().Context(), level, msg,
				slog.String("request_id", v.RequestID),
				slog.String("method", v.Method),
				slog.String("uri", v.URI),
				slog.String("route", v.RoutePath),
				slog.Int("status", v.Status),
				slog.String("latency", v.Latency.String()),
				slog.String("remote_ip", v.RemoteIP),
				slog.String("user_agent", v.UserAgent),
			)
			return nil
		},
	})
}
