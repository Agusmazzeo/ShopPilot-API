package repositories

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/yourorg/shoppilot/internal/models"
)

// CustomerRepository defines the interface for customer database operations
type CustomerRepository interface {
	Create(ctx context.Context, customer *models.Customer) error
	GetByID(ctx context.Context, clientID uuid.UUID, id uuid.UUID) (*models.Customer, error)
	GetByCode(ctx context.Context, clientID uuid.UUID, code string) (*models.Customer, error)
	GetByEmail(ctx context.Context, clientID uuid.UUID, email string) (*models.Customer, error)
	Update(ctx context.Context, customer *models.Customer) error
	Delete(ctx context.Context, clientID uuid.UUID, id uuid.UUID) error
	ListByClient(ctx context.Context, clientID uuid.UUID, limit, offset int) ([]*models.Customer, error)
	Search(ctx context.Context, clientID uuid.UUID, query string, limit, offset int) ([]*models.Customer, error)
}

// customerRepository handles database operations for customers
type customerRepository struct {
	*BaseRepository
}

// NewCustomerRepository creates a new customer repository
func NewCustomerRepository(pool *pgxpool.Pool) CustomerRepository {
	return &customerRepository{
		BaseRepository: NewBaseRepository(pool),
	}
}

// Create inserts a new customer into the database
func (r *customerRepository) Create(ctx context.Context, customer *models.Customer) error {
	query := `
		INSERT INTO customers (
			client_id, code, first_name, last_name, email, phone,
			shipping_address, shipping_city, shipping_state, shipping_postal_code, shipping_country,
			billing_address, billing_city, billing_state, billing_postal_code, billing_country,
			tax_id, notes, metadata, is_active
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20)
		RETURNING id, created_at, updated_at
	`

	err := r.pool.QueryRow(
		ctx,
		query,
		customer.ClientID,
		customer.Code,
		customer.FirstName,
		customer.LastName,
		customer.Email,
		customer.Phone,
		customer.ShippingAddress,
		customer.ShippingCity,
		customer.ShippingState,
		customer.ShippingPostalCode,
		customer.ShippingCountry,
		customer.BillingAddress,
		customer.BillingCity,
		customer.BillingState,
		customer.BillingPostalCode,
		customer.BillingCountry,
		customer.TaxID,
		customer.Notes,
		customer.Metadata,
		customer.IsActive,
	).Scan(&customer.ID, &customer.CreatedAt, &customer.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create customer: %w", err)
	}

	return nil
}

// GetByID retrieves a customer by ID
func (r *customerRepository) GetByID(ctx context.Context, clientID uuid.UUID, id uuid.UUID) (*models.Customer, error) {
	query := `
		SELECT id, client_id, code, first_name, last_name, email, phone,
		       shipping_address, shipping_city, shipping_state, shipping_postal_code, shipping_country,
		       billing_address, billing_city, billing_state, billing_postal_code, billing_country,
		       tax_id, notes, metadata, is_active, created_at, updated_at
		FROM customers
		WHERE client_id = $1 AND id = $2
	`

	var customer models.Customer
	err := r.pool.QueryRow(ctx, query, clientID, id).Scan(
		&customer.ID,
		&customer.ClientID,
		&customer.Code,
		&customer.FirstName,
		&customer.LastName,
		&customer.Email,
		&customer.Phone,
		&customer.ShippingAddress,
		&customer.ShippingCity,
		&customer.ShippingState,
		&customer.ShippingPostalCode,
		&customer.ShippingCountry,
		&customer.BillingAddress,
		&customer.BillingCity,
		&customer.BillingState,
		&customer.BillingPostalCode,
		&customer.BillingCountry,
		&customer.TaxID,
		&customer.Notes,
		&customer.Metadata,
		&customer.IsActive,
		&customer.CreatedAt,
		&customer.UpdatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("customer not found: %s", id)
		}
		return nil, fmt.Errorf("failed to get customer: %w", err)
	}

	return &customer, nil
}

// GetByCode retrieves a customer by code
func (r *customerRepository) GetByCode(ctx context.Context, clientID uuid.UUID, code string) (*models.Customer, error) {
	query := `
		SELECT id, client_id, code, first_name, last_name, email, phone,
		       shipping_address, shipping_city, shipping_state, shipping_postal_code, shipping_country,
		       billing_address, billing_city, billing_state, billing_postal_code, billing_country,
		       tax_id, notes, metadata, is_active, created_at, updated_at
		FROM customers
		WHERE client_id = $1 AND code = $2
	`

	var customer models.Customer
	err := r.pool.QueryRow(ctx, query, clientID, code).Scan(
		&customer.ID,
		&customer.ClientID,
		&customer.Code,
		&customer.FirstName,
		&customer.LastName,
		&customer.Email,
		&customer.Phone,
		&customer.ShippingAddress,
		&customer.ShippingCity,
		&customer.ShippingState,
		&customer.ShippingPostalCode,
		&customer.ShippingCountry,
		&customer.BillingAddress,
		&customer.BillingCity,
		&customer.BillingState,
		&customer.BillingPostalCode,
		&customer.BillingCountry,
		&customer.TaxID,
		&customer.Notes,
		&customer.Metadata,
		&customer.IsActive,
		&customer.CreatedAt,
		&customer.UpdatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("customer not found: %s", code)
		}
		return nil, fmt.Errorf("failed to get customer: %w", err)
	}

	return &customer, nil
}

// GetByEmail retrieves a customer by email
func (r *customerRepository) GetByEmail(ctx context.Context, clientID uuid.UUID, email string) (*models.Customer, error) {
	query := `
		SELECT id, client_id, code, first_name, last_name, email, phone,
		       shipping_address, shipping_city, shipping_state, shipping_postal_code, shipping_country,
		       billing_address, billing_city, billing_state, billing_postal_code, billing_country,
		       tax_id, notes, metadata, is_active, created_at, updated_at
		FROM customers
		WHERE client_id = $1 AND email = $2
	`

	var customer models.Customer
	err := r.pool.QueryRow(ctx, query, clientID, email).Scan(
		&customer.ID,
		&customer.ClientID,
		&customer.Code,
		&customer.FirstName,
		&customer.LastName,
		&customer.Email,
		&customer.Phone,
		&customer.ShippingAddress,
		&customer.ShippingCity,
		&customer.ShippingState,
		&customer.ShippingPostalCode,
		&customer.ShippingCountry,
		&customer.BillingAddress,
		&customer.BillingCity,
		&customer.BillingState,
		&customer.BillingPostalCode,
		&customer.BillingCountry,
		&customer.TaxID,
		&customer.Notes,
		&customer.Metadata,
		&customer.IsActive,
		&customer.CreatedAt,
		&customer.UpdatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("customer not found: %s", email)
		}
		return nil, fmt.Errorf("failed to get customer: %w", err)
	}

	return &customer, nil
}

// Update updates an existing customer
func (r *customerRepository) Update(ctx context.Context, customer *models.Customer) error {
	query := `
		UPDATE customers
		SET first_name = $3, last_name = $4, email = $5, phone = $6,
		    shipping_address = $7, shipping_city = $8, shipping_state = $9,
		    shipping_postal_code = $10, shipping_country = $11,
		    billing_address = $12, billing_city = $13, billing_state = $14,
		    billing_postal_code = $15, billing_country = $16,
		    tax_id = $17, notes = $18, metadata = $19, is_active = $20,
		    updated_at = NOW()
		WHERE client_id = $1 AND id = $2
	`

	result, err := r.pool.Exec(
		ctx,
		query,
		customer.ClientID,
		customer.ID,
		customer.FirstName,
		customer.LastName,
		customer.Email,
		customer.Phone,
		customer.ShippingAddress,
		customer.ShippingCity,
		customer.ShippingState,
		customer.ShippingPostalCode,
		customer.ShippingCountry,
		customer.BillingAddress,
		customer.BillingCity,
		customer.BillingState,
		customer.BillingPostalCode,
		customer.BillingCountry,
		customer.TaxID,
		customer.Notes,
		customer.Metadata,
		customer.IsActive,
	)

	if err != nil {
		return fmt.Errorf("failed to update customer: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("customer not found: %s", customer.ID)
	}

	return nil
}

// Delete removes a customer
func (r *customerRepository) Delete(ctx context.Context, clientID uuid.UUID, id uuid.UUID) error {
	query := `DELETE FROM customers WHERE client_id = $1 AND id = $2`

	result, err := r.pool.Exec(ctx, query, clientID, id)
	if err != nil {
		return fmt.Errorf("failed to delete customer: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("customer not found: %s", id)
	}

	return nil
}

// ListByClient retrieves all customers for a client with pagination
func (r *customerRepository) ListByClient(ctx context.Context, clientID uuid.UUID, limit, offset int) ([]*models.Customer, error) {
	query := `
		SELECT id, client_id, code, first_name, last_name, email, phone,
		       shipping_address, shipping_city, shipping_state, shipping_postal_code, shipping_country,
		       billing_address, billing_city, billing_state, billing_postal_code, billing_country,
		       tax_id, notes, metadata, is_active, created_at, updated_at
		FROM customers
		WHERE client_id = $1
		ORDER BY last_name, first_name
		LIMIT $2 OFFSET $3
	`

	rows, err := r.pool.Query(ctx, query, clientID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list customers: %w", err)
	}
	defer rows.Close()

	return r.scanCustomers(rows)
}

// Search searches customers by name or email
func (r *customerRepository) Search(ctx context.Context, clientID uuid.UUID, query string, limit, offset int) ([]*models.Customer, error) {
	searchQuery := `
		SELECT id, client_id, code, first_name, last_name, email, phone,
		       shipping_address, shipping_city, shipping_state, shipping_postal_code, shipping_country,
		       billing_address, billing_city, billing_state, billing_postal_code, billing_country,
		       tax_id, notes, metadata, is_active, created_at, updated_at
		FROM customers
		WHERE client_id = $1
		  AND (first_name ILIKE $2 OR last_name ILIKE $2 OR email ILIKE $2 OR code ILIKE $2)
		ORDER BY last_name, first_name
		LIMIT $3 OFFSET $4
	`

	searchPattern := "%" + query + "%"
	rows, err := r.pool.Query(ctx, searchQuery, clientID, searchPattern, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to search customers: %w", err)
	}
	defer rows.Close()

	return r.scanCustomers(rows)
}

// scanCustomers is a helper to scan customer rows
func (r *customerRepository) scanCustomers(rows pgx.Rows) ([]*models.Customer, error) {
	var customers []*models.Customer
	for rows.Next() {
		var customer models.Customer
		err := rows.Scan(
			&customer.ID,
			&customer.ClientID,
			&customer.Code,
			&customer.FirstName,
			&customer.LastName,
			&customer.Email,
			&customer.Phone,
			&customer.ShippingAddress,
			&customer.ShippingCity,
			&customer.ShippingState,
			&customer.ShippingPostalCode,
			&customer.ShippingCountry,
			&customer.BillingAddress,
			&customer.BillingCity,
			&customer.BillingState,
			&customer.BillingPostalCode,
			&customer.BillingCountry,
			&customer.TaxID,
			&customer.Notes,
			&customer.Metadata,
			&customer.IsActive,
			&customer.CreatedAt,
			&customer.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan customer: %w", err)
		}
		customers = append(customers, &customer)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating customers: %w", err)
	}

	return customers, nil
}
