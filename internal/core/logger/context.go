package logger

import (
	"context"
	"log/slog"
)

type ctxKey struct{}

var loggerKey = ctxKey{}

func ToContext(ctx context.Context, log *slog.Logger) context.Context {
	return context.WithValue(ctx, loggerKey, log)
}

func FromContext(ctx context.Context) *slog.Logger {
	if log, ok := ctx.Value(loggerKey).(*slog.Logger); ok {
		return log
	}
	return slog.Default()
}
