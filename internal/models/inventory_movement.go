package models

import (
	"time"

	"github.com/google/uuid"
)

type InventoryMovementType string

const (
	MovementTypePurchase          InventoryMovementType = "purchase"
	MovementTypeSale              InventoryMovementType = "sale"
	MovementTypeAdjustment        InventoryMovementType = "adjustment"
	MovementTypeReturnFromCustomer InventoryMovementType = "return_from_customer"
	MovementTypeReturnToSupplier  InventoryMovementType = "return_to_supplier"
	MovementTypeDamaged           InventoryMovementType = "damaged"
	MovementTypeTransfer          InventoryMovementType = "transfer"
)

type InventoryMovement struct {
	ID               uuid.UUID             `db:"id" json:"id"`
	ClientID         uuid.UUID             `db:"client_id" json:"clientId"`
	VariantID        uuid.UUID             `db:"variant_id" json:"variantId"`
	ShopID           uuid.UUID             `db:"shop_id" json:"shopId"`
	MovementType     InventoryMovementType `db:"movement_type" json:"movementType"`
	Quantity         int                   `db:"quantity" json:"quantity"`
	PreviousQuantity int                   `db:"previous_quantity" json:"previousQuantity"`
	NewQuantity      int                   `db:"new_quantity" json:"newQuantity"`
	ReferenceType    string                `db:"reference_type" json:"referenceType"`
	ReferenceID      *uuid.UUID            `db:"reference_id" json:"referenceId"`
	Notes            string                `db:"notes" json:"notes"`
	PerformedBy      *uuid.UUID            `db:"performed_by" json:"performedBy"`
	CreatedAt        time.Time             `db:"created_at" json:"createdAt"`
}
