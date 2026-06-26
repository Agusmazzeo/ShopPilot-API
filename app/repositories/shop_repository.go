package repositories

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/yourorg/shoppilot/internal/models"
)

// ShopRepository defines the interface for shop data operations
type ShopRepository interface {
	// Shop CRUD
	Create(ctx context.Context, shop *models.Shop) error
	GetByID(ctx context.Context, clientID uuid.UUID, shopID uuid.UUID) (*models.Shop, error)
	GetBySlug(ctx context.Context, clientID uuid.UUID, slug string) (*models.Shop, error)
	Update(ctx context.Context, shop *models.Shop) error
	Delete(ctx context.Context, clientID uuid.UUID, shopID uuid.UUID) error
	ListByClient(ctx context.Context, clientID uuid.UUID, limit, offset int) ([]*models.Shop, error)
	ListActiveByClient(ctx context.Context, clientID uuid.UUID, limit, offset int) ([]*models.Shop, error)

	// Shop user assignments
	AssignUser(ctx context.Context, shopID uuid.UUID, clientUserRoleID int) error
	RemoveUser(ctx context.Context, shopID uuid.UUID, clientUserRoleID int) error
	GetShopUsers(ctx context.Context, shopID uuid.UUID) ([]*models.ShopUser, error)
}

// shopRepository implements ShopRepository interface
type shopRepository struct {
	*BaseRepository
}

// NewShopRepository creates a new shop repository
func NewShopRepository(pool *pgxpool.Pool) ShopRepository {
	return &shopRepository{
		BaseRepository: NewBaseRepository(pool),
	}
}

// Create inserts a new shop into the database
func (r *shopRepository) Create(ctx context.Context, shop *models.Shop) error {
	query := `
		INSERT INTO shops (
			client_id, name, slug, description, webpage_url,
			address, city, state, country, postal_code,
			phone, email, logo_url, is_active
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
		RETURNING id, created_at, updated_at
	`

	err := r.pool.QueryRow(
		ctx,
		query,
		shop.ClientID,
		shop.Name,
		shop.Slug,
		shop.Description,
		shop.WebpageURL,
		shop.Address,
		shop.City,
		shop.State,
		shop.Country,
		shop.PostalCode,
		shop.Phone,
		shop.Email,
		NullableString(shop.LogoURL),
		shop.IsActive,
	).Scan(&shop.ID, &shop.CreatedAt, &shop.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create shop: %w", err)
	}

	return nil
}

// GetByID retrieves a shop by composite key (client_id, id)
func (r *shopRepository) GetByID(ctx context.Context, clientID uuid.UUID, shopID uuid.UUID) (*models.Shop, error) {
	query := `
		SELECT
			id, client_id, name, slug, description, webpage_url,
			address, city, state, country, postal_code,
			phone, email, logo_url, is_active, created_at, updated_at
		FROM shops
		WHERE client_id = $1 AND id = $2
	`

	var shop models.Shop
	err := r.pool.QueryRow(ctx, query, clientID, shopID).Scan(
		&shop.ID,
		&shop.ClientID,
		&shop.Name,
		&shop.Slug,
		&shop.Description,
		&shop.WebpageURL,
		&shop.Address,
		&shop.City,
		&shop.State,
		&shop.Country,
		&shop.PostalCode,
		&shop.Phone,
		&shop.Email,
		&shop.LogoURL,
		&shop.IsActive,
		&shop.CreatedAt,
		&shop.UpdatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("shop not found: client_id=%s, shop_id=%s", clientID, shopID)
		}
		return nil, fmt.Errorf("failed to get shop: %w", err)
	}

	return &shop, nil
}

// GetBySlug retrieves a shop by slug (unique per client)
func (r *shopRepository) GetBySlug(ctx context.Context, clientID uuid.UUID, slug string) (*models.Shop, error) {
	query := `
		SELECT
			id, client_id, name, slug, description, webpage_url,
			address, city, state, country, postal_code,
			phone, email, logo_url, is_active, created_at, updated_at
		FROM shops
		WHERE client_id = $1 AND slug = $2
	`

	var shop models.Shop
	err := r.pool.QueryRow(ctx, query, clientID, slug).Scan(
		&shop.ID,
		&shop.ClientID,
		&shop.Name,
		&shop.Slug,
		&shop.Description,
		&shop.WebpageURL,
		&shop.Address,
		&shop.City,
		&shop.State,
		&shop.Country,
		&shop.PostalCode,
		&shop.Phone,
		&shop.Email,
		&shop.LogoURL,
		&shop.IsActive,
		&shop.CreatedAt,
		&shop.UpdatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("shop not found: client_id=%s, slug=%s", clientID, slug)
		}
		return nil, fmt.Errorf("failed to get shop by slug: %w", err)
	}

	return &shop, nil
}

// Update updates an existing shop
func (r *shopRepository) Update(ctx context.Context, shop *models.Shop) error {
	query := `
		UPDATE shops
		SET
			name = $1,
			slug = $2,
			description = $3,
			webpage_url = $4,
			address = $5,
			city = $6,
			state = $7,
			country = $8,
			postal_code = $9,
			phone = $10,
			email = $11,
			logo_url = $12,
			is_active = $13,
			updated_at = NOW()
		WHERE client_id = $14 AND id = $15
		RETURNING updated_at
	`

	err := r.pool.QueryRow(
		ctx,
		query,
		shop.Name,
		shop.Slug,
		shop.Description,
		shop.WebpageURL,
		shop.Address,
		shop.City,
		shop.State,
		shop.Country,
		shop.PostalCode,
		shop.Phone,
		shop.Email,
		NullableString(shop.LogoURL),
		shop.IsActive,
		shop.ClientID,
		shop.ID,
	).Scan(&shop.UpdatedAt)

	if err != nil {
		if err == pgx.ErrNoRows {
			return fmt.Errorf("shop not found: client_id=%s, shop_id=%s", shop.ClientID, shop.ID)
		}
		return fmt.Errorf("failed to update shop: %w", err)
	}

	return nil
}

// Delete deletes a shop from partition by composite key
func (r *shopRepository) Delete(ctx context.Context, clientID uuid.UUID, shopID uuid.UUID) error {
	query := `DELETE FROM shops WHERE client_id = $1 AND id = $2`

	result, err := r.pool.Exec(ctx, query, clientID, shopID)
	if err != nil {
		return fmt.Errorf("failed to delete shop: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("shop not found: client_id=%s, shop_id=%s", clientID, shopID)
	}

	return nil
}

// ListByClient retrieves all shops for a client with pagination
func (r *shopRepository) ListByClient(ctx context.Context, clientID uuid.UUID, limit, offset int) ([]*models.Shop, error) {
	query := `
		SELECT
			id, client_id, name, slug, description, webpage_url,
			address, city, state, country, postal_code,
			phone, email, logo_url, is_active, created_at, updated_at
		FROM shops
		WHERE client_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.pool.Query(ctx, query, clientID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list shops by client: %w", err)
	}
	defer rows.Close()

	var shops []*models.Shop
	for rows.Next() {
		var shop models.Shop
		err := rows.Scan(
			&shop.ID,
			&shop.ClientID,
			&shop.Name,
			&shop.Slug,
			&shop.Description,
			&shop.WebpageURL,
			&shop.Address,
			&shop.City,
			&shop.State,
			&shop.Country,
			&shop.PostalCode,
			&shop.Phone,
			&shop.Email,
			&shop.LogoURL,
			&shop.IsActive,
			&shop.CreatedAt,
			&shop.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan shop: %w", err)
		}
		shops = append(shops, &shop)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating shops: %w", err)
	}

	return shops, nil
}

// ListActiveByClient retrieves active shops for a client with pagination
func (r *shopRepository) ListActiveByClient(ctx context.Context, clientID uuid.UUID, limit, offset int) ([]*models.Shop, error) {
	query := `
		SELECT
			id, client_id, name, slug, description, webpage_url,
			address, city, state, country, postal_code,
			phone, email, logo_url, is_active, created_at, updated_at
		FROM shops
		WHERE client_id = $1 AND is_active = true
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.pool.Query(ctx, query, clientID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list active shops by client: %w", err)
	}
	defer rows.Close()

	var shops []*models.Shop
	for rows.Next() {
		var shop models.Shop
		err := rows.Scan(
			&shop.ID,
			&shop.ClientID,
			&shop.Name,
			&shop.Slug,
			&shop.Description,
			&shop.WebpageURL,
			&shop.Address,
			&shop.City,
			&shop.State,
			&shop.Country,
			&shop.PostalCode,
			&shop.Phone,
			&shop.Email,
			&shop.LogoURL,
			&shop.IsActive,
			&shop.CreatedAt,
			&shop.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan shop: %w", err)
		}
		shops = append(shops, &shop)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating active shops: %w", err)
	}

	return shops, nil
}

// AssignUser assigns a user to a shop via client_user_role
func (r *shopRepository) AssignUser(ctx context.Context, shopID uuid.UUID, clientUserRoleID int) error {
	// First, get the client_id from the shop
	var clientID uuid.UUID
	err := r.pool.QueryRow(ctx, `SELECT client_id FROM shops WHERE id = $1`, shopID).Scan(&clientID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return fmt.Errorf("shop not found: %s", shopID)
		}
		return fmt.Errorf("failed to get shop client_id: %w", err)
	}

	query := `
		INSERT INTO shop_users (client_id, shop_id, client_user_role_id)
		VALUES ($1, $2, $3)
		ON CONFLICT (shop_id, client_user_role_id) DO NOTHING
	`

	_, err = r.pool.Exec(ctx, query, clientID, shopID, clientUserRoleID)
	if err != nil {
		return fmt.Errorf("failed to assign user to shop: %w", err)
	}

	return nil
}

// RemoveUser removes a user from a shop
func (r *shopRepository) RemoveUser(ctx context.Context, shopID uuid.UUID, clientUserRoleID int) error {
	query := `
		DELETE FROM shop_users
		WHERE shop_id = $1 AND client_user_role_id = $2
	`

	result, err := r.pool.Exec(ctx, query, shopID, clientUserRoleID)
	if err != nil {
		return fmt.Errorf("failed to remove user from shop: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("shop user not found: shop_id=%s, client_user_role_id=%d", shopID, clientUserRoleID)
	}

	return nil
}

// GetShopUsers retrieves all users assigned to a shop
func (r *shopRepository) GetShopUsers(ctx context.Context, shopID uuid.UUID) ([]*models.ShopUser, error) {
	query := `
		SELECT id, client_id, shop_id, client_user_role_id, created_at, updated_at
		FROM shop_users
		WHERE shop_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.pool.Query(ctx, query, shopID)
	if err != nil {
		return nil, fmt.Errorf("failed to get shop users: %w", err)
	}
	defer rows.Close()

	var shopUsers []*models.ShopUser
	for rows.Next() {
		var shopUser models.ShopUser
		err := rows.Scan(
			&shopUser.ID,
			&shopUser.ClientID,
			&shopUser.ShopID,
			&shopUser.ClientUserRoleID,
			&shopUser.CreatedAt,
			&shopUser.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan shop user: %w", err)
		}
		shopUsers = append(shopUsers, &shopUser)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating shop users: %w", err)
	}

	return shopUsers, nil
}
