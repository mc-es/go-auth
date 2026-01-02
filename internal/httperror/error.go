// Package httperror provides a structured error handling system for HTTP APIs.
//
// Usage:
//
//	// Simple error
//	return httperror.BadRequest("Invalid email format")
//
//	// Error with cause
//	return httperror.InternalServerError("Database connection failed", dbErr)
//
//	// Error with code
//	return httperror.BadRequest("Invalid email format").WithCode(httperror.ErrCodeInvalidJSON)
package httperror

import (
	"fmt"
	"net/http"
)

// Error represents a structured HTTP error with status code, message, and optional error code.
type Error struct {
	Status  int    `json:"-"`
	Code    string `json:"code,omitempty"`
	Message string `json:"message"`
	Cause   error  `json:"-"`
}

// Error implements the error interface.
func (e *Error) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Cause)
	}

	return e.Message
}

// Unwrap returns the underlying error.
func (e *Error) Unwrap() error {
	return e.Cause
}

// WithCode sets the error code and returns the error for method chaining.
func (e *Error) WithCode(code ErrorCode) *Error {
	if code != "" {
		e.Code = string(code)
	}

	return e
}

// New creates a new httperror.Error with the specified HTTP status code and message.
//
// Example:
//
//	err := httperror.New(http.StatusBadRequest, "Invalid input")
//	err := httperror.New(http.StatusNotFound, "User not found", dbErr)
func New(status int, message string, err ...error) *Error {
	if status < 100 || status > 599 {
		status = http.StatusInternalServerError
	}

	if message == "" {
		message = http.StatusText(status)
	}

	var cause error
	if len(err) > 0 {
		cause = err[0]
	}

	return &Error{
		Status:  status,
		Message: message,
		Cause:   cause,
	}
}

// BadRequest creates a 400 Bad Request error.
func BadRequest(message string, err ...error) *Error {
	return New(http.StatusBadRequest, message, err...)
}

// Unauthorized creates a 401 Unauthorized error.
func Unauthorized(message string, err ...error) *Error {
	return New(http.StatusUnauthorized, message, err...)
}

// Forbidden creates a 403 Forbidden error.
func Forbidden(message string, err ...error) *Error {
	return New(http.StatusForbidden, message, err...)
}

// NotFound creates a 404 Not Found error.
func NotFound(message string, err ...error) *Error {
	return New(http.StatusNotFound, message, err...)
}

// MethodNotAllowed creates a 405 Method Not Allowed error.
func MethodNotAllowed(message string, err ...error) *Error {
	return New(http.StatusMethodNotAllowed, message, err...)
}

// RequestTimeout creates a 408 Request Timeout error.
func RequestTimeout(message string, err ...error) *Error {
	return New(http.StatusRequestTimeout, message, err...)
}

// Conflict creates a 409 Conflict error.
func Conflict(message string, err ...error) *Error {
	return New(http.StatusConflict, message, err...)
}

// UnprocessableEntity creates a 422 Unprocessable Entity error.
func UnprocessableEntity(message string, err ...error) *Error {
	return New(http.StatusUnprocessableEntity, message, err...)
}

// TooManyRequests creates a 429 Too Many Requests error.
func TooManyRequests(message string, err ...error) *Error {
	return New(http.StatusTooManyRequests, message, err...)
}

// InternalServerError creates a 500 Internal Server Error.
func InternalServerError(message string, err ...error) *Error {
	return New(http.StatusInternalServerError, message, err...)
}

// NotImplemented creates a 501 Not Implemented error.
func NotImplemented(message string, err ...error) *Error {
	return New(http.StatusNotImplemented, message, err...)
}

// BadGateway creates a 502 Bad Gateway error.
func BadGateway(message string, err ...error) *Error {
	return New(http.StatusBadGateway, message, err...)
}

// ServiceUnavailable creates a 503 Service Unavailable error.
func ServiceUnavailable(message string, err ...error) *Error {
	return New(http.StatusServiceUnavailable, message, err...)
}

// GatewayTimeout creates a 504 Gateway Timeout error.
func GatewayTimeout(message string, err ...error) *Error {
	return New(http.StatusGatewayTimeout, message, err...)
}
