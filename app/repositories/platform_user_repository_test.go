package repositories

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yourorg/shoppilot/internal/models"
)

func TestPlatformUserRepository_Create(t *testing.T) {
	pool := SetupTestDB(t)
	defer CleanupTestDB(t, pool)

	repo := NewPlatformUserRepository(pool)

	// Clean up test data
	TruncateTable(t, pool, "platform_user_roles")
	TruncateTable(t, pool, "platform_users")

	user := &models.PlatformUser{
		Email:     "test@example.com",
		Username:  "testuser",
		Password:  "hashedpassword123",
		FirstName: "Test",
		LastName:  "User",
		Phone:     "+1234567890",
	}

	err := repo.Create(context.Background(), user)
	require.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, user.ID)
	assert.NotZero(t, user.CreatedAt)
	assert.NotZero(t, user.UpdatedAt)
	assert.Equal(t, 1, user.UserStatusID) // Default status

	// Verify in database
	retrieved, err := repo.GetByID(context.Background(), user.ID)
	require.NoError(t, err)
	assert.Equal(t, user.Email, retrieved.Email)
	assert.Equal(t, user.Username, retrieved.Username)
	assert.Equal(t, user.Password, retrieved.Password)
	assert.Equal(t, user.FirstName, retrieved.FirstName)
	assert.Equal(t, user.LastName, retrieved.LastName)
}

func TestPlatformUserRepository_GetByID(t *testing.T) {
	pool := SetupTestDB(t)
	defer CleanupTestDB(t, pool)

	repo := NewPlatformUserRepository(pool)

	// Clean up test data
	TruncateTable(t, pool, "platform_user_roles")
	TruncateTable(t, pool, "platform_users")

	user := &models.PlatformUser{
		Email:    "test@example.com",
		Username: "testuser",
		Password: "hashedpassword123",
	}

	err := repo.Create(context.Background(), user)
	require.NoError(t, err)

	// Test successful retrieval
	retrieved, err := repo.GetByID(context.Background(), user.ID)
	require.NoError(t, err)
	assert.Equal(t, user.ID, retrieved.ID)
	assert.Equal(t, user.Email, retrieved.Email)

	// Test non-existent user
	_, err = repo.GetByID(context.Background(), uuid.New())
	assert.Error(t, err)
}

func TestPlatformUserRepository_GetByEmail(t *testing.T) {
	pool := SetupTestDB(t)
	defer CleanupTestDB(t, pool)

	repo := NewPlatformUserRepository(pool)

	// Clean up test data
	TruncateTable(t, pool, "platform_user_roles")
	TruncateTable(t, pool, "platform_users")

	user := &models.PlatformUser{
		Email:    "test@example.com",
		Username: "testuser",
		Password: "hashedpassword123",
	}

	err := repo.Create(context.Background(), user)
	require.NoError(t, err)

	// Test successful retrieval
	retrieved, err := repo.GetByEmail(context.Background(), "test@example.com")
	require.NoError(t, err)
	assert.Equal(t, user.ID, retrieved.ID)
	assert.Equal(t, user.Email, retrieved.Email)

	// Test non-existent email
	_, err = repo.GetByEmail(context.Background(), "nonexistent@example.com")
	assert.Error(t, err)
}

func TestPlatformUserRepository_GetByUsername(t *testing.T) {
	pool := SetupTestDB(t)
	defer CleanupTestDB(t, pool)

	repo := NewPlatformUserRepository(pool)

	// Clean up test data
	TruncateTable(t, pool, "platform_user_roles")
	TruncateTable(t, pool, "platform_users")

	user := &models.PlatformUser{
		Email:    "test@example.com",
		Username: "testuser",
		Password: "hashedpassword123",
	}

	err := repo.Create(context.Background(), user)
	require.NoError(t, err)

	// Test successful retrieval
	retrieved, err := repo.GetByUsername(context.Background(), "testuser")
	require.NoError(t, err)
	assert.Equal(t, user.ID, retrieved.ID)
	assert.Equal(t, user.Username, retrieved.Username)

	// Test non-existent username
	_, err = repo.GetByUsername(context.Background(), "nonexistent")
	assert.Error(t, err)
}

func TestPlatformUserRepository_Update(t *testing.T) {
	pool := SetupTestDB(t)
	defer CleanupTestDB(t, pool)

	repo := NewPlatformUserRepository(pool)

	// Clean up test data
	TruncateTable(t, pool, "platform_user_roles")
	TruncateTable(t, pool, "platform_users")

	user := &models.PlatformUser{
		Email:    "test@example.com",
		Username: "testuser",
		Password: "hashedpassword123",
	}

	err := repo.Create(context.Background(), user)
	require.NoError(t, err)

	// Update user
	user.Email = "updated@example.com"
	user.FirstName = "Updated"
	user.LastName = "Name"
	user.Phone = "+9876543210"

	originalUpdatedAt := user.UpdatedAt
	time.Sleep(10 * time.Millisecond) // Ensure timestamp changes

	err = repo.Update(context.Background(), user)
	require.NoError(t, err)
	assert.True(t, user.UpdatedAt.After(originalUpdatedAt))

	// Verify changes
	retrieved, err := repo.GetByID(context.Background(), user.ID)
	require.NoError(t, err)
	assert.Equal(t, "updated@example.com", retrieved.Email)
	assert.Equal(t, "Updated", retrieved.FirstName)
	assert.Equal(t, "Name", retrieved.LastName)
	assert.Equal(t, "+9876543210", retrieved.Phone)

	// Test update non-existent user
	nonExistent := &models.PlatformUser{
		ID:       uuid.New(),
		Email:    "test@example.com",
		Username: "test",
		Password: "test",
	}
	err = repo.Update(context.Background(), nonExistent)
	assert.Error(t, err)
}

func TestPlatformUserRepository_Delete(t *testing.T) {
	pool := SetupTestDB(t)
	defer CleanupTestDB(t, pool)

	repo := NewPlatformUserRepository(pool)

	// Clean up test data
	TruncateTable(t, pool, "platform_user_roles")
	TruncateTable(t, pool, "platform_users")

	user := &models.PlatformUser{
		Email:    "test@example.com",
		Username: "testuser",
		Password: "hashedpassword123",
	}

	err := repo.Create(context.Background(), user)
	require.NoError(t, err)

	// Delete user
	err = repo.Delete(context.Background(), user.ID)
	require.NoError(t, err)

	// Verify deletion
	_, err = repo.GetByID(context.Background(), user.ID)
	assert.Error(t, err)

	// Test delete non-existent user
	err = repo.Delete(context.Background(), uuid.New())
	assert.Error(t, err)
}

func TestPlatformUserRepository_List(t *testing.T) {
	pool := SetupTestDB(t)
	defer CleanupTestDB(t, pool)

	repo := NewPlatformUserRepository(pool)

	// Clean up test data
	TruncateTable(t, pool, "platform_user_roles")
	TruncateTable(t, pool, "platform_users")

	// Create multiple users
	for i := 0; i < 5; i++ {
		user := &models.PlatformUser{
			Email:    uuid.New().String() + "@example.com",
			Username: "user" + uuid.New().String()[:8],
			Password: "password123",
		}
		err := repo.Create(context.Background(), user)
		require.NoError(t, err)
	}

	// Test pagination - first page
	users, err := repo.List(context.Background(), 3, 0)
	require.NoError(t, err)
	assert.Len(t, users, 3)

	// Test pagination - second page
	users, err = repo.List(context.Background(), 3, 3)
	require.NoError(t, err)
	assert.Len(t, users, 2)

	// Test empty result
	users, err = repo.List(context.Background(), 10, 100)
	require.NoError(t, err)
	assert.Len(t, users, 0)
}

func TestPlatformUserRepository_AssignRole(t *testing.T) {
	pool := SetupTestDB(t)
	defer CleanupTestDB(t, pool)

	repo := NewPlatformUserRepository(pool)

	// Clean up test data
	TruncateTable(t, pool, "platform_user_roles")
	TruncateTable(t, pool, "platform_users")

	user := &models.PlatformUser{
		Email:    "test@example.com",
		Username: "testuser",
		Password: "hashedpassword123",
	}

	err := repo.Create(context.Background(), user)
	require.NoError(t, err)

	// Assign role (assuming role with ID 1 exists from seed data)
	err = repo.AssignRole(context.Background(), user.ID, 1)
	require.NoError(t, err)

	// Verify role assignment
	roles, err := repo.GetUserRoles(context.Background(), user.ID)
	require.NoError(t, err)
	assert.Len(t, roles, 1)
	assert.Equal(t, 1, roles[0].ID)

	// Test duplicate assignment (should not error due to ON CONFLICT DO NOTHING)
	err = repo.AssignRole(context.Background(), user.ID, 1)
	require.NoError(t, err)

	// Verify still only one role
	roles, err = repo.GetUserRoles(context.Background(), user.ID)
	require.NoError(t, err)
	assert.Len(t, roles, 1)
}

func TestPlatformUserRepository_RemoveRole(t *testing.T) {
	pool := SetupTestDB(t)
	defer CleanupTestDB(t, pool)

	repo := NewPlatformUserRepository(pool)

	// Clean up test data
	TruncateTable(t, pool, "platform_user_roles")
	TruncateTable(t, pool, "platform_users")

	user := &models.PlatformUser{
		Email:    "test@example.com",
		Username: "testuser",
		Password: "hashedpassword123",
	}

	err := repo.Create(context.Background(), user)
	require.NoError(t, err)

	// Assign role
	err = repo.AssignRole(context.Background(), user.ID, 1)
	require.NoError(t, err)

	// Remove role
	err = repo.RemoveRole(context.Background(), user.ID, 1)
	require.NoError(t, err)

	// Verify removal
	roles, err := repo.GetUserRoles(context.Background(), user.ID)
	require.NoError(t, err)
	assert.Len(t, roles, 0)

	// Test remove non-existent role assignment
	err = repo.RemoveRole(context.Background(), user.ID, 1)
	assert.Error(t, err)
}

func TestPlatformUserRepository_GetUserRoles(t *testing.T) {
	pool := SetupTestDB(t)
	defer CleanupTestDB(t, pool)

	repo := NewPlatformUserRepository(pool)

	// Clean up test data
	TruncateTable(t, pool, "platform_user_roles")
	TruncateTable(t, pool, "platform_users")

	user := &models.PlatformUser{
		Email:    "test@example.com",
		Username: "testuser",
		Password: "hashedpassword123",
	}

	err := repo.Create(context.Background(), user)
	require.NoError(t, err)

	// Test with no roles
	roles, err := repo.GetUserRoles(context.Background(), user.ID)
	require.NoError(t, err)
	assert.Len(t, roles, 0)

	// Assign multiple roles (assuming roles with IDs 1 and 2 exist)
	err = repo.AssignRole(context.Background(), user.ID, 1)
	require.NoError(t, err)
	err = repo.AssignRole(context.Background(), user.ID, 2)
	require.NoError(t, err)

	// Get roles
	roles, err = repo.GetUserRoles(context.Background(), user.ID)
	require.NoError(t, err)
	assert.Len(t, roles, 2)

	// Verify role data
	for _, role := range roles {
		assert.NotEmpty(t, role.Name)
		assert.NotZero(t, role.ID)
	}
}

func TestPlatformUserRepository_GetUserPermissions(t *testing.T) {
	pool := SetupTestDB(t)
	defer CleanupTestDB(t, pool)

	repo := NewPlatformUserRepository(pool)

	// Clean up test data
	TruncateTable(t, pool, "platform_user_roles")
	TruncateTable(t, pool, "platform_users")

	user := &models.PlatformUser{
		Email:    "test@example.com",
		Username: "testuser",
		Password: "hashedpassword123",
	}

	err := repo.Create(context.Background(), user)
	require.NoError(t, err)

	// Test with no roles (no permissions)
	permissions, err := repo.GetUserPermissions(context.Background(), user.ID)
	require.NoError(t, err)
	assert.Len(t, permissions, 0)

	// Assign role (assuming role 1 has permissions)
	err = repo.AssignRole(context.Background(), user.ID, 1)
	require.NoError(t, err)

	// Get permissions
	permissions, err = repo.GetUserPermissions(context.Background(), user.ID)
	require.NoError(t, err)

	// Verify permissions (if role has permissions assigned in seed data)
	for _, perm := range permissions {
		assert.NotEmpty(t, perm.Name)
		assert.NotEmpty(t, perm.Resource)
		assert.NotEmpty(t, perm.Action)
		assert.NotZero(t, perm.ID)
	}
}
