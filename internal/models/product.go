package models

import (
	"time"

	"github.com/google/uuid"
)

type Product struct {
	ID          uuid.UUID              `db:"id" json:"id"`
	ShopID      uuid.UUID              `db:"shop_id" json:"shopId"`
	ClientID    uuid.UUID              `db:"client_id" json:"clientId"`
	Code        string                 `db:"code" json:"code"`
	Name        string                 `db:"name" json:"name"`
	Description string                 `db:"description" json:"description"`
	Metadata    map[string]interface{} `db:"metadata" json:"metadata"`
	IsActive    bool                   `db:"is_active" json:"isActive"`
	CreatedAt   time.Time              `db:"created_at" json:"createdAt"`
	UpdatedAt   time.Time              `db:"updated_at" json:"updatedAt"`
}

type ProductVariant struct {
	ID               uuid.UUID              `db:"id" json:"id"`
	ShopID           uuid.UUID              `db:"shop_id" json:"shopId"`
	ProductID        uuid.UUID              `db:"product_id" json:"productId"`
	SKU              string                 `db:"sku" json:"sku"`
	Name             string                 `db:"name" json:"name"`
	Price            float64                `db:"price" json:"price"`
	CompareAtPrice   *float64               `db:"compare_at_price" json:"compareAtPrice"`
	Cost             *float64               `db:"cost" json:"cost"`
	Quantity         int                    `db:"quantity" json:"quantity"`
	Weight           *float64               `db:"weight" json:"weight"`
	WeightUnit       string                 `db:"weight_unit" json:"weightUnit"`
	RequiresShipping bool                   `db:"requires_shipping" json:"requiresShipping"`
	IsDefault        bool                   `db:"is_default" json:"isDefault"`
	Attributes       map[string]interface{} `db:"attributes" json:"attributes"`
	IsActive         bool                   `db:"is_active" json:"isActive"`
	CreatedAt        time.Time              `db:"created_at" json:"createdAt"`
	UpdatedAt        time.Time              `db:"updated_at" json:"updatedAt"`
}
