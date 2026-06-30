package repositories

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/yourorg/shoppilot/internal/models"
)

// InventoryAlertRepository defines the interface for inventory alert database operations
type InventoryAlertRepository interface {
	Create(ctx context.Context, alert *models.InventoryAlert) error
	GetByID(ctx context.Context, clientID uuid.UUID, id uuid.UUID) (*models.InventoryAlert, error)
	GetByVariant(ctx context.Context, clientID uuid.UUID, variantID uuid.UUID, shopID uuid.UUID) (*models.InventoryAlert, error)
	Update(ctx context.Context, alert *models.InventoryAlert) error
	Delete(ctx context.Context, clientID uuid.UUID, id uuid.UUID) error
	ListByShop(ctx context.Context, clientID uuid.UUID, shopID uuid.UUID, limit, offset int) ([]*models.InventoryAlert, error)
	ListTriggered(ctx context.Context, clientID uuid.UUID) ([]*models.InventoryAlert, error)
}

// inventoryAlertRepository handles database operations for inventory alerts
type inventoryAlertRepository struct {
	*BaseRepository
}

// NewInventoryAlertRepository creates a new inventory alert repository
func NewInventoryAlertRepository(pool *pgxpool.Pool) InventoryAlertRepository {
	return &inventoryAlertRepository{
		BaseRepository: NewBaseRepository(pool),
	}
}

// Create inserts a new inventory alert into the database
func (r *inventoryAlertRepository) Create(ctx context.Context, alert *models.InventoryAlert) error {
	query := `
		INSERT INTO inventory_alerts (
			client_id, variant_id, shop_id, reorder_point, reorder_quantity,
			low_stock_threshold, is_enabled, metadata
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, created_at, updated_at
	`

	err := r.pool.QueryRow(
		ctx,
		query,
		alert.ClientID,
		alert.VariantID,
		alert.ShopID,
		alert.ReorderPoint,
		alert.ReorderQuantity,
		alert.LowStockThreshold,
		alert.IsEnabled,
		alert.Metadata,
	).Scan(&alert.ID, &alert.CreatedAt, &alert.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create inventory alert: %w", err)
	}

	return nil
}

// GetByID retrieves an inventory alert by ID
func (r *inventoryAlertRepository) GetByID(ctx context.Context, clientID uuid.UUID, id uuid.UUID) (*models.InventoryAlert, error) {
	query := `
		SELECT id, client_id, variant_id, shop_id, reorder_point, reorder_quantity,
		       low_stock_threshold, is_enabled, last_alert_sent_at, metadata,
		       created_at, updated_at
		FROM inventory_alerts
		WHERE client_id = $1 AND id = $2
	`

	var alert models.InventoryAlert
	err := r.pool.QueryRow(ctx, query, clientID, id).Scan(
		&alert.ID,
		&alert.ClientID,
		&alert.VariantID,
		&alert.ShopID,
		&alert.ReorderPoint,
		&alert.ReorderQuantity,
		&alert.LowStockThreshold,
		&alert.IsEnabled,
		&alert.LastAlertSentAt,
		&alert.Metadata,
		&alert.CreatedAt,
		&alert.UpdatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("inventory alert not found: %s", id)
		}
		return nil, fmt.Errorf("failed to get inventory alert: %w", err)
	}

	return &alert, nil
}

// GetByVariant retrieves an inventory alert for a specific variant and shop
func (r *inventoryAlertRepository) GetByVariant(ctx context.Context, clientID uuid.UUID, variantID uuid.UUID, shopID uuid.UUID) (*models.InventoryAlert, error) {
	query := `
		SELECT id, client_id, variant_id, shop_id, reorder_point, reorder_quantity,
		       low_stock_threshold, is_enabled, last_alert_sent_at, metadata,
		       created_at, updated_at
		FROM inventory_alerts
		WHERE client_id = $1 AND variant_id = $2 AND shop_id = $3
	`

	var alert models.InventoryAlert
	err := r.pool.QueryRow(ctx, query, clientID, variantID, shopID).Scan(
		&alert.ID,
		&alert.ClientID,
		&alert.VariantID,
		&alert.ShopID,
		&alert.ReorderPoint,
		&alert.ReorderQuantity,
		&alert.LowStockThreshold,
		&alert.IsEnabled,
		&alert.LastAlertSentAt,
		&alert.Metadata,
		&alert.CreatedAt,
		&alert.UpdatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("inventory alert not found for variant: %s", variantID)
		}
		return nil, fmt.Errorf("failed to get inventory alert: %w", err)
	}

	return &alert, nil
}

// Update updates an existing inventory alert
func (r *inventoryAlertRepository) Update(ctx context.Context, alert *models.InventoryAlert) error {
	query := `
		UPDATE inventory_alerts
		SET reorder_point = $3, reorder_quantity = $4, low_stock_threshold = $5,
		    is_enabled = $6, last_alert_sent_at = $7, metadata = $8, updated_at = NOW()
		WHERE client_id = $1 AND id = $2
	`

	result, err := r.pool.Exec(
		ctx,
		query,
		alert.ClientID,
		alert.ID,
		alert.ReorderPoint,
		alert.ReorderQuantity,
		alert.LowStockThreshold,
		alert.IsEnabled,
		alert.LastAlertSentAt,
		alert.Metadata,
	)

	if err != nil {
		return fmt.Errorf("failed to update inventory alert: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("inventory alert not found: %s", alert.ID)
	}

	return nil
}

// Delete removes an inventory alert
func (r *inventoryAlertRepository) Delete(ctx context.Context, clientID uuid.UUID, id uuid.UUID) error {
	query := `DELETE FROM inventory_alerts WHERE client_id = $1 AND id = $2`

	result, err := r.pool.Exec(ctx, query, clientID, id)
	if err != nil {
		return fmt.Errorf("failed to delete inventory alert: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("inventory alert not found: %s", id)
	}

	return nil
}

// ListByShop retrieves inventory alerts for a specific shop
func (r *inventoryAlertRepository) ListByShop(ctx context.Context, clientID uuid.UUID, shopID uuid.UUID, limit, offset int) ([]*models.InventoryAlert, error) {
	query := `
		SELECT id, client_id, variant_id, shop_id, reorder_point, reorder_quantity,
		       low_stock_threshold, is_enabled, last_alert_sent_at, metadata,
		       created_at, updated_at
		FROM inventory_alerts
		WHERE client_id = $1 AND shop_id = $2
		ORDER BY created_at DESC
		LIMIT $3 OFFSET $4
	`

	rows, err := r.pool.Query(ctx, query, clientID, shopID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list inventory alerts: %w", err)
	}
	defer rows.Close()

	return r.scanAlerts(rows)
}

// ListTriggered retrieves inventory alerts that have been triggered (quantity <= threshold)
func (r *inventoryAlertRepository) ListTriggered(ctx context.Context, clientID uuid.UUID) ([]*models.InventoryAlert, error) {
	query := `
		SELECT ia.id, ia.client_id, ia.variant_id, ia.shop_id, ia.reorder_point,
		       ia.reorder_quantity, ia.low_stock_threshold, ia.is_enabled,
		       ia.last_alert_sent_at, ia.metadata, ia.created_at, ia.updated_at
		FROM inventory_alerts ia
		INNER JOIN product_variants pv ON ia.client_id = pv.client_id AND ia.variant_id = pv.id
		WHERE ia.client_id = $1
		  AND ia.is_enabled = true
		  AND pv.quantity <= ia.low_stock_threshold
		ORDER BY pv.quantity ASC, ia.created_at DESC
	`

	rows, err := r.pool.Query(ctx, query, clientID)
	if err != nil {
		return nil, fmt.Errorf("failed to list triggered inventory alerts: %w", err)
	}
	defer rows.Close()

	return r.scanAlerts(rows)
}

// scanAlerts is a helper to scan inventory alert rows
func (r *inventoryAlertRepository) scanAlerts(rows pgx.Rows) ([]*models.InventoryAlert, error) {
	var alerts []*models.InventoryAlert
	for rows.Next() {
		var alert models.InventoryAlert
		err := rows.Scan(
			&alert.ID,
			&alert.ClientID,
			&alert.VariantID,
			&alert.ShopID,
			&alert.ReorderPoint,
			&alert.ReorderQuantity,
			&alert.LowStockThreshold,
			&alert.IsEnabled,
			&alert.LastAlertSentAt,
			&alert.Metadata,
			&alert.CreatedAt,
			&alert.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan inventory alert: %w", err)
		}
		alerts = append(alerts, &alert)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating inventory alerts: %w", err)
	}

	return alerts, nil
}
