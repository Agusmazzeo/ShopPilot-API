package services

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/yourorg/shoppilot/internal/models"
)

// TestCreateClient_Success tests successful client creation
func TestCreateClient_Success(t *testing.T) {
	mockRepo := new(MockClientRepository)
	service := NewClientService(mockRepo)
	ctx := context.Background()

	req := &CreateClientRequest{
		Name:         "Test Client",
		Description:  "Test Description",
		ContactEmail: "test@example.com",
		ContactPhone: "+1234567890",
		WebsiteURL:   "https://test.com",
		LogoURL:      nil,
	}

	// Mock GetBySlug to return not found (slug doesn't exist)
	mockRepo.On("GetBySlug", ctx, "test-client").Return(nil, errors.New("not found"))

	// Mock Create to succeed
	mockRepo.On("Create", ctx, mock.MatchedBy(func(c *models.Client) bool {
		return c.Name == req.Name &&
			c.Slug == "test-client" &&
			c.Description == req.Description &&
			c.ContactEmail == req.ContactEmail &&
			c.ContactPhone == req.ContactPhone &&
			c.WebsiteURL == req.WebsiteURL &&
			c.IsActive == true
	})).Return(nil)

	client, err := service.CreateClient(ctx, req)

	assert.NoError(t, err)
	assert.NotNil(t, client)
	assert.Equal(t, "Test Client", client.Name)
	assert.Equal(t, "test-client", client.Slug)
	assert.Equal(t, "test@example.com", client.ContactEmail)
	assert.True(t, client.IsActive)
	mockRepo.AssertExpectations(t)
}

// TestCreateClient_InvalidEmail tests client creation with invalid email
func TestCreateClient_InvalidEmail(t *testing.T) {
	mockRepo := new(MockClientRepository)
	service := NewClientService(mockRepo)
	ctx := context.Background()

	req := &CreateClientRequest{
		Name:         "Test Client",
		Description:  "Test Description",
		ContactEmail: "invalid-email",
		ContactPhone: "",
		WebsiteURL:   "",
	}

	client, err := service.CreateClient(ctx, req)

	assert.Error(t, err)
	assert.Nil(t, client)
	assert.Contains(t, err.Error(), "invalid email")
	mockRepo.AssertNotCalled(t, "Create")
}

// TestCreateClient_InvalidPhone tests client creation with invalid phone
func TestCreateClient_InvalidPhone(t *testing.T) {
	mockRepo := new(MockClientRepository)
	service := NewClientService(mockRepo)
	ctx := context.Background()

	req := &CreateClientRequest{
		Name:         "Test Client",
		Description:  "Test Description",
		ContactEmail: "test@example.com",
		ContactPhone: "123", // Too short
		WebsiteURL:   "",
	}

	client, err := service.CreateClient(ctx, req)

	assert.Error(t, err)
	assert.Nil(t, client)
	assert.Contains(t, err.Error(), "invalid phone")
	mockRepo.AssertNotCalled(t, "Create")
}

// TestCreateClient_DuplicateSlug tests client creation when slug already exists
func TestCreateClient_DuplicateSlug(t *testing.T) {
	mockRepo := new(MockClientRepository)
	service := NewClientService(mockRepo)
	ctx := context.Background()

	existingClient := &models.Client{
		ID:   uuid.New(),
		Name: "Test Client",
		Slug: "test-client",
	}

	req := &CreateClientRequest{
		Name:         "Test Client",
		Description:  "Test Description",
		ContactEmail: "test@example.com",
		ContactPhone: "",
		WebsiteURL:   "",
	}

	// Mock GetBySlug to return existing client
	mockRepo.On("GetBySlug", ctx, "test-client").Return(existingClient, nil)

	client, err := service.CreateClient(ctx, req)

	assert.Error(t, err)
	assert.Nil(t, client)
	assert.Contains(t, err.Error(), "already exists")
	mockRepo.AssertNotCalled(t, "Create")
	mockRepo.AssertExpectations(t)
}

// TestCreateClient_RepositoryError tests client creation when repository fails
func TestCreateClient_RepositoryError(t *testing.T) {
	mockRepo := new(MockClientRepository)
	service := NewClientService(mockRepo)
	ctx := context.Background()

	req := &CreateClientRequest{
		Name:         "Test Client",
		ContactEmail: "test@example.com",
	}

	mockRepo.On("GetBySlug", ctx, "test-client").Return(nil, errors.New("not found"))
	mockRepo.On("Create", ctx, mock.Anything).Return(errors.New("database error"))

	client, err := service.CreateClient(ctx, req)

	assert.Error(t, err)
	assert.Nil(t, client)
	assert.Contains(t, err.Error(), "failed to create client")
	mockRepo.AssertExpectations(t)
}

// TestGetClient_Success tests successful retrieval of a client by ID
func TestGetClient_Success(t *testing.T) {
	mockRepo := new(MockClientRepository)
	service := NewClientService(mockRepo)
	ctx := context.Background()

	clientID := uuid.New()
	expectedClient := &models.Client{
		ID:           clientID,
		Name:         "Test Client",
		Slug:         "test-client",
		ContactEmail: "test@example.com",
		IsActive:     true,
	}

	mockRepo.On("GetByID", ctx, clientID).Return(expectedClient, nil)

	client, err := service.GetClient(ctx, clientID)

	assert.NoError(t, err)
	assert.NotNil(t, client)
	assert.Equal(t, clientID, client.ID)
	assert.Equal(t, "Test Client", client.Name)
	mockRepo.AssertExpectations(t)
}

// TestGetClient_NotFound tests retrieval when client doesn't exist
func TestGetClient_NotFound(t *testing.T) {
	mockRepo := new(MockClientRepository)
	service := NewClientService(mockRepo)
	ctx := context.Background()

	clientID := uuid.New()
	mockRepo.On("GetByID", ctx, clientID).Return(nil, errors.New("not found"))

	client, err := service.GetClient(ctx, clientID)

	assert.Error(t, err)
	assert.Nil(t, client)
	assert.Contains(t, err.Error(), "failed to get client")
	mockRepo.AssertExpectations(t)
}

// TestGetClientBySlug_Success tests successful retrieval of a client by slug
func TestGetClientBySlug_Success(t *testing.T) {
	mockRepo := new(MockClientRepository)
	service := NewClientService(mockRepo)
	ctx := context.Background()

	expectedClient := &models.Client{
		ID:           uuid.New(),
		Name:         "Test Client",
		Slug:         "test-client",
		ContactEmail: "test@example.com",
		IsActive:     true,
	}

	mockRepo.On("GetBySlug", ctx, "test-client").Return(expectedClient, nil)

	client, err := service.GetClientBySlug(ctx, "test-client")

	assert.NoError(t, err)
	assert.NotNil(t, client)
	assert.Equal(t, "test-client", client.Slug)
	assert.Equal(t, "Test Client", client.Name)
	mockRepo.AssertExpectations(t)
}

// TestGetClientBySlug_NotFound tests retrieval when slug doesn't exist
func TestGetClientBySlug_NotFound(t *testing.T) {
	mockRepo := new(MockClientRepository)
	service := NewClientService(mockRepo)
	ctx := context.Background()

	mockRepo.On("GetBySlug", ctx, "nonexistent").Return(nil, errors.New("not found"))

	client, err := service.GetClientBySlug(ctx, "nonexistent")

	assert.Error(t, err)
	assert.Nil(t, client)
	assert.Contains(t, err.Error(), "failed to get client by slug")
	mockRepo.AssertExpectations(t)
}

// TestUpdateClient_Success tests successful client update
func TestUpdateClient_Success(t *testing.T) {
	mockRepo := new(MockClientRepository)
	service := NewClientService(mockRepo)
	ctx := context.Background()

	clientID := uuid.New()
	existingClient := &models.Client{
		ID:           clientID,
		Name:         "Old Name",
		Slug:         "old-name",
		ContactEmail: "old@example.com",
		IsActive:     true,
	}

	newName := "New Name"
	newEmail := "new@example.com"
	req := &UpdateClientRequest{
		Name:         &newName,
		ContactEmail: &newEmail,
	}

	mockRepo.On("GetByID", ctx, clientID).Return(existingClient, nil)
	mockRepo.On("GetBySlug", ctx, "new-name").Return(nil, errors.New("not found"))
	mockRepo.On("Update", ctx, mock.MatchedBy(func(c *models.Client) bool {
		return c.ID == clientID &&
			c.Name == "New Name" &&
			c.Slug == "new-name" &&
			c.ContactEmail == "new@example.com"
	})).Return(nil)

	err := service.UpdateClient(ctx, clientID, req)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

// TestUpdateClient_InvalidEmail tests update with invalid email
func TestUpdateClient_InvalidEmail(t *testing.T) {
	mockRepo := new(MockClientRepository)
	service := NewClientService(mockRepo)
	ctx := context.Background()

	clientID := uuid.New()
	existingClient := &models.Client{
		ID:           clientID,
		Name:         "Test Client",
		Slug:         "test-client",
		ContactEmail: "old@example.com",
	}

	invalidEmail := "invalid-email"
	req := &UpdateClientRequest{
		ContactEmail: &invalidEmail,
	}

	mockRepo.On("GetByID", ctx, clientID).Return(existingClient, nil)

	err := service.UpdateClient(ctx, clientID, req)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid email")
	mockRepo.AssertNotCalled(t, "Update")
	mockRepo.AssertExpectations(t)
}

// TestUpdateClient_InvalidPhone tests update with invalid phone
func TestUpdateClient_InvalidPhone(t *testing.T) {
	mockRepo := new(MockClientRepository)
	service := NewClientService(mockRepo)
	ctx := context.Background()

	clientID := uuid.New()
	existingClient := &models.Client{
		ID:           clientID,
		Name:         "Test Client",
		Slug:         "test-client",
		ContactEmail: "test@example.com",
	}

	invalidPhone := "12"
	req := &UpdateClientRequest{
		ContactPhone: &invalidPhone,
	}

	mockRepo.On("GetByID", ctx, clientID).Return(existingClient, nil)

	err := service.UpdateClient(ctx, clientID, req)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid phone")
	mockRepo.AssertNotCalled(t, "Update")
	mockRepo.AssertExpectations(t)
}

// TestUpdateClient_DuplicateSlug tests update when new slug already exists
func TestUpdateClient_DuplicateSlug(t *testing.T) {
	mockRepo := new(MockClientRepository)
	service := NewClientService(mockRepo)
	ctx := context.Background()

	clientID := uuid.New()
	existingClient := &models.Client{
		ID:   clientID,
		Name: "Old Name",
		Slug: "old-name",
	}

	otherClient := &models.Client{
		ID:   uuid.New(),
		Name: "New Name",
		Slug: "new-name",
	}

	newName := "New Name"
	req := &UpdateClientRequest{
		Name: &newName,
	}

	mockRepo.On("GetByID", ctx, clientID).Return(existingClient, nil)
	mockRepo.On("GetBySlug", ctx, "new-name").Return(otherClient, nil)

	err := service.UpdateClient(ctx, clientID, req)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already exists")
	mockRepo.AssertNotCalled(t, "Update")
	mockRepo.AssertExpectations(t)
}

// TestUpdateClient_NotFound tests update when client doesn't exist
func TestUpdateClient_NotFound(t *testing.T) {
	mockRepo := new(MockClientRepository)
	service := NewClientService(mockRepo)
	ctx := context.Background()

	clientID := uuid.New()
	newName := "New Name"
	req := &UpdateClientRequest{
		Name: &newName,
	}

	mockRepo.On("GetByID", ctx, clientID).Return(nil, errors.New("not found"))

	err := service.UpdateClient(ctx, clientID, req)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "client not found")
	mockRepo.AssertNotCalled(t, "Update")
	mockRepo.AssertExpectations(t)
}

// TestDeleteClient_Success tests successful soft delete
func TestDeleteClient_Success(t *testing.T) {
	mockRepo := new(MockClientRepository)
	service := NewClientService(mockRepo)
	ctx := context.Background()

	clientID := uuid.New()
	existingClient := &models.Client{
		ID:       clientID,
		Name:     "Test Client",
		Slug:     "test-client",
		IsActive: true,
	}

	mockRepo.On("GetByID", ctx, clientID).Return(existingClient, nil)
	mockRepo.On("Update", ctx, mock.MatchedBy(func(c *models.Client) bool {
		return c.ID == clientID && c.IsActive == false
	})).Return(nil)

	err := service.DeleteClient(ctx, clientID)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

// TestDeleteClient_NotFound tests delete when client doesn't exist
func TestDeleteClient_NotFound(t *testing.T) {
	mockRepo := new(MockClientRepository)
	service := NewClientService(mockRepo)
	ctx := context.Background()

	clientID := uuid.New()
	mockRepo.On("GetByID", ctx, clientID).Return(nil, errors.New("not found"))

	err := service.DeleteClient(ctx, clientID)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "client not found")
	mockRepo.AssertNotCalled(t, "Update")
	mockRepo.AssertExpectations(t)
}

// TestDeleteClient_RepositoryError tests delete when update fails
func TestDeleteClient_RepositoryError(t *testing.T) {
	mockRepo := new(MockClientRepository)
	service := NewClientService(mockRepo)
	ctx := context.Background()

	clientID := uuid.New()
	existingClient := &models.Client{
		ID:       clientID,
		Name:     "Test Client",
		IsActive: true,
	}

	mockRepo.On("GetByID", ctx, clientID).Return(existingClient, nil)
	mockRepo.On("Update", ctx, mock.Anything).Return(errors.New("database error"))

	err := service.DeleteClient(ctx, clientID)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to delete client")
	mockRepo.AssertExpectations(t)
}

// TestListClients_Success tests successful listing with pagination
func TestListClients_Success(t *testing.T) {
	mockRepo := new(MockClientRepository)
	service := NewClientService(mockRepo)
	ctx := context.Background()

	expectedClients := []*models.Client{
		{
			ID:   uuid.New(),
			Name: "Client 1",
			Slug: "client-1",
		},
		{
			ID:   uuid.New(),
			Name: "Client 2",
			Slug: "client-2",
		},
	}

	mockRepo.On("List", ctx, 10, 0).Return(expectedClients, nil)

	clients, total, err := service.ListClients(ctx, 1, 10)

	assert.NoError(t, err)
	assert.NotNil(t, clients)
	assert.Len(t, clients, 2)
	assert.Equal(t, 2, total)
	mockRepo.AssertExpectations(t)
}

// TestListClients_WithPagination tests listing with custom pagination
func TestListClients_WithPagination(t *testing.T) {
	mockRepo := new(MockClientRepository)
	service := NewClientService(mockRepo)
	ctx := context.Background()

	expectedClients := []*models.Client{
		{ID: uuid.New(), Name: "Client 3"},
		{ID: uuid.New(), Name: "Client 4"},
		{ID: uuid.New(), Name: "Client 5"},
	}

	// Page 2, pageSize 3 = offset 3
	mockRepo.On("List", ctx, 3, 3).Return(expectedClients, nil)

	clients, _, err := service.ListClients(ctx, 2, 3)

	assert.NoError(t, err)
	assert.NotNil(t, clients)
	assert.Len(t, clients, 3)
	mockRepo.AssertExpectations(t)
}

// TestListClients_InvalidPagination tests pagination parameter validation
func TestListClients_InvalidPagination(t *testing.T) {
	mockRepo := new(MockClientRepository)
	service := NewClientService(mockRepo)
	ctx := context.Background()

	expectedClients := []*models.Client{}

	// Invalid page (0) should default to 1, invalid pageSize (0) should default to 10
	mockRepo.On("List", ctx, 10, 0).Return(expectedClients, nil)

	clients, _, err := service.ListClients(ctx, 0, 0)

	assert.NoError(t, err)
	assert.NotNil(t, clients)
	mockRepo.AssertExpectations(t)
}

// TestListClients_MaxPageSize tests page size capping at 100
func TestListClients_MaxPageSize(t *testing.T) {
	mockRepo := new(MockClientRepository)
	service := NewClientService(mockRepo)
	ctx := context.Background()

	expectedClients := []*models.Client{}

	// PageSize > 100 should be capped at 100
	mockRepo.On("List", ctx, 100, 0).Return(expectedClients, nil)

	clients, _, err := service.ListClients(ctx, 1, 200)

	assert.NoError(t, err)
	assert.NotNil(t, clients)
	mockRepo.AssertExpectations(t)
}

// TestListClients_RepositoryError tests listing when repository fails
func TestListClients_RepositoryError(t *testing.T) {
	mockRepo := new(MockClientRepository)
	service := NewClientService(mockRepo)
	ctx := context.Background()

	mockRepo.On("List", ctx, 10, 0).Return(nil, errors.New("database error"))

	clients, total, err := service.ListClients(ctx, 1, 10)

	assert.Error(t, err)
	assert.Nil(t, clients)
	assert.Equal(t, 0, total)
	assert.Contains(t, err.Error(), "failed to list clients")
	mockRepo.AssertExpectations(t)
}

// TestActivateClient_Success tests successful client activation
func TestActivateClient_Success(t *testing.T) {
	mockRepo := new(MockClientRepository)
	service := NewClientService(mockRepo)
	ctx := context.Background()

	clientID := uuid.New()
	existingClient := &models.Client{
		ID:       clientID,
		Name:     "Test Client",
		Slug:     "test-client",
		IsActive: false,
	}

	mockRepo.On("GetByID", ctx, clientID).Return(existingClient, nil)
	mockRepo.On("Update", ctx, mock.MatchedBy(func(c *models.Client) bool {
		return c.ID == clientID && c.IsActive == true
	})).Return(nil)

	err := service.ActivateClient(ctx, clientID)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

// TestActivateClient_NotFound tests activation when client doesn't exist
func TestActivateClient_NotFound(t *testing.T) {
	mockRepo := new(MockClientRepository)
	service := NewClientService(mockRepo)
	ctx := context.Background()

	clientID := uuid.New()
	mockRepo.On("GetByID", ctx, clientID).Return(nil, errors.New("not found"))

	err := service.ActivateClient(ctx, clientID)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "client not found")
	mockRepo.AssertNotCalled(t, "Update")
	mockRepo.AssertExpectations(t)
}

// TestActivateClient_RepositoryError tests activation when update fails
func TestActivateClient_RepositoryError(t *testing.T) {
	mockRepo := new(MockClientRepository)
	service := NewClientService(mockRepo)
	ctx := context.Background()

	clientID := uuid.New()
	existingClient := &models.Client{
		ID:       clientID,
		IsActive: false,
	}

	mockRepo.On("GetByID", ctx, clientID).Return(existingClient, nil)
	mockRepo.On("Update", ctx, mock.Anything).Return(errors.New("database error"))

	err := service.ActivateClient(ctx, clientID)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to activate client")
	mockRepo.AssertExpectations(t)
}

// TestDeactivateClient_Success tests successful client deactivation
func TestDeactivateClient_Success(t *testing.T) {
	mockRepo := new(MockClientRepository)
	service := NewClientService(mockRepo)
	ctx := context.Background()

	clientID := uuid.New()
	existingClient := &models.Client{
		ID:       clientID,
		Name:     "Test Client",
		Slug:     "test-client",
		IsActive: true,
	}

	mockRepo.On("GetByID", ctx, clientID).Return(existingClient, nil)
	mockRepo.On("Update", ctx, mock.MatchedBy(func(c *models.Client) bool {
		return c.ID == clientID && c.IsActive == false
	})).Return(nil)

	err := service.DeactivateClient(ctx, clientID)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

// TestDeactivateClient_NotFound tests deactivation when client doesn't exist
func TestDeactivateClient_NotFound(t *testing.T) {
	mockRepo := new(MockClientRepository)
	service := NewClientService(mockRepo)
	ctx := context.Background()

	clientID := uuid.New()
	mockRepo.On("GetByID", ctx, clientID).Return(nil, errors.New("not found"))

	err := service.DeactivateClient(ctx, clientID)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "client not found")
	mockRepo.AssertNotCalled(t, "Update")
	mockRepo.AssertExpectations(t)
}

// TestDeactivateClient_RepositoryError tests deactivation when update fails
func TestDeactivateClient_RepositoryError(t *testing.T) {
	mockRepo := new(MockClientRepository)
	service := NewClientService(mockRepo)
	ctx := context.Background()

	clientID := uuid.New()
	existingClient := &models.Client{
		ID:       clientID,
		IsActive: true,
	}

	mockRepo.On("GetByID", ctx, clientID).Return(existingClient, nil)
	mockRepo.On("Update", ctx, mock.Anything).Return(errors.New("database error"))

	err := service.DeactivateClient(ctx, clientID)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to deactivate client")
	mockRepo.AssertExpectations(t)
}
