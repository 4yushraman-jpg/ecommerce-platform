package handler

import (
	"net/http"

	"github.com/4yushraman-jpg/auth-service/internal/dto"
	appErrors "github.com/4yushraman-jpg/auth-service/internal/errors"
	"github.com/4yushraman-jpg/auth-service/internal/middleware"
	"github.com/4yushraman-jpg/auth-service/internal/response"
	"github.com/4yushraman-jpg/auth-service/internal/repository"
)

type UserHandler struct {
	userRepo *repository.UserRepository
}

func NewUserHandler(userRepo *repository.UserRepository) *UserHandler {
	return &UserHandler{userRepo: userRepo}
}

func (h *UserHandler) Me(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		response.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	user, err := h.userRepo.GetByID(r.Context(), userCtx.UserID)
	if err != nil {
		httpErr := appErrors.ToHTTPError(err)
		response.Error(w, httpErr.Status, httpErr.Message)
		return
	}

	response.Success(w, http.StatusOK, dto.MeResponse{
		ID:        user.ID.String(),
		Email:     user.Email,
		Role:      user.Role,
		CreatedAt: user.CreatedAt,
	})
}
