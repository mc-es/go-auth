package database

import (
	"context"
	"errors"
	"fmt"
	"maps"
	"net"
	"slices"
	"syscall"

	"go.mongodb.org/mongo-driver/mongo"
)

// dbError represents a database error.
type dbError struct {
	operation string         // Name of the database operation that failed
	err       error          // Underlying error that occurred
	metadata  map[string]any // Additional context about the error
}

// Operation names.
const (
	// Database.
	operationConnect    = "connect"
	operationDisconnect = "disconnect"
	operationPing       = "ping"
	operationGetVersion = "get_version"

	// Monitor.
	operationMonitor = "monitor"

	// Retry.
	operationRetry = "retry"
)

// retryableErrorLabels contains MongoDB error labels.
var retryableErrorLabels = []string{
	"TransientTransactionError",
	"UnknownTransactionCommitResult",
	"RetryableWriteError",
	"NetworkError",
}

// retryableSyscallErrors contains system call errors.
var retryableSyscallErrors = []error{
	syscall.ECONNREFUSED,
	syscall.ECONNRESET,
	syscall.EPIPE,
	syscall.ENETUNREACH,
	syscall.EHOSTUNREACH,
}

// newDbError creates a new database error.
func newDbError(operation string, err error, metadata ...map[string]any) *dbError {
	dbErr := &dbError{
		operation: operation,
		err:       err,
	}

	if len(metadata) > 0 && metadata[0] != nil {
		dbErr.metadata = maps.Clone(metadata[0])
	}

	return dbErr
}

// Error returns a formatted error message.
func (e *dbError) Error() string {
	if len(e.metadata) > 0 {
		return fmt.Sprintf("database %s failed: %v (metadata: %v)", e.operation, e.err, e.metadata)
	}

	return fmt.Sprintf("database %s failed: %v", e.operation, e.err)
}

// Unwrap returns the underlying error.
func (e *dbError) Unwrap() error {
	return e.err
}

// wrapDbError wraps a database error with metadata.
func wrapDbError(operation string, err error, metadata map[string]any) error {
	if err == nil {
		return nil
	}

	var dbErr *dbError
	if errors.As(err, &dbErr) {
		merged := make(map[string]any)
		if dbErr.metadata != nil {
			maps.Copy(merged, dbErr.metadata)
		}

		if metadata != nil {
			maps.Copy(merged, metadata)
		}

		return &dbError{
			operation: operation,
			err:       err,
			metadata:  merged,
		}
	}

	return newDbError(operation, err, metadata)
}

// isContextError checks if the error is due to context cancellation or deadline.
func isContextError(err error) bool {
	return errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded)
}

// isRetryableError determines if an error should trigger a retry.
func isRetryableError(err error) bool {
	if err == nil {
		return false
	}

	if isContextError(err) {
		return false
	}

	return isNetworkError(err) || isMongoError(err)
}

// isNetworkError checks if the error is a network-related error.
func isNetworkError(err error) bool {
	if errors.Is(err, mongo.ErrClientDisconnected) {
		return true
	}

	var netErr net.Error
	if errors.As(err, &netErr) && netErr.Timeout() {
		return true
	}

	return slices.ContainsFunc(retryableSyscallErrors, func(e error) bool {
		return errors.Is(err, e)
	})
}

// isMongoError checks if the error is a MongoDB-specific error.
func isMongoError(err error) bool {
	var serverErr mongo.ServerError
	if errors.As(err, &serverErr) {
		return slices.ContainsFunc(retryableErrorLabels, func(label string) bool {
			return serverErr.HasErrorLabel(label)
		})
	}

	return false
}
