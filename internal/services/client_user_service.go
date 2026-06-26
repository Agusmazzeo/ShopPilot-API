package services

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/yourorg/shoppilot/app/repositories"
	"github.com/yourorg/shoppilot/internal/models"
	"golang.org/x/crypto/bcrypt"
)

// ClientUserService defines the interface for client user business logic
type ClientUserService interface {
	// User management
	CreateUser(ctx context.Context, clientID uuid.UUID, req *CreateClientUserRequest) (*models.ClientUser, error)
	GetUser(ctx context.Context, userID uuid.UUID) (*models.ClientUser, error)
	UpdateUser(ctx context.Context, userID uuid.UUID, req *UpdateClientUserRequest) error
	DeleteUser(ctx context.Context, userID uuid.UUID) error
	ListUsers(ctx context.Context, clientID uuid.UUID, page, pageSize int) ([]*models.ClientUser, int, error)

	// Authentication
	Authenticate(ctx context.Context, clientID uuid.UUID, username, password string) (*models.ClientUser, error)
	ChangePassword(ctx context.Context, userID uuid.UUID, oldPassword, newPassword string) error

	// Role management
	AssignRole(ctx context.Context, userID uuid.UUID, roleName string) error
	RemoveRole(ctx context.Context, userID uuid.UUID, roleName string) error
	GetUserPermissions(ctx context.Context, userID uuid.UUID) ([]*models.ClientPermission, error)
	HasPermission(ctx context.Context, userID uuid.UUID, resource, action string) (bool, error)
}

// CreateClientUserRequest contains the data needed to create a client user
type CreateClientUserRequest struct {
	Email        string
	Username     string
	Password     string
	FirstName    string
	LastName     string
	Phone        string
	AvatarURL    *string
	UserStatusID int
}

// UpdateClientUserRequest contains the data that can be updated for a client user
type UpdateClientUserRequest struct {
	Email        *string
	Username     *string
	FirstName    *string
	LastName     *string
	Phone        *string
	AvatarURL    *string
	UserStatusID *int
}

// clientUserService is the concrete implementation
type clientUserService struct {
	repo repositories.ClientUserRepository
}

// NewClientUserService creates a new client user service
func NewClientUserService(repo repositories.ClientUserRepository) ClientUserService {
	return &clientUserService{
		repo: repo,
	}
}

// CreateUser creates a new client user with hashed password
func (s *clientUserService) CreateUser(ctx context.Context, clientID uuid.UUID, req *CreateClientUserRequest) (*models.ClientUser, error) {
	// Validate input
	if err := s.validateCreateRequest(req); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Check if email already exists for this client
	existingUser, err := s.repo.GetByEmail(ctx, clientID, req.Email)
	if err == nil && existingUser != nil {
		return nil, fmt.Errorf("email already exists for this client")
	}

	// Check if username already exists for this client
	existingUser, err = s.repo.GetByUsername(ctx, clientID, req.Username)
	if err == nil && existingUser != nil {
		return nil, fmt.Errorf("username already exists for this client")
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Create user model
	user := &models.ClientUser{
		ClientID:     clientID,
		Email:        req.Email,
		Username:     req.Username,
		Password:     string(hashedPassword),
		FirstName:    req.FirstName,
		LastName:     req.LastName,
		Phone:        req.Phone,
		AvatarURL:    req.AvatarURL,
		UserStatusID: req.UserStatusID,
	}

	// Create in database
	if err := s.repo.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return user, nil
}

// GetUser retrieves a client user by ID
func (s *clientUserService) GetUser(ctx context.Context, userID uuid.UUID) (*models.ClientUser, error) {
	user, err := s.repo.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	return user, nil
}

// UpdateUser updates an existing client user
func (s *clientUserService) UpdateUser(ctx context.Context, userID uuid.UUID, req *UpdateClientUserRequest) error {
	// Get existing user
	user, err := s.repo.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	// Update fields if provided
	if req.Email != nil {
		// Check if new email already exists for this client (excluding current user)
		existingUser, err := s.repo.GetByEmail(ctx, user.ClientID, *req.Email)
		if err == nil && existingUser != nil && existingUser.ID != userID {
			return fmt.Errorf("email already exists for this client")
		}
		user.Email = *req.Email
	}

	if req.Username != nil {
		// Check if new username already exists for this client (excluding current user)
		existingUser, err := s.repo.GetByUsername(ctx, user.ClientID, *req.Username)
		if err == nil && existingUser != nil && existingUser.ID != userID {
			return fmt.Errorf("username already exists for this client")
		}
		user.Username = *req.Username
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

	if req.UserStatusID != nil {
		user.UserStatusID = *req.UserStatusID
	}

	// Update in database
	if err := s.repo.Update(ctx, user); err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	return nil
}

// DeleteUser deletes a client user, preventing deletion of the last client_admin
func (s *clientUserService) DeleteUser(ctx context.Context, userID uuid.UUID) error {
	// Get user to check client
	user, err := s.repo.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	// Check if user has client_admin role
	roles, err := s.repo.GetUserRoles(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get user roles: %w", err)
	}

	hasClientAdmin := false
	for _, role := range roles {
		if role.Name == "client_admin" {
			hasClientAdmin = true
			break
		}
	}

	// If user is a client_admin, check if they're the last one
	if hasClientAdmin {
		// Get all users for this client
		allUsers, err := s.repo.ListByClient(ctx, user.ClientID, 1000, 0) // Large limit to get all
		if err != nil {
			return fmt.Errorf("failed to list client users: %w", err)
		}

		// Count client_admin users
		adminCount := 0
		for _, u := range allUsers {
			userRoles, err := s.repo.GetUserRoles(ctx, u.ID)
			if err != nil {
				continue
			}
			for _, role := range userRoles {
				if role.Name == "client_admin" {
					adminCount++
					break
				}
			}
		}

		// Prevent deletion of last client_admin
		if adminCount <= 1 {
			return fmt.Errorf("cannot delete the last client_admin user")
		}
	}

	// Delete user
	if err := s.repo.Delete(ctx, userID); err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	return nil
}

// ListUsers retrieves all users for a client with pagination
func (s *clientUserService) ListUsers(ctx context.Context, clientID uuid.UUID, page, pageSize int) ([]*models.ClientUser, int, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	offset := (page - 1) * pageSize

	users, err := s.repo.ListByClient(ctx, clientID, pageSize, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list users: %w", err)
	}

	// For simplicity, we're not implementing total count here
	// In production, you'd want a separate count query
	total := len(users)

	return users, total, nil
}

// Authenticate verifies username and password for a client user
func (s *clientUserService) Authenticate(ctx context.Context, clientID uuid.UUID, username, password string) (*models.ClientUser, error) {
	// Get user by username within client
	user, err := s.repo.GetByUsername(ctx, clientID, username)
	if err != nil {
		return nil, fmt.Errorf("invalid credentials")
	}

	// Compare password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return nil, fmt.Errorf("invalid credentials")
	}

	return user, nil
}

// ChangePassword changes a user's password after verifying the old password
func (s *clientUserService) ChangePassword(ctx context.Context, userID uuid.UUID, oldPassword, newPassword string) error {
	// Get user
	user, err := s.repo.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	// Verify old password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(oldPassword)); err != nil {
		return fmt.Errorf("invalid old password")
	}

	// Validate new password
	if len(newPassword) < 8 {
		return fmt.Errorf("new password must be at least 8 characters")
	}

	// Hash new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash new password: %w", err)
	}

	// Update password
	user.Password = string(hashedPassword)
	if err := s.repo.Update(ctx, user); err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	return nil
}

// AssignRole assigns a role to a user by role name
func (s *clientUserService) AssignRole(ctx context.Context, userID uuid.UUID, roleName string) error {
	// For this implementation, we need to map role name to role ID
	// In a real implementation, you'd query the client_roles table
	roleIDMap := map[string]int{
		"client_admin":      1,
		"shop_manager":      2,
		"inventory_manager": 3,
		"viewer":            4,
	}

	roleID, ok := roleIDMap[roleName]
	if !ok {
		return fmt.Errorf("invalid role name: %s", roleName)
	}

	if err := s.repo.AssignRole(ctx, userID, roleID); err != nil {
		return fmt.Errorf("failed to assign role: %w", err)
	}

	return nil
}

// RemoveRole removes a role from a user by role name
func (s *clientUserService) RemoveRole(ctx context.Context, userID uuid.UUID, roleName string) error {
	// Prevent removing client_admin if it's the last one
	if roleName == "client_admin" {
		user, err := s.repo.GetByID(ctx, userID)
		if err != nil {
			return fmt.Errorf("user not found: %w", err)
		}

		// Get all users for this client
		allUsers, err := s.repo.ListByClient(ctx, user.ClientID, 1000, 0)
		if err != nil {
			return fmt.Errorf("failed to list client users: %w", err)
		}

		// Count client_admin users
		adminCount := 0
		for _, u := range allUsers {
			userRoles, err := s.repo.GetUserRoles(ctx, u.ID)
			if err != nil {
				continue
			}
			for _, role := range userRoles {
				if role.Name == "client_admin" {
					adminCount++
					break
				}
			}
		}

		// Prevent removal of last client_admin role
		if adminCount <= 1 {
			return fmt.Errorf("cannot remove the last client_admin role")
		}
	}

	// Map role name to role ID
	roleIDMap := map[string]int{
		"client_admin":      1,
		"shop_manager":      2,
		"inventory_manager": 3,
		"viewer":            4,
	}

	roleID, ok := roleIDMap[roleName]
	if !ok {
		return fmt.Errorf("invalid role name: %s", roleName)
	}

	if err := s.repo.RemoveRole(ctx, userID, roleID); err != nil {
		return fmt.Errorf("failed to remove role: %w", err)
	}

	return nil
}

// GetUserPermissions retrieves all permissions for a user
func (s *clientUserService) GetUserPermissions(ctx context.Context, userID uuid.UUID) ([]*models.ClientPermission, error) {
	permissions, err := s.repo.GetUserPermissions(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user permissions: %w", err)
	}

	return permissions, nil
}

// HasPermission checks if a user has a specific permission
func (s *clientUserService) HasPermission(ctx context.Context, userID uuid.UUID, resource, action string) (bool, error) {
	permissions, err := s.repo.GetUserPermissions(ctx, userID)
	if err != nil {
		return false, fmt.Errorf("failed to get user permissions: %w", err)
	}

	for _, perm := range permissions {
		if perm.Resource == resource && perm.Action == action {
			return true, nil
		}
	}

	return false, nil
}

// validateCreateRequest validates the create user request
func (s *clientUserService) validateCreateRequest(req *CreateClientUserRequest) error {
	if req.Email == "" {
		return fmt.Errorf("email is required")
	}

	if !strings.Contains(req.Email, "@") {
		return fmt.Errorf("invalid email format")
	}

	if req.Username == "" {
		return fmt.Errorf("username is required")
	}

	if len(req.Username) < 3 || len(req.Username) > 50 {
		return fmt.Errorf("username must be between 3 and 50 characters")
	}

	if req.Password == "" {
		return fmt.Errorf("password is required")
	}

	if len(req.Password) < 8 {
		return fmt.Errorf("password must be at least 8 characters")
	}

	if req.UserStatusID <= 0 {
		return fmt.Errorf("user status ID is required")
	}

	return nil
}
