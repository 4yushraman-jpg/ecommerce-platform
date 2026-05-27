package dto

import "time"

type ProductResponse struct {
	ID             string    `json:"id"`
	Name           string    `json:"name"`
	Description    string    `json:"description"`
	Price          float64   `json:"price"`
	InventoryCount int       `json:"inventory_count"`
	Category       string    `json:"category"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type ListProductsResponse struct {
	Items      []ProductResponse `json:"items"`
	NextCursor string            `json:"next_cursor,omitempty"`
}
