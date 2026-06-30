package models

import (
	"time"

	"github.com/google/uuid"
)

type PurchaseOrderStatus string

const (
	POStatusDraft              PurchaseOrderStatus = "draft"
	POStatusSubmitted          PurchaseOrderStatus = "submitted"
	POStatusPartiallyReceived  PurchaseOrderStatus = "partially_received"
	POStatusReceived           PurchaseOrderStatus = "received"
	POStatusCancelled          PurchaseOrderStatus = "cancelled"
)

type PurchaseOrder struct {
	ID                   uuid.UUID              `db:"id" json:"id"`
	ClientID             uuid.UUID              `db:"client_id" json:"clientId"`
	SupplierID           uuid.UUID              `db:"supplier_id" json:"supplierId"`
	ShopID               uuid.UUID              `db:"shop_id" json:"shopId"`
	PONumber             string                 `db:"po_number" json:"poNumber"`
	Status               PurchaseOrderStatus    `db:"status" json:"status"`
	OrderDate            time.Time              `db:"order_date" json:"orderDate"`
	ExpectedDeliveryDate *time.Time             `db:"expected_delivery_date" json:"expectedDeliveryDate"`
	ReceivedDate         *time.Time             `db:"received_date" json:"receivedDate"`
	TotalAmount          float64                `db:"total_amount" json:"totalAmount"`
	Currency             string                 `db:"currency" json:"currency"`
	Notes                string                 `db:"notes" json:"notes"`
	Metadata             map[string]interface{} `db:"metadata" json:"metadata"`
	CreatedAt            time.Time              `db:"created_at" json:"createdAt"`
	UpdatedAt            time.Time              `db:"updated_at" json:"updatedAt"`
}

type PurchaseOrderItem struct {
	ID               uuid.UUID `db:"id" json:"id"`
	ClientID         uuid.UUID `db:"client_id" json:"clientId"`
	PurchaseOrderID  uuid.UUID `db:"purchase_order_id" json:"purchaseOrderId"`
	VariantID        uuid.UUID `db:"variant_id" json:"variantId"`
	QuantityOrdered  int       `db:"quantity_ordered" json:"quantityOrdered"`
	QuantityReceived int       `db:"quantity_received" json:"quantityReceived"`
	UnitCost         float64   `db:"unit_cost" json:"unitCost"`
	TotalCost        float64   `db:"total_cost" json:"totalCost"`
	Notes            string    `db:"notes" json:"notes"`
	CreatedAt        time.Time `db:"created_at" json:"createdAt"`
	UpdatedAt        time.Time `db:"updated_at" json:"updatedAt"`
}
