package repositories

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/yourorg/shoppilot/app/models"
)

// ShopRepository handles database operations for shops
type ShopRepository struct {
	*BaseRepository
}

// NewShopRepository creates a new shop repository
func NewShopRepository(pool *pgxpool.Pool) *ShopRepository {
	return &ShopRepository{
		BaseRepository: NewBaseRepository(pool),
	}
}

// Create inserts a new shop into the database
func (r *ShopRepository) Create(ctx context.Context, shop *models.Shop) error {
	query := `
		INSERT INTO shops (client_id, user_id, name, slug, description, domain, is_active)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, created_at, updated_at
	`

	err := r.pool.QueryRow(
		ctx,
		query,
		shop.ClientID,
		shop.UserID,
		shop.Name,
		shop.Slug,
		shop.Description,
		shop.Domain,
		shop.IsActive,
	).Scan(&shop.ID, &shop.CreatedAt, &shop.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create shop: %w", err)
	}

	return nil
}

// GetByID retrieves a shop by ID
func (r *ShopRepository) GetByID(ctx context.Context, id int) (*models.Shop, error) {
	query := `
		SELECT id, client_id, user_id, name, slug, description, domain, is_active, created_at, updated_at
		FROM shops
		WHERE id = $1
	`

	var shop models.Shop
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&shop.ID,
		&shop.ClientID,
		&shop.UserID,
		&shop.Name,
		&shop.Slug,
		&shop.Description,
		&shop.Domain,
		&shop.IsActive,
		&shop.CreatedAt,
		&shop.UpdatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("shop not found: %d", id)
		}
		return nil, fmt.Errorf("failed to get shop: %w", err)
	}

	return &shop, nil
}

// GetBySlug retrieves a shop by slug within a client scope
func (r *ShopRepository) GetBySlug(ctx context.Context, clientID int, slug string) (*models.Shop, error) {
	query := `
		SELECT id, client_id, user_id, name, slug, description, domain, is_active, created_at, updated_at
		FROM shops
		WHERE client_id = $1 AND slug = $2
	`

	var shop models.Shop
	err := r.pool.QueryRow(ctx, query, clientID, slug).Scan(
		&shop.ID,
		&shop.ClientID,
		&shop.UserID,
		&shop.Name,
		&shop.Slug,
		&shop.Description,
		&shop.Domain,
		&shop.IsActive,
		&shop.CreatedAt,
		&shop.UpdatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("shop not found: %s", slug)
		}
		return nil, fmt.Errorf("failed to get shop: %w", err)
	}

	return &shop, nil
}

// ListByClientID retrieves all shops for a specific client
func (r *ShopRepository) ListByClientID(ctx context.Context, clientID int) ([]*models.Shop, error) {
	query := `
		SELECT id, client_id, user_id, name, slug, description, domain, is_active, created_at, updated_at
		FROM shops
		WHERE client_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.pool.Query(ctx, query, clientID)
	if err != nil {
		return nil, fmt.Errorf("failed to list shops: %w", err)
	}
	defer rows.Close()

	var shops []*models.Shop
	for rows.Next() {
		var shop models.Shop
		err := rows.Scan(
			&shop.ID,
			&shop.ClientID,
			&shop.UserID,
			&shop.Name,
			&shop.Slug,
			&shop.Description,
			&shop.Domain,
			&shop.IsActive,
			&shop.CreatedAt,
			&shop.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan shop: %w", err)
		}
		shops = append(shops, &shop)
	}

	return shops, nil
}

// ListByUserID retrieves all shops for a specific user
func (r *ShopRepository) ListByUserID(ctx context.Context, userID int) ([]*models.Shop, error) {
	query := `
		SELECT id, client_id, user_id, name, slug, description, domain, is_active, created_at, updated_at
		FROM shops
		WHERE user_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to list shops: %w", err)
	}
	defer rows.Close()

	var shops []*models.Shop
	for rows.Next() {
		var shop models.Shop
		err := rows.Scan(
			&shop.ID,
			&shop.ClientID,
			&shop.UserID,
			&shop.Name,
			&shop.Slug,
			&shop.Description,
			&shop.Domain,
			&shop.IsActive,
			&shop.CreatedAt,
			&shop.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan shop: %w", err)
		}
		shops = append(shops, &shop)
	}

	return shops, nil
}

// Update updates an existing shop
func (r *ShopRepository) Update(ctx context.Context, shop *models.Shop) error {
	query := `
		UPDATE shops
		SET name = $1, description = $2, domain = $3,
		    is_active = $4, updated_at = NOW()
		WHERE id = $5
		RETURNING updated_at
	`

	err := r.pool.QueryRow(
		ctx,
		query,
		shop.Name,
		shop.Description,
		shop.Domain,
		shop.IsActive,
		shop.ID,
	).Scan(&shop.UpdatedAt)

	if err != nil {
		if err == pgx.ErrNoRows {
			return fmt.Errorf("shop not found: %d", shop.ID)
		}
		return fmt.Errorf("failed to update shop: %w", err)
	}

	return nil
}

// Delete deletes a shop by ID
func (r *ShopRepository) Delete(ctx context.Context, id int) error {
	query := `DELETE FROM shops WHERE id = $1`

	result, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete shop: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("shop not found: %d", id)
	}

	return nil
}
