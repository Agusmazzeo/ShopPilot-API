package utils

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/yourorg/shoppilot/internal/models"
)

// contextKey is a custom type for context keys to avoid collisions
type contextKey string

const (
	UserContextKey            contextKey = "user"
	UserIDContextKey          contextKey = "userID"
	UserEmailContextKey       contextKey = "userEmail"
	UserRoleContextKey        contextKey = "userRole"
	UserPermissionsContextKey contextKey = "userPermissions"
)

// GetUserFromContext extracts the authenticated user from context
func GetUserFromContext(ctx context.Context) (*models.AuthUser, bool) {
	user, ok := ctx.Value(UserContextKey).(*models.AuthUser)
	return user, ok
}

// GetUserIDFromContext extracts the user ID from context
func GetUserIDFromContext(ctx context.Context) (uuid.UUID, bool) {
	userID, ok := ctx.Value(UserIDContextKey).(uuid.UUID)
	return userID, ok
}

// HasPermissionInContext checks if the current user has a specific permission
func HasPermissionInContext(ctx context.Context, resource, action string) bool {
	user, exists := GetUserFromContext(ctx)
	if !exists {
		return false
	}

	for _, permission := range user.Permissions {
		if permission.Resource == resource && permission.Action == action {
			return true
		}
	}
	return false
}

// GetClientIDFromContext extracts the client ID from the authenticated user
// Returns error if user is a platform user (no client context)
func GetClientIDFromContext(ctx context.Context) (uuid.UUID, error) {
	user, ok := GetUserFromContext(ctx)
	if !ok {
		return uuid.Nil, errors.New("no authenticated user in context")
	}

	if user.ClientID == nil {
		return uuid.Nil, errors.New("user has no client context")
	}

	return *user.ClientID, nil
}

// MustGetClientIDFromContext extracts client ID or panics
// Use only in handlers where client context is guaranteed
func MustGetClientIDFromContext(ctx context.Context) uuid.UUID {
	clientID, err := GetClientIDFromContext(ctx)
	if err != nil {
		panic(err) // Will be caught by recovery middleware
	}
	return clientID
}
