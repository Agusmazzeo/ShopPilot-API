package models

import "time"

// Client represents a tenant in the multi-tenant system
type Client struct {
	ID               int       `json:"id" db:"id"`
	Name             string    `json:"name" db:"name"`
	Slug             string    `json:"slug" db:"slug"`
	ContactEmail     string    `json:"contact_email" db:"contact_email"`
	ContactPhone     *string   `json:"contact_phone,omitempty" db:"contact_phone"`
	SubscriptionTier string    `json:"subscription_tier" db:"subscription_tier"`
	IsActive         bool      `json:"is_active" db:"is_active"`
	CreatedAt        time.Time `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time `json:"updated_at" db:"updated_at"`
}

// CreateClientRequest represents a request to create a new client
type CreateClientRequest struct {
	Name             string  `json:"name" validate:"required,min=2,max=255"`
	Slug             string  `json:"slug" validate:"required,min=2,max=255,alphanum"`
	ContactEmail     string  `json:"contact_email" validate:"required,email"`
	ContactPhone     *string `json:"contact_phone,omitempty"`
	SubscriptionTier string  `json:"subscription_tier" validate:"required,oneof=free basic pro enterprise"`
}

// UpdateClientRequest represents a request to update a client
type UpdateClientRequest struct {
	Name             *string `json:"name,omitempty" validate:"omitempty,min=2,max=255"`
	ContactEmail     *string `json:"contact_email,omitempty" validate:"omitempty,email"`
	ContactPhone     *string `json:"contact_phone,omitempty"`
	SubscriptionTier *string `json:"subscription_tier,omitempty" validate:"omitempty,oneof=free basic pro enterprise"`
	IsActive         *bool   `json:"is_active,omitempty"`
}
