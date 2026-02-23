package response

import (
	"encoding/json"
	"errors"
	"net/http"

	"go-auth/internal/apperror"
)

type Meta struct {
	Total int `json:"total"`
	Page  int `json:"page"`
	Limit int `json:"limit"`
}

type successResponse struct {
	Data any   `json:"data"`
	Meta *Meta `json:"meta,omitempty"`
}

type errorBody struct {
	Code    apperror.Code `json:"code"`
	Message string        `json:"message"`
}

type errorResponse struct {
	Error errorBody `json:"error"`
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	js, err := json.Marshal(data)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)

		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, _ = w.Write(js)
}

func OK(w http.ResponseWriter, data any) {
	writeJSON(w, http.StatusOK, successResponse{Data: data})
}

func OKWithMeta(w http.ResponseWriter, data any, meta *Meta) {
	writeJSON(w, http.StatusOK, successResponse{Data: data, Meta: meta})
}

func Created(w http.ResponseWriter, data any) {
	writeJSON(w, http.StatusCreated, successResponse{Data: data})
}

func Accepted(w http.ResponseWriter, data any) {
	writeJSON(w, http.StatusAccepted, successResponse{Data: data})
}

func NoContent(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNoContent)
}

func Error(w http.ResponseWriter, err error) {
	var appErr *apperror.Error

	if errors.As(err, &appErr) {
		writeJSON(w, appErr.Status, errorResponse{
			Error: errorBody{
				Code:    appErr.Code,
				Message: appErr.Message,
			},
		})

		return
	}

	writeJSON(w, http.StatusInternalServerError, errorResponse{
		Error: errorBody{
			Code:    apperror.ErrCodeInternalServer,
			Message: "Something went wrong",
		},
	})
}
