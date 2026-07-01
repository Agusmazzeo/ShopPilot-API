package utils

import (
	"context"

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
