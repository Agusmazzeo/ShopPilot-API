package repositories

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/yourorg/shoppilot/internal/models"
)

// SupplierRepository defines the interface for supplier database operations
type SupplierRepository interface {
	Create(ctx context.Context, supplier *models.Supplier) error
	GetByID(ctx context.Context, clientID uuid.UUID, id uuid.UUID) (*models.Supplier, error)
	GetByCode(ctx context.Context, clientID uuid.UUID, code string) (*models.Supplier, error)
	Update(ctx context.Context, supplier *models.Supplier) error
	Delete(ctx context.Context, clientID uuid.UUID, id uuid.UUID) error
	ListByClient(ctx context.Context, clientID uuid.UUID, limit, offset int) ([]*models.Supplier, error)
	ListActive(ctx context.Context, clientID uuid.UUID, limit, offset int) ([]*models.Supplier, error)
}

// supplierRepository handles database operations for suppliers
type supplierRepository struct {
	*BaseRepository
}

// NewSupplierRepository creates a new supplier repository
func NewSupplierRepository(pool *pgxpool.Pool) SupplierRepository {
	return &supplierRepository{
		BaseRepository: NewBaseRepository(pool),
	}
}

// Create inserts a new supplier into the database
func (r *supplierRepository) Create(ctx context.Context, supplier *models.Supplier) error {
	query := `
		INSERT INTO suppliers (
			client_id, code, name, email, phone, address, city, state,
			postal_code, country, tax_id, payment_terms, currency, notes,
			metadata, is_active
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)
		RETURNING id, created_at, updated_at
	`

	err := r.pool.QueryRow(
		ctx,
		query,
		supplier.ClientID,
		supplier.Code,
		supplier.Name,
		supplier.Email,
		supplier.Phone,
		supplier.Address,
		supplier.City,
		supplier.State,
		supplier.PostalCode,
		supplier.Country,
		supplier.TaxID,
		supplier.PaymentTerms,
		supplier.Currency,
		supplier.Notes,
		supplier.Metadata,
		supplier.IsActive,
	).Scan(&supplier.ID, &supplier.CreatedAt, &supplier.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create supplier: %w", err)
	}

	return nil
}

// GetByID retrieves a supplier by ID
func (r *supplierRepository) GetByID(ctx context.Context, clientID uuid.UUID, id uuid.UUID) (*models.Supplier, error) {
	query := `
		SELECT id, client_id, code, name, email, phone, address, city, state,
		       postal_code, country, tax_id, payment_terms, currency, notes,
		       metadata, is_active, created_at, updated_at
		FROM suppliers
		WHERE client_id = $1 AND id = $2
	`

	var supplier models.Supplier
	err := r.pool.QueryRow(ctx, query, clientID, id).Scan(
		&supplier.ID,
		&supplier.ClientID,
		&supplier.Code,
		&supplier.Name,
		&supplier.Email,
		&supplier.Phone,
		&supplier.Address,
		&supplier.City,
		&supplier.State,
		&supplier.PostalCode,
		&supplier.Country,
		&supplier.TaxID,
		&supplier.PaymentTerms,
		&supplier.Currency,
		&supplier.Notes,
		&supplier.Metadata,
		&supplier.IsActive,
		&supplier.CreatedAt,
		&supplier.UpdatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("supplier not found: %s", id)
		}
		return nil, fmt.Errorf("failed to get supplier: %w", err)
	}

	return &supplier, nil
}

// GetByCode retrieves a supplier by code
func (r *supplierRepository) GetByCode(ctx context.Context, clientID uuid.UUID, code string) (*models.Supplier, error) {
	query := `
		SELECT id, client_id, code, name, email, phone, address, city, state,
		       postal_code, country, tax_id, payment_terms, currency, notes,
		       metadata, is_active, created_at, updated_at
		FROM suppliers
		WHERE client_id = $1 AND code = $2
	`

	var supplier models.Supplier
	err := r.pool.QueryRow(ctx, query, clientID, code).Scan(
		&supplier.ID,
		&supplier.ClientID,
		&supplier.Code,
		&supplier.Name,
		&supplier.Email,
		&supplier.Phone,
		&supplier.Address,
		&supplier.City,
		&supplier.State,
		&supplier.PostalCode,
		&supplier.Country,
		&supplier.TaxID,
		&supplier.PaymentTerms,
		&supplier.Currency,
		&supplier.Notes,
		&supplier.Metadata,
		&supplier.IsActive,
		&supplier.CreatedAt,
		&supplier.UpdatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("supplier not found: %s", code)
		}
		return nil, fmt.Errorf("failed to get supplier: %w", err)
	}

	return &supplier, nil
}

// Update updates an existing supplier
func (r *supplierRepository) Update(ctx context.Context, supplier *models.Supplier) error {
	query := `
		UPDATE suppliers
		SET name = $3, email = $4, phone = $5, address = $6, city = $7, state = $8,
		    postal_code = $9, country = $10, tax_id = $11, payment_terms = $12,
		    currency = $13, notes = $14, metadata = $15, is_active = $16,
		    updated_at = NOW()
		WHERE client_id = $1 AND id = $2
	`

	result, err := r.pool.Exec(
		ctx,
		query,
		supplier.ClientID,
		supplier.ID,
		supplier.Name,
		supplier.Email,
		supplier.Phone,
		supplier.Address,
		supplier.City,
		supplier.State,
		supplier.PostalCode,
		supplier.Country,
		supplier.TaxID,
		supplier.PaymentTerms,
		supplier.Currency,
		supplier.Notes,
		supplier.Metadata,
		supplier.IsActive,
	)

	if err != nil {
		return fmt.Errorf("failed to update supplier: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("supplier not found: %s", supplier.ID)
	}

	return nil
}

// Delete removes a supplier
func (r *supplierRepository) Delete(ctx context.Context, clientID uuid.UUID, id uuid.UUID) error {
	query := `DELETE FROM suppliers WHERE client_id = $1 AND id = $2`

	result, err := r.pool.Exec(ctx, query, clientID, id)
	if err != nil {
		return fmt.Errorf("failed to delete supplier: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("supplier not found: %s", id)
	}

	return nil
}

// ListByClient retrieves all suppliers for a client with pagination
func (r *supplierRepository) ListByClient(ctx context.Context, clientID uuid.UUID, limit, offset int) ([]*models.Supplier, error) {
	query := `
		SELECT id, client_id, code, name, email, phone, address, city, state,
		       postal_code, country, tax_id, payment_terms, currency, notes,
		       metadata, is_active, created_at, updated_at
		FROM suppliers
		WHERE client_id = $1
		ORDER BY name
		LIMIT $2 OFFSET $3
	`

	rows, err := r.pool.Query(ctx, query, clientID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list suppliers: %w", err)
	}
	defer rows.Close()

	var suppliers []*models.Supplier
	for rows.Next() {
		var supplier models.Supplier
		err := rows.Scan(
			&supplier.ID,
			&supplier.ClientID,
			&supplier.Code,
			&supplier.Name,
			&supplier.Email,
			&supplier.Phone,
			&supplier.Address,
			&supplier.City,
			&supplier.State,
			&supplier.PostalCode,
			&supplier.Country,
			&supplier.TaxID,
			&supplier.PaymentTerms,
			&supplier.Currency,
			&supplier.Notes,
			&supplier.Metadata,
			&supplier.IsActive,
			&supplier.CreatedAt,
			&supplier.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan supplier: %w", err)
		}
		suppliers = append(suppliers, &supplier)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating suppliers: %w", err)
	}

	return suppliers, nil
}

// ListActive retrieves active suppliers for a client with pagination
func (r *supplierRepository) ListActive(ctx context.Context, clientID uuid.UUID, limit, offset int) ([]*models.Supplier, error) {
	query := `
		SELECT id, client_id, code, name, email, phone, address, city, state,
		       postal_code, country, tax_id, payment_terms, currency, notes,
		       metadata, is_active, created_at, updated_at
		FROM suppliers
		WHERE client_id = $1 AND is_active = true
		ORDER BY name
		LIMIT $2 OFFSET $3
	`

	rows, err := r.pool.Query(ctx, query, clientID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list active suppliers: %w", err)
	}
	defer rows.Close()

	var suppliers []*models.Supplier
	for rows.Next() {
		var supplier models.Supplier
		err := rows.Scan(
			&supplier.ID,
			&supplier.ClientID,
			&supplier.Code,
			&supplier.Name,
			&supplier.Email,
			&supplier.Phone,
			&supplier.Address,
			&supplier.City,
			&supplier.State,
			&supplier.PostalCode,
			&supplier.Country,
			&supplier.TaxID,
			&supplier.PaymentTerms,
			&supplier.Currency,
			&supplier.Notes,
			&supplier.Metadata,
			&supplier.IsActive,
			&supplier.CreatedAt,
			&supplier.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan supplier: %w", err)
		}
		suppliers = append(suppliers, &supplier)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating active suppliers: %w", err)
	}

	return suppliers, nil
}
