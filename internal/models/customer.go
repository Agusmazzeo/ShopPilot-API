package models

import (
	"time"

	"github.com/google/uuid"
)

type Customer struct {
	ID                  uuid.UUID              `db:"id" json:"id"`
	ClientID            uuid.UUID              `db:"client_id" json:"clientId"`
	Code                string                 `db:"code" json:"code"`
	FirstName           string                 `db:"first_name" json:"firstName"`
	LastName            string                 `db:"last_name" json:"lastName"`
	Email               string                 `db:"email" json:"email"`
	Phone               string                 `db:"phone" json:"phone"`
	ShippingAddress     string                 `db:"shipping_address" json:"shippingAddress"`
	ShippingCity        string                 `db:"shipping_city" json:"shippingCity"`
	ShippingState       string                 `db:"shipping_state" json:"shippingState"`
	ShippingPostalCode  string                 `db:"shipping_postal_code" json:"shippingPostalCode"`
	ShippingCountry     string                 `db:"shipping_country" json:"shippingCountry"`
	BillingAddress      string                 `db:"billing_address" json:"billingAddress"`
	BillingCity         string                 `db:"billing_city" json:"billingCity"`
	BillingState        string                 `db:"billing_state" json:"billingState"`
	BillingPostalCode   string                 `db:"billing_postal_code" json:"billingPostalCode"`
	BillingCountry      string                 `db:"billing_country" json:"billingCountry"`
	TaxID               string                 `db:"tax_id" json:"taxId"`
	Notes               string                 `db:"notes" json:"notes"`
	Metadata            map[string]interface{} `db:"metadata" json:"metadata"`
	IsActive            bool                   `db:"is_active" json:"isActive"`
	CreatedAt           time.Time              `db:"created_at" json:"createdAt"`
	UpdatedAt           time.Time              `db:"updated_at" json:"updatedAt"`
}
