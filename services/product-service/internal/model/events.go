package model

import "time"

type ProductEvent struct {
	Type      string    `json:"type"`
	ProductID string    `json:"product_id"`
	At        time.Time `json:"at"`
}

const (
	ProductCreatedEventType   = "product.created"
	ProductUpdatedEventType   = "product.updated"
	InventoryUpdatedEventType = "inventory.updated"
)
