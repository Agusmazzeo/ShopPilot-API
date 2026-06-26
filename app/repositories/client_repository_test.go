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

func TestClientRepository_Create(t *testing.T) {
	pool := SetupTestDB(t)
	defer CleanupTestDB(t, pool)

	// Clean up before test
	TruncateTable(t, pool, "clients")

	repo := NewClientRepository(pool)

	t.Run("successfully creates a client", func(t *testing.T) {
		client := &models.Client{
			Name:         "Test Client",
			Slug:         "test-client",
			Description:  "A test client",
			ContactEmail: "test@example.com",
			ContactPhone: "+1234567890",
			WebsiteURL:   "https://example.com",
			IsActive:     true,
		}

		err := repo.Create(context.Background(), client)
		require.NoError(t, err)
		assert.NotEqual(t, uuid.Nil, client.ID)
		assert.NotZero(t, client.CreatedAt)
		assert.NotZero(t, client.UpdatedAt)
	})

	t.Run("fails with duplicate slug", func(t *testing.T) {
		client1 := &models.Client{
			Name:         "Client One",
			Slug:         "unique-slug",
			Description:  "First client",
			ContactEmail: "client1@example.com",
			ContactPhone: "+1234567890",
			WebsiteURL:   "https://client1.com",
			IsActive:     true,
		}
		err := repo.Create(context.Background(), client1)
		require.NoError(t, err)

		client2 := &models.Client{
			Name:         "Client Two",
			Slug:         "unique-slug", // Same slug
			Description:  "Second client",
			ContactEmail: "client2@example.com",
			ContactPhone: "+0987654321",
			WebsiteURL:   "https://client2.com",
			IsActive:     true,
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
		Name:         "Test Client",
		Slug:         "test-client",
		Description:  "A test client",
		ContactEmail: "test@example.com",
		ContactPhone: "+1234567890",
		WebsiteURL:   "https://example.com",
		IsActive:     true,
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
		_, err := repo.GetByID(context.Background(), uuid.New())
		assert.Error(t, err)
	})
}

func TestClientRepository_GetBySlug(t *testing.T) {
	pool := SetupTestDB(t)
	defer CleanupTestDB(t, pool)

	TruncateTable(t, pool, "clients")

	repo := NewClientRepository(pool)

	client := &models.Client{
		Name:         "Test Client",
		Slug:         "test-slug",
		Description:  "A test client",
		ContactEmail: "test@example.com",
		ContactPhone: "+1234567890",
		WebsiteURL:   "https://example.com",
		IsActive:     true,
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
			Name:         fmt.Sprintf("Client %d", i),
			Slug:         fmt.Sprintf("client-%d", i),
			Description:  fmt.Sprintf("Client %d description", i),
			ContactEmail: fmt.Sprintf("client%d@example.com", i),
			ContactPhone: fmt.Sprintf("+123456789%d", i),
			WebsiteURL:   fmt.Sprintf("https://client%d.com", i),
			IsActive:     i%2 == 1, // Odd clients are active
		}
		err := repo.Create(context.Background(), client)
		require.NoError(t, err)
	}

	t.Run("successfully lists all clients with pagination", func(t *testing.T) {
		clients, err := repo.List(context.Background(), 10, 0)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(clients), 3)
	})

	t.Run("respects limit parameter", func(t *testing.T) {
		clients, err := repo.List(context.Background(), 2, 0)
		require.NoError(t, err)
		assert.LessOrEqual(t, len(clients), 2)
	})

	t.Run("respects offset parameter", func(t *testing.T) {
		allClients, err := repo.List(context.Background(), 100, 0)
		require.NoError(t, err)

		if len(allClients) > 1 {
			offsetClients, err := repo.List(context.Background(), 100, 1)
			require.NoError(t, err)
			assert.Less(t, len(offsetClients), len(allClients))
		}
	})
}

func TestClientRepository_ListActive(t *testing.T) {
	pool := SetupTestDB(t)
	defer CleanupTestDB(t, pool)

	TruncateTable(t, pool, "clients")

	repo := NewClientRepository(pool)

	// Create multiple clients with different active states
	for i := 1; i <= 5; i++ {
		client := &models.Client{
			Name:         fmt.Sprintf("Client %d", i),
			Slug:         fmt.Sprintf("client-active-%d", i),
			Description:  fmt.Sprintf("Client %d description", i),
			ContactEmail: fmt.Sprintf("client%d@example.com", i),
			ContactPhone: fmt.Sprintf("+123456789%d", i),
			WebsiteURL:   fmt.Sprintf("https://client%d.com", i),
			IsActive:     i <= 3, // First 3 are active, last 2 are not
		}
		err := repo.Create(context.Background(), client)
		require.NoError(t, err)
	}

	t.Run("successfully lists only active clients", func(t *testing.T) {
		clients, err := repo.ListActive(context.Background(), 10, 0)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(clients), 3)

		// Verify all returned clients are active
		for _, client := range clients {
			assert.True(t, client.IsActive)
		}
	})

	t.Run("respects limit parameter", func(t *testing.T) {
		clients, err := repo.ListActive(context.Background(), 2, 0)
		require.NoError(t, err)
		assert.LessOrEqual(t, len(clients), 2)

		// Verify all returned clients are active
		for _, client := range clients {
			assert.True(t, client.IsActive)
		}
	})
}

func TestClientRepository_Update(t *testing.T) {
	pool := SetupTestDB(t)
	defer CleanupTestDB(t, pool)

	TruncateTable(t, pool, "clients")

	repo := NewClientRepository(pool)

	client := &models.Client{
		Name:         "Original Name",
		Slug:         "original-slug",
		Description:  "Original description",
		ContactEmail: "original@example.com",
		ContactPhone: "+1234567890",
		WebsiteURL:   "https://original.com",
		IsActive:     true,
	}
	err := repo.Create(context.Background(), client)
	require.NoError(t, err)

	t.Run("successfully updates client", func(t *testing.T) {
		client.Name = "Updated Name"
		client.Description = "Updated description"
		client.WebsiteURL = "https://updated.com"
		err := repo.Update(context.Background(), client)
		require.NoError(t, err)

		// Verify update
		found, err := repo.GetByID(context.Background(), client.ID)
		require.NoError(t, err)
		assert.Equal(t, "Updated Name", found.Name)
		assert.Equal(t, "Updated description", found.Description)
		assert.Equal(t, "https://updated.com", found.WebsiteURL)
	})
}

func TestClientRepository_Delete(t *testing.T) {
	pool := SetupTestDB(t)
	defer CleanupTestDB(t, pool)

	TruncateTable(t, pool, "clients")

	repo := NewClientRepository(pool)

	client := &models.Client{
		Name:         "To Delete",
		Slug:         "to-delete",
		Description:  "Will be deleted",
		ContactEmail: "delete@example.com",
		ContactPhone: "+1234567890",
		WebsiteURL:   "https://delete.com",
		IsActive:     true,
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

	t.Run("returns error when deleting non-existent client", func(t *testing.T) {
		err := repo.Delete(context.Background(), uuid.New())
		assert.Error(t, err)
	})
}
