package logger

import (
	"context"
	"log/slog"
)

// ILogger 日志接口，便于单测注入 mock 或 no-op
type ILogger interface {
	InfoContext(ctx context.Context, msg string, args ...any)
	ErrorContext(ctx context.Context, msg string, args ...any)
	WarnContext(ctx context.Context, msg string, args ...any)
	Log(ctx context.Context, level slog.Level, msg string, args ...any)
}
