package handler

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"

	"go-auth/internal/apperror"
	"go-auth/internal/response"
	"go-auth/internal/service"
)

type AuthHandler struct {
	svc      service.Service
	validate *validator.Validate
}

func NewAuthHandler(svc service.Service) *AuthHandler {
	return &AuthHandler{
		svc:      svc,
		validate: validator.New(),
	}
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	body, ok := decodeAndValidate[registerRequest](w, r, h.validate)
	if !ok {
		return
	}

	resp, err := h.svc.Register(r.Context(), &service.RegisterRequest{
		Username:  strings.TrimSpace(body.Username),
		Email:     strings.TrimSpace(body.Email),
		Password:  body.Password,
		FirstName: strings.TrimSpace(body.FirstName),
		LastName:  strings.TrimSpace(body.LastName),
	})
	if err != nil {
		response.Error(w, err)

		return
	}

	response.Created(w, registerResponse{UserID: resp.UserID.String()})
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	body, ok := decodeAndValidate[loginRequest](w, r, h.validate)
	if !ok {
		return
	}

	userAgent := r.Header.Get("User-Agent")
	clientIP := clientIP(r)

	resp, err := h.svc.Login(r.Context(), &service.LoginRequest{
		Login:     strings.TrimSpace(body.Login),
		Password:  body.Password,
		UserAgent: userAgent,
		ClientIP:  clientIP,
	})
	if err != nil {
		response.Error(w, err)

		return
	}

	response.OK(w, loginResponse{
		UserID:           resp.UserID.String(),
		AccessToken:      resp.AccessToken,
		RefreshToken:     resp.RefreshToken,
		AccessExpiresAt:  resp.AccessExpiresAt.Format(time.RFC3339),
		RefreshExpiresAt: resp.RefreshExpiresAt.Format(time.RFC3339),
	})
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	body, ok := decodeAndValidate[logoutRequest](w, r, h.validate)
	if !ok {
		return
	}

	if err := h.svc.Logout(r.Context(), strings.TrimSpace(body.RefreshToken)); err != nil {
		response.Error(w, err)

		return
	}

	response.NoContent(w)
}

func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	body, ok := decodeAndValidate[refreshRequest](w, r, h.validate)
	if !ok {
		return
	}

	userAgent := r.Header.Get("User-Agent")
	clientIP := clientIP(r)

	resp, err := h.svc.Refresh(r.Context(), &service.RefreshRequest{
		RefreshToken: strings.TrimSpace(body.RefreshToken),
		UserAgent:    userAgent,
		ClientIP:     clientIP,
	})
	if err != nil {
		response.Error(w, err)

		return
	}

	response.OK(w, refreshResponse{
		AccessToken:      resp.AccessToken,
		RefreshToken:     resp.RefreshToken,
		AccessExpiresAt:  resp.AccessExpiresAt.Format(time.RFC3339),
		RefreshExpiresAt: resp.RefreshExpiresAt.Format(time.RFC3339),
	})
}

func clientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		if i := strings.Index(xff, ","); i > 0 {
			return strings.TrimSpace(xff[:i])
		}

		return strings.TrimSpace(xff)
	}

	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return strings.TrimSpace(xri)
	}

	return strings.TrimSpace(r.RemoteAddr)
}

func decodeAndValidate[T any](w http.ResponseWriter, r *http.Request, validate *validator.Validate) (T, bool) {
	var body T

	if r.Body == nil {
		response.Error(w, apperror.BadRequest(apperror.ErrCodeInvalidJSON, apperror.MsgRequestBodyRequired, nil))

		return body, false
	}

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		response.Error(w, apperror.BadRequest(apperror.ErrCodeInvalidJSON, apperror.MsgInvalidJSON, err))

		return body, false
	}

	if validate != nil {
		if err := validate.Struct(&body); err != nil {
			response.Error(w, apperror.BadRequest(apperror.ErrCodeInvalidParam, err.Error(), err))

			return body, false
		}
	}

	return body, true
}
