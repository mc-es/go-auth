package database

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"go-auth/pkg/logger"
)

// HealthCheckConfig configures periodic health check behavior.
type HealthCheckConfig struct {
	Interval time.Duration // The interval at which to perform the health check
	Timeout  time.Duration // The timeout for the health check
}

// HealthStatus represents the current health status of the database.
type HealthStatus struct {
	healthy      bool      // Whether the database is healthy
	lastCheck    time.Time // The time the last health check was performed
	lastError    error     // The error from the last health check
	checkCount   int64     // The number of health checks performed
	failureCount int64     // The number of health checks that have failed
}

// HealthMonitor manages periodic health checks and status tracking.
type HealthMonitor struct {
	status     atomic.Pointer[HealthStatus] // The current health status of the database
	waitGroup  sync.WaitGroup               // The wait group for the health check goroutine
	ctxCancel  context.CancelFunc           // The context cancel function for the health check goroutine
	running    atomic.Bool                  // Whether the health check goroutine is running
	inProgress atomic.Bool                  // Whether a health check is in progress
	pingFunc   func(context.Context) error  // The function to perform the health check
}

// Default values for health check configuration.
const (
	defaultHealthCheckInterval    = 10 * time.Minute
	defaultHealthCheckTimeout     = 10 * time.Second
	defaultHealthCheckStopTimeout = 5 * time.Second
)

// NewHealthMonitor creates a new HealthMonitor instance.
func NewHealthMonitor(pingFunc func(context.Context) error) *HealthMonitor {
	if pingFunc == nil {
		panic("pingFunc cannot be nil")
	}

	hm := &HealthMonitor{
		pingFunc: pingFunc,
	}
	hm.status.Store(&HealthStatus{healthy: false})

	return hm
}

// GetHealthStats returns health check statistics from the database.
func (hm *HealthMonitor) GetHealthStats() HealthStatus {
	val := hm.status.Load()
	if val == nil {
		return HealthStatus{}
	}

	return *val
}

// GetHealthStatus returns the current health status of the database.
func (hm *HealthMonitor) GetHealthStatus(ctx context.Context) (bool, error) {
	timeout := defaultHealthCheckTimeout

	if d, ok := ctx.Deadline(); ok {
		if remaining := time.Until(d); remaining > 100*time.Millisecond {
			timeout = remaining
		}
	}

	return hm.executeCheck(ctx, timeout, operationHealthCheck)
}

// StartHealthCheck starts the health check goroutine.
func (hm *HealthMonitor) StartHealthCheck(parentCtx context.Context, cfgs ...HealthCheckConfig) {
	if !hm.running.CompareAndSwap(false, true) {
		logger.S().Warnw("Health check already running; ignoring StartHealthCheck")

		return
	}

	cfg := defaultHealthCheckConfig()
	if len(cfgs) > 0 {
		cfg = cfgs[0]
	}

	ctx, cancel := context.WithCancel(parentCtx)
	hm.ctxCancel = cancel

	logger.S().Infow("Starting periodic health check",
		"interval", cfg.Interval,
		"timeout", cfg.Timeout,
	)

	hm.waitGroup.Go(func() {
		ticker := time.NewTicker(cfg.Interval)
		defer ticker.Stop()

		_, _ = hm.executeCheck(ctx, cfg.Timeout, operationPeriodicCheck)

		for {
			select {
			case <-ctx.Done():
				logger.S().Infow("Health check stopped", "reason", ctx.Err())

				return
			case <-ticker.C:
				_, _ = hm.executeCheck(ctx, cfg.Timeout, operationPeriodicCheck)
			}
		}
	})
}

// StopHealthCheck stops the health check goroutine and waits for it to complete.
func (hm *HealthMonitor) StopHealthCheck() {
	if !hm.running.CompareAndSwap(true, false) {
		return
	}

	if hm.ctxCancel != nil {
		hm.ctxCancel()
	}

	done := make(chan struct{})

	go func() {
		hm.waitGroup.Wait()
		close(done)
	}()

	select {
	case <-done:
		logger.S().Debugw("Health check stopped")
	case <-time.After(defaultHealthCheckStopTimeout):
		logger.S().Warnw("Timeout waiting for health check to stop")
	}
}

// defaultHealthCheckConfig returns a default health check configuration.
func defaultHealthCheckConfig() HealthCheckConfig {
	return HealthCheckConfig{
		Interval: defaultHealthCheckInterval,
		Timeout:  defaultHealthCheckTimeout,
	}
}

// executeCheck performs a single health check and updates the status.
func (hm *HealthMonitor) executeCheck(
	ctx context.Context,
	timeout time.Duration,
	opSource string,
) (bool, error) {
	if !hm.inProgress.CompareAndSwap(false, true) {
		logger.S().Debugw("Health check skipped (in progress)", "source", opSource)

		return false, WrapErrorWithMetadata(opSource, errCheckInProgress, map[string]any{
			"source": opSource,
		})
	}
	defer hm.inProgress.Store(false)

	checkCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	start := time.Now()
	err := hm.pingFunc(checkCtx)
	duration := time.Since(start)

	hm.updateStatus(err == nil, err)

	if err != nil {
		logger.S().
			Warnw("Health check failed", "source", opSource, "error", err, "duration", duration)

		return hm.GetHealthStats().healthy, WrapErrorWithMetadata(
			opSource,
			errPingFailed,
			map[string]any{
				"error":    err.Error(),
				"duration": duration,
				"source":   opSource,
			},
		)
	}

	logger.S().Debugw("Health check succeeded", "source", opSource, "duration", duration)

	return hm.GetHealthStats().healthy, nil
}

// updateStatus updates the health status.
func (hm *HealthMonitor) updateStatus(healthy bool, err error) {
	prev := hm.status.Load()
	next := *prev

	next.checkCount++
	next.lastCheck = time.Now()
	next.healthy = healthy

	if !healthy {
		next.failureCount++
		next.lastError = err
	} else {
		next.lastError = nil
	}

	if prev.healthy != healthy {
		if healthy {
			logger.S().Infow("Database status changed: HEALTHY")
		} else {
			logger.S().Warnw("Database status changed: UNHEALTHY", "error", err)
		}
	}

	hm.status.Store(&next)
}
