package models

import "time"

// Product represents a product in the system
type Product struct {
	ID               int       `json:"id" db:"id"`
	ClientID         int       `json:"client_id" db:"client_id"`
	ShopID           int       `json:"shop_id" db:"shop_id"`
	CategoryID       *int      `json:"category_id,omitempty" db:"category_id"`
	SKU              *string   `json:"sku,omitempty" db:"sku"`
	Name             string    `json:"name" db:"name"`
	Slug             string    `json:"slug" db:"slug"`
	Description      *string   `json:"description,omitempty" db:"description"`
	ShortDescription *string   `json:"short_description,omitempty" db:"short_description"`
	Price            float64   `json:"price" db:"price"`
	CompareAtPrice   *float64  `json:"compare_at_price,omitempty" db:"compare_at_price"`
	CostPerItem      *float64  `json:"cost_per_item,omitempty" db:"cost_per_item"`
	Weight           *float64  `json:"weight,omitempty" db:"weight"`
	WeightUnit       string    `json:"weight_unit" db:"weight_unit"`
	RequiresShipping bool      `json:"requires_shipping" db:"requires_shipping"`
	IsActive         bool      `json:"is_active" db:"is_active"`
	CreatedAt        time.Time `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time `json:"updated_at" db:"updated_at"`
}

// CreateProductRequest represents a request to create a new product
type CreateProductRequest struct {
	ClientID         int      `json:"client_id" validate:"required,gt=0"`
	ShopID           int      `json:"shop_id" validate:"required,gt=0"`
	CategoryID       *int     `json:"category_id,omitempty"`
	SKU              *string  `json:"sku,omitempty"`
	Name             string   `json:"name" validate:"required,min=2,max=255"`
	Slug             string   `json:"slug" validate:"required,min=2,max=255"`
	Description      *string  `json:"description,omitempty"`
	ShortDescription *string  `json:"short_description,omitempty" validate:"omitempty,max=500"`
	Price            float64  `json:"price" validate:"required,gte=0"`
	CompareAtPrice   *float64 `json:"compare_at_price,omitempty" validate:"omitempty,gte=0"`
	CostPerItem      *float64 `json:"cost_per_item,omitempty" validate:"omitempty,gte=0"`
	Weight           *float64 `json:"weight,omitempty" validate:"omitempty,gte=0"`
	WeightUnit       string   `json:"weight_unit" validate:"required,oneof=kg lb oz g"`
	RequiresShipping bool     `json:"requires_shipping"`
}

// UpdateProductRequest represents a request to update a product
type UpdateProductRequest struct {
	CategoryID       *int     `json:"category_id,omitempty"`
	SKU              *string  `json:"sku,omitempty"`
	Name             *string  `json:"name,omitempty" validate:"omitempty,min=2,max=255"`
	Description      *string  `json:"description,omitempty"`
	ShortDescription *string  `json:"short_description,omitempty" validate:"omitempty,max=500"`
	Price            *float64 `json:"price,omitempty" validate:"omitempty,gte=0"`
	CompareAtPrice   *float64 `json:"compare_at_price,omitempty" validate:"omitempty,gte=0"`
	CostPerItem      *float64 `json:"cost_per_item,omitempty" validate:"omitempty,gte=0"`
	Weight           *float64 `json:"weight,omitempty" validate:"omitempty,gte=0"`
	WeightUnit       *string  `json:"weight_unit,omitempty" validate:"omitempty,oneof=kg lb oz g"`
	RequiresShipping *bool    `json:"requires_shipping,omitempty"`
	IsActive         *bool    `json:"is_active,omitempty"`
}
