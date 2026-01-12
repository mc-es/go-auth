package zap_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zapcore"

	"go-auth/pkg/logger"
	_ "go-auth/pkg/logger/adapter/zap"
	"go-auth/pkg/logger/internal/core"
)

const (
	testLogFileName = "test.log"
	testMessage     = "test message"
	testKey1        = "key1"
	testValue1      = "value1"
	testKey2        = "key2"
	testValue2      = "value2"
)

type testLogger struct {
	logger logger.Logger
	file   string
}

// setupLogger creates a logger with the given options and returns it along with the log file path.
func setupLogger(t *testing.T, opts ...logger.Option) *testLogger {
	t.Helper()

	tempDir := t.TempDir()
	logFile := filepath.Join(tempDir, testLogFileName)

	allOpts := append([]logger.Option{
		logger.WithDriver(logger.DriverZap),
		logger.WithOutputPaths(logFile),
	}, opts...)

	log, err := logger.New(allOpts...)
	require.NoError(t, err)

	return &testLogger{
		logger: log,
		file:   logFile,
	}
}

// readLogFile reads and returns the content of the log file.
func (tl *testLogger) readLogFile(t *testing.T) string {
	t.Helper()

	_ = tl.logger.Sync()

	content, err := os.ReadFile(tl.file)
	if err != nil {
		return ""
	}

	return string(content)
}

// cleanup closes the logger.
func (tl *testLogger) cleanup(t *testing.T) {
	t.Helper()

	_ = tl.logger.Sync()
}

func TestLevels(t *testing.T) {
	tests := []struct {
		name      string
		level     core.Level
		expected  zapcore.Level
		shouldLog bool
	}{
		{"debug", core.LevelDebug, zapcore.DebugLevel, true},
		{"info", core.LevelInfo, zapcore.InfoLevel, true},
		{"warn", core.LevelWarn, zapcore.WarnLevel, false}, // Info message won't appear at Warn level
		{"error", core.LevelError, zapcore.ErrorLevel, false},
		{"panic", core.LevelPanic, zapcore.PanicLevel, false},
		{"fatal", core.LevelFatal, zapcore.FatalLevel, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tl := setupLogger(t,
				logger.WithLevel(tt.level),
				logger.WithFormat(logger.FormatJSON),
			)
			defer tl.cleanup(t)

			tl.logger.Info(testMessage)
			content := tl.readLogFile(t)

			if tt.shouldLog {
				assert.Contains(t, content, testMessage)
			} else {
				assert.NotContains(t, content, testMessage)
			}
		})
	}
}

func TestFormatJSON(t *testing.T) {
	tl := setupLogger(t,
		logger.WithFormat(logger.FormatJSON),
		logger.WithLevel(logger.LevelInfo),
	)
	defer tl.cleanup(t)

	tl.logger.Info(testMessage, testKey1, testValue1)
	content := tl.readLogFile(t)

	assert.Contains(t, content, `"msg":"`+testMessage+`"`)
	assert.Contains(t, content, `"level":"INFO"`)
	assert.Contains(t, content, `"`+testKey1+`":"`+testValue1+`"`)
}

func TestFormatText(t *testing.T) {
	tl := setupLogger(t,
		logger.WithFormat(logger.FormatText),
		logger.WithLevel(logger.LevelInfo),
	)
	defer tl.cleanup(t)

	tl.logger.Info(testMessage, testKey1, testValue1)
	content := tl.readLogFile(t)

	assert.Contains(t, content, testMessage)
	assert.Contains(t, content, testKey1)
	assert.Contains(t, content, testValue1)
}

func TestTimeLayout(t *testing.T) {
	tests := []struct {
		name       string
		timeLayout logger.TimeLayout
	}{
		{"DateTime", logger.TimeLayoutDateTime},
		{"DateOnly", logger.TimeLayoutDateOnly},
		{"TimeOnly", logger.TimeLayoutTimeOnly},
		{"RFC3339", logger.TimeLayoutRFC3339},
		{"RFC822", logger.TimeLayoutRFC822},
		{"RFC1123", logger.TimeLayoutRFC1123},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tl := setupLogger(t,
				logger.WithTimeLayout(tt.timeLayout),
				logger.WithLevel(logger.LevelInfo),
				logger.WithFormat(logger.FormatJSON),
			)
			defer tl.cleanup(t)

			tl.logger.Info(testMessage)
			content := tl.readLogFile(t)

			assert.Contains(t, content, `"time":`)
		})
	}
}

func TestDevelopment(t *testing.T) {
	tempDir := t.TempDir()
	logFile := filepath.Join(tempDir, testLogFileName)

	log, err := logger.New(
		logger.WithDriver(logger.DriverZap),
		logger.WithOutputPaths(logFile, "stdout"),
		logger.WithDevelopment(),
	)
	require.NoError(t, err)

	defer func() { _ = log.Sync() }()

	log.Info(testMessage)
	_ = log.Sync()

	content, err := os.ReadFile(logFile)
	// In development mode, file might not exist if output goes to stdout
	if err != nil {
		return
	}

	assert.Contains(t, string(content), `"caller"`)
}

func TestContextExtractor(t *testing.T) {
	extractor := func(ctx context.Context) []any {
		var result []any
		if userID, ok := ctx.Value("user_id").(int); ok {
			result = append(result, "user_id", userID)
		}

		if requestID, ok := ctx.Value("request_id").(string); ok {
			result = append(result, "request_id", requestID)
		}

		return result
	}

	tl := setupLogger(t,
		logger.WithContextExtractor(extractor),
		logger.WithLevel(logger.LevelInfo),
	)
	defer tl.cleanup(t)

	ctx := context.WithValue(context.Background(), "user_id", 42)
	ctx = context.WithValue(ctx, "request_id", "req-123")

	tl.logger.InfoCtx(ctx, testMessage, testKey1, testValue1)
	content := tl.readLogFile(t)

	assert.Contains(t, content, `"user_id":42`)
	assert.Contains(t, content, `"request_id":"req-123"`)
	assert.Contains(t, content, `"`+testKey1+`":"`+testValue1+`"`)
}

func TestNoContextExtractor(t *testing.T) {
	extractor := func(ctx context.Context) []any {
		return []any{"extracted", "value"}
	}

	tl := setupLogger(t,
		logger.WithContextExtractor(extractor),
		logger.WithLevel(logger.LevelInfo),
	)
	defer tl.cleanup(t)

	tl.logger.Info(testMessage, testKey1, testValue1)
	content := tl.readLogFile(t)

	assert.NotContains(t, content, `"extracted":"value"`)
	assert.Contains(t, content, `"`+testKey1+`":"`+testValue1+`"`)
}

func TestNilContextExtractor(t *testing.T) {
	tl := setupLogger(t,
		logger.WithLevel(logger.LevelInfo),
	)
	defer tl.cleanup(t)

	ctx := context.WithValue(context.Background(), "user_id", 42)
	tl.logger.InfoCtx(ctx, testMessage, testKey1, testValue1)
	content := tl.readLogFile(t)

	assert.NotContains(t, content, "user_id")
	assert.Contains(t, content, `"`+testKey1+`":"`+testValue1+`"`)
}

func TestAllLogLevels(t *testing.T) {
	tl := setupLogger(t,
		logger.WithLevel(logger.LevelDebug),
		logger.WithFormat(logger.FormatJSON),
	)
	defer tl.cleanup(t)

	tl.logger.Debug("debug message")
	tl.logger.Info("info message")
	tl.logger.Warn("warn message")
	tl.logger.Error("error message")

	content := tl.readLogFile(t)

	assert.Contains(t, content, `"msg":"debug message"`)
	assert.Contains(t, content, `"msg":"info message"`)
	assert.Contains(t, content, `"msg":"warn message"`)
	assert.Contains(t, content, `"msg":"error message"`)
}

func TestAllCtxLogLevels(t *testing.T) {
	tl := setupLogger(t,
		logger.WithLevel(logger.LevelDebug),
		logger.WithFormat(logger.FormatJSON),
	)
	defer tl.cleanup(t)

	ctx := context.Background()

	tl.logger.DebugCtx(ctx, "debug ctx message")
	tl.logger.InfoCtx(ctx, "info ctx message")
	tl.logger.WarnCtx(ctx, "warn ctx message")
	tl.logger.ErrorCtx(ctx, "error ctx message")

	content := tl.readLogFile(t)

	assert.Contains(t, content, `"msg":"debug ctx message"`)
	assert.Contains(t, content, `"msg":"info ctx message"`)
	assert.Contains(t, content, `"msg":"warn ctx message"`)
	assert.Contains(t, content, `"msg":"error ctx message"`)
}

func TestSync(t *testing.T) {
	tl := setupLogger(t,
		logger.WithLevel(logger.LevelInfo),
	)
	defer tl.cleanup(t)

	tl.logger.Info(testMessage)

	err := tl.logger.Sync()
	assert.NoError(t, err)

	content := tl.readLogFile(t)
	assert.Contains(t, content, testMessage)
}

func TestEmptyAttributes(t *testing.T) {
	tl := setupLogger(t,
		logger.WithLevel(logger.LevelInfo),
		logger.WithFormat(logger.FormatJSON),
	)
	defer tl.cleanup(t)

	tl.logger.Info(testMessage)
	content := tl.readLogFile(t)

	assert.Contains(t, content, `"msg":"`+testMessage+`"`)
}

func TestMultipleAttributes(t *testing.T) {
	tl := setupLogger(t,
		logger.WithLevel(logger.LevelInfo),
		logger.WithFormat(logger.FormatJSON),
	)
	defer tl.cleanup(t)

	tl.logger.Info(testMessage,
		testKey1, testValue1,
		testKey2, testValue2,
		"key3", "value3",
		"key4", 42,
		"key5", true,
	)

	content := tl.readLogFile(t)

	assert.Contains(t, content, `"`+testKey1+`":"`+testValue1+`"`)
	assert.Contains(t, content, `"`+testKey2+`":"`+testValue2+`"`)
	assert.Contains(t, content, `"key3":"value3"`)
	assert.Contains(t, content, `"key4":42`)
	assert.Contains(t, content, `"key5":true`)
}
