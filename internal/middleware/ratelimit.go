package middleware

import (
	"net/http"
	"strings"
	"sync"

	"golang.org/x/time/rate"

	"go-auth/internal/config"
)

func RateLimit(cfg config.RateLimit) func(next http.Handler) http.Handler {
	if cfg.Limit <= 0 {
		return func(next http.Handler) http.Handler { return next }
	}

	tokensPerSec := float64(cfg.Limit) / cfg.Period.Seconds()
	burst := max(cfg.Limit, 1)

	var (
		limiters sync.Map
		mu       sync.Mutex
	)

	getLimiter := func(key string) *rate.Limiter {
		if v, ok := limiters.Load(key); ok {
			return v.(*rate.Limiter)
		}

		mu.Lock()
		defer mu.Unlock()

		if v, ok := limiters.Load(key); ok {
			return v.(*rate.Limiter)
		}

		lim := rate.NewLimiter(rate.Limit(tokensPerSec), burst)
		limiters.Store(key, lim)

		return lim
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			key := clientIP(r)

			lim := getLimiter(key)
			if !lim.Allow() {
				w.Header().Set("Retry-After", "60")
				http.Error(w, http.StatusText(http.StatusTooManyRequests), http.StatusTooManyRequests)

				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func clientIP(r *http.Request) string {
	if x := r.Header.Get("X-Real-IP"); x != "" {
		return strings.TrimSpace(strings.Split(x, ",")[0])
	}

	if x := r.Header.Get("X-Forwarded-For"); x != "" {
		return strings.TrimSpace(strings.Split(x, ",")[0])
	}

	addr := r.RemoteAddr
	if i := strings.LastIndex(addr, ":"); i >= 0 {
		addr = addr[:i]
	}

	return addr
}
