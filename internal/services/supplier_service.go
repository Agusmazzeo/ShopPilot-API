package services

import (
	"context"
	"fmt"
	"net/mail"

	"github.com/google/uuid"
	"github.com/yourorg/shoppilot/app/repositories"
	"github.com/yourorg/shoppilot/internal/models"
)

// SupplierService defines the interface for supplier business logic
type SupplierService interface {
	CreateSupplier(ctx context.Context, clientID uuid.UUID, req *CreateSupplierRequest) (*models.Supplier, error)
	GetSupplier(ctx context.Context, clientID uuid.UUID, id uuid.UUID) (*models.Supplier, error)
	UpdateSupplier(ctx context.Context, clientID uuid.UUID, id uuid.UUID, req *UpdateSupplierRequest) error
	DeleteSupplier(ctx context.Context, clientID uuid.UUID, id uuid.UUID) error
	ListSuppliers(ctx context.Context, clientID uuid.UUID, page, pageSize int) ([]*models.Supplier, int, error)
	ListActiveSuppliers(ctx context.Context, clientID uuid.UUID, page, pageSize int) ([]*models.Supplier, int, error)
}

// supplierService implements SupplierService interface
type supplierService struct {
	repo repositories.SupplierRepository
}

// NewSupplierService creates a new supplier service
func NewSupplierService(repo repositories.SupplierRepository) SupplierService {
	return &supplierService{
		repo: repo,
	}
}

// Request/Response types

type CreateSupplierRequest struct {
	Code         string
	Name         string
	Email        string
	Phone        string
	Address      string
	City         string
	State        string
	PostalCode   string
	Country      string
	TaxID        string
	PaymentTerms string
	Currency     string
	Notes        string
	IsActive     bool
	Metadata     map[string]interface{}
}

type UpdateSupplierRequest struct {
	Name         *string
	Email        *string
	Phone        *string
	Address      *string
	City         *string
	State        *string
	PostalCode   *string
	Country      *string
	TaxID        *string
	PaymentTerms *string
	Currency     *string
	Notes        *string
	IsActive     *bool
	Metadata     map[string]interface{}
}

// CreateSupplier creates a new supplier with business rule validations
func (s *supplierService) CreateSupplier(ctx context.Context, clientID uuid.UUID, req *CreateSupplierRequest) (*models.Supplier, error) {
	// Business rule: Validate email format
	if req.Email != "" {
		if _, err := mail.ParseAddress(req.Email); err != nil {
			return nil, fmt.Errorf("invalid email format: %s", req.Email)
		}
	}

	// Business rule: Supplier code unique per client
	existingSupplier, err := s.repo.GetByCode(ctx, clientID, req.Code)
	if err == nil && existingSupplier != nil {
		return nil, fmt.Errorf("supplier code '%s' already exists for this client", req.Code)
	}

	supplier := &models.Supplier{
		ClientID:     clientID,
		Code:         req.Code,
		Name:         req.Name,
		Email:        req.Email,
		Phone:        req.Phone,
		Address:      req.Address,
		City:         req.City,
		State:        req.State,
		PostalCode:   req.PostalCode,
		Country:      req.Country,
		TaxID:        req.TaxID,
		PaymentTerms: req.PaymentTerms,
		Currency:     req.Currency,
		Notes:        req.Notes,
		IsActive:     req.IsActive,
		Metadata:     req.Metadata,
	}

	if err := s.repo.Create(ctx, supplier); err != nil {
		return nil, fmt.Errorf("failed to create supplier: %w", err)
	}

	return supplier, nil
}

// GetSupplier retrieves a supplier by ID
func (s *supplierService) GetSupplier(ctx context.Context, clientID uuid.UUID, id uuid.UUID) (*models.Supplier, error) {
	supplier, err := s.repo.GetByID(ctx, clientID, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get supplier: %w", err)
	}

	return supplier, nil
}

// UpdateSupplier updates an existing supplier
func (s *supplierService) UpdateSupplier(ctx context.Context, clientID uuid.UUID, id uuid.UUID, req *UpdateSupplierRequest) error {
	// Get existing supplier
	supplier, err := s.repo.GetByID(ctx, clientID, id)
	if err != nil {
		return fmt.Errorf("failed to get supplier: %w", err)
	}

	// Business rule: Validate email format if being updated
	if req.Email != nil && *req.Email != "" {
		if _, err := mail.ParseAddress(*req.Email); err != nil {
			return fmt.Errorf("invalid email format: %s", *req.Email)
		}
	}

	// Apply updates
	if req.Name != nil {
		supplier.Name = *req.Name
	}
	if req.Email != nil {
		supplier.Email = *req.Email
	}
	if req.Phone != nil {
		supplier.Phone = *req.Phone
	}
	if req.Address != nil {
		supplier.Address = *req.Address
	}
	if req.City != nil {
		supplier.City = *req.City
	}
	if req.State != nil {
		supplier.State = *req.State
	}
	if req.PostalCode != nil {
		supplier.PostalCode = *req.PostalCode
	}
	if req.Country != nil {
		supplier.Country = *req.Country
	}
	if req.TaxID != nil {
		supplier.TaxID = *req.TaxID
	}
	if req.PaymentTerms != nil {
		supplier.PaymentTerms = *req.PaymentTerms
	}
	if req.Currency != nil {
		supplier.Currency = *req.Currency
	}
	if req.Notes != nil {
		supplier.Notes = *req.Notes
	}
	if req.IsActive != nil {
		supplier.IsActive = *req.IsActive
	}
	if req.Metadata != nil {
		supplier.Metadata = req.Metadata
	}

	if err := s.repo.Update(ctx, supplier); err != nil {
		return fmt.Errorf("failed to update supplier: %w", err)
	}

	return nil
}

// DeleteSupplier deletes a supplier with business rule validation
func (s *supplierService) DeleteSupplier(ctx context.Context, clientID uuid.UUID, id uuid.UUID) error {
	// Business rule: Cannot delete supplier with active purchase orders
	// Note: This would require checking the PurchaseOrderRepository
	// For now, we rely on database foreign key constraints to prevent deletion
	// The database will return an error if there are related purchase orders

	if err := s.repo.Delete(ctx, clientID, id); err != nil {
		return fmt.Errorf("failed to delete supplier: %w", err)
	}

	return nil
}

// ListSuppliers retrieves suppliers with pagination
func (s *supplierService) ListSuppliers(ctx context.Context, clientID uuid.UUID, page, pageSize int) ([]*models.Supplier, int, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	offset := (page - 1) * pageSize

	suppliers, err := s.repo.ListByClient(ctx, clientID, pageSize, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list suppliers: %w", err)
	}

	// Note: Total count not implemented in repository yet, returning 0 for now
	total := len(suppliers)

	return suppliers, total, nil
}

// ListActiveSuppliers retrieves only active suppliers with pagination
func (s *supplierService) ListActiveSuppliers(ctx context.Context, clientID uuid.UUID, page, pageSize int) ([]*models.Supplier, int, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	offset := (page - 1) * pageSize

	suppliers, err := s.repo.ListActive(ctx, clientID, pageSize, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list active suppliers: %w", err)
	}

	// Note: Total count not implemented in repository yet, returning 0 for now
	total := len(suppliers)

	return suppliers, total, nil
}
