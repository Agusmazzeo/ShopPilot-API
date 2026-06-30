package services

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/yourorg/shoppilot/app/repositories"
	"github.com/yourorg/shoppilot/internal/models"
)

// PurchaseOrderService defines the interface for purchase order business logic
type PurchaseOrderService interface {
	// Purchase Order management
	CreatePurchaseOrder(ctx context.Context, clientID uuid.UUID, req *CreatePurchaseOrderRequest) (*models.PurchaseOrder, error)
	GetPurchaseOrder(ctx context.Context, clientID uuid.UUID, id uuid.UUID) (*models.PurchaseOrder, error)
	UpdatePurchaseOrder(ctx context.Context, clientID uuid.UUID, id uuid.UUID, req *UpdatePurchaseOrderRequest) error
	DeletePurchaseOrder(ctx context.Context, clientID uuid.UUID, id uuid.UUID) error
	ListPurchaseOrders(ctx context.Context, clientID uuid.UUID, filters *PurchaseOrderFilters, page, pageSize int) ([]*models.PurchaseOrder, int, error)

	// Status transitions
	SubmitPurchaseOrder(ctx context.Context, clientID uuid.UUID, id uuid.UUID) error
	CancelPurchaseOrder(ctx context.Context, clientID uuid.UUID, id uuid.UUID) error

	// Line items
	AddItem(ctx context.Context, clientID uuid.UUID, poID uuid.UUID, req *AddPurchaseOrderItemRequest) (*models.PurchaseOrderItem, error)
	RemoveItem(ctx context.Context, clientID uuid.UUID, poID uuid.UUID, itemID uuid.UUID) error
	ListItems(ctx context.Context, clientID uuid.UUID, poID uuid.UUID) ([]*models.PurchaseOrderItem, error)

	// Receiving
	ReceiveItems(ctx context.Context, clientID uuid.UUID, poID uuid.UUID, items []ReceiveItemRequest) error
}

// purchaseOrderService implements PurchaseOrderService interface
type purchaseOrderService struct {
	poRepo       repositories.PurchaseOrderRepository
	productRepo  repositories.ProductRepository
	movementRepo repositories.InventoryMovementRepository
}

// NewPurchaseOrderService creates a new purchase order service
func NewPurchaseOrderService(
	poRepo repositories.PurchaseOrderRepository,
	productRepo repositories.ProductRepository,
	movementRepo repositories.InventoryMovementRepository,
) PurchaseOrderService {
	return &purchaseOrderService{
		poRepo:       poRepo,
		productRepo:  productRepo,
		movementRepo: movementRepo,
	}
}

// Request/Response types

type CreatePurchaseOrderRequest struct {
	SupplierID    uuid.UUID
	ShopID        uuid.UUID
	PONumber      string // Optional - auto-generated if empty
	OrderDate     time.Time
	ExpectedDate  *time.Time
	Notes         string
	TotalAmount   float64
	Metadata      map[string]interface{}
	Items         []CreatePurchaseOrderItemRequest
}

type CreatePurchaseOrderItemRequest struct {
	VariantID       uuid.UUID
	QuantityOrdered int
	UnitPrice       float64
	TotalPrice      float64
	Notes           string
}

type UpdatePurchaseOrderRequest struct {
	ExpectedDate *time.Time
	Notes        *string
	Metadata     map[string]interface{}
}

type PurchaseOrderFilters struct {
	SupplierID *uuid.UUID
	ShopID     *uuid.UUID
	Status     *models.PurchaseOrderStatus
}

type AddPurchaseOrderItemRequest struct {
	VariantID       uuid.UUID
	QuantityOrdered int
	UnitPrice       float64
	Notes           string
}

type ReceiveItemRequest struct {
	ItemID           uuid.UUID
	QuantityReceived int
}

// CreatePurchaseOrder creates a new purchase order with business rule validations
func (s *purchaseOrderService) CreatePurchaseOrder(ctx context.Context, clientID uuid.UUID, req *CreatePurchaseOrderRequest) (*models.PurchaseOrder, error) {
	// Business rule: Auto-generate PO number if not provided
	poNumber := req.PONumber
	if poNumber == "" {
		// Simple auto-generation: PO-{timestamp}
		poNumber = fmt.Sprintf("PO-%d", time.Now().Unix())
	}

	// Business rule: PO number unique per client
	existingPO, err := s.poRepo.GetByPONumber(ctx, clientID, poNumber)
	if err == nil && existingPO != nil {
		return nil, fmt.Errorf("purchase order number '%s' already exists for this client", poNumber)
	}

	po := &models.PurchaseOrder{
		ClientID:             clientID,
		SupplierID:           req.SupplierID,
		ShopID:               req.ShopID,
		PONumber:             poNumber,
		OrderDate:            req.OrderDate,
		ExpectedDeliveryDate: req.ExpectedDate,
		Status:               models.POStatusDraft,
		Notes:                req.Notes,
		TotalAmount:          req.TotalAmount,
		Currency:             "USD", // Default currency
		Metadata:             req.Metadata,
	}

	if err := s.poRepo.Create(ctx, po); err != nil {
		return nil, fmt.Errorf("failed to create purchase order: %w", err)
	}

	// Create line items if provided
	for _, itemReq := range req.Items {
		item := &models.PurchaseOrderItem{
			ClientID:         clientID,
			PurchaseOrderID:  po.ID,
			VariantID:        itemReq.VariantID,
			QuantityOrdered:  itemReq.QuantityOrdered,
			QuantityReceived: 0,
			UnitCost:         itemReq.UnitPrice,
			TotalCost:        itemReq.TotalPrice,
			Notes:            itemReq.Notes,
		}

		if err := s.poRepo.CreateItem(ctx, item); err != nil {
			return nil, fmt.Errorf("failed to create purchase order item: %w", err)
		}
	}

	return po, nil
}

// GetPurchaseOrder retrieves a purchase order by ID
func (s *purchaseOrderService) GetPurchaseOrder(ctx context.Context, clientID uuid.UUID, id uuid.UUID) (*models.PurchaseOrder, error) {
	po, err := s.poRepo.GetByID(ctx, clientID, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get purchase order: %w", err)
	}

	return po, nil
}

// UpdatePurchaseOrder updates an existing purchase order
func (s *purchaseOrderService) UpdatePurchaseOrder(ctx context.Context, clientID uuid.UUID, id uuid.UUID, req *UpdatePurchaseOrderRequest) error {
	// Get existing purchase order
	po, err := s.poRepo.GetByID(ctx, clientID, id)
	if err != nil {
		return fmt.Errorf("failed to get purchase order: %w", err)
	}

	// Business rule: Only draft POs can be modified
	if po.Status != models.POStatusDraft {
		return fmt.Errorf("only draft purchase orders can be modified")
	}

	// Apply updates
	if req.ExpectedDate != nil {
		po.ExpectedDeliveryDate = req.ExpectedDate
	}
	if req.Notes != nil {
		po.Notes = *req.Notes
	}
	if req.Metadata != nil {
		po.Metadata = req.Metadata
	}

	if err := s.poRepo.Update(ctx, po); err != nil {
		return fmt.Errorf("failed to update purchase order: %w", err)
	}

	return nil
}

// DeletePurchaseOrder deletes a purchase order
func (s *purchaseOrderService) DeletePurchaseOrder(ctx context.Context, clientID uuid.UUID, id uuid.UUID) error {
	// Get existing purchase order
	po, err := s.poRepo.GetByID(ctx, clientID, id)
	if err != nil {
		return fmt.Errorf("failed to get purchase order: %w", err)
	}

	// Business rule: Only draft POs can be deleted
	if po.Status != models.POStatusDraft {
		return fmt.Errorf("only draft purchase orders can be deleted")
	}

	if err := s.poRepo.Delete(ctx, clientID, id); err != nil {
		return fmt.Errorf("failed to delete purchase order: %w", err)
	}

	return nil
}

// ListPurchaseOrders retrieves purchase orders with filtering and pagination
func (s *purchaseOrderService) ListPurchaseOrders(ctx context.Context, clientID uuid.UUID, filters *PurchaseOrderFilters, page, pageSize int) ([]*models.PurchaseOrder, int, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	offset := (page - 1) * pageSize

	var pos []*models.PurchaseOrder
	var err error

	if filters != nil {
		if filters.SupplierID != nil {
			pos, err = s.poRepo.ListBySupplier(ctx, clientID, *filters.SupplierID, pageSize, offset)
		} else if filters.Status != nil {
			pos, err = s.poRepo.ListByStatus(ctx, clientID, *filters.Status, pageSize, offset)
		} else {
			pos, err = s.poRepo.ListByClient(ctx, clientID, pageSize, offset)
		}
	} else {
		pos, err = s.poRepo.ListByClient(ctx, clientID, pageSize, offset)
	}

	if err != nil {
		return nil, 0, fmt.Errorf("failed to list purchase orders: %w", err)
	}

	total := len(pos)

	return pos, total, nil
}

// SubmitPurchaseOrder transitions a purchase order from draft to submitted
func (s *purchaseOrderService) SubmitPurchaseOrder(ctx context.Context, clientID uuid.UUID, id uuid.UUID) error {
	po, err := s.poRepo.GetByID(ctx, clientID, id)
	if err != nil {
		return fmt.Errorf("failed to get purchase order: %w", err)
	}

	if po.Status != models.POStatusDraft {
		return fmt.Errorf("only draft purchase orders can be submitted")
	}

	po.Status = models.POStatusSubmitted
	if err := s.poRepo.Update(ctx, po); err != nil {
		return fmt.Errorf("failed to submit purchase order: %w", err)
	}

	return nil
}

// CancelPurchaseOrder transitions a purchase order to cancelled
func (s *purchaseOrderService) CancelPurchaseOrder(ctx context.Context, clientID uuid.UUID, id uuid.UUID) error {
	po, err := s.poRepo.GetByID(ctx, clientID, id)
	if err != nil {
		return fmt.Errorf("failed to get purchase order: %w", err)
	}

	if po.Status == models.POStatusReceived || po.Status == models.POStatusCancelled {
		return fmt.Errorf("cannot cancel a received or already cancelled purchase order")
	}

	po.Status = models.POStatusCancelled
	if err := s.poRepo.Update(ctx, po); err != nil {
		return fmt.Errorf("failed to cancel purchase order: %w", err)
	}

	return nil
}

// AddItem adds a line item to a purchase order
func (s *purchaseOrderService) AddItem(ctx context.Context, clientID uuid.UUID, poID uuid.UUID, req *AddPurchaseOrderItemRequest) (*models.PurchaseOrderItem, error) {
	po, err := s.poRepo.GetByID(ctx, clientID, poID)
	if err != nil {
		return nil, fmt.Errorf("failed to get purchase order: %w", err)
	}

	// Business rule: Only draft POs can be modified
	if po.Status != models.POStatusDraft {
		return nil, fmt.Errorf("can only add items to draft purchase orders")
	}

	item := &models.PurchaseOrderItem{
		ClientID:         clientID,
		PurchaseOrderID:  poID,
		VariantID:        req.VariantID,
		QuantityOrdered:  req.QuantityOrdered,
		QuantityReceived: 0,
		UnitCost:         req.UnitPrice,
		TotalCost:        req.UnitPrice * float64(req.QuantityOrdered),
		Notes:            req.Notes,
	}

	if err := s.poRepo.CreateItem(ctx, item); err != nil {
		return nil, fmt.Errorf("failed to add item to purchase order: %w", err)
	}

	return item, nil
}

// RemoveItem removes a line item from a purchase order
func (s *purchaseOrderService) RemoveItem(ctx context.Context, clientID uuid.UUID, poID uuid.UUID, itemID uuid.UUID) error {
	po, err := s.poRepo.GetByID(ctx, clientID, poID)
	if err != nil {
		return fmt.Errorf("failed to get purchase order: %w", err)
	}

	// Business rule: Only draft POs can be modified
	if po.Status != models.POStatusDraft {
		return fmt.Errorf("can only remove items from draft purchase orders")
	}

	if err := s.poRepo.DeleteItem(ctx, clientID, itemID); err != nil {
		return fmt.Errorf("failed to remove item from purchase order: %w", err)
	}

	return nil
}

// ListItems retrieves all line items for a purchase order
func (s *purchaseOrderService) ListItems(ctx context.Context, clientID uuid.UUID, poID uuid.UUID) ([]*models.PurchaseOrderItem, error) {
	items, err := s.poRepo.ListItemsByPO(ctx, clientID, poID)
	if err != nil {
		return nil, fmt.Errorf("failed to list purchase order items: %w", err)
	}

	return items, nil
}

// ReceiveItems processes received items for a purchase order
// This is a multi-step transaction that:
// 1. Updates purchase_order_items.quantity_received
// 2. Updates product_variants.quantity (increase)
// 3. Creates inventory_movements records
// 4. Updates purchase_orders.status if fully received
func (s *purchaseOrderService) ReceiveItems(ctx context.Context, clientID uuid.UUID, poID uuid.UUID, items []ReceiveItemRequest) error {
	// Note: In a production system, this should use database transactions
	// For now, we'll process sequentially and rely on individual operations

	po, err := s.poRepo.GetByID(ctx, clientID, poID)
	if err != nil {
		return fmt.Errorf("failed to get purchase order: %w", err)
	}

	if po.Status != models.POStatusSubmitted && po.Status != models.POStatusPartiallyReceived {
		return fmt.Errorf("can only receive items for submitted or partially received purchase orders")
	}

	// Process each item
	for _, receiveReq := range items {
		// Get the item
		item, err := s.poRepo.GetItem(ctx, clientID, receiveReq.ItemID)
		if err != nil {
			return fmt.Errorf("failed to get purchase order item: %w", err)
		}

		// Verify item belongs to this PO
		if item.PurchaseOrderID != poID {
			return fmt.Errorf("item does not belong to this purchase order")
		}

		// Get the variant to update inventory
		variant, err := s.productRepo.GetVariantByID(ctx, clientID, item.VariantID)
		if err != nil {
			return fmt.Errorf("failed to get variant: %w", err)
		}

		// 1. Update quantity_received on the item
		if err := s.poRepo.ReceiveItem(ctx, clientID, receiveReq.ItemID, receiveReq.QuantityReceived); err != nil {
			return fmt.Errorf("failed to update received quantity: %w", err)
		}

		// 2. Update variant quantity (increase)
		previousQuantity := variant.Quantity
		newQuantity := previousQuantity + receiveReq.QuantityReceived
		if err := s.productRepo.UpdateInventory(ctx, clientID, item.VariantID, newQuantity); err != nil {
			return fmt.Errorf("failed to update inventory: %w", err)
		}

		// 3. Create inventory movement record
		refID := poID
		movement := &models.InventoryMovement{
			ClientID:         clientID,
			VariantID:        item.VariantID,
			ShopID:           po.ShopID,
			MovementType:     models.MovementTypePurchase,
			Quantity:         receiveReq.QuantityReceived,
			PreviousQuantity: previousQuantity,
			NewQuantity:      newQuantity,
			ReferenceType:    "purchase_order",
			ReferenceID:      &refID,
			Notes:            fmt.Sprintf("Received from PO %s", po.PONumber),
		}

		if err := s.movementRepo.Create(ctx, movement); err != nil {
			return fmt.Errorf("failed to create inventory movement: %w", err)
		}
	}

	// 4. Check if all items are fully received and update PO status
	allItems, err := s.poRepo.ListItemsByPO(ctx, clientID, poID)
	if err != nil {
		return fmt.Errorf("failed to list purchase order items: %w", err)
	}

	fullyReceived := true
	partiallyReceived := false
	for _, item := range allItems {
		if item.QuantityReceived < item.QuantityOrdered {
			fullyReceived = false
		}
		if item.QuantityReceived > 0 {
			partiallyReceived = true
		}
	}

	if fullyReceived {
		po.Status = models.POStatusReceived
	} else if partiallyReceived {
		po.Status = models.POStatusPartiallyReceived
	}

	if err := s.poRepo.Update(ctx, po); err != nil {
		return fmt.Errorf("failed to update purchase order status: %w", err)
	}

	return nil
}
