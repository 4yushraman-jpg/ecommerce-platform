package handler

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"

	"github.com/4yushraman-jpg/product-service/internal/dto"
	"github.com/4yushraman-jpg/product-service/internal/validator"
)

type fakeProductService struct{}

func (f *fakeProductService) Create(ctx context.Context, req dto.CreateProductRequest) (*dto.ProductResponse, error) {
	return &dto.ProductResponse{ID: uuid.NewString(), Name: req.Name, Description: req.Description, Price: req.Price, InventoryCount: req.InventoryCount, Category: req.Category}, nil
}
func (f *fakeProductService) GetByID(ctx context.Context, id uuid.UUID) (*dto.ProductResponse, error) {
	return &dto.ProductResponse{ID: id.String(), Name: "N"}, nil
}
func (f *fakeProductService) List(ctx context.Context, query dto.ListProductsQuery) (*dto.ListProductsResponse, error) {
	return &dto.ListProductsResponse{}, nil
}
func (f *fakeProductService) Update(ctx context.Context, id uuid.UUID, req dto.UpdateProductRequest) error {
	return nil
}
func (f *fakeProductService) Delete(ctx context.Context, id uuid.UUID) error {
	return nil
}

func TestProductHandlerCreate(t *testing.T) {
	h := NewProductHandler(&fakeProductService{}, validator.New())
	tests := []struct {
		name       string
		body       string
		wantStatus int
	}{
		{name: "invalid json", body: "{", wantStatus: http.StatusBadRequest},
		{name: "validation error", body: `{"name":"","price":0,"inventory_count":-1,"category":""}`, wantStatus: http.StatusBadRequest},
		{name: "success", body: `{"name":"Mouse","description":"Wireless","price":49.99,"inventory_count":10,"category":"electronics"}`, wantStatus: http.StatusCreated},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/api/v1/products", bytes.NewBufferString(tc.body))
			rec := httptest.NewRecorder()
			h.Create(rec, req)
			if rec.Code != tc.wantStatus {
				t.Fatalf("expected %d, got %d", tc.wantStatus, rec.Code)
			}
		})
	}
}

func TestProductHandlerList(t *testing.T) {
	h := NewProductHandler(&fakeProductService{}, validator.New())
	tests := []struct {
		name       string
		target     string
		wantStatus int
	}{
		{name: "success default query", target: "/api/v1/products", wantStatus: http.StatusOK},
		{name: "invalid limit", target: "/api/v1/products?limit=0", wantStatus: http.StatusBadRequest},
		{name: "invalid sort", target: "/api/v1/products?sort=sideways", wantStatus: http.StatusBadRequest},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tc.target, nil)
			rec := httptest.NewRecorder()
			h.List(rec, req)
			if rec.Code != tc.wantStatus {
				t.Fatalf("expected %d, got %d", tc.wantStatus, rec.Code)
			}
		})
	}
}
