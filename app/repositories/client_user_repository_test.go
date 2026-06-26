package repositories

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yourorg/shoppilot/internal/models"
)

// Helper function to create a test client for client user tests
func createTestClient(t *testing.T, repo ClientUserRepository, pool interface{}) uuid.UUID {
	// Create a client using raw SQL since we're testing the client user repository
	ctx := context.Background()

	// Use type assertion to get the pool
	var clientID uuid.UUID
	query := `
		INSERT INTO clients (name, slug, contact_email, is_active)
		VALUES ($1, $2, $3, $4)
		RETURNING id
	`

	// Get pool from the repository
	r := repo.(*clientUserRepository)
	err := r.pool.QueryRow(
		ctx,
		query,
		"Test Client",
		fmt.Sprintf("test-client-%s", uuid.New().String()[:8]),
		"test@example.com",
		true,
	).Scan(&clientID)

	require.NoError(t, err)
	return clientID
}

// Helper function to create a test role
func createTestRole(t *testing.T, repo ClientUserRepository, roleName string) int {
	ctx := context.Background()

	var roleID int
	query := `
		INSERT INTO client_roles (name, description, is_system_role)
		VALUES ($1, $2, $3)
		ON CONFLICT (name) DO UPDATE SET name = EXCLUDED.name
		RETURNING id
	`

	r := repo.(*clientUserRepository)
	err := r.pool.QueryRow(ctx, query, roleName, "Test role", false).Scan(&roleID)
	require.NoError(t, err)
	return roleID
}

// Helper function to create a test permission
func createTestPermission(t *testing.T, repo ClientUserRepository, permName, resource, action string) int {
	ctx := context.Background()

	var permID int
	query := `
		INSERT INTO client_permissions (name, description, resource, action)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (name) DO UPDATE SET name = EXCLUDED.name
		RETURNING id
	`

	r := repo.(*clientUserRepository)
	err := r.pool.QueryRow(ctx, query, permName, "Test permission", resource, action).Scan(&permID)
	require.NoError(t, err)
	return permID
}

// Helper function to assign permission to role
func assignPermissionToRole(t *testing.T, repo ClientUserRepository, roleID, permID int) {
	ctx := context.Background()

	query := `
		INSERT INTO client_role_permissions (role_id, permission_id)
		VALUES ($1, $2)
		ON CONFLICT (role_id, permission_id) DO NOTHING
	`

	r := repo.(*clientUserRepository)
	_, err := r.pool.Exec(ctx, query, roleID, permID)
	require.NoError(t, err)
}

func TestClientUserRepository_Create(t *testing.T) {
	pool := SetupTestDB(t)
	defer CleanupTestDB(t, pool)

	// Clean up tables
	TruncateTable(t, pool, "client_users")
	TruncateTable(t, pool, "clients")

	repo := NewClientUserRepository(pool)
	clientID := createTestClient(t, repo, pool)

	t.Run("successfully creates a client user", func(t *testing.T) {
		user := &models.ClientUser{
			ClientID:     clientID,
			Email:        "user@example.com",
			Username:     "testuser",
			Password:     "hashedpassword123",
			FirstName:    "Test",
			LastName:     "User",
			Phone:        "1234567890",
			UserStatusID: 1,
		}

		err := repo.Create(context.Background(), user)
		require.NoError(t, err)
		assert.NotEqual(t, uuid.Nil, user.ID)
		assert.NotZero(t, user.CreatedAt)
		assert.NotZero(t, user.UpdatedAt)
	})

	t.Run("fails with duplicate email in same client", func(t *testing.T) {
		user1 := &models.ClientUser{
			ClientID:     clientID,
			Email:        "duplicate@example.com",
			Username:     "user1",
			Password:     "password",
			UserStatusID: 1,
		}
		err := repo.Create(context.Background(), user1)
		require.NoError(t, err)

		user2 := &models.ClientUser{
			ClientID:     clientID,
			Email:        "duplicate@example.com", // Same email
			Username:     "user2",
			Password:     "password",
			UserStatusID: 1,
		}
		err = repo.Create(context.Background(), user2)
		assert.Error(t, err)
	})

	t.Run("fails with duplicate username in same client", func(t *testing.T) {
		user1 := &models.ClientUser{
			ClientID:     clientID,
			Email:        "email1@example.com",
			Username:     "uniqueuser",
			Password:     "password",
			UserStatusID: 1,
		}
		err := repo.Create(context.Background(), user1)
		require.NoError(t, err)

		user2 := &models.ClientUser{
			ClientID:     clientID,
			Email:        "email2@example.com",
			Username:     "uniqueuser", // Same username
			Password:     "password",
			UserStatusID: 1,
		}
		err = repo.Create(context.Background(), user2)
		assert.Error(t, err)
	})

	t.Run("allows same email/username in different clients", func(t *testing.T) {
		// Create second client
		clientID2 := createTestClient(t, repo, pool)

		user1 := &models.ClientUser{
			ClientID:     clientID,
			Email:        "shared@example.com",
			Username:     "shareduser",
			Password:     "password",
			UserStatusID: 1,
		}
		err := repo.Create(context.Background(), user1)
		require.NoError(t, err)

		user2 := &models.ClientUser{
			ClientID:     clientID2,
			Email:        "shared@example.com", // Same email, different client
			Username:     "shareduser",         // Same username, different client
			Password:     "password",
			UserStatusID: 1,
		}
		err = repo.Create(context.Background(), user2)
		require.NoError(t, err)
	})
}

func TestClientUserRepository_GetByID(t *testing.T) {
	pool := SetupTestDB(t)
	defer CleanupTestDB(t, pool)

	TruncateTable(t, pool, "client_users")
	TruncateTable(t, pool, "clients")

	repo := NewClientUserRepository(pool)
	clientID := createTestClient(t, repo, pool)

	// Create a user
	user := &models.ClientUser{
		ClientID:     clientID,
		Email:        "getbyid@example.com",
		Username:     "getbyiduser",
		Password:     "password",
		FirstName:    "Get",
		LastName:     "ByID",
		UserStatusID: 1,
	}
	err := repo.Create(context.Background(), user)
	require.NoError(t, err)

	t.Run("successfully retrieves user by ID", func(t *testing.T) {
		found, err := repo.GetByID(context.Background(), user.ID)
		require.NoError(t, err)
		assert.Equal(t, user.ID, found.ID)
		assert.Equal(t, user.Email, found.Email)
		assert.Equal(t, user.Username, found.Username)
		assert.Equal(t, user.ClientID, found.ClientID)
	})

	t.Run("returns error for non-existent ID", func(t *testing.T) {
		_, err := repo.GetByID(context.Background(), uuid.New())
		assert.Error(t, err)
	})
}

func TestClientUserRepository_GetByEmail(t *testing.T) {
	pool := SetupTestDB(t)
	defer CleanupTestDB(t, pool)

	TruncateTable(t, pool, "client_users")
	TruncateTable(t, pool, "clients")

	repo := NewClientUserRepository(pool)
	clientID := createTestClient(t, repo, pool)

	user := &models.ClientUser{
		ClientID:     clientID,
		Email:        "unique-email@example.com",
		Username:     "emailuser",
		Password:     "password",
		UserStatusID: 1,
	}
	err := repo.Create(context.Background(), user)
	require.NoError(t, err)

	t.Run("successfully retrieves user by email", func(t *testing.T) {
		found, err := repo.GetByEmail(context.Background(), clientID, "unique-email@example.com")
		require.NoError(t, err)
		assert.Equal(t, user.ID, found.ID)
		assert.Equal(t, "unique-email@example.com", found.Email)
	})

	t.Run("returns error for non-existent email", func(t *testing.T) {
		_, err := repo.GetByEmail(context.Background(), clientID, "nonexistent@example.com")
		assert.Error(t, err)
	})

	t.Run("returns error for wrong client", func(t *testing.T) {
		clientID2 := createTestClient(t, repo, pool)
		_, err := repo.GetByEmail(context.Background(), clientID2, "unique-email@example.com")
		assert.Error(t, err)
	})
}

func TestClientUserRepository_GetByUsername(t *testing.T) {
	pool := SetupTestDB(t)
	defer CleanupTestDB(t, pool)

	TruncateTable(t, pool, "client_users")
	TruncateTable(t, pool, "clients")

	repo := NewClientUserRepository(pool)
	clientID := createTestClient(t, repo, pool)

	user := &models.ClientUser{
		ClientID:     clientID,
		Email:        "usernametest@example.com",
		Username:     "uniqueusername",
		Password:     "password",
		UserStatusID: 1,
	}
	err := repo.Create(context.Background(), user)
	require.NoError(t, err)

	t.Run("successfully retrieves user by username", func(t *testing.T) {
		found, err := repo.GetByUsername(context.Background(), clientID, "uniqueusername")
		require.NoError(t, err)
		assert.Equal(t, user.ID, found.ID)
		assert.Equal(t, "uniqueusername", found.Username)
	})

	t.Run("returns error for non-existent username", func(t *testing.T) {
		_, err := repo.GetByUsername(context.Background(), clientID, "nonexistent")
		assert.Error(t, err)
	})

	t.Run("returns error for wrong client", func(t *testing.T) {
		clientID2 := createTestClient(t, repo, pool)
		_, err := repo.GetByUsername(context.Background(), clientID2, "uniqueusername")
		assert.Error(t, err)
	})
}

func TestClientUserRepository_Update(t *testing.T) {
	pool := SetupTestDB(t)
	defer CleanupTestDB(t, pool)

	TruncateTable(t, pool, "client_users")
	TruncateTable(t, pool, "clients")

	repo := NewClientUserRepository(pool)
	clientID := createTestClient(t, repo, pool)

	user := &models.ClientUser{
		ClientID:     clientID,
		Email:        "original@example.com",
		Username:     "original",
		Password:     "password",
		FirstName:    "Original",
		LastName:     "Name",
		UserStatusID: 1,
	}
	err := repo.Create(context.Background(), user)
	require.NoError(t, err)

	t.Run("successfully updates user", func(t *testing.T) {
		user.Email = "updated@example.com"
		user.FirstName = "Updated"
		user.LastName = "User"
		user.Phone = "9876543210"

		err := repo.Update(context.Background(), user)
		require.NoError(t, err)

		// Verify update
		found, err := repo.GetByID(context.Background(), user.ID)
		require.NoError(t, err)
		assert.Equal(t, "updated@example.com", found.Email)
		assert.Equal(t, "Updated", found.FirstName)
		assert.Equal(t, "User", found.LastName)
		assert.Equal(t, "9876543210", found.Phone)
	})

	t.Run("returns error for non-existent user", func(t *testing.T) {
		nonExistent := &models.ClientUser{
			ID:           uuid.New(),
			ClientID:     clientID,
			Email:        "none@example.com",
			Username:     "none",
			Password:     "password",
			UserStatusID: 1,
		}
		err := repo.Update(context.Background(), nonExistent)
		assert.Error(t, err)
	})
}

func TestClientUserRepository_Delete(t *testing.T) {
	pool := SetupTestDB(t)
	defer CleanupTestDB(t, pool)

	TruncateTable(t, pool, "client_users")
	TruncateTable(t, pool, "clients")

	repo := NewClientUserRepository(pool)
	clientID := createTestClient(t, repo, pool)

	user := &models.ClientUser{
		ClientID:     clientID,
		Email:        "todelete@example.com",
		Username:     "todelete",
		Password:     "password",
		UserStatusID: 1,
	}
	err := repo.Create(context.Background(), user)
	require.NoError(t, err)

	t.Run("successfully deletes user", func(t *testing.T) {
		err := repo.Delete(context.Background(), user.ID)
		require.NoError(t, err)

		// Verify deletion
		_, err = repo.GetByID(context.Background(), user.ID)
		assert.Error(t, err)
	})

	t.Run("returns error for non-existent user", func(t *testing.T) {
		err := repo.Delete(context.Background(), uuid.New())
		assert.Error(t, err)
	})
}

func TestClientUserRepository_ListByClient(t *testing.T) {
	pool := SetupTestDB(t)
	defer CleanupTestDB(t, pool)

	TruncateTable(t, pool, "client_users")
	TruncateTable(t, pool, "clients")

	repo := NewClientUserRepository(pool)
	clientID := createTestClient(t, repo, pool)
	clientID2 := createTestClient(t, repo, pool)

	// Create users for first client
	for i := 1; i <= 5; i++ {
		user := &models.ClientUser{
			ClientID:     clientID,
			Email:        fmt.Sprintf("user%d@example.com", i),
			Username:     fmt.Sprintf("user%d", i),
			Password:     "password",
			UserStatusID: 1,
		}
		err := repo.Create(context.Background(), user)
		require.NoError(t, err)
	}

	// Create users for second client
	for i := 1; i <= 3; i++ {
		user := &models.ClientUser{
			ClientID:     clientID2,
			Email:        fmt.Sprintf("client2user%d@example.com", i),
			Username:     fmt.Sprintf("client2user%d", i),
			Password:     "password",
			UserStatusID: 1,
		}
		err := repo.Create(context.Background(), user)
		require.NoError(t, err)
	}

	t.Run("successfully lists users for specific client", func(t *testing.T) {
		users, err := repo.ListByClient(context.Background(), clientID, 10, 0)
		require.NoError(t, err)
		assert.Equal(t, 5, len(users))

		// All users should belong to clientID
		for _, u := range users {
			assert.Equal(t, clientID, u.ClientID)
		}
	})

	t.Run("pagination works correctly", func(t *testing.T) {
		// Get first page
		page1, err := repo.ListByClient(context.Background(), clientID, 2, 0)
		require.NoError(t, err)
		assert.Equal(t, 2, len(page1))

		// Get second page
		page2, err := repo.ListByClient(context.Background(), clientID, 2, 2)
		require.NoError(t, err)
		assert.Equal(t, 2, len(page2))

		// Ensure different users
		assert.NotEqual(t, page1[0].ID, page2[0].ID)
	})

	t.Run("returns empty list for client with no users", func(t *testing.T) {
		emptyClientID := createTestClient(t, repo, pool)
		users, err := repo.ListByClient(context.Background(), emptyClientID, 10, 0)
		require.NoError(t, err)
		assert.Equal(t, 0, len(users))
	})
}

func TestClientUserRepository_AssignRole(t *testing.T) {
	pool := SetupTestDB(t)
	defer CleanupTestDB(t, pool)

	TruncateTable(t, pool, "client_user_roles")
	TruncateTable(t, pool, "client_users")
	TruncateTable(t, pool, "clients")

	repo := NewClientUserRepository(pool)
	clientID := createTestClient(t, repo, pool)

	user := &models.ClientUser{
		ClientID:     clientID,
		Email:        "roleuser@example.com",
		Username:     "roleuser",
		Password:     "password",
		UserStatusID: 1,
	}
	err := repo.Create(context.Background(), user)
	require.NoError(t, err)

	roleID := createTestRole(t, repo, "test_role")

	t.Run("successfully assigns role to user", func(t *testing.T) {
		err := repo.AssignRole(context.Background(), user.ID, roleID)
		require.NoError(t, err)

		// Verify role assignment
		roles, err := repo.GetUserRoles(context.Background(), user.ID)
		require.NoError(t, err)
		assert.Equal(t, 1, len(roles))
		assert.Equal(t, roleID, roles[0].ID)
	})

	t.Run("handles duplicate role assignment gracefully", func(t *testing.T) {
		err := repo.AssignRole(context.Background(), user.ID, roleID)
		require.NoError(t, err) // Should not error due to ON CONFLICT DO NOTHING
	})
}

func TestClientUserRepository_RemoveRole(t *testing.T) {
	pool := SetupTestDB(t)
	defer CleanupTestDB(t, pool)

	TruncateTable(t, pool, "client_user_roles")
	TruncateTable(t, pool, "client_users")
	TruncateTable(t, pool, "clients")

	repo := NewClientUserRepository(pool)
	clientID := createTestClient(t, repo, pool)

	user := &models.ClientUser{
		ClientID:     clientID,
		Email:        "removerole@example.com",
		Username:     "removerole",
		Password:     "password",
		UserStatusID: 1,
	}
	err := repo.Create(context.Background(), user)
	require.NoError(t, err)

	roleID := createTestRole(t, repo, "role_to_remove")
	err = repo.AssignRole(context.Background(), user.ID, roleID)
	require.NoError(t, err)

	t.Run("successfully removes role from user", func(t *testing.T) {
		err := repo.RemoveRole(context.Background(), user.ID, roleID)
		require.NoError(t, err)

		// Verify role removal
		roles, err := repo.GetUserRoles(context.Background(), user.ID)
		require.NoError(t, err)
		assert.Equal(t, 0, len(roles))
	})

	t.Run("returns error for non-existent role assignment", func(t *testing.T) {
		err := repo.RemoveRole(context.Background(), user.ID, roleID)
		assert.Error(t, err)
	})
}

func TestClientUserRepository_GetUserRoles(t *testing.T) {
	pool := SetupTestDB(t)
	defer CleanupTestDB(t, pool)

	TruncateTable(t, pool, "client_user_roles")
	TruncateTable(t, pool, "client_users")
	TruncateTable(t, pool, "clients")

	repo := NewClientUserRepository(pool)
	clientID := createTestClient(t, repo, pool)

	user := &models.ClientUser{
		ClientID:     clientID,
		Email:        "multirole@example.com",
		Username:     "multirole",
		Password:     "password",
		UserStatusID: 1,
	}
	err := repo.Create(context.Background(), user)
	require.NoError(t, err)

	// Create and assign multiple roles
	role1ID := createTestRole(t, repo, "admin_role")
	role2ID := createTestRole(t, repo, "manager_role")
	role3ID := createTestRole(t, repo, "viewer_role")

	err = repo.AssignRole(context.Background(), user.ID, role1ID)
	require.NoError(t, err)
	err = repo.AssignRole(context.Background(), user.ID, role2ID)
	require.NoError(t, err)
	err = repo.AssignRole(context.Background(), user.ID, role3ID)
	require.NoError(t, err)

	t.Run("successfully retrieves all user roles", func(t *testing.T) {
		roles, err := repo.GetUserRoles(context.Background(), user.ID)
		require.NoError(t, err)
		assert.Equal(t, 3, len(roles))
	})

	t.Run("returns empty list for user with no roles", func(t *testing.T) {
		userNoRoles := &models.ClientUser{
			ClientID:     clientID,
			Email:        "noroles@example.com",
			Username:     "noroles",
			Password:     "password",
			UserStatusID: 1,
		}
		err := repo.Create(context.Background(), userNoRoles)
		require.NoError(t, err)

		roles, err := repo.GetUserRoles(context.Background(), userNoRoles.ID)
		require.NoError(t, err)
		assert.Equal(t, 0, len(roles))
	})
}

func TestClientUserRepository_GetUserPermissions(t *testing.T) {
	pool := SetupTestDB(t)
	defer CleanupTestDB(t, pool)

	TruncateTable(t, pool, "client_user_roles")
	TruncateTable(t, pool, "client_role_permissions")
	TruncateTable(t, pool, "client_users")
	TruncateTable(t, pool, "clients")

	repo := NewClientUserRepository(pool)
	clientID := createTestClient(t, repo, pool)

	user := &models.ClientUser{
		ClientID:     clientID,
		Email:        "permissions@example.com",
		Username:     "permissions",
		Password:     "password",
		UserStatusID: 1,
	}
	err := repo.Create(context.Background(), user)
	require.NoError(t, err)

	// Create roles and permissions
	role1ID := createTestRole(t, repo, "perm_role_1")
	role2ID := createTestRole(t, repo, "perm_role_2")

	perm1ID := createTestPermission(t, repo, "create_shop", "shop", "create")
	perm2ID := createTestPermission(t, repo, "read_shop", "shop", "read")
	perm3ID := createTestPermission(t, repo, "update_product", "product", "update")

	// Assign permissions to roles
	assignPermissionToRole(t, repo, role1ID, perm1ID)
	assignPermissionToRole(t, repo, role1ID, perm2ID)
	assignPermissionToRole(t, repo, role2ID, perm3ID)
	assignPermissionToRole(t, repo, role2ID, perm2ID) // perm2 is in both roles

	// Assign roles to user
	err = repo.AssignRole(context.Background(), user.ID, role1ID)
	require.NoError(t, err)
	err = repo.AssignRole(context.Background(), user.ID, role2ID)
	require.NoError(t, err)

	t.Run("successfully retrieves all user permissions", func(t *testing.T) {
		permissions, err := repo.GetUserPermissions(context.Background(), user.ID)
		require.NoError(t, err)

		// Should get 3 unique permissions (DISTINCT in query)
		assert.Equal(t, 3, len(permissions))

		// Verify permissions are present
		permNames := make(map[string]bool)
		for _, p := range permissions {
			permNames[p.Name] = true
		}
		assert.True(t, permNames["create_shop"])
		assert.True(t, permNames["read_shop"])
		assert.True(t, permNames["update_product"])
	})

	t.Run("returns empty list for user with no permissions", func(t *testing.T) {
		userNoPerms := &models.ClientUser{
			ClientID:     clientID,
			Email:        "noperms@example.com",
			Username:     "noperms",
			Password:     "password",
			UserStatusID: 1,
		}
		err := repo.Create(context.Background(), userNoPerms)
		require.NoError(t, err)

		permissions, err := repo.GetUserPermissions(context.Background(), userNoPerms.ID)
		require.NoError(t, err)
		assert.Equal(t, 0, len(permissions))
	})
}
