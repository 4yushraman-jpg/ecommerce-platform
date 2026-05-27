package mapper

import (
	"github.com/4yushraman-jpg/product-service/internal/dto"
	"github.com/4yushraman-jpg/product-service/internal/model"
)

func ToProductResponse(product *model.Product) dto.ProductResponse {
	return dto.ProductResponse{
		ID:             product.ID.String(),
		Name:           product.Name,
		Description:    product.Description,
		Price:          product.Price,
		InventoryCount: product.InventoryCount,
		Category:       product.Category,
		CreatedAt:      product.CreatedAt,
		UpdatedAt:      product.UpdatedAt,
	}
}
