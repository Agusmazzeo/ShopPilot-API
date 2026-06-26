package repositories

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/yourorg/shoppilot/internal/models"
)

// ClientRepository defines the interface for client database operations
type ClientRepository interface {
	// Client CRUD
	Create(ctx context.Context, client *models.Client) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.Client, error)
	GetBySlug(ctx context.Context, slug string) (*models.Client, error)
	Update(ctx context.Context, client *models.Client) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, limit, offset int) ([]*models.Client, error)
	ListActive(ctx context.Context, limit, offset int) ([]*models.Client, error)
}

// clientRepository handles database operations for clients
type clientRepository struct {
	*BaseRepository
}

// NewClientRepository creates a new client repository
func NewClientRepository(pool *pgxpool.Pool) ClientRepository {
	return &clientRepository{
		BaseRepository: NewBaseRepository(pool),
	}
}

// Create inserts a new client into the database
func (r *clientRepository) Create(ctx context.Context, client *models.Client) error {
	query := `
		INSERT INTO clients (name, slug, description, contact_email, contact_phone, website_url, logo_url, is_active)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, created_at, updated_at
	`

	err := r.pool.QueryRow(
		ctx,
		query,
		client.Name,
		client.Slug,
		client.Description,
		client.ContactEmail,
		client.ContactPhone,
		client.WebsiteURL,
		client.LogoURL,
		client.IsActive,
	).Scan(&client.ID, &client.CreatedAt, &client.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}

	return nil
}

// GetByID retrieves a client by ID
func (r *clientRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Client, error) {
	query := `
		SELECT id, name, slug, description, contact_email, contact_phone, website_url, logo_url, is_active, created_at, updated_at
		FROM clients
		WHERE id = $1
	`

	var client models.Client
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&client.ID,
		&client.Name,
		&client.Slug,
		&client.Description,
		&client.ContactEmail,
		&client.ContactPhone,
		&client.WebsiteURL,
		&client.LogoURL,
		&client.IsActive,
		&client.CreatedAt,
		&client.UpdatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("client not found: %s", id)
		}
		return nil, fmt.Errorf("failed to get client: %w", err)
	}

	return &client, nil
}

// GetBySlug retrieves a client by slug
func (r *clientRepository) GetBySlug(ctx context.Context, slug string) (*models.Client, error) {
	query := `
		SELECT id, name, slug, description, contact_email, contact_phone, website_url, logo_url, is_active, created_at, updated_at
		FROM clients
		WHERE slug = $1
	`

	var client models.Client
	err := r.pool.QueryRow(ctx, query, slug).Scan(
		&client.ID,
		&client.Name,
		&client.Slug,
		&client.Description,
		&client.ContactEmail,
		&client.ContactPhone,
		&client.WebsiteURL,
		&client.LogoURL,
		&client.IsActive,
		&client.CreatedAt,
		&client.UpdatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("client not found: %s", slug)
		}
		return nil, fmt.Errorf("failed to get client: %w", err)
	}

	return &client, nil
}

// List retrieves all clients with pagination
func (r *clientRepository) List(ctx context.Context, limit, offset int) ([]*models.Client, error) {
	query := `
		SELECT id, name, slug, description, contact_email, contact_phone, website_url, logo_url, is_active, created_at, updated_at
		FROM clients
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := r.pool.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list clients: %w", err)
	}
	defer rows.Close()

	var clients []*models.Client
	for rows.Next() {
		var client models.Client
		err := rows.Scan(
			&client.ID,
			&client.Name,
			&client.Slug,
			&client.Description,
			&client.ContactEmail,
			&client.ContactPhone,
			&client.WebsiteURL,
			&client.LogoURL,
			&client.IsActive,
			&client.CreatedAt,
			&client.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan client: %w", err)
		}
		clients = append(clients, &client)
	}

	return clients, nil
}

// ListActive retrieves only active clients with pagination
func (r *clientRepository) ListActive(ctx context.Context, limit, offset int) ([]*models.Client, error) {
	query := `
		SELECT id, name, slug, description, contact_email, contact_phone, website_url, logo_url, is_active, created_at, updated_at
		FROM clients
		WHERE is_active = true
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := r.pool.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list active clients: %w", err)
	}
	defer rows.Close()

	var clients []*models.Client
	for rows.Next() {
		var client models.Client
		err := rows.Scan(
			&client.ID,
			&client.Name,
			&client.Slug,
			&client.Description,
			&client.ContactEmail,
			&client.ContactPhone,
			&client.WebsiteURL,
			&client.LogoURL,
			&client.IsActive,
			&client.CreatedAt,
			&client.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan client: %w", err)
		}
		clients = append(clients, &client)
	}

	return clients, nil
}

// Update updates an existing client
func (r *clientRepository) Update(ctx context.Context, client *models.Client) error {
	query := `
		UPDATE clients
		SET name = $1, slug = $2, description = $3, contact_email = $4,
		    contact_phone = $5, website_url = $6, logo_url = $7,
		    is_active = $8, updated_at = NOW()
		WHERE id = $9
		RETURNING updated_at
	`

	err := r.pool.QueryRow(
		ctx,
		query,
		client.Name,
		client.Slug,
		client.Description,
		client.ContactEmail,
		client.ContactPhone,
		client.WebsiteURL,
		client.LogoURL,
		client.IsActive,
		client.ID,
	).Scan(&client.UpdatedAt)

	if err != nil {
		if err == pgx.ErrNoRows {
			return fmt.Errorf("client not found: %s", client.ID)
		}
		return fmt.Errorf("failed to update client: %w", err)
	}

	return nil
}

// Delete deletes a client by ID
func (r *clientRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM clients WHERE id = $1`

	result, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete client: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("client not found: %s", id)
	}

	return nil
}
