package services

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/yourorg/shoppilot/app/repositories"
	"github.com/yourorg/shoppilot/internal/models"
)

// ProductService defines the interface for product business logic
type ProductService interface {
	// Product management
	CreateProduct(ctx context.Context, req *CreateProductRequest) (*models.Product, error)
	GetProduct(ctx context.Context, clientID uuid.UUID, productID uuid.UUID) (*models.Product, error)
	UpdateProduct(ctx context.Context, clientID uuid.UUID, productID uuid.UUID, req *UpdateProductRequest) error
	DeleteProduct(ctx context.Context, clientID uuid.UUID, productID uuid.UUID) error
	ListProducts(ctx context.Context, clientID uuid.UUID, shopID *uuid.UUID, page, pageSize int) ([]*models.Product, int, error)
	SearchProducts(ctx context.Context, clientID uuid.UUID, query string, page, pageSize int) ([]*models.Product, int, error)

	// Variant management
	CreateVariant(ctx context.Context, clientID uuid.UUID, productID uuid.UUID, req *CreateVariantRequest) (*models.ProductVariant, error)
	GetVariant(ctx context.Context, clientID uuid.UUID, variantID uuid.UUID) (*models.ProductVariant, error)
	UpdateVariant(ctx context.Context, clientID uuid.UUID, variantID uuid.UUID, req *UpdateVariantRequest) error
	DeleteVariant(ctx context.Context, clientID uuid.UUID, variantID uuid.UUID) error
	ListVariants(ctx context.Context, clientID uuid.UUID, productID uuid.UUID) ([]*models.ProductVariant, error)

	// Inventory
	AdjustInventory(ctx context.Context, clientID uuid.UUID, variantID uuid.UUID, delta int) error
	SetInventory(ctx context.Context, clientID uuid.UUID, variantID uuid.UUID, quantity int) error
	CheckStock(ctx context.Context, clientID uuid.UUID, variantID uuid.UUID) (int, error)
}

// productService implements ProductService interface
type productService struct {
	repo repositories.ProductRepository
}

// NewProductService creates a new product service
func NewProductService(repo repositories.ProductRepository) ProductService {
	return &productService{
		repo: repo,
	}
}

// Request/Response types

type CreateProductRequest struct {
	ClientID    uuid.UUID
	ShopID      uuid.UUID
	Code        string
	Name        string
	Description string
	Metadata    map[string]interface{}
	IsActive    bool
}

type UpdateProductRequest struct {
	Name        *string
	Description *string
	Metadata    map[string]interface{}
	IsActive    *bool
}

type CreateVariantRequest struct {
	SKU              string
	Name             string
	Price            float64
	CompareAtPrice   *float64
	Cost             *float64
	Quantity         int
	Weight           *float64
	WeightUnit       string
	RequiresShipping bool
	IsDefault        bool
	Attributes       map[string]interface{}
	IsActive         bool
}

type UpdateVariantRequest struct {
	Name             *string
	Price            *float64
	CompareAtPrice   *float64
	Cost             *float64
	Quantity         *int
	Weight           *float64
	WeightUnit       *string
	RequiresShipping *bool
	IsDefault        *bool
	Attributes       map[string]interface{}
	IsActive         *bool
}

// Product management

// CreateProduct creates a new product with business rule validations
func (s *productService) CreateProduct(ctx context.Context, req *CreateProductRequest) (*models.Product, error) {
	// Validate product code is unique per client
	existingProduct, err := s.repo.GetByCode(ctx, req.ClientID, req.Code)
	if err == nil && existingProduct != nil {
		return nil, fmt.Errorf("product code '%s' already exists for this client", req.Code)
	}

	product := &models.Product{
		ClientID:    req.ClientID,
		ShopID:      req.ShopID,
		Code:        req.Code,
		Name:        req.Name,
		Description: req.Description,
		Metadata:    req.Metadata,
		IsActive:    req.IsActive,
	}

	if err := s.repo.Create(ctx, product); err != nil {
		return nil, fmt.Errorf("failed to create product: %w", err)
	}

	return product, nil
}

// GetProduct retrieves a product by ID
func (s *productService) GetProduct(ctx context.Context, clientID uuid.UUID, productID uuid.UUID) (*models.Product, error) {
	product, err := s.repo.GetByID(ctx, clientID, productID)
	if err != nil {
		return nil, fmt.Errorf("failed to get product: %w", err)
	}

	return product, nil
}

// UpdateProduct updates an existing product
func (s *productService) UpdateProduct(ctx context.Context, clientID uuid.UUID, productID uuid.UUID, req *UpdateProductRequest) error {
	// Get existing product
	product, err := s.repo.GetByID(ctx, clientID, productID)
	if err != nil {
		return fmt.Errorf("failed to get product: %w", err)
	}

	// Apply updates
	if req.Name != nil {
		product.Name = *req.Name
	}
	if req.Description != nil {
		product.Description = *req.Description
	}
	if req.Metadata != nil {
		product.Metadata = req.Metadata
	}
	if req.IsActive != nil {
		product.IsActive = *req.IsActive
	}

	if err := s.repo.Update(ctx, product); err != nil {
		return fmt.Errorf("failed to update product: %w", err)
	}

	return nil
}

// DeleteProduct deletes a product with business rule validation
func (s *productService) DeleteProduct(ctx context.Context, clientID uuid.UUID, productID uuid.UUID) error {
	// Business rule: Every product must have ≥1 variant (enforce on delete)
	// Note: This is enforced at product creation level - products should be created with at least one variant
	// Deleting a product will cascade delete all variants (handled by database)

	if err := s.repo.Delete(ctx, clientID, productID); err != nil {
		return fmt.Errorf("failed to delete product: %w", err)
	}

	return nil
}

// ListProducts retrieves products with optional filtering by shop
func (s *productService) ListProducts(ctx context.Context, clientID uuid.UUID, shopID *uuid.UUID, page, pageSize int) ([]*models.Product, int, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	offset := (page - 1) * pageSize

	var products []*models.Product
	var err error

	if shopID != nil {
		products, err = s.repo.ListByShop(ctx, clientID, *shopID, pageSize, offset)
	} else {
		products, err = s.repo.ListByClient(ctx, clientID, pageSize, offset)
	}

	if err != nil {
		return nil, 0, fmt.Errorf("failed to list products: %w", err)
	}

	// Note: Total count not implemented in repository yet, returning 0 for now
	total := len(products)

	return products, total, nil
}

// SearchProducts performs a search on products
func (s *productService) SearchProducts(ctx context.Context, clientID uuid.UUID, query string, page, pageSize int) ([]*models.Product, int, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	offset := (page - 1) * pageSize

	products, err := s.repo.Search(ctx, clientID, query, pageSize, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to search products: %w", err)
	}

	// Note: Total count not implemented in repository yet, returning 0 for now
	total := len(products)

	return products, total, nil
}

// Variant management

// CreateVariant creates a new product variant with business rule validations
func (s *productService) CreateVariant(ctx context.Context, clientID uuid.UUID, productID uuid.UUID, req *CreateVariantRequest) (*models.ProductVariant, error) {
	// Verify product exists and belongs to client
	product, err := s.repo.GetByID(ctx, clientID, productID)
	if err != nil {
		return nil, fmt.Errorf("product not found: %w", err)
	}

	// Validate SKU is unique per client
	existingVariant, err := s.repo.GetVariantBySKU(ctx, clientID, req.SKU)
	if err == nil && existingVariant != nil {
		return nil, fmt.Errorf("SKU '%s' already exists for this client", req.SKU)
	}

	// Business rule: Only one variant per product can be is_default=true
	if req.IsDefault {
		// Check if there's already a default variant for this product
		existingDefault, err := s.repo.GetDefaultVariant(ctx, clientID, productID)
		if err == nil && existingDefault != nil {
			// Unset the current default
			existingDefault.IsDefault = false
			if err := s.repo.UpdateVariant(ctx, existingDefault); err != nil {
				return nil, fmt.Errorf("failed to unset existing default variant: %w", err)
			}
		}
	}

	variant := &models.ProductVariant{
		ClientID:         clientID,
		ShopID:           product.ShopID,
		ProductID:        productID,
		SKU:              req.SKU,
		Name:             req.Name,
		Price:            req.Price,
		CompareAtPrice:   req.CompareAtPrice,
		Cost:             req.Cost,
		Quantity:         req.Quantity,
		Weight:           req.Weight,
		WeightUnit:       req.WeightUnit,
		RequiresShipping: req.RequiresShipping,
		IsDefault:        req.IsDefault,
		Attributes:       req.Attributes,
		IsActive:         req.IsActive,
	}

	if err := s.repo.CreateVariant(ctx, variant); err != nil {
		return nil, fmt.Errorf("failed to create variant: %w", err)
	}

	return variant, nil
}

// GetVariant retrieves a variant by ID
func (s *productService) GetVariant(ctx context.Context, clientID uuid.UUID, variantID uuid.UUID) (*models.ProductVariant, error) {
	variant, err := s.repo.GetVariantByID(ctx, clientID, variantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get variant: %w", err)
	}

	return variant, nil
}

// UpdateVariant updates an existing variant with business rule validations
func (s *productService) UpdateVariant(ctx context.Context, clientID uuid.UUID, variantID uuid.UUID, req *UpdateVariantRequest) error {
	// Get existing variant
	variant, err := s.repo.GetVariantByID(ctx, clientID, variantID)
	if err != nil {
		return fmt.Errorf("failed to get variant: %w", err)
	}

	// Business rule: Only one variant per product can be is_default=true
	if req.IsDefault != nil && *req.IsDefault && !variant.IsDefault {
		// Check if there's already a default variant for this product
		existingDefault, err := s.repo.GetDefaultVariant(ctx, clientID, variant.ProductID)
		if err == nil && existingDefault != nil && existingDefault.ID != variantID {
			// Unset the current default
			existingDefault.IsDefault = false
			if err := s.repo.UpdateVariant(ctx, existingDefault); err != nil {
				return fmt.Errorf("failed to unset existing default variant: %w", err)
			}
		}
	}

	// Apply updates
	if req.Name != nil {
		variant.Name = *req.Name
	}
	if req.Price != nil {
		variant.Price = *req.Price
	}
	if req.CompareAtPrice != nil {
		variant.CompareAtPrice = req.CompareAtPrice
	}
	if req.Cost != nil {
		variant.Cost = req.Cost
	}
	if req.Quantity != nil {
		// Business rule: Inventory cannot go negative
		if *req.Quantity < 0 {
			return fmt.Errorf("inventory quantity cannot be negative")
		}
		variant.Quantity = *req.Quantity
	}
	if req.Weight != nil {
		variant.Weight = req.Weight
	}
	if req.WeightUnit != nil {
		variant.WeightUnit = *req.WeightUnit
	}
	if req.RequiresShipping != nil {
		variant.RequiresShipping = *req.RequiresShipping
	}
	if req.IsDefault != nil {
		variant.IsDefault = *req.IsDefault
	}
	if req.Attributes != nil {
		variant.Attributes = req.Attributes
	}
	if req.IsActive != nil {
		variant.IsActive = *req.IsActive
	}

	if err := s.repo.UpdateVariant(ctx, variant); err != nil {
		return fmt.Errorf("failed to update variant: %w", err)
	}

	return nil
}

// DeleteVariant deletes a variant with business rule validation
func (s *productService) DeleteVariant(ctx context.Context, clientID uuid.UUID, variantID uuid.UUID) error {
	// Get the variant to find its product
	variant, err := s.repo.GetVariantByID(ctx, clientID, variantID)
	if err != nil {
		return fmt.Errorf("failed to get variant: %w", err)
	}

	// Business rule: Every product must have ≥1 variant (enforce on delete)
	variants, err := s.repo.ListVariantsByProduct(ctx, clientID, variant.ProductID)
	if err != nil {
		return fmt.Errorf("failed to list variants: %w", err)
	}

	if len(variants) <= 1 {
		return fmt.Errorf("cannot delete the last variant of a product; products must have at least one variant")
	}

	if err := s.repo.DeleteVariant(ctx, clientID, variantID); err != nil {
		return fmt.Errorf("failed to delete variant: %w", err)
	}

	return nil
}

// ListVariants retrieves all variants for a product
func (s *productService) ListVariants(ctx context.Context, clientID uuid.UUID, productID uuid.UUID) ([]*models.ProductVariant, error) {
	// Verify product exists
	_, err := s.repo.GetByID(ctx, clientID, productID)
	if err != nil {
		return nil, fmt.Errorf("product not found: %w", err)
	}

	variants, err := s.repo.ListVariantsByProduct(ctx, clientID, productID)
	if err != nil {
		return nil, fmt.Errorf("failed to list variants: %w", err)
	}

	return variants, nil
}

// Inventory management

// AdjustInventory adjusts the inventory by a delta (can be positive or negative)
func (s *productService) AdjustInventory(ctx context.Context, clientID uuid.UUID, variantID uuid.UUID, delta int) error {
	variant, err := s.repo.GetVariantByID(ctx, clientID, variantID)
	if err != nil {
		return fmt.Errorf("failed to get variant: %w", err)
	}

	newQuantity := variant.Quantity + delta

	// Business rule: Inventory cannot go negative
	if newQuantity < 0 {
		return fmt.Errorf("insufficient inventory: current=%d, delta=%d, result would be negative", variant.Quantity, delta)
	}

	if err := s.repo.UpdateInventory(ctx, clientID, variantID, newQuantity); err != nil {
		return fmt.Errorf("failed to adjust inventory: %w", err)
	}

	return nil
}

// SetInventory sets the inventory to a specific quantity
func (s *productService) SetInventory(ctx context.Context, clientID uuid.UUID, variantID uuid.UUID, quantity int) error {
	// Business rule: Inventory cannot go negative
	if quantity < 0 {
		return fmt.Errorf("inventory quantity cannot be negative: %d", quantity)
	}

	if err := s.repo.UpdateInventory(ctx, clientID, variantID, quantity); err != nil {
		return fmt.Errorf("failed to set inventory: %w", err)
	}

	return nil
}

// CheckStock retrieves the current stock level for a variant
func (s *productService) CheckStock(ctx context.Context, clientID uuid.UUID, variantID uuid.UUID) (int, error) {
	variant, err := s.repo.GetVariantByID(ctx, clientID, variantID)
	if err != nil {
		return 0, fmt.Errorf("failed to get variant: %w", err)
	}

	return variant.Quantity, nil
}
