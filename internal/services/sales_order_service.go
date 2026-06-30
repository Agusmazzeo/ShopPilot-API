package services

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/yourorg/shoppilot/app/repositories"
	"github.com/yourorg/shoppilot/internal/models"
)

// SalesOrderService defines the interface for sales order business logic
type SalesOrderService interface {
	// Sales Order management
	CreateSalesOrder(ctx context.Context, clientID uuid.UUID, req *CreateSalesOrderRequest) (*models.SalesOrder, error)
	GetSalesOrder(ctx context.Context, clientID uuid.UUID, id uuid.UUID) (*models.SalesOrder, error)
	UpdateSalesOrder(ctx context.Context, clientID uuid.UUID, id uuid.UUID, req *UpdateSalesOrderRequest) error
	DeleteSalesOrder(ctx context.Context, clientID uuid.UUID, id uuid.UUID) error
	ListSalesOrders(ctx context.Context, clientID uuid.UUID, filters *SalesOrderFilters, page, pageSize int) ([]*models.SalesOrder, int, error)

	// Status transitions
	ConfirmSalesOrder(ctx context.Context, clientID uuid.UUID, id uuid.UUID) error
	CancelSalesOrder(ctx context.Context, clientID uuid.UUID, id uuid.UUID) error

	// Line items
	AddItem(ctx context.Context, clientID uuid.UUID, soID uuid.UUID, req *AddSalesOrderItemRequest) (*models.SalesOrderItem, error)
	RemoveItem(ctx context.Context, clientID uuid.UUID, soID uuid.UUID, itemID uuid.UUID) error
	ListItems(ctx context.Context, clientID uuid.UUID, soID uuid.UUID) ([]*models.SalesOrderItem, error)

	// Fulfillment
	FulfillItems(ctx context.Context, clientID uuid.UUID, soID uuid.UUID, items []FulfillItemRequest) error
}

// salesOrderService implements SalesOrderService interface
type salesOrderService struct {
	soRepo       repositories.SalesOrderRepository
	productRepo  repositories.ProductRepository
	movementRepo repositories.InventoryMovementRepository
}

// NewSalesOrderService creates a new sales order service
func NewSalesOrderService(
	soRepo repositories.SalesOrderRepository,
	productRepo repositories.ProductRepository,
	movementRepo repositories.InventoryMovementRepository,
) SalesOrderService {
	return &salesOrderService{
		soRepo:       soRepo,
		productRepo:  productRepo,
		movementRepo: movementRepo,
	}
}

// Request/Response types

type CreateSalesOrderRequest struct {
	CustomerID  uuid.UUID
	ShopID      uuid.UUID
	OrderNumber string // Optional - auto-generated if empty
	OrderDate   time.Time
	Subtotal    float64
	TaxAmount   float64
	TotalAmount float64
	Notes       string
	Metadata    map[string]interface{}
	Items       []CreateSalesOrderItemRequest
}

type CreateSalesOrderItemRequest struct {
	VariantID       uuid.UUID
	QuantityOrdered int
	UnitPrice       float64
	TotalPrice      float64
	Notes           string
}

type UpdateSalesOrderRequest struct {
	Notes    *string
	Metadata map[string]interface{}
}

type SalesOrderFilters struct {
	CustomerID *uuid.UUID
	ShopID     *uuid.UUID
	Status     *models.SalesOrderStatus
}

type AddSalesOrderItemRequest struct {
	VariantID       uuid.UUID
	QuantityOrdered int
	UnitPrice       float64
	Notes           string
}

type FulfillItemRequest struct {
	ItemID            uuid.UUID
	QuantityFulfilled int
}

// CreateSalesOrder creates a new sales order with business rule validations
func (s *salesOrderService) CreateSalesOrder(ctx context.Context, clientID uuid.UUID, req *CreateSalesOrderRequest) (*models.SalesOrder, error) {
	// Business rule: Auto-generate order number if not provided
	orderNumber := req.OrderNumber
	if orderNumber == "" {
		// Simple auto-generation: SO-{timestamp}
		orderNumber = fmt.Sprintf("SO-%d", time.Now().Unix())
	}

	// Business rule: Order number unique per client
	existingSO, err := s.soRepo.GetByOrderNumber(ctx, clientID, orderNumber)
	if err == nil && existingSO != nil {
		return nil, fmt.Errorf("sales order number '%s' already exists for this client", orderNumber)
	}

	so := &models.SalesOrder{
		ClientID:    clientID,
		CustomerID:  req.CustomerID,
		ShopID:      req.ShopID,
		OrderNumber: orderNumber,
		OrderDate:   req.OrderDate,
		Status:      models.SOStatusPending,
		Subtotal:    req.Subtotal,
		TaxAmount:   req.TaxAmount,
		TotalAmount: req.TotalAmount,
		Currency:    "USD", // Default currency
		Notes:       req.Notes,
		Metadata:    req.Metadata,
	}

	if err := s.soRepo.Create(ctx, so); err != nil {
		return nil, fmt.Errorf("failed to create sales order: %w", err)
	}

	// Create line items if provided
	for _, itemReq := range req.Items {
		item := &models.SalesOrderItem{
			ClientID:          clientID,
			SalesOrderID:      so.ID,
			VariantID:         itemReq.VariantID,
			QuantityOrdered:   itemReq.QuantityOrdered,
			QuantityFulfilled: 0,
			UnitPrice:         itemReq.UnitPrice,
			TotalPrice:        itemReq.TotalPrice,
			Notes:             itemReq.Notes,
		}

		if err := s.soRepo.CreateItem(ctx, item); err != nil {
			return nil, fmt.Errorf("failed to create sales order item: %w", err)
		}
	}

	return so, nil
}

// GetSalesOrder retrieves a sales order by ID
func (s *salesOrderService) GetSalesOrder(ctx context.Context, clientID uuid.UUID, id uuid.UUID) (*models.SalesOrder, error) {
	so, err := s.soRepo.GetByID(ctx, clientID, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get sales order: %w", err)
	}

	return so, nil
}

// UpdateSalesOrder updates an existing sales order
func (s *salesOrderService) UpdateSalesOrder(ctx context.Context, clientID uuid.UUID, id uuid.UUID, req *UpdateSalesOrderRequest) error {
	// Get existing sales order
	so, err := s.soRepo.GetByID(ctx, clientID, id)
	if err != nil {
		return fmt.Errorf("failed to get sales order: %w", err)
	}

	// Business rule: Only pending SOs can be modified
	if so.Status != models.SOStatusPending {
		return fmt.Errorf("only pending sales orders can be modified")
	}

	// Apply updates
	if req.Notes != nil {
		so.Notes = *req.Notes
	}
	if req.Metadata != nil {
		so.Metadata = req.Metadata
	}

	if err := s.soRepo.Update(ctx, so); err != nil {
		return fmt.Errorf("failed to update sales order: %w", err)
	}

	return nil
}

// DeleteSalesOrder deletes a sales order
func (s *salesOrderService) DeleteSalesOrder(ctx context.Context, clientID uuid.UUID, id uuid.UUID) error {
	// Get existing sales order
	so, err := s.soRepo.GetByID(ctx, clientID, id)
	if err != nil {
		return fmt.Errorf("failed to get sales order: %w", err)
	}

	// Business rule: Only pending SOs can be deleted
	if so.Status != models.SOStatusPending {
		return fmt.Errorf("only pending sales orders can be deleted")
	}

	if err := s.soRepo.Delete(ctx, clientID, id); err != nil {
		return fmt.Errorf("failed to delete sales order: %w", err)
	}

	return nil
}

// ListSalesOrders retrieves sales orders with filtering and pagination
func (s *salesOrderService) ListSalesOrders(ctx context.Context, clientID uuid.UUID, filters *SalesOrderFilters, page, pageSize int) ([]*models.SalesOrder, int, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	offset := (page - 1) * pageSize

	var sos []*models.SalesOrder
	var err error

	if filters != nil {
		if filters.CustomerID != nil {
			sos, err = s.soRepo.ListByCustomer(ctx, clientID, *filters.CustomerID, pageSize, offset)
		} else if filters.Status != nil {
			sos, err = s.soRepo.ListByStatus(ctx, clientID, *filters.Status, pageSize, offset)
		} else {
			sos, err = s.soRepo.ListByClient(ctx, clientID, pageSize, offset)
		}
	} else {
		sos, err = s.soRepo.ListByClient(ctx, clientID, pageSize, offset)
	}

	if err != nil {
		return nil, 0, fmt.Errorf("failed to list sales orders: %w", err)
	}

	total := len(sos)

	return sos, total, nil
}

// ConfirmSalesOrder transitions a sales order from pending to confirmed
func (s *salesOrderService) ConfirmSalesOrder(ctx context.Context, clientID uuid.UUID, id uuid.UUID) error {
	so, err := s.soRepo.GetByID(ctx, clientID, id)
	if err != nil {
		return fmt.Errorf("failed to get sales order: %w", err)
	}

	if so.Status != models.SOStatusPending {
		return fmt.Errorf("only pending sales orders can be confirmed")
	}

	// Business rule: Check inventory availability before confirming
	items, err := s.soRepo.ListItemsBySO(ctx, clientID, id)
	if err != nil {
		return fmt.Errorf("failed to list sales order items: %w", err)
	}

	for _, item := range items {
		variant, err := s.productRepo.GetVariantByID(ctx, clientID, item.VariantID)
		if err != nil {
			return fmt.Errorf("failed to get variant: %w", err)
		}

		if variant.Quantity < item.QuantityOrdered {
			return fmt.Errorf("insufficient inventory for variant %s: available=%d, required=%d",
				item.VariantID, variant.Quantity, item.QuantityOrdered)
		}
	}

	so.Status = models.SOStatusConfirmed
	if err := s.soRepo.Update(ctx, so); err != nil {
		return fmt.Errorf("failed to confirm sales order: %w", err)
	}

	return nil
}

// CancelSalesOrder transitions a sales order to cancelled
func (s *salesOrderService) CancelSalesOrder(ctx context.Context, clientID uuid.UUID, id uuid.UUID) error {
	so, err := s.soRepo.GetByID(ctx, clientID, id)
	if err != nil {
		return fmt.Errorf("failed to get sales order: %w", err)
	}

	if so.Status == models.SOStatusFulfilled || so.Status == models.SOStatusCancelled {
		return fmt.Errorf("cannot cancel a fulfilled or already cancelled sales order")
	}

	so.Status = models.SOStatusCancelled
	if err := s.soRepo.Update(ctx, so); err != nil {
		return fmt.Errorf("failed to cancel sales order: %w", err)
	}

	return nil
}

// AddItem adds a line item to a sales order
func (s *salesOrderService) AddItem(ctx context.Context, clientID uuid.UUID, soID uuid.UUID, req *AddSalesOrderItemRequest) (*models.SalesOrderItem, error) {
	so, err := s.soRepo.GetByID(ctx, clientID, soID)
	if err != nil {
		return nil, fmt.Errorf("failed to get sales order: %w", err)
	}

	// Business rule: Only pending SOs can be modified
	if so.Status != models.SOStatusPending {
		return nil, fmt.Errorf("can only add items to pending sales orders")
	}

	item := &models.SalesOrderItem{
		ClientID:          clientID,
		SalesOrderID:      soID,
		VariantID:         req.VariantID,
		QuantityOrdered:   req.QuantityOrdered,
		QuantityFulfilled: 0,
		UnitPrice:         req.UnitPrice,
		TotalPrice:        req.UnitPrice * float64(req.QuantityOrdered),
		Notes:             req.Notes,
	}

	if err := s.soRepo.CreateItem(ctx, item); err != nil {
		return nil, fmt.Errorf("failed to add item to sales order: %w", err)
	}

	return item, nil
}

// RemoveItem removes a line item from a sales order
func (s *salesOrderService) RemoveItem(ctx context.Context, clientID uuid.UUID, soID uuid.UUID, itemID uuid.UUID) error {
	so, err := s.soRepo.GetByID(ctx, clientID, soID)
	if err != nil {
		return fmt.Errorf("failed to get sales order: %w", err)
	}

	// Business rule: Only pending SOs can be modified
	if so.Status != models.SOStatusPending {
		return fmt.Errorf("can only remove items from pending sales orders")
	}

	if err := s.soRepo.DeleteItem(ctx, clientID, itemID); err != nil {
		return fmt.Errorf("failed to remove item from sales order: %w", err)
	}

	return nil
}

// ListItems retrieves all line items for a sales order
func (s *salesOrderService) ListItems(ctx context.Context, clientID uuid.UUID, soID uuid.UUID) ([]*models.SalesOrderItem, error) {
	items, err := s.soRepo.ListItemsBySO(ctx, clientID, soID)
	if err != nil {
		return nil, fmt.Errorf("failed to list sales order items: %w", err)
	}

	return items, nil
}

// FulfillItems processes fulfilled items for a sales order
// This is a multi-step transaction that:
// 1. Checks inventory availability
// 2. Updates sales_order_items.quantity_fulfilled
// 3. Updates product_variants.quantity (decrease)
// 4. Creates inventory_movements records
// 5. Updates sales_orders.status if fully fulfilled
func (s *salesOrderService) FulfillItems(ctx context.Context, clientID uuid.UUID, soID uuid.UUID, items []FulfillItemRequest) error {
	// Note: In a production system, this should use database transactions
	// For now, we'll process sequentially and rely on individual operations

	so, err := s.soRepo.GetByID(ctx, clientID, soID)
	if err != nil {
		return fmt.Errorf("failed to get sales order: %w", err)
	}

	if so.Status != models.SOStatusConfirmed && so.Status != models.SOStatusPartiallyFulfilled {
		return fmt.Errorf("can only fulfill items for confirmed or partially fulfilled sales orders")
	}

	// Process each item
	for _, fulfillReq := range items {
		// Get the item
		item, err := s.soRepo.GetItem(ctx, clientID, fulfillReq.ItemID)
		if err != nil {
			return fmt.Errorf("failed to get sales order item: %w", err)
		}

		// Verify item belongs to this SO
		if item.SalesOrderID != soID {
			return fmt.Errorf("item does not belong to this sales order")
		}

		// Business rule: Cannot fulfill more than ordered
		if item.QuantityFulfilled+fulfillReq.QuantityFulfilled > item.QuantityOrdered {
			return fmt.Errorf("cannot fulfill more than ordered: ordered=%d, already_fulfilled=%d, attempting=%d",
				item.QuantityOrdered, item.QuantityFulfilled, fulfillReq.QuantityFulfilled)
		}

		// Get the variant to update inventory
		variant, err := s.productRepo.GetVariantByID(ctx, clientID, item.VariantID)
		if err != nil {
			return fmt.Errorf("failed to get variant: %w", err)
		}

		// Business rule: Check inventory availability
		if variant.Quantity < fulfillReq.QuantityFulfilled {
			return fmt.Errorf("insufficient inventory: available=%d, required=%d",
				variant.Quantity, fulfillReq.QuantityFulfilled)
		}

		// 1. Update quantity_fulfilled on the item
		if err := s.soRepo.FulfillItem(ctx, clientID, fulfillReq.ItemID, fulfillReq.QuantityFulfilled); err != nil {
			return fmt.Errorf("failed to update fulfilled quantity: %w", err)
		}

		// 2. Update variant quantity (decrease)
		previousQuantity := variant.Quantity
		newQuantity := previousQuantity - fulfillReq.QuantityFulfilled
		if err := s.productRepo.UpdateInventory(ctx, clientID, item.VariantID, newQuantity); err != nil {
			return fmt.Errorf("failed to update inventory: %w", err)
		}

		// 3. Create inventory movement record
		refID := soID
		movement := &models.InventoryMovement{
			ClientID:         clientID,
			VariantID:        item.VariantID,
			ShopID:           so.ShopID,
			MovementType:     models.MovementTypeSale,
			Quantity:         -fulfillReq.QuantityFulfilled, // Negative for sales
			PreviousQuantity: previousQuantity,
			NewQuantity:      newQuantity,
			ReferenceType:    "sales_order",
			ReferenceID:      &refID,
			Notes:            fmt.Sprintf("Fulfilled from SO %s", so.OrderNumber),
		}

		if err := s.movementRepo.Create(ctx, movement); err != nil {
			return fmt.Errorf("failed to create inventory movement: %w", err)
		}
	}

	// 4. Check if all items are fully fulfilled and update SO status
	allItems, err := s.soRepo.ListItemsBySO(ctx, clientID, soID)
	if err != nil {
		return fmt.Errorf("failed to list sales order items: %w", err)
	}

	fullyFulfilled := true
	partiallyFulfilled := false
	for _, item := range allItems {
		if item.QuantityFulfilled < item.QuantityOrdered {
			fullyFulfilled = false
		}
		if item.QuantityFulfilled > 0 {
			partiallyFulfilled = true
		}
	}

	if fullyFulfilled {
		so.Status = models.SOStatusFulfilled
	} else if partiallyFulfilled {
		so.Status = models.SOStatusPartiallyFulfilled
	}

	if err := s.soRepo.Update(ctx, so); err != nil {
		return fmt.Errorf("failed to update sales order status: %w", err)
	}

	return nil
}
