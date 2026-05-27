package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.opentelemetry.io/otel"

	appErrors "github.com/4yushraman-jpg/product-service/internal/errors"
	"github.com/4yushraman-jpg/product-service/internal/model"
	"github.com/4yushraman-jpg/product-service/internal/tracing"
)

type ProductListParams struct {
	Limit    int
	CursorAt *time.Time
	CursorID *uuid.UUID
	Category string
	Search   string
	SortAsc  bool
}

type ProductUpdateParams struct {
	Name           *string
	Description    *string
	Price          *float64
	InventoryCount *int
	Category       *string
}

type ProductRepository struct {
	db *pgxpool.Pool
}

const (
	createProductQuery = `
		INSERT INTO products (
			id,
			name,
			description,
			price,
			inventory_count,
			category,
			created_at,
			updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	getProductByIDQuery = `
		SELECT id, name, description, price::float8, inventory_count, category, created_at, updated_at
		FROM products WHERE id = $1
	`
	deleteProductQuery = `DELETE FROM products WHERE id = $1`
)

func NewProductRepository(db *pgxpool.Pool) *ProductRepository {
	return &ProductRepository{db: db}
}

func (r *ProductRepository) Create(ctx context.Context, product *model.Product) error {
	ctx, span := otel.Tracer("product-service/repository").Start(ctx, tracing.SpanCreateProduct)
	defer span.End()
	_, err := r.db.Exec(
		ctx,
		createProductQuery,
		product.ID,
		product.Name,
		product.Description,
		product.Price,
		product.InventoryCount,
		product.Category,
		product.CreatedAt,
		product.UpdatedAt,
	)
	return err
}

func (r *ProductRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.Product, error) {
	ctx, span := otel.Tracer("product-service/repository").Start(ctx, tracing.SpanGetProduct)
	defer span.End()
	var product model.Product
	err := r.db.QueryRow(ctx, getProductByIDQuery, id).Scan(
		&product.ID,
		&product.Name,
		&product.Description,
		&product.Price,
		&product.InventoryCount,
		&product.Category,
		&product.CreatedAt,
		&product.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, appErrors.ErrProductNotFound
		}
		return nil, err
	}
	return &product, nil
}

func (r *ProductRepository) List(ctx context.Context, params ProductListParams) ([]model.Product, error) {
	ctx, span := otel.Tracer("product-service/repository").Start(ctx, tracing.SpanListProducts)
	defer span.End()
	args := make([]any, 0, 4)
	where := make([]string, 0, 3)

	if params.Category != "" {
		args = append(args, params.Category)
		where = append(where, fmt.Sprintf("category = $%d", len(args)))
	}
	if params.Search != "" {
		args = append(args, "%"+strings.ToLower(params.Search)+"%")
		where = append(where, fmt.Sprintf("name ILIKE $%d", len(args)))
	}
	if params.CursorAt != nil && params.CursorID != nil {
		args = append(args, *params.CursorAt, *params.CursorID)
		if params.SortAsc {
			where = append(where, fmt.Sprintf("(created_at, id) > ($%d, $%d)", len(args)-1, len(args)))
		} else {
			where = append(where, fmt.Sprintf("(created_at, id) < ($%d, $%d)", len(args)-1, len(args)))
		}
	}

	order := "ORDER BY created_at DESC, id DESC"
	if params.SortAsc {
		order = "ORDER BY created_at ASC, id ASC"
	}

	query := `
		SELECT id, name, description, price::float8, inventory_count, category, created_at, updated_at
		FROM products
	`
	if len(where) > 0 {
		query += " WHERE " + strings.Join(where, " AND ")
	}

	args = append(args, params.Limit+1)
	query += " " + order + fmt.Sprintf(" LIMIT $%d", len(args))

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]model.Product, 0, params.Limit+1)
	for rows.Next() {
		var p model.Product
		if err := rows.Scan(&p.ID, &p.Name, &p.Description, &p.Price, &p.InventoryCount, &p.Category, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, err
		}
		items = append(items, p)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

func (r *ProductRepository) Update(ctx context.Context, id uuid.UUID, req ProductUpdateParams) error {
	ctx, span := otel.Tracer("product-service/repository").Start(ctx, tracing.SpanUpdateProduct)
	defer span.End()
	setParts := make([]string, 0, 6)
	args := make([]any, 0, 5)

	if req.Name != nil {
		args = append(args, *req.Name)
		setParts = append(setParts, fmt.Sprintf("name = $%d", len(args)))
	}
	if req.Description != nil {
		args = append(args, *req.Description)
		setParts = append(setParts, fmt.Sprintf("description = $%d", len(args)))
	}
	if req.Price != nil {
		args = append(args, *req.Price)
		setParts = append(setParts, fmt.Sprintf("price = $%d", len(args)))
	}
	if req.InventoryCount != nil {
		args = append(args, *req.InventoryCount)
		setParts = append(setParts, fmt.Sprintf("inventory_count = $%d", len(args)))
	}
	if req.Category != nil {
		args = append(args, *req.Category)
		setParts = append(setParts, fmt.Sprintf("category = $%d", len(args)))
	}
	if len(setParts) == 0 {
		return appErrors.ErrBadRequest
	}

	setParts = append(setParts, "updated_at = NOW()")
	args = append(args, id)
	query := fmt.Sprintf("UPDATE products SET %s WHERE id = $%d", strings.Join(setParts, ", "), len(args))
	cmd, err := r.db.Exec(ctx, query, args...)
	if err != nil {
		return err
	}
	if cmd.RowsAffected() == 0 {
		return appErrors.ErrProductNotFound
	}
	return nil
}

func (r *ProductRepository) Delete(ctx context.Context, id uuid.UUID) error {
	ctx, span := otel.Tracer("product-service/repository").Start(ctx, tracing.SpanDeleteProduct)
	defer span.End()
	cmd, err := r.db.Exec(ctx, deleteProductQuery, id)
	if err != nil {
		return err
	}
	if cmd.RowsAffected() == 0 {
		return appErrors.ErrProductNotFound
	}
	return nil
}
