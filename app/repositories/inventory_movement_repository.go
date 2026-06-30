package repositories

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/yourorg/shoppilot/internal/models"
)

// InventoryMovementRepository defines the interface for inventory movement database operations
type InventoryMovementRepository interface {
	Create(ctx context.Context, movement *models.InventoryMovement) error
	GetByID(ctx context.Context, clientID uuid.UUID, id uuid.UUID) (*models.InventoryMovement, error)
	ListByVariant(ctx context.Context, clientID uuid.UUID, variantID uuid.UUID, limit, offset int) ([]*models.InventoryMovement, error)
	ListByShop(ctx context.Context, clientID uuid.UUID, shopID uuid.UUID, limit, offset int) ([]*models.InventoryMovement, error)
	ListByType(ctx context.Context, clientID uuid.UUID, movementType models.InventoryMovementType, limit, offset int) ([]*models.InventoryMovement, error)
	ListByReference(ctx context.Context, clientID uuid.UUID, referenceType string, referenceID uuid.UUID) ([]*models.InventoryMovement, error)
}

// inventoryMovementRepository handles database operations for inventory movements
type inventoryMovementRepository struct {
	*BaseRepository
}

// NewInventoryMovementRepository creates a new inventory movement repository
func NewInventoryMovementRepository(pool *pgxpool.Pool) InventoryMovementRepository {
	return &inventoryMovementRepository{
		BaseRepository: NewBaseRepository(pool),
	}
}

// Create inserts a new inventory movement into the database
func (r *inventoryMovementRepository) Create(ctx context.Context, movement *models.InventoryMovement) error {
	query := `
		INSERT INTO inventory_movements (
			client_id, variant_id, shop_id, movement_type, quantity,
			previous_quantity, new_quantity, reference_type, reference_id,
			notes, performed_by
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING id, created_at
	`

	err := r.pool.QueryRow(
		ctx,
		query,
		movement.ClientID,
		movement.VariantID,
		movement.ShopID,
		movement.MovementType,
		movement.Quantity,
		movement.PreviousQuantity,
		movement.NewQuantity,
		movement.ReferenceType,
		movement.ReferenceID,
		movement.Notes,
		movement.PerformedBy,
	).Scan(&movement.ID, &movement.CreatedAt)

	if err != nil {
		return fmt.Errorf("failed to create inventory movement: %w", err)
	}

	return nil
}

// GetByID retrieves an inventory movement by ID
func (r *inventoryMovementRepository) GetByID(ctx context.Context, clientID uuid.UUID, id uuid.UUID) (*models.InventoryMovement, error) {
	query := `
		SELECT id, client_id, variant_id, shop_id, movement_type, quantity,
		       previous_quantity, new_quantity, reference_type, reference_id,
		       notes, performed_by, created_at
		FROM inventory_movements
		WHERE client_id = $1 AND id = $2
	`

	var movement models.InventoryMovement
	err := r.pool.QueryRow(ctx, query, clientID, id).Scan(
		&movement.ID,
		&movement.ClientID,
		&movement.VariantID,
		&movement.ShopID,
		&movement.MovementType,
		&movement.Quantity,
		&movement.PreviousQuantity,
		&movement.NewQuantity,
		&movement.ReferenceType,
		&movement.ReferenceID,
		&movement.Notes,
		&movement.PerformedBy,
		&movement.CreatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("inventory movement not found: %s", id)
		}
		return nil, fmt.Errorf("failed to get inventory movement: %w", err)
	}

	return &movement, nil
}

// ListByVariant retrieves inventory movements for a specific variant
func (r *inventoryMovementRepository) ListByVariant(ctx context.Context, clientID uuid.UUID, variantID uuid.UUID, limit, offset int) ([]*models.InventoryMovement, error) {
	query := `
		SELECT id, client_id, variant_id, shop_id, movement_type, quantity,
		       previous_quantity, new_quantity, reference_type, reference_id,
		       notes, performed_by, created_at
		FROM inventory_movements
		WHERE client_id = $1 AND variant_id = $2
		ORDER BY created_at DESC
		LIMIT $3 OFFSET $4
	`

	rows, err := r.pool.Query(ctx, query, clientID, variantID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list inventory movements by variant: %w", err)
	}
	defer rows.Close()

	return r.scanMovements(rows)
}

// ListByShop retrieves inventory movements for a specific shop
func (r *inventoryMovementRepository) ListByShop(ctx context.Context, clientID uuid.UUID, shopID uuid.UUID, limit, offset int) ([]*models.InventoryMovement, error) {
	query := `
		SELECT id, client_id, variant_id, shop_id, movement_type, quantity,
		       previous_quantity, new_quantity, reference_type, reference_id,
		       notes, performed_by, created_at
		FROM inventory_movements
		WHERE client_id = $1 AND shop_id = $2
		ORDER BY created_at DESC
		LIMIT $3 OFFSET $4
	`

	rows, err := r.pool.Query(ctx, query, clientID, shopID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list inventory movements by shop: %w", err)
	}
	defer rows.Close()

	return r.scanMovements(rows)
}

// ListByType retrieves inventory movements of a specific type
func (r *inventoryMovementRepository) ListByType(ctx context.Context, clientID uuid.UUID, movementType models.InventoryMovementType, limit, offset int) ([]*models.InventoryMovement, error) {
	query := `
		SELECT id, client_id, variant_id, shop_id, movement_type, quantity,
		       previous_quantity, new_quantity, reference_type, reference_id,
		       notes, performed_by, created_at
		FROM inventory_movements
		WHERE client_id = $1 AND movement_type = $2
		ORDER BY created_at DESC
		LIMIT $3 OFFSET $4
	`

	rows, err := r.pool.Query(ctx, query, clientID, movementType, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list inventory movements by type: %w", err)
	}
	defer rows.Close()

	return r.scanMovements(rows)
}

// ListByReference retrieves inventory movements for a specific reference document
func (r *inventoryMovementRepository) ListByReference(ctx context.Context, clientID uuid.UUID, referenceType string, referenceID uuid.UUID) ([]*models.InventoryMovement, error) {
	query := `
		SELECT id, client_id, variant_id, shop_id, movement_type, quantity,
		       previous_quantity, new_quantity, reference_type, reference_id,
		       notes, performed_by, created_at
		FROM inventory_movements
		WHERE client_id = $1 AND reference_type = $2 AND reference_id = $3
		ORDER BY created_at DESC
	`

	rows, err := r.pool.Query(ctx, query, clientID, referenceType, referenceID)
	if err != nil {
		return nil, fmt.Errorf("failed to list inventory movements by reference: %w", err)
	}
	defer rows.Close()

	return r.scanMovements(rows)
}

// scanMovements is a helper to scan inventory movement rows
func (r *inventoryMovementRepository) scanMovements(rows pgx.Rows) ([]*models.InventoryMovement, error) {
	var movements []*models.InventoryMovement
	for rows.Next() {
		var movement models.InventoryMovement
		err := rows.Scan(
			&movement.ID,
			&movement.ClientID,
			&movement.VariantID,
			&movement.ShopID,
			&movement.MovementType,
			&movement.Quantity,
			&movement.PreviousQuantity,
			&movement.NewQuantity,
			&movement.ReferenceType,
			&movement.ReferenceID,
			&movement.Notes,
			&movement.PerformedBy,
			&movement.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan inventory movement: %w", err)
		}
		movements = append(movements, &movement)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating inventory movements: %w", err)
	}

	return movements, nil
}
