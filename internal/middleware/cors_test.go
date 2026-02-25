package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"go-auth/internal/config"
	"go-auth/internal/middleware"
)

func TestCORS(t *testing.T) {
	const (
		allowedOrigin = "http://allowed.com"
		localOrigin   = "http://localhost:3000"
		evilOrigin    = "http://evil.com"
		headerContent = "Content-Type"
	)

	tests := []struct {
		name             string
		cfg              *config.CORS
		origin           string
		method           string
		wantOrigin       string
		wantCreds        string
		wantStatus       int
		expectNextCalled bool
	}{
		{
			name:             "nil config",
			cfg:              nil,
			origin:           localOrigin,
			method:           "GET",
			wantOrigin:       "",
			wantCreds:        "",
			wantStatus:       http.StatusOK,
			expectNextCalled: true,
		},
		{
			name: "Wildcard origin no creds",
			cfg: &config.CORS{
				Origins:    []string{"*"},
				Methods:    []string{"GET", "POST"},
				Headers:    []string{headerContent},
				MaxAge:     600,
				AllowCreds: false,
			},
			origin:           localOrigin,
			method:           "GET",
			wantOrigin:       "*",
			wantCreds:        "",
			wantStatus:       http.StatusOK,
			expectNextCalled: true,
		},
		{
			name: "Wildcard origin with creds",
			cfg: &config.CORS{
				Origins:    []string{"*"},
				Methods:    []string{"GET"},
				Headers:    []string{headerContent},
				MaxAge:     600,
				AllowCreds: true,
			},
			origin:           localOrigin,
			method:           "GET",
			wantOrigin:       localOrigin,
			wantCreds:        "true",
			wantStatus:       http.StatusOK,
			expectNextCalled: true,
		},
		{
			name: "Specific origin allowed",
			cfg: &config.CORS{
				Origins:    []string{allowedOrigin},
				Methods:    []string{"GET"},
				Headers:    []string{headerContent},
				MaxAge:     600,
				AllowCreds: true,
			},
			origin:           allowedOrigin,
			method:           "GET",
			wantOrigin:       allowedOrigin,
			wantCreds:        "true",
			wantStatus:       http.StatusOK,
			expectNextCalled: true,
		},
		{
			name: "Specific origin NOT allowed",
			cfg: &config.CORS{
				Origins:    []string{allowedOrigin},
				Methods:    []string{"GET"},
				Headers:    []string{headerContent},
				MaxAge:     600,
				AllowCreds: false,
			},
			origin:           evilOrigin,
			method:           "GET",
			wantOrigin:       "",
			wantCreds:        "",
			wantStatus:       http.StatusOK,
			expectNextCalled: true,
		},
		{
			name: "OPTIONS request preflight",
			cfg: &config.CORS{
				Origins:    []string{"*"},
				Methods:    []string{"GET"},
				Headers:    []string{headerContent},
				MaxAge:     600,
				AllowCreds: false,
			},
			origin:           localOrigin,
			method:           "OPTIONS",
			wantOrigin:       "*",
			wantCreds:        "",
			wantStatus:       http.StatusNoContent,
			expectNextCalled: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			called := false
			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				called = true

				w.WriteHeader(http.StatusOK)
			})

			handler := middleware.CORS(tt.cfg)(next)

			req := httptest.NewRequest(tt.method, "/", http.NoBody)
			if tt.origin != "" {
				req.Header.Set("Origin", tt.origin)
			}

			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)
			assert.Equal(t, tt.wantOrigin, rr.Header().Get("Access-Control-Allow-Origin"))
			assert.Equal(t, tt.wantCreds, rr.Header().Get("Access-Control-Allow-Credentials"))
			assert.Equal(t, tt.wantStatus, rr.Code)
			assert.Equal(t, tt.expectNextCalled, called)
		})
	}
}
