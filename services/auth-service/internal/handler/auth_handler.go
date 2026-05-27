package handler

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/go-playground/validator/v10"

	"github.com/4yushraman-jpg/auth-service/internal/dto"
	appErrors "github.com/4yushraman-jpg/auth-service/internal/errors"
	"github.com/4yushraman-jpg/auth-service/internal/response"
)

type AuthHandler struct {
	authService AuthService
	validate    *validator.Validate
}

type AuthService interface {
	Register(ctx context.Context, email, password string) error
	Login(ctx context.Context, email, password string) (*dto.AuthTokensResponse, error)
	Refresh(ctx context.Context, refreshToken string) (*dto.AuthTokensResponse, error)
	Logout(ctx context.Context, refreshToken string) error
}

func NewAuthHandler(
	authService AuthService,
) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		validate:    validator.New(),
	}
}

func (h *AuthHandler) Register(
	w http.ResponseWriter,
	r *http.Request,
) {
	var request dto.RegisterRequest

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := h.validate.Struct(request); err != nil {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	err := h.authService.Register(
		r.Context(),
		request.Email,
		request.Password,
	)

	if err != nil {
		httpErr := appErrors.ToHTTPError(err)
		response.Error(w, httpErr.Status, httpErr.Message)
		return
	}

	response.Success(w, http.StatusCreated, map[string]string{
		"message": "user registered successfully",
	})
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var request dto.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if err := h.validate.Struct(request); err != nil {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	tokens, err := h.authService.Login(r.Context(), request.Email, request.Password)
	if err != nil {
		httpErr := appErrors.ToHTTPError(err)
		response.Error(w, httpErr.Status, httpErr.Message)
		return
	}
	response.Success(w, http.StatusOK, tokens)
}

func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	var request dto.RefreshRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if err := h.validate.Struct(request); err != nil {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	tokens, err := h.authService.Refresh(r.Context(), request.RefreshToken)
	if err != nil {
		httpErr := appErrors.ToHTTPError(err)
		response.Error(w, httpErr.Status, httpErr.Message)
		return
	}
	response.Success(w, http.StatusOK, tokens)
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	var request dto.LogoutRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if err := h.validate.Struct(request); err != nil {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := h.authService.Logout(r.Context(), request.RefreshToken); err != nil {
		httpErr := appErrors.ToHTTPError(err)
		response.Error(w, httpErr.Status, httpErr.Message)
		return
	}
	response.Success(w, http.StatusOK, map[string]string{"message": "logged out"})
}
