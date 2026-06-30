package services

import (
	"context"
	"fmt"
	"net/mail"

	"github.com/google/uuid"
	"github.com/yourorg/shoppilot/app/repositories"
	"github.com/yourorg/shoppilot/internal/models"
)

// CustomerService defines the interface for customer business logic
type CustomerService interface {
	CreateCustomer(ctx context.Context, clientID uuid.UUID, req *CreateCustomerRequest) (*models.Customer, error)
	GetCustomer(ctx context.Context, clientID uuid.UUID, id uuid.UUID) (*models.Customer, error)
	UpdateCustomer(ctx context.Context, clientID uuid.UUID, id uuid.UUID, req *UpdateCustomerRequest) error
	DeleteCustomer(ctx context.Context, clientID uuid.UUID, id uuid.UUID) error
	ListCustomers(ctx context.Context, clientID uuid.UUID, page, pageSize int) ([]*models.Customer, int, error)
	SearchCustomers(ctx context.Context, clientID uuid.UUID, query string, page, pageSize int) ([]*models.Customer, int, error)
}

// customerService implements CustomerService interface
type customerService struct {
	repo repositories.CustomerRepository
}

// NewCustomerService creates a new customer service
func NewCustomerService(repo repositories.CustomerRepository) CustomerService {
	return &customerService{
		repo: repo,
	}
}

// Request/Response types

type CreateCustomerRequest struct {
	Code                string
	FirstName           string
	LastName            string
	Email               string
	Phone               string
	ShippingAddress     string
	ShippingCity        string
	ShippingState       string
	ShippingPostalCode  string
	ShippingCountry     string
	BillingAddress      string
	BillingCity         string
	BillingState        string
	BillingPostalCode   string
	BillingCountry      string
	TaxID               string
	Notes               string
	IsActive            bool
	Metadata            map[string]interface{}
}

type UpdateCustomerRequest struct {
	FirstName          *string
	LastName           *string
	Email              *string
	Phone              *string
	ShippingAddress    *string
	ShippingCity       *string
	ShippingState      *string
	ShippingPostalCode *string
	ShippingCountry    *string
	BillingAddress     *string
	BillingCity        *string
	BillingState       *string
	BillingPostalCode  *string
	BillingCountry     *string
	TaxID              *string
	Notes              *string
	IsActive           *bool
	Metadata           map[string]interface{}
}

// CreateCustomer creates a new customer with business rule validations
func (s *customerService) CreateCustomer(ctx context.Context, clientID uuid.UUID, req *CreateCustomerRequest) (*models.Customer, error) {
	// Business rule: Validate email format
	if req.Email != "" {
		if _, err := mail.ParseAddress(req.Email); err != nil {
			return nil, fmt.Errorf("invalid email format: %s", req.Email)
		}
	}

	// Business rule: Customer code unique per client
	existingCustomer, err := s.repo.GetByCode(ctx, clientID, req.Code)
	if err == nil && existingCustomer != nil {
		return nil, fmt.Errorf("customer code '%s' already exists for this client", req.Code)
	}

	customer := &models.Customer{
		ClientID:           clientID,
		Code:               req.Code,
		FirstName:          req.FirstName,
		LastName:           req.LastName,
		Email:              req.Email,
		Phone:              req.Phone,
		ShippingAddress:    req.ShippingAddress,
		ShippingCity:       req.ShippingCity,
		ShippingState:      req.ShippingState,
		ShippingPostalCode: req.ShippingPostalCode,
		ShippingCountry:    req.ShippingCountry,
		BillingAddress:     req.BillingAddress,
		BillingCity:        req.BillingCity,
		BillingState:       req.BillingState,
		BillingPostalCode:  req.BillingPostalCode,
		BillingCountry:     req.BillingCountry,
		TaxID:              req.TaxID,
		Notes:              req.Notes,
		IsActive:           req.IsActive,
		Metadata:           req.Metadata,
	}

	if err := s.repo.Create(ctx, customer); err != nil {
		return nil, fmt.Errorf("failed to create customer: %w", err)
	}

	return customer, nil
}

// GetCustomer retrieves a customer by ID
func (s *customerService) GetCustomer(ctx context.Context, clientID uuid.UUID, id uuid.UUID) (*models.Customer, error) {
	customer, err := s.repo.GetByID(ctx, clientID, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get customer: %w", err)
	}

	return customer, nil
}

// UpdateCustomer updates an existing customer
func (s *customerService) UpdateCustomer(ctx context.Context, clientID uuid.UUID, id uuid.UUID, req *UpdateCustomerRequest) error {
	// Get existing customer
	customer, err := s.repo.GetByID(ctx, clientID, id)
	if err != nil {
		return fmt.Errorf("failed to get customer: %w", err)
	}

	// Business rule: Validate email format if being updated
	if req.Email != nil && *req.Email != "" {
		if _, err := mail.ParseAddress(*req.Email); err != nil {
			return fmt.Errorf("invalid email format: %s", *req.Email)
		}
	}

	// Apply updates
	if req.FirstName != nil {
		customer.FirstName = *req.FirstName
	}
	if req.LastName != nil {
		customer.LastName = *req.LastName
	}
	if req.Email != nil {
		customer.Email = *req.Email
	}
	if req.Phone != nil {
		customer.Phone = *req.Phone
	}
	if req.ShippingAddress != nil {
		customer.ShippingAddress = *req.ShippingAddress
	}
	if req.ShippingCity != nil {
		customer.ShippingCity = *req.ShippingCity
	}
	if req.ShippingState != nil {
		customer.ShippingState = *req.ShippingState
	}
	if req.ShippingPostalCode != nil {
		customer.ShippingPostalCode = *req.ShippingPostalCode
	}
	if req.ShippingCountry != nil {
		customer.ShippingCountry = *req.ShippingCountry
	}
	if req.BillingAddress != nil {
		customer.BillingAddress = *req.BillingAddress
	}
	if req.BillingCity != nil {
		customer.BillingCity = *req.BillingCity
	}
	if req.BillingState != nil {
		customer.BillingState = *req.BillingState
	}
	if req.BillingPostalCode != nil {
		customer.BillingPostalCode = *req.BillingPostalCode
	}
	if req.BillingCountry != nil {
		customer.BillingCountry = *req.BillingCountry
	}
	if req.TaxID != nil {
		customer.TaxID = *req.TaxID
	}
	if req.Notes != nil {
		customer.Notes = *req.Notes
	}
	if req.IsActive != nil {
		customer.IsActive = *req.IsActive
	}
	if req.Metadata != nil {
		customer.Metadata = req.Metadata
	}

	if err := s.repo.Update(ctx, customer); err != nil {
		return fmt.Errorf("failed to update customer: %w", err)
	}

	return nil
}

// DeleteCustomer deletes a customer with business rule validation
func (s *customerService) DeleteCustomer(ctx context.Context, clientID uuid.UUID, id uuid.UUID) error {
	// Business rule: Cannot delete customer with active orders
	// Note: This would require checking the SalesOrderRepository
	// For now, we rely on database foreign key constraints to prevent deletion
	// The database will return an error if there are related sales orders

	if err := s.repo.Delete(ctx, clientID, id); err != nil {
		return fmt.Errorf("failed to delete customer: %w", err)
	}

	return nil
}

// ListCustomers retrieves customers with pagination
func (s *customerService) ListCustomers(ctx context.Context, clientID uuid.UUID, page, pageSize int) ([]*models.Customer, int, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	offset := (page - 1) * pageSize

	customers, err := s.repo.ListByClient(ctx, clientID, pageSize, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list customers: %w", err)
	}

	// Note: Total count not implemented in repository yet, returning 0 for now
	total := len(customers)

	return customers, total, nil
}

// SearchCustomers performs a search on customers
func (s *customerService) SearchCustomers(ctx context.Context, clientID uuid.UUID, query string, page, pageSize int) ([]*models.Customer, int, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	offset := (page - 1) * pageSize

	customers, err := s.repo.Search(ctx, clientID, query, pageSize, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to search customers: %w", err)
	}

	// Note: Total count not implemented in repository yet, returning 0 for now
	total := len(customers)

	return customers, total, nil
}
