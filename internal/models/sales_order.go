package models

import (
	"time"

	"github.com/google/uuid"
)

type SalesOrderStatus string

const (
	SOStatusPending           SalesOrderStatus = "pending"
	SOStatusConfirmed         SalesOrderStatus = "confirmed"
	SOStatusProcessing        SalesOrderStatus = "processing"
	SOStatusPartiallyFulfilled SalesOrderStatus = "partially_fulfilled"
	SOStatusFulfilled         SalesOrderStatus = "fulfilled"
	SOStatusCancelled         SalesOrderStatus = "cancelled"
)

type SalesOrder struct {
	ID              uuid.UUID              `db:"id" json:"id"`
	ClientID        uuid.UUID              `db:"client_id" json:"clientId"`
	CustomerID      uuid.UUID              `db:"customer_id" json:"customerId"`
	ShopID          uuid.UUID              `db:"shop_id" json:"shopId"`
	OrderNumber     string                 `db:"order_number" json:"orderNumber"`
	Status          SalesOrderStatus       `db:"status" json:"status"`
	OrderDate       time.Time              `db:"order_date" json:"orderDate"`
	ShippingDate    *time.Time             `db:"shipping_date" json:"shippingDate"`
	DeliveryDate    *time.Time             `db:"delivery_date" json:"deliveryDate"`
	Subtotal        float64                `db:"subtotal" json:"subtotal"`
	TaxAmount       float64                `db:"tax_amount" json:"taxAmount"`
	ShippingAmount  float64                `db:"shipping_amount" json:"shippingAmount"`
	TotalAmount     float64                `db:"total_amount" json:"totalAmount"`
	Currency        string                 `db:"currency" json:"currency"`
	ShippingAddress string                 `db:"shipping_address" json:"shippingAddress"`
	BillingAddress  string                 `db:"billing_address" json:"billingAddress"`
	Notes           string                 `db:"notes" json:"notes"`
	Metadata        map[string]interface{} `db:"metadata" json:"metadata"`
	CreatedAt       time.Time              `db:"created_at" json:"createdAt"`
	UpdatedAt       time.Time              `db:"updated_at" json:"updatedAt"`
}

type SalesOrderItem struct {
	ID                uuid.UUID `db:"id" json:"id"`
	ClientID          uuid.UUID `db:"client_id" json:"clientId"`
	SalesOrderID      uuid.UUID `db:"sales_order_id" json:"salesOrderId"`
	VariantID         uuid.UUID `db:"variant_id" json:"variantId"`
	QuantityOrdered   int       `db:"quantity_ordered" json:"quantityOrdered"`
	QuantityFulfilled int       `db:"quantity_fulfilled" json:"quantityFulfilled"`
	UnitPrice         float64   `db:"unit_price" json:"unitPrice"`
	TaxRate           float64   `db:"tax_rate" json:"taxRate"`
	DiscountAmount    float64   `db:"discount_amount" json:"discountAmount"`
	TotalPrice        float64   `db:"total_price" json:"totalPrice"`
	Notes             string    `db:"notes" json:"notes"`
	CreatedAt         time.Time `db:"created_at" json:"createdAt"`
	UpdatedAt         time.Time `db:"updated_at" json:"updatedAt"`
}
