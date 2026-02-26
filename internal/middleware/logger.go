package middleware

import (
	"net/http"
	"time"

	chimw "github.com/go-chi/chi/v5/middleware"

	"go-auth/pkg/logger"
)

func Logger(log logger.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			ww := chimw.NewWrapResponseWriter(w, r.ProtoMajor)
			reqID := chimw.GetReqID(r.Context())

			defer func() {
				dur := time.Since(start)
				status := ww.Status()

				attrs := []any{
					"method", r.Method,
					"uri", r.URL.RequestURI(),
					"request_id", reqID,
					"status", status,
					"duration_ms", dur.Milliseconds(),
					"ip", clientIP(r),
					"user_agent", r.UserAgent(),
				}
				if query := r.URL.Query(); len(query) > 0 {
					attrs = append(attrs, "query", query)
				}

				switch {
				case status >= http.StatusInternalServerError:
					log.Error("request", attrs...)
				case status >= http.StatusBadRequest:
					log.Warn("request", attrs...)
				default:
					log.Info("request", attrs...)
				}
			}()

			next.ServeHTTP(ww, r)
		})
	}
}
