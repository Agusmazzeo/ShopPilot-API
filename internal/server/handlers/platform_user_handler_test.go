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

// MockPlatformUserService is a mock implementation of PlatformUserService
type MockPlatformUserService struct {
	mock.Mock
}

func (m *MockPlatformUserService) CreateUser(ctx context.Context, req *services.CreatePlatformUserRequest) (*models.PlatformUser, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.PlatformUser), args.Error(1)
}

func (m *MockPlatformUserService) GetUser(ctx context.Context, id uuid.UUID) (*models.PlatformUser, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.PlatformUser), args.Error(1)
}

func (m *MockPlatformUserService) UpdateUser(ctx context.Context, id uuid.UUID, req *services.UpdatePlatformUserRequest) error {
	args := m.Called(ctx, id, req)
	return args.Error(0)
}

func (m *MockPlatformUserService) DeleteUser(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockPlatformUserService) ListUsers(ctx context.Context, page, pageSize int) ([]*models.PlatformUser, int, error) {
	args := m.Called(ctx, page, pageSize)
	if args.Get(0) == nil {
		return nil, args.Int(1), args.Error(2)
	}
	return args.Get(0).([]*models.PlatformUser), args.Int(1), args.Error(2)
}

func (m *MockPlatformUserService) Authenticate(ctx context.Context, username, password string) (*models.PlatformUser, error) {
	args := m.Called(ctx, username, password)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.PlatformUser), args.Error(1)
}

func (m *MockPlatformUserService) ChangePassword(ctx context.Context, userID uuid.UUID, oldPassword, newPassword string) error {
	args := m.Called(ctx, userID, oldPassword, newPassword)
	return args.Error(0)
}

func (m *MockPlatformUserService) AssignRole(ctx context.Context, userID uuid.UUID, roleName string) error {
	args := m.Called(ctx, userID, roleName)
	return args.Error(0)
}

func (m *MockPlatformUserService) RemoveRole(ctx context.Context, userID uuid.UUID, roleName string) error {
	args := m.Called(ctx, userID, roleName)
	return args.Error(0)
}

func (m *MockPlatformUserService) GetUserPermissions(ctx context.Context, userID uuid.UUID) ([]*models.PlatformPermission, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.PlatformPermission), args.Error(1)
}

func (m *MockPlatformUserService) HasPermission(ctx context.Context, userID uuid.UUID, resource, action string) (bool, error) {
	args := m.Called(ctx, userID, resource, action)
	return args.Bool(0), args.Error(1)
}

// Test helpers

func setupTestRouter(handler *PlatformUserHandler) *mux.Router {
	router := mux.NewRouter()
	router.HandleFunc("/api/v1/platform/users", handler.Create).Methods("POST")
	router.HandleFunc("/api/v1/platform/users/{id}", handler.Get).Methods("GET")
	router.HandleFunc("/api/v1/platform/users/{id}", handler.Update).Methods("PUT")
	router.HandleFunc("/api/v1/platform/users/{id}", handler.Delete).Methods("DELETE")
	router.HandleFunc("/api/v1/platform/users", handler.List).Methods("GET")
	router.HandleFunc("/api/v1/platform/users/{id}/roles", handler.AssignRole).Methods("POST")
	router.HandleFunc("/api/v1/platform/users/{id}/roles/{roleId}", handler.RemoveRole).Methods("DELETE")
	router.HandleFunc("/api/v1/platform/users/{id}/permissions", handler.GetPermissions).Methods("GET")
	return router
}

// Tests

func TestPlatformUserHandler_Create(t *testing.T) {
	t.Run("successful creation", func(t *testing.T) {
		mockService := new(MockPlatformUserService)
		handler := NewPlatformUserHandler(mockService)
		router := setupTestRouter(handler)

		userID := uuid.New()
		now := time.Now()

		expectedUser := &models.PlatformUser{
			ID:           userID,
			Email:        "test@example.com",
			Username:     "testuser",
			FirstName:    "Test",
			LastName:     "User",
			Phone:        "1234567890",
			UserStatusID: 1,
			CreatedAt:    now,
			UpdatedAt:    now,
		}

		mockService.On("CreateUser", mock.Anything, mock.MatchedBy(func(req *services.CreatePlatformUserRequest) bool {
			return req.Email == "test@example.com" &&
				req.Username == "testuser" &&
				req.Password == "password123"
		})).Return(expectedUser, nil)

		reqBody := CreatePlatformUserRequest{
			Email:     "test@example.com",
			Username:  "testuser",
			Password:  "password123",
			FirstName: "Test",
			LastName:  "User",
			Phone:     "1234567890",
		}

		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest("POST", "/api/v1/platform/users", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusCreated, rr.Code)

		var response APIResponse
		err := json.NewDecoder(rr.Body).Decode(&response)
		assert.NoError(t, err)
		assert.True(t, response.Success)

		mockService.AssertExpectations(t)
	})

	t.Run("invalid email", func(t *testing.T) {
		mockService := new(MockPlatformUserService)
		handler := NewPlatformUserHandler(mockService)
		router := setupTestRouter(handler)

		mockService.On("CreateUser", mock.Anything, mock.Anything).
			Return(nil, services.ErrInvalidEmail)

		reqBody := CreatePlatformUserRequest{
			Email:    "invalid-email",
			Username: "testuser",
			Password: "password123",
		}

		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest("POST", "/api/v1/platform/users", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)

		var response APIResponse
		err := json.NewDecoder(rr.Body).Decode(&response)
		assert.NoError(t, err)
		assert.False(t, response.Success)
		assert.Equal(t, "invalid_email", response.Error.Code)

		mockService.AssertExpectations(t)
	})

	t.Run("invalid username", func(t *testing.T) {
		mockService := new(MockPlatformUserService)
		handler := NewPlatformUserHandler(mockService)
		router := setupTestRouter(handler)

		mockService.On("CreateUser", mock.Anything, mock.Anything).
			Return(nil, services.ErrInvalidUsername)

		reqBody := CreatePlatformUserRequest{
			Email:    "test@example.com",
			Username: "ab",
			Password: "password123",
		}

		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest("POST", "/api/v1/platform/users", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)

		var response APIResponse
		err := json.NewDecoder(rr.Body).Decode(&response)
		assert.NoError(t, err)
		assert.False(t, response.Success)
		assert.Equal(t, "invalid_username", response.Error.Code)

		mockService.AssertExpectations(t)
	})

	t.Run("password too short", func(t *testing.T) {
		mockService := new(MockPlatformUserService)
		handler := NewPlatformUserHandler(mockService)
		router := setupTestRouter(handler)

		mockService.On("CreateUser", mock.Anything, mock.Anything).
			Return(nil, services.ErrPasswordTooShort)

		reqBody := CreatePlatformUserRequest{
			Email:    "test@example.com",
			Username: "testuser",
			Password: "pass",
		}

		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest("POST", "/api/v1/platform/users", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)

		var response APIResponse
		err := json.NewDecoder(rr.Body).Decode(&response)
		assert.NoError(t, err)
		assert.False(t, response.Success)
		assert.Equal(t, "password_too_short", response.Error.Code)

		mockService.AssertExpectations(t)
	})

	t.Run("invalid JSON", func(t *testing.T) {
		mockService := new(MockPlatformUserService)
		handler := NewPlatformUserHandler(mockService)
		router := setupTestRouter(handler)

		req := httptest.NewRequest("POST", "/api/v1/platform/users", bytes.NewBufferString("invalid json"))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)

		var response APIResponse
		err := json.NewDecoder(rr.Body).Decode(&response)
		assert.NoError(t, err)
		assert.False(t, response.Success)
		assert.Equal(t, "invalid_request", response.Error.Code)
	})
}

func TestPlatformUserHandler_Get(t *testing.T) {
	t.Run("successful get", func(t *testing.T) {
		mockService := new(MockPlatformUserService)
		handler := NewPlatformUserHandler(mockService)
		router := setupTestRouter(handler)

		userID := uuid.New()
		now := time.Now()

		expectedUser := &models.PlatformUser{
			ID:           userID,
			Email:        "test@example.com",
			Username:     "testuser",
			FirstName:    "Test",
			LastName:     "User",
			Phone:        "1234567890",
			UserStatusID: 1,
			CreatedAt:    now,
			UpdatedAt:    now,
		}

		mockService.On("GetUser", mock.Anything, userID).Return(expectedUser, nil)

		req := httptest.NewRequest("GET", "/api/v1/platform/users/"+userID.String(), nil)
		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		var response APIResponse
		err := json.NewDecoder(rr.Body).Decode(&response)
		assert.NoError(t, err)
		assert.True(t, response.Success)

		mockService.AssertExpectations(t)
	})

	t.Run("user not found", func(t *testing.T) {
		mockService := new(MockPlatformUserService)
		handler := NewPlatformUserHandler(mockService)
		router := setupTestRouter(handler)

		userID := uuid.New()

		mockService.On("GetUser", mock.Anything, userID).
			Return(nil, errors.New("user not found"))

		req := httptest.NewRequest("GET", "/api/v1/platform/users/"+userID.String(), nil)
		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusNotFound, rr.Code)

		var response APIResponse
		err := json.NewDecoder(rr.Body).Decode(&response)
		assert.NoError(t, err)
		assert.False(t, response.Success)
		assert.Equal(t, "user_not_found", response.Error.Code)

		mockService.AssertExpectations(t)
	})

	t.Run("invalid UUID format", func(t *testing.T) {
		mockService := new(MockPlatformUserService)
		handler := NewPlatformUserHandler(mockService)
		router := setupTestRouter(handler)

		req := httptest.NewRequest("GET", "/api/v1/platform/users/invalid-uuid", nil)
		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)

		var response APIResponse
		err := json.NewDecoder(rr.Body).Decode(&response)
		assert.NoError(t, err)
		assert.False(t, response.Success)
		assert.Equal(t, "invalid_id", response.Error.Code)
	})
}

func TestPlatformUserHandler_Update(t *testing.T) {
	t.Run("successful update", func(t *testing.T) {
		mockService := new(MockPlatformUserService)
		handler := NewPlatformUserHandler(mockService)
		router := setupTestRouter(handler)

		userID := uuid.New()

		mockService.On("UpdateUser", mock.Anything, userID, mock.Anything).Return(nil)

		email := "newemail@example.com"
		reqBody := UpdatePlatformUserRequest{
			Email: &email,
		}

		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest("PUT", "/api/v1/platform/users/"+userID.String(), bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		var response APIResponse
		err := json.NewDecoder(rr.Body).Decode(&response)
		assert.NoError(t, err)
		assert.True(t, response.Success)

		mockService.AssertExpectations(t)
	})

	t.Run("invalid email on update", func(t *testing.T) {
		mockService := new(MockPlatformUserService)
		handler := NewPlatformUserHandler(mockService)
		router := setupTestRouter(handler)

		userID := uuid.New()

		mockService.On("UpdateUser", mock.Anything, userID, mock.Anything).
			Return(services.ErrInvalidEmail)

		email := "invalid-email"
		reqBody := UpdatePlatformUserRequest{
			Email: &email,
		}

		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest("PUT", "/api/v1/platform/users/"+userID.String(), bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)

		var response APIResponse
		err := json.NewDecoder(rr.Body).Decode(&response)
		assert.NoError(t, err)
		assert.False(t, response.Success)
		assert.Equal(t, "invalid_email", response.Error.Code)

		mockService.AssertExpectations(t)
	})
}

func TestPlatformUserHandler_Delete(t *testing.T) {
	t.Run("successful delete", func(t *testing.T) {
		mockService := new(MockPlatformUserService)
		handler := NewPlatformUserHandler(mockService)
		router := setupTestRouter(handler)

		userID := uuid.New()

		mockService.On("DeleteUser", mock.Anything, userID).Return(nil)

		req := httptest.NewRequest("DELETE", "/api/v1/platform/users/"+userID.String(), nil)
		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		var response APIResponse
		err := json.NewDecoder(rr.Body).Decode(&response)
		assert.NoError(t, err)
		assert.True(t, response.Success)

		mockService.AssertExpectations(t)
	})

	t.Run("cannot delete super admin", func(t *testing.T) {
		mockService := new(MockPlatformUserService)
		handler := NewPlatformUserHandler(mockService)
		router := setupTestRouter(handler)

		userID := uuid.New()

		mockService.On("DeleteUser", mock.Anything, userID).
			Return(services.ErrCannotDeleteSuperAdmin)

		req := httptest.NewRequest("DELETE", "/api/v1/platform/users/"+userID.String(), nil)
		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)

		var response APIResponse
		err := json.NewDecoder(rr.Body).Decode(&response)
		assert.NoError(t, err)
		assert.False(t, response.Success)
		assert.Equal(t, "cannot_delete_super_admin", response.Error.Code)

		mockService.AssertExpectations(t)
	})
}

func TestPlatformUserHandler_List(t *testing.T) {
	t.Run("successful list with defaults", func(t *testing.T) {
		mockService := new(MockPlatformUserService)
		handler := NewPlatformUserHandler(mockService)
		router := setupTestRouter(handler)

		now := time.Now()
		users := []*models.PlatformUser{
			{
				ID:           uuid.New(),
				Email:        "user1@example.com",
				Username:     "user1",
				FirstName:    "User",
				LastName:     "One",
				UserStatusID: 1,
				CreatedAt:    now,
				UpdatedAt:    now,
			},
			{
				ID:           uuid.New(),
				Email:        "user2@example.com",
				Username:     "user2",
				FirstName:    "User",
				LastName:     "Two",
				UserStatusID: 1,
				CreatedAt:    now,
				UpdatedAt:    now,
			},
		}

		mockService.On("ListUsers", mock.Anything, 1, 10).Return(users, 2, nil)

		req := httptest.NewRequest("GET", "/api/v1/platform/users", nil)
		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		var response APIResponse
		err := json.NewDecoder(rr.Body).Decode(&response)
		assert.NoError(t, err)
		assert.True(t, response.Success)

		mockService.AssertExpectations(t)
	})

	t.Run("successful list with pagination", func(t *testing.T) {
		mockService := new(MockPlatformUserService)
		handler := NewPlatformUserHandler(mockService)
		router := setupTestRouter(handler)

		users := []*models.PlatformUser{}

		mockService.On("ListUsers", mock.Anything, 2, 20).Return(users, 0, nil)

		req := httptest.NewRequest("GET", "/api/v1/platform/users?page=2&page_size=20", nil)
		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		var response APIResponse
		err := json.NewDecoder(rr.Body).Decode(&response)
		assert.NoError(t, err)
		assert.True(t, response.Success)

		mockService.AssertExpectations(t)
	})
}

func TestPlatformUserHandler_AssignRole(t *testing.T) {
	t.Run("successful role assignment", func(t *testing.T) {
		mockService := new(MockPlatformUserService)
		handler := NewPlatformUserHandler(mockService)
		router := setupTestRouter(handler)

		userID := uuid.New()

		mockService.On("AssignRole", mock.Anything, userID, "platform_admin").Return(nil)

		reqBody := AssignRoleRequest{
			RoleName: "platform_admin",
		}

		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest("POST", "/api/v1/platform/users/"+userID.String()+"/roles", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		var response APIResponse
		err := json.NewDecoder(rr.Body).Decode(&response)
		assert.NoError(t, err)
		assert.True(t, response.Success)

		mockService.AssertExpectations(t)
	})

	t.Run("role not found", func(t *testing.T) {
		mockService := new(MockPlatformUserService)
		handler := NewPlatformUserHandler(mockService)
		router := setupTestRouter(handler)

		userID := uuid.New()

		mockService.On("AssignRole", mock.Anything, userID, "nonexistent_role").
			Return(services.ErrRoleNotFound)

		reqBody := AssignRoleRequest{
			RoleName: "nonexistent_role",
		}

		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest("POST", "/api/v1/platform/users/"+userID.String()+"/roles", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)

		var response APIResponse
		err := json.NewDecoder(rr.Body).Decode(&response)
		assert.NoError(t, err)
		assert.False(t, response.Success)
		assert.Equal(t, "role_not_found", response.Error.Code)

		mockService.AssertExpectations(t)
	})

	t.Run("missing role name", func(t *testing.T) {
		mockService := new(MockPlatformUserService)
		handler := NewPlatformUserHandler(mockService)
		router := setupTestRouter(handler)

		userID := uuid.New()

		reqBody := AssignRoleRequest{
			RoleName: "",
		}

		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest("POST", "/api/v1/platform/users/"+userID.String()+"/roles", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)

		var response APIResponse
		err := json.NewDecoder(rr.Body).Decode(&response)
		assert.NoError(t, err)
		assert.False(t, response.Success)
		assert.Equal(t, "missing_role_name", response.Error.Code)
	})
}

func TestPlatformUserHandler_RemoveRole(t *testing.T) {
	t.Run("successful role removal", func(t *testing.T) {
		mockService := new(MockPlatformUserService)
		handler := NewPlatformUserHandler(mockService)
		router := setupTestRouter(handler)

		userID := uuid.New()

		mockService.On("RemoveRole", mock.Anything, userID, "platform_admin").Return(nil)

		req := httptest.NewRequest("DELETE", "/api/v1/platform/users/"+userID.String()+"/roles/2", nil)
		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		var response APIResponse
		err := json.NewDecoder(rr.Body).Decode(&response)
		assert.NoError(t, err)
		assert.True(t, response.Success)

		mockService.AssertExpectations(t)
	})

	t.Run("cannot remove last super admin", func(t *testing.T) {
		mockService := new(MockPlatformUserService)
		handler := NewPlatformUserHandler(mockService)
		router := setupTestRouter(handler)

		userID := uuid.New()

		mockService.On("RemoveRole", mock.Anything, userID, "super_admin").
			Return(services.ErrCannotRemoveLastSuperAdmin)

		req := httptest.NewRequest("DELETE", "/api/v1/platform/users/"+userID.String()+"/roles/1", nil)
		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)

		var response APIResponse
		err := json.NewDecoder(rr.Body).Decode(&response)
		assert.NoError(t, err)
		assert.False(t, response.Success)
		assert.Equal(t, "cannot_remove_last_super_admin", response.Error.Code)

		mockService.AssertExpectations(t)
	})

	t.Run("invalid role ID", func(t *testing.T) {
		mockService := new(MockPlatformUserService)
		handler := NewPlatformUserHandler(mockService)
		router := setupTestRouter(handler)

		userID := uuid.New()

		req := httptest.NewRequest("DELETE", "/api/v1/platform/users/"+userID.String()+"/roles/999", nil)
		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)

		var response APIResponse
		err := json.NewDecoder(rr.Body).Decode(&response)
		assert.NoError(t, err)
		assert.False(t, response.Success)
		assert.Equal(t, "invalid_role_id", response.Error.Code)
	})
}

func TestPlatformUserHandler_GetPermissions(t *testing.T) {
	t.Run("successful get permissions", func(t *testing.T) {
		mockService := new(MockPlatformUserService)
		handler := NewPlatformUserHandler(mockService)
		router := setupTestRouter(handler)

		userID := uuid.New()
		now := time.Now()

		permissions := []*models.PlatformPermission{
			{
				ID:          1,
				Name:        "read_users",
				Description: "Read user data",
				Resource:    "users",
				Action:      "read",
				CreatedAt:   now,
			},
			{
				ID:          2,
				Name:        "write_users",
				Description: "Write user data",
				Resource:    "users",
				Action:      "write",
				CreatedAt:   now,
			},
		}

		mockService.On("GetUserPermissions", mock.Anything, userID).Return(permissions, nil)

		req := httptest.NewRequest("GET", "/api/v1/platform/users/"+userID.String()+"/permissions", nil)
		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		var response APIResponse
		err := json.NewDecoder(rr.Body).Decode(&response)
		assert.NoError(t, err)
		assert.True(t, response.Success)

		mockService.AssertExpectations(t)
	})

	t.Run("error getting permissions", func(t *testing.T) {
		mockService := new(MockPlatformUserService)
		handler := NewPlatformUserHandler(mockService)
		router := setupTestRouter(handler)

		userID := uuid.New()

		mockService.On("GetUserPermissions", mock.Anything, userID).
			Return(nil, errors.New("database error"))

		req := httptest.NewRequest("GET", "/api/v1/platform/users/"+userID.String()+"/permissions", nil)
		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusInternalServerError, rr.Code)

		var response APIResponse
		err := json.NewDecoder(rr.Body).Decode(&response)
		assert.NoError(t, err)
		assert.False(t, response.Success)
		assert.Equal(t, "internal_error", response.Error.Code)

		mockService.AssertExpectations(t)
	})
}
