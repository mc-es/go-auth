package apperror

import (
	"errors"
	"fmt"
	"net/http"
)

type Error struct {
	Status  int
	Code    Code
	Message string
	Cause   error
}

func newError(status int, code Code, message string, cause error) *Error {
	if code == "" {
		code = ErrCodeInternalServer
	}

	if message == "" {
		message = http.StatusText(status)
	}

	return &Error{
		Status:  status,
		Code:    code,
		Message: message,
		Cause:   cause,
	}
}

func (e *Error) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("[%s] %s: %v", e.Code, e.Message, e.Cause)
	}

	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

func (e *Error) Unwrap() error {
	return e.Cause
}

func (e *Error) Is(target error) bool {
	t, ok := target.(*Error)
	if !ok || t == nil {
		return errors.Is(e.Cause, target)
	}

	return e.Code == t.Code
}

func BadRequest(code Code, message string, cause error) *Error {
	return newError(http.StatusBadRequest, code, message, cause)
}

func Unauthorized(code Code, message string, cause error) *Error {
	return newError(http.StatusUnauthorized, code, message, cause)
}

func Forbidden(code Code, message string, cause error) *Error {
	return newError(http.StatusForbidden, code, message, cause)
}

func NotFound(code Code, message string, cause error) *Error {
	return newError(http.StatusNotFound, code, message, cause)
}

func MethodNotAllowed(code Code, message string, cause error) *Error {
	return newError(http.StatusMethodNotAllowed, code, message, cause)
}

func RequestTimeout(code Code, message string, cause error) *Error {
	return newError(http.StatusRequestTimeout, code, message, cause)
}

func Conflict(code Code, message string, cause error) *Error {
	return newError(http.StatusConflict, code, message, cause)
}

func PreconditionFailed(code Code, message string, cause error) *Error {
	return newError(http.StatusPreconditionFailed, code, message, cause)
}

func UnprocessableEntity(code Code, message string, cause error) *Error {
	return newError(http.StatusUnprocessableEntity, code, message, cause)
}

func PreconditionRequired(code Code, message string, cause error) *Error {
	return newError(http.StatusPreconditionRequired, code, message, cause)
}

func TooManyRequests(code Code, message string, cause error) *Error {
	return newError(http.StatusTooManyRequests, code, message, cause)
}

func InternalServerError(code Code, message string, cause error) *Error {
	return newError(http.StatusInternalServerError, code, message, cause)
}

func NotImplemented(code Code, message string, cause error) *Error {
	return newError(http.StatusNotImplemented, code, message, cause)
}

func BadGateway(code Code, message string, cause error) *Error {
	return newError(http.StatusBadGateway, code, message, cause)
}

func ServiceUnavailable(code Code, message string, cause error) *Error {
	return newError(http.StatusServiceUnavailable, code, message, cause)
}

func GatewayTimeout(code Code, message string, cause error) *Error {
	return newError(http.StatusGatewayTimeout, code, message, cause)
}
