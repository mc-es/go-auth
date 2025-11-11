package main

import (
	"fmt"
	"net/http"
	"os"
	"strconv"

	"github.com/joho/godotenv"
	"github.com/mc-es/go-auth/pkg/logger"
)

func main() {
	if err := logger.Init(
		logger.WithDevelopmentMode(),
		logger.WithLevel("debug"),
		logger.WithInitialFields(map[string]any{
			"app":     "go-auth",
			"version": "1.0.0",
		}),
	); err != nil {
		panic(err)
	}
	defer logger.Sync()

	port := getPort()
	startServer(port)
}

func getPort() int {
	err := godotenv.Load()
	if err != nil {
		fmt.Println("No .env file found")
	}

	val := os.Getenv("PORT")
	if val == "" {
		return 8080
	}
	p, err := strconv.Atoi(val)
	if err != nil || p <= 0 {
		logger.S().Warnw("Invalid PORT value, using default", "PORT", val)
		return 8080
	}
	return p
}

func startServer(port int) {
	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		logger.S().Infow("Incoming request",
			"method", r.Method,
			"path", r.URL.Path,
			"remote_addr", r.RemoteAddr,
		)
		fmt.Fprintf(w, "Hello from go-auth 👋")
	})

	addr := fmt.Sprintf(":%d", port)
	logger.S().Infow("Server starting...", "port", port)

	if err := http.ListenAndServe(addr, mux); err != nil {
		logger.S().Errorw("Server failed...", "error", err)
	}
}
