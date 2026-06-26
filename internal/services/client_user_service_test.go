package services

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/yourorg/shoppilot/internal/models"
	"golang.org/x/crypto/bcrypt"
)

// Test CreateUser
func TestClientUserService_CreateUser(t *testing.T) {
	ctx := context.Background()
	clientID := uuid.New()

	t.Run("successful creation", func(t *testing.T) {
		mockRepo := new(MockClientUserRepository)
		service := NewClientUserService(mockRepo)

		req := &CreateClientUserRequest{
			Email:        "test@example.com",
			Username:     "testuser",
			Password:     "password123",
			FirstName:    "Test",
			LastName:     "User",
			Phone:        "1234567890",
			UserStatusID: 1,
		}

		// Mock email doesn't exist
		mockRepo.On("GetByEmail", ctx, clientID, req.Email).Return(nil, errors.New("not found"))
		// Mock username doesn't exist
		mockRepo.On("GetByUsername", ctx, clientID, req.Username).Return(nil, errors.New("not found"))
		// Mock successful creation
		mockRepo.On("Create", ctx, mock.AnythingOfType("*models.ClientUser")).Return(nil)

		user, err := service.CreateUser(ctx, clientID, req)

		assert.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, clientID, user.ClientID)
		assert.Equal(t, req.Email, user.Email)
		assert.Equal(t, req.Username, user.Username)
		assert.NotEqual(t, req.Password, user.Password) // Password should be hashed

		// Verify password is hashed correctly
		err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password))
		assert.NoError(t, err)

		mockRepo.AssertExpectations(t)
	})

	t.Run("email already exists", func(t *testing.T) {
		mockRepo := new(MockClientUserRepository)
		service := NewClientUserService(mockRepo)

		req := &CreateClientUserRequest{
			Email:        "existing@example.com",
			Username:     "newuser",
			Password:     "password123",
			UserStatusID: 1,
		}

		existingUser := &models.ClientUser{
			ID:       uuid.New(),
			Email:    req.Email,
			ClientID: clientID,
		}

		mockRepo.On("GetByEmail", ctx, clientID, req.Email).Return(existingUser, nil)

		user, err := service.CreateUser(ctx, clientID, req)

		assert.Error(t, err)
		assert.Nil(t, user)
		assert.Contains(t, err.Error(), "email already exists")

		mockRepo.AssertExpectations(t)
	})

	t.Run("username already exists", func(t *testing.T) {
		mockRepo := new(MockClientUserRepository)
		service := NewClientUserService(mockRepo)

		req := &CreateClientUserRequest{
			Email:        "new@example.com",
			Username:     "existinguser",
			Password:     "password123",
			UserStatusID: 1,
		}

		existingUser := &models.ClientUser{
			ID:       uuid.New(),
			Username: req.Username,
			ClientID: clientID,
		}

		mockRepo.On("GetByEmail", ctx, clientID, req.Email).Return(nil, errors.New("not found"))
		mockRepo.On("GetByUsername", ctx, clientID, req.Username).Return(existingUser, nil)

		user, err := service.CreateUser(ctx, clientID, req)

		assert.Error(t, err)
		assert.Nil(t, user)
		assert.Contains(t, err.Error(), "username already exists")

		mockRepo.AssertExpectations(t)
	})

	t.Run("validation errors", func(t *testing.T) {
		mockRepo := new(MockClientUserRepository)
		service := NewClientUserService(mockRepo)

		tests := []struct {
			name   string
			req    *CreateClientUserRequest
			errMsg string
		}{
			{
				name: "empty email",
				req: &CreateClientUserRequest{
					Email:        "",
					Username:     "testuser",
					Password:     "password123",
					UserStatusID: 1,
				},
				errMsg: "email is required",
			},
			{
				name: "invalid email format",
				req: &CreateClientUserRequest{
					Email:        "notanemail",
					Username:     "testuser",
					Password:     "password123",
					UserStatusID: 1,
				},
				errMsg: "invalid email format",
			},
			{
				name: "empty username",
				req: &CreateClientUserRequest{
					Email:        "test@example.com",
					Username:     "",
					Password:     "password123",
					UserStatusID: 1,
				},
				errMsg: "username is required",
			},
			{
				name: "username too short",
				req: &CreateClientUserRequest{
					Email:        "test@example.com",
					Username:     "ab",
					Password:     "password123",
					UserStatusID: 1,
				},
				errMsg: "username must be between 3 and 50 characters",
			},
			{
				name: "password too short",
				req: &CreateClientUserRequest{
					Email:        "test@example.com",
					Username:     "testuser",
					Password:     "short",
					UserStatusID: 1,
				},
				errMsg: "password must be at least 8 characters",
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				user, err := service.CreateUser(ctx, clientID, tt.req)
				assert.Error(t, err)
				assert.Nil(t, user)
				assert.Contains(t, err.Error(), tt.errMsg)
			})
		}
	})
}

// Test Authenticate
func TestClientUserService_Authenticate(t *testing.T) {
	ctx := context.Background()
	clientID := uuid.New()
	password := "password123"
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	t.Run("successful authentication", func(t *testing.T) {
		mockRepo := new(MockClientUserRepository)
		service := NewClientUserService(mockRepo)

		user := &models.ClientUser{
			ID:       uuid.New(),
			ClientID: clientID,
			Username: "testuser",
			Password: string(hashedPassword),
		}

		mockRepo.On("GetByUsername", ctx, clientID, "testuser").Return(user, nil)

		authenticatedUser, err := service.Authenticate(ctx, clientID, "testuser", password)

		assert.NoError(t, err)
		assert.NotNil(t, authenticatedUser)
		assert.Equal(t, user.ID, authenticatedUser.ID)

		mockRepo.AssertExpectations(t)
	})

	t.Run("user not found", func(t *testing.T) {
		mockRepo := new(MockClientUserRepository)
		service := NewClientUserService(mockRepo)

		mockRepo.On("GetByUsername", ctx, clientID, "nonexistent").Return(nil, errors.New("not found"))

		user, err := service.Authenticate(ctx, clientID, "nonexistent", password)

		assert.Error(t, err)
		assert.Nil(t, user)
		assert.Contains(t, err.Error(), "invalid credentials")

		mockRepo.AssertExpectations(t)
	})

	t.Run("wrong password", func(t *testing.T) {
		mockRepo := new(MockClientUserRepository)
		service := NewClientUserService(mockRepo)

		user := &models.ClientUser{
			ID:       uuid.New(),
			ClientID: clientID,
			Username: "testuser",
			Password: string(hashedPassword),
		}

		mockRepo.On("GetByUsername", ctx, clientID, "testuser").Return(user, nil)

		authenticatedUser, err := service.Authenticate(ctx, clientID, "testuser", "wrongpassword")

		assert.Error(t, err)
		assert.Nil(t, authenticatedUser)
		assert.Contains(t, err.Error(), "invalid credentials")

		mockRepo.AssertExpectations(t)
	})
}

// Test ChangePassword
func TestClientUserService_ChangePassword(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	oldPassword := "oldpassword123"
	newPassword := "newpassword123"
	hashedOldPassword, _ := bcrypt.GenerateFromPassword([]byte(oldPassword), bcrypt.DefaultCost)

	t.Run("successful password change", func(t *testing.T) {
		mockRepo := new(MockClientUserRepository)
		service := NewClientUserService(mockRepo)

		user := &models.ClientUser{
			ID:       userID,
			Password: string(hashedOldPassword),
		}

		mockRepo.On("GetByID", ctx, userID).Return(user, nil)
		mockRepo.On("Update", ctx, mock.AnythingOfType("*models.ClientUser")).Return(nil)

		err := service.ChangePassword(ctx, userID, oldPassword, newPassword)

		assert.NoError(t, err)

		// Verify new password is hashed correctly
		err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(newPassword))
		assert.NoError(t, err)

		mockRepo.AssertExpectations(t)
	})

	t.Run("wrong old password", func(t *testing.T) {
		mockRepo := new(MockClientUserRepository)
		service := NewClientUserService(mockRepo)

		user := &models.ClientUser{
			ID:       userID,
			Password: string(hashedOldPassword),
		}

		mockRepo.On("GetByID", ctx, userID).Return(user, nil)

		err := service.ChangePassword(ctx, userID, "wrongoldpassword", newPassword)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid old password")

		mockRepo.AssertExpectations(t)
	})

	t.Run("new password too short", func(t *testing.T) {
		mockRepo := new(MockClientUserRepository)
		service := NewClientUserService(mockRepo)

		user := &models.ClientUser{
			ID:       userID,
			Password: string(hashedOldPassword),
		}

		mockRepo.On("GetByID", ctx, userID).Return(user, nil)

		err := service.ChangePassword(ctx, userID, oldPassword, "short")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "must be at least 8 characters")

		mockRepo.AssertExpectations(t)
	})
}

// Test DeleteUser - cannot delete last client_admin
func TestClientUserService_DeleteUser(t *testing.T) {
	ctx := context.Background()
	clientID := uuid.New()
	adminUserID := uuid.New()

	t.Run("prevent deleting last client_admin", func(t *testing.T) {
		mockRepo := new(MockClientUserRepository)
		service := NewClientUserService(mockRepo)

		adminUser := &models.ClientUser{
			ID:       adminUserID,
			ClientID: clientID,
		}

		adminRole := &models.ClientRole{
			ID:   1,
			Name: "client_admin",
		}

		// Mock GetByID
		mockRepo.On("GetByID", ctx, adminUserID).Return(adminUser, nil)

		// Mock GetUserRoles - user is admin
		mockRepo.On("GetUserRoles", ctx, adminUserID).Return([]*models.ClientRole{adminRole}, nil)

		// Mock ListByClient - only one user (this admin)
		mockRepo.On("ListByClient", ctx, clientID, 1000, 0).Return([]*models.ClientUser{adminUser}, nil)

		err := service.DeleteUser(ctx, adminUserID)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot delete the last client_admin user")

		mockRepo.AssertExpectations(t)
	})

	t.Run("allow deleting non-admin user", func(t *testing.T) {
		mockRepo := new(MockClientUserRepository)
		service := NewClientUserService(mockRepo)

		regularUserID := uuid.New()
		regularUser := &models.ClientUser{
			ID:       regularUserID,
			ClientID: clientID,
		}

		viewerRole := &models.ClientRole{
			ID:   4,
			Name: "viewer",
		}

		mockRepo.On("GetByID", ctx, regularUserID).Return(regularUser, nil)
		mockRepo.On("GetUserRoles", ctx, regularUserID).Return([]*models.ClientRole{viewerRole}, nil)
		mockRepo.On("Delete", ctx, regularUserID).Return(nil)

		err := service.DeleteUser(ctx, regularUserID)

		assert.NoError(t, err)

		mockRepo.AssertExpectations(t)
	})

	t.Run("allow deleting admin when multiple admins exist", func(t *testing.T) {
		mockRepo := new(MockClientUserRepository)
		service := NewClientUserService(mockRepo)

		admin1ID := uuid.New()
		admin2ID := uuid.New()

		admin1 := &models.ClientUser{
			ID:       admin1ID,
			ClientID: clientID,
		}

		admin2 := &models.ClientUser{
			ID:       admin2ID,
			ClientID: clientID,
		}

		adminRole := &models.ClientRole{
			ID:   1,
			Name: "client_admin",
		}

		mockRepo.On("GetByID", ctx, admin1ID).Return(admin1, nil)
		mockRepo.On("GetUserRoles", ctx, admin1ID).Return([]*models.ClientRole{adminRole}, nil)
		mockRepo.On("ListByClient", ctx, clientID, 1000, 0).Return([]*models.ClientUser{admin1, admin2}, nil)
		mockRepo.On("GetUserRoles", ctx, admin2ID).Return([]*models.ClientRole{adminRole}, nil)
		mockRepo.On("Delete", ctx, admin1ID).Return(nil)

		err := service.DeleteUser(ctx, admin1ID)

		assert.NoError(t, err)

		mockRepo.AssertExpectations(t)
	})
}

// Test AssignRole
func TestClientUserService_AssignRole(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()

	t.Run("successful role assignment", func(t *testing.T) {
		mockRepo := new(MockClientUserRepository)
		service := NewClientUserService(mockRepo)

		mockRepo.On("AssignRole", ctx, userID, 2).Return(nil) // shop_manager = 2

		err := service.AssignRole(ctx, userID, "shop_manager")

		assert.NoError(t, err)

		mockRepo.AssertExpectations(t)
	})

	t.Run("invalid role name", func(t *testing.T) {
		mockRepo := new(MockClientUserRepository)
		service := NewClientUserService(mockRepo)

		err := service.AssignRole(ctx, userID, "invalid_role")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid role name")

		mockRepo.AssertExpectations(t)
	})
}

// Test RemoveRole - cannot remove last client_admin
func TestClientUserService_RemoveRole(t *testing.T) {
	ctx := context.Background()
	clientID := uuid.New()
	adminUserID := uuid.New()

	t.Run("prevent removing last client_admin role", func(t *testing.T) {
		mockRepo := new(MockClientUserRepository)
		service := NewClientUserService(mockRepo)

		adminUser := &models.ClientUser{
			ID:       adminUserID,
			ClientID: clientID,
		}

		adminRole := &models.ClientRole{
			ID:   1,
			Name: "client_admin",
		}

		mockRepo.On("GetByID", ctx, adminUserID).Return(adminUser, nil)
		mockRepo.On("ListByClient", ctx, clientID, 1000, 0).Return([]*models.ClientUser{adminUser}, nil)
		mockRepo.On("GetUserRoles", ctx, adminUserID).Return([]*models.ClientRole{adminRole}, nil)

		err := service.RemoveRole(ctx, adminUserID, "client_admin")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot remove the last client_admin role")

		mockRepo.AssertExpectations(t)
	})

	t.Run("allow removing non-admin role", func(t *testing.T) {
		mockRepo := new(MockClientUserRepository)
		service := NewClientUserService(mockRepo)

		mockRepo.On("RemoveRole", ctx, adminUserID, 4).Return(nil) // viewer = 4

		err := service.RemoveRole(ctx, adminUserID, "viewer")

		assert.NoError(t, err)

		mockRepo.AssertExpectations(t)
	})
}

// Test HasPermission
func TestClientUserService_HasPermission(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()

	t.Run("user has permission", func(t *testing.T) {
		mockRepo := new(MockClientUserRepository)
		service := NewClientUserService(mockRepo)

		permissions := []*models.ClientPermission{
			{
				ID:       1,
				Resource: "shop",
				Action:   "create",
			},
			{
				ID:       2,
				Resource: "product",
				Action:   "read",
			},
		}

		mockRepo.On("GetUserPermissions", ctx, userID).Return(permissions, nil)

		has, err := service.HasPermission(ctx, userID, "shop", "create")

		assert.NoError(t, err)
		assert.True(t, has)

		mockRepo.AssertExpectations(t)
	})

	t.Run("user does not have permission", func(t *testing.T) {
		mockRepo := new(MockClientUserRepository)
		service := NewClientUserService(mockRepo)

		permissions := []*models.ClientPermission{
			{
				ID:       2,
				Resource: "product",
				Action:   "read",
			},
		}

		mockRepo.On("GetUserPermissions", ctx, userID).Return(permissions, nil)

		has, err := service.HasPermission(ctx, userID, "shop", "delete")

		assert.NoError(t, err)
		assert.False(t, has)

		mockRepo.AssertExpectations(t)
	})
}

// Test GetUserPermissions
func TestClientUserService_GetUserPermissions(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()

	t.Run("successful retrieval", func(t *testing.T) {
		mockRepo := new(MockClientUserRepository)
		service := NewClientUserService(mockRepo)

		expectedPermissions := []*models.ClientPermission{
			{
				ID:       1,
				Resource: "shop",
				Action:   "create",
			},
			{
				ID:       2,
				Resource: "product",
				Action:   "read",
			},
		}

		mockRepo.On("GetUserPermissions", ctx, userID).Return(expectedPermissions, nil)

		permissions, err := service.GetUserPermissions(ctx, userID)

		assert.NoError(t, err)
		assert.Equal(t, expectedPermissions, permissions)
		assert.Len(t, permissions, 2)

		mockRepo.AssertExpectations(t)
	})
}

// Test UpdateUser
func TestClientUserService_UpdateUser(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	clientID := uuid.New()

	t.Run("successful update", func(t *testing.T) {
		mockRepo := new(MockClientUserRepository)
		service := NewClientUserService(mockRepo)

		existingUser := &models.ClientUser{
			ID:        userID,
			ClientID:  clientID,
			Email:     "old@example.com",
			Username:  "oldusername",
			FirstName: "Old",
			LastName:  "Name",
		}

		newEmail := "new@example.com"
		newUsername := "newusername"
		req := &UpdateClientUserRequest{
			Email:    &newEmail,
			Username: &newUsername,
		}

		mockRepo.On("GetByID", ctx, userID).Return(existingUser, nil)
		mockRepo.On("GetByEmail", ctx, clientID, newEmail).Return(nil, errors.New("not found"))
		mockRepo.On("GetByUsername", ctx, clientID, newUsername).Return(nil, errors.New("not found"))
		mockRepo.On("Update", ctx, mock.AnythingOfType("*models.ClientUser")).Return(nil)

		err := service.UpdateUser(ctx, userID, req)

		assert.NoError(t, err)

		mockRepo.AssertExpectations(t)
	})

	t.Run("email already exists for another user", func(t *testing.T) {
		mockRepo := new(MockClientUserRepository)
		service := NewClientUserService(mockRepo)

		existingUser := &models.ClientUser{
			ID:       userID,
			ClientID: clientID,
			Email:    "old@example.com",
		}

		otherUserID := uuid.New()
		otherUser := &models.ClientUser{
			ID:       otherUserID,
			ClientID: clientID,
			Email:    "new@example.com",
		}

		newEmail := "new@example.com"
		req := &UpdateClientUserRequest{
			Email: &newEmail,
		}

		mockRepo.On("GetByID", ctx, userID).Return(existingUser, nil)
		mockRepo.On("GetByEmail", ctx, clientID, newEmail).Return(otherUser, nil)

		err := service.UpdateUser(ctx, userID, req)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "email already exists")

		mockRepo.AssertExpectations(t)
	})
}

// Test ListUsers
func TestClientUserService_ListUsers(t *testing.T) {
	ctx := context.Background()
	clientID := uuid.New()

	t.Run("successful listing with pagination", func(t *testing.T) {
		mockRepo := new(MockClientUserRepository)
		service := NewClientUserService(mockRepo)

		users := []*models.ClientUser{
			{ID: uuid.New(), ClientID: clientID, Email: "user1@example.com"},
			{ID: uuid.New(), ClientID: clientID, Email: "user2@example.com"},
		}

		mockRepo.On("ListByClient", ctx, clientID, 20, 0).Return(users, nil)

		result, total, err := service.ListUsers(ctx, clientID, 1, 20)

		assert.NoError(t, err)
		assert.Equal(t, users, result)
		assert.Equal(t, 2, total)

		mockRepo.AssertExpectations(t)
	})

	t.Run("default pagination values", func(t *testing.T) {
		mockRepo := new(MockClientUserRepository)
		service := NewClientUserService(mockRepo)

		users := []*models.ClientUser{}

		mockRepo.On("ListByClient", ctx, clientID, 20, 0).Return(users, nil)

		result, total, err := service.ListUsers(ctx, clientID, 0, 0)

		assert.NoError(t, err)
		assert.Equal(t, users, result)
		assert.Equal(t, 0, total)

		mockRepo.AssertExpectations(t)
	})
}
