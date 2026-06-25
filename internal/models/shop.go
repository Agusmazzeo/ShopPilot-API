package models

import (
	"time"

	"github.com/google/uuid"
)

type Shop struct {
	ID          uuid.UUID `db:"id" json:"id"`
	ClientID    uuid.UUID `db:"client_id" json:"clientId"`
	Name        string    `db:"name" json:"name"`
	Slug        string    `db:"slug" json:"slug"`
	Description string    `db:"description" json:"description"`
	WebpageURL  string    `db:"webpage_url" json:"webpageUrl"`
	Address     string    `db:"address" json:"address"`
	City        string    `db:"city" json:"city"`
	State       string    `db:"state" json:"state"`
	Country     string    `db:"country" json:"country"`
	PostalCode  string    `db:"postal_code" json:"postalCode"`
	Phone       string    `db:"phone" json:"phone"`
	Email       string    `db:"email" json:"email"`
	LogoURL     *string   `db:"logo_url" json:"logoUrl"`
	IsActive    bool      `db:"is_active" json:"isActive"`
	CreatedAt   time.Time `db:"created_at" json:"createdAt"`
	UpdatedAt   time.Time `db:"updated_at" json:"updatedAt"`
}

type ShopUser struct {
	ID           int       `db:"id" json:"id"`
	ClientID     uuid.UUID `db:"client_id" json:"clientId"`
	ShopID       uuid.UUID `db:"shop_id" json:"shopId"`
	ClientUserID uuid.UUID `db:"client_user_id" json:"clientUserId"`
	Role         string    `db:"role" json:"role"`
	CreatedAt    time.Time `db:"created_at" json:"createdAt"`
	UpdatedAt    time.Time `db:"updated_at" json:"updatedAt"`
}
