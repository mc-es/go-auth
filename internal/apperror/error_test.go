package apperror_test

import (
	"errors"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"

	"go-auth/internal/apperror"
)

func TestError(t *testing.T) {
	causeErr := errors.New("sql: no rows in result set")
	tests := []struct {
		name    string
		err     *apperror.Error
		wantStr string
	}{
		{
			name: "format with cause",
			err: apperror.Forbidden(
				apperror.ErrCodeInternalServer,
				"Database query failed",
				causeErr,
			),
			wantStr: fmt.Sprintf("[%s] Database query failed: sql: no rows in result set", apperror.ErrCodeInternalServer),
		},
		{
			name: "format without cause",
			err: apperror.NotFound(
				apperror.ErrCodeUserNotFound,
				"User not found",
				nil,
			),
			wantStr: fmt.Sprintf("[%s] User not found", apperror.ErrCodeUserNotFound),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := tt.err.Error()
			assert.Equal(t, tt.wantStr, got)
		})
	}
}

func TestUnwrap(t *testing.T) {
	causeErr := errors.New("bcrypt: hashedPassword is not the hash of the given password")
	tests := []struct {
		name string
		err  error
		want error
	}{
		{
			name: "unwrap returns the cause",
			err:  apperror.Unauthorized(apperror.ErrCodeInvalidCredentials, "Invalid credentials", causeErr),
			want: causeErr,
		},
		{
			name: "unwrap returns nil if no cause",
			err:  apperror.BadRequest(apperror.ErrCodeInvalidJSON, "Invalid JSON", nil),
			want: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := errors.Unwrap(tt.err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestIs(t *testing.T) {
	causeErr := errors.New("underlying db error")
	tests := []struct {
		name   string
		err    error
		target error
		want   bool
	}{
		{
			name:   "is returns true if the error code is the same",
			err:    apperror.Conflict(apperror.ErrCodeEmailAlreadyUsed, "Email already used", nil),
			target: &apperror.Error{Code: apperror.ErrCodeEmailAlreadyUsed},
			want:   true,
		},
		{
			name:   "is returns false if the error code is different",
			err:    apperror.Conflict(apperror.ErrCodeEmailAlreadyUsed, "Email already used", nil),
			target: &apperror.Error{Code: apperror.ErrCodeInvalidParam},
			want:   false,
		},
		{
			name:   "is returns true when target is the cause",
			err:    apperror.InternalServerError(apperror.ErrCodeInternalServer, "DB failed", causeErr),
			target: causeErr,
			want:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := errors.Is(tt.err, tt.target)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestNormalization(t *testing.T) {
	tests := []struct {
		name        string
		err         *apperror.Error
		wantStatus  int
		wantCode    apperror.Code
		wantMessage string
	}{
		{
			name:        "empty code becomes ErrCodeInternalServer",
			err:         apperror.BadRequest("", "some message", nil),
			wantStatus:  http.StatusBadRequest,
			wantCode:    apperror.ErrCodeInternalServer,
			wantMessage: "some message",
		},
		{
			name:        "empty message becomes http.StatusText(status)",
			err:         apperror.NotFound(apperror.ErrCodeUserNotFound, "", nil),
			wantStatus:  http.StatusNotFound,
			wantCode:    apperror.ErrCodeUserNotFound,
			wantMessage: http.StatusText(http.StatusNotFound),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := tt.err

			assert.Equal(t, tt.wantStatus, got.Status)
			assert.Equal(t, tt.wantCode, got.Code)
			assert.Equal(t, tt.wantMessage, got.Message)
		})
	}
}

func TestErrorFactories(t *testing.T) {
	tests := []struct {
		name       string
		factory    func(code apperror.Code, message string, cause error) *apperror.Error
		wantStatus int
	}{
		{name: "BadRequest", factory: apperror.BadRequest, wantStatus: http.StatusBadRequest},
		{name: "Unauthorized", factory: apperror.Unauthorized, wantStatus: http.StatusUnauthorized},
		{name: "Forbidden", factory: apperror.Forbidden, wantStatus: http.StatusForbidden},
		{name: "NotFound", factory: apperror.NotFound, wantStatus: http.StatusNotFound},
		{name: "MethodNotAllowed", factory: apperror.MethodNotAllowed, wantStatus: http.StatusMethodNotAllowed},
		{name: "RequestTimeout", factory: apperror.RequestTimeout, wantStatus: http.StatusRequestTimeout},
		{name: "Conflict", factory: apperror.Conflict, wantStatus: http.StatusConflict},
		{name: "PreconditionFailed", factory: apperror.PreconditionFailed, wantStatus: http.StatusPreconditionFailed},
		{name: "UnprocessableEntity", factory: apperror.UnprocessableEntity, wantStatus: http.StatusUnprocessableEntity},
		{name: "PreconditionRequired", factory: apperror.PreconditionRequired, wantStatus: http.StatusPreconditionRequired},
		{name: "TooManyRequests", factory: apperror.TooManyRequests, wantStatus: http.StatusTooManyRequests},
		{name: "InternalServerError", factory: apperror.InternalServerError, wantStatus: http.StatusInternalServerError},
		{name: "NotImplemented", factory: apperror.NotImplemented, wantStatus: http.StatusNotImplemented},
		{name: "BadGateway", factory: apperror.BadGateway, wantStatus: http.StatusBadGateway},
		{name: "ServiceUnavailable", factory: apperror.ServiceUnavailable, wantStatus: http.StatusServiceUnavailable},
		{name: "GatewayTimeout", factory: apperror.GatewayTimeout, wantStatus: http.StatusGatewayTimeout},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := tt.factory("TEST_CODE", "test message", nil)

			assert.Equal(t, tt.wantStatus, got.Status)
			assert.Equal(t, apperror.Code("TEST_CODE"), got.Code)
		})
	}
}
