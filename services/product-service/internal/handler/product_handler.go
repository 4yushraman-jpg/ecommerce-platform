package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/4yushraman-jpg/product-service/internal/dto"
	appErrors "github.com/4yushraman-jpg/product-service/internal/errors"
	"github.com/4yushraman-jpg/product-service/internal/response"
	"github.com/4yushraman-jpg/product-service/internal/validator"
)

type ProductService interface {
	Create(ctx context.Context, req dto.CreateProductRequest) (*dto.ProductResponse, error)
	GetByID(ctx context.Context, id uuid.UUID) (*dto.ProductResponse, error)
	List(ctx context.Context, query dto.ListProductsQuery) (*dto.ListProductsResponse, error)
	Update(ctx context.Context, id uuid.UUID, req dto.UpdateProductRequest) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type ProductHandler struct {
	service   ProductService
	validator *validator.Validator
}

func NewProductHandler(service ProductService, validator *validator.Validator) *ProductHandler {
	return &ProductHandler{service: service, validator: validator}
}

func (h *ProductHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateProductRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if err := h.validator.Struct(req); err != nil {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}
	out, err := h.service.Create(r.Context(), req)
	if err != nil {
		httpErr := appErrors.ToHTTPError(err)
		response.Error(w, httpErr.Status, httpErr.Message)
		return
	}
	response.Success(w, http.StatusCreated, out)
}

func (h *ProductHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid product id")
		return
	}
	out, err := h.service.GetByID(r.Context(), id)
	if err != nil {
		httpErr := appErrors.ToHTTPError(err)
		response.Error(w, httpErr.Status, httpErr.Message)
		return
	}
	response.Success(w, http.StatusOK, out)
}

func (h *ProductHandler) List(w http.ResponseWriter, r *http.Request) {
	limit := 20
	if value := r.URL.Query().Get("limit"); value != "" {
		parsed, err := strconv.Atoi(value)
		if err != nil || parsed < 1 || parsed > 100 {
			response.Error(w, http.StatusBadRequest, "invalid limit")
			return
		}
		limit = parsed
	}
	sort := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("sort")))
	if sort != "" && sort != "asc" && sort != "desc" {
		response.Error(w, http.StatusBadRequest, "invalid sort")
		return
	}

	query := dto.ListProductsQuery{
		Limit:    limit,
		Cursor:   r.URL.Query().Get("cursor"),
		Category: r.URL.Query().Get("category"),
		Sort:     sort,
		Search:   r.URL.Query().Get("search"),
	}
	out, err := h.service.List(r.Context(), query)
	if err != nil {
		httpErr := appErrors.ToHTTPError(err)
		response.Error(w, httpErr.Status, httpErr.Message)
		return
	}
	response.Success(w, http.StatusOK, out)
}

func (h *ProductHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid product id")
		return
	}
	var req dto.UpdateProductRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if err := h.validator.Struct(req); err != nil {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := h.service.Update(r.Context(), id, req); err != nil {
		httpErr := appErrors.ToHTTPError(err)
		response.Error(w, httpErr.Status, httpErr.Message)
		return
	}
	response.Success(w, http.StatusOK, map[string]string{"message": "product updated"})
}

func (h *ProductHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid product id")
		return
	}
	if err := h.service.Delete(r.Context(), id); err != nil {
		httpErr := appErrors.ToHTTPError(err)
		response.Error(w, httpErr.Status, httpErr.Message)
		return
	}
	response.Success(w, http.StatusOK, map[string]string{"message": "product deleted"})
}
