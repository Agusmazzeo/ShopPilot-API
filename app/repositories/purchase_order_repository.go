package repositories

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/yourorg/shoppilot/internal/models"
)

// PurchaseOrderRepository defines the interface for purchase order database operations
type PurchaseOrderRepository interface {
	// Purchase Orders
	Create(ctx context.Context, po *models.PurchaseOrder) error
	GetByID(ctx context.Context, clientID uuid.UUID, id uuid.UUID) (*models.PurchaseOrder, error)
	GetByPONumber(ctx context.Context, clientID uuid.UUID, poNumber string) (*models.PurchaseOrder, error)
	Update(ctx context.Context, po *models.PurchaseOrder) error
	Delete(ctx context.Context, clientID uuid.UUID, id uuid.UUID) error
	ListByClient(ctx context.Context, clientID uuid.UUID, limit, offset int) ([]*models.PurchaseOrder, error)
	ListBySupplier(ctx context.Context, clientID uuid.UUID, supplierID uuid.UUID, limit, offset int) ([]*models.PurchaseOrder, error)
	ListByStatus(ctx context.Context, clientID uuid.UUID, status models.PurchaseOrderStatus, limit, offset int) ([]*models.PurchaseOrder, error)

	// Purchase Order Items
	CreateItem(ctx context.Context, item *models.PurchaseOrderItem) error
	GetItem(ctx context.Context, clientID uuid.UUID, id uuid.UUID) (*models.PurchaseOrderItem, error)
	UpdateItem(ctx context.Context, item *models.PurchaseOrderItem) error
	DeleteItem(ctx context.Context, clientID uuid.UUID, id uuid.UUID) error
	ListItemsByPO(ctx context.Context, clientID uuid.UUID, poID uuid.UUID) ([]*models.PurchaseOrderItem, error)

	// Receiving
	ReceiveItem(ctx context.Context, clientID uuid.UUID, itemID uuid.UUID, quantityReceived int) error
}

// purchaseOrderRepository handles database operations for purchase orders
type purchaseOrderRepository struct {
	*BaseRepository
}

// NewPurchaseOrderRepository creates a new purchase order repository
func NewPurchaseOrderRepository(pool *pgxpool.Pool) PurchaseOrderRepository {
	return &purchaseOrderRepository{
		BaseRepository: NewBaseRepository(pool),
	}
}

// Create inserts a new purchase order into the database
func (r *purchaseOrderRepository) Create(ctx context.Context, po *models.PurchaseOrder) error {
	query := `
		INSERT INTO purchase_orders (
			client_id, supplier_id, shop_id, po_number, status, order_date,
			expected_delivery_date, total_amount, currency, notes, metadata
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING id, created_at, updated_at
	`

	err := r.pool.QueryRow(
		ctx,
		query,
		po.ClientID,
		po.SupplierID,
		po.ShopID,
		po.PONumber,
		po.Status,
		po.OrderDate,
		po.ExpectedDeliveryDate,
		po.TotalAmount,
		po.Currency,
		po.Notes,
		po.Metadata,
	).Scan(&po.ID, &po.CreatedAt, &po.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create purchase order: %w", err)
	}

	return nil
}

// GetByID retrieves a purchase order by ID
func (r *purchaseOrderRepository) GetByID(ctx context.Context, clientID uuid.UUID, id uuid.UUID) (*models.PurchaseOrder, error) {
	query := `
		SELECT id, client_id, supplier_id, shop_id, po_number, status, order_date,
		       expected_delivery_date, received_date, total_amount, currency,
		       notes, metadata, created_at, updated_at
		FROM purchase_orders
		WHERE client_id = $1 AND id = $2
	`

	var po models.PurchaseOrder
	err := r.pool.QueryRow(ctx, query, clientID, id).Scan(
		&po.ID,
		&po.ClientID,
		&po.SupplierID,
		&po.ShopID,
		&po.PONumber,
		&po.Status,
		&po.OrderDate,
		&po.ExpectedDeliveryDate,
		&po.ReceivedDate,
		&po.TotalAmount,
		&po.Currency,
		&po.Notes,
		&po.Metadata,
		&po.CreatedAt,
		&po.UpdatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("purchase order not found: %s", id)
		}
		return nil, fmt.Errorf("failed to get purchase order: %w", err)
	}

	return &po, nil
}

// GetByPONumber retrieves a purchase order by PO number
func (r *purchaseOrderRepository) GetByPONumber(ctx context.Context, clientID uuid.UUID, poNumber string) (*models.PurchaseOrder, error) {
	query := `
		SELECT id, client_id, supplier_id, shop_id, po_number, status, order_date,
		       expected_delivery_date, received_date, total_amount, currency,
		       notes, metadata, created_at, updated_at
		FROM purchase_orders
		WHERE client_id = $1 AND po_number = $2
	`

	var po models.PurchaseOrder
	err := r.pool.QueryRow(ctx, query, clientID, poNumber).Scan(
		&po.ID,
		&po.ClientID,
		&po.SupplierID,
		&po.ShopID,
		&po.PONumber,
		&po.Status,
		&po.OrderDate,
		&po.ExpectedDeliveryDate,
		&po.ReceivedDate,
		&po.TotalAmount,
		&po.Currency,
		&po.Notes,
		&po.Metadata,
		&po.CreatedAt,
		&po.UpdatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("purchase order not found: %s", poNumber)
		}
		return nil, fmt.Errorf("failed to get purchase order: %w", err)
	}

	return &po, nil
}

// Update updates an existing purchase order
func (r *purchaseOrderRepository) Update(ctx context.Context, po *models.PurchaseOrder) error {
	query := `
		UPDATE purchase_orders
		SET status = $3, expected_delivery_date = $4, received_date = $5,
		    total_amount = $6, notes = $7, metadata = $8, updated_at = NOW()
		WHERE client_id = $1 AND id = $2
	`

	result, err := r.pool.Exec(
		ctx,
		query,
		po.ClientID,
		po.ID,
		po.Status,
		po.ExpectedDeliveryDate,
		po.ReceivedDate,
		po.TotalAmount,
		po.Notes,
		po.Metadata,
	)

	if err != nil {
		return fmt.Errorf("failed to update purchase order: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("purchase order not found: %s", po.ID)
	}

	return nil
}

// Delete removes a purchase order
func (r *purchaseOrderRepository) Delete(ctx context.Context, clientID uuid.UUID, id uuid.UUID) error {
	query := `DELETE FROM purchase_orders WHERE client_id = $1 AND id = $2`

	result, err := r.pool.Exec(ctx, query, clientID, id)
	if err != nil {
		return fmt.Errorf("failed to delete purchase order: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("purchase order not found: %s", id)
	}

	return nil
}

// ListByClient retrieves purchase orders for a client
func (r *purchaseOrderRepository) ListByClient(ctx context.Context, clientID uuid.UUID, limit, offset int) ([]*models.PurchaseOrder, error) {
	query := `
		SELECT id, client_id, supplier_id, shop_id, po_number, status, order_date,
		       expected_delivery_date, received_date, total_amount, currency,
		       notes, metadata, created_at, updated_at
		FROM purchase_orders
		WHERE client_id = $1
		ORDER BY order_date DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.pool.Query(ctx, query, clientID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list purchase orders: %w", err)
	}
	defer rows.Close()

	return r.scanPurchaseOrders(rows)
}

// ListBySupplier retrieves purchase orders for a specific supplier
func (r *purchaseOrderRepository) ListBySupplier(ctx context.Context, clientID uuid.UUID, supplierID uuid.UUID, limit, offset int) ([]*models.PurchaseOrder, error) {
	query := `
		SELECT id, client_id, supplier_id, shop_id, po_number, status, order_date,
		       expected_delivery_date, received_date, total_amount, currency,
		       notes, metadata, created_at, updated_at
		FROM purchase_orders
		WHERE client_id = $1 AND supplier_id = $2
		ORDER BY order_date DESC
		LIMIT $3 OFFSET $4
	`

	rows, err := r.pool.Query(ctx, query, clientID, supplierID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list purchase orders by supplier: %w", err)
	}
	defer rows.Close()

	return r.scanPurchaseOrders(rows)
}

// ListByStatus retrieves purchase orders with a specific status
func (r *purchaseOrderRepository) ListByStatus(ctx context.Context, clientID uuid.UUID, status models.PurchaseOrderStatus, limit, offset int) ([]*models.PurchaseOrder, error) {
	query := `
		SELECT id, client_id, supplier_id, shop_id, po_number, status, order_date,
		       expected_delivery_date, received_date, total_amount, currency,
		       notes, metadata, created_at, updated_at
		FROM purchase_orders
		WHERE client_id = $1 AND status = $2
		ORDER BY order_date DESC
		LIMIT $3 OFFSET $4
	`

	rows, err := r.pool.Query(ctx, query, clientID, status, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list purchase orders by status: %w", err)
	}
	defer rows.Close()

	return r.scanPurchaseOrders(rows)
}

// CreateItem inserts a new purchase order item
func (r *purchaseOrderRepository) CreateItem(ctx context.Context, item *models.PurchaseOrderItem) error {
	query := `
		INSERT INTO purchase_order_items (
			client_id, purchase_order_id, variant_id, quantity_ordered,
			unit_cost, total_cost, notes
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, quantity_received, created_at, updated_at
	`

	err := r.pool.QueryRow(
		ctx,
		query,
		item.ClientID,
		item.PurchaseOrderID,
		item.VariantID,
		item.QuantityOrdered,
		item.UnitCost,
		item.TotalCost,
		item.Notes,
	).Scan(&item.ID, &item.QuantityReceived, &item.CreatedAt, &item.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create purchase order item: %w", err)
	}

	return nil
}

// GetItem retrieves a purchase order item by ID
func (r *purchaseOrderRepository) GetItem(ctx context.Context, clientID uuid.UUID, id uuid.UUID) (*models.PurchaseOrderItem, error) {
	query := `
		SELECT id, client_id, purchase_order_id, variant_id, quantity_ordered,
		       quantity_received, unit_cost, total_cost, notes, created_at, updated_at
		FROM purchase_order_items
		WHERE client_id = $1 AND id = $2
	`

	var item models.PurchaseOrderItem
	err := r.pool.QueryRow(ctx, query, clientID, id).Scan(
		&item.ID,
		&item.ClientID,
		&item.PurchaseOrderID,
		&item.VariantID,
		&item.QuantityOrdered,
		&item.QuantityReceived,
		&item.UnitCost,
		&item.TotalCost,
		&item.Notes,
		&item.CreatedAt,
		&item.UpdatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("purchase order item not found: %s", id)
		}
		return nil, fmt.Errorf("failed to get purchase order item: %w", err)
	}

	return &item, nil
}

// UpdateItem updates a purchase order item
func (r *purchaseOrderRepository) UpdateItem(ctx context.Context, item *models.PurchaseOrderItem) error {
	query := `
		UPDATE purchase_order_items
		SET quantity_ordered = $3, quantity_received = $4, unit_cost = $5,
		    total_cost = $6, notes = $7, updated_at = NOW()
		WHERE client_id = $1 AND id = $2
	`

	result, err := r.pool.Exec(
		ctx,
		query,
		item.ClientID,
		item.ID,
		item.QuantityOrdered,
		item.QuantityReceived,
		item.UnitCost,
		item.TotalCost,
		item.Notes,
	)

	if err != nil {
		return fmt.Errorf("failed to update purchase order item: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("purchase order item not found: %s", item.ID)
	}

	return nil
}

// DeleteItem removes a purchase order item
func (r *purchaseOrderRepository) DeleteItem(ctx context.Context, clientID uuid.UUID, id uuid.UUID) error {
	query := `DELETE FROM purchase_order_items WHERE client_id = $1 AND id = $2`

	result, err := r.pool.Exec(ctx, query, clientID, id)
	if err != nil {
		return fmt.Errorf("failed to delete purchase order item: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("purchase order item not found: %s", id)
	}

	return nil
}

// ListItemsByPO retrieves all items for a purchase order
func (r *purchaseOrderRepository) ListItemsByPO(ctx context.Context, clientID uuid.UUID, poID uuid.UUID) ([]*models.PurchaseOrderItem, error) {
	query := `
		SELECT id, client_id, purchase_order_id, variant_id, quantity_ordered,
		       quantity_received, unit_cost, total_cost, notes, created_at, updated_at
		FROM purchase_order_items
		WHERE client_id = $1 AND purchase_order_id = $2
		ORDER BY created_at
	`

	rows, err := r.pool.Query(ctx, query, clientID, poID)
	if err != nil {
		return nil, fmt.Errorf("failed to list purchase order items: %w", err)
	}
	defer rows.Close()

	var items []*models.PurchaseOrderItem
	for rows.Next() {
		var item models.PurchaseOrderItem
		err := rows.Scan(
			&item.ID,
			&item.ClientID,
			&item.PurchaseOrderID,
			&item.VariantID,
			&item.QuantityOrdered,
			&item.QuantityReceived,
			&item.UnitCost,
			&item.TotalCost,
			&item.Notes,
			&item.CreatedAt,
			&item.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan purchase order item: %w", err)
		}
		items = append(items, &item)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating purchase order items: %w", err)
	}

	return items, nil
}

// ReceiveItem updates the quantity received for a purchase order item
func (r *purchaseOrderRepository) ReceiveItem(ctx context.Context, clientID uuid.UUID, itemID uuid.UUID, quantityReceived int) error {
	query := `
		UPDATE purchase_order_items
		SET quantity_received = quantity_received + $3, updated_at = NOW()
		WHERE client_id = $1 AND id = $2
		  AND quantity_received + $3 <= quantity_ordered
	`

	result, err := r.pool.Exec(ctx, query, clientID, itemID, quantityReceived)
	if err != nil {
		return fmt.Errorf("failed to receive purchase order item: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("purchase order item not found or quantity exceeds ordered amount: %s", itemID)
	}

	return nil
}

// scanPurchaseOrders is a helper to scan purchase order rows
func (r *purchaseOrderRepository) scanPurchaseOrders(rows pgx.Rows) ([]*models.PurchaseOrder, error) {
	var orders []*models.PurchaseOrder
	for rows.Next() {
		var po models.PurchaseOrder
		err := rows.Scan(
			&po.ID,
			&po.ClientID,
			&po.SupplierID,
			&po.ShopID,
			&po.PONumber,
			&po.Status,
			&po.OrderDate,
			&po.ExpectedDeliveryDate,
			&po.ReceivedDate,
			&po.TotalAmount,
			&po.Currency,
			&po.Notes,
			&po.Metadata,
			&po.CreatedAt,
			&po.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan purchase order: %w", err)
		}
		orders = append(orders, &po)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating purchase orders: %w", err)
	}

	return orders, nil
}
