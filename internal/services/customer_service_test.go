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

func TestCustomerService_CreateCustomer_Success(t *testing.T) {
	fakeRepo := &fakes.FakeCustomerRepository{}
	service := NewCustomerService(fakeRepo)

	clientID := uuid.New()

	// Configure fake behavior - code doesn't exist yet
	fakeRepo.GetByCodeReturns(nil, errors.New("not found"))
	fakeRepo.CreateReturns(nil)

	req := &CreateCustomerRequest{
		Code:      "CUST001",
		FirstName: "John",
		LastName:  "Doe",
		Email:     "john.doe@example.com",
		Phone:     "+1234567890",
		IsActive:  true,
	}

	customer, err := service.CreateCustomer(context.Background(), clientID, req)
	require.NoError(t, err)
	assert.NotNil(t, customer)
	assert.Equal(t, "CUST001", customer.Code)
	assert.Equal(t, "John", customer.FirstName)
	assert.Equal(t, "Doe", customer.LastName)
	assert.Equal(t, "john.doe@example.com", customer.Email)

	// Verify Create was called
	assert.Equal(t, 1, fakeRepo.CreateCallCount())
}

func TestCustomerService_CreateCustomer_InvalidEmail(t *testing.T) {
	fakeRepo := &fakes.FakeCustomerRepository{}
	service := NewCustomerService(fakeRepo)

	clientID := uuid.New()

	req := &CreateCustomerRequest{
		Code:      "CUST001",
		FirstName: "John",
		LastName:  "Doe",
		Email:     "not-a-valid-email",
		IsActive:  true,
	}

	customer, err := service.CreateCustomer(context.Background(), clientID, req)
	require.Error(t, err)
	assert.Nil(t, customer)
	assert.Contains(t, err.Error(), "invalid email format")

	// Create should not be called if validation fails
	assert.Equal(t, 0, fakeRepo.CreateCallCount())
}

func TestCustomerService_CreateCustomer_DuplicateCode(t *testing.T) {
	fakeRepo := &fakes.FakeCustomerRepository{}
	service := NewCustomerService(fakeRepo)

	clientID := uuid.New()
	existingCustomer := &models.Customer{
		ID:        uuid.New(),
		ClientID:  clientID,
		Code:      "CUST001",
		FirstName: "Jane",
		LastName:  "Smith",
	}

	// Configure fake - customer with this code already exists
	fakeRepo.GetByCodeReturns(existingCustomer, nil)

	req := &CreateCustomerRequest{
		Code:      "CUST001",
		FirstName: "John",
		LastName:  "Doe",
		Email:     "john.doe@example.com",
		IsActive:  true,
	}

	customer, err := service.CreateCustomer(context.Background(), clientID, req)
	require.Error(t, err)
	assert.Nil(t, customer)
	assert.Contains(t, err.Error(), "already exists")

	// Create should not be called if code already exists
	assert.Equal(t, 0, fakeRepo.CreateCallCount())
}

func TestCustomerService_GetCustomer_Success(t *testing.T) {
	fakeRepo := &fakes.FakeCustomerRepository{}
	service := NewCustomerService(fakeRepo)

	clientID := uuid.New()
	customerID := uuid.New()

	expectedCustomer := &models.Customer{
		ID:        customerID,
		ClientID:  clientID,
		Code:      "CUST001",
		FirstName: "John",
		LastName:  "Doe",
		Email:     "john.doe@example.com",
	}

	fakeRepo.GetByIDReturns(expectedCustomer, nil)

	customer, err := service.GetCustomer(context.Background(), clientID, customerID)
	require.NoError(t, err)
	assert.Equal(t, expectedCustomer, customer)
	assert.Equal(t, 1, fakeRepo.GetByIDCallCount())
}

func TestCustomerService_GetCustomer_NotFound(t *testing.T) {
	fakeRepo := &fakes.FakeCustomerRepository{}
	service := NewCustomerService(fakeRepo)

	clientID := uuid.New()
	customerID := uuid.New()

	fakeRepo.GetByIDReturns(nil, errors.New("customer not found"))

	customer, err := service.GetCustomer(context.Background(), clientID, customerID)
	require.Error(t, err)
	assert.Nil(t, customer)
	assert.Contains(t, err.Error(), "failed to get customer")
}

func TestCustomerService_UpdateCustomer_Success(t *testing.T) {
	fakeRepo := &fakes.FakeCustomerRepository{}
	service := NewCustomerService(fakeRepo)

	clientID := uuid.New()
	customerID := uuid.New()

	existingCustomer := &models.Customer{
		ID:        customerID,
		ClientID:  clientID,
		Code:      "CUST001",
		FirstName: "John",
		LastName:  "Doe",
		Email:     "john.doe@example.com",
		IsActive:  true,
	}

	fakeRepo.GetByIDReturns(existingCustomer, nil)
	fakeRepo.UpdateReturns(nil)

	newFirstName := "Jane"
	newEmail := "jane.doe@example.com"
	req := &UpdateCustomerRequest{
		FirstName: &newFirstName,
		Email:     &newEmail,
	}

	err := service.UpdateCustomer(context.Background(), clientID, customerID, req)
	require.NoError(t, err)
	assert.Equal(t, 1, fakeRepo.UpdateCallCount())
}

func TestCustomerService_UpdateCustomer_InvalidEmail(t *testing.T) {
	fakeRepo := &fakes.FakeCustomerRepository{}
	service := NewCustomerService(fakeRepo)

	clientID := uuid.New()
	customerID := uuid.New()

	existingCustomer := &models.Customer{
		ID:        customerID,
		ClientID:  clientID,
		Code:      "CUST001",
		FirstName: "John",
		LastName:  "Doe",
		Email:     "john.doe@example.com",
	}

	fakeRepo.GetByIDReturns(existingCustomer, nil)

	invalidEmail := "invalid-email"
	req := &UpdateCustomerRequest{
		Email: &invalidEmail,
	}

	err := service.UpdateCustomer(context.Background(), clientID, customerID, req)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid email format")

	// Update should not be called if validation fails
	assert.Equal(t, 0, fakeRepo.UpdateCallCount())
}

func TestCustomerService_DeleteCustomer_Success(t *testing.T) {
	fakeRepo := &fakes.FakeCustomerRepository{}
	service := NewCustomerService(fakeRepo)

	clientID := uuid.New()
	customerID := uuid.New()

	fakeRepo.DeleteReturns(nil)

	err := service.DeleteCustomer(context.Background(), clientID, customerID)
	require.NoError(t, err)
	assert.Equal(t, 1, fakeRepo.DeleteCallCount())
}

func TestCustomerService_DeleteCustomer_WithActiveOrders(t *testing.T) {
	fakeRepo := &fakes.FakeCustomerRepository{}
	service := NewCustomerService(fakeRepo)

	clientID := uuid.New()
	customerID := uuid.New()

	// Simulate database FK constraint error
	fakeRepo.DeleteReturns(errors.New("violates foreign key constraint"))

	err := service.DeleteCustomer(context.Background(), clientID, customerID)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to delete customer")
}

func TestCustomerService_ListCustomers_Success(t *testing.T) {
	fakeRepo := &fakes.FakeCustomerRepository{}
	service := NewCustomerService(fakeRepo)

	clientID := uuid.New()

	expectedCustomers := []*models.Customer{
		{ID: uuid.New(), ClientID: clientID, Code: "CUST001", FirstName: "John", LastName: "Doe"},
		{ID: uuid.New(), ClientID: clientID, Code: "CUST002", FirstName: "Jane", LastName: "Smith"},
		{ID: uuid.New(), ClientID: clientID, Code: "CUST003", FirstName: "Bob", LastName: "Johnson"},
	}

	fakeRepo.ListByClientReturns(expectedCustomers, nil)

	customers, total, err := service.ListCustomers(context.Background(), clientID, 1, 20)
	require.NoError(t, err)
	assert.Equal(t, expectedCustomers, customers)
	assert.Equal(t, 3, total)
	assert.Equal(t, 1, fakeRepo.ListByClientCallCount())
}

func TestCustomerService_ListCustomers_PaginationDefaults(t *testing.T) {
	fakeRepo := &fakes.FakeCustomerRepository{}
	service := NewCustomerService(fakeRepo)

	clientID := uuid.New()
	fakeRepo.ListByClientReturns([]*models.Customer{}, nil)

	// Test invalid page (< 1)
	_, _, err := service.ListCustomers(context.Background(), clientID, 0, 20)
	require.NoError(t, err)

	// Verify page was corrected to 1 (offset = 0)
	_, _, limit, offset := fakeRepo.ListByClientArgsForCall(0)
	assert.Equal(t, 20, limit)
	assert.Equal(t, 0, offset)

	// Test invalid pageSize (> 100)
	_, _, err = service.ListCustomers(context.Background(), clientID, 1, 150)
	require.NoError(t, err)

	// Verify pageSize was corrected to 20
	_, _, limit, _ = fakeRepo.ListByClientArgsForCall(1)
	assert.Equal(t, 20, limit)
}

func TestCustomerService_SearchCustomers_Success(t *testing.T) {
	fakeRepo := &fakes.FakeCustomerRepository{}
	service := NewCustomerService(fakeRepo)

	clientID := uuid.New()

	expectedCustomers := []*models.Customer{
		{ID: uuid.New(), ClientID: clientID, Code: "CUST001", FirstName: "John", LastName: "Doe", Email: "john@example.com"},
		{ID: uuid.New(), ClientID: clientID, Code: "CUST002", FirstName: "Johnny", LastName: "Smith", Email: "johnny@example.com"},
	}

	fakeRepo.SearchReturns(expectedCustomers, nil)

	customers, total, err := service.SearchCustomers(context.Background(), clientID, "john", 1, 20)
	require.NoError(t, err)
	assert.Equal(t, expectedCustomers, customers)
	assert.Equal(t, 2, total)
	assert.Equal(t, 1, fakeRepo.SearchCallCount())

	// Verify search was called with correct parameters
	_, _, query, limit, offset := fakeRepo.SearchArgsForCall(0)
	assert.Equal(t, "john", query)
	assert.Equal(t, 20, limit)
	assert.Equal(t, 0, offset)
}

func TestCustomerService_SearchCustomers_PaginationDefaults(t *testing.T) {
	fakeRepo := &fakes.FakeCustomerRepository{}
	service := NewCustomerService(fakeRepo)

	clientID := uuid.New()
	fakeRepo.SearchReturns([]*models.Customer{}, nil)

	// Test invalid page (< 1)
	_, _, err := service.SearchCustomers(context.Background(), clientID, "test", 0, 20)
	require.NoError(t, err)

	// Verify page was corrected to 1 (offset = 0)
	_, _, _, limit, offset := fakeRepo.SearchArgsForCall(0)
	assert.Equal(t, 20, limit)
	assert.Equal(t, 0, offset)

	// Test invalid pageSize (> 100)
	_, _, err = service.SearchCustomers(context.Background(), clientID, "test", 1, 150)
	require.NoError(t, err)

	// Verify pageSize was corrected to 20
	_, _, _, limit, _ = fakeRepo.SearchArgsForCall(1)
	assert.Equal(t, 20, limit)
}
