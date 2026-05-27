package service

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"go.opentelemetry.io/otel"

	"github.com/4yushraman-jpg/product-service/internal/dto"
	appErrors "github.com/4yushraman-jpg/product-service/internal/errors"
	"github.com/4yushraman-jpg/product-service/internal/mapper"
	"github.com/4yushraman-jpg/product-service/internal/model"
	"github.com/4yushraman-jpg/product-service/internal/repository"
	"github.com/4yushraman-jpg/product-service/internal/tracing"
)

type ProductService struct {
	repo      productRepository
	redis     *redis.Client
	cacheTTL  time.Duration
	publisher EventPublisher
}

type productRepository interface {
	Create(ctx context.Context, product *model.Product) error
	GetByID(ctx context.Context, id uuid.UUID) (*model.Product, error)
	List(ctx context.Context, params repository.ProductListParams) ([]model.Product, error)
	Update(ctx context.Context, id uuid.UUID, req repository.ProductUpdateParams) error
	Delete(ctx context.Context, id uuid.UUID) error
}

func NewProductService(repo productRepository, redis *redis.Client, cacheTTL time.Duration, publisher EventPublisher) *ProductService {
	return &ProductService{repo: repo, redis: redis, cacheTTL: cacheTTL, publisher: publisher}
}

func (s *ProductService) Create(ctx context.Context, req dto.CreateProductRequest) (*dto.ProductResponse, error) {
	ctx, span := otel.Tracer("product-service/service").Start(ctx, tracing.SpanCreateProduct)
	defer span.End()

	now := time.Now().UTC()
	product := &model.Product{
		ID:             uuid.New(),
		Name:           req.Name,
		Description:    req.Description,
		Price:          req.Price,
		InventoryCount: req.InventoryCount,
		Category:       req.Category,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
	if err := s.repo.Create(ctx, product); err != nil {
		return nil, err
	}
	s.invalidateListCache(ctx)

	_ = s.publisher.Publish(ctx, model.ProductEvent{
		Type:      model.ProductCreatedEventType,
		ProductID: product.ID.String(),
		At:        now,
	})
	resp := mapper.ToProductResponse(product)
	return &resp, nil
}

func (s *ProductService) GetByID(ctx context.Context, id uuid.UUID) (*dto.ProductResponse, error) {
	ctx, span := otel.Tracer("product-service/service").Start(ctx, tracing.SpanGetProduct)
	defer span.End()

	cacheKey := s.detailCacheKey(id)
	if cached, ok := s.getCachedProduct(ctx, cacheKey); ok {
		return cached, nil
	}

	product, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	out := mapper.ToProductResponse(product)
	s.setCachedProduct(ctx, cacheKey, &out)
	return &out, nil
}

func (s *ProductService) List(ctx context.Context, query dto.ListProductsQuery) (*dto.ListProductsResponse, error) {
	ctx, span := otel.Tracer("product-service/service").Start(ctx, tracing.SpanListProducts)
	defer span.End()

	if query.Limit <= 0 {
		return nil, appErrors.ErrBadRequest
	}
	sort := strings.ToLower(strings.TrimSpace(query.Sort))
	if sort == "" {
		sort = "desc"
	}
	if sort != "asc" && sort != "desc" {
		return nil, appErrors.ErrInvalidSort
	}
	sortAsc := sort == "asc"
	cursorAt, cursorID, err := decodeCursor(query.Cursor)
	if err != nil {
		return nil, err
	}

	cacheKey := s.listCacheKey(query.Limit, query.Cursor, query.Category, sort, query.Search)
	if cached, ok := s.getCachedList(ctx, cacheKey); ok {
		return cached, nil
	}

	items, err := s.repo.List(ctx, repository.ProductListParams{
		Limit:    query.Limit,
		CursorAt: cursorAt,
		CursorID: cursorID,
		Category: query.Category,
		SortAsc:  sortAsc,
		Search:   query.Search,
	})
	if err != nil {
		return nil, err
	}

	hasMore := len(items) > query.Limit
	if hasMore {
		items = items[:query.Limit]
	}

	respItems := make([]dto.ProductResponse, 0, len(items))
	for i := range items {
		respItems = append(respItems, mapper.ToProductResponse(&items[i]))
	}

	resp := &dto.ListProductsResponse{Items: respItems}
	if hasMore {
		last := items[len(items)-1]
		resp.NextCursor = encodeCursor(last.CreatedAt, last.ID)
	}

	s.setCachedList(ctx, cacheKey, resp)
	return resp, nil
}

func (s *ProductService) Update(ctx context.Context, id uuid.UUID, req dto.UpdateProductRequest) error {
	ctx, span := otel.Tracer("product-service/service").Start(ctx, tracing.SpanUpdateProduct)
	defer span.End()
	updates := repository.ProductUpdateParams{}
	publishProductUpdated := false
	publishInventoryUpdated := false
	if req.Name != nil {
		updates.Name = req.Name
		publishProductUpdated = true
	}
	if req.Description != nil {
		updates.Description = req.Description
		publishProductUpdated = true
	}
	if req.Price != nil {
		updates.Price = req.Price
		publishProductUpdated = true
	}
	if req.InventoryCount != nil {
		updates.InventoryCount = req.InventoryCount
		publishInventoryUpdated = true
	}
	if req.Category != nil {
		updates.Category = req.Category
		publishProductUpdated = true
	}
	if !publishProductUpdated && !publishInventoryUpdated {
		return appErrors.ErrNoUpdateFields
	}

	if err := s.repo.Update(ctx, id, updates); err != nil {
		return err
	}

	s.deleteCachedProduct(ctx, s.detailCacheKey(id))
	s.invalidateListCache(ctx)
	if publishProductUpdated {
		_ = s.publisher.Publish(ctx, model.ProductEvent{Type: model.ProductUpdatedEventType, ProductID: id.String(), At: time.Now().UTC()})
	}
	if publishInventoryUpdated {
		_ = s.publisher.Publish(ctx, model.ProductEvent{Type: model.InventoryUpdatedEventType, ProductID: id.String(), At: time.Now().UTC()})
	}
	return nil
}

func (s *ProductService) Delete(ctx context.Context, id uuid.UUID) error {
	ctx, span := otel.Tracer("product-service/service").Start(ctx, tracing.SpanDeleteProduct)
	defer span.End()
	if err := s.repo.Delete(ctx, id); err != nil {
		return err
	}
	s.deleteCachedProduct(ctx, s.detailCacheKey(id))
	s.invalidateListCache(ctx)
	return nil
}

func (s *ProductService) detailCacheKey(id uuid.UUID) string {
	return "product:detail:" + id.String()
}

func (s *ProductService) listCacheKey(limit int, cursor, category, sort, search string) string {
	return fmt.Sprintf("product:list:%d:%s:%s:%s:%s", limit, cursor, category, sort, search)
}

func (s *ProductService) getCachedProduct(ctx context.Context, cacheKey string) (*dto.ProductResponse, bool) {
	if s.redis == nil {
		return nil, false
	}
	ctx, span := otel.Tracer("product-service/cache").Start(ctx, "cache.product.detail.get")
	defer span.End()
	cached, err := s.redis.Get(ctx, cacheKey).Result()
	if err != nil {
		return nil, false
	}
	var out dto.ProductResponse
	if err := json.Unmarshal([]byte(cached), &out); err != nil {
		return nil, false
	}
	return &out, true
}

func (s *ProductService) setCachedProduct(ctx context.Context, cacheKey string, resp *dto.ProductResponse) {
	if s.redis == nil {
		return
	}
	ctx, span := otel.Tracer("product-service/cache").Start(ctx, "cache.product.detail.set")
	defer span.End()
	payload, err := json.Marshal(resp)
	if err != nil {
		return
	}
	_ = s.redis.Set(ctx, cacheKey, payload, s.cacheTTL).Err()
}

func (s *ProductService) getCachedList(ctx context.Context, cacheKey string) (*dto.ListProductsResponse, bool) {
	if s.redis == nil {
		return nil, false
	}
	ctx, span := otel.Tracer("product-service/cache").Start(ctx, "cache.product.list.get")
	defer span.End()
	cached, err := s.redis.Get(ctx, cacheKey).Result()
	if err != nil {
		return nil, false
	}
	var out dto.ListProductsResponse
	if err := json.Unmarshal([]byte(cached), &out); err != nil {
		return nil, false
	}
	return &out, true
}

func (s *ProductService) setCachedList(ctx context.Context, cacheKey string, resp *dto.ListProductsResponse) {
	if s.redis == nil {
		return
	}
	ctx, span := otel.Tracer("product-service/cache").Start(ctx, "cache.product.list.set")
	defer span.End()
	payload, err := json.Marshal(resp)
	if err != nil {
		return
	}
	_ = s.redis.Set(ctx, cacheKey, payload, s.cacheTTL).Err()
}

func encodeCursor(createdAt time.Time, id uuid.UUID) string {
	raw := fmt.Sprintf("%d|%s", createdAt.UnixNano(), id.String())
	return base64.RawURLEncoding.EncodeToString([]byte(raw))
}

func decodeCursor(cursor string) (*time.Time, *uuid.UUID, error) {
	if cursor == "" {
		return nil, nil, nil
	}
	decoded, err := base64.RawURLEncoding.DecodeString(cursor)
	if err != nil {
		return nil, nil, appErrors.ErrInvalidCursor
	}
	parts := strings.Split(string(decoded), "|")
	if len(parts) != 2 {
		return nil, nil, appErrors.ErrInvalidCursor
	}
	ns, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return nil, nil, appErrors.ErrInvalidCursor
	}
	id, err := uuid.Parse(parts[1])
	if err != nil {
		return nil, nil, appErrors.ErrInvalidCursor
	}
	t := time.Unix(0, ns).UTC()
	return &t, &id, nil
}

func (s *ProductService) invalidateListCache(ctx context.Context) {
	if s.redis == nil {
		return
	}
	ctx, span := otel.Tracer("product-service/cache").Start(ctx, "cache.product.list.invalidate")
	defer span.End()
	var cursor uint64
	for {
		keys, nextCursor, err := s.redis.Scan(ctx, cursor, "product:list:*", 50).Result()
		if err != nil {
			return
		}
		if len(keys) > 0 {
			_ = s.redis.Del(ctx, keys...).Err()
		}
		cursor = nextCursor
		if cursor == 0 {
			return
		}
	}
}

func (s *ProductService) deleteCachedProduct(ctx context.Context, cacheKey string) {
	if s.redis == nil {
		return
	}
	ctx, span := otel.Tracer("product-service/cache").Start(ctx, "cache.product.detail.delete")
	defer span.End()
	_ = s.redis.Del(ctx, cacheKey).Err()
}
