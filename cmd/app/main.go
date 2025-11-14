// Package main is the app's entry point and starts auth service.
package main

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"

	"github.com/mc-es/go-auth/pkg/logger"
)

const defaultPort = 8080
const (
	readTimeout  = 5 * time.Second
	writeTimeout = 10 * time.Second
	idleTimeout  = 120 * time.Second
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

	defer func() {
		_ = logger.Sync()
	}()

	port := getPort()
	startServer(port)
}

func getPort() int {
	err := godotenv.Load()
	if err != nil {
		logger.S().Warnw("No .env file found")
	}

	port, err := strconv.Atoi(os.Getenv("PORT"))
	if err != nil || port <= 0 {
		logger.S().Warnw("Invalid PORT value, using default", "PORT", os.Getenv("PORT"))

		return defaultPort
	}

	return port
}

func startServer(port int) {
	mux := http.NewServeMux()

	mux.HandleFunc("/", func(res http.ResponseWriter, req *http.Request) {
		logger.S().Infow("Incoming request",
			"method", req.Method,
			"path", req.URL.Path,
			"remote_addr", req.RemoteAddr,
		)

		_, _ = fmt.Fprintf(res, "Hello from go-auth 👋")
	})

	addr := fmt.Sprintf(":%d", port)
	server := &http.Server{
		Addr:         addr,
		Handler:      mux,
		ReadTimeout:  readTimeout,
		WriteTimeout: writeTimeout,
		IdleTimeout:  idleTimeout,
	}

	logger.S().Infow("Server starting...", "port", port)

	err := server.ListenAndServe()
	if err != nil {
		logger.S().Errorw("Server failed...", "error", err)
	}
}
