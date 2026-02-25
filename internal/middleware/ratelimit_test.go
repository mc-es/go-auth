package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"go-auth/internal/config"
	"go-auth/internal/middleware"
)

func TestRateLimit(t *testing.T) {
	const (
		ip1 = "192.168.1.1"
		ip2 = "10.0.0.1"
		ip3 = "1.1.1.1"
		ip4 = "2.2.2.2"
	)

	tests := []struct {
		name           string
		cfg            config.RateLimit
		requests       []string
		expectedStatus []int
		delay          time.Duration
	}{
		{
			name: "Under Limit",
			cfg: config.RateLimit{
				Limit:  3,
				Period: 100 * time.Millisecond,
			},
			requests:       []string{ip1, ip1},
			expectedStatus: []int{http.StatusOK, http.StatusOK},
		},
		{
			name: "Hit Limit",
			cfg: config.RateLimit{
				Limit:  3,
				Period: 100 * time.Millisecond,
			},
			requests:       []string{ip2, ip2, ip2, ip2},
			expectedStatus: []int{http.StatusOK, http.StatusOK, http.StatusOK, http.StatusTooManyRequests},
		},
		{
			name: "IP Isolation",
			cfg: config.RateLimit{
				Limit:  1,
				Period: 100 * time.Millisecond,
			},
			requests:       []string{ip3, ip3, ip4},
			expectedStatus: []int{http.StatusOK, http.StatusTooManyRequests, http.StatusOK},
		},
		{
			name: "Reset After Period",
			cfg: config.RateLimit{
				Limit:  1,
				Period: 50 * time.Millisecond,
			},
			requests:       []string{ip1, ip1},
			expectedStatus: []int{http.StatusOK, http.StatusOK},
			delay:          60 * time.Millisecond,
		},
		{
			name: "Garbage Collection Eviction",
			cfg: config.RateLimit{
				Limit:  1,
				Period: 15 * time.Millisecond,
			},
			requests:       []string{ip1, ip1},
			expectedStatus: []int{http.StatusOK, http.StatusOK},
			delay:          60 * time.Millisecond,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctx := t.Context()

			assert.Len(t, tt.expectedStatus, len(tt.requests))

			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			})

			rlMiddleware := middleware.RateLimit(ctx, tt.cfg)
			handler := rlMiddleware(next)

			for idx, ip := range tt.requests {
				if idx > 0 && tt.delay > 0 {
					time.Sleep(tt.delay)
				}

				req := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
				req.Header.Set("X-Forwarded-For", ip)

				rr := httptest.NewRecorder()
				handler.ServeHTTP(rr, req)
				assert.Equal(t, tt.expectedStatus[idx], rr.Code)

				if rr.Code == http.StatusTooManyRequests {
					retryAfter := rr.Header().Get("Retry-After")
					assert.NotEmpty(t, retryAfter)
				}
			}
		})
	}
}
