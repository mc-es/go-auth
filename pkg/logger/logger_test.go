package logger_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go-auth/pkg/logger"
	// Adapters must be imported to register factories.
	_ "go-auth/pkg/logger/adapter/logrus"
	_ "go-auth/pkg/logger/adapter/zap"
	"go-auth/pkg/logger/internal/core"
)

func TestNewWithDefaults(t *testing.T) {
	log, err := logger.New()
	require.NoError(t, err)
	assert.NotNil(t, log)
}

func TestNewWithOptions(t *testing.T) {
	log, err := logger.New(
		logger.WithDriver(logger.DriverLogrus),
		logger.WithLevel(logger.LevelInfo),
		logger.WithFormat(logger.FormatText),
		logger.WithTimeLayout(logger.TimeLayoutTimeOnly),
		logger.WithOutputPaths("stdout"),
	)
	require.NoError(t, err)
	assert.NotNil(t, log)
}

func TestNewWithValidationErrors(t *testing.T) {
	tests := []struct {
		name        string
		opts        []logger.Option
		expectedErr error
	}{
		{
			name: "missing driver",
			opts: []logger.Option{
				logger.WithDriver(logger.Driver("")), // Empty driver
			},
			expectedErr: core.ErrMissingDriver,
		},
		{
			name: "unknown driver",
			opts: []logger.Option{
				logger.WithDriver(logger.Driver("unknown")),
			},
			expectedErr: core.ErrUnknownDriver,
		},
		{
			name: "invalid level",
			opts: []logger.Option{
				logger.WithLevel(logger.Level(99)), // Invalid level
			},
			expectedErr: core.ErrInvalidLevel,
		},
		{
			name: "invalid format",
			opts: []logger.Option{
				logger.WithFormat(logger.Format("xml")), // Invalid format
			},
			expectedErr: core.ErrInvalidFormat,
		},
		{
			name: "invalid time layout",
			opts: []logger.Option{
				logger.WithTimeLayout(logger.TimeLayout("")), // Empty time layout
			},
			expectedErr: core.ErrInvalidTimeLayout,
		},
		{
			name: "empty output paths",
			opts: []logger.Option{
				logger.WithOutputPaths(), // Empty paths
			},
			expectedErr: core.ErrInvalidPaths,
		},
		{
			name: "invalid output paths",
			opts: []logger.Option{
				logger.WithOutputPaths(""), // Empty paths
			},
			expectedErr: core.ErrInvalidPaths,
		},
		{
			name: "invalid file rotation",
			opts: []logger.Option{
				logger.WithFileRotation(0, -1, -100, false, false), // Invalid file rotation
			},
			expectedErr: core.ErrInvalidFileRotation,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			log, err := logger.New(tt.opts...)
			assert.ErrorIs(t, err, tt.expectedErr)
			assert.Nil(t, log)
		})
	}
}

func TestNewWithDevelopment(t *testing.T) {
	log, err := logger.New(logger.WithDevelopment())
	require.NoError(t, err)
	assert.NotNil(t, log)
}

func TestNewWithContextExtraction(t *testing.T) {
	// Create a temp file for logging
	tempDir := t.TempDir()
	logFile := filepath.Join(tempDir, "test.log")

	// Define extractor
	extractor := func(ctx context.Context) []any {
		if uid, ok := ctx.Value("user_id").(int); ok {
			return []any{"user_id", uid}
		}

		return nil
	}

	// Initialize logger with zap
	log, err := logger.New(
		logger.WithDriver(logger.DriverZap),
		logger.WithOutputPaths(logFile),
		logger.WithContextExtractor(extractor),
		logger.WithLevel(logger.LevelInfo),
	)
	require.NoError(t, err)

	// Create context with value
	ctx := context.WithValue(context.Background(), "user_id", 42)

	// Log with context
	log.InfoCtx(ctx, "test message", "foo", "bar")

	// Sync to ensure flush
	_ = log.Sync()

	// Read file content
	content, err := os.ReadFile(logFile)
	require.NoError(t, err)

	logStr := string(content)

	// Verify content
	assert.Contains(t, logStr, `"msg":"test message"`)
	assert.Contains(t, logStr, `"user_id":42`) // Extracted from context
	assert.Contains(t, logStr, `"foo":"bar"`)
}
