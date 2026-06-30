package models

import (
	"time"

	"github.com/google/uuid"
)

type Supplier struct {
	ID           uuid.UUID              `db:"id" json:"id"`
	ClientID     uuid.UUID              `db:"client_id" json:"clientId"`
	Code         string                 `db:"code" json:"code"`
	Name         string                 `db:"name" json:"name"`
	Email        string                 `db:"email" json:"email"`
	Phone        string                 `db:"phone" json:"phone"`
	Address      string                 `db:"address" json:"address"`
	City         string                 `db:"city" json:"city"`
	State        string                 `db:"state" json:"state"`
	PostalCode   string                 `db:"postal_code" json:"postalCode"`
	Country      string                 `db:"country" json:"country"`
	TaxID        string                 `db:"tax_id" json:"taxId"`
	PaymentTerms string                 `db:"payment_terms" json:"paymentTerms"`
	Currency     string                 `db:"currency" json:"currency"`
	Notes        string                 `db:"notes" json:"notes"`
	Metadata     map[string]interface{} `db:"metadata" json:"metadata"`
	IsActive     bool                   `db:"is_active" json:"isActive"`
	CreatedAt    time.Time              `db:"created_at" json:"createdAt"`
	UpdatedAt    time.Time              `db:"updated_at" json:"updatedAt"`
}
