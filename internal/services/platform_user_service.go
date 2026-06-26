package services

import (
	"context"
	"errors"
	"fmt"
	"net/mail"
	"regexp"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/yourorg/shoppilot/app/repositories"
	"github.com/yourorg/shoppilot/internal/models"
)

var (
	// ErrInvalidEmail indicates the email format is invalid
	ErrInvalidEmail = errors.New("invalid email format")
	// ErrInvalidUsername indicates the username does not meet constraints
	ErrInvalidUsername = errors.New("username must be alphanumeric and 3-50 characters long")
	// ErrPasswordTooShort indicates the password is too short
	ErrPasswordTooShort = errors.New("password must be at least 8 characters")
	// ErrCannotDeleteSuperAdmin indicates attempting to delete a super admin user
	ErrCannotDeleteSuperAdmin = errors.New("cannot delete super_admin user")
	// ErrCannotRemoveLastSuperAdmin indicates attempting to remove the last super_admin role
	ErrCannotRemoveLastSuperAdmin = errors.New("cannot remove last super_admin role")
	// ErrInvalidCredentials indicates authentication failed
	ErrInvalidCredentials = errors.New("invalid username or password")
	// ErrOldPasswordIncorrect indicates the old password is incorrect
	ErrOldPasswordIncorrect = errors.New("old password is incorrect")
	// ErrRoleNotFound indicates the role was not found
	ErrRoleNotFound = errors.New("role not found")
)

// CreatePlatformUserRequest contains the data needed to create a new platform user
type CreatePlatformUserRequest struct {
	Email     string
	Username  string
	Password  string
	FirstName string
	LastName  string
	Phone     string
}

// UpdatePlatformUserRequest contains the data needed to update a platform user
type UpdatePlatformUserRequest struct {
	Email     *string
	FirstName *string
	LastName  *string
	Phone     *string
	AvatarURL *string
}

// PlatformUserService defines the interface for platform user business logic
type PlatformUserService interface {
	// User management
	CreateUser(ctx context.Context, req *CreatePlatformUserRequest) (*models.PlatformUser, error)
	GetUser(ctx context.Context, id uuid.UUID) (*models.PlatformUser, error)
	UpdateUser(ctx context.Context, id uuid.UUID, req *UpdatePlatformUserRequest) error
	DeleteUser(ctx context.Context, id uuid.UUID) error
	ListUsers(ctx context.Context, page, pageSize int) ([]*models.PlatformUser, int, error)

	// Authentication
	Authenticate(ctx context.Context, username, password string) (*models.PlatformUser, error)
	ChangePassword(ctx context.Context, userID uuid.UUID, oldPassword, newPassword string) error

	// Role management
	AssignRole(ctx context.Context, userID uuid.UUID, roleName string) error
	RemoveRole(ctx context.Context, userID uuid.UUID, roleName string) error
	GetUserPermissions(ctx context.Context, userID uuid.UUID) ([]*models.PlatformPermission, error)
	HasPermission(ctx context.Context, userID uuid.UUID, resource, action string) (bool, error)
}

// platformUserService is the concrete implementation of PlatformUserService
type platformUserService struct {
	repo repositories.PlatformUserRepository
}

// NewPlatformUserService creates a new instance of platform user service
func NewPlatformUserService(repo repositories.PlatformUserRepository) PlatformUserService {
	return &platformUserService{repo: repo}
}

// CreateUser creates a new platform user with validation and password hashing
func (s *platformUserService) CreateUser(ctx context.Context, req *CreatePlatformUserRequest) (*models.PlatformUser, error) {
	// Validate email format
	if err := validateEmail(req.Email); err != nil {
		return nil, err
	}

	// Validate username constraints
	if err := validateUsername(req.Username); err != nil {
		return nil, err
	}

	// Validate password length
	if len(req.Password) < 8 {
		return nil, ErrPasswordTooShort
	}

	// Hash password
	hashedPassword, err := hashPassword(req.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Create user model
	user := &models.PlatformUser{
		Email:        req.Email,
		Username:     req.Username,
		Password:     hashedPassword,
		FirstName:    req.FirstName,
		LastName:     req.LastName,
		Phone:        req.Phone,
		UserStatusID: 1, // Default to active
	}

	// Save to repository
	if err := s.repo.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return user, nil
}

// GetUser retrieves a platform user by ID
func (s *platformUserService) GetUser(ctx context.Context, id uuid.UUID) (*models.PlatformUser, error) {
	user, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	return user, nil
}

// UpdateUser updates an existing platform user
func (s *platformUserService) UpdateUser(ctx context.Context, id uuid.UUID, req *UpdatePlatformUserRequest) error {
	// Get existing user
	user, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}

	// Update fields if provided
	if req.Email != nil {
		if err := validateEmail(*req.Email); err != nil {
			return err
		}
		user.Email = *req.Email
	}

	if req.FirstName != nil {
		user.FirstName = *req.FirstName
	}

	if req.LastName != nil {
		user.LastName = *req.LastName
	}

	if req.Phone != nil {
		user.Phone = *req.Phone
	}

	if req.AvatarURL != nil {
		user.AvatarURL = req.AvatarURL
	}

	// Save updates
	if err := s.repo.Update(ctx, user); err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	return nil
}

// DeleteUser deletes a platform user with super_admin protection
func (s *platformUserService) DeleteUser(ctx context.Context, id uuid.UUID) error {
	// Check if user has super_admin role
	roles, err := s.repo.GetUserRoles(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get user roles: %w", err)
	}

	for _, role := range roles {
		if role.Name == "super_admin" {
			return ErrCannotDeleteSuperAdmin
		}
	}

	// Delete user
	if err := s.repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	return nil
}

// ListUsers retrieves a paginated list of platform users
func (s *platformUserService) ListUsers(ctx context.Context, page, pageSize int) ([]*models.PlatformUser, int, error) {
	// Validate pagination parameters
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	if pageSize > 100 {
		pageSize = 100
	}

	// Calculate offset
	offset := (page - 1) * pageSize

	// Get users
	users, err := s.repo.List(ctx, pageSize, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list users: %w", err)
	}

	// For now, we return users without total count
	// In a production system, you'd query the total count
	return users, len(users), nil
}

// Authenticate verifies user credentials and returns the user if valid
func (s *platformUserService) Authenticate(ctx context.Context, username, password string) (*models.PlatformUser, error) {
	// Get user by username
	user, err := s.repo.GetByUsername(ctx, username)
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	// Verify password
	if err := verifyPassword(user.Password, password); err != nil {
		return nil, ErrInvalidCredentials
	}

	return user, nil
}

// ChangePassword changes a user's password after verifying the old password
func (s *platformUserService) ChangePassword(ctx context.Context, userID uuid.UUID, oldPassword, newPassword string) error {
	// Get user
	user, err := s.repo.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}

	// Verify old password
	if err := verifyPassword(user.Password, oldPassword); err != nil {
		return ErrOldPasswordIncorrect
	}

	// Validate new password length
	if len(newPassword) < 8 {
		return ErrPasswordTooShort
	}

	// Hash new password
	hashedPassword, err := hashPassword(newPassword)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	// Update password
	user.Password = hashedPassword
	if err := s.repo.Update(ctx, user); err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	return nil
}

// AssignRole assigns a role to a user by role name
func (s *platformUserService) AssignRole(ctx context.Context, userID uuid.UUID, roleName string) error {
	// Get role ID by name (we'll need to add this to repository or handle it differently)
	// For now, we'll use a simple mapping - in production, you'd query the platform_roles table
	roleID, err := getRoleIDByName(roleName)
	if err != nil {
		return err
	}

	// Assign role
	if err := s.repo.AssignRole(ctx, userID, roleID); err != nil {
		return fmt.Errorf("failed to assign role: %w", err)
	}

	return nil
}

// RemoveRole removes a role from a user with protection for last super_admin
func (s *platformUserService) RemoveRole(ctx context.Context, userID uuid.UUID, roleName string) error {
	// Get role ID by name
	roleID, err := getRoleIDByName(roleName)
	if err != nil {
		return err
	}

	// If removing super_admin role, check if this is the last one
	if roleName == "super_admin" {
		// Get current user roles
		roles, err := s.repo.GetUserRoles(ctx, userID)
		if err != nil {
			return fmt.Errorf("failed to get user roles: %w", err)
		}

		// Check if user has super_admin role
		hasSuperAdmin := false
		for _, role := range roles {
			if role.Name == "super_admin" {
				hasSuperAdmin = true
				break
			}
		}

		if hasSuperAdmin {
			// TODO: Check if this is the last super_admin in the system
			// This would require a repository method to count super_admins
			// For now, we'll just prevent removal
			return ErrCannotRemoveLastSuperAdmin
		}
	}

	// Remove role
	if err := s.repo.RemoveRole(ctx, userID, roleID); err != nil {
		return fmt.Errorf("failed to remove role: %w", err)
	}

	return nil
}

// GetUserPermissions retrieves all permissions for a user
func (s *platformUserService) GetUserPermissions(ctx context.Context, userID uuid.UUID) ([]*models.PlatformPermission, error) {
	permissions, err := s.repo.GetUserPermissions(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user permissions: %w", err)
	}
	return permissions, nil
}

// HasPermission checks if a user has a specific permission
func (s *platformUserService) HasPermission(ctx context.Context, userID uuid.UUID, resource, action string) (bool, error) {
	permissions, err := s.repo.GetUserPermissions(ctx, userID)
	if err != nil {
		return false, fmt.Errorf("failed to get user permissions: %w", err)
	}

	for _, permission := range permissions {
		if permission.Resource == resource && permission.Action == action {
			return true, nil
		}
	}

	return false, nil
}

// Helper functions

// validateEmail validates email format
func validateEmail(email string) error {
	_, err := mail.ParseAddress(email)
	if err != nil {
		return ErrInvalidEmail
	}
	return nil
}

// validateUsername validates username constraints (alphanumeric, 3-50 chars)
func validateUsername(username string) error {
	if len(username) < 3 || len(username) > 50 {
		return ErrInvalidUsername
	}

	// Check if alphanumeric (letters, numbers, underscore allowed)
	matched, err := regexp.MatchString(`^[a-zA-Z0-9_]+$`, username)
	if err != nil {
		return fmt.Errorf("failed to validate username: %w", err)
	}

	if !matched {
		return ErrInvalidUsername
	}

	return nil
}

// hashPassword hashes a password using bcrypt
func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// verifyPassword verifies a password against a hash
func verifyPassword(hashedPassword, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}

// getRoleIDByName maps role names to IDs
// TODO: This should be replaced with a repository method to query platform_roles table
func getRoleIDByName(roleName string) (int, error) {
	roleMap := map[string]int{
		"super_admin":    1,
		"platform_admin": 2,
		"support":        3,
	}

	if id, exists := roleMap[roleName]; exists {
		return id, nil
	}

	return 0, ErrRoleNotFound
}
