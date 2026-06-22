package repositories

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/yourorg/shoppilot/app/models"
)

// ProductCategoryRepository handles database operations for product categories
type ProductCategoryRepository struct {
	*BaseRepository
}

// NewProductCategoryRepository creates a new product category repository
func NewProductCategoryRepository(pool *pgxpool.Pool) *ProductCategoryRepository {
	return &ProductCategoryRepository{
		BaseRepository: NewBaseRepository(pool),
	}
}

// Create inserts a new product category into the database
func (r *ProductCategoryRepository) Create(ctx context.Context, category *models.ProductCategory) error {
	query := `
		INSERT INTO product_categories (client_id, shop_id, name, slug, description, parent_id, display_order, is_active)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, created_at, updated_at
	`

	err := r.pool.QueryRow(
		ctx,
		query,
		category.ClientID,
		category.ShopID,
		category.Name,
		category.Slug,
		category.Description,
		category.ParentID,
		category.DisplayOrder,
		category.IsActive,
	).Scan(&category.ID, &category.CreatedAt, &category.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create product category: %w", err)
	}

	return nil
}

// GetByID retrieves a product category by ID
func (r *ProductCategoryRepository) GetByID(ctx context.Context, id int) (*models.ProductCategory, error) {
	query := `
		SELECT id, client_id, shop_id, name, slug, description, parent_id, display_order, is_active, created_at, updated_at
		FROM product_categories
		WHERE id = $1
	`

	var category models.ProductCategory
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&category.ID,
		&category.ClientID,
		&category.ShopID,
		&category.Name,
		&category.Slug,
		&category.Description,
		&category.ParentID,
		&category.DisplayOrder,
		&category.IsActive,
		&category.CreatedAt,
		&category.UpdatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("product category not found: %d", id)
		}
		return nil, fmt.Errorf("failed to get product category: %w", err)
	}

	return &category, nil
}

// ListByShopID retrieves all product categories for a specific shop
func (r *ProductCategoryRepository) ListByShopID(ctx context.Context, shopID int) ([]*models.ProductCategory, error) {
	query := `
		SELECT id, client_id, shop_id, name, slug, description, parent_id, display_order, is_active, created_at, updated_at
		FROM product_categories
		WHERE shop_id = $1
		ORDER BY display_order ASC, name ASC
	`

	rows, err := r.pool.Query(ctx, query, shopID)
	if err != nil {
		return nil, fmt.Errorf("failed to list product categories: %w", err)
	}
	defer rows.Close()

	var categories []*models.ProductCategory
	for rows.Next() {
		var category models.ProductCategory
		err := rows.Scan(
			&category.ID,
			&category.ClientID,
			&category.ShopID,
			&category.Name,
			&category.Slug,
			&category.Description,
			&category.ParentID,
			&category.DisplayOrder,
			&category.IsActive,
			&category.CreatedAt,
			&category.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan product category: %w", err)
		}
		categories = append(categories, &category)
	}

	return categories, nil
}

// ListByClientID retrieves all product categories for a specific client
func (r *ProductCategoryRepository) ListByClientID(ctx context.Context, clientID int) ([]*models.ProductCategory, error) {
	query := `
		SELECT id, client_id, shop_id, name, slug, description, parent_id, display_order, is_active, created_at, updated_at
		FROM product_categories
		WHERE client_id = $1
		ORDER BY shop_id, display_order ASC, name ASC
	`

	rows, err := r.pool.Query(ctx, query, clientID)
	if err != nil {
		return nil, fmt.Errorf("failed to list product categories: %w", err)
	}
	defer rows.Close()

	var categories []*models.ProductCategory
	for rows.Next() {
		var category models.ProductCategory
		err := rows.Scan(
			&category.ID,
			&category.ClientID,
			&category.ShopID,
			&category.Name,
			&category.Slug,
			&category.Description,
			&category.ParentID,
			&category.DisplayOrder,
			&category.IsActive,
			&category.CreatedAt,
			&category.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan product category: %w", err)
		}
		categories = append(categories, &category)
	}

	return categories, nil
}

// Update updates an existing product category
func (r *ProductCategoryRepository) Update(ctx context.Context, category *models.ProductCategory) error {
	query := `
		UPDATE product_categories
		SET name = $1, description = $2, parent_id = $3,
		    display_order = $4, is_active = $5, updated_at = NOW()
		WHERE id = $6
		RETURNING updated_at
	`

	err := r.pool.QueryRow(
		ctx,
		query,
		category.Name,
		category.Description,
		category.ParentID,
		category.DisplayOrder,
		category.IsActive,
		category.ID,
	).Scan(&category.UpdatedAt)

	if err != nil {
		if err == pgx.ErrNoRows {
			return fmt.Errorf("product category not found: %d", category.ID)
		}
		return fmt.Errorf("failed to update product category: %w", err)
	}

	return nil
}

// Delete deletes a product category by ID
func (r *ProductCategoryRepository) Delete(ctx context.Context, id int) error {
	query := `DELETE FROM product_categories WHERE id = $1`

	result, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete product category: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("product category not found: %d", id)
	}

	return nil
}
