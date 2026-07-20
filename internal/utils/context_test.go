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

func TestGetClientIDFromContext(t *testing.T) {
	clientID := uuid.New()
	user := &models.AuthUser{
		ID:       uuid.New(),
		Email:    "client@example.com",
		Username: "clientuser",
		ClientID: &clientID,
	}

	ctx := context.WithValue(context.Background(), UserContextKey, user)

	retrievedClientID, err := GetClientIDFromContext(ctx)
	assert.NoError(t, err)
	assert.Equal(t, clientID, retrievedClientID)
}

func TestGetClientIDFromContext_NoUser(t *testing.T) {
	ctx := context.Background()

	clientID, err := GetClientIDFromContext(ctx)
	assert.Error(t, err)
	assert.Equal(t, "no authenticated user in context", err.Error())
	assert.Equal(t, uuid.Nil, clientID)
}

func TestGetClientIDFromContext_PlatformUser(t *testing.T) {
	user := &models.AuthUser{
		ID:       uuid.New(),
		Email:    "platform@example.com",
		Username: "platformuser",
		ClientID: nil, // Platform user has no client context
	}

	ctx := context.WithValue(context.Background(), UserContextKey, user)

	clientID, err := GetClientIDFromContext(ctx)
	assert.Error(t, err)
	assert.Equal(t, "user has no client context", err.Error())
	assert.Equal(t, uuid.Nil, clientID)
}

func TestMustGetClientIDFromContext(t *testing.T) {
	clientID := uuid.New()
	user := &models.AuthUser{
		ID:       uuid.New(),
		Email:    "client@example.com",
		ClientID: &clientID,
	}

	ctx := context.WithValue(context.Background(), UserContextKey, user)

	retrievedClientID := MustGetClientIDFromContext(ctx)
	assert.Equal(t, clientID, retrievedClientID)
}

func TestMustGetClientIDFromContext_Panic(t *testing.T) {
	// Platform user with no client context should panic
	user := &models.AuthUser{
		ID:       uuid.New(),
		Email:    "platform@example.com",
		ClientID: nil,
	}

	ctx := context.WithValue(context.Background(), UserContextKey, user)

	assert.Panics(t, func() {
		MustGetClientIDFromContext(ctx)
	})
}
