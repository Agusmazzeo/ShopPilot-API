package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
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

// MockClientUserService is a mock implementation of ClientUserService
type MockClientUserService struct {
	mock.Mock
}

func (m *MockClientUserService) CreateUser(ctx context.Context, clientID uuid.UUID, req *services.CreateClientUserRequest) (*models.ClientUser, error) {
	args := m.Called(ctx, clientID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.ClientUser), args.Error(1)
}

func (m *MockClientUserService) GetUser(ctx context.Context, userID uuid.UUID) (*models.ClientUser, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.ClientUser), args.Error(1)
}

func (m *MockClientUserService) UpdateUser(ctx context.Context, userID uuid.UUID, req *services.UpdateClientUserRequest) error {
	args := m.Called(ctx, userID, req)
	return args.Error(0)
}

func (m *MockClientUserService) DeleteUser(ctx context.Context, userID uuid.UUID) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func (m *MockClientUserService) ListUsers(ctx context.Context, clientID uuid.UUID, page, pageSize int) ([]*models.ClientUser, int, error) {
	args := m.Called(ctx, clientID, page, pageSize)
	if args.Get(0) == nil {
		return nil, args.Int(1), args.Error(2)
	}
	return args.Get(0).([]*models.ClientUser), args.Int(1), args.Error(2)
}

func (m *MockClientUserService) Authenticate(ctx context.Context, clientID uuid.UUID, username, password string) (*models.ClientUser, error) {
	args := m.Called(ctx, clientID, username, password)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.ClientUser), args.Error(1)
}

func (m *MockClientUserService) ChangePassword(ctx context.Context, userID uuid.UUID, oldPassword, newPassword string) error {
	args := m.Called(ctx, userID, oldPassword, newPassword)
	return args.Error(0)
}

func (m *MockClientUserService) AssignRole(ctx context.Context, userID uuid.UUID, roleName string) error {
	args := m.Called(ctx, userID, roleName)
	return args.Error(0)
}

func (m *MockClientUserService) RemoveRole(ctx context.Context, userID uuid.UUID, roleName string) error {
	args := m.Called(ctx, userID, roleName)
	return args.Error(0)
}

func (m *MockClientUserService) GetUserPermissions(ctx context.Context, userID uuid.UUID) ([]*models.ClientPermission, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.ClientPermission), args.Error(1)
}

func (m *MockClientUserService) HasPermission(ctx context.Context, userID uuid.UUID, resource, action string) (bool, error) {
	args := m.Called(ctx, userID, resource, action)
	return args.Bool(0), args.Error(1)
}

// TestClientUserHandler_Create tests the Create handler
func TestClientUserHandler_Create(t *testing.T) {
	clientID := uuid.New()
	userID := uuid.New()
	now := time.Now()

	tests := []struct {
		name           string
		clientID       string
		requestBody    interface{}
		mockSetup      func(*MockClientUserService)
		expectedStatus int
		expectedError  string
	}{
		{
			name:     "successful creation",
			clientID: clientID.String(),
			requestBody: CreateClientUserRequest{
				Email:        "test@example.com",
				Username:     "testuser",
				Password:     "password123",
				FirstName:    "Test",
				LastName:     "User",
				Phone:        "1234567890",
				UserStatusID: 1,
			},
			mockSetup: func(m *MockClientUserService) {
				m.On("CreateUser", mock.Anything, clientID, mock.AnythingOfType("*services.CreateClientUserRequest")).
					Return(&models.ClientUser{
						ID:           userID,
						ClientID:     clientID,
						Email:        "test@example.com",
						Username:     "testuser",
						FirstName:    "Test",
						LastName:     "User",
						Phone:        "1234567890",
						UserStatusID: 1,
						CreatedAt:    now,
						UpdatedAt:    now,
					}, nil)
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name:           "invalid client ID",
			clientID:       "invalid-uuid",
			requestBody:    CreateClientUserRequest{},
			mockSetup:      func(m *MockClientUserService) {},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "INVALID_CLIENT_ID",
		},
		{
			name:     "invalid request body",
			clientID: clientID.String(),
			requestBody: map[string]interface{}{
				"email": 12345, // Invalid type
			},
			mockSetup:      func(m *MockClientUserService) {},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "INVALID_REQUEST",
		},
		{
			name:     "service error - duplicate email",
			clientID: clientID.String(),
			requestBody: CreateClientUserRequest{
				Email:        "test@example.com",
				Username:     "testuser",
				Password:     "password123",
				UserStatusID: 1,
			},
			mockSetup: func(m *MockClientUserService) {
				m.On("CreateUser", mock.Anything, clientID, mock.AnythingOfType("*services.CreateClientUserRequest")).
					Return(nil, errors.New("email already exists for this client"))
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "CREATE_FAILED",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockClientUserService)
			tt.mockSetup(mockService)

			handler := NewClientUserHandler(mockService)

			body, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest("POST", "/api/v1/clients/"+tt.clientID+"/users", bytes.NewBuffer(body))
			req = mux.SetURLVars(req, map[string]string{"clientId": tt.clientID})

			rr := httptest.NewRecorder()
			handler.Create(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)

			var response APIResponse
			json.Unmarshal(rr.Body.Bytes(), &response)

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

// TestClientUserHandler_Get tests the Get handler
func TestClientUserHandler_Get(t *testing.T) {
	clientID := uuid.New()
	userID := uuid.New()
	now := time.Now()

	tests := []struct {
		name           string
		userID         string
		mockSetup      func(*MockClientUserService)
		expectedStatus int
		expectedError  string
	}{
		{
			name:   "successful get",
			userID: userID.String(),
			mockSetup: func(m *MockClientUserService) {
				m.On("GetUser", mock.Anything, userID).
					Return(&models.ClientUser{
						ID:           userID,
						ClientID:     clientID,
						Email:        "test@example.com",
						Username:     "testuser",
						FirstName:    "Test",
						LastName:     "User",
						Phone:        "1234567890",
						UserStatusID: 1,
						CreatedAt:    now,
						UpdatedAt:    now,
					}, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "invalid user ID",
			userID:         "invalid-uuid",
			mockSetup:      func(m *MockClientUserService) {},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "INVALID_USER_ID",
		},
		{
			name:   "user not found",
			userID: userID.String(),
			mockSetup: func(m *MockClientUserService) {
				m.On("GetUser", mock.Anything, userID).
					Return(nil, errors.New("user not found"))
			},
			expectedStatus: http.StatusNotFound,
			expectedError:  "USER_NOT_FOUND",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockClientUserService)
			tt.mockSetup(mockService)

			handler := NewClientUserHandler(mockService)

			req := httptest.NewRequest("GET", "/api/v1/clients/"+clientID.String()+"/users/"+tt.userID, nil)
			req = mux.SetURLVars(req, map[string]string{"clientId": clientID.String(), "id": tt.userID})

			rr := httptest.NewRecorder()
			handler.Get(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)

			var response APIResponse
			json.Unmarshal(rr.Body.Bytes(), &response)

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

// TestClientUserHandler_Update tests the Update handler
func TestClientUserHandler_Update(t *testing.T) {
	userID := uuid.New()
	email := "updated@example.com"

	tests := []struct {
		name           string
		userID         string
		requestBody    interface{}
		mockSetup      func(*MockClientUserService)
		expectedStatus int
		expectedError  string
	}{
		{
			name:   "successful update",
			userID: userID.String(),
			requestBody: UpdateClientUserRequest{
				Email: &email,
			},
			mockSetup: func(m *MockClientUserService) {
				m.On("UpdateUser", mock.Anything, userID, mock.AnythingOfType("*services.UpdateClientUserRequest")).
					Return(nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "invalid user ID",
			userID:         "invalid-uuid",
			requestBody:    UpdateClientUserRequest{},
			mockSetup:      func(m *MockClientUserService) {},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "INVALID_USER_ID",
		},
		{
			name:   "service error",
			userID: userID.String(),
			requestBody: UpdateClientUserRequest{
				Email: &email,
			},
			mockSetup: func(m *MockClientUserService) {
				m.On("UpdateUser", mock.Anything, userID, mock.AnythingOfType("*services.UpdateClientUserRequest")).
					Return(errors.New("email already exists for this client"))
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "UPDATE_FAILED",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockClientUserService)
			tt.mockSetup(mockService)

			handler := NewClientUserHandler(mockService)

			body, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest("PUT", "/api/v1/clients/"+uuid.New().String()+"/users/"+tt.userID, bytes.NewBuffer(body))
			req = mux.SetURLVars(req, map[string]string{"id": tt.userID})

			rr := httptest.NewRecorder()
			handler.Update(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)

			var response APIResponse
			json.Unmarshal(rr.Body.Bytes(), &response)

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

// TestClientUserHandler_Delete tests the Delete handler
func TestClientUserHandler_Delete(t *testing.T) {
	userID := uuid.New()

	tests := []struct {
		name           string
		userID         string
		mockSetup      func(*MockClientUserService)
		expectedStatus int
		expectedError  string
	}{
		{
			name:   "successful delete",
			userID: userID.String(),
			mockSetup: func(m *MockClientUserService) {
				m.On("DeleteUser", mock.Anything, userID).
					Return(nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "invalid user ID",
			userID:         "invalid-uuid",
			mockSetup:      func(m *MockClientUserService) {},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "INVALID_USER_ID",
		},
		{
			name:   "cannot delete last admin",
			userID: userID.String(),
			mockSetup: func(m *MockClientUserService) {
				m.On("DeleteUser", mock.Anything, userID).
					Return(errors.New("cannot delete the last client_admin user"))
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "DELETE_FAILED",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockClientUserService)
			tt.mockSetup(mockService)

			handler := NewClientUserHandler(mockService)

			req := httptest.NewRequest("DELETE", "/api/v1/clients/"+uuid.New().String()+"/users/"+tt.userID, nil)
			req = mux.SetURLVars(req, map[string]string{"id": tt.userID})

			rr := httptest.NewRecorder()
			handler.Delete(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)

			var response APIResponse
			json.Unmarshal(rr.Body.Bytes(), &response)

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

// TestClientUserHandler_List tests the List handler
func TestClientUserHandler_List(t *testing.T) {
	clientID := uuid.New()
	userID1 := uuid.New()
	userID2 := uuid.New()
	now := time.Now()

	tests := []struct {
		name           string
		clientID       string
		queryParams    string
		mockSetup      func(*MockClientUserService)
		expectedStatus int
		expectedError  string
	}{
		{
			name:     "successful list with default pagination",
			clientID: clientID.String(),
			mockSetup: func(m *MockClientUserService) {
				m.On("ListUsers", mock.Anything, clientID, 1, 20).
					Return([]*models.ClientUser{
						{
							ID:           userID1,
							ClientID:     clientID,
							Email:        "user1@example.com",
							Username:     "user1",
							FirstName:    "User",
							LastName:     "One",
							UserStatusID: 1,
							CreatedAt:    now,
							UpdatedAt:    now,
						},
						{
							ID:           userID2,
							ClientID:     clientID,
							Email:        "user2@example.com",
							Username:     "user2",
							FirstName:    "User",
							LastName:     "Two",
							UserStatusID: 1,
							CreatedAt:    now,
							UpdatedAt:    now,
						},
					}, 2, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:        "successful list with custom pagination",
			clientID:    clientID.String(),
			queryParams: "?page=2&page_size=10",
			mockSetup: func(m *MockClientUserService) {
				m.On("ListUsers", mock.Anything, clientID, 2, 10).
					Return([]*models.ClientUser{}, 0, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "invalid client ID",
			clientID:       "invalid-uuid",
			mockSetup:      func(m *MockClientUserService) {},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "INVALID_CLIENT_ID",
		},
		{
			name:     "service error",
			clientID: clientID.String(),
			mockSetup: func(m *MockClientUserService) {
				m.On("ListUsers", mock.Anything, clientID, 1, 20).
					Return(nil, 0, errors.New("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedError:  "LIST_FAILED",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockClientUserService)
			tt.mockSetup(mockService)

			handler := NewClientUserHandler(mockService)

			req := httptest.NewRequest("GET", "/api/v1/clients/"+tt.clientID+"/users"+tt.queryParams, nil)
			req = mux.SetURLVars(req, map[string]string{"clientId": tt.clientID})

			rr := httptest.NewRecorder()
			handler.List(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)

			var response APIResponse
			json.Unmarshal(rr.Body.Bytes(), &response)

			if tt.expectedError != "" {
				assert.False(t, response.Success)
				assert.NotNil(t, response.Error)
				assert.Equal(t, tt.expectedError, response.Error.Code)
			} else {
				assert.True(t, response.Success)
				assert.Nil(t, response.Error)
				assert.NotNil(t, response.Meta)
			}

			mockService.AssertExpectations(t)
		})
	}
}

// TestClientUserHandler_AssignRole tests the AssignRole handler
func TestClientUserHandler_AssignRole(t *testing.T) {
	userID := uuid.New()

	tests := []struct {
		name           string
		userID         string
		requestBody    interface{}
		mockSetup      func(*MockClientUserService)
		expectedStatus int
		expectedError  string
	}{
		{
			name:   "successful role assignment",
			userID: userID.String(),
			requestBody: AssignRoleRequest{
				RoleName: "client_admin",
			},
			mockSetup: func(m *MockClientUserService) {
				m.On("AssignRole", mock.Anything, userID, "client_admin").
					Return(nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "invalid user ID",
			userID:         "invalid-uuid",
			requestBody:    AssignRoleRequest{RoleName: "client_admin"},
			mockSetup:      func(m *MockClientUserService) {},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "INVALID_USER_ID",
		},
		{
			name:   "invalid role name",
			userID: userID.String(),
			requestBody: AssignRoleRequest{
				RoleName: "invalid_role",
			},
			mockSetup: func(m *MockClientUserService) {
				m.On("AssignRole", mock.Anything, userID, "invalid_role").
					Return(errors.New("invalid role name: invalid_role"))
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "ASSIGN_ROLE_FAILED",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockClientUserService)
			tt.mockSetup(mockService)

			handler := NewClientUserHandler(mockService)

			body, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest("POST", "/api/v1/clients/"+uuid.New().String()+"/users/"+tt.userID+"/roles", bytes.NewBuffer(body))
			req = mux.SetURLVars(req, map[string]string{"id": tt.userID})

			rr := httptest.NewRecorder()
			handler.AssignRole(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)

			var response APIResponse
			json.Unmarshal(rr.Body.Bytes(), &response)

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

// TestClientUserHandler_RemoveRole tests the RemoveRole handler
func TestClientUserHandler_RemoveRole(t *testing.T) {
	userID := uuid.New()

	tests := []struct {
		name           string
		userID         string
		roleID         string
		mockSetup      func(*MockClientUserService)
		expectedStatus int
		expectedError  string
	}{
		{
			name:   "successful role removal",
			userID: userID.String(),
			roleID: "shop_manager",
			mockSetup: func(m *MockClientUserService) {
				m.On("RemoveRole", mock.Anything, userID, "shop_manager").
					Return(nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "invalid user ID",
			userID:         "invalid-uuid",
			roleID:         "shop_manager",
			mockSetup:      func(m *MockClientUserService) {},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "INVALID_USER_ID",
		},
		{
			name:   "cannot remove last admin",
			userID: userID.String(),
			roleID: "client_admin",
			mockSetup: func(m *MockClientUserService) {
				m.On("RemoveRole", mock.Anything, userID, "client_admin").
					Return(errors.New("cannot remove the last client_admin role"))
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "REMOVE_ROLE_FAILED",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockClientUserService)
			tt.mockSetup(mockService)

			handler := NewClientUserHandler(mockService)

			req := httptest.NewRequest("DELETE", "/api/v1/clients/"+uuid.New().String()+"/users/"+tt.userID+"/roles/"+tt.roleID, nil)
			req = mux.SetURLVars(req, map[string]string{"id": tt.userID, "roleId": tt.roleID})

			rr := httptest.NewRecorder()
			handler.RemoveRole(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)

			var response APIResponse
			json.Unmarshal(rr.Body.Bytes(), &response)

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
