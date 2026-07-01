package utils

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yourorg/shoppilot/internal/models"
)

func TestJWTManager_GenerateToken(t *testing.T) {
	manager := NewJWTManager("test-secret", 24*time.Hour)

	user := &models.AuthUser{
		ID:       uuid.New(),
		Email:    "test@example.com",
		Username: "testuser",
		RoleID:   1,
		RoleName: "Admin",
		Permissions: []models.Permission{
			{ID: 1, Name: "read_users", Resource: "users", Action: "read"},
		},
	}

	token, err := manager.GenerateToken(user)
	require.NoError(t, err)
	assert.NotEmpty(t, token)
}

func TestJWTManager_ValidateToken(t *testing.T) {
	manager := NewJWTManager("test-secret", 24*time.Hour)
	userID := uuid.New()

	user := &models.AuthUser{
		ID:       userID,
		Email:    "test@example.com",
		Username: "testuser",
		RoleID:   1,
		RoleName: "Admin",
		Permissions: []models.Permission{
			{ID: 1, Name: "read_users", Resource: "users", Action: "read"},
		},
	}

	token, err := manager.GenerateToken(user)
	require.NoError(t, err)

	claims, err := manager.ValidateToken(token)
	require.NoError(t, err)
	assert.Equal(t, userID, claims.UserID)
	assert.Equal(t, "test@example.com", claims.Email)
	assert.Equal(t, "Admin", claims.RoleName)
	assert.Len(t, claims.Permissions, 1)
}

func TestJWTManager_ValidateToken_Expired(t *testing.T) {
	manager := NewJWTManager("test-secret", -1*time.Hour) // Already expired

	user := &models.AuthUser{
		ID:       uuid.New(),
		Email:    "test@example.com",
		Username: "testuser",
		RoleID:   1,
		RoleName: "Admin",
	}

	token, err := manager.GenerateToken(user)
	require.NoError(t, err)

	_, err = manager.ValidateToken(token)
	assert.Error(t, err)
}

func TestJWTManager_ValidateToken_InvalidSecret(t *testing.T) {
	manager1 := NewJWTManager("secret1", 24*time.Hour)
	manager2 := NewJWTManager("secret2", 24*time.Hour)

	user := &models.AuthUser{
		ID:       uuid.New(),
		Email:    "test@example.com",
		Username: "testuser",
	}

	token, err := manager1.GenerateToken(user)
	require.NoError(t, err)

	_, err = manager2.ValidateToken(token)
	assert.Error(t, err)
}

func TestJWTClaims_HasPermission(t *testing.T) {
	claims := &JWTClaims{
		Permissions: []PermissionClaim{
			{Resource: "users", Action: "read"},
			{Resource: "clients", Action: "write"},
		},
	}

	assert.True(t, claims.HasPermission("users", "read"))
	assert.True(t, claims.HasPermission("clients", "write"))
	assert.False(t, claims.HasPermission("users", "write"))
	assert.False(t, claims.HasPermission("shops", "read"))
}
