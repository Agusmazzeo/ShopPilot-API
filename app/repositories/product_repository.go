package repositories

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/yourorg/shoppilot/app/models"
)

// ProductRepository handles database operations for products
type ProductRepository struct {
	*BaseRepository
}

// NewProductRepository creates a new product repository
func NewProductRepository(pool *pgxpool.Pool) *ProductRepository {
	return &ProductRepository{
		BaseRepository: NewBaseRepository(pool),
	}
}

// Create inserts a new product into the database
func (r *ProductRepository) Create(ctx context.Context, product *models.Product) error {
	query := `
		INSERT INTO products (
			client_id, shop_id, category_id, sku, name, slug, description,
			short_description, price, compare_at_price, cost_per_item,
			weight, weight_unit, requires_shipping, is_active
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
		RETURNING id, created_at, updated_at
	`

	err := r.pool.QueryRow(
		ctx,
		query,
		product.ClientID,
		product.ShopID,
		product.CategoryID,
		product.SKU,
		product.Name,
		product.Slug,
		product.Description,
		product.ShortDescription,
		product.Price,
		product.CompareAtPrice,
		product.CostPerItem,
		product.Weight,
		product.WeightUnit,
		product.RequiresShipping,
		product.IsActive,
	).Scan(&product.ID, &product.CreatedAt, &product.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create product: %w", err)
	}

	return nil
}

// GetByID retrieves a product by ID
func (r *ProductRepository) GetByID(ctx context.Context, id int) (*models.Product, error) {
	query := `
		SELECT id, client_id, shop_id, category_id, sku, name, slug, description,
		       short_description, price, compare_at_price, cost_per_item,
		       weight, weight_unit, requires_shipping, is_active, created_at, updated_at
		FROM products
		WHERE id = $1
	`

	var product models.Product
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&product.ID,
		&product.ClientID,
		&product.ShopID,
		&product.CategoryID,
		&product.SKU,
		&product.Name,
		&product.Slug,
		&product.Description,
		&product.ShortDescription,
		&product.Price,
		&product.CompareAtPrice,
		&product.CostPerItem,
		&product.Weight,
		&product.WeightUnit,
		&product.RequiresShipping,
		&product.IsActive,
		&product.CreatedAt,
		&product.UpdatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("product not found: %d", id)
		}
		return nil, fmt.Errorf("failed to get product: %w", err)
	}

	return &product, nil
}

// GetBySlug retrieves a product by slug within a client and shop scope
func (r *ProductRepository) GetBySlug(ctx context.Context, clientID, shopID int, slug string) (*models.Product, error) {
	query := `
		SELECT id, client_id, shop_id, category_id, sku, name, slug, description,
		       short_description, price, compare_at_price, cost_per_item,
		       weight, weight_unit, requires_shipping, is_active, created_at, updated_at
		FROM products
		WHERE client_id = $1 AND shop_id = $2 AND slug = $3
	`

	var product models.Product
	err := r.pool.QueryRow(ctx, query, clientID, shopID, slug).Scan(
		&product.ID,
		&product.ClientID,
		&product.ShopID,
		&product.CategoryID,
		&product.SKU,
		&product.Name,
		&product.Slug,
		&product.Description,
		&product.ShortDescription,
		&product.Price,
		&product.CompareAtPrice,
		&product.CostPerItem,
		&product.Weight,
		&product.WeightUnit,
		&product.RequiresShipping,
		&product.IsActive,
		&product.CreatedAt,
		&product.UpdatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("product not found: %s", slug)
		}
		return nil, fmt.Errorf("failed to get product: %w", err)
	}

	return &product, nil
}

// ListByShopID retrieves all products for a specific shop
func (r *ProductRepository) ListByShopID(ctx context.Context, shopID int) ([]*models.Product, error) {
	query := `
		SELECT id, client_id, shop_id, category_id, sku, name, slug, description,
		       short_description, price, compare_at_price, cost_per_item,
		       weight, weight_unit, requires_shipping, is_active, created_at, updated_at
		FROM products
		WHERE shop_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.pool.Query(ctx, query, shopID)
	if err != nil {
		return nil, fmt.Errorf("failed to list products: %w", err)
	}
	defer rows.Close()

	return r.scanProducts(rows)
}

// ListByCategoryID retrieves all products for a specific category
func (r *ProductRepository) ListByCategoryID(ctx context.Context, categoryID int) ([]*models.Product, error) {
	query := `
		SELECT id, client_id, shop_id, category_id, sku, name, slug, description,
		       short_description, price, compare_at_price, cost_per_item,
		       weight, weight_unit, requires_shipping, is_active, created_at, updated_at
		FROM products
		WHERE category_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.pool.Query(ctx, query, categoryID)
	if err != nil {
		return nil, fmt.Errorf("failed to list products: %w", err)
	}
	defer rows.Close()

	return r.scanProducts(rows)
}

// ListByClientID retrieves all products for a specific client
func (r *ProductRepository) ListByClientID(ctx context.Context, clientID int) ([]*models.Product, error) {
	query := `
		SELECT id, client_id, shop_id, category_id, sku, name, slug, description,
		       short_description, price, compare_at_price, cost_per_item,
		       weight, weight_unit, requires_shipping, is_active, created_at, updated_at
		FROM products
		WHERE client_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.pool.Query(ctx, query, clientID)
	if err != nil {
		return nil, fmt.Errorf("failed to list products: %w", err)
	}
	defer rows.Close()

	return r.scanProducts(rows)
}

// Update updates an existing product
func (r *ProductRepository) Update(ctx context.Context, product *models.Product) error {
	query := `
		UPDATE products
		SET category_id = $1, sku = $2, name = $3, description = $4,
		    short_description = $5, price = $6, compare_at_price = $7,
		    cost_per_item = $8, weight = $9, weight_unit = $10,
		    requires_shipping = $11, is_active = $12, updated_at = NOW()
		WHERE id = $13
		RETURNING updated_at
	`

	err := r.pool.QueryRow(
		ctx,
		query,
		product.CategoryID,
		product.SKU,
		product.Name,
		product.Description,
		product.ShortDescription,
		product.Price,
		product.CompareAtPrice,
		product.CostPerItem,
		product.Weight,
		product.WeightUnit,
		product.RequiresShipping,
		product.IsActive,
		product.ID,
	).Scan(&product.UpdatedAt)

	if err != nil {
		if err == pgx.ErrNoRows {
			return fmt.Errorf("product not found: %d", product.ID)
		}
		return fmt.Errorf("failed to update product: %w", err)
	}

	return nil
}

// Delete deletes a product by ID
func (r *ProductRepository) Delete(ctx context.Context, id int) error {
	query := `DELETE FROM products WHERE id = $1`

	result, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete product: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("product not found: %d", id)
	}

	return nil
}

// scanProducts is a helper function to scan multiple products from rows
func (r *ProductRepository) scanProducts(rows pgx.Rows) ([]*models.Product, error) {
	var products []*models.Product
	for rows.Next() {
		var product models.Product
		err := rows.Scan(
			&product.ID,
			&product.ClientID,
			&product.ShopID,
			&product.CategoryID,
			&product.SKU,
			&product.Name,
			&product.Slug,
			&product.Description,
			&product.ShortDescription,
			&product.Price,
			&product.CompareAtPrice,
			&product.CostPerItem,
			&product.Weight,
			&product.WeightUnit,
			&product.RequiresShipping,
			&product.IsActive,
			&product.CreatedAt,
			&product.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan product: %w", err)
		}
		products = append(products, &product)
	}

	return products, nil
}
