package models

import (
	"encoding/json"
	"time"
)

// UserRole represents a role that can be assigned to users
type UserRole struct {
	ID          int             `json:"id" db:"id"`
	Name        string          `json:"name" db:"name"`
	Description string          `json:"description" db:"description"`
	Permissions json.RawMessage `json:"permissions" db:"permissions"`
	CreatedAt   time.Time       `json:"created_at" db:"created_at"`
}

// ShopUser represents the relationship between a user and a shop with a role
type ShopUser struct {
	ID        int       `json:"id" db:"id"`
	ClientID  int       `json:"client_id" db:"client_id"`
	ShopID    int       `json:"shop_id" db:"shop_id"`
	UserID    int       `json:"user_id" db:"user_id"`
	RoleID    int       `json:"role_id" db:"role_id"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// AssignUserToShopRequest represents a request to assign a user to a shop
type AssignUserToShopRequest struct {
	ClientID int `json:"client_id" validate:"required,gt=0"`
	ShopID   int `json:"shop_id" validate:"required,gt=0"`
	UserID   int `json:"user_id" validate:"required,gt=0"`
	RoleID   int `json:"role_id" validate:"required,gt=0"`
}
