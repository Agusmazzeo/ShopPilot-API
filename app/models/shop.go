package models

import "time"

// Shop represents a shop/store in the system
type Shop struct {
	ID           int       `json:"id" db:"id"`
	ClientID     int       `json:"client_id" db:"client_id"`
	UserID       int       `json:"user_id" db:"user_id"`
	Name         string    `json:"name" db:"name"`
	Slug         string    `json:"slug" db:"slug"`
	Description  *string   `json:"description,omitempty" db:"description"`
	LogoURL      *string   `json:"logo_url,omitempty" db:"logo_url"`
	Theme        string    `json:"theme" db:"theme"`
	CustomDomain *string   `json:"custom_domain,omitempty" db:"custom_domain"`
	IsActive     bool      `json:"is_active" db:"is_active"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}

// CreateShopRequest represents a request to create a new shop
type CreateShopRequest struct {
	ClientID     int     `json:"client_id" validate:"required,gt=0"`
	UserID       int     `json:"user_id" validate:"required,gt=0"`
	Name         string  `json:"name" validate:"required,min=2,max=255"`
	Slug         string  `json:"slug" validate:"required,min=2,max=255,alphanum"`
	Description  *string `json:"description,omitempty"`
	LogoURL      *string `json:"logo_url,omitempty"`
	Theme        string  `json:"theme" validate:"required"`
	CustomDomain *string `json:"custom_domain,omitempty"`
}

// UpdateShopRequest represents a request to update a shop
type UpdateShopRequest struct {
	Name         *string `json:"name,omitempty" validate:"omitempty,min=2,max=255"`
	Description  *string `json:"description,omitempty"`
	LogoURL      *string `json:"logo_url,omitempty"`
	Theme        *string `json:"theme,omitempty"`
	CustomDomain *string `json:"custom_domain,omitempty"`
	IsActive     *bool   `json:"is_active,omitempty"`
}
