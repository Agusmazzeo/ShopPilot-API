package services

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/yourorg/shoppilot/app/repositories"
	"github.com/yourorg/shoppilot/internal/models"
)

// CreateClientRequest represents the request to create a new client
type CreateClientRequest struct {
	Name         string  `json:"name" validate:"required,min=2,max=255"`
	Description  string  `json:"description"`
	ContactEmail string  `json:"contact_email" validate:"required,email"`
	ContactPhone string  `json:"contact_phone"`
	WebsiteURL   string  `json:"website_url"`
	LogoURL      *string `json:"logo_url"`
}

// UpdateClientRequest represents the request to update a client
type UpdateClientRequest struct {
	Name         *string `json:"name,omitempty" validate:"omitempty,min=2,max=255"`
	Description  *string `json:"description,omitempty"`
	ContactEmail *string `json:"contact_email,omitempty" validate:"omitempty,email"`
	ContactPhone *string `json:"contact_phone,omitempty"`
	WebsiteURL   *string `json:"website_url,omitempty"`
	LogoURL      *string `json:"logo_url,omitempty"`
}

// ClientService defines the interface for client business logic
type ClientService interface {
	// Client management
	CreateClient(ctx context.Context, req *CreateClientRequest) (*models.Client, error)
	GetClient(ctx context.Context, id uuid.UUID) (*models.Client, error)
	GetClientBySlug(ctx context.Context, slug string) (*models.Client, error)
	UpdateClient(ctx context.Context, id uuid.UUID, req *UpdateClientRequest) error
	DeleteClient(ctx context.Context, id uuid.UUID) error
	ListClients(ctx context.Context, page, pageSize int) ([]*models.Client, int, error)
	ActivateClient(ctx context.Context, id uuid.UUID) error
	DeactivateClient(ctx context.Context, id uuid.UUID) error
}

// clientService implements the ClientService interface
type clientService struct {
	repo repositories.ClientRepository
}

// NewClientService creates a new client service instance
func NewClientService(repo repositories.ClientRepository) ClientService {
	return &clientService{repo: repo}
}

// CreateClient creates a new client with business rule validation
func (s *clientService) CreateClient(ctx context.Context, req *CreateClientRequest) (*models.Client, error) {
	// Validate email format
	if err := validateEmailFormat(req.ContactEmail); err != nil {
		return nil, fmt.Errorf("invalid email: %w", err)
	}

	// Validate phone format if provided
	if req.ContactPhone != "" {
		if err := validatePhone(req.ContactPhone); err != nil {
			return nil, fmt.Errorf("invalid phone: %w", err)
		}
	}

	// Generate slug from name
	slug := generateSlug(req.Name)

	// Check if slug already exists
	existing, err := s.repo.GetBySlug(ctx, slug)
	if err == nil && existing != nil {
		return nil, fmt.Errorf("client with slug '%s' already exists", slug)
	}

	// Create client model
	client := &models.Client{
		Name:         req.Name,
		Slug:         slug,
		Description:  req.Description,
		ContactEmail: req.ContactEmail,
		ContactPhone: req.ContactPhone,
		WebsiteURL:   req.WebsiteURL,
		LogoURL:      req.LogoURL,
		IsActive:     true, // New clients are active by default
	}

	// Save to repository
	if err := s.repo.Create(ctx, client); err != nil {
		return nil, fmt.Errorf("failed to create client: %w", err)
	}

	return client, nil
}

// GetClient retrieves a client by ID
func (s *clientService) GetClient(ctx context.Context, id uuid.UUID) (*models.Client, error) {
	client, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get client: %w", err)
	}
	return client, nil
}

// GetClientBySlug retrieves a client by slug
func (s *clientService) GetClientBySlug(ctx context.Context, slug string) (*models.Client, error) {
	client, err := s.repo.GetBySlug(ctx, slug)
	if err != nil {
		return nil, fmt.Errorf("failed to get client by slug: %w", err)
	}
	return client, nil
}

// UpdateClient updates an existing client
func (s *clientService) UpdateClient(ctx context.Context, id uuid.UUID, req *UpdateClientRequest) error {
	// Get existing client
	client, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("client not found: %w", err)
	}

	// Update fields if provided
	if req.Name != nil {
		// Regenerate slug if name changes
		newSlug := generateSlug(*req.Name)
		if newSlug != client.Slug {
			// Check if new slug already exists
			existing, err := s.repo.GetBySlug(ctx, newSlug)
			if err == nil && existing != nil && existing.ID != id {
				return fmt.Errorf("client with slug '%s' already exists", newSlug)
			}
			client.Slug = newSlug
		}
		client.Name = *req.Name
	}

	if req.Description != nil {
		client.Description = *req.Description
	}

	if req.ContactEmail != nil {
		if err := validateEmailFormat(*req.ContactEmail); err != nil {
			return fmt.Errorf("invalid email: %w", err)
		}
		client.ContactEmail = *req.ContactEmail
	}

	if req.ContactPhone != nil {
		if *req.ContactPhone != "" {
			if err := validatePhone(*req.ContactPhone); err != nil {
				return fmt.Errorf("invalid phone: %w", err)
			}
		}
		client.ContactPhone = *req.ContactPhone
	}

	if req.WebsiteURL != nil {
		client.WebsiteURL = *req.WebsiteURL
	}

	if req.LogoURL != nil {
		client.LogoURL = req.LogoURL
	}

	// Save to repository
	if err := s.repo.Update(ctx, client); err != nil {
		return fmt.Errorf("failed to update client: %w", err)
	}

	return nil
}

// DeleteClient performs a soft delete by setting is_active to false
func (s *clientService) DeleteClient(ctx context.Context, id uuid.UUID) error {
	// Get existing client
	client, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("client not found: %w", err)
	}

	// Soft delete: set is_active to false
	client.IsActive = false

	// Save to repository
	if err := s.repo.Update(ctx, client); err != nil {
		return fmt.Errorf("failed to delete client: %w", err)
	}

	return nil
}

// ListClients retrieves all clients with pagination
func (s *clientService) ListClients(ctx context.Context, page, pageSize int) ([]*models.Client, int, error) {
	// Validate pagination parameters
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	if pageSize > 100 {
		pageSize = 100 // Cap at 100 items per page
	}

	offset := (page - 1) * pageSize

	// Get clients from repository
	clients, err := s.repo.List(ctx, pageSize, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list clients: %w", err)
	}

	// Get total count (for now, we'll return the count of results)
	// In production, you'd want a separate Count() method in the repository
	total := len(clients)
	if len(clients) == pageSize {
		// If we got a full page, there might be more
		total = page * pageSize // This is a simplification
	}

	return clients, total, nil
}

// ActivateClient sets a client's is_active status to true
func (s *clientService) ActivateClient(ctx context.Context, id uuid.UUID) error {
	// Get existing client
	client, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("client not found: %w", err)
	}

	// Set active status
	client.IsActive = true

	// Save to repository
	if err := s.repo.Update(ctx, client); err != nil {
		return fmt.Errorf("failed to activate client: %w", err)
	}

	return nil
}

// DeactivateClient sets a client's is_active status to false
func (s *clientService) DeactivateClient(ctx context.Context, id uuid.UUID) error {
	// Get existing client
	client, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("client not found: %w", err)
	}

	// Set inactive status
	client.IsActive = false

	// Save to repository
	if err := s.repo.Update(ctx, client); err != nil {
		return fmt.Errorf("failed to deactivate client: %w", err)
	}

	return nil
}
