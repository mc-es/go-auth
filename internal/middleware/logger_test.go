package middleware_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go-auth/internal/middleware"
	"go-auth/pkg/logger"
)

type logCall struct {
	level string
	msg   string
	attrs []any
}

type spyLogger struct {
	mu    sync.Mutex
	calls []logCall
}

func (s *spyLogger) Debug(msg string, attrs ...any) { s.record("debug", msg, attrs...) }
func (s *spyLogger) Info(msg string, attrs ...any)  { s.record("info", msg, attrs...) }
func (s *spyLogger) Warn(msg string, attrs ...any)  { s.record("warn", msg, attrs...) }
func (s *spyLogger) Error(msg string, attrs ...any) { s.record("error", msg, attrs...) }
func (s *spyLogger) Panic(msg string, attrs ...any) { s.record("panic", msg, attrs...) }
func (s *spyLogger) Fatal(msg string, attrs ...any) { s.record("fatal", msg, attrs...) }

func (s *spyLogger) DebugCtx(context.Context, string, ...any) { /* No-op */ }
func (s *spyLogger) InfoCtx(context.Context, string, ...any)  { /* No-op */ }
func (s *spyLogger) WarnCtx(context.Context, string, ...any)  { /* No-op */ }
func (s *spyLogger) ErrorCtx(context.Context, string, ...any) { /* No-op */ }
func (s *spyLogger) PanicCtx(context.Context, string, ...any) { /* No-op */ }
func (s *spyLogger) FatalCtx(context.Context, string, ...any) { /* No-op */ }

func (s *spyLogger) Named(string) logger.Logger { return s }
func (s *spyLogger) Sync() error                { return nil }

func (s *spyLogger) record(level, msg string, attrs ...any) {
	s.mu.Lock()
	defer s.mu.Unlock()

	cloned := append([]any(nil), attrs...)
	s.calls = append(s.calls, logCall{
		level: level,
		msg:   msg,
		attrs: cloned,
	})
}

func TestLogger(t *testing.T) {
	t.Parallel()

	const requestURIWithQuery = "/auth/login?foo=bar&foo=baz"

	tests := []struct {
		name          string
		requestURI    string
		status        int
		expectedLevel string
		expectQuery   bool
	}{
		{
			name:          "info for success responses",
			requestURI:    requestURIWithQuery,
			status:        http.StatusOK,
			expectedLevel: "info",
			expectQuery:   true,
		},
		{
			name:          "warn for client errors",
			requestURI:    requestURIWithQuery,
			status:        http.StatusBadRequest,
			expectedLevel: "warn",
			expectQuery:   true,
		},
		{
			name:          "error for server errors",
			requestURI:    requestURIWithQuery,
			status:        http.StatusInternalServerError,
			expectedLevel: "error",
			expectQuery:   true,
		},
		{
			name:          "omits query attribute when query is empty",
			requestURI:    "/auth/login",
			status:        http.StatusOK,
			expectedLevel: "info",
			expectQuery:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			log := &spyLogger{}
			nextCalled := false
			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				nextCalled = true

				w.WriteHeader(tt.status)
			})

			handler := middleware.Logger(log)(next)

			req := httptest.NewRequest(http.MethodGet, tt.requestURI, http.NoBody)
			req.Header.Set("User-Agent", "logger-test-agent")
			req.RemoteAddr = "203.0.113.1:12345"

			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)

			require.True(t, nextCalled)
			require.Equal(t, tt.status, rr.Code)
			require.Len(t, log.calls, 1)

			call := log.calls[0]
			assert.Equal(t, tt.expectedLevel, call.level)
			assert.Equal(t, "request", call.msg)

			attrs := attrsToMap(call.attrs)
			assert.Equal(t, http.MethodGet, attrs["method"])
			assert.Equal(t, tt.requestURI, attrs["uri"])
			assert.Equal(t, tt.status, attrs["status"])
			assert.Equal(t, "203.0.113.1", attrs["ip"])
			assert.Equal(t, "logger-test-agent", attrs["user_agent"])
			assert.Equal(t, "", attrs["request_id"])

			if tt.expectQuery {
				assert.Equal(t, url.Values{"foo": {"bar", "baz"}}, attrs["query"])
			} else {
				_, exists := attrs["query"]
				assert.False(t, exists)
			}

			durationMs, ok := attrs["duration_ms"].(int64)
			require.True(t, ok)
			assert.GreaterOrEqual(t, durationMs, int64(0))
		})
	}
}

func attrsToMap(attrs []any) map[string]any {
	result := make(map[string]any, len(attrs)/2)
	for i := 0; i+1 < len(attrs); i += 2 {
		key, ok := attrs[i].(string)
		if !ok {
			continue
		}

		result[key] = attrs[i+1]
	}

	return result
}
