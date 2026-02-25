package middleware

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"golang.org/x/time/rate"

	"go-auth/internal/config"
)

type client struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

func RateLimit(ctx context.Context, cfg config.RateLimit) func(next http.Handler) http.Handler {
	tokensPerSec := rate.Limit(float64(cfg.Limit) / cfg.Period.Seconds())
	burst := max(cfg.Limit, 1)

	var (
		mu      sync.Mutex
		clients = make(map[string]*client)
	)

	go func() {
		ticker := time.NewTicker(cfg.Period)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				mu.Lock()

				for ip, c := range clients {
					if time.Since(c.lastSeen) > cfg.Period*3 {
						delete(clients, ip)
					}
				}

				mu.Unlock()
			case <-ctx.Done():
				return
			}
		}
	}()

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := clientIP(r)

			mu.Lock()

			cl, exists := clients[ip]
			if !exists {
				cl = &client{
					limiter: rate.NewLimiter(tokensPerSec, burst),
				}
				clients[ip] = cl
			}

			cl.lastSeen = time.Now()

			mu.Unlock()

			if !cl.limiter.Allow() {
				retryAfter := fmt.Sprintf("%.0f", cfg.Period.Seconds())
				w.Header().Set("Retry-After", retryAfter)

				http.Error(w, http.StatusText(http.StatusTooManyRequests), http.StatusTooManyRequests)

				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func clientIP(r *http.Request) string {
	if x := r.Header.Get("X-Forwarded-For"); x != "" {
		return strings.TrimSpace(strings.Split(x, ",")[0])
	}

	if x := r.Header.Get("X-Real-IP"); x != "" {
		return strings.TrimSpace(strings.Split(x, ",")[0])
	}

	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}

	return ip
}
