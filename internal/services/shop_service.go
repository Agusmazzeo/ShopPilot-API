package services

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/yourorg/shoppilot/app/repositories"
	"github.com/yourorg/shoppilot/internal/models"
)

// CreateShopRequest represents the request to create a shop
type CreateShopRequest struct {
	Name        string  `json:"name" validate:"required,min=1,max=255"`
	Description string  `json:"description"`
	WebpageURL  string  `json:"webpage_url"`
	Address     string  `json:"address"`
	City        string  `json:"city"`
	State       string  `json:"state"`
	Country     string  `json:"country"`
	PostalCode  string  `json:"postal_code"`
	Phone       string  `json:"phone"`
	Email       string  `json:"email" validate:"omitempty,email"`
	LogoURL     *string `json:"logo_url"`
}

// UpdateShopRequest represents the request to update a shop
type UpdateShopRequest struct {
	Name        *string `json:"name" validate:"omitempty,min=1,max=255"`
	Description *string `json:"description"`
	WebpageURL  *string `json:"webpage_url"`
	Address     *string `json:"address"`
	City        *string `json:"city"`
	State       *string `json:"state"`
	Country     *string `json:"country"`
	PostalCode  *string `json:"postal_code"`
	Phone       *string `json:"phone"`
	Email       *string `json:"email" validate:"omitempty,email"`
	LogoURL     *string `json:"logo_url"`
	IsActive    *bool   `json:"is_active"`
}

// ShopService defines the interface for shop business logic
type ShopService interface {
	// Shop management
	CreateShop(ctx context.Context, clientID uuid.UUID, req *CreateShopRequest) (*models.Shop, error)
	GetShop(ctx context.Context, clientID uuid.UUID, shopID uuid.UUID) (*models.Shop, error)
	GetShopBySlug(ctx context.Context, clientID uuid.UUID, slug string) (*models.Shop, error)
	UpdateShop(ctx context.Context, clientID uuid.UUID, shopID uuid.UUID, req *UpdateShopRequest) error
	DeleteShop(ctx context.Context, clientID uuid.UUID, shopID uuid.UUID) error
	ListShops(ctx context.Context, clientID uuid.UUID, page, pageSize int) ([]*models.Shop, int, error)

	// Shop users
	AssignUserToShop(ctx context.Context, shopID uuid.UUID, clientUserID uuid.UUID, roleName string) error
	RemoveUserFromShop(ctx context.Context, shopID uuid.UUID, clientUserRoleID int) error
	GetShopUsers(ctx context.Context, shopID uuid.UUID) ([]*models.ShopUser, error)
}

// shopService implements ShopService interface
type shopService struct {
	shopRepo   repositories.ShopRepository
	clientRepo repositories.ClientRepository
}

// NewShopService creates a new shop service
func NewShopService(shopRepo repositories.ShopRepository, clientRepo repositories.ClientRepository) ShopService {
	return &shopService{
		shopRepo:   shopRepo,
		clientRepo: clientRepo,
	}
}

// CreateShop creates a new shop with business rules validation
func (s *shopService) CreateShop(ctx context.Context, clientID uuid.UUID, req *CreateShopRequest) (*models.Shop, error) {
	// Validate client exists
	client, err := s.clientRepo.GetByID(ctx, clientID)
	if err != nil {
		return nil, fmt.Errorf("client validation failed: %w", err)
	}
	if client == nil {
		return nil, fmt.Errorf("client not found: %s", clientID)
	}

	// Generate slug from name (unique per client)
	slug := generateSlug(req.Name)

	// Check if slug already exists for this client
	existingShop, err := s.shopRepo.GetBySlug(ctx, clientID, slug)
	if err == nil && existingShop != nil {
		// Slug exists, append a number to make it unique
		slug = s.ensureUniqueSlug(ctx, clientID, slug)
	}

	// Create shop model
	shop := &models.Shop{
		ClientID:    clientID,
		Name:        req.Name,
		Slug:        slug,
		Description: req.Description,
		WebpageURL:  req.WebpageURL,
		Address:     req.Address,
		City:        req.City,
		State:       req.State,
		Country:     req.Country,
		PostalCode:  req.PostalCode,
		Phone:       req.Phone,
		Email:       req.Email,
		LogoURL:     req.LogoURL,
		IsActive:    true, // New shops are active by default
	}

	// Create shop in repository
	if err := s.shopRepo.Create(ctx, shop); err != nil {
		return nil, fmt.Errorf("failed to create shop: %w", err)
	}

	return shop, nil
}

// GetShop retrieves a shop by ID
func (s *shopService) GetShop(ctx context.Context, clientID uuid.UUID, shopID uuid.UUID) (*models.Shop, error) {
	shop, err := s.shopRepo.GetByID(ctx, clientID, shopID)
	if err != nil {
		return nil, fmt.Errorf("failed to get shop: %w", err)
	}
	return shop, nil
}

// GetShopBySlug retrieves a shop by slug
func (s *shopService) GetShopBySlug(ctx context.Context, clientID uuid.UUID, slug string) (*models.Shop, error) {
	shop, err := s.shopRepo.GetBySlug(ctx, clientID, slug)
	if err != nil {
		return nil, fmt.Errorf("failed to get shop by slug: %w", err)
	}
	return shop, nil
}

// UpdateShop updates an existing shop
func (s *shopService) UpdateShop(ctx context.Context, clientID uuid.UUID, shopID uuid.UUID, req *UpdateShopRequest) error {
	// Get existing shop
	shop, err := s.shopRepo.GetByID(ctx, clientID, shopID)
	if err != nil {
		return fmt.Errorf("failed to get shop: %w", err)
	}

	// Update fields if provided
	if req.Name != nil {
		shop.Name = *req.Name
		// Regenerate slug if name changed
		newSlug := generateSlug(*req.Name)
		if newSlug != shop.Slug {
			// Check if new slug already exists for this client
			existingShop, err := s.shopRepo.GetBySlug(ctx, clientID, newSlug)
			if err == nil && existingShop != nil && existingShop.ID != shopID {
				// Slug exists for a different shop, make it unique
				newSlug = s.ensureUniqueSlug(ctx, clientID, newSlug)
			}
			shop.Slug = newSlug
		}
	}
	if req.Description != nil {
		shop.Description = *req.Description
	}
	if req.WebpageURL != nil {
		shop.WebpageURL = *req.WebpageURL
	}
	if req.Address != nil {
		shop.Address = *req.Address
	}
	if req.City != nil {
		shop.City = *req.City
	}
	if req.State != nil {
		shop.State = *req.State
	}
	if req.Country != nil {
		shop.Country = *req.Country
	}
	if req.PostalCode != nil {
		shop.PostalCode = *req.PostalCode
	}
	if req.Phone != nil {
		shop.Phone = *req.Phone
	}
	if req.Email != nil {
		shop.Email = *req.Email
	}
	if req.LogoURL != nil {
		shop.LogoURL = req.LogoURL
	}
	if req.IsActive != nil {
		shop.IsActive = *req.IsActive
	}

	// Update shop in repository
	if err := s.shopRepo.Update(ctx, shop); err != nil {
		return fmt.Errorf("failed to update shop: %w", err)
	}

	return nil
}

// DeleteShop deletes a shop (cascades to products)
func (s *shopService) DeleteShop(ctx context.Context, clientID uuid.UUID, shopID uuid.UUID) error {
	// Verify shop exists before deleting
	_, err := s.shopRepo.GetByID(ctx, clientID, shopID)
	if err != nil {
		return fmt.Errorf("failed to get shop: %w", err)
	}

	// Delete shop (cascade will handle products)
	if err := s.shopRepo.Delete(ctx, clientID, shopID); err != nil {
		return fmt.Errorf("failed to delete shop: %w", err)
	}

	return nil
}

// ListShops lists shops for a client with pagination
func (s *shopService) ListShops(ctx context.Context, clientID uuid.UUID, page, pageSize int) ([]*models.Shop, int, error) {
	// Validate pagination parameters
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20 // Default page size
	}

	offset := (page - 1) * pageSize

	// Get shops
	shops, err := s.shopRepo.ListByClient(ctx, clientID, pageSize, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list shops: %w", err)
	}

	// For a complete implementation, we would query total count
	// For now, return 0 as total (handler would need to implement count separately)
	total := 0

	return shops, total, nil
}

// AssignUserToShop assigns a user to a shop via client_user_role
func (s *shopService) AssignUserToShop(ctx context.Context, shopID uuid.UUID, clientUserID uuid.UUID, roleName string) error {
	// Note: This is a simplified implementation
	// In a complete implementation, we would:
	// 1. Validate the client user exists
	// 2. Look up the client_user_role by client_user_id and role_name
	// 3. Assign that client_user_role_id to the shop

	// For now, we're assuming clientUserRoleID is passed directly
	// This would need to be refactored with a ClientUserRepository
	return fmt.Errorf("not implemented: need ClientUserRepository to lookup client_user_role_id")
}

// RemoveUserFromShop removes a user from a shop
func (s *shopService) RemoveUserFromShop(ctx context.Context, shopID uuid.UUID, clientUserRoleID int) error {
	if err := s.shopRepo.RemoveUser(ctx, shopID, clientUserRoleID); err != nil {
		return fmt.Errorf("failed to remove user from shop: %w", err)
	}
	return nil
}

// GetShopUsers retrieves all users assigned to a shop
func (s *shopService) GetShopUsers(ctx context.Context, shopID uuid.UUID) ([]*models.ShopUser, error) {
	shopUsers, err := s.shopRepo.GetShopUsers(ctx, shopID)
	if err != nil {
		return nil, fmt.Errorf("failed to get shop users: %w", err)
	}
	return shopUsers, nil
}

// ensureUniqueSlug appends a number to make the slug unique for this client
func (s *shopService) ensureUniqueSlug(ctx context.Context, clientID uuid.UUID, baseSlug string) string {
	slug := baseSlug
	counter := 1

	for {
		testSlug := fmt.Sprintf("%s-%d", slug, counter)
		existingShop, err := s.shopRepo.GetBySlug(ctx, clientID, testSlug)
		if err != nil || existingShop == nil {
			// Slug is available
			return testSlug
		}
		counter++

		// Safety check to prevent infinite loop
		if counter > 1000 {
			// Fallback to UUID-based slug
			return fmt.Sprintf("%s-%s", slug, uuid.New().String()[:8])
		}
	}
}
