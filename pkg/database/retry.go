package database

import (
	"context"
	"math"
	"math/rand/v2"
	"time"
)

// newRetrier creates a new retrier instance.
func newRetrier(cfg retryConfig) *retrier {
	return &retrier{
		cfg: cfg,
	}
}

// Do performs a retrier operation with exponential backoff.
func (r *retrier) Do(ctx context.Context, operation func(context.Context) error) error {
	for attempt := 0; attempt <= r.cfg.MaxRetries; attempt++ {
		if err := r.checkContext(ctx, attempt); err != nil {
			return err
		}

		err := operation(ctx)
		if err == nil {
			return nil
		}

		if isContextError(err) {
			return wrapDbError(operationRetry, err, map[string]any{
				"attempt": attempt,
			})
		}

		if !isRetryableError(err) || attempt == r.cfg.MaxRetries {
			return wrapDbError(operationRetry, err, map[string]any{
				"attempt":      attempt,
				"max_attempts": r.cfg.MaxRetries + 1,
			})
		}

		if err := r.sleepWithBackoff(ctx, attempt); err != nil {
			return err
		}
	}

	return wrapDbError(operationRetry, ctx.Err(), map[string]any{
		"max_attempts": r.cfg.MaxRetries + 1,
	})
}

// checkContext checks if the context is canceled and returns an error if so.
func (r *retrier) checkContext(ctx context.Context, attempt int) error {
	if err := ctx.Err(); err != nil {
		return wrapDbError(operationRetry, err, map[string]any{
			"attempt": attempt,
		})
	}

	if deadline, ok := ctx.Deadline(); ok {
		if time.Until(deadline) < r.cfg.InitialRetryQuota {
			return wrapDbError(operationRetry, context.DeadlineExceeded, map[string]any{
				"attempt":  attempt,
				"deadline": deadline.Format(time.RFC3339),
			})
		}
	}

	return nil
}

// sleepWithBackoff sleeps for the calculated backoff duration with context awareness.
func (r *retrier) sleepWithBackoff(ctx context.Context, attempt int) error {
	sleepDuration := r.calculateBackoff(attempt)

	timer := time.NewTimer(sleepDuration)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return wrapDbError(operationRetry, ctx.Err(), map[string]any{
			"attempt": attempt + 1,
		})
	case <-timer.C:
		return nil
	}
}

// calculateBackoff calculates the backoff duration for a given attempt using equal jitter strategy.
func (r *retrier) calculateBackoff(attempt int) time.Duration {
	exp := float64(r.cfg.InitialBackoff) * math.Pow(r.cfg.BackoffMultiplier, float64(attempt))
	cappedDuration := time.Duration(min(exp, float64(r.cfg.MaxBackoffTime)))

	halfExp := cappedDuration / jitterDivisor
	randomPart := time.Duration(rand.Float64() * float64(halfExp)) //nolint:gosec

	return halfExp + randomPart
}
