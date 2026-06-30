package models

import (
	"time"

	"github.com/google/uuid"
)

type InventoryAlert struct {
	ID                uuid.UUID              `db:"id" json:"id"`
	ClientID          uuid.UUID              `db:"client_id" json:"clientId"`
	VariantID         uuid.UUID              `db:"variant_id" json:"variantId"`
	ShopID            uuid.UUID              `db:"shop_id" json:"shopId"`
	ReorderPoint      int                    `db:"reorder_point" json:"reorderPoint"`
	ReorderQuantity   int                    `db:"reorder_quantity" json:"reorderQuantity"`
	LowStockThreshold int                    `db:"low_stock_threshold" json:"lowStockThreshold"`
	IsEnabled         bool                   `db:"is_enabled" json:"isEnabled"`
	LastAlertSentAt   *time.Time             `db:"last_alert_sent_at" json:"lastAlertSentAt"`
	Metadata          map[string]interface{} `db:"metadata" json:"metadata"`
	CreatedAt         time.Time              `db:"created_at" json:"createdAt"`
	UpdatedAt         time.Time              `db:"updated_at" json:"updatedAt"`
}
