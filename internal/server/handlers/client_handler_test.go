package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/yourorg/shoppilot/internal/models"
	"github.com/yourorg/shoppilot/internal/services"
)

// MockClientService is a mock implementation of services.ClientService
type MockClientService struct {
	mock.Mock
}

func (m *MockClientService) CreateClient(ctx context.Context, req *services.CreateClientRequest) (*models.Client, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Client), args.Error(1)
}

func (m *MockClientService) GetClient(ctx context.Context, id uuid.UUID) (*models.Client, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Client), args.Error(1)
}

func (m *MockClientService) GetClientBySlug(ctx context.Context, slug string) (*models.Client, error) {
	args := m.Called(ctx, slug)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Client), args.Error(1)
}

func (m *MockClientService) UpdateClient(ctx context.Context, id uuid.UUID, req *services.UpdateClientRequest) error {
	args := m.Called(ctx, id, req)
	return args.Error(0)
}

func (m *MockClientService) DeleteClient(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockClientService) ListClients(ctx context.Context, page, pageSize int) ([]*models.Client, int, error) {
	args := m.Called(ctx, page, pageSize)
	if args.Get(0) == nil {
		return nil, args.Int(1), args.Error(2)
	}
	return args.Get(0).([]*models.Client), args.Int(1), args.Error(2)
}

func (m *MockClientService) ActivateClient(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockClientService) DeactivateClient(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// Helper function to create a test client model
func createTestClient() *models.Client {
	logoURL := "https://example.com/logo.png"
	return &models.Client{
		ID:           uuid.New(),
		Name:         "Test Client",
		Slug:         "test-client",
		Description:  "A test client",
		ContactEmail: "test@example.com",
		ContactPhone: "+1234567890",
		WebsiteURL:   "https://example.com",
		LogoURL:      &logoURL,
		IsActive:     true,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
}

func TestClientHandler_Create(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    interface{}
		mockSetup      func(*MockClientService)
		expectedStatus int
		expectedError  *string
	}{
		{
			name: "successful creation",
			requestBody: services.CreateClientRequest{
				Name:         "New Client",
				ContactEmail: "new@example.com",
				ContactPhone: "+1234567890",
				WebsiteURL:   "https://example.com",
			},
			mockSetup: func(m *MockClientService) {
				client := createTestClient()
				client.Name = "New Client"
				client.ContactEmail = "new@example.com"
				m.On("CreateClient", mock.Anything, mock.AnythingOfType("*services.CreateClientRequest")).
					Return(client, nil)
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name:        "invalid request body",
			requestBody: "invalid json",
			mockSetup: func(m *MockClientService) {
				// No mock setup needed - should fail before service call
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "slug already exists",
			requestBody: services.CreateClientRequest{
				Name:         "Duplicate Client",
				ContactEmail: "dup@example.com",
			},
			mockSetup: func(m *MockClientService) {
				m.On("CreateClient", mock.Anything, mock.AnythingOfType("*services.CreateClientRequest")).
					Return(nil, fmt.Errorf("client with slug already exists"))
			},
			expectedStatus: http.StatusConflict,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			mockService := new(MockClientService)
			tt.mockSetup(mockService)
			handler := NewClientHandler(mockService)

			// Create request
			body, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest(http.MethodPost, "/api/v1/clients", bytes.NewBuffer(body))
			rec := httptest.NewRecorder()

			// Execute
			handler.Create(rec, req)

			// Assert
			assert.Equal(t, tt.expectedStatus, rec.Code)

			var response APIResponse
			json.NewDecoder(rec.Body).Decode(&response)

			if tt.expectedStatus >= 400 {
				assert.False(t, response.Success)
				assert.NotNil(t, response.Error)
			} else {
				assert.True(t, response.Success)
				assert.NotNil(t, response.Data)
			}

			mockService.AssertExpectations(t)
		})
	}
}

func TestClientHandler_Get(t *testing.T) {
	validID := uuid.New()

	tests := []struct {
		name           string
		clientID       string
		mockSetup      func(*MockClientService)
		expectedStatus int
	}{
		{
			name:     "successful get",
			clientID: validID.String(),
			mockSetup: func(m *MockClientService) {
				client := createTestClient()
				client.ID = validID
				m.On("GetClient", mock.Anything, validID).Return(client, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:     "invalid UUID",
			clientID: "not-a-uuid",
			mockSetup: func(m *MockClientService) {
				// No mock setup needed - should fail before service call
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:     "client not found",
			clientID: validID.String(),
			mockSetup: func(m *MockClientService) {
				m.On("GetClient", mock.Anything, validID).
					Return(nil, fmt.Errorf("client not found"))
			},
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			mockService := new(MockClientService)
			tt.mockSetup(mockService)
			handler := NewClientHandler(mockService)

			// Create request with mux vars
			req := httptest.NewRequest(http.MethodGet, "/api/v1/clients/"+tt.clientID, nil)
			req = mux.SetURLVars(req, map[string]string{"id": tt.clientID})
			rec := httptest.NewRecorder()

			// Execute
			handler.Get(rec, req)

			// Assert
			assert.Equal(t, tt.expectedStatus, rec.Code)
			mockService.AssertExpectations(t)
		})
	}
}

func TestClientHandler_GetBySlug(t *testing.T) {
	tests := []struct {
		name           string
		slug           string
		mockSetup      func(*MockClientService)
		expectedStatus int
	}{
		{
			name: "successful get by slug",
			slug: "test-client",
			mockSetup: func(m *MockClientService) {
				client := createTestClient()
				m.On("GetClientBySlug", mock.Anything, "test-client").Return(client, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "empty slug",
			slug: "",
			mockSetup: func(m *MockClientService) {
				// No mock setup needed
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "client not found",
			slug: "nonexistent",
			mockSetup: func(m *MockClientService) {
				m.On("GetClientBySlug", mock.Anything, "nonexistent").
					Return(nil, fmt.Errorf("client not found"))
			},
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			mockService := new(MockClientService)
			tt.mockSetup(mockService)
			handler := NewClientHandler(mockService)

			// Create request with mux vars
			req := httptest.NewRequest(http.MethodGet, "/api/v1/clients/slug/"+tt.slug, nil)
			req = mux.SetURLVars(req, map[string]string{"slug": tt.slug})
			rec := httptest.NewRecorder()

			// Execute
			handler.GetBySlug(rec, req)

			// Assert
			assert.Equal(t, tt.expectedStatus, rec.Code)
			mockService.AssertExpectations(t)
		})
	}
}

func TestClientHandler_Update(t *testing.T) {
	validID := uuid.New()
	newName := "Updated Name"

	tests := []struct {
		name           string
		clientID       string
		requestBody    interface{}
		mockSetup      func(*MockClientService)
		expectedStatus int
	}{
		{
			name:     "successful update",
			clientID: validID.String(),
			requestBody: services.UpdateClientRequest{
				Name: &newName,
			},
			mockSetup: func(m *MockClientService) {
				m.On("UpdateClient", mock.Anything, validID, mock.AnythingOfType("*services.UpdateClientRequest")).
					Return(nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:        "invalid UUID",
			clientID:    "not-a-uuid",
			requestBody: services.UpdateClientRequest{},
			mockSetup: func(m *MockClientService) {
				// No mock setup needed
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:        "invalid request body",
			clientID:    validID.String(),
			requestBody: "invalid json",
			mockSetup: func(m *MockClientService) {
				// No mock setup needed
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:     "client not found",
			clientID: validID.String(),
			requestBody: services.UpdateClientRequest{
				Name: &newName,
			},
			mockSetup: func(m *MockClientService) {
				m.On("UpdateClient", mock.Anything, validID, mock.AnythingOfType("*services.UpdateClientRequest")).
					Return(fmt.Errorf("client not found"))
			},
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			mockService := new(MockClientService)
			tt.mockSetup(mockService)
			handler := NewClientHandler(mockService)

			// Create request
			body, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest(http.MethodPut, "/api/v1/clients/"+tt.clientID, bytes.NewBuffer(body))
			req = mux.SetURLVars(req, map[string]string{"id": tt.clientID})
			rec := httptest.NewRecorder()

			// Execute
			handler.Update(rec, req)

			// Assert
			assert.Equal(t, tt.expectedStatus, rec.Code)
			mockService.AssertExpectations(t)
		})
	}
}

func TestClientHandler_Delete(t *testing.T) {
	validID := uuid.New()

	tests := []struct {
		name           string
		clientID       string
		mockSetup      func(*MockClientService)
		expectedStatus int
	}{
		{
			name:     "successful delete",
			clientID: validID.String(),
			mockSetup: func(m *MockClientService) {
				m.On("DeleteClient", mock.Anything, validID).Return(nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:     "invalid UUID",
			clientID: "not-a-uuid",
			mockSetup: func(m *MockClientService) {
				// No mock setup needed
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:     "client not found",
			clientID: validID.String(),
			mockSetup: func(m *MockClientService) {
				m.On("DeleteClient", mock.Anything, validID).
					Return(fmt.Errorf("client not found"))
			},
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			mockService := new(MockClientService)
			tt.mockSetup(mockService)
			handler := NewClientHandler(mockService)

			// Create request
			req := httptest.NewRequest(http.MethodDelete, "/api/v1/clients/"+tt.clientID, nil)
			req = mux.SetURLVars(req, map[string]string{"id": tt.clientID})
			rec := httptest.NewRecorder()

			// Execute
			handler.Delete(rec, req)

			// Assert
			assert.Equal(t, tt.expectedStatus, rec.Code)
			mockService.AssertExpectations(t)
		})
	}
}

func TestClientHandler_List(t *testing.T) {
	tests := []struct {
		name           string
		queryParams    map[string]string
		mockSetup      func(*MockClientService)
		expectedStatus int
		expectedPage   int
		expectedSize   int
	}{
		{
			name:        "successful list with defaults",
			queryParams: map[string]string{},
			mockSetup: func(m *MockClientService) {
				clients := []*models.Client{createTestClient(), createTestClient()}
				m.On("ListClients", mock.Anything, 1, 10).Return(clients, 2, nil)
			},
			expectedStatus: http.StatusOK,
			expectedPage:   1,
			expectedSize:   10,
		},
		{
			name: "successful list with custom pagination",
			queryParams: map[string]string{
				"page":      "2",
				"page_size": "20",
			},
			mockSetup: func(m *MockClientService) {
				clients := []*models.Client{createTestClient()}
				m.On("ListClients", mock.Anything, 2, 20).Return(clients, 1, nil)
			},
			expectedStatus: http.StatusOK,
			expectedPage:   2,
			expectedSize:   20,
		},
		{
			name: "invalid pagination parameters use defaults",
			queryParams: map[string]string{
				"page":      "invalid",
				"page_size": "invalid",
			},
			mockSetup: func(m *MockClientService) {
				clients := []*models.Client{}
				m.On("ListClients", mock.Anything, 1, 10).Return(clients, 0, nil)
			},
			expectedStatus: http.StatusOK,
			expectedPage:   1,
			expectedSize:   10,
		},
		{
			name:        "service error",
			queryParams: map[string]string{},
			mockSetup: func(m *MockClientService) {
				m.On("ListClients", mock.Anything, 1, 10).
					Return(nil, 0, fmt.Errorf("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			mockService := new(MockClientService)
			tt.mockSetup(mockService)
			handler := NewClientHandler(mockService)

			// Create request with query params
			url := "/api/v1/clients"
			if len(tt.queryParams) > 0 {
				url += "?"
				first := true
				for k, v := range tt.queryParams {
					if !first {
						url += "&"
					}
					url += k + "=" + v
					first = false
				}
			}
			req := httptest.NewRequest(http.MethodGet, url, nil)
			rec := httptest.NewRecorder()

			// Execute
			handler.List(rec, req)

			// Assert
			assert.Equal(t, tt.expectedStatus, rec.Code)

			if tt.expectedStatus == http.StatusOK {
				var response APIResponse
				json.NewDecoder(rec.Body).Decode(&response)
				assert.True(t, response.Success)
				assert.NotNil(t, response.Meta)
				assert.Equal(t, tt.expectedPage, response.Meta.Page)
				assert.Equal(t, tt.expectedSize, response.Meta.PageSize)
			}

			mockService.AssertExpectations(t)
		})
	}
}

func TestClientHandler_Activate(t *testing.T) {
	validID := uuid.New()

	tests := []struct {
		name           string
		clientID       string
		mockSetup      func(*MockClientService)
		expectedStatus int
	}{
		{
			name:     "successful activation",
			clientID: validID.String(),
			mockSetup: func(m *MockClientService) {
				m.On("ActivateClient", mock.Anything, validID).Return(nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:     "invalid UUID",
			clientID: "not-a-uuid",
			mockSetup: func(m *MockClientService) {
				// No mock setup needed
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:     "client not found",
			clientID: validID.String(),
			mockSetup: func(m *MockClientService) {
				m.On("ActivateClient", mock.Anything, validID).
					Return(fmt.Errorf("client not found"))
			},
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			mockService := new(MockClientService)
			tt.mockSetup(mockService)
			handler := NewClientHandler(mockService)

			// Create request
			req := httptest.NewRequest(http.MethodPost, "/api/v1/clients/"+tt.clientID+"/activate", nil)
			req = mux.SetURLVars(req, map[string]string{"id": tt.clientID})
			rec := httptest.NewRecorder()

			// Execute
			handler.Activate(rec, req)

			// Assert
			assert.Equal(t, tt.expectedStatus, rec.Code)
			mockService.AssertExpectations(t)
		})
	}
}

func TestClientHandler_Deactivate(t *testing.T) {
	validID := uuid.New()

	tests := []struct {
		name           string
		clientID       string
		mockSetup      func(*MockClientService)
		expectedStatus int
	}{
		{
			name:     "successful deactivation",
			clientID: validID.String(),
			mockSetup: func(m *MockClientService) {
				m.On("DeactivateClient", mock.Anything, validID).Return(nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:     "invalid UUID",
			clientID: "not-a-uuid",
			mockSetup: func(m *MockClientService) {
				// No mock setup needed
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:     "client not found",
			clientID: validID.String(),
			mockSetup: func(m *MockClientService) {
				m.On("DeactivateClient", mock.Anything, validID).
					Return(fmt.Errorf("client not found"))
			},
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			mockService := new(MockClientService)
			tt.mockSetup(mockService)
			handler := NewClientHandler(mockService)

			// Create request
			req := httptest.NewRequest(http.MethodPost, "/api/v1/clients/"+tt.clientID+"/deactivate", nil)
			req = mux.SetURLVars(req, map[string]string{"id": tt.clientID})
			rec := httptest.NewRecorder()

			// Execute
			handler.Deactivate(rec, req)

			// Assert
			assert.Equal(t, tt.expectedStatus, rec.Code)
			mockService.AssertExpectations(t)
		})
	}
}
