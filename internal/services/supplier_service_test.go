package services

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yourorg/shoppilot/internal/models"
	"github.com/yourorg/shoppilot/internal/services/fakes"
)

func TestSupplierService_CreateSupplier_Success(t *testing.T) {
	fakeRepo := &fakes.FakeSupplierRepository{}
	service := NewSupplierService(fakeRepo)

	clientID := uuid.New()

	// Configure fake behavior - code doesn't exist yet
	fakeRepo.GetByCodeReturns(nil, errors.New("not found"))
	fakeRepo.CreateReturns(nil)

	req := &CreateSupplierRequest{
		Code:     "SUP001",
		Name:     "Test Supplier",
		Email:    "test@supplier.com",
		Phone:    "+1234567890",
		IsActive: true,
	}

	supplier, err := service.CreateSupplier(context.Background(), clientID, req)
	require.NoError(t, err)
	assert.NotNil(t, supplier)
	assert.Equal(t, "SUP001", supplier.Code)
	assert.Equal(t, "Test Supplier", supplier.Name)
	assert.Equal(t, "test@supplier.com", supplier.Email)

	// Verify Create was called
	assert.Equal(t, 1, fakeRepo.CreateCallCount())
}

func TestSupplierService_CreateSupplier_InvalidEmail(t *testing.T) {
	fakeRepo := &fakes.FakeSupplierRepository{}
	service := NewSupplierService(fakeRepo)

	clientID := uuid.New()

	req := &CreateSupplierRequest{
		Code:     "SUP001",
		Name:     "Test Supplier",
		Email:    "invalid-email",
		IsActive: true,
	}

	supplier, err := service.CreateSupplier(context.Background(), clientID, req)
	require.Error(t, err)
	assert.Nil(t, supplier)
	assert.Contains(t, err.Error(), "invalid email format")

	// Create should not be called if validation fails
	assert.Equal(t, 0, fakeRepo.CreateCallCount())
}

func TestSupplierService_CreateSupplier_DuplicateCode(t *testing.T) {
	fakeRepo := &fakes.FakeSupplierRepository{}
	service := NewSupplierService(fakeRepo)

	clientID := uuid.New()
	existingSupplier := &models.Supplier{
		ID:       uuid.New(),
		ClientID: clientID,
		Code:     "SUP001",
		Name:     "Existing Supplier",
	}

	// Configure fake - supplier with this code already exists
	fakeRepo.GetByCodeReturns(existingSupplier, nil)

	req := &CreateSupplierRequest{
		Code:     "SUP001",
		Name:     "New Supplier",
		Email:    "test@supplier.com",
		IsActive: true,
	}

	supplier, err := service.CreateSupplier(context.Background(), clientID, req)
	require.Error(t, err)
	assert.Nil(t, supplier)
	assert.Contains(t, err.Error(), "already exists")

	// Create should not be called if code already exists
	assert.Equal(t, 0, fakeRepo.CreateCallCount())
}

func TestSupplierService_GetSupplier_Success(t *testing.T) {
	fakeRepo := &fakes.FakeSupplierRepository{}
	service := NewSupplierService(fakeRepo)

	clientID := uuid.New()
	supplierID := uuid.New()

	expectedSupplier := &models.Supplier{
		ID:       supplierID,
		ClientID: clientID,
		Code:     "SUP001",
		Name:     "Test Supplier",
		Email:    "test@supplier.com",
	}

	fakeRepo.GetByIDReturns(expectedSupplier, nil)

	supplier, err := service.GetSupplier(context.Background(), clientID, supplierID)
	require.NoError(t, err)
	assert.Equal(t, expectedSupplier, supplier)
	assert.Equal(t, 1, fakeRepo.GetByIDCallCount())
}

func TestSupplierService_GetSupplier_NotFound(t *testing.T) {
	fakeRepo := &fakes.FakeSupplierRepository{}
	service := NewSupplierService(fakeRepo)

	clientID := uuid.New()
	supplierID := uuid.New()

	fakeRepo.GetByIDReturns(nil, errors.New("supplier not found"))

	supplier, err := service.GetSupplier(context.Background(), clientID, supplierID)
	require.Error(t, err)
	assert.Nil(t, supplier)
	assert.Contains(t, err.Error(), "failed to get supplier")
}

func TestSupplierService_UpdateSupplier_Success(t *testing.T) {
	fakeRepo := &fakes.FakeSupplierRepository{}
	service := NewSupplierService(fakeRepo)

	clientID := uuid.New()
	supplierID := uuid.New()

	existingSupplier := &models.Supplier{
		ID:       supplierID,
		ClientID: clientID,
		Code:     "SUP001",
		Name:     "Original Name",
		Email:    "original@supplier.com",
		IsActive: true,
	}

	fakeRepo.GetByIDReturns(existingSupplier, nil)
	fakeRepo.UpdateReturns(nil)

	newName := "Updated Name"
	newEmail := "updated@supplier.com"
	req := &UpdateSupplierRequest{
		Name:  &newName,
		Email: &newEmail,
	}

	err := service.UpdateSupplier(context.Background(), clientID, supplierID, req)
	require.NoError(t, err)
	assert.Equal(t, 1, fakeRepo.UpdateCallCount())
}

func TestSupplierService_UpdateSupplier_InvalidEmail(t *testing.T) {
	fakeRepo := &fakes.FakeSupplierRepository{}
	service := NewSupplierService(fakeRepo)

	clientID := uuid.New()
	supplierID := uuid.New()

	existingSupplier := &models.Supplier{
		ID:       supplierID,
		ClientID: clientID,
		Code:     "SUP001",
		Name:     "Test Supplier",
		Email:    "test@supplier.com",
	}

	fakeRepo.GetByIDReturns(existingSupplier, nil)

	invalidEmail := "not-an-email"
	req := &UpdateSupplierRequest{
		Email: &invalidEmail,
	}

	err := service.UpdateSupplier(context.Background(), clientID, supplierID, req)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid email format")

	// Update should not be called if validation fails
	assert.Equal(t, 0, fakeRepo.UpdateCallCount())
}

func TestSupplierService_DeleteSupplier_Success(t *testing.T) {
	fakeRepo := &fakes.FakeSupplierRepository{}
	service := NewSupplierService(fakeRepo)

	clientID := uuid.New()
	supplierID := uuid.New()

	fakeRepo.DeleteReturns(nil)

	err := service.DeleteSupplier(context.Background(), clientID, supplierID)
	require.NoError(t, err)
	assert.Equal(t, 1, fakeRepo.DeleteCallCount())
}

func TestSupplierService_DeleteSupplier_WithActivePurchaseOrders(t *testing.T) {
	fakeRepo := &fakes.FakeSupplierRepository{}
	service := NewSupplierService(fakeRepo)

	clientID := uuid.New()
	supplierID := uuid.New()

	// Simulate database FK constraint error
	fakeRepo.DeleteReturns(errors.New("violates foreign key constraint"))

	err := service.DeleteSupplier(context.Background(), clientID, supplierID)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to delete supplier")
}

func TestSupplierService_ListSuppliers_Success(t *testing.T) {
	fakeRepo := &fakes.FakeSupplierRepository{}
	service := NewSupplierService(fakeRepo)

	clientID := uuid.New()

	expectedSuppliers := []*models.Supplier{
		{ID: uuid.New(), ClientID: clientID, Code: "SUP001", Name: "Supplier 1"},
		{ID: uuid.New(), ClientID: clientID, Code: "SUP002", Name: "Supplier 2"},
		{ID: uuid.New(), ClientID: clientID, Code: "SUP003", Name: "Supplier 3"},
	}

	fakeRepo.ListByClientReturns(expectedSuppliers, nil)

	suppliers, total, err := service.ListSuppliers(context.Background(), clientID, 1, 20)
	require.NoError(t, err)
	assert.Equal(t, expectedSuppliers, suppliers)
	assert.Equal(t, 3, total)
	assert.Equal(t, 1, fakeRepo.ListByClientCallCount())
}

func TestSupplierService_ListSuppliers_PaginationDefaults(t *testing.T) {
	fakeRepo := &fakes.FakeSupplierRepository{}
	service := NewSupplierService(fakeRepo)

	clientID := uuid.New()
	fakeRepo.ListByClientReturns([]*models.Supplier{}, nil)

	// Test invalid page (< 1)
	_, _, err := service.ListSuppliers(context.Background(), clientID, 0, 20)
	require.NoError(t, err)

	// Verify page was corrected to 1 (offset = 0)
	_, _, limit, offset := fakeRepo.ListByClientArgsForCall(0)
	assert.Equal(t, 20, limit)
	assert.Equal(t, 0, offset)

	// Test invalid pageSize (> 100)
	_, _, err = service.ListSuppliers(context.Background(), clientID, 1, 150)
	require.NoError(t, err)

	// Verify pageSize was corrected to 20
	_, _, limit, _ = fakeRepo.ListByClientArgsForCall(1)
	assert.Equal(t, 20, limit)
}

func TestSupplierService_ListActiveSuppliers_Success(t *testing.T) {
	fakeRepo := &fakes.FakeSupplierRepository{}
	service := NewSupplierService(fakeRepo)

	clientID := uuid.New()

	expectedSuppliers := []*models.Supplier{
		{ID: uuid.New(), ClientID: clientID, Code: "SUP001", Name: "Active Supplier 1", IsActive: true},
		{ID: uuid.New(), ClientID: clientID, Code: "SUP002", Name: "Active Supplier 2", IsActive: true},
	}

	fakeRepo.ListActiveReturns(expectedSuppliers, nil)

	suppliers, total, err := service.ListActiveSuppliers(context.Background(), clientID, 1, 20)
	require.NoError(t, err)
	assert.Equal(t, expectedSuppliers, suppliers)
	assert.Equal(t, 2, total)
	assert.Equal(t, 1, fakeRepo.ListActiveCallCount())
}
