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

// Error represents a database error.
type Error struct {
	operation string         // The operation that caused the error
	err       error          // The underlying error
	metadata  map[string]any // Additional metadata about the error
}

// Operation constants for database operations.
const (
	operationConnect       = "connect"
	operationDisconnect    = "disconnect"
	operationPing          = "ping"
	operationRetry         = "retry"
	operationHealthCheck   = "health_check"
	operationPeriodicCheck = "periodic_check"
)

// Database error constants.
var (
	errConnectionFailed = errors.New("database connection failed")
	errDisconnectFailed = errors.New("database disconnect failed")
	errPingFailed       = errors.New("database ping failed")
	errRetryExhausted   = errors.New("retry attempts exhausted")
	errContextCancelled = errors.New("operation cancelled by context")
	errCheckInProgress  = errors.New("check in progress")
	errOperationFailed  = errors.New("database operation failed")
	errDeadlineExceeded = errors.New("operation deadline exceeded")
)

// retryableErrorLabels is a list of labels that indicate a retryable error.
var retryableErrorLabels = []string{
	"TransientTransactionError",
	"UnknownTransactionCommitResult",
	"RetryableWriteError",
	"NetworkError",
}

// retryableSyscallErrors is a list of syscall errors that indicate a retryable error.
var retryableSyscallErrors = []error{
	syscall.ECONNREFUSED,
	syscall.ECONNRESET,
	syscall.EPIPE,
	syscall.ENETUNREACH,
	syscall.EHOSTUNREACH,
}

// NewError creates a new database error.
func NewError(operation string, err error, metadata ...map[string]any) *Error {
	dbErr := &Error{
		operation: operation,
		err:       err,
	}

	if len(metadata) > 0 && metadata[0] != nil {
		dbErr.metadata = maps.Clone(metadata[0])
	}

	return dbErr
}

// Error returns the error message.
func (e *Error) Error() string {
	if len(e.metadata) > 0 {
		return fmt.Sprintf("database %s failed: %v (metadata: %v)", e.operation, e.err, e.metadata)
	}

	return fmt.Sprintf("database %s failed: %v", e.operation, e.err)
}

// Unwrap returns the underlying error.
func (e *Error) Unwrap() error {
	return e.err
}

// WrapErrorWithMetadata wraps an error with metadata.
func WrapErrorWithMetadata(operation string, err error, metadata map[string]any) error {
	if err == nil {
		return nil
	}

	var dbErr *Error
	if errors.As(err, &dbErr) {
		merged := make(map[string]any)
		if dbErr.metadata != nil {
			maps.Copy(merged, dbErr.metadata)
		}

		if metadata != nil {
			maps.Copy(merged, metadata)
		}

		return &Error{
			operation: operation,
			err:       err,
			metadata:  merged,
		}
	}

	return NewError(operation, err, metadata)
}

// IsRetryableError checks if the error is retryable.
func IsRetryableError(err error) bool {
	if err == nil {
		return false
	}

	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
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
