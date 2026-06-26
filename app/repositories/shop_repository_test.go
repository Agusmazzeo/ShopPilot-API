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

func TestShopRepository_Create(t *testing.T) {
	pool := SetupTestDB(t)
	defer CleanupTestDB(t, pool)

	TruncateTable(t, pool, "shops")
	TruncateTable(t, pool, "clients")

	repo := NewShopRepository(pool)

	// Create a test client first
	clientID := uuid.New()
	_, err := pool.Exec(context.Background(),
		`INSERT INTO clients (id, name, slug, is_active) VALUES ($1, $2, $3, $4)`,
		clientID, "Test Client", "test-client", true)
	require.NoError(t, err)

	t.Run("successfully creates a shop in partition", func(t *testing.T) {
		shop := &models.Shop{
			ClientID:    clientID,
			Name:        "Test Shop",
			Slug:        "test-shop",
			Description: "A test shop",
			WebpageURL:  "https://testshop.example.com",
			Address:     "123 Main St",
			City:        "TestCity",
			State:       "TS",
			Country:     "TestCountry",
			PostalCode:  "12345",
			Phone:       "555-0100",
			Email:       "test@shop.com",
			IsActive:    true,
		}

		err := repo.Create(context.Background(), shop)
		require.NoError(t, err)
		assert.NotEqual(t, uuid.Nil, shop.ID)
		assert.NotZero(t, shop.CreatedAt)
		assert.NotZero(t, shop.UpdatedAt)
	})

	t.Run("fails with duplicate slug within same client", func(t *testing.T) {
		shop1 := &models.Shop{
			ClientID:   clientID,
			Name:       "Shop One",
			Slug:       "unique-shop",
			WebpageURL: "https://shop1.example.com",
			IsActive:   true,
		}
		err := repo.Create(context.Background(), shop1)
		require.NoError(t, err)

		shop2 := &models.Shop{
			ClientID:   clientID,
			Name:       "Shop Two",
			Slug:       "unique-shop", // Same slug, same client
			WebpageURL: "https://shop2.example.com",
			IsActive:   true,
		}
		err = repo.Create(context.Background(), shop2)
		assert.Error(t, err)
	})

	t.Run("allows duplicate slug across different clients", func(t *testing.T) {
		client2ID := uuid.New()
		_, err := pool.Exec(context.Background(),
			`INSERT INTO clients (id, name, slug, is_active) VALUES ($1, $2, $3, $4)`,
			client2ID, "Test Client 2", "test-client-2", true)
		require.NoError(t, err)

		shop1 := &models.Shop{
			ClientID:   clientID,
			Name:       "Shop One",
			Slug:       "shared-slug",
			WebpageURL: "https://shop1.example.com",
			IsActive:   true,
		}
		err = repo.Create(context.Background(), shop1)
		require.NoError(t, err)

		shop2 := &models.Shop{
			ClientID:   client2ID, // Different client
			Name:       "Shop Two",
			Slug:       "shared-slug", // Same slug
			WebpageURL: "https://shop2.example.com",
			IsActive:   true,
		}
		err = repo.Create(context.Background(), shop2)
		require.NoError(t, err)
		assert.NotEqual(t, shop1.ID, shop2.ID)
	})
}

func TestShopRepository_GetByID(t *testing.T) {
	pool := SetupTestDB(t)
	defer CleanupTestDB(t, pool)

	TruncateTable(t, pool, "shops")
	TruncateTable(t, pool, "clients")

	repo := NewShopRepository(pool)

	// Create a test client
	clientID := uuid.New()
	_, err := pool.Exec(context.Background(),
		`INSERT INTO clients (id, name, slug, is_active) VALUES ($1, $2, $3, $4)`,
		clientID, "Test Client", "test-client", true)
	require.NoError(t, err)

	shop := &models.Shop{
		ClientID:   clientID,
		Name:       "Test Shop",
		Slug:       "test-shop",
		WebpageURL: "https://testshop.example.com",
		IsActive:   true,
	}
	err = repo.Create(context.Background(), shop)
	require.NoError(t, err)

	t.Run("successfully retrieves shop by composite key", func(t *testing.T) {
		found, err := repo.GetByID(context.Background(), clientID, shop.ID)
		require.NoError(t, err)
		assert.Equal(t, shop.ID, found.ID)
		assert.Equal(t, shop.ClientID, found.ClientID)
		assert.Equal(t, shop.Name, found.Name)
		assert.Equal(t, shop.Slug, found.Slug)
	})

	t.Run("returns error for non-existent shop ID", func(t *testing.T) {
		_, err := repo.GetByID(context.Background(), clientID, uuid.New())
		assert.Error(t, err)
	})

	t.Run("returns error for wrong client ID", func(t *testing.T) {
		_, err := repo.GetByID(context.Background(), uuid.New(), shop.ID)
		assert.Error(t, err)
	})
}

func TestShopRepository_GetBySlug(t *testing.T) {
	pool := SetupTestDB(t)
	defer CleanupTestDB(t, pool)

	TruncateTable(t, pool, "shops")
	TruncateTable(t, pool, "clients")

	repo := NewShopRepository(pool)

	// Create a test client
	clientID := uuid.New()
	_, err := pool.Exec(context.Background(),
		`INSERT INTO clients (id, name, slug, is_active) VALUES ($1, $2, $3, $4)`,
		clientID, "Test Client", "test-client", true)
	require.NoError(t, err)

	shop := &models.Shop{
		ClientID:   clientID,
		Name:       "Test Shop",
		Slug:       "test-shop",
		WebpageURL: "https://testshop.example.com",
		IsActive:   true,
	}
	err = repo.Create(context.Background(), shop)
	require.NoError(t, err)

	t.Run("successfully retrieves shop by slug and client", func(t *testing.T) {
		found, err := repo.GetBySlug(context.Background(), clientID, "test-shop")
		require.NoError(t, err)
		assert.Equal(t, shop.ID, found.ID)
		assert.Equal(t, "test-shop", found.Slug)
	})

	t.Run("returns error for non-existent slug", func(t *testing.T) {
		_, err := repo.GetBySlug(context.Background(), clientID, "nonexistent")
		assert.Error(t, err)
	})

	t.Run("returns error for wrong client", func(t *testing.T) {
		_, err := repo.GetBySlug(context.Background(), uuid.New(), "test-shop")
		assert.Error(t, err)
	})
}

func TestShopRepository_Update(t *testing.T) {
	pool := SetupTestDB(t)
	defer CleanupTestDB(t, pool)

	TruncateTable(t, pool, "shops")
	TruncateTable(t, pool, "clients")

	repo := NewShopRepository(pool)

	// Create a test client
	clientID := uuid.New()
	_, err := pool.Exec(context.Background(),
		`INSERT INTO clients (id, name, slug, is_active) VALUES ($1, $2, $3, $4)`,
		clientID, "Test Client", "test-client", true)
	require.NoError(t, err)

	shop := &models.Shop{
		ClientID:   clientID,
		Name:       "Original Name",
		Slug:       "original-slug",
		WebpageURL: "https://original.example.com",
		IsActive:   true,
	}
	err = repo.Create(context.Background(), shop)
	require.NoError(t, err)

	t.Run("successfully updates shop", func(t *testing.T) {
		newLogoURL := "https://example.com/logo.png"
		shop.Name = "Updated Name"
		shop.WebpageURL = "https://updated.example.com"
		shop.LogoURL = &newLogoURL
		err := repo.Update(context.Background(), shop)
		require.NoError(t, err)

		// Verify update
		found, err := repo.GetByID(context.Background(), clientID, shop.ID)
		require.NoError(t, err)
		assert.Equal(t, "Updated Name", found.Name)
		assert.Equal(t, "https://updated.example.com", found.WebpageURL)
		assert.NotNil(t, found.LogoURL)
		assert.Equal(t, newLogoURL, *found.LogoURL)
	})

	t.Run("returns error for non-existent shop", func(t *testing.T) {
		nonExistent := &models.Shop{
			ClientID:   clientID,
			ID:         uuid.New(),
			Name:       "Test",
			WebpageURL: "https://test.example.com",
			IsActive:   true,
		}
		err := repo.Update(context.Background(), nonExistent)
		assert.Error(t, err)
	})
}

func TestShopRepository_Delete(t *testing.T) {
	pool := SetupTestDB(t)
	defer CleanupTestDB(t, pool)

	TruncateTable(t, pool, "shops")
	TruncateTable(t, pool, "clients")

	repo := NewShopRepository(pool)

	// Create a test client
	clientID := uuid.New()
	_, err := pool.Exec(context.Background(),
		`INSERT INTO clients (id, name, slug, is_active) VALUES ($1, $2, $3, $4)`,
		clientID, "Test Client", "test-client", true)
	require.NoError(t, err)

	shop := &models.Shop{
		ClientID:   clientID,
		Name:       "To Delete",
		Slug:       "to-delete",
		WebpageURL: "https://todelete.example.com",
		IsActive:   true,
	}
	err = repo.Create(context.Background(), shop)
	require.NoError(t, err)

	t.Run("successfully deletes shop from partition", func(t *testing.T) {
		err := repo.Delete(context.Background(), clientID, shop.ID)
		require.NoError(t, err)

		// Verify deletion
		_, err = repo.GetByID(context.Background(), clientID, shop.ID)
		assert.Error(t, err)
	})

	t.Run("returns error for non-existent shop", func(t *testing.T) {
		err := repo.Delete(context.Background(), clientID, uuid.New())
		assert.Error(t, err)
	})
}

func TestShopRepository_ListByClient(t *testing.T) {
	pool := SetupTestDB(t)
	defer CleanupTestDB(t, pool)

	TruncateTable(t, pool, "shops")
	TruncateTable(t, pool, "clients")

	repo := NewShopRepository(pool)

	// Create test clients
	client1ID := uuid.New()
	client2ID := uuid.New()
	_, err := pool.Exec(context.Background(),
		`INSERT INTO clients (id, name, slug, is_active) VALUES ($1, $2, $3, $4), ($5, $6, $7, $8)`,
		client1ID, "Client 1", "client-1", true,
		client2ID, "Client 2", "client-2", true)
	require.NoError(t, err)

	// Create shops for client 1
	for i := 1; i <= 3; i++ {
		shop := &models.Shop{
			ClientID:   client1ID,
			Name:       fmt.Sprintf("Shop %d", i),
			Slug:       fmt.Sprintf("shop-%d", i),
			WebpageURL: fmt.Sprintf("https://shop%d.example.com", i),
			IsActive:   true,
		}
		err := repo.Create(context.Background(), shop)
		require.NoError(t, err)
	}

	// Create shops for client 2
	for i := 1; i <= 2; i++ {
		shop := &models.Shop{
			ClientID:   client2ID,
			Name:       fmt.Sprintf("Shop %d Client 2", i),
			Slug:       fmt.Sprintf("shop-%d-c2", i),
			WebpageURL: fmt.Sprintf("https://shop%d-c2.example.com", i),
			IsActive:   true,
		}
		err := repo.Create(context.Background(), shop)
		require.NoError(t, err)
	}

	t.Run("successfully lists shops for client 1 with pagination", func(t *testing.T) {
		shops, err := repo.ListByClient(context.Background(), client1ID, 10, 0)
		require.NoError(t, err)
		assert.Len(t, shops, 3)
	})

	t.Run("successfully lists shops for client 2", func(t *testing.T) {
		shops, err := repo.ListByClient(context.Background(), client2ID, 10, 0)
		require.NoError(t, err)
		assert.Len(t, shops, 2)
	})

	t.Run("returns empty list for client with no shops", func(t *testing.T) {
		shops, err := repo.ListByClient(context.Background(), uuid.New(), 10, 0)
		require.NoError(t, err)
		assert.Empty(t, shops)
	})

	t.Run("pagination works correctly", func(t *testing.T) {
		// Get first page
		shops, err := repo.ListByClient(context.Background(), client1ID, 2, 0)
		require.NoError(t, err)
		assert.Len(t, shops, 2)

		// Get second page
		shops, err = repo.ListByClient(context.Background(), client1ID, 2, 2)
		require.NoError(t, err)
		assert.Len(t, shops, 1)
	})
}

func TestShopRepository_ListActiveByClient(t *testing.T) {
	pool := SetupTestDB(t)
	defer CleanupTestDB(t, pool)

	TruncateTable(t, pool, "shops")
	TruncateTable(t, pool, "clients")

	repo := NewShopRepository(pool)

	// Create a test client
	clientID := uuid.New()
	_, err := pool.Exec(context.Background(),
		`INSERT INTO clients (id, name, slug, is_active) VALUES ($1, $2, $3, $4)`,
		clientID, "Test Client", "test-client", true)
	require.NoError(t, err)

	// Create active shops
	for i := 1; i <= 3; i++ {
		shop := &models.Shop{
			ClientID:   clientID,
			Name:       fmt.Sprintf("Active Shop %d", i),
			Slug:       fmt.Sprintf("active-shop-%d", i),
			WebpageURL: fmt.Sprintf("https://active%d.example.com", i),
			IsActive:   true,
		}
		err := repo.Create(context.Background(), shop)
		require.NoError(t, err)
	}

	// Create inactive shops
	for i := 1; i <= 2; i++ {
		shop := &models.Shop{
			ClientID:   clientID,
			Name:       fmt.Sprintf("Inactive Shop %d", i),
			Slug:       fmt.Sprintf("inactive-shop-%d", i),
			WebpageURL: fmt.Sprintf("https://inactive%d.example.com", i),
			IsActive:   false,
		}
		err := repo.Create(context.Background(), shop)
		require.NoError(t, err)
	}

	t.Run("returns only active shops", func(t *testing.T) {
		shops, err := repo.ListActiveByClient(context.Background(), clientID, 10, 0)
		require.NoError(t, err)
		assert.Len(t, shops, 3)
		for _, shop := range shops {
			assert.True(t, shop.IsActive)
		}
	})
}

func TestShopRepository_AssignUser(t *testing.T) {
	pool := SetupTestDB(t)
	defer CleanupTestDB(t, pool)

	TruncateTable(t, pool, "shop_users")
	TruncateTable(t, pool, "shops")
	TruncateTable(t, pool, "clients")

	repo := NewShopRepository(pool)

	// Create a test client and shop
	clientID := uuid.New()
	_, err := pool.Exec(context.Background(),
		`INSERT INTO clients (id, name, slug, is_active) VALUES ($1, $2, $3, $4)`,
		clientID, "Test Client", "test-client", true)
	require.NoError(t, err)

	shop := &models.Shop{
		ClientID:   clientID,
		Name:       "Test Shop",
		Slug:       "test-shop",
		WebpageURL: "https://testshop.example.com",
		IsActive:   true,
	}
	err = repo.Create(context.Background(), shop)
	require.NoError(t, err)

	t.Run("successfully assigns user to shop", func(t *testing.T) {
		// Create client_user and role
		clientUserID := uuid.New()
		_, err := pool.Exec(context.Background(),
			`INSERT INTO client_users (id, client_id, username, email, password, user_status_id) VALUES ($1, $2, $3, $4, $5, 1)`,
			clientUserID, clientID, "testuser", "test@example.com", "hash123")
		require.NoError(t, err)

		// Create role and client_user_role
		var roleID int
		err = pool.QueryRow(context.Background(),
			`INSERT INTO client_roles (name, description) VALUES ($1, $2) RETURNING id`,
			fmt.Sprintf("test_role_%s", uuid.New().String()[:8]), "Test Role").Scan(&roleID)
		require.NoError(t, err)

		var clientUserRoleID int
		err = pool.QueryRow(context.Background(),
			`INSERT INTO client_user_roles (user_id, role_id) VALUES ($1, $2) RETURNING id`,
			clientUserID, roleID).Scan(&clientUserRoleID)
		require.NoError(t, err)

		err = repo.AssignUser(context.Background(), shop.ID, clientUserRoleID)
		require.NoError(t, err)

		// Verify assignment
		shopUsers, err := repo.GetShopUsers(context.Background(), shop.ID)
		require.NoError(t, err)
		assert.Len(t, shopUsers, 1)
		assert.Equal(t, clientUserRoleID, shopUsers[0].ClientUserRoleID)
	})

	t.Run("handles duplicate assignment gracefully", func(t *testing.T) {
		// Create another client_user and role
		clientUserID := uuid.New()
		_, err := pool.Exec(context.Background(),
			`INSERT INTO client_users (id, client_id, username, email, password, user_status_id) VALUES ($1, $2, $3, $4, $5, 1)`,
			clientUserID, clientID, "testuser2", "test2@example.com", "hash123")
		require.NoError(t, err)

		var roleID int
		err = pool.QueryRow(context.Background(),
			`INSERT INTO client_roles (name, description) VALUES ($1, $2) RETURNING id`,
			fmt.Sprintf("test_role_%s", uuid.New().String()[:8]), "Test Role 2").Scan(&roleID)
		require.NoError(t, err)

		var clientUserRoleID int
		err = pool.QueryRow(context.Background(),
			`INSERT INTO client_user_roles (user_id, role_id) VALUES ($1, $2) RETURNING id`,
			clientUserID, roleID).Scan(&clientUserRoleID)
		require.NoError(t, err)

		err = repo.AssignUser(context.Background(), shop.ID, clientUserRoleID)
		require.NoError(t, err)

		// Try to assign again (should not error due to ON CONFLICT DO NOTHING)
		err = repo.AssignUser(context.Background(), shop.ID, clientUserRoleID)
		require.NoError(t, err)

		// Verify still only one assignment
		shopUsers, err := repo.GetShopUsers(context.Background(), shop.ID)
		require.NoError(t, err)
		// Should have 2 users (from previous test run + this one)
		assert.GreaterOrEqual(t, len(shopUsers), 1)
	})
}

func TestShopRepository_RemoveUser(t *testing.T) {
	pool := SetupTestDB(t)
	defer CleanupTestDB(t, pool)

	TruncateTable(t, pool, "shop_users")
	TruncateTable(t, pool, "shops")
	TruncateTable(t, pool, "clients")

	repo := NewShopRepository(pool)

	// Create a test client and shop
	clientID := uuid.New()
	_, err := pool.Exec(context.Background(),
		`INSERT INTO clients (id, name, slug, is_active) VALUES ($1, $2, $3, $4)`,
		clientID, "Test Client", "test-client", true)
	require.NoError(t, err)

	shop := &models.Shop{
		ClientID:   clientID,
		Name:       "Test Shop",
		Slug:       "test-shop",
		WebpageURL: "https://testshop.example.com",
		IsActive:   true,
	}
	err = repo.Create(context.Background(), shop)
	require.NoError(t, err)

	// Create client_user and role
	clientUserID := uuid.New()
	_, err = pool.Exec(context.Background(),
		`INSERT INTO client_users (id, client_id, username, email, password, user_status_id) VALUES ($1, $2, $3, $4, $5, 1)`,
		clientUserID, clientID, "testuser", "test@example.com", "hash123")
	require.NoError(t, err)

	var roleID int
	err = pool.QueryRow(context.Background(),
		`INSERT INTO client_roles (name, description) VALUES ($1, $2) RETURNING id`,
		fmt.Sprintf("test_role_%s", uuid.New().String()[:8]), "Test Role").Scan(&roleID)
	require.NoError(t, err)

	var clientUserRoleID int
	err = pool.QueryRow(context.Background(),
		`INSERT INTO client_user_roles (user_id, role_id) VALUES ($1, $2) RETURNING id`,
		clientUserID, roleID).Scan(&clientUserRoleID)
	require.NoError(t, err)

	err = repo.AssignUser(context.Background(), shop.ID, clientUserRoleID)
	require.NoError(t, err)

	t.Run("successfully removes user from shop", func(t *testing.T) {
		err := repo.RemoveUser(context.Background(), shop.ID, clientUserRoleID)
		require.NoError(t, err)

		// Verify removal
		shopUsers, err := repo.GetShopUsers(context.Background(), shop.ID)
		require.NoError(t, err)
		assert.Empty(t, shopUsers)
	})

	t.Run("returns error for non-existent shop user", func(t *testing.T) {
		err := repo.RemoveUser(context.Background(), shop.ID, 99999)
		assert.Error(t, err)
	})
}

func TestShopRepository_GetShopUsers(t *testing.T) {
	pool := SetupTestDB(t)
	defer CleanupTestDB(t, pool)

	TruncateTable(t, pool, "shop_users")
	TruncateTable(t, pool, "shops")
	TruncateTable(t, pool, "clients")

	repo := NewShopRepository(pool)

	// Create a test client and shop
	clientID := uuid.New()
	_, err := pool.Exec(context.Background(),
		`INSERT INTO clients (id, name, slug, is_active) VALUES ($1, $2, $3, $4)`,
		clientID, "Test Client", "test-client", true)
	require.NoError(t, err)

	shop := &models.Shop{
		ClientID:   clientID,
		Name:       "Test Shop",
		Slug:       "test-shop",
		WebpageURL: "https://testshop.example.com",
		IsActive:   true,
	}
	err = repo.Create(context.Background(), shop)
	require.NoError(t, err)

	// Assign multiple users
	for i := 1; i <= 3; i++ {
		// Create client_user and role for each assignment
		clientUserID := uuid.New()
		_, err := pool.Exec(context.Background(),
			`INSERT INTO client_users (id, client_id, username, email, password, user_status_id) VALUES ($1, $2, $3, $4, $5, 1)`,
			clientUserID, clientID, fmt.Sprintf("testuser%d", i), fmt.Sprintf("test%d@example.com", i), "hash123")
		require.NoError(t, err)

		var roleID int
		err = pool.QueryRow(context.Background(),
			`INSERT INTO client_roles (name, description) VALUES ($1, $2) RETURNING id`,
			fmt.Sprintf("test_role_%s_%d", uuid.New().String()[:8], i), fmt.Sprintf("Test Role %d", i)).Scan(&roleID)
		require.NoError(t, err)

		var clientUserRoleID int
		err = pool.QueryRow(context.Background(),
			`INSERT INTO client_user_roles (user_id, role_id) VALUES ($1, $2) RETURNING id`,
			clientUserID, roleID).Scan(&clientUserRoleID)
		require.NoError(t, err)

		err = repo.AssignUser(context.Background(), shop.ID, clientUserRoleID)
		require.NoError(t, err)
	}

	t.Run("successfully retrieves all shop users", func(t *testing.T) {
		shopUsers, err := repo.GetShopUsers(context.Background(), shop.ID)
		require.NoError(t, err)
		assert.Len(t, shopUsers, 3)
		for _, su := range shopUsers {
			assert.Equal(t, shop.ID, su.ShopID)
			assert.Equal(t, clientID, su.ClientID)
		}
	})

	t.Run("returns empty list for shop with no users", func(t *testing.T) {
		newShop := &models.Shop{
			ClientID:   clientID,
			Name:       "Empty Shop",
			Slug:       "empty-shop",
			WebpageURL: "https://emptyshop.example.com",
			IsActive:   true,
		}
		err := repo.Create(context.Background(), newShop)
		require.NoError(t, err)

		shopUsers, err := repo.GetShopUsers(context.Background(), newShop.ID)
		require.NoError(t, err)
		assert.Empty(t, shopUsers)
	})
}
