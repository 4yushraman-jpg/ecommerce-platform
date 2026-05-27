package repository

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/4yushraman-jpg/product-service/internal/model"
)

func TestProductRepositoryIntegration_CreateGetDelete(t *testing.T) {
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("TEST_DATABASE_URL not set")
	}

	ctx := context.Background()
	db, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Fatalf("connect db: %v", err)
	}
	defer db.Close()

	repo := NewProductRepository(db)
	product := &model.Product{
		ID:             uuid.New(),
		Name:           "Integration Product",
		Description:    "integration",
		Price:          10.5,
		InventoryCount: 7,
		Category:       "test",
		CreatedAt:      time.Now().UTC(),
		UpdatedAt:      time.Now().UTC(),
	}

	if err := repo.Create(ctx, product); err != nil {
		t.Fatalf("create product: %v", err)
	}
	got, err := repo.GetByID(ctx, product.ID)
	if err != nil {
		t.Fatalf("get product: %v", err)
	}
	if got.ID != product.ID {
		t.Fatalf("expected id %s, got %s", product.ID, got.ID)
	}
	if err := repo.Delete(ctx, product.ID); err != nil {
		t.Fatalf("delete product: %v", err)
	}
}

func TestProductRepositoryIntegration_UpdateAndList(t *testing.T) {
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("TEST_DATABASE_URL not set")
	}

	ctx := context.Background()
	db, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Fatalf("connect db: %v", err)
	}
	defer db.Close()

	repo := NewProductRepository(db)
	product := &model.Product{
		ID:             uuid.New(),
		Name:           "Integration Update Product",
		Description:    "integration",
		Price:          10.5,
		InventoryCount: 7,
		Category:       "test",
		CreatedAt:      time.Now().UTC(),
		UpdatedAt:      time.Now().UTC(),
	}

	if err := repo.Create(ctx, product); err != nil {
		t.Fatalf("create product: %v", err)
	}
	newName := "Updated Integration Product"
	newCount := 11
	if err := repo.Update(ctx, product.ID, ProductUpdateParams{Name: &newName, InventoryCount: &newCount}); err != nil {
		t.Fatalf("update product: %v", err)
	}
	items, err := repo.List(ctx, ProductListParams{Limit: 10, Category: "test", SortAsc: false})
	if err != nil {
		t.Fatalf("list products: %v", err)
	}
	if len(items) == 0 {
		t.Fatalf("expected list results")
	}
	if err := repo.Delete(ctx, product.ID); err != nil {
		t.Fatalf("delete product: %v", err)
	}
}
