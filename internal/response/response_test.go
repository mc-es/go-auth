package response_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"go-auth/internal/apperror"
	"go-auth/internal/response"
)

type testSuccessBody struct {
	Data any `json:"data"`
	Meta *struct {
		Total int `json:"total"`
		Page  int `json:"page"`
		Limit int `json:"limit"`
	} `json:"meta"`
}

type testErrorBody struct {
	Error struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

const (
	headerContentType = "Content-Type"
	contentTypeJSON   = "application/json"
)

func TestOK(t *testing.T) {
	tests := []struct {
		name       string
		data       any
		wantStatus int
		wantData   any
	}{
		{
			name:       "object",
			data:       map[string]string{"id": "1", "name": "test"},
			wantStatus: http.StatusOK,
			wantData:   map[string]any{"id": "1", "name": "test"},
		},
		{
			name:       "slice",
			data:       []int{1, 2, 3},
			wantStatus: http.StatusOK,
			wantData:   []any{float64(1), float64(2), float64(3)},
		},
		{
			name:       "null",
			data:       nil,
			wantStatus: http.StatusOK,
			wantData:   nil,
		},
		{
			name:       "marshal failure returns 500",
			data:       make(chan int),
			wantStatus: http.StatusInternalServerError,
			wantData:   nil, // body is "Internal Server Error\n", not JSON
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			rec := httptest.NewRecorder()
			response.OK(rec, tt.data)

			assert.Equal(t, tt.wantStatus, rec.Code)

			if tt.name == "marshal failure returns 500" {
				assert.Equal(t, "text/plain; charset=utf-8", rec.Header().Get(headerContentType))
				assert.Equal(t, "Internal Server Error\n", rec.Body.String())

				return
			}

			assert.Equal(t, contentTypeJSON, rec.Header().Get(headerContentType))

			var body testSuccessBody

			err := json.Unmarshal(rec.Body.Bytes(), &body)
			assert.NoError(t, err)
			assert.Equal(t, tt.wantData, body.Data)
		})
	}
}

func TestOKWithMeta(t *testing.T) {
	t.Run("with meta", func(t *testing.T) {
		t.Parallel()

		rec := httptest.NewRecorder()
		data := map[string]string{"id": "1"}
		meta := &response.Meta{Total: 100, Page: 1, Limit: 10}

		response.OKWithMeta(rec, data, meta)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, contentTypeJSON, rec.Header().Get(headerContentType))

		var body testSuccessBody

		err := json.Unmarshal(rec.Body.Bytes(), &body)
		assert.NoError(t, err)
		assert.Equal(t, map[string]any{"id": "1"}, body.Data)
		assert.NotNil(t, body.Meta)
		assert.Equal(t, 100, body.Meta.Total)
		assert.Equal(t, 1, body.Meta.Page)
		assert.Equal(t, 10, body.Meta.Limit)
	})

	t.Run("with nil meta", func(t *testing.T) {
		t.Parallel()

		rec := httptest.NewRecorder()
		response.OKWithMeta(rec, "ok", nil)

		assert.Equal(t, http.StatusOK, rec.Code)

		var body testSuccessBody

		err := json.Unmarshal(rec.Body.Bytes(), &body)
		assert.NoError(t, err)
		assert.Equal(t, "ok", body.Data)
		assert.Nil(t, body.Meta)
	})
}

func TestCreated(t *testing.T) {
	t.Run("writes 201 with data", func(t *testing.T) {
		t.Parallel()

		rec := httptest.NewRecorder()
		response.Created(rec, map[string]int{"id": 42})

		assert.Equal(t, http.StatusCreated, rec.Code)
		assert.Equal(t, contentTypeJSON, rec.Header().Get(headerContentType))

		var body testSuccessBody

		err := json.Unmarshal(rec.Body.Bytes(), &body)
		assert.NoError(t, err)
		assert.Equal(t, map[string]any{"id": float64(42)}, body.Data)
	})
}

func TestAccepted(t *testing.T) {
	t.Run("writes 202 with data", func(t *testing.T) {
		t.Parallel()

		rec := httptest.NewRecorder()
		response.Accepted(rec, map[string]string{"status": "pending"})

		assert.Equal(t, http.StatusAccepted, rec.Code)
		assert.Equal(t, contentTypeJSON, rec.Header().Get(headerContentType))

		var body testSuccessBody

		err := json.Unmarshal(rec.Body.Bytes(), &body)
		assert.NoError(t, err)
		assert.Equal(t, map[string]any{"status": "pending"}, body.Data)
	})
}

func TestNoContent(t *testing.T) {
	t.Run("writes 204 with empty body", func(t *testing.T) {
		t.Parallel()

		rec := httptest.NewRecorder()
		response.NoContent(rec)

		assert.Equal(t, http.StatusNoContent, rec.Code)
		assert.Empty(t, rec.Body.Bytes())
	})
}

func TestError(t *testing.T) {
	tests := []struct {
		name        string
		err         error
		wantStatus  int
		wantCode    apperror.Code
		wantMessage string
	}{
		{
			name:        "apperror writes status and body",
			err:         apperror.BadRequest(apperror.ErrCodeInvalidJSON, "Invalid JSON body", nil),
			wantStatus:  http.StatusBadRequest,
			wantCode:    apperror.ErrCodeInvalidJSON,
			wantMessage: "Invalid JSON body",
		},
		{
			name:        "apperror with status 404",
			err:         apperror.NotFound(apperror.ErrCodeUserNotFound, "User not found", nil),
			wantStatus:  http.StatusNotFound,
			wantCode:    apperror.ErrCodeUserNotFound,
			wantMessage: "User not found",
		},
		{
			name:        "generic error writes 500",
			err:         errors.New("something broke"),
			wantStatus:  http.StatusInternalServerError,
			wantCode:    apperror.ErrCodeInternalServer,
			wantMessage: "Something went wrong",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			rec := httptest.NewRecorder()
			response.Error(rec, tt.err)

			assert.Equal(t, tt.wantStatus, rec.Code)
			assert.Equal(t, contentTypeJSON, rec.Header().Get(headerContentType))

			var body testErrorBody

			err := json.Unmarshal(rec.Body.Bytes(), &body)
			assert.NoError(t, err)
			assert.Equal(t, string(tt.wantCode), body.Error.Code)
			assert.Equal(t, tt.wantMessage, body.Error.Message)
		})
	}
}
