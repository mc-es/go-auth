package middleware

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"

	"go-auth/pkg/logger"
)

func Logger(log logger.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
			reqID := middleware.GetReqID(r.Context())

			next.ServeHTTP(ww, r)

			dur := time.Since(start)
			log.Info("request",
				"method", r.Method,
				"path", r.URL.Path,
				"request_id", reqID,
				"status", ww.Status(),
				"duration_ms", dur.Milliseconds(),
			)
		})
	}
}
