package dto

type CreateProductRequest struct {
	Name           string  `json:"name" validate:"required,min=2,max=120"`
	Description    string  `json:"description" validate:"max=1000"`
	Price          float64 `json:"price" validate:"required,gt=0"`
	InventoryCount int     `json:"inventory_count" validate:"gte=0"`
	Category       string  `json:"category" validate:"required,min=2,max=80"`
}

type UpdateProductRequest struct {
	Name           *string  `json:"name" validate:"omitempty,min=2,max=120"`
	Description    *string  `json:"description" validate:"omitempty,max=1000"`
	Price          *float64 `json:"price" validate:"omitempty,gt=0"`
	InventoryCount *int     `json:"inventory_count" validate:"omitempty,gte=0"`
	Category       *string  `json:"category" validate:"omitempty,min=2,max=80"`
}

type ListProductsQuery struct {
	Limit    int
	Cursor   string
	Category string
	Sort     string
	Search   string
}
