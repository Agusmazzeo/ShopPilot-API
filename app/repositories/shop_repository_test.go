package repositories

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yourorg/shoppilot/app/models"
)

func TestShopRepository_Create(t *testing.T) {
	pool := SetupTestDB(t)
	defer CleanupTestDB(t, pool)

	TruncateTable(t, pool, "shops")

	repo := NewShopRepository(pool)

	t.Run("successfully creates a shop", func(t *testing.T) {
		shop := &models.Shop{
			ClientID: 1,
			UserID:   1,
			Name:     "Test Shop",
			Slug:     "test-shop",
			Domain:   stringPtr("testshop.example.com"),
			IsActive: true,
		}

		err := repo.Create(context.Background(), shop)
		require.NoError(t, err)
		assert.Greater(t, shop.ID, 0)
		assert.NotZero(t, shop.CreatedAt)
		assert.NotZero(t, shop.UpdatedAt)
	})

	t.Run("fails with duplicate slug within same client", func(t *testing.T) {
		shop1 := &models.Shop{
			ClientID: 1,
			UserID:   1,
			Name:     "Shop One",
			Slug:     "unique-shop",
			IsActive: true,
		}
		err := repo.Create(context.Background(), shop1)
		require.NoError(t, err)

		shop2 := &models.Shop{
			ClientID: 1,
			UserID:   1,
			Name:     "Shop Two",
			Slug:     "unique-shop", // Same slug, same client
			IsActive: true,
		}
		err = repo.Create(context.Background(), shop2)
		assert.Error(t, err)
	})

	t.Run("allows duplicate slug across different clients", func(t *testing.T) {
		shop1 := &models.Shop{
			ClientID: 1,
			UserID:   1,
			Name:     "Shop One",
			Slug:     "shared-slug",
			IsActive: true,
		}
		err := repo.Create(context.Background(), shop1)
		require.NoError(t, err)

		shop2 := &models.Shop{
			ClientID: 2, // Different client
			UserID:   2,
			Name:     "Shop Two",
			Slug:     "shared-slug", // Same slug
			IsActive: true,
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

	repo := NewShopRepository(pool)

	shop := &models.Shop{
		ClientID: 1,
		UserID:   1,
		Name:     "Test Shop",
		Slug:     "test-shop",
		IsActive: true,
	}
	err := repo.Create(context.Background(), shop)
	require.NoError(t, err)

	t.Run("successfully retrieves shop by ID", func(t *testing.T) {
		found, err := repo.GetByID(context.Background(), shop.ID)
		require.NoError(t, err)
		assert.Equal(t, shop.ID, found.ID)
		assert.Equal(t, shop.Name, found.Name)
		assert.Equal(t, shop.Slug, found.Slug)
	})

	t.Run("returns error for non-existent ID", func(t *testing.T) {
		_, err := repo.GetByID(context.Background(), 99999)
		assert.Error(t, err)
	})
}

func TestShopRepository_GetBySlug(t *testing.T) {
	pool := SetupTestDB(t)
	defer CleanupTestDB(t, pool)

	TruncateTable(t, pool, "shops")

	repo := NewShopRepository(pool)

	shop := &models.Shop{
		ClientID: 1,
		UserID:   1,
		Name:     "Test Shop",
		Slug:     "test-shop",
		IsActive: true,
	}
	err := repo.Create(context.Background(), shop)
	require.NoError(t, err)

	t.Run("successfully retrieves shop by slug and client", func(t *testing.T) {
		found, err := repo.GetBySlug(context.Background(), 1, "test-shop")
		require.NoError(t, err)
		assert.Equal(t, shop.ID, found.ID)
		assert.Equal(t, "test-shop", found.Slug)
	})

	t.Run("returns error for non-existent slug", func(t *testing.T) {
		_, err := repo.GetBySlug(context.Background(), 1, "nonexistent")
		assert.Error(t, err)
	})

	t.Run("returns error for wrong client", func(t *testing.T) {
		_, err := repo.GetBySlug(context.Background(), 2, "test-shop")
		assert.Error(t, err)
	})
}

func TestShopRepository_ListByClientID(t *testing.T) {
	pool := SetupTestDB(t)
	defer CleanupTestDB(t, pool)

	TruncateTable(t, pool, "shops")

	repo := NewShopRepository(pool)

	// Create shops for client 1
	for i := 1; i <= 3; i++ {
		shop := &models.Shop{
			ClientID: 1,
			UserID:   1,
			Name:     fmt.Sprintf("Shop %d", i),
			Slug:     fmt.Sprintf("shop-%d", i),
			IsActive: true,
		}
		err := repo.Create(context.Background(), shop)
		require.NoError(t, err)
	}

	// Create shops for client 2
	for i := 1; i <= 2; i++ {
		shop := &models.Shop{
			ClientID: 2,
			UserID:   2,
			Name:     fmt.Sprintf("Shop %d Client 2", i),
			Slug:     fmt.Sprintf("shop-%d-c2", i),
			IsActive: true,
		}
		err := repo.Create(context.Background(), shop)
		require.NoError(t, err)
	}

	t.Run("successfully lists shops for client 1", func(t *testing.T) {
		shops, err := repo.ListByClientID(context.Background(), 1)
		require.NoError(t, err)
		assert.Len(t, shops, 3)
	})

	t.Run("successfully lists shops for client 2", func(t *testing.T) {
		shops, err := repo.ListByClientID(context.Background(), 2)
		require.NoError(t, err)
		assert.Len(t, shops, 2)
	})

	t.Run("returns empty list for client with no shops", func(t *testing.T) {
		shops, err := repo.ListByClientID(context.Background(), 99)
		require.NoError(t, err)
		assert.Empty(t, shops)
	})
}

func TestShopRepository_ListByUserID(t *testing.T) {
	pool := SetupTestDB(t)
	defer CleanupTestDB(t, pool)

	TruncateTable(t, pool, "shops")

	repo := NewShopRepository(pool)

	// Create shops for user 1
	for i := 1; i <= 2; i++ {
		shop := &models.Shop{
			ClientID: 1,
			UserID:   1,
			Name:     fmt.Sprintf("User1 Shop %d", i),
			Slug:     fmt.Sprintf("user1-shop-%d", i),
			IsActive: true,
		}
		err := repo.Create(context.Background(), shop)
		require.NoError(t, err)
	}

	// Create shops for user 2
	shop := &models.Shop{
		ClientID: 1,
		UserID:   2,
		Name:     "User2 Shop",
		Slug:     "user2-shop",
		IsActive: true,
	}
	err := repo.Create(context.Background(), shop)
	require.NoError(t, err)

	t.Run("successfully lists shops for user 1", func(t *testing.T) {
		shops, err := repo.ListByUserID(context.Background(), 1)
		require.NoError(t, err)
		assert.Len(t, shops, 2)
	})

	t.Run("successfully lists shops for user 2", func(t *testing.T) {
		shops, err := repo.ListByUserID(context.Background(), 2)
		require.NoError(t, err)
		assert.Len(t, shops, 1)
	})

	t.Run("returns empty list for user with no shops", func(t *testing.T) {
		shops, err := repo.ListByUserID(context.Background(), 99)
		require.NoError(t, err)
		assert.Empty(t, shops)
	})
}

func TestShopRepository_Update(t *testing.T) {
	pool := SetupTestDB(t)
	defer CleanupTestDB(t, pool)

	TruncateTable(t, pool, "shops")

	repo := NewShopRepository(pool)

	shop := &models.Shop{
		ClientID: 1,
		UserID:   1,
		Name:     "Original Name",
		Slug:     "original-slug",
		IsActive: true,
	}
	err := repo.Create(context.Background(), shop)
	require.NoError(t, err)

	t.Run("successfully updates shop", func(t *testing.T) {
		shop.Name = "Updated Name"
		shop.Domain = stringPtr("updated.example.com")
		err := repo.Update(context.Background(), shop)
		require.NoError(t, err)

		// Verify update
		found, err := repo.GetByID(context.Background(), shop.ID)
		require.NoError(t, err)
		assert.Equal(t, "Updated Name", found.Name)
		assert.Equal(t, "updated.example.com", *found.Domain)
	})

	t.Run("returns error for non-existent shop", func(t *testing.T) {
		nonExistent := &models.Shop{
			ID:       99999,
			Name:     "Test",
			IsActive: true,
		}
		err := repo.Update(context.Background(), nonExistent)
		assert.Error(t, err)
	})
}

func TestShopRepository_Delete(t *testing.T) {
	pool := SetupTestDB(t)
	defer CleanupTestDB(t, pool)

	TruncateTable(t, pool, "shops")

	repo := NewShopRepository(pool)

	shop := &models.Shop{
		ClientID: 1,
		UserID:   1,
		Name:     "To Delete",
		Slug:     "to-delete",
		IsActive: true,
	}
	err := repo.Create(context.Background(), shop)
	require.NoError(t, err)

	t.Run("successfully deletes shop", func(t *testing.T) {
		err := repo.Delete(context.Background(), shop.ID)
		require.NoError(t, err)

		// Verify deletion
		_, err = repo.GetByID(context.Background(), shop.ID)
		assert.Error(t, err)
	})

	t.Run("returns error for non-existent shop", func(t *testing.T) {
		err := repo.Delete(context.Background(), 99999)
		assert.Error(t, err)
	})
}

// Helper function for creating string pointers
func stringPtr(s string) *string {
	return &s
}
