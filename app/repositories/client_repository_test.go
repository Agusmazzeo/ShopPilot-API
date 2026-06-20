package repositories

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yourorg/shoppilot/app/models"
)

func TestClientRepository_Create(t *testing.T) {
	pool := SetupTestDB(t)
	defer CleanupTestDB(t, pool)

	// Clean up before test
	TruncateTable(t, pool, "clients")

	repo := NewClientRepository(pool)

	t.Run("successfully creates a client", func(t *testing.T) {
		client := &models.Client{
			Name:             "Test Client",
			Slug:             "test-client",
			ContactEmail:     "test@example.com",
			SubscriptionTier: "free",
			IsActive:         true,
		}

		err := repo.Create(context.Background(), client)
		require.NoError(t, err)
		assert.Greater(t, client.ID, 0)
		assert.NotZero(t, client.CreatedAt)
		assert.NotZero(t, client.UpdatedAt)
	})

	t.Run("fails with duplicate slug", func(t *testing.T) {
		client1 := &models.Client{
			Name:             "Client One",
			Slug:             "unique-slug",
			ContactEmail:     "client1@example.com",
			SubscriptionTier: "free",
			IsActive:         true,
		}
		err := repo.Create(context.Background(), client1)
		require.NoError(t, err)

		client2 := &models.Client{
			Name:             "Client Two",
			Slug:             "unique-slug", // Same slug
			ContactEmail:     "client2@example.com",
			SubscriptionTier: "free",
			IsActive:         true,
		}
		err = repo.Create(context.Background(), client2)
		assert.Error(t, err)
	})
}

func TestClientRepository_GetByID(t *testing.T) {
	pool := SetupTestDB(t)
	defer CleanupTestDB(t, pool)

	TruncateTable(t, pool, "clients")

	repo := NewClientRepository(pool)

	// Create a client
	client := &models.Client{
		Name:             "Test Client",
		Slug:             "test-client",
		ContactEmail:     "test@example.com",
		SubscriptionTier: "free",
		IsActive:         true,
	}
	err := repo.Create(context.Background(), client)
	require.NoError(t, err)

	t.Run("successfully retrieves client by ID", func(t *testing.T) {
		found, err := repo.GetByID(context.Background(), client.ID)
		require.NoError(t, err)
		assert.Equal(t, client.ID, found.ID)
		assert.Equal(t, client.Slug, found.Slug)
		assert.Equal(t, client.ContactEmail, found.ContactEmail)
	})

	t.Run("returns error for non-existent ID", func(t *testing.T) {
		_, err := repo.GetByID(context.Background(), 99999)
		assert.Error(t, err)
	})
}

func TestClientRepository_GetBySlug(t *testing.T) {
	pool := SetupTestDB(t)
	defer CleanupTestDB(t, pool)

	TruncateTable(t, pool, "clients")

	repo := NewClientRepository(pool)

	client := &models.Client{
		Name:             "Test Client",
		Slug:             "test-slug",
		ContactEmail:     "test@example.com",
		SubscriptionTier: "free",
		IsActive:         true,
	}
	err := repo.Create(context.Background(), client)
	require.NoError(t, err)

	t.Run("successfully retrieves client by slug", func(t *testing.T) {
		found, err := repo.GetBySlug(context.Background(), "test-slug")
		require.NoError(t, err)
		assert.Equal(t, client.ID, found.ID)
		assert.Equal(t, "test-slug", found.Slug)
	})

	t.Run("returns error for non-existent slug", func(t *testing.T) {
		_, err := repo.GetBySlug(context.Background(), "non-existent")
		assert.Error(t, err)
	})
}

func TestClientRepository_List(t *testing.T) {
	pool := SetupTestDB(t)
	defer CleanupTestDB(t, pool)

	TruncateTable(t, pool, "clients")

	repo := NewClientRepository(pool)

	// Create multiple clients
	for i := 1; i <= 3; i++ {
		client := &models.Client{
			Name:             fmt.Sprintf("Client %d", i),
			Slug:             fmt.Sprintf("client-%d", i),
			ContactEmail:     fmt.Sprintf("client%d@example.com", i),
			SubscriptionTier: "free",
			IsActive:         true,
		}
		err := repo.Create(context.Background(), client)
		require.NoError(t, err)
	}

	t.Run("successfully lists all clients", func(t *testing.T) {
		clients, err := repo.List(context.Background())
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(clients), 3)
	})
}

func TestClientRepository_Update(t *testing.T) {
	pool := SetupTestDB(t)
	defer CleanupTestDB(t, pool)

	TruncateTable(t, pool, "clients")

	repo := NewClientRepository(pool)

	client := &models.Client{
		Name:             "Original Name",
		Slug:             "original-slug",
		ContactEmail:     "original@example.com",
		SubscriptionTier: "free",
		IsActive:         true,
	}
	err := repo.Create(context.Background(), client)
	require.NoError(t, err)

	t.Run("successfully updates client", func(t *testing.T) {
		client.Name = "Updated Name"
		client.SubscriptionTier = "pro"
		err := repo.Update(context.Background(), client)
		require.NoError(t, err)

		// Verify update
		found, err := repo.GetByID(context.Background(), client.ID)
		require.NoError(t, err)
		assert.Equal(t, "Updated Name", found.Name)
		assert.Equal(t, "pro", found.SubscriptionTier)
	})
}

func TestClientRepository_Delete(t *testing.T) {
	pool := SetupTestDB(t)
	defer CleanupTestDB(t, pool)

	TruncateTable(t, pool, "clients")

	repo := NewClientRepository(pool)

	client := &models.Client{
		Name:             "To Delete",
		Slug:             "to-delete",
		ContactEmail:     "delete@example.com",
		SubscriptionTier: "free",
		IsActive:         true,
	}
	err := repo.Create(context.Background(), client)
	require.NoError(t, err)

	t.Run("successfully deletes client", func(t *testing.T) {
		err := repo.Delete(context.Background(), client.ID)
		require.NoError(t, err)

		// Verify deletion
		_, err = repo.GetByID(context.Background(), client.ID)
		assert.Error(t, err)
	})
}
