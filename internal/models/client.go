package models

import (
	"time"

	"github.com/google/uuid"
)

type Client struct {
	ID           uuid.UUID `db:"id" json:"id"`
	Name         string    `db:"name" json:"name"`
	Slug         string    `db:"slug" json:"slug"`
	Description  string    `db:"description" json:"description"`
	ContactEmail string    `db:"contact_email" json:"contactEmail"`
	ContactPhone string    `db:"contact_phone" json:"contactPhone"`
	WebsiteURL   string    `db:"website_url" json:"websiteUrl"`
	LogoURL      *string   `db:"logo_url" json:"logoUrl"`
	IsActive     bool      `db:"is_active" json:"isActive"`
	CreatedAt    time.Time `db:"created_at" json:"createdAt"`
	UpdatedAt    time.Time `db:"updated_at" json:"updatedAt"`
}
