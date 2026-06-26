package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/yourorg/shoppilot/internal/models"
	"github.com/yourorg/shoppilot/internal/services"
)

// MockShopService is a mock implementation of ShopService
type MockShopService struct {
	mock.Mock
}

func (m *MockShopService) CreateShop(ctx context.Context, clientID uuid.UUID, req *services.CreateShopRequest) (*models.Shop, error) {
	args := m.Called(ctx, clientID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Shop), args.Error(1)
}

func (m *MockShopService) GetShop(ctx context.Context, clientID uuid.UUID, shopID uuid.UUID) (*models.Shop, error) {
	args := m.Called(ctx, clientID, shopID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Shop), args.Error(1)
}

func (m *MockShopService) GetShopBySlug(ctx context.Context, clientID uuid.UUID, slug string) (*models.Shop, error) {
	args := m.Called(ctx, clientID, slug)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Shop), args.Error(1)
}

func (m *MockShopService) UpdateShop(ctx context.Context, clientID uuid.UUID, shopID uuid.UUID, req *services.UpdateShopRequest) error {
	args := m.Called(ctx, clientID, shopID, req)
	return args.Error(0)
}

func (m *MockShopService) DeleteShop(ctx context.Context, clientID uuid.UUID, shopID uuid.UUID) error {
	args := m.Called(ctx, clientID, shopID)
	return args.Error(0)
}

func (m *MockShopService) ListShops(ctx context.Context, clientID uuid.UUID, page, pageSize int) ([]*models.Shop, int, error) {
	args := m.Called(ctx, clientID, page, pageSize)
	if args.Get(0) == nil {
		return nil, args.Int(1), args.Error(2)
	}
	return args.Get(0).([]*models.Shop), args.Int(1), args.Error(2)
}

func (m *MockShopService) AssignUserToShop(ctx context.Context, shopID uuid.UUID, clientUserID uuid.UUID, roleName string) error {
	args := m.Called(ctx, shopID, clientUserID, roleName)
	return args.Error(0)
}

func (m *MockShopService) RemoveUserFromShop(ctx context.Context, shopID uuid.UUID, clientUserRoleID int) error {
	args := m.Called(ctx, shopID, clientUserRoleID)
	return args.Error(0)
}

func (m *MockShopService) GetShopUsers(ctx context.Context, shopID uuid.UUID) ([]*models.ShopUser, error) {
	args := m.Called(ctx, shopID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.ShopUser), args.Error(1)
}

func TestShopHandler_Create(t *testing.T) {
	clientID := uuid.New()
	shopID := uuid.New()

	tests := []struct {
		name           string
		clientIDParam  string
		requestBody    interface{}
		mockSetup      func(*MockShopService)
		expectedStatus int
		expectedError  string
	}{
		{
			name:          "success",
			clientIDParam: clientID.String(),
			requestBody: services.CreateShopRequest{
				Name:        "Test Shop",
				Description: "A test shop",
				Email:       "test@shop.com",
			},
			mockSetup: func(m *MockShopService) {
				m.On("CreateShop", mock.Anything, clientID, mock.AnythingOfType("*services.CreateShopRequest")).
					Return(&models.Shop{
						ID:       shopID,
						ClientID: clientID,
						Name:     "Test Shop",
						Slug:     "test-shop",
					}, nil)
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name:           "invalid client ID",
			clientIDParam:  "invalid-uuid",
			requestBody:    services.CreateShopRequest{Name: "Test Shop"},
			mockSetup:      func(m *MockShopService) {},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "INVALID_CLIENT_ID",
		},
		{
			name:           "invalid request body",
			clientIDParam:  clientID.String(),
			requestBody:    "invalid json",
			mockSetup:      func(m *MockShopService) {},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "INVALID_REQUEST",
		},
		{
			name:          "service error",
			clientIDParam: clientID.String(),
			requestBody: services.CreateShopRequest{
				Name: "Test Shop",
			},
			mockSetup: func(m *MockShopService) {
				m.On("CreateShop", mock.Anything, clientID, mock.AnythingOfType("*services.CreateShopRequest")).
					Return(nil, fmt.Errorf("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedError:  "CREATE_FAILED",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			mockService := new(MockShopService)
			tt.mockSetup(mockService)
			handler := NewShopHandler(mockService)

			// Create request
			body, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest(http.MethodPost, "/api/v1/clients/"+tt.clientIDParam+"/shops", bytes.NewBuffer(body))
			req = mux.SetURLVars(req, map[string]string{"clientId": tt.clientIDParam})
			w := httptest.NewRecorder()

			// Execute
			handler.Create(w, req)

			// Assert
			assert.Equal(t, tt.expectedStatus, w.Code)

			var response APIResponse
			json.NewDecoder(w.Body).Decode(&response)

			if tt.expectedError != "" {
				assert.False(t, response.Success)
				assert.NotNil(t, response.Error)
				assert.Equal(t, tt.expectedError, response.Error.Code)
			} else {
				assert.True(t, response.Success)
				assert.Nil(t, response.Error)
			}

			mockService.AssertExpectations(t)
		})
	}
}

func TestShopHandler_Get(t *testing.T) {
	clientID := uuid.New()
	shopID := uuid.New()

	tests := []struct {
		name           string
		clientIDParam  string
		shopIDParam    string
		mockSetup      func(*MockShopService)
		expectedStatus int
		expectedError  string
	}{
		{
			name:          "success",
			clientIDParam: clientID.String(),
			shopIDParam:   shopID.String(),
			mockSetup: func(m *MockShopService) {
				m.On("GetShop", mock.Anything, clientID, shopID).
					Return(&models.Shop{
						ID:       shopID,
						ClientID: clientID,
						Name:     "Test Shop",
					}, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "invalid client ID",
			clientIDParam:  "invalid-uuid",
			shopIDParam:    shopID.String(),
			mockSetup:      func(m *MockShopService) {},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "INVALID_CLIENT_ID",
		},
		{
			name:           "invalid shop ID",
			clientIDParam:  clientID.String(),
			shopIDParam:    "invalid-uuid",
			mockSetup:      func(m *MockShopService) {},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "INVALID_SHOP_ID",
		},
		{
			name:          "shop not found",
			clientIDParam: clientID.String(),
			shopIDParam:   shopID.String(),
			mockSetup: func(m *MockShopService) {
				m.On("GetShop", mock.Anything, clientID, shopID).
					Return(nil, fmt.Errorf("shop not found"))
			},
			expectedStatus: http.StatusNotFound,
			expectedError:  "SHOP_NOT_FOUND",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			mockService := new(MockShopService)
			tt.mockSetup(mockService)
			handler := NewShopHandler(mockService)

			// Create request
			req := httptest.NewRequest(http.MethodGet, "/api/v1/clients/"+tt.clientIDParam+"/shops/"+tt.shopIDParam, nil)
			req = mux.SetURLVars(req, map[string]string{
				"clientId": tt.clientIDParam,
				"id":       tt.shopIDParam,
			})
			w := httptest.NewRecorder()

			// Execute
			handler.Get(w, req)

			// Assert
			assert.Equal(t, tt.expectedStatus, w.Code)

			var response APIResponse
			json.NewDecoder(w.Body).Decode(&response)

			if tt.expectedError != "" {
				assert.False(t, response.Success)
				assert.NotNil(t, response.Error)
				assert.Equal(t, tt.expectedError, response.Error.Code)
			} else {
				assert.True(t, response.Success)
			}

			mockService.AssertExpectations(t)
		})
	}
}

func TestShopHandler_GetBySlug(t *testing.T) {
	clientID := uuid.New()
	shopID := uuid.New()

	tests := []struct {
		name           string
		clientIDParam  string
		slugParam      string
		mockSetup      func(*MockShopService)
		expectedStatus int
		expectedError  string
	}{
		{
			name:          "success",
			clientIDParam: clientID.String(),
			slugParam:     "test-shop",
			mockSetup: func(m *MockShopService) {
				m.On("GetShopBySlug", mock.Anything, clientID, "test-shop").
					Return(&models.Shop{
						ID:       shopID,
						ClientID: clientID,
						Name:     "Test Shop",
						Slug:     "test-shop",
					}, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "invalid client ID",
			clientIDParam:  "invalid-uuid",
			slugParam:      "test-shop",
			mockSetup:      func(m *MockShopService) {},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "INVALID_CLIENT_ID",
		},
		{
			name:           "empty slug",
			clientIDParam:  clientID.String(),
			slugParam:      "",
			mockSetup:      func(m *MockShopService) {},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "INVALID_SLUG",
		},
		{
			name:          "shop not found",
			clientIDParam: clientID.String(),
			slugParam:     "non-existent",
			mockSetup: func(m *MockShopService) {
				m.On("GetShopBySlug", mock.Anything, clientID, "non-existent").
					Return(nil, fmt.Errorf("shop not found"))
			},
			expectedStatus: http.StatusNotFound,
			expectedError:  "SHOP_NOT_FOUND",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			mockService := new(MockShopService)
			tt.mockSetup(mockService)
			handler := NewShopHandler(mockService)

			// Create request
			req := httptest.NewRequest(http.MethodGet, "/api/v1/clients/"+tt.clientIDParam+"/shops/slug/"+tt.slugParam, nil)
			req = mux.SetURLVars(req, map[string]string{
				"clientId": tt.clientIDParam,
				"slug":     tt.slugParam,
			})
			w := httptest.NewRecorder()

			// Execute
			handler.GetBySlug(w, req)

			// Assert
			assert.Equal(t, tt.expectedStatus, w.Code)

			var response APIResponse
			json.NewDecoder(w.Body).Decode(&response)

			if tt.expectedError != "" {
				assert.False(t, response.Success)
				assert.NotNil(t, response.Error)
				assert.Equal(t, tt.expectedError, response.Error.Code)
			} else {
				assert.True(t, response.Success)
			}

			mockService.AssertExpectations(t)
		})
	}
}

func TestShopHandler_Update(t *testing.T) {
	clientID := uuid.New()
	shopID := uuid.New()
	newName := "Updated Shop"

	tests := []struct {
		name           string
		clientIDParam  string
		shopIDParam    string
		requestBody    interface{}
		mockSetup      func(*MockShopService)
		expectedStatus int
		expectedError  string
	}{
		{
			name:          "success",
			clientIDParam: clientID.String(),
			shopIDParam:   shopID.String(),
			requestBody: services.UpdateShopRequest{
				Name: &newName,
			},
			mockSetup: func(m *MockShopService) {
				m.On("UpdateShop", mock.Anything, clientID, shopID, mock.AnythingOfType("*services.UpdateShopRequest")).
					Return(nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "invalid client ID",
			clientIDParam:  "invalid-uuid",
			shopIDParam:    shopID.String(),
			requestBody:    services.UpdateShopRequest{Name: &newName},
			mockSetup:      func(m *MockShopService) {},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "INVALID_CLIENT_ID",
		},
		{
			name:           "invalid shop ID",
			clientIDParam:  clientID.String(),
			shopIDParam:    "invalid-uuid",
			requestBody:    services.UpdateShopRequest{Name: &newName},
			mockSetup:      func(m *MockShopService) {},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "INVALID_SHOP_ID",
		},
		{
			name:           "invalid request body",
			clientIDParam:  clientID.String(),
			shopIDParam:    shopID.String(),
			requestBody:    "invalid json",
			mockSetup:      func(m *MockShopService) {},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "INVALID_REQUEST",
		},
		{
			name:          "service error",
			clientIDParam: clientID.String(),
			shopIDParam:   shopID.String(),
			requestBody: services.UpdateShopRequest{
				Name: &newName,
			},
			mockSetup: func(m *MockShopService) {
				m.On("UpdateShop", mock.Anything, clientID, shopID, mock.AnythingOfType("*services.UpdateShopRequest")).
					Return(fmt.Errorf("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedError:  "UPDATE_FAILED",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			mockService := new(MockShopService)
			tt.mockSetup(mockService)
			handler := NewShopHandler(mockService)

			// Create request
			body, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest(http.MethodPut, "/api/v1/clients/"+tt.clientIDParam+"/shops/"+tt.shopIDParam, bytes.NewBuffer(body))
			req = mux.SetURLVars(req, map[string]string{
				"clientId": tt.clientIDParam,
				"id":       tt.shopIDParam,
			})
			w := httptest.NewRecorder()

			// Execute
			handler.Update(w, req)

			// Assert
			assert.Equal(t, tt.expectedStatus, w.Code)

			var response APIResponse
			json.NewDecoder(w.Body).Decode(&response)

			if tt.expectedError != "" {
				assert.False(t, response.Success)
				assert.NotNil(t, response.Error)
				assert.Equal(t, tt.expectedError, response.Error.Code)
			} else {
				assert.True(t, response.Success)
			}

			mockService.AssertExpectations(t)
		})
	}
}

func TestShopHandler_Delete(t *testing.T) {
	clientID := uuid.New()
	shopID := uuid.New()

	tests := []struct {
		name           string
		clientIDParam  string
		shopIDParam    string
		mockSetup      func(*MockShopService)
		expectedStatus int
		expectedError  string
	}{
		{
			name:          "success",
			clientIDParam: clientID.String(),
			shopIDParam:   shopID.String(),
			mockSetup: func(m *MockShopService) {
				m.On("DeleteShop", mock.Anything, clientID, shopID).
					Return(nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "invalid client ID",
			clientIDParam:  "invalid-uuid",
			shopIDParam:    shopID.String(),
			mockSetup:      func(m *MockShopService) {},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "INVALID_CLIENT_ID",
		},
		{
			name:           "invalid shop ID",
			clientIDParam:  clientID.String(),
			shopIDParam:    "invalid-uuid",
			mockSetup:      func(m *MockShopService) {},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "INVALID_SHOP_ID",
		},
		{
			name:          "service error",
			clientIDParam: clientID.String(),
			shopIDParam:   shopID.String(),
			mockSetup: func(m *MockShopService) {
				m.On("DeleteShop", mock.Anything, clientID, shopID).
					Return(fmt.Errorf("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedError:  "DELETE_FAILED",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			mockService := new(MockShopService)
			tt.mockSetup(mockService)
			handler := NewShopHandler(mockService)

			// Create request
			req := httptest.NewRequest(http.MethodDelete, "/api/v1/clients/"+tt.clientIDParam+"/shops/"+tt.shopIDParam, nil)
			req = mux.SetURLVars(req, map[string]string{
				"clientId": tt.clientIDParam,
				"id":       tt.shopIDParam,
			})
			w := httptest.NewRecorder()

			// Execute
			handler.Delete(w, req)

			// Assert
			assert.Equal(t, tt.expectedStatus, w.Code)

			var response APIResponse
			json.NewDecoder(w.Body).Decode(&response)

			if tt.expectedError != "" {
				assert.False(t, response.Success)
				assert.NotNil(t, response.Error)
				assert.Equal(t, tt.expectedError, response.Error.Code)
			} else {
				assert.True(t, response.Success)
			}

			mockService.AssertExpectations(t)
		})
	}
}

func TestShopHandler_List(t *testing.T) {
	clientID := uuid.New()
	shopID1 := uuid.New()
	shopID2 := uuid.New()

	tests := []struct {
		name           string
		clientIDParam  string
		queryParams    map[string]string
		mockSetup      func(*MockShopService)
		expectedStatus int
		expectedError  string
		expectedPage   int
		expectedSize   int
	}{
		{
			name:          "success with default pagination",
			clientIDParam: clientID.String(),
			queryParams:   map[string]string{},
			mockSetup: func(m *MockShopService) {
				m.On("ListShops", mock.Anything, clientID, 1, 20).
					Return([]*models.Shop{
						{ID: shopID1, ClientID: clientID, Name: "Shop 1"},
						{ID: shopID2, ClientID: clientID, Name: "Shop 2"},
					}, 2, nil)
			},
			expectedStatus: http.StatusOK,
			expectedPage:   1,
			expectedSize:   20,
		},
		{
			name:          "success with custom pagination",
			clientIDParam: clientID.String(),
			queryParams:   map[string]string{"page": "2", "pageSize": "10"},
			mockSetup: func(m *MockShopService) {
				m.On("ListShops", mock.Anything, clientID, 2, 10).
					Return([]*models.Shop{
						{ID: shopID1, ClientID: clientID, Name: "Shop 1"},
					}, 11, nil)
			},
			expectedStatus: http.StatusOK,
			expectedPage:   2,
			expectedSize:   10,
		},
		{
			name:           "invalid client ID",
			clientIDParam:  "invalid-uuid",
			queryParams:    map[string]string{},
			mockSetup:      func(m *MockShopService) {},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "INVALID_CLIENT_ID",
		},
		{
			name:          "service error",
			clientIDParam: clientID.String(),
			queryParams:   map[string]string{},
			mockSetup: func(m *MockShopService) {
				m.On("ListShops", mock.Anything, clientID, 1, 20).
					Return(nil, 0, fmt.Errorf("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedError:  "LIST_FAILED",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			mockService := new(MockShopService)
			tt.mockSetup(mockService)
			handler := NewShopHandler(mockService)

			// Create request
			url := "/api/v1/clients/" + tt.clientIDParam + "/shops"
			if len(tt.queryParams) > 0 {
				url += "?"
				for k, v := range tt.queryParams {
					url += k + "=" + v + "&"
				}
			}
			req := httptest.NewRequest(http.MethodGet, url, nil)
			req = mux.SetURLVars(req, map[string]string{"clientId": tt.clientIDParam})
			w := httptest.NewRecorder()

			// Execute
			handler.List(w, req)

			// Assert
			assert.Equal(t, tt.expectedStatus, w.Code)

			var response APIResponse
			json.NewDecoder(w.Body).Decode(&response)

			if tt.expectedError != "" {
				assert.False(t, response.Success)
				assert.NotNil(t, response.Error)
				assert.Equal(t, tt.expectedError, response.Error.Code)
			} else {
				assert.True(t, response.Success)
				assert.NotNil(t, response.Meta)
				assert.Equal(t, tt.expectedPage, response.Meta.Page)
				assert.Equal(t, tt.expectedSize, response.Meta.PageSize)
			}

			mockService.AssertExpectations(t)
		})
	}
}

func TestShopHandler_AssignUser(t *testing.T) {
	shopID := uuid.New()
	userID := uuid.New()

	tests := []struct {
		name           string
		shopIDParam    string
		requestBody    interface{}
		mockSetup      func(*MockShopService)
		expectedStatus int
		expectedError  string
	}{
		{
			name:        "success",
			shopIDParam: shopID.String(),
			requestBody: AssignUserRequest{
				ClientUserID: userID,
				RoleName:     "shop_manager",
			},
			mockSetup: func(m *MockShopService) {
				m.On("AssignUserToShop", mock.Anything, shopID, userID, "shop_manager").
					Return(nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:        "invalid shop ID",
			shopIDParam: "invalid-uuid",
			requestBody: AssignUserRequest{
				ClientUserID: userID,
				RoleName:     "shop_manager",
			},
			mockSetup:      func(m *MockShopService) {},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "INVALID_SHOP_ID",
		},
		{
			name:           "invalid request body",
			shopIDParam:    shopID.String(),
			requestBody:    "invalid json",
			mockSetup:      func(m *MockShopService) {},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "INVALID_REQUEST",
		},
		{
			name:        "service error",
			shopIDParam: shopID.String(),
			requestBody: AssignUserRequest{
				ClientUserID: userID,
				RoleName:     "shop_manager",
			},
			mockSetup: func(m *MockShopService) {
				m.On("AssignUserToShop", mock.Anything, shopID, userID, "shop_manager").
					Return(fmt.Errorf("user assignment failed"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedError:  "ASSIGN_FAILED",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			mockService := new(MockShopService)
			tt.mockSetup(mockService)
			handler := NewShopHandler(mockService)

			// Create request
			body, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest(http.MethodPost, "/api/v1/shops/"+tt.shopIDParam+"/users", bytes.NewBuffer(body))
			req = mux.SetURLVars(req, map[string]string{"id": tt.shopIDParam})
			w := httptest.NewRecorder()

			// Execute
			handler.AssignUser(w, req)

			// Assert
			assert.Equal(t, tt.expectedStatus, w.Code)

			var response APIResponse
			json.NewDecoder(w.Body).Decode(&response)

			if tt.expectedError != "" {
				assert.False(t, response.Success)
				assert.NotNil(t, response.Error)
				assert.Equal(t, tt.expectedError, response.Error.Code)
			} else {
				assert.True(t, response.Success)
			}

			mockService.AssertExpectations(t)
		})
	}
}

func TestShopHandler_RemoveUser(t *testing.T) {
	shopID := uuid.New()

	tests := []struct {
		name           string
		shopIDParam    string
		userRoleIDStr  string
		mockSetup      func(*MockShopService)
		expectedStatus int
		expectedError  string
	}{
		{
			name:          "success",
			shopIDParam:   shopID.String(),
			userRoleIDStr: "123",
			mockSetup: func(m *MockShopService) {
				m.On("RemoveUserFromShop", mock.Anything, shopID, 123).
					Return(nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "invalid shop ID",
			shopIDParam:    "invalid-uuid",
			userRoleIDStr:  "123",
			mockSetup:      func(m *MockShopService) {},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "INVALID_SHOP_ID",
		},
		{
			name:           "invalid user role ID",
			shopIDParam:    shopID.String(),
			userRoleIDStr:  "invalid",
			mockSetup:      func(m *MockShopService) {},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "INVALID_USER_ROLE_ID",
		},
		{
			name:          "service error",
			shopIDParam:   shopID.String(),
			userRoleIDStr: "123",
			mockSetup: func(m *MockShopService) {
				m.On("RemoveUserFromShop", mock.Anything, shopID, 123).
					Return(fmt.Errorf("removal failed"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedError:  "REMOVE_FAILED",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			mockService := new(MockShopService)
			tt.mockSetup(mockService)
			handler := NewShopHandler(mockService)

			// Create request
			req := httptest.NewRequest(http.MethodDelete, "/api/v1/shops/"+tt.shopIDParam+"/users/"+tt.userRoleIDStr, nil)
			req = mux.SetURLVars(req, map[string]string{
				"id":         tt.shopIDParam,
				"userRoleId": tt.userRoleIDStr,
			})
			w := httptest.NewRecorder()

			// Execute
			handler.RemoveUser(w, req)

			// Assert
			assert.Equal(t, tt.expectedStatus, w.Code)

			var response APIResponse
			json.NewDecoder(w.Body).Decode(&response)

			if tt.expectedError != "" {
				assert.False(t, response.Success)
				assert.NotNil(t, response.Error)
				assert.Equal(t, tt.expectedError, response.Error.Code)
			} else {
				assert.True(t, response.Success)
			}

			mockService.AssertExpectations(t)
		})
	}
}

func TestShopHandler_GetUsers(t *testing.T) {
	shopID := uuid.New()
	clientID := uuid.New()

	tests := []struct {
		name           string
		shopIDParam    string
		mockSetup      func(*MockShopService)
		expectedStatus int
		expectedError  string
	}{
		{
			name:        "success",
			shopIDParam: shopID.String(),
			mockSetup: func(m *MockShopService) {
				m.On("GetShopUsers", mock.Anything, shopID).
					Return([]*models.ShopUser{
						{ID: 1, ClientID: clientID, ShopID: shopID, ClientUserRoleID: 123},
						{ID: 2, ClientID: clientID, ShopID: shopID, ClientUserRoleID: 124},
					}, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "invalid shop ID",
			shopIDParam:    "invalid-uuid",
			mockSetup:      func(m *MockShopService) {},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "INVALID_SHOP_ID",
		},
		{
			name:        "service error",
			shopIDParam: shopID.String(),
			mockSetup: func(m *MockShopService) {
				m.On("GetShopUsers", mock.Anything, shopID).
					Return(nil, fmt.Errorf("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedError:  "GET_USERS_FAILED",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			mockService := new(MockShopService)
			tt.mockSetup(mockService)
			handler := NewShopHandler(mockService)

			// Create request
			req := httptest.NewRequest(http.MethodGet, "/api/v1/shops/"+tt.shopIDParam+"/users", nil)
			req = mux.SetURLVars(req, map[string]string{"id": tt.shopIDParam})
			w := httptest.NewRecorder()

			// Execute
			handler.GetUsers(w, req)

			// Assert
			assert.Equal(t, tt.expectedStatus, w.Code)

			var response APIResponse
			json.NewDecoder(w.Body).Decode(&response)

			if tt.expectedError != "" {
				assert.False(t, response.Success)
				assert.NotNil(t, response.Error)
				assert.Equal(t, tt.expectedError, response.Error.Code)
			} else {
				assert.True(t, response.Success)
				assert.NotNil(t, response.Data)
			}

			mockService.AssertExpectations(t)
		})
	}
}
