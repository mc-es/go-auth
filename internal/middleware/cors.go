package middleware

import (
	"net/http"
	"strconv"
	"strings"

	"go-auth/internal/config"
)

const accessControlAllowOrigin = "Access-Control-Allow-Origin"

func CORS(cfg *config.CORS) func(next http.Handler) http.Handler {
	if cfg == nil {
		return func(next http.Handler) http.Handler { return next }
	}

	allowedOrigins, allowAll := buildAllowedOrigins(cfg.Origins)

	methods := strings.Join(cfg.Methods, ", ")
	headers := strings.Join(cfg.Headers, ", ")
	maxAge := strconv.Itoa(cfg.MaxAge)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			setOriginHeaders(w, r.Header.Get("Origin"), allowAll, cfg.AllowCreds, allowedOrigins)
			setCORSHeaders(w, methods, headers, maxAge, cfg.AllowCreds)

			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)

				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func buildAllowedOrigins(origins []string) (map[string]struct{}, bool) {
	allowedOrigins := make(map[string]struct{}, len(origins))
	allowAll := false

	for _, origin := range origins {
		if origin == "*" {
			allowAll = true

			break
		}

		allowedOrigins[origin] = struct{}{}
	}

	return allowedOrigins, allowAll
}

func setOriginHeaders(w http.ResponseWriter, origin string, allowAll, allowCreds bool, allowed map[string]struct{}) {
	if origin == "" {
		return
	}

	w.Header().Add("Vary", "Origin")

	if allowAll {
		if allowCreds {
			w.Header().Set(accessControlAllowOrigin, origin)
		} else {
			w.Header().Set(accessControlAllowOrigin, "*")
		}

		return
	}

	if _, ok := allowed[origin]; ok {
		w.Header().Set(accessControlAllowOrigin, origin)
	}
}

func setCORSHeaders(w http.ResponseWriter, methods, headers, maxAge string, allowCreds bool) {
	if allowCreds {
		w.Header().Set("Access-Control-Allow-Credentials", "true")
	}

	w.Header().Set("Access-Control-Allow-Methods", methods)
	w.Header().Set("Access-Control-Allow-Headers", headers)
	w.Header().Set("Access-Control-Max-Age", maxAge)
}
