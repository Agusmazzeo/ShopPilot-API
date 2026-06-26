package repositories

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/yourorg/shoppilot/internal/models"
)

// ProductRepository handles database operations for products and variants
type ProductRepository interface {
	// Product CRUD
	Create(ctx context.Context, product *models.Product) error
	GetByID(ctx context.Context, clientID uuid.UUID, productID uuid.UUID) (*models.Product, error)
	GetByCode(ctx context.Context, clientID uuid.UUID, code string) (*models.Product, error)
	Update(ctx context.Context, product *models.Product) error
	Delete(ctx context.Context, clientID uuid.UUID, productID uuid.UUID) error
	ListByShop(ctx context.Context, clientID uuid.UUID, shopID uuid.UUID, limit, offset int) ([]*models.Product, error)
	ListByClient(ctx context.Context, clientID uuid.UUID, limit, offset int) ([]*models.Product, error)
	Search(ctx context.Context, clientID uuid.UUID, query string, limit, offset int) ([]*models.Product, error)

	// Variant operations
	CreateVariant(ctx context.Context, variant *models.ProductVariant) error
	GetVariantByID(ctx context.Context, clientID uuid.UUID, variantID uuid.UUID) (*models.ProductVariant, error)
	GetVariantBySKU(ctx context.Context, clientID uuid.UUID, sku string) (*models.ProductVariant, error)
	UpdateVariant(ctx context.Context, variant *models.ProductVariant) error
	DeleteVariant(ctx context.Context, clientID uuid.UUID, variantID uuid.UUID) error
	ListVariantsByProduct(ctx context.Context, clientID uuid.UUID, productID uuid.UUID) ([]*models.ProductVariant, error)
	GetDefaultVariant(ctx context.Context, clientID uuid.UUID, productID uuid.UUID) (*models.ProductVariant, error)
	UpdateInventory(ctx context.Context, clientID uuid.UUID, variantID uuid.UUID, quantity int) error
}

type productRepository struct {
	*BaseRepository
}

// NewProductRepository creates a new product repository
func NewProductRepository(pool *pgxpool.Pool) ProductRepository {
	return &productRepository{
		BaseRepository: NewBaseRepository(pool),
	}
}

// Product CRUD operations

// Create inserts a new product into the database
func (r *productRepository) Create(ctx context.Context, product *models.Product) error {
	query := `
		INSERT INTO products (client_id, shop_id, code, name, description, metadata, is_active)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, created_at, updated_at
	`

	err := r.pool.QueryRow(
		ctx,
		query,
		product.ClientID,
		product.ShopID,
		product.Code,
		product.Name,
		product.Description,
		product.Metadata,
		product.IsActive,
	).Scan(&product.ID, &product.CreatedAt, &product.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create product: %w", err)
	}

	return nil
}

// GetByID retrieves a product by composite key (client_id, id)
func (r *productRepository) GetByID(ctx context.Context, clientID uuid.UUID, productID uuid.UUID) (*models.Product, error) {
	query := `
		SELECT id, client_id, shop_id, code, name, description, metadata, is_active, created_at, updated_at
		FROM products
		WHERE client_id = $1 AND id = $2
	`

	var product models.Product
	err := r.pool.QueryRow(ctx, query, clientID, productID).Scan(
		&product.ID,
		&product.ClientID,
		&product.ShopID,
		&product.Code,
		&product.Name,
		&product.Description,
		&product.Metadata,
		&product.IsActive,
		&product.CreatedAt,
		&product.UpdatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("product not found: %s", productID)
		}
		return nil, fmt.Errorf("failed to get product: %w", err)
	}

	return &product, nil
}

// GetByCode retrieves a product by code (unique per client)
func (r *productRepository) GetByCode(ctx context.Context, clientID uuid.UUID, code string) (*models.Product, error) {
	query := `
		SELECT id, client_id, shop_id, code, name, description, metadata, is_active, created_at, updated_at
		FROM products
		WHERE client_id = $1 AND code = $2
	`

	var product models.Product
	err := r.pool.QueryRow(ctx, query, clientID, code).Scan(
		&product.ID,
		&product.ClientID,
		&product.ShopID,
		&product.Code,
		&product.Name,
		&product.Description,
		&product.Metadata,
		&product.IsActive,
		&product.CreatedAt,
		&product.UpdatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("product not found with code: %s", code)
		}
		return nil, fmt.Errorf("failed to get product: %w", err)
	}

	return &product, nil
}

// Update updates an existing product
func (r *productRepository) Update(ctx context.Context, product *models.Product) error {
	query := `
		UPDATE products
		SET name = $1, description = $2, metadata = $3, is_active = $4, updated_at = NOW()
		WHERE client_id = $5 AND id = $6
		RETURNING updated_at
	`

	err := r.pool.QueryRow(
		ctx,
		query,
		product.Name,
		product.Description,
		product.Metadata,
		product.IsActive,
		product.ClientID,
		product.ID,
	).Scan(&product.UpdatedAt)

	if err != nil {
		if err == pgx.ErrNoRows {
			return fmt.Errorf("product not found: %s", product.ID)
		}
		return fmt.Errorf("failed to update product: %w", err)
	}

	return nil
}

// Delete deletes a product by composite key (cascades to variants)
func (r *productRepository) Delete(ctx context.Context, clientID uuid.UUID, productID uuid.UUID) error {
	query := `DELETE FROM products WHERE client_id = $1 AND id = $2`

	result, err := r.pool.Exec(ctx, query, clientID, productID)
	if err != nil {
		return fmt.Errorf("failed to delete product: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("product not found: %s", productID)
	}

	return nil
}

// ListByShop retrieves all products for a specific shop
func (r *productRepository) ListByShop(ctx context.Context, clientID uuid.UUID, shopID uuid.UUID, limit, offset int) ([]*models.Product, error) {
	query := `
		SELECT id, client_id, shop_id, code, name, description, metadata, is_active, created_at, updated_at
		FROM products
		WHERE client_id = $1 AND shop_id = $2
		ORDER BY created_at DESC
		LIMIT $3 OFFSET $4
	`

	rows, err := r.pool.Query(ctx, query, clientID, shopID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list products by shop: %w", err)
	}
	defer rows.Close()

	var products []*models.Product
	for rows.Next() {
		var product models.Product
		err := rows.Scan(
			&product.ID,
			&product.ClientID,
			&product.ShopID,
			&product.Code,
			&product.Name,
			&product.Description,
			&product.Metadata,
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

// ListByClient retrieves all products for a specific client
func (r *productRepository) ListByClient(ctx context.Context, clientID uuid.UUID, limit, offset int) ([]*models.Product, error) {
	query := `
		SELECT id, client_id, shop_id, code, name, description, metadata, is_active, created_at, updated_at
		FROM products
		WHERE client_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.pool.Query(ctx, query, clientID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list products by client: %w", err)
	}
	defer rows.Close()

	var products []*models.Product
	for rows.Next() {
		var product models.Product
		err := rows.Scan(
			&product.ID,
			&product.ClientID,
			&product.ShopID,
			&product.Code,
			&product.Name,
			&product.Description,
			&product.Metadata,
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

// Search performs full-text search on products
func (r *productRepository) Search(ctx context.Context, clientID uuid.UUID, query string, limit, offset int) ([]*models.Product, error) {
	searchQuery := `
		SELECT id, client_id, shop_id, code, name, description, metadata, is_active, created_at, updated_at
		FROM products
		WHERE client_id = $1
		  AND (name ILIKE $2 OR description ILIKE $2 OR code ILIKE $2)
		ORDER BY created_at DESC
		LIMIT $3 OFFSET $4
	`

	searchPattern := "%" + query + "%"
	rows, err := r.pool.Query(ctx, searchQuery, clientID, searchPattern, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to search products: %w", err)
	}
	defer rows.Close()

	var products []*models.Product
	for rows.Next() {
		var product models.Product
		err := rows.Scan(
			&product.ID,
			&product.ClientID,
			&product.ShopID,
			&product.Code,
			&product.Name,
			&product.Description,
			&product.Metadata,
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

// Variant operations

// CreateVariant inserts a new product variant
func (r *productRepository) CreateVariant(ctx context.Context, variant *models.ProductVariant) error {
	query := `
		INSERT INTO product_variants (
			client_id, shop_id, product_id, sku, name, price, compare_at_price, cost,
			quantity, weight, weight_unit, requires_shipping, is_default, attributes, is_active
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
		RETURNING id, created_at, updated_at
	`

	err := r.pool.QueryRow(
		ctx,
		query,
		variant.ClientID,
		variant.ShopID,
		variant.ProductID,
		variant.SKU,
		variant.Name,
		variant.Price,
		variant.CompareAtPrice,
		variant.Cost,
		variant.Quantity,
		variant.Weight,
		variant.WeightUnit,
		variant.RequiresShipping,
		variant.IsDefault,
		variant.Attributes,
		variant.IsActive,
	).Scan(&variant.ID, &variant.CreatedAt, &variant.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create variant: %w", err)
	}

	return nil
}

// GetVariantByID retrieves a variant by composite key (client_id, id)
func (r *productRepository) GetVariantByID(ctx context.Context, clientID uuid.UUID, variantID uuid.UUID) (*models.ProductVariant, error) {
	query := `
		SELECT id, client_id, shop_id, product_id, sku, name, price, compare_at_price, cost,
		       quantity, weight, weight_unit, requires_shipping, is_default, attributes, is_active,
		       created_at, updated_at
		FROM product_variants
		WHERE client_id = $1 AND id = $2
	`

	var variant models.ProductVariant
	err := r.pool.QueryRow(ctx, query, clientID, variantID).Scan(
		&variant.ID,
		&variant.ClientID,
		&variant.ShopID,
		&variant.ProductID,
		&variant.SKU,
		&variant.Name,
		&variant.Price,
		&variant.CompareAtPrice,
		&variant.Cost,
		&variant.Quantity,
		&variant.Weight,
		&variant.WeightUnit,
		&variant.RequiresShipping,
		&variant.IsDefault,
		&variant.Attributes,
		&variant.IsActive,
		&variant.CreatedAt,
		&variant.UpdatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("variant not found: %s", variantID)
		}
		return nil, fmt.Errorf("failed to get variant: %w", err)
	}

	return &variant, nil
}

// GetVariantBySKU retrieves a variant by SKU (unique per client)
func (r *productRepository) GetVariantBySKU(ctx context.Context, clientID uuid.UUID, sku string) (*models.ProductVariant, error) {
	query := `
		SELECT id, client_id, shop_id, product_id, sku, name, price, compare_at_price, cost,
		       quantity, weight, weight_unit, requires_shipping, is_default, attributes, is_active,
		       created_at, updated_at
		FROM product_variants
		WHERE client_id = $1 AND sku = $2
	`

	var variant models.ProductVariant
	err := r.pool.QueryRow(ctx, query, clientID, sku).Scan(
		&variant.ID,
		&variant.ClientID,
		&variant.ShopID,
		&variant.ProductID,
		&variant.SKU,
		&variant.Name,
		&variant.Price,
		&variant.CompareAtPrice,
		&variant.Cost,
		&variant.Quantity,
		&variant.Weight,
		&variant.WeightUnit,
		&variant.RequiresShipping,
		&variant.IsDefault,
		&variant.Attributes,
		&variant.IsActive,
		&variant.CreatedAt,
		&variant.UpdatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("variant not found with SKU: %s", sku)
		}
		return nil, fmt.Errorf("failed to get variant: %w", err)
	}

	return &variant, nil
}

// UpdateVariant updates an existing product variant
func (r *productRepository) UpdateVariant(ctx context.Context, variant *models.ProductVariant) error {
	query := `
		UPDATE product_variants
		SET name = $1, price = $2, compare_at_price = $3, cost = $4, quantity = $5,
		    weight = $6, weight_unit = $7, requires_shipping = $8, is_default = $9,
		    attributes = $10, is_active = $11, updated_at = NOW()
		WHERE client_id = $12 AND id = $13
		RETURNING updated_at
	`

	err := r.pool.QueryRow(
		ctx,
		query,
		variant.Name,
		variant.Price,
		variant.CompareAtPrice,
		variant.Cost,
		variant.Quantity,
		variant.Weight,
		variant.WeightUnit,
		variant.RequiresShipping,
		variant.IsDefault,
		variant.Attributes,
		variant.IsActive,
		variant.ClientID,
		variant.ID,
	).Scan(&variant.UpdatedAt)

	if err != nil {
		if err == pgx.ErrNoRows {
			return fmt.Errorf("variant not found: %s", variant.ID)
		}
		return fmt.Errorf("failed to update variant: %w", err)
	}

	return nil
}

// DeleteVariant deletes a product variant by composite key
func (r *productRepository) DeleteVariant(ctx context.Context, clientID uuid.UUID, variantID uuid.UUID) error {
	query := `DELETE FROM product_variants WHERE client_id = $1 AND id = $2`

	result, err := r.pool.Exec(ctx, query, clientID, variantID)
	if err != nil {
		return fmt.Errorf("failed to delete variant: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("variant not found: %s", variantID)
	}

	return nil
}

// ListVariantsByProduct retrieves all variants for a specific product
func (r *productRepository) ListVariantsByProduct(ctx context.Context, clientID uuid.UUID, productID uuid.UUID) ([]*models.ProductVariant, error) {
	query := `
		SELECT id, client_id, shop_id, product_id, sku, name, price, compare_at_price, cost,
		       quantity, weight, weight_unit, requires_shipping, is_default, attributes, is_active,
		       created_at, updated_at
		FROM product_variants
		WHERE client_id = $1 AND product_id = $2
		ORDER BY is_default DESC, created_at ASC
	`

	rows, err := r.pool.Query(ctx, query, clientID, productID)
	if err != nil {
		return nil, fmt.Errorf("failed to list variants: %w", err)
	}
	defer rows.Close()

	var variants []*models.ProductVariant
	for rows.Next() {
		var variant models.ProductVariant
		err := rows.Scan(
			&variant.ID,
			&variant.ClientID,
			&variant.ShopID,
			&variant.ProductID,
			&variant.SKU,
			&variant.Name,
			&variant.Price,
			&variant.CompareAtPrice,
			&variant.Cost,
			&variant.Quantity,
			&variant.Weight,
			&variant.WeightUnit,
			&variant.RequiresShipping,
			&variant.IsDefault,
			&variant.Attributes,
			&variant.IsActive,
			&variant.CreatedAt,
			&variant.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan variant: %w", err)
		}
		variants = append(variants, &variant)
	}

	return variants, nil
}

// GetDefaultVariant retrieves the default variant for a product
func (r *productRepository) GetDefaultVariant(ctx context.Context, clientID uuid.UUID, productID uuid.UUID) (*models.ProductVariant, error) {
	query := `
		SELECT id, client_id, shop_id, product_id, sku, name, price, compare_at_price, cost,
		       quantity, weight, weight_unit, requires_shipping, is_default, attributes, is_active,
		       created_at, updated_at
		FROM product_variants
		WHERE client_id = $1 AND product_id = $2 AND is_default = true
	`

	var variant models.ProductVariant
	err := r.pool.QueryRow(ctx, query, clientID, productID).Scan(
		&variant.ID,
		&variant.ClientID,
		&variant.ShopID,
		&variant.ProductID,
		&variant.SKU,
		&variant.Name,
		&variant.Price,
		&variant.CompareAtPrice,
		&variant.Cost,
		&variant.Quantity,
		&variant.Weight,
		&variant.WeightUnit,
		&variant.RequiresShipping,
		&variant.IsDefault,
		&variant.Attributes,
		&variant.IsActive,
		&variant.CreatedAt,
		&variant.UpdatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("default variant not found for product: %s", productID)
		}
		return nil, fmt.Errorf("failed to get default variant: %w", err)
	}

	return &variant, nil
}

// UpdateInventory adjusts the quantity of a variant
func (r *productRepository) UpdateInventory(ctx context.Context, clientID uuid.UUID, variantID uuid.UUID, quantity int) error {
	query := `
		UPDATE product_variants
		SET quantity = $1, updated_at = NOW()
		WHERE client_id = $2 AND id = $3
	`

	result, err := r.pool.Exec(ctx, query, quantity, clientID, variantID)
	if err != nil {
		return fmt.Errorf("failed to update inventory: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("variant not found: %s", variantID)
	}

	return nil
}
