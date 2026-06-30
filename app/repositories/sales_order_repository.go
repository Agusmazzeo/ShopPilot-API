package repositories

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/yourorg/shoppilot/internal/models"
)

// SalesOrderRepository defines the interface for sales order database operations
type SalesOrderRepository interface {
	// Sales Orders
	Create(ctx context.Context, so *models.SalesOrder) error
	GetByID(ctx context.Context, clientID uuid.UUID, id uuid.UUID) (*models.SalesOrder, error)
	GetByOrderNumber(ctx context.Context, clientID uuid.UUID, orderNumber string) (*models.SalesOrder, error)
	Update(ctx context.Context, so *models.SalesOrder) error
	Delete(ctx context.Context, clientID uuid.UUID, id uuid.UUID) error
	ListByClient(ctx context.Context, clientID uuid.UUID, limit, offset int) ([]*models.SalesOrder, error)
	ListByCustomer(ctx context.Context, clientID uuid.UUID, customerID uuid.UUID, limit, offset int) ([]*models.SalesOrder, error)
	ListByStatus(ctx context.Context, clientID uuid.UUID, status models.SalesOrderStatus, limit, offset int) ([]*models.SalesOrder, error)

	// Sales Order Items
	CreateItem(ctx context.Context, item *models.SalesOrderItem) error
	GetItem(ctx context.Context, clientID uuid.UUID, id uuid.UUID) (*models.SalesOrderItem, error)
	UpdateItem(ctx context.Context, item *models.SalesOrderItem) error
	DeleteItem(ctx context.Context, clientID uuid.UUID, id uuid.UUID) error
	ListItemsBySO(ctx context.Context, clientID uuid.UUID, soID uuid.UUID) ([]*models.SalesOrderItem, error)

	// Fulfillment
	FulfillItem(ctx context.Context, clientID uuid.UUID, itemID uuid.UUID, quantityFulfilled int) error
}

// salesOrderRepository handles database operations for sales orders
type salesOrderRepository struct {
	*BaseRepository
}

// NewSalesOrderRepository creates a new sales order repository
func NewSalesOrderRepository(pool *pgxpool.Pool) SalesOrderRepository {
	return &salesOrderRepository{
		BaseRepository: NewBaseRepository(pool),
	}
}

// Create inserts a new sales order into the database
func (r *salesOrderRepository) Create(ctx context.Context, so *models.SalesOrder) error {
	query := `
		INSERT INTO sales_orders (
			client_id, customer_id, shop_id, order_number, status, order_date,
			shipping_date, delivery_date, subtotal, tax_amount, shipping_amount,
			total_amount, currency, shipping_address, billing_address, notes, metadata
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17)
		RETURNING id, created_at, updated_at
	`

	err := r.pool.QueryRow(
		ctx,
		query,
		so.ClientID,
		so.CustomerID,
		so.ShopID,
		so.OrderNumber,
		so.Status,
		so.OrderDate,
		so.ShippingDate,
		so.DeliveryDate,
		so.Subtotal,
		so.TaxAmount,
		so.ShippingAmount,
		so.TotalAmount,
		so.Currency,
		so.ShippingAddress,
		so.BillingAddress,
		so.Notes,
		so.Metadata,
	).Scan(&so.ID, &so.CreatedAt, &so.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create sales order: %w", err)
	}

	return nil
}

// GetByID retrieves a sales order by ID
func (r *salesOrderRepository) GetByID(ctx context.Context, clientID uuid.UUID, id uuid.UUID) (*models.SalesOrder, error) {
	query := `
		SELECT id, client_id, customer_id, shop_id, order_number, status, order_date,
		       shipping_date, delivery_date, subtotal, tax_amount, shipping_amount,
		       total_amount, currency, shipping_address, billing_address,
		       notes, metadata, created_at, updated_at
		FROM sales_orders
		WHERE client_id = $1 AND id = $2
	`

	var so models.SalesOrder
	err := r.pool.QueryRow(ctx, query, clientID, id).Scan(
		&so.ID,
		&so.ClientID,
		&so.CustomerID,
		&so.ShopID,
		&so.OrderNumber,
		&so.Status,
		&so.OrderDate,
		&so.ShippingDate,
		&so.DeliveryDate,
		&so.Subtotal,
		&so.TaxAmount,
		&so.ShippingAmount,
		&so.TotalAmount,
		&so.Currency,
		&so.ShippingAddress,
		&so.BillingAddress,
		&so.Notes,
		&so.Metadata,
		&so.CreatedAt,
		&so.UpdatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("sales order not found: %s", id)
		}
		return nil, fmt.Errorf("failed to get sales order: %w", err)
	}

	return &so, nil
}

// GetByOrderNumber retrieves a sales order by order number
func (r *salesOrderRepository) GetByOrderNumber(ctx context.Context, clientID uuid.UUID, orderNumber string) (*models.SalesOrder, error) {
	query := `
		SELECT id, client_id, customer_id, shop_id, order_number, status, order_date,
		       shipping_date, delivery_date, subtotal, tax_amount, shipping_amount,
		       total_amount, currency, shipping_address, billing_address,
		       notes, metadata, created_at, updated_at
		FROM sales_orders
		WHERE client_id = $1 AND order_number = $2
	`

	var so models.SalesOrder
	err := r.pool.QueryRow(ctx, query, clientID, orderNumber).Scan(
		&so.ID,
		&so.ClientID,
		&so.CustomerID,
		&so.ShopID,
		&so.OrderNumber,
		&so.Status,
		&so.OrderDate,
		&so.ShippingDate,
		&so.DeliveryDate,
		&so.Subtotal,
		&so.TaxAmount,
		&so.ShippingAmount,
		&so.TotalAmount,
		&so.Currency,
		&so.ShippingAddress,
		&so.BillingAddress,
		&so.Notes,
		&so.Metadata,
		&so.CreatedAt,
		&so.UpdatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("sales order not found: %s", orderNumber)
		}
		return nil, fmt.Errorf("failed to get sales order: %w", err)
	}

	return &so, nil
}

// Update updates an existing sales order
func (r *salesOrderRepository) Update(ctx context.Context, so *models.SalesOrder) error {
	query := `
		UPDATE sales_orders
		SET status = $3, shipping_date = $4, delivery_date = $5,
		    subtotal = $6, tax_amount = $7, shipping_amount = $8,
		    total_amount = $9, shipping_address = $10, billing_address = $11,
		    notes = $12, metadata = $13, updated_at = NOW()
		WHERE client_id = $1 AND id = $2
	`

	result, err := r.pool.Exec(
		ctx,
		query,
		so.ClientID,
		so.ID,
		so.Status,
		so.ShippingDate,
		so.DeliveryDate,
		so.Subtotal,
		so.TaxAmount,
		so.ShippingAmount,
		so.TotalAmount,
		so.ShippingAddress,
		so.BillingAddress,
		so.Notes,
		so.Metadata,
	)

	if err != nil {
		return fmt.Errorf("failed to update sales order: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("sales order not found: %s", so.ID)
	}

	return nil
}

// Delete removes a sales order
func (r *salesOrderRepository) Delete(ctx context.Context, clientID uuid.UUID, id uuid.UUID) error {
	query := `DELETE FROM sales_orders WHERE client_id = $1 AND id = $2`

	result, err := r.pool.Exec(ctx, query, clientID, id)
	if err != nil {
		return fmt.Errorf("failed to delete sales order: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("sales order not found: %s", id)
	}

	return nil
}

// ListByClient retrieves sales orders for a client
func (r *salesOrderRepository) ListByClient(ctx context.Context, clientID uuid.UUID, limit, offset int) ([]*models.SalesOrder, error) {
	query := `
		SELECT id, client_id, customer_id, shop_id, order_number, status, order_date,
		       shipping_date, delivery_date, subtotal, tax_amount, shipping_amount,
		       total_amount, currency, shipping_address, billing_address,
		       notes, metadata, created_at, updated_at
		FROM sales_orders
		WHERE client_id = $1
		ORDER BY order_date DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.pool.Query(ctx, query, clientID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list sales orders: %w", err)
	}
	defer rows.Close()

	return r.scanSalesOrders(rows)
}

// ListByCustomer retrieves sales orders for a specific customer
func (r *salesOrderRepository) ListByCustomer(ctx context.Context, clientID uuid.UUID, customerID uuid.UUID, limit, offset int) ([]*models.SalesOrder, error) {
	query := `
		SELECT id, client_id, customer_id, shop_id, order_number, status, order_date,
		       shipping_date, delivery_date, subtotal, tax_amount, shipping_amount,
		       total_amount, currency, shipping_address, billing_address,
		       notes, metadata, created_at, updated_at
		FROM sales_orders
		WHERE client_id = $1 AND customer_id = $2
		ORDER BY order_date DESC
		LIMIT $3 OFFSET $4
	`

	rows, err := r.pool.Query(ctx, query, clientID, customerID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list sales orders by customer: %w", err)
	}
	defer rows.Close()

	return r.scanSalesOrders(rows)
}

// ListByStatus retrieves sales orders with a specific status
func (r *salesOrderRepository) ListByStatus(ctx context.Context, clientID uuid.UUID, status models.SalesOrderStatus, limit, offset int) ([]*models.SalesOrder, error) {
	query := `
		SELECT id, client_id, customer_id, shop_id, order_number, status, order_date,
		       shipping_date, delivery_date, subtotal, tax_amount, shipping_amount,
		       total_amount, currency, shipping_address, billing_address,
		       notes, metadata, created_at, updated_at
		FROM sales_orders
		WHERE client_id = $1 AND status = $2
		ORDER BY order_date DESC
		LIMIT $3 OFFSET $4
	`

	rows, err := r.pool.Query(ctx, query, clientID, status, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list sales orders by status: %w", err)
	}
	defer rows.Close()

	return r.scanSalesOrders(rows)
}

// CreateItem inserts a new sales order item
func (r *salesOrderRepository) CreateItem(ctx context.Context, item *models.SalesOrderItem) error {
	query := `
		INSERT INTO sales_order_items (
			client_id, sales_order_id, variant_id, quantity_ordered,
			unit_price, tax_rate, discount_amount, total_price, notes
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, quantity_fulfilled, created_at, updated_at
	`

	err := r.pool.QueryRow(
		ctx,
		query,
		item.ClientID,
		item.SalesOrderID,
		item.VariantID,
		item.QuantityOrdered,
		item.UnitPrice,
		item.TaxRate,
		item.DiscountAmount,
		item.TotalPrice,
		item.Notes,
	).Scan(&item.ID, &item.QuantityFulfilled, &item.CreatedAt, &item.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create sales order item: %w", err)
	}

	return nil
}

// GetItem retrieves a sales order item by ID
func (r *salesOrderRepository) GetItem(ctx context.Context, clientID uuid.UUID, id uuid.UUID) (*models.SalesOrderItem, error) {
	query := `
		SELECT id, client_id, sales_order_id, variant_id, quantity_ordered,
		       quantity_fulfilled, unit_price, tax_rate, discount_amount,
		       total_price, notes, created_at, updated_at
		FROM sales_order_items
		WHERE client_id = $1 AND id = $2
	`

	var item models.SalesOrderItem
	err := r.pool.QueryRow(ctx, query, clientID, id).Scan(
		&item.ID,
		&item.ClientID,
		&item.SalesOrderID,
		&item.VariantID,
		&item.QuantityOrdered,
		&item.QuantityFulfilled,
		&item.UnitPrice,
		&item.TaxRate,
		&item.DiscountAmount,
		&item.TotalPrice,
		&item.Notes,
		&item.CreatedAt,
		&item.UpdatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("sales order item not found: %s", id)
		}
		return nil, fmt.Errorf("failed to get sales order item: %w", err)
	}

	return &item, nil
}

// UpdateItem updates a sales order item
func (r *salesOrderRepository) UpdateItem(ctx context.Context, item *models.SalesOrderItem) error {
	query := `
		UPDATE sales_order_items
		SET quantity_ordered = $3, quantity_fulfilled = $4, unit_price = $5,
		    tax_rate = $6, discount_amount = $7, total_price = $8,
		    notes = $9, updated_at = NOW()
		WHERE client_id = $1 AND id = $2
	`

	result, err := r.pool.Exec(
		ctx,
		query,
		item.ClientID,
		item.ID,
		item.QuantityOrdered,
		item.QuantityFulfilled,
		item.UnitPrice,
		item.TaxRate,
		item.DiscountAmount,
		item.TotalPrice,
		item.Notes,
	)

	if err != nil {
		return fmt.Errorf("failed to update sales order item: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("sales order item not found: %s", item.ID)
	}

	return nil
}

// DeleteItem removes a sales order item
func (r *salesOrderRepository) DeleteItem(ctx context.Context, clientID uuid.UUID, id uuid.UUID) error {
	query := `DELETE FROM sales_order_items WHERE client_id = $1 AND id = $2`

	result, err := r.pool.Exec(ctx, query, clientID, id)
	if err != nil {
		return fmt.Errorf("failed to delete sales order item: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("sales order item not found: %s", id)
	}

	return nil
}

// ListItemsBySO retrieves all items for a sales order
func (r *salesOrderRepository) ListItemsBySO(ctx context.Context, clientID uuid.UUID, soID uuid.UUID) ([]*models.SalesOrderItem, error) {
	query := `
		SELECT id, client_id, sales_order_id, variant_id, quantity_ordered,
		       quantity_fulfilled, unit_price, tax_rate, discount_amount,
		       total_price, notes, created_at, updated_at
		FROM sales_order_items
		WHERE client_id = $1 AND sales_order_id = $2
		ORDER BY created_at
	`

	rows, err := r.pool.Query(ctx, query, clientID, soID)
	if err != nil {
		return nil, fmt.Errorf("failed to list sales order items: %w", err)
	}
	defer rows.Close()

	var items []*models.SalesOrderItem
	for rows.Next() {
		var item models.SalesOrderItem
		err := rows.Scan(
			&item.ID,
			&item.ClientID,
			&item.SalesOrderID,
			&item.VariantID,
			&item.QuantityOrdered,
			&item.QuantityFulfilled,
			&item.UnitPrice,
			&item.TaxRate,
			&item.DiscountAmount,
			&item.TotalPrice,
			&item.Notes,
			&item.CreatedAt,
			&item.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan sales order item: %w", err)
		}
		items = append(items, &item)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating sales order items: %w", err)
	}

	return items, nil
}

// FulfillItem updates the quantity fulfilled for a sales order item
func (r *salesOrderRepository) FulfillItem(ctx context.Context, clientID uuid.UUID, itemID uuid.UUID, quantityFulfilled int) error {
	query := `
		UPDATE sales_order_items
		SET quantity_fulfilled = quantity_fulfilled + $3, updated_at = NOW()
		WHERE client_id = $1 AND id = $2
		  AND quantity_fulfilled + $3 <= quantity_ordered
	`

	result, err := r.pool.Exec(ctx, query, clientID, itemID, quantityFulfilled)
	if err != nil {
		return fmt.Errorf("failed to fulfill sales order item: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("sales order item not found or quantity exceeds ordered amount: %s", itemID)
	}

	return nil
}

// scanSalesOrders is a helper to scan sales order rows
func (r *salesOrderRepository) scanSalesOrders(rows pgx.Rows) ([]*models.SalesOrder, error) {
	var orders []*models.SalesOrder
	for rows.Next() {
		var so models.SalesOrder
		err := rows.Scan(
			&so.ID,
			&so.ClientID,
			&so.CustomerID,
			&so.ShopID,
			&so.OrderNumber,
			&so.Status,
			&so.OrderDate,
			&so.ShippingDate,
			&so.DeliveryDate,
			&so.Subtotal,
			&so.TaxAmount,
			&so.ShippingAmount,
			&so.TotalAmount,
			&so.Currency,
			&so.ShippingAddress,
			&so.BillingAddress,
			&so.Notes,
			&so.Metadata,
			&so.CreatedAt,
			&so.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan sales order: %w", err)
		}
		orders = append(orders, &so)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating sales orders: %w", err)
	}

	return orders, nil
}
