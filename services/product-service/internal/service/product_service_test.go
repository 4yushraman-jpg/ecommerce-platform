package service

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"

	"github.com/4yushraman-jpg/product-service/internal/dto"
	appErrors "github.com/4yushraman-jpg/product-service/internal/errors"
	"github.com/4yushraman-jpg/product-service/internal/model"
	"github.com/4yushraman-jpg/product-service/internal/repository"
)

type fakeRepo struct {
	product   *model.Product
	getCalls  int
	listCalls int
}

func (f *fakeRepo) Create(ctx context.Context, product *model.Product) error { return nil }
func (f *fakeRepo) GetByID(ctx context.Context, id uuid.UUID) (*model.Product, error) {
	f.getCalls++
	return f.product, nil
}
func (f *fakeRepo) List(ctx context.Context, params repository.ProductListParams) ([]model.Product, error) {
	f.listCalls++
	return []model.Product{*f.product}, nil
}
func (f *fakeRepo) Update(ctx context.Context, id uuid.UUID, req repository.ProductUpdateParams) error {
	return nil
}
func (f *fakeRepo) Delete(ctx context.Context, id uuid.UUID) error { return nil }

func TestGetByID_UsesCacheAside(t *testing.T) {
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("start miniredis: %v", err)
	}
	defer mr.Close()

	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	id := uuid.New()
	repo := &fakeRepo{product: &model.Product{
		ID: id, Name: "Keyboard", Description: "Mechanical", Price: 199.99, InventoryCount: 5, Category: "electronics",
		CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC(),
	}}
	svc := NewProductService(repo, rdb, time.Minute, &NoopEventPublisher{})

	if _, err := svc.GetByID(context.Background(), id); err != nil {
		t.Fatalf("first get failed: %v", err)
	}
	if _, err := svc.GetByID(context.Background(), id); err != nil {
		t.Fatalf("second get failed: %v", err)
	}
	if repo.getCalls != 1 {
		t.Fatalf("expected 1 db call with cache hit, got %d", repo.getCalls)
	}
}

func TestList_InvalidCursor(t *testing.T) {
	mr, _ := miniredis.Run()
	defer mr.Close()
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	svc := NewProductService(&fakeRepo{}, rdb, time.Minute, &NoopEventPublisher{})
	_, err := svc.List(context.Background(), dto.ListProductsQuery{Limit: 20, Cursor: "invalid"})
	if err == nil || err != appErrors.ErrInvalidCursor {
		t.Fatalf("expected invalid cursor error, got %v", err)
	}
}

func TestList_UsesCacheAside(t *testing.T) {
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("start miniredis: %v", err)
	}
	defer mr.Close()

	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	product := &model.Product{
		ID:             uuid.New(),
		Name:           "Mouse",
		Description:    "Wireless",
		Price:          49.99,
		InventoryCount: 10,
		Category:       "electronics",
		CreatedAt:      time.Now().UTC(),
		UpdatedAt:      time.Now().UTC(),
	}
	repo := &fakeRepo{product: product}
	svc := NewProductService(repo, rdb, time.Minute, &NoopEventPublisher{})

	query := dto.ListProductsQuery{Limit: 20, Sort: "desc", Category: "electronics", Search: "mouse"}
	if _, err := svc.List(context.Background(), query); err != nil {
		t.Fatalf("first list failed: %v", err)
	}
	if _, err := svc.List(context.Background(), query); err != nil {
		t.Fatalf("second list failed: %v", err)
	}
	if repo.listCalls != 1 {
		t.Fatalf("expected 1 db call with cache hit, got %d", repo.listCalls)
	}
}
