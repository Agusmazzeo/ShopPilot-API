package utils

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/yourorg/shoppilot/internal/models"
)

func TestGetUserFromContext(t *testing.T) {
	userID := uuid.New()
	user := &models.AuthUser{
		ID:       userID,
		Email:    "test@example.com",
		Username: "testuser",
		RoleName: "Admin",
		Permissions: []models.Permission{
			{Resource: "users", Action: "read"},
		},
	}

	ctx := context.WithValue(context.Background(), UserContextKey, user)

	retrievedUser, ok := GetUserFromContext(ctx)
	assert.True(t, ok)
	assert.Equal(t, user.ID, retrievedUser.ID)
	assert.Equal(t, user.Email, retrievedUser.Email)
}

func TestGetUserFromContext_NotFound(t *testing.T) {
	ctx := context.Background()

	user, ok := GetUserFromContext(ctx)
	assert.False(t, ok)
	assert.Nil(t, user)
}

func TestGetUserIDFromContext(t *testing.T) {
	userID := uuid.New()
	ctx := context.WithValue(context.Background(), UserIDContextKey, userID)

	retrievedID, ok := GetUserIDFromContext(ctx)
	assert.True(t, ok)
	assert.Equal(t, userID, retrievedID)
}

func TestHasPermissionInContext(t *testing.T) {
	user := &models.AuthUser{
		ID:    uuid.New(),
		Email: "test@example.com",
		Permissions: []models.Permission{
			{Resource: "users", Action: "read"},
			{Resource: "clients", Action: "write"},
		},
	}

	ctx := context.WithValue(context.Background(), UserContextKey, user)

	assert.True(t, HasPermissionInContext(ctx, "users", "read"))
	assert.True(t, HasPermissionInContext(ctx, "clients", "write"))
	assert.False(t, HasPermissionInContext(ctx, "users", "write"))
	assert.False(t, HasPermissionInContext(ctx, "shops", "read"))
}

func TestHasPermissionInContext_NoUser(t *testing.T) {
	ctx := context.Background()
	assert.False(t, HasPermissionInContext(ctx, "users", "read"))
}
