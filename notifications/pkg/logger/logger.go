// Package logger
package logger

import (
	"context"
	stdlog "log"
	"log/slog"
	"notifications/internal/config"
	"os"
	"strings"
)

type Logger struct {
	*slog.Logger
}

func New(cfg *config.LoggerConfig) *Logger {
	if cfg.Output == nil {
		cfg.Output = os.Stdout
	}

	opts := &slog.HandlerOptions{
		Level: cfg.Level,
	}

	var handler slog.Handler
	if strings.ToLower(string(cfg.Format)) == "json" {
		handler = slog.NewJSONHandler(cfg.Output, opts)
	} else {
		handler = slog.NewTextHandler(cfg.Output, opts)
	}

	return &Logger{
		Logger: slog.New(handler),
	}
}

func (l *Logger) With(args ...any) *Logger {
	return &Logger{
		Logger: l.Logger.With(args...),
	}
}

func (l *Logger) AsStandardLogger(level slog.Level) *stdlog.Logger {
	return slog.NewLogLogger(l.Logger.Handler(), level)
}

func (l *Logger) Debug(ctx context.Context, op, msg string, args ...any) {
	l.Logger.DebugContext(ctx, msg, withOp(op, args)...)
}

func (l *Logger) Info(ctx context.Context, op, msg string, args ...any) {
	l.Logger.InfoContext(ctx, msg, withOp(op, args)...)
}

func (l *Logger) Warn(ctx context.Context, op, msg string, args ...any) {
	l.Logger.WarnContext(ctx, msg, withOp(op, args)...)
}

func (l *Logger) Error(ctx context.Context, op, msg string, args ...any) {
	l.Logger.ErrorContext(ctx, msg, withOp(op, args)...)
}

func withOp(op string, args []any) []any {
	if op == "" {
		return args
	}

	attrs := make([]any, 0, len(args)+1)
	attrs = append(attrs, slog.String("op", op))
	attrs = append(attrs, args...)
	return attrs
}
