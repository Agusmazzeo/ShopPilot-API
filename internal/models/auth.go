package models

import (
	"time"

	"github.com/google/uuid"
)

// AuthUser represents a user in an authentication context (can be PlatformUser or ClientUser)
type AuthUser struct {
	ID           uuid.UUID    `json:"id"`
	Email        string       `json:"email"`
	Username     string       `json:"username"`
	FirstName    string       `json:"firstName"`
	LastName     string       `json:"lastName"`
	Phone        string       `json:"phone"`
	AvatarURL    *string      `json:"avatarUrl"`
	UserStatusID int          `json:"userStatusId"`
	RoleID       int          `json:"roleId"`
	RoleName     string       `json:"roleName"`
	Permissions  []Permission `json:"permissions"`
	LastLoginAt  *time.Time   `json:"lastLoginAt"`
	CreatedAt    time.Time    `json:"createdAt"`
	UpdatedAt    time.Time    `json:"updatedAt"`

	// Context-specific fields
	ClientID *uuid.UUID `json:"clientId,omitempty"` // Set for ClientUser, nil for PlatformUser
}

// Permission represents a permission for RBAC
type Permission struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Resource    string    `json:"resource"`
	Action      string    `json:"action"`
	CreatedAt   time.Time `json:"createdAt"`
}

// Role represents a user role
type Role struct {
	ID           int          `json:"id"`
	Name         string       `json:"name"`
	Description  string       `json:"description"`
	IsSystemRole bool         `json:"isSystemRole"`
	Permissions  []Permission `json:"permissions"`
	CreatedAt    time.Time    `json:"createdAt"`
	UpdatedAt    time.Time    `json:"updatedAt"`
}

// HasGlobalPermission checks if the user has a specific permission
func (u *AuthUser) HasGlobalPermission(resource, action string) bool {
	for _, permission := range u.Permissions {
		if permission.Resource == resource && permission.Action == action {
			return true
		}
	}
	return false
}

// ToPlatformUser converts AuthUser to PlatformUser (only if it was created from one)
func (u *AuthUser) ToPlatformUser() *PlatformUser {
	if u.ClientID != nil {
		return nil // This is a ClientUser, not a PlatformUser
	}
	return &PlatformUser{
		ID:           u.ID,
		Email:        u.Email,
		Username:     u.Username,
		FirstName:    u.FirstName,
		LastName:     u.LastName,
		Phone:        u.Phone,
		AvatarURL:    u.AvatarURL,
		UserStatusID: u.UserStatusID,
		LastLoginAt:  u.LastLoginAt,
		CreatedAt:    u.CreatedAt,
		UpdatedAt:    u.UpdatedAt,
	}
}

// ToClientUser converts AuthUser to ClientUser (only if it was created from one)
func (u *AuthUser) ToClientUser() *ClientUser {
	if u.ClientID == nil {
		return nil // This is a PlatformUser, not a ClientUser
	}
	return &ClientUser{
		ID:           u.ID,
		ClientID:     *u.ClientID,
		Email:        u.Email,
		Username:     u.Username,
		FirstName:    u.FirstName,
		LastName:     u.LastName,
		Phone:        u.Phone,
		AvatarURL:    u.AvatarURL,
		UserStatusID: u.UserStatusID,
		LastLoginAt:  u.LastLoginAt,
		CreatedAt:    u.CreatedAt,
		UpdatedAt:    u.UpdatedAt,
	}
}
