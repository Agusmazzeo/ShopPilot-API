package repositories

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/yourorg/shoppilot/app/models"
)

// ClientRepository handles database operations for clients
type ClientRepository struct {
	*BaseRepository
}

// NewClientRepository creates a new client repository
func NewClientRepository(pool *pgxpool.Pool) *ClientRepository {
	return &ClientRepository{
		BaseRepository: NewBaseRepository(pool),
	}
}

// Create inserts a new client into the database
func (r *ClientRepository) Create(ctx context.Context, client *models.Client) error {
	query := `
		INSERT INTO clients (name, slug, contact_email, contact_phone, subscription_tier, is_active)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at, updated_at
	`

	err := r.pool.QueryRow(
		ctx,
		query,
		client.Name,
		client.Slug,
		client.ContactEmail,
		client.ContactPhone,
		client.SubscriptionTier,
		client.IsActive,
	).Scan(&client.ID, &client.CreatedAt, &client.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}

	return nil
}

// GetByID retrieves a client by ID
func (r *ClientRepository) GetByID(ctx context.Context, id int) (*models.Client, error) {
	query := `
		SELECT id, name, slug, contact_email, contact_phone, subscription_tier, is_active, created_at, updated_at
		FROM clients
		WHERE id = $1
	`

	var client models.Client
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&client.ID,
		&client.Name,
		&client.Slug,
		&client.ContactEmail,
		&client.ContactPhone,
		&client.SubscriptionTier,
		&client.IsActive,
		&client.CreatedAt,
		&client.UpdatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("client not found: %d", id)
		}
		return nil, fmt.Errorf("failed to get client: %w", err)
	}

	return &client, nil
}

// GetBySlug retrieves a client by slug
func (r *ClientRepository) GetBySlug(ctx context.Context, slug string) (*models.Client, error) {
	query := `
		SELECT id, name, slug, contact_email, contact_phone, subscription_tier, is_active, created_at, updated_at
		FROM clients
		WHERE slug = $1
	`

	var client models.Client
	err := r.pool.QueryRow(ctx, query, slug).Scan(
		&client.ID,
		&client.Name,
		&client.Slug,
		&client.ContactEmail,
		&client.ContactPhone,
		&client.SubscriptionTier,
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

// List retrieves all clients
func (r *ClientRepository) List(ctx context.Context) ([]*models.Client, error) {
	query := `
		SELECT id, name, slug, contact_email, contact_phone, subscription_tier, is_active, created_at, updated_at
		FROM clients
		ORDER BY created_at DESC
	`

	rows, err := r.pool.Query(ctx, query)
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
			&client.ContactEmail,
			&client.ContactPhone,
			&client.SubscriptionTier,
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
func (r *ClientRepository) Update(ctx context.Context, client *models.Client) error {
	query := `
		UPDATE clients
		SET name = $1, contact_email = $2, contact_phone = $3,
		    subscription_tier = $4, is_active = $5, updated_at = NOW()
		WHERE id = $6
		RETURNING updated_at
	`

	err := r.pool.QueryRow(
		ctx,
		query,
		client.Name,
		client.ContactEmail,
		client.ContactPhone,
		client.SubscriptionTier,
		client.IsActive,
		client.ID,
	).Scan(&client.UpdatedAt)

	if err != nil {
		if err == pgx.ErrNoRows {
			return fmt.Errorf("client not found: %d", client.ID)
		}
		return fmt.Errorf("failed to update client: %w", err)
	}

	return nil
}

// Delete deletes a client by ID
func (r *ClientRepository) Delete(ctx context.Context, id int) error {
	query := `DELETE FROM clients WHERE id = $1`

	result, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete client: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("client not found: %d", id)
	}

	return nil
}
