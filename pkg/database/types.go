package database

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"go.mongodb.org/mongo-driver/mongo"

	"go-auth/config"
)

// Database defines the interface for database connection management.
type Database interface {
	Ping(ctx context.Context) error
	Close() error
}

// Monitor defines the interface for database monitoring.
type Monitor interface {
	Start(ctx context.Context)
	IsHealthy() bool
	Stop()
}

// database is the implementation of Database.
type database struct {
	client  *mongo.Client   // MongoDB client instance
	config  config.Database // Database configuration
	retrier *retrier        // Retrier used for retries
}

// monitor is the implementation of Monitor.
type monitor struct {
	healthy    atomic.Bool                 // Current health status (true = healthy, false = unhealthy)
	running    atomic.Bool                 // Whether the monitor goroutine is running
	inProgress atomic.Bool                 // Whether a monitor is currently in progress
	waitGroup  sync.WaitGroup              // Waits for the monitor goroutine to finish
	ping       func(context.Context) error // Function to ping the database
	interval   time.Duration               // Time between monitors
	timeout    time.Duration               // Timeout for each monitor
	mu         sync.Mutex                  // Mutex for protecting the monitor state
	ctxCancel  context.CancelFunc          // Cancels the monitor context
}

// retrier handles retry logic with exponential backoff for database operations.
type retrier struct {
	cfg retryConfig
}

// retryConfig is the configuration for the retry operation.
type retryConfig struct {
	MaxRetries        int           // Maximum number of retries
	InitialBackoff    time.Duration // Initial backoff time
	MaxBackoffTime    time.Duration // Maximum backoff time
	BackoffMultiplier float64       // Backoff multiplier
	InitialRetryQuota time.Duration // Initial retry quota
}

// Default values for retrier configuration.
const (
	defaultMaxBackoff        = 5 * time.Second
	defaultBackoffMultiplier = 2.0
	defaultInitialRetryQuota = 50 * time.Millisecond

	jitterDivisor = 2               // Divisor for jitter backoff calculation
	stopTimeout   = 5 * time.Second // Timeout for stopping health check goroutine
)

// Compile-time assertions.
var (
	_ Database = (*database)(nil)
	_ Monitor  = (*monitor)(nil)
)
