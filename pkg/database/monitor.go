package database

import (
	"context"
	"errors"
	"time"

	"go-auth/pkg/logger"
)

// NewMonitor creates a new Monitor instance.
func NewMonitor(prober Prober, interval, timeout time.Duration) (Monitor, error) {
	if prober == nil {
		return nil, wrapDbError(
			operationMonitor,
			errors.New("prober cannot be nil"),
			map[string]any{
				"interval": interval.String(),
				"timeout":  timeout.String(),
			},
		)
	}

	monitor := &monitor{
		prober:   prober,
		interval: interval,
		timeout:  timeout,
	}
	monitor.healthy.Store(true)

	return monitor, nil
}

// Start starts the monitor goroutine.
func (m *monitor) Start(parentCtx context.Context) {
	if !m.running.CompareAndSwap(false, true) {
		logger.S().Warnw("Monitor already running")

		return
	}

	m.mu.Lock()

	ctx, cancel := context.WithCancel(parentCtx)
	m.ctxCancel = cancel
	m.mu.Unlock()

	logger.S().Debugw("Starting periodic monitor",
		"interval", m.interval.String(),
		"timeout", m.timeout.String(),
	)

	m.waitGroup.Go(func() {
		ticker := time.NewTicker(m.interval)
		defer ticker.Stop()

		if err := m.checkMonitor(ctx, m.timeout); err != nil {
			logger.S().Warnw("Initial monitor failed", "error", err)
		}

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				_ = m.checkMonitor(ctx, m.timeout)
			}
		}
	})
}

// IsHealthy returns the current monitor status.
func (m *monitor) IsHealthy() bool {
	return m.healthy.Load()
}

// Stop stops the monitor goroutine.
func (m *monitor) Stop() {
	if !m.running.CompareAndSwap(true, false) {
		logger.S().Warnw("Monitor not running")

		return
	}

	m.mu.Lock()
	cancel := m.ctxCancel
	m.ctxCancel = nil
	m.mu.Unlock()

	if cancel != nil {
		cancel()
	}

	done := make(chan struct{})

	go func() {
		m.waitGroup.Wait()
		close(done)
	}()

	select {
	case <-done:
		logger.S().Debugw("Monitor stopped")
	case <-time.After(stopTimeout):
		logger.S().Warnw("Timeout waiting for monitor to stop")
	}
}

// checkMonitor checks the monitor of the database.
func (m *monitor) checkMonitor(ctx context.Context, timeout time.Duration) error {
	if !m.inProgress.CompareAndSwap(false, true) {
		logger.S().Debugw("Monitor skipped, previous monitor in progress")

		return nil
	}
	defer m.inProgress.Store(false)

	checkCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	start := time.Now()
	err := m.prober.Ping(checkCtx)
	duration := time.Since(start)
	healthy := err == nil
	m.healthy.Store(healthy)

	if err != nil {
		return wrapDbError(operationMonitor, err, map[string]any{
			"duration": duration.String(),
			"timeout":  timeout.String(),
		})
	}

	logger.S().Debugw("Monitor succeeded", "duration", duration.String())

	return nil
}
