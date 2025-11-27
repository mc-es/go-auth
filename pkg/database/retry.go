package database

import (
	"context"
	"math"
	"math/rand/v2"
	"time"
)

// RetryConfig configures retry behavior.
type RetryConfig struct {
	MaxRetries        int           // The maximum number of retries allowed
	InitialBackoff    time.Duration // The initial wait time before the first retry
	MaxBackoffTime    time.Duration // The maximum wait time between retries
	BackoffMultiplier float64       // The factor by which the wait time increases
	InitialRetryQuota time.Duration // The minimum remaining context deadline required to retry
}

// Default values for retry configuration.
const (
	defaultMaxRetries        = 3
	defaultInitialBackoff    = 100 * time.Millisecond
	defaultMaxBackoff        = 5 * time.Second
	defaultBackoffMultiplier = 2.0
	defaultInitialRetryQuota = 50 * time.Millisecond
	equalJitterDivisor       = 2.0
)

// Retry executes an operation with retry logic and exponential backoff.
func Retry(
	ctx context.Context,
	operation func(context.Context) error,
	config ...RetryConfig,
) error {
	cfg := defaultRetryConfig()
	if len(config) > 0 {
		cfg = config[0]
	}

	for attempt := 0; attempt <= cfg.MaxRetries; attempt++ {
		if err := checkContext(ctx, cfg, attempt); err != nil {
			return err
		}

		err := operation(ctx)
		if err == nil {
			return nil
		}

		if !IsRetryableError(err) || attempt == cfg.MaxRetries {
			return WrapErrorWithMetadata(operationRetry, errOperationFailed, map[string]any{
				"attempt":      attempt,
				"retryable":    false,
				"max_attempts": cfg.MaxRetries + 1,
				"error":        err.Error(),
			})
		}

		if err := sleepWithBackoff(ctx, cfg, attempt); err != nil {
			return err
		}
	}

	return WrapErrorWithMetadata(operationRetry, errRetryExhausted, map[string]any{
		"max_attempts": cfg.MaxRetries + 1,
	})
}

// defaultRetryConfig returns the default retry configuration.
func defaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxRetries:        defaultMaxRetries,
		InitialBackoff:    defaultInitialBackoff,
		MaxBackoffTime:    defaultMaxBackoff,
		BackoffMultiplier: defaultBackoffMultiplier,
		InitialRetryQuota: defaultInitialRetryQuota,
	}
}

// checkContext checks if the context is cancelled and returns an error if so.
func checkContext(ctx context.Context, cfg RetryConfig, attempt int) error {
	if err := ctx.Err(); err != nil {
		return WrapErrorWithMetadata(operationRetry, errContextCancelled, map[string]any{
			"attempt": attempt,
			"error":   err.Error(),
		})
	}

	if deadline, ok := ctx.Deadline(); ok {
		if time.Until(deadline) < cfg.InitialRetryQuota {
			return WrapErrorWithMetadata(operationRetry, errDeadlineExceeded, map[string]any{
				"attempt": attempt,
			})
		}
	}

	return nil
}

// sleepWithBackoff sleeps for the calculated backoff duration with context awareness.
func sleepWithBackoff(ctx context.Context, cfg RetryConfig, attempt int) error {
	sleepDuration := calculateBackoff(cfg, attempt)

	timer := time.NewTimer(sleepDuration)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return WrapErrorWithMetadata(operationRetry, errContextCancelled, map[string]any{
			"attempt": attempt + 1,
			"error":   ctx.Err().Error(),
		})
	case <-timer.C:
		return nil
	}
}

// calculateBackoff calculates the backoff duration for a given attempt using equal jitter strategy.
func calculateBackoff(cfg RetryConfig, attempt int) time.Duration {
	exp := float64(cfg.InitialBackoff) * math.Pow(cfg.BackoffMultiplier, float64(attempt))
	cappedDuration := time.Duration(min(exp, float64(cfg.MaxBackoffTime)))

	halfExp := cappedDuration / time.Duration(equalJitterDivisor)
	randomPart := time.Duration(rand.Float64() * float64(halfExp)) //nolint:gosec

	return halfExp + randomPart
}
