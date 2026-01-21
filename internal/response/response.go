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

func writeJSON(writer http.ResponseWriter, status int, data any) {
	js, err := json.Marshal(data)
	if err != nil {
		http.Error(writer, "Internal Server Error", http.StatusInternalServerError)

		return
	}

	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(status)
	_, _ = writer.Write(js)
}

func OK(writer http.ResponseWriter, data any) {
	writeJSON(writer, http.StatusOK, successResponse{Data: data})
}

func OKWithMeta(writer http.ResponseWriter, data any, meta *Meta) {
	writeJSON(writer, http.StatusOK, successResponse{Data: data, Meta: meta})
}

func Created(writer http.ResponseWriter, data any) {
	writeJSON(writer, http.StatusCreated, successResponse{Data: data})
}

func Accepted(writer http.ResponseWriter, data any) {
	writeJSON(writer, http.StatusAccepted, successResponse{Data: data})
}

func NoContent(writer http.ResponseWriter) {
	writer.WriteHeader(http.StatusNoContent)
}

func Error(writer http.ResponseWriter, err error) {
	var appErr *apperror.Error

	if errors.As(err, &appErr) {
		writeJSON(writer, appErr.Status, errorResponse{
			Error: errorBody{
				Code:    appErr.Code,
				Message: appErr.Message,
			},
		})

		return
	}

	writeJSON(writer, http.StatusInternalServerError, errorResponse{
		Error: errorBody{
			Code:    apperror.ErrCodeInternalServer,
			Message: "Something went wrong",
		},
	})
}
