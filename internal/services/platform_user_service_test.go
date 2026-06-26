package services

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"

	"github.com/yourorg/shoppilot/internal/models"
)

// MockPlatformUserRepository is a mock implementation of PlatformUserRepository
type MockPlatformUserRepository struct {
	mock.Mock
}

func (m *MockPlatformUserRepository) Create(ctx context.Context, user *models.PlatformUser) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockPlatformUserRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.PlatformUser, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.PlatformUser), args.Error(1)
}

func (m *MockPlatformUserRepository) GetByEmail(ctx context.Context, email string) (*models.PlatformUser, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.PlatformUser), args.Error(1)
}

func (m *MockPlatformUserRepository) GetByUsername(ctx context.Context, username string) (*models.PlatformUser, error) {
	args := m.Called(ctx, username)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.PlatformUser), args.Error(1)
}

func (m *MockPlatformUserRepository) Update(ctx context.Context, user *models.PlatformUser) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockPlatformUserRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockPlatformUserRepository) List(ctx context.Context, limit, offset int) ([]*models.PlatformUser, error) {
	args := m.Called(ctx, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.PlatformUser), args.Error(1)
}

func (m *MockPlatformUserRepository) AssignRole(ctx context.Context, userID uuid.UUID, roleID int) error {
	args := m.Called(ctx, userID, roleID)
	return args.Error(0)
}

func (m *MockPlatformUserRepository) RemoveRole(ctx context.Context, userID uuid.UUID, roleID int) error {
	args := m.Called(ctx, userID, roleID)
	return args.Error(0)
}

func (m *MockPlatformUserRepository) GetUserRoles(ctx context.Context, userID uuid.UUID) ([]*models.PlatformRole, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.PlatformRole), args.Error(1)
}

func (m *MockPlatformUserRepository) GetUserPermissions(ctx context.Context, userID uuid.UUID) ([]*models.PlatformPermission, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.PlatformPermission), args.Error(1)
}

// TestPlatformUserService_CreateUser tests user creation with password hashing and validation
func TestPlatformUserService_CreateUser(t *testing.T) {
	tests := []struct {
		name        string
		request     *CreatePlatformUserRequest
		setupMock   func(*MockPlatformUserRepository)
		wantErr     bool
		expectedErr error
	}{
		{
			name: "successful user creation",
			request: &CreatePlatformUserRequest{
				Email:     "test@example.com",
				Username:  "testuser",
				Password:  "password123",
				FirstName: "Test",
				LastName:  "User",
				Phone:     "1234567890",
			},
			setupMock: func(m *MockPlatformUserRepository) {
				m.On("Create", mock.Anything, mock.MatchedBy(func(u *models.PlatformUser) bool {
					// Verify password is hashed
					err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte("password123"))
					return err == nil &&
						u.Email == "test@example.com" &&
						u.Username == "testuser" &&
						u.UserStatusID == 1
				})).Return(nil)
			},
			wantErr: false,
		},
		{
			name: "invalid email format",
			request: &CreatePlatformUserRequest{
				Email:    "invalid-email",
				Username: "testuser",
				Password: "password123",
			},
			setupMock:   func(m *MockPlatformUserRepository) {},
			wantErr:     true,
			expectedErr: ErrInvalidEmail,
		},
		{
			name: "username too short",
			request: &CreatePlatformUserRequest{
				Email:    "test@example.com",
				Username: "ab",
				Password: "password123",
			},
			setupMock:   func(m *MockPlatformUserRepository) {},
			wantErr:     true,
			expectedErr: ErrInvalidUsername,
		},
		{
			name: "username too long",
			request: &CreatePlatformUserRequest{
				Email:    "test@example.com",
				Username: "abcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyz",
				Password: "password123",
			},
			setupMock:   func(m *MockPlatformUserRepository) {},
			wantErr:     true,
			expectedErr: ErrInvalidUsername,
		},
		{
			name: "username with special characters",
			request: &CreatePlatformUserRequest{
				Email:    "test@example.com",
				Username: "test@user!",
				Password: "password123",
			},
			setupMock:   func(m *MockPlatformUserRepository) {},
			wantErr:     true,
			expectedErr: ErrInvalidUsername,
		},
		{
			name: "password too short",
			request: &CreatePlatformUserRequest{
				Email:    "test@example.com",
				Username: "testuser",
				Password: "short",
			},
			setupMock:   func(m *MockPlatformUserRepository) {},
			wantErr:     true,
			expectedErr: ErrPasswordTooShort,
		},
		{
			name: "repository error",
			request: &CreatePlatformUserRequest{
				Email:    "test@example.com",
				Username: "testuser",
				Password: "password123",
			},
			setupMock: func(m *MockPlatformUserRepository) {
				m.On("Create", mock.Anything, mock.Anything).Return(errors.New("database error"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockPlatformUserRepository)
			tt.setupMock(mockRepo)

			service := NewPlatformUserService(mockRepo)
			user, err := service.CreateUser(context.Background(), tt.request)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.expectedErr != nil {
					assert.ErrorIs(t, err, tt.expectedErr)
				}
				assert.Nil(t, user)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, user)
				assert.Equal(t, tt.request.Email, user.Email)
				assert.Equal(t, tt.request.Username, user.Username)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

// TestPlatformUserService_Authenticate tests user authentication
func TestPlatformUserService_Authenticate(t *testing.T) {
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("correctpassword"), bcrypt.DefaultCost)

	tests := []struct {
		name        string
		username    string
		password    string
		setupMock   func(*MockPlatformUserRepository)
		wantErr     bool
		expectedErr error
	}{
		{
			name:     "successful authentication",
			username: "testuser",
			password: "correctpassword",
			setupMock: func(m *MockPlatformUserRepository) {
				m.On("GetByUsername", mock.Anything, "testuser").Return(&models.PlatformUser{
					ID:       uuid.New(),
					Username: "testuser",
					Password: string(hashedPassword),
				}, nil)
			},
			wantErr: false,
		},
		{
			name:     "user not found",
			username: "nonexistent",
			password: "password",
			setupMock: func(m *MockPlatformUserRepository) {
				m.On("GetByUsername", mock.Anything, "nonexistent").Return(nil, errors.New("user not found"))
			},
			wantErr:     true,
			expectedErr: ErrInvalidCredentials,
		},
		{
			name:     "incorrect password",
			username: "testuser",
			password: "wrongpassword",
			setupMock: func(m *MockPlatformUserRepository) {
				m.On("GetByUsername", mock.Anything, "testuser").Return(&models.PlatformUser{
					ID:       uuid.New(),
					Username: "testuser",
					Password: string(hashedPassword),
				}, nil)
			},
			wantErr:     true,
			expectedErr: ErrInvalidCredentials,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockPlatformUserRepository)
			tt.setupMock(mockRepo)

			service := NewPlatformUserService(mockRepo)
			user, err := service.Authenticate(context.Background(), tt.username, tt.password)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.expectedErr != nil {
					assert.ErrorIs(t, err, tt.expectedErr)
				}
				assert.Nil(t, user)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, user)
				assert.Equal(t, tt.username, user.Username)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

// TestPlatformUserService_ChangePassword tests password change functionality
func TestPlatformUserService_ChangePassword(t *testing.T) {
	userID := uuid.New()
	oldHashedPassword, _ := bcrypt.GenerateFromPassword([]byte("oldpassword"), bcrypt.DefaultCost)

	tests := []struct {
		name        string
		userID      uuid.UUID
		oldPassword string
		newPassword string
		setupMock   func(*MockPlatformUserRepository)
		wantErr     bool
		expectedErr error
	}{
		{
			name:        "successful password change",
			userID:      userID,
			oldPassword: "oldpassword",
			newPassword: "newpassword123",
			setupMock: func(m *MockPlatformUserRepository) {
				m.On("GetByID", mock.Anything, userID).Return(&models.PlatformUser{
					ID:       userID,
					Password: string(oldHashedPassword),
				}, nil)
				m.On("Update", mock.Anything, mock.MatchedBy(func(u *models.PlatformUser) bool {
					// Verify new password is hashed correctly
					err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte("newpassword123"))
					return err == nil
				})).Return(nil)
			},
			wantErr: false,
		},
		{
			name:        "incorrect old password",
			userID:      userID,
			oldPassword: "wrongoldpassword",
			newPassword: "newpassword123",
			setupMock: func(m *MockPlatformUserRepository) {
				m.On("GetByID", mock.Anything, userID).Return(&models.PlatformUser{
					ID:       userID,
					Password: string(oldHashedPassword),
				}, nil)
			},
			wantErr:     true,
			expectedErr: ErrOldPasswordIncorrect,
		},
		{
			name:        "new password too short",
			userID:      userID,
			oldPassword: "oldpassword",
			newPassword: "short",
			setupMock: func(m *MockPlatformUserRepository) {
				m.On("GetByID", mock.Anything, userID).Return(&models.PlatformUser{
					ID:       userID,
					Password: string(oldHashedPassword),
				}, nil)
			},
			wantErr:     true,
			expectedErr: ErrPasswordTooShort,
		},
		{
			name:        "user not found",
			userID:      userID,
			oldPassword: "oldpassword",
			newPassword: "newpassword123",
			setupMock: func(m *MockPlatformUserRepository) {
				m.On("GetByID", mock.Anything, userID).Return(nil, errors.New("user not found"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockPlatformUserRepository)
			tt.setupMock(mockRepo)

			service := NewPlatformUserService(mockRepo)
			err := service.ChangePassword(context.Background(), tt.userID, tt.oldPassword, tt.newPassword)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.expectedErr != nil {
					assert.ErrorIs(t, err, tt.expectedErr)
				}
			} else {
				assert.NoError(t, err)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

// TestPlatformUserService_DeleteUser tests user deletion with super_admin protection
func TestPlatformUserService_DeleteUser(t *testing.T) {
	regularUserID := uuid.New()
	superAdminUserID := uuid.New()

	tests := []struct {
		name        string
		userID      uuid.UUID
		setupMock   func(*MockPlatformUserRepository)
		wantErr     bool
		expectedErr error
	}{
		{
			name:   "successfully delete regular user",
			userID: regularUserID,
			setupMock: func(m *MockPlatformUserRepository) {
				m.On("GetUserRoles", mock.Anything, regularUserID).Return([]*models.PlatformRole{
					{ID: 2, Name: "platform_admin"},
				}, nil)
				m.On("Delete", mock.Anything, regularUserID).Return(nil)
			},
			wantErr: false,
		},
		{
			name:   "cannot delete super_admin user",
			userID: superAdminUserID,
			setupMock: func(m *MockPlatformUserRepository) {
				m.On("GetUserRoles", mock.Anything, superAdminUserID).Return([]*models.PlatformRole{
					{ID: 1, Name: "super_admin"},
				}, nil)
			},
			wantErr:     true,
			expectedErr: ErrCannotDeleteSuperAdmin,
		},
		{
			name:   "error getting user roles",
			userID: regularUserID,
			setupMock: func(m *MockPlatformUserRepository) {
				m.On("GetUserRoles", mock.Anything, regularUserID).Return(nil, errors.New("database error"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockPlatformUserRepository)
			tt.setupMock(mockRepo)

			service := NewPlatformUserService(mockRepo)
			err := service.DeleteUser(context.Background(), tt.userID)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.expectedErr != nil {
					assert.ErrorIs(t, err, tt.expectedErr)
				}
			} else {
				assert.NoError(t, err)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

// TestPlatformUserService_AssignRole tests role assignment
func TestPlatformUserService_AssignRole(t *testing.T) {
	userID := uuid.New()

	tests := []struct {
		name        string
		userID      uuid.UUID
		roleName    string
		setupMock   func(*MockPlatformUserRepository)
		wantErr     bool
		expectedErr error
	}{
		{
			name:     "successfully assign role",
			userID:   userID,
			roleName: "super_admin",
			setupMock: func(m *MockPlatformUserRepository) {
				m.On("AssignRole", mock.Anything, userID, 1).Return(nil)
			},
			wantErr: false,
		},
		{
			name:        "invalid role name",
			userID:      userID,
			roleName:    "nonexistent_role",
			setupMock:   func(m *MockPlatformUserRepository) {},
			wantErr:     true,
			expectedErr: ErrRoleNotFound,
		},
		{
			name:     "repository error",
			userID:   userID,
			roleName: "super_admin",
			setupMock: func(m *MockPlatformUserRepository) {
				m.On("AssignRole", mock.Anything, userID, 1).Return(errors.New("database error"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockPlatformUserRepository)
			tt.setupMock(mockRepo)

			service := NewPlatformUserService(mockRepo)
			err := service.AssignRole(context.Background(), tt.userID, tt.roleName)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.expectedErr != nil {
					assert.ErrorIs(t, err, tt.expectedErr)
				}
			} else {
				assert.NoError(t, err)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

// TestPlatformUserService_HasPermission tests permission checking
func TestPlatformUserService_HasPermission(t *testing.T) {
	userID := uuid.New()

	tests := []struct {
		name       string
		userID     uuid.UUID
		resource   string
		action     string
		setupMock  func(*MockPlatformUserRepository)
		wantResult bool
		wantErr    bool
	}{
		{
			name:     "user has permission",
			userID:   userID,
			resource: "users",
			action:   "create",
			setupMock: func(m *MockPlatformUserRepository) {
				m.On("GetUserPermissions", mock.Anything, userID).Return([]*models.PlatformPermission{
					{
						ID:       1,
						Resource: "users",
						Action:   "create",
					},
					{
						ID:       2,
						Resource: "users",
						Action:   "read",
					},
				}, nil)
			},
			wantResult: true,
			wantErr:    false,
		},
		{
			name:     "user does not have permission",
			userID:   userID,
			resource: "clients",
			action:   "delete",
			setupMock: func(m *MockPlatformUserRepository) {
				m.On("GetUserPermissions", mock.Anything, userID).Return([]*models.PlatformPermission{
					{
						ID:       1,
						Resource: "users",
						Action:   "create",
					},
				}, nil)
			},
			wantResult: false,
			wantErr:    false,
		},
		{
			name:     "error getting permissions",
			userID:   userID,
			resource: "users",
			action:   "create",
			setupMock: func(m *MockPlatformUserRepository) {
				m.On("GetUserPermissions", mock.Anything, userID).Return(nil, errors.New("database error"))
			},
			wantResult: false,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockPlatformUserRepository)
			tt.setupMock(mockRepo)

			service := NewPlatformUserService(mockRepo)
			hasPermission, err := service.HasPermission(context.Background(), tt.userID, tt.resource, tt.action)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantResult, hasPermission)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

// TestPlatformUserService_GetUser tests getting a user by ID
func TestPlatformUserService_GetUser(t *testing.T) {
	userID := uuid.New()

	tests := []struct {
		name      string
		userID    uuid.UUID
		setupMock func(*MockPlatformUserRepository)
		wantErr   bool
	}{
		{
			name:   "successfully get user",
			userID: userID,
			setupMock: func(m *MockPlatformUserRepository) {
				m.On("GetByID", mock.Anything, userID).Return(&models.PlatformUser{
					ID:       userID,
					Email:    "test@example.com",
					Username: "testuser",
				}, nil)
			},
			wantErr: false,
		},
		{
			name:   "user not found",
			userID: userID,
			setupMock: func(m *MockPlatformUserRepository) {
				m.On("GetByID", mock.Anything, userID).Return(nil, errors.New("user not found"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockPlatformUserRepository)
			tt.setupMock(mockRepo)

			service := NewPlatformUserService(mockRepo)
			user, err := service.GetUser(context.Background(), tt.userID)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, user)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, user)
				assert.Equal(t, tt.userID, user.ID)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

// TestPlatformUserService_UpdateUser tests updating a user
func TestPlatformUserService_UpdateUser(t *testing.T) {
	userID := uuid.New()
	newEmail := "newemail@example.com"
	newFirstName := "NewFirst"

	tests := []struct {
		name      string
		userID    uuid.UUID
		request   *UpdatePlatformUserRequest
		setupMock func(*MockPlatformUserRepository)
		wantErr   bool
	}{
		{
			name:   "successfully update user",
			userID: userID,
			request: &UpdatePlatformUserRequest{
				Email:     &newEmail,
				FirstName: &newFirstName,
			},
			setupMock: func(m *MockPlatformUserRepository) {
				m.On("GetByID", mock.Anything, userID).Return(&models.PlatformUser{
					ID:       userID,
					Email:    "old@example.com",
					Username: "testuser",
				}, nil)
				m.On("Update", mock.Anything, mock.MatchedBy(func(u *models.PlatformUser) bool {
					return u.Email == newEmail && u.FirstName == newFirstName
				})).Return(nil)
			},
			wantErr: false,
		},
		{
			name:   "user not found",
			userID: userID,
			request: &UpdatePlatformUserRequest{
				Email: &newEmail,
			},
			setupMock: func(m *MockPlatformUserRepository) {
				m.On("GetByID", mock.Anything, userID).Return(nil, errors.New("user not found"))
			},
			wantErr: true,
		},
		{
			name:   "invalid email format",
			userID: userID,
			request: &UpdatePlatformUserRequest{
				Email: func() *string { s := "invalid-email"; return &s }(),
			},
			setupMock: func(m *MockPlatformUserRepository) {
				m.On("GetByID", mock.Anything, userID).Return(&models.PlatformUser{
					ID: userID,
				}, nil)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockPlatformUserRepository)
			tt.setupMock(mockRepo)

			service := NewPlatformUserService(mockRepo)
			err := service.UpdateUser(context.Background(), tt.userID, tt.request)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

// TestPlatformUserService_ListUsers tests listing users with pagination
func TestPlatformUserService_ListUsers(t *testing.T) {
	tests := []struct {
		name         string
		page         int
		pageSize     int
		setupMock    func(*MockPlatformUserRepository)
		wantErr      bool
		expectedPage int
		expectedSize int
	}{
		{
			name:     "successfully list users",
			page:     1,
			pageSize: 10,
			setupMock: func(m *MockPlatformUserRepository) {
				m.On("List", mock.Anything, 10, 0).Return([]*models.PlatformUser{
					{ID: uuid.New(), Username: "user1"},
					{ID: uuid.New(), Username: "user2"},
				}, nil)
			},
			wantErr:      false,
			expectedPage: 1,
			expectedSize: 10,
		},
		{
			name:     "default to page 1 if invalid",
			page:     0,
			pageSize: 10,
			setupMock: func(m *MockPlatformUserRepository) {
				m.On("List", mock.Anything, 10, 0).Return([]*models.PlatformUser{}, nil)
			},
			wantErr:      false,
			expectedPage: 1,
			expectedSize: 10,
		},
		{
			name:     "cap page size at 100",
			page:     1,
			pageSize: 200,
			setupMock: func(m *MockPlatformUserRepository) {
				m.On("List", mock.Anything, 100, 0).Return([]*models.PlatformUser{}, nil)
			},
			wantErr:      false,
			expectedPage: 1,
			expectedSize: 100,
		},
		{
			name:     "page 2 with correct offset",
			page:     2,
			pageSize: 10,
			setupMock: func(m *MockPlatformUserRepository) {
				m.On("List", mock.Anything, 10, 10).Return([]*models.PlatformUser{}, nil)
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockPlatformUserRepository)
			tt.setupMock(mockRepo)

			service := NewPlatformUserService(mockRepo)
			users, count, err := service.ListUsers(context.Background(), tt.page, tt.pageSize)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, users)
				assert.GreaterOrEqual(t, count, 0)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

// TestPlatformUserService_GetUserPermissions tests getting user permissions
func TestPlatformUserService_GetUserPermissions(t *testing.T) {
	userID := uuid.New()

	tests := []struct {
		name      string
		userID    uuid.UUID
		setupMock func(*MockPlatformUserRepository)
		wantErr   bool
	}{
		{
			name:   "successfully get permissions",
			userID: userID,
			setupMock: func(m *MockPlatformUserRepository) {
				m.On("GetUserPermissions", mock.Anything, userID).Return([]*models.PlatformPermission{
					{ID: 1, Resource: "users", Action: "create"},
					{ID: 2, Resource: "users", Action: "read"},
				}, nil)
			},
			wantErr: false,
		},
		{
			name:   "repository error",
			userID: userID,
			setupMock: func(m *MockPlatformUserRepository) {
				m.On("GetUserPermissions", mock.Anything, userID).Return(nil, errors.New("database error"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockPlatformUserRepository)
			tt.setupMock(mockRepo)

			service := NewPlatformUserService(mockRepo)
			permissions, err := service.GetUserPermissions(context.Background(), tt.userID)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, permissions)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, permissions)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

// TestPlatformUserService_RemoveRole tests role removal with protection
func TestPlatformUserService_RemoveRole(t *testing.T) {
	userID := uuid.New()

	tests := []struct {
		name        string
		userID      uuid.UUID
		roleName    string
		setupMock   func(*MockPlatformUserRepository)
		wantErr     bool
		expectedErr error
	}{
		{
			name:     "successfully remove non-super_admin role",
			userID:   userID,
			roleName: "platform_admin",
			setupMock: func(m *MockPlatformUserRepository) {
				m.On("RemoveRole", mock.Anything, userID, 2).Return(nil)
			},
			wantErr: false,
		},
		{
			name:     "prevent removing super_admin role",
			userID:   userID,
			roleName: "super_admin",
			setupMock: func(m *MockPlatformUserRepository) {
				m.On("GetUserRoles", mock.Anything, userID).Return([]*models.PlatformRole{
					{ID: 1, Name: "super_admin"},
				}, nil)
			},
			wantErr:     true,
			expectedErr: ErrCannotRemoveLastSuperAdmin,
		},
		{
			name:        "invalid role name",
			userID:      userID,
			roleName:    "invalid_role",
			setupMock:   func(m *MockPlatformUserRepository) {},
			wantErr:     true,
			expectedErr: ErrRoleNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockPlatformUserRepository)
			tt.setupMock(mockRepo)

			service := NewPlatformUserService(mockRepo)
			err := service.RemoveRole(context.Background(), tt.userID, tt.roleName)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.expectedErr != nil {
					assert.ErrorIs(t, err, tt.expectedErr)
				}
			} else {
				assert.NoError(t, err)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

// Test helper functions

func TestValidateEmail(t *testing.T) {
	tests := []struct {
		name    string
		email   string
		wantErr bool
	}{
		{"valid email", "test@example.com", false},
		{"valid email with subdomain", "user@mail.example.com", false},
		{"invalid email no @", "notanemail", true},
		{"invalid email no domain", "test@", true},
		{"invalid email no user", "@example.com", true},
		{"empty email", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateEmail(tt.email)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateUsername(t *testing.T) {
	tests := []struct {
		name     string
		username string
		wantErr  bool
	}{
		{"valid username", "testuser", false},
		{"valid with numbers", "user123", false},
		{"valid with underscore", "test_user", false},
		{"too short", "ab", true},
		{"too long", "abcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyz", true},
		{"with special chars", "test@user", true},
		{"with spaces", "test user", true},
		{"with dash", "test-user", true},
		{"empty", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateUsername(tt.username)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestHashPassword(t *testing.T) {
	password := "testpassword123"
	hash, err := hashPassword(password)

	assert.NoError(t, err)
	assert.NotEmpty(t, hash)
	assert.NotEqual(t, password, hash)

	// Verify hash is valid
	err = bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	assert.NoError(t, err)
}

func TestVerifyPassword(t *testing.T) {
	password := "testpassword123"
	hash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	tests := []struct {
		name     string
		hash     string
		password string
		wantErr  bool
	}{
		{"correct password", string(hash), password, false},
		{"incorrect password", string(hash), "wrongpassword", true},
		{"empty password", string(hash), "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := verifyPassword(tt.hash, tt.password)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestGetRoleIDByName(t *testing.T) {
	tests := []struct {
		name     string
		roleName string
		wantID   int
		wantErr  bool
	}{
		{"super_admin", "super_admin", 1, false},
		{"platform_admin", "platform_admin", 2, false},
		{"support", "support", 3, false},
		{"invalid role", "invalid_role", 0, true},
		{"empty role", "", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id, err := getRoleIDByName(tt.roleName)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Equal(t, 0, id)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantID, id)
			}
		})
	}
}
