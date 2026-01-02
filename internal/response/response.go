// Package response provides a structured response handling system for HTTP APIs.
//
// Usage:
//
//	// Success responses
//	response.OK(writer, user)
//	response.Created(writer, newUser)
//	response.Accepted(writer, map[string]string{"status": "processing"})
//	response.NoContent(writer)
//
//	// Error handling
//	err := httperror.NotFound("User not found")
//	response.Error(writer, request, err)
package response

import (
	"encoding/json"
	"errors"
	"net/http"

	"go-auth/internal/httperror"
	"go-auth/pkg/logger"
)

type customError struct {
	Code    string `json:"code,omitempty"`
	Message string `json:"message"`
}

type errorResponse struct {
	Error customError `json:"error"`
}

func jsonResponse(writer http.ResponseWriter, status int, data any) {
	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(status)

	if data != nil {
		if err := json.NewEncoder(writer).Encode(data); err != nil {
			logger.Error("JSON encode failed", "error", err)
		}
	}
}

// OK sends a 200 OK response.
func OK(writer http.ResponseWriter, data any) {
	jsonResponse(writer, http.StatusOK, data)
}

// Created sends a 201 Created response.
func Created(writer http.ResponseWriter, data any) {
	jsonResponse(writer, http.StatusCreated, data)
}

// Accepted sends a 202 Accepted response.
func Accepted(writer http.ResponseWriter, data any) {
	jsonResponse(writer, http.StatusAccepted, data)
}

// NoContent sends a 204 No Content response.
func NoContent(writer http.ResponseWriter) {
	jsonResponse(writer, http.StatusNoContent, nil)
}

// Error sends a standardized JSON error response and handles logging.
//
// Behavior:
//   - Logs 5xx errors as ERROR and 4xx errors as WARN.
//   - Defaults to 500 Internal Server Error for unknown error types.
//
// Output: { "error": { "code": "...", "message": "..." } }
//
// Example:
//
//	err := httperror.NotFound("User not found").
//	    WithCode(httperror.ErrCodeUserNotFound)
//	response.Error(writer, request, err) // in handler function
func Error(writer http.ResponseWriter, request *http.Request, err error) {
	statusCode := http.StatusInternalServerError

	errBody := customError{
		Code:    string(httperror.ErrCodeInternalServer),
		Message: http.StatusText(statusCode),
	}

	var customErr *httperror.Error
	if errors.As(err, &customErr) {
		statusCode = customErr.Status
		errBody = customError{
			Code:    string(customErr.Code),
			Message: customErr.Message,
		}
	}

	if statusCode >= 500 {
		logger.Error("Server Error",
			"path", request.URL.Path,
			"status", statusCode,
			"error", err,
		)
	} else {
		logger.Warn("Client Error",
			"path", request.URL.Path,
			"status", statusCode,
			"code", errBody.Code,
		)
	}

	jsonResponse(writer, statusCode, errorResponse{Error: errBody})
}
