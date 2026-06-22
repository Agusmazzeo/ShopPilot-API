package repositories

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yourorg/shoppilot/app/models"
)

func TestUserRepository_Create(t *testing.T) {
	pool := SetupTestDB(t)
	defer CleanupTestDB(t, pool)

	TruncateTable(t, pool, "users")

	repo := NewUserRepository(pool)

	t.Run("successfully creates a user", func(t *testing.T) {
		user := &models.User{
			ClientID:     1,
			Email:        "test@example.com",
			PasswordHash: "hashedpassword",
			FirstName:    "John",
			LastName:     "Doe",
			StatusID:     1,
			IsActive:     true,
		}

		err := repo.Create(context.Background(), user)
		require.NoError(t, err)
		assert.Greater(t, user.ID, 0)
		assert.NotZero(t, user.CreatedAt)
		assert.NotZero(t, user.UpdatedAt)
	})

	t.Run("fails with duplicate email within same client", func(t *testing.T) {
		user1 := &models.User{
			ClientID:     1,
			Email:        "duplicate@example.com",
			PasswordHash: "hashedpassword",
			FirstName:    "User",
			LastName:     "One",
			StatusID:     1,
			IsActive:     true,
		}
		err := repo.Create(context.Background(), user1)
		require.NoError(t, err)

		user2 := &models.User{
			ClientID:     1,
			Email:        "duplicate@example.com", // Same email, same client
			PasswordHash: "hashedpassword",
			FirstName:    "User",
			LastName:     "Two",
			StatusID:     1,
			IsActive:     true,
		}
		err = repo.Create(context.Background(), user2)
		assert.Error(t, err)
	})

	t.Run("allows duplicate email across different clients", func(t *testing.T) {
		user1 := &models.User{
			ClientID:     1,
			Email:        "shared@example.com",
			PasswordHash: "hashedpassword",
			FirstName:    "User",
			LastName:     "One",
			StatusID:     1,
			IsActive:     true,
		}
		err := repo.Create(context.Background(), user1)
		require.NoError(t, err)

		user2 := &models.User{
			ClientID:     2, // Different client
			Email:        "shared@example.com", // Same email
			PasswordHash: "hashedpassword",
			FirstName:    "User",
			LastName:     "Two",
			StatusID:     1,
			IsActive:     true,
		}
		err = repo.Create(context.Background(), user2)
		require.NoError(t, err)
		assert.NotEqual(t, user1.ID, user2.ID)
	})
}

func TestUserRepository_GetByID(t *testing.T) {
	pool := SetupTestDB(t)
	defer CleanupTestDB(t, pool)

	TruncateTable(t, pool, "users")

	repo := NewUserRepository(pool)

	user := &models.User{
		ClientID:     1,
		Email:        "test@example.com",
		PasswordHash: "hashedpassword",
		FirstName:    "John",
		LastName:     "Doe",
		StatusID:     1,
		IsActive:     true,
	}
	err := repo.Create(context.Background(), user)
	require.NoError(t, err)

	t.Run("successfully retrieves user by ID", func(t *testing.T) {
		found, err := repo.GetByID(context.Background(), user.ID)
		require.NoError(t, err)
		assert.Equal(t, user.ID, found.ID)
		assert.Equal(t, user.Email, found.Email)
		assert.Equal(t, user.FirstName, found.FirstName)
	})

	t.Run("returns error for non-existent ID", func(t *testing.T) {
		_, err := repo.GetByID(context.Background(), 99999)
		assert.Error(t, err)
	})
}

func TestUserRepository_GetByEmail(t *testing.T) {
	pool := SetupTestDB(t)
	defer CleanupTestDB(t, pool)

	TruncateTable(t, pool, "users")

	repo := NewUserRepository(pool)

	user := &models.User{
		ClientID:     1,
		Email:        "test@example.com",
		PasswordHash: "hashedpassword",
		FirstName:    "John",
		LastName:     "Doe",
		StatusID:     1,
		IsActive:     true,
	}
	err := repo.Create(context.Background(), user)
	require.NoError(t, err)

	t.Run("successfully retrieves user by email and client", func(t *testing.T) {
		found, err := repo.GetByEmail(context.Background(), 1, "test@example.com")
		require.NoError(t, err)
		assert.Equal(t, user.ID, found.ID)
		assert.Equal(t, user.Email, found.Email)
	})

	t.Run("returns error for non-existent email", func(t *testing.T) {
		_, err := repo.GetByEmail(context.Background(), 1, "nonexistent@example.com")
		assert.Error(t, err)
	})

	t.Run("returns error for wrong client", func(t *testing.T) {
		_, err := repo.GetByEmail(context.Background(), 2, "test@example.com")
		assert.Error(t, err)
	})
}

func TestUserRepository_ListByClientID(t *testing.T) {
	pool := SetupTestDB(t)
	defer CleanupTestDB(t, pool)

	TruncateTable(t, pool, "users")

	repo := NewUserRepository(pool)

	// Create users for client 1
	for i := 1; i <= 3; i++ {
		user := &models.User{
			ClientID:     1,
			Email:        fmt.Sprintf("user%d@client1.com", i),
			PasswordHash: "hashedpassword",
			FirstName:    fmt.Sprintf("User%d", i),
			LastName:     "Client1",
			StatusID:     1,
			IsActive:     true,
		}
		err := repo.Create(context.Background(), user)
		require.NoError(t, err)
	}

	// Create users for client 2
	for i := 1; i <= 2; i++ {
		user := &models.User{
			ClientID:     2,
			Email:        fmt.Sprintf("user%d@client2.com", i),
			PasswordHash: "hashedpassword",
			FirstName:    fmt.Sprintf("User%d", i),
			LastName:     "Client2",
			StatusID:     1,
			IsActive:     true,
		}
		err := repo.Create(context.Background(), user)
		require.NoError(t, err)
	}

	t.Run("successfully lists users for client 1", func(t *testing.T) {
		users, err := repo.ListByClientID(context.Background(), 1)
		require.NoError(t, err)
		assert.Len(t, users, 3)
	})

	t.Run("successfully lists users for client 2", func(t *testing.T) {
		users, err := repo.ListByClientID(context.Background(), 2)
		require.NoError(t, err)
		assert.Len(t, users, 2)
	})

	t.Run("returns empty list for client with no users", func(t *testing.T) {
		users, err := repo.ListByClientID(context.Background(), 99)
		require.NoError(t, err)
		assert.Empty(t, users)
	})
}

func TestUserRepository_Update(t *testing.T) {
	pool := SetupTestDB(t)
	defer CleanupTestDB(t, pool)

	TruncateTable(t, pool, "users")

	repo := NewUserRepository(pool)

	user := &models.User{
		ClientID:     1,
		Email:        "test@example.com",
		PasswordHash: "hashedpassword",
		FirstName:    "John",
		LastName:     "Doe",
		StatusID:     1,
		IsActive:     true,
	}
	err := repo.Create(context.Background(), user)
	require.NoError(t, err)

	t.Run("successfully updates user", func(t *testing.T) {
		user.FirstName = "Jane"
		user.LastName = "Smith"
		user.StatusID = 2
		err := repo.Update(context.Background(), user)
		require.NoError(t, err)

		// Verify update
		found, err := repo.GetByID(context.Background(), user.ID)
		require.NoError(t, err)
		assert.Equal(t, "Jane", found.FirstName)
		assert.Equal(t, "Smith", found.LastName)
		assert.Equal(t, 2, found.StatusID)
	})

	t.Run("returns error for non-existent user", func(t *testing.T) {
		nonExistent := &models.User{
			ID:        99999,
			FirstName: "Test",
			LastName:  "User",
			StatusID:  1,
			IsActive:  true,
		}
		err := repo.Update(context.Background(), nonExistent)
		assert.Error(t, err)
	})
}

func TestUserRepository_Delete(t *testing.T) {
	pool := SetupTestDB(t)
	defer CleanupTestDB(t, pool)

	TruncateTable(t, pool, "users")

	repo := NewUserRepository(pool)

	user := &models.User{
		ClientID:     1,
		Email:        "delete@example.com",
		PasswordHash: "hashedpassword",
		FirstName:    "Delete",
		LastName:     "Me",
		StatusID:     1,
		IsActive:     true,
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
		err := repo.Delete(context.Background(), 99999)
		assert.Error(t, err)
	})
}
