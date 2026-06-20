package models

import "time"

// ProductCategory represents a product category
type ProductCategory struct {
	ID           int       `json:"id" db:"id"`
	ClientID     int       `json:"client_id" db:"client_id"`
	ShopID       int       `json:"shop_id" db:"shop_id"`
	Name         string    `json:"name" db:"name"`
	Slug         string    `json:"slug" db:"slug"`
	Description  *string   `json:"description,omitempty" db:"description"`
	ParentID     *int      `json:"parent_id,omitempty" db:"parent_id"`
	DisplayOrder int       `json:"display_order" db:"display_order"`
	IsActive     bool      `json:"is_active" db:"is_active"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}

// CreateProductCategoryRequest represents a request to create a new category
type CreateProductCategoryRequest struct {
	ClientID     int     `json:"client_id" validate:"required,gt=0"`
	ShopID       int     `json:"shop_id" validate:"required,gt=0"`
	Name         string  `json:"name" validate:"required,min=2,max=255"`
	Slug         string  `json:"slug" validate:"required,min=2,max=255,alphanum"`
	Description  *string `json:"description,omitempty"`
	ParentID     *int    `json:"parent_id,omitempty"`
	DisplayOrder int     `json:"display_order" validate:"gte=0"`
}

// UpdateProductCategoryRequest represents a request to update a category
type UpdateProductCategoryRequest struct {
	Name         *string `json:"name,omitempty" validate:"omitempty,min=2,max=255"`
	Description  *string `json:"description,omitempty"`
	ParentID     *int    `json:"parent_id,omitempty"`
	DisplayOrder *int    `json:"display_order,omitempty" validate:"omitempty,gte=0"`
	IsActive     *bool   `json:"is_active,omitempty"`
}
