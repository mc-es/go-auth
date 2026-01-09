package main

import (
	"context"

	"go-auth/pkg/logger"
	_ "go-auth/pkg/logger/adapter/logrus"
	_ "go-auth/pkg/logger/adapter/zap"
)

type contextKey string

const (
	traceIDKey contextKey = "trace_id"
	userIDKey  contextKey = "user_id"
)

func main() {
	log, err := logger.New(
		logger.WithDriver(logger.DriverZap),
		logger.WithDevelopment(),
		logger.WithContextExtractor(extractor),
	)
	if err != nil {
		panic(err)
	}

	defer func() { _ = log.Sync() }()

	ctx := context.Background()
	ctx = context.WithValue(ctx, traceIDKey, "test-12345")
	ctx = context.WithValue(ctx, userIDKey, 12345)
	log.InfoCtx(ctx, "Hello, World!")
}

func extractor(ctx context.Context) []logger.Attr {
	var attrs []logger.Attr

	if traceID, ok := ctx.Value(traceIDKey).(string); ok {
		attrs = append(attrs, logger.Str(string(traceIDKey), traceID))
	}

	if userID, ok := ctx.Value(userIDKey).(int); ok {
		attrs = append(attrs, logger.Int(string(userIDKey), userID))
	}

	return attrs
}
