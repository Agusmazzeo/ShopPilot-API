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

func TestSupplierRepository_Create(t *testing.T) {
	pool := SetupTestDB(t)
	defer CleanupTestDB(t, pool)

	TruncateTable(t, pool, "suppliers")
	TruncateTable(t, pool, "clients")

	// Create test client first (suppliers FK to clients)
	clientRepo := NewClientRepository(pool)
	client := &models.Client{
		Name:         "Test Client",
		Slug:         "test-client-suppliers",
		Description:  "Test client for supplier tests",
		ContactEmail: "client@test.com",
		ContactPhone: "+1234567890",
		WebsiteURL:   "https://test.com",
		IsActive:     true,
	}
	err := clientRepo.Create(context.Background(), client)
	require.NoError(t, err)

	repo := NewSupplierRepository(pool)

	t.Run("successfully creates a supplier", func(t *testing.T) {
		supplier := &models.Supplier{
			ClientID:     client.ID,
			Code:         "SUP001",
			Name:         "Test Supplier",
			Email:        "supplier@test.com",
			Phone:        "+1234567890",
			Address:      "123 Test St",
			City:         "Test City",
			State:        "TS",
			PostalCode:   "12345",
			Country:      "Test Country",
			TaxID:        "TAX123",
			PaymentTerms: "Net 30",
			Currency:     "USD",
			Notes:        "Test supplier notes",
			IsActive:     true,
		}

		err := repo.Create(context.Background(), supplier)
		require.NoError(t, err)
		assert.NotEqual(t, uuid.Nil, supplier.ID)
		assert.NotZero(t, supplier.CreatedAt)
		assert.NotZero(t, supplier.UpdatedAt)
	})

	t.Run("fails with duplicate code", func(t *testing.T) {
		supplier1 := &models.Supplier{
			ClientID: client.ID,
			Code:     "UNIQUE001",
			Name:     "First Supplier",
			Email:    "first@test.com",
			IsActive: true,
		}
		err := repo.Create(context.Background(), supplier1)
		require.NoError(t, err)

		supplier2 := &models.Supplier{
			ClientID: client.ID,
			Code:     "UNIQUE001", // Same code
			Name:     "Second Supplier",
			Email:    "second@test.com",
			IsActive: true,
		}
		err = repo.Create(context.Background(), supplier2)
		assert.Error(t, err)
	})
}

func TestSupplierRepository_GetByID(t *testing.T) {
	pool := SetupTestDB(t)
	defer CleanupTestDB(t, pool)

	TruncateTable(t, pool, "suppliers")
	TruncateTable(t, pool, "clients")

	clientRepo := NewClientRepository(pool)
	client := &models.Client{
		Name:         "Test Client",
		Slug:         "test-client-get",
		Description:  "Test client",
		ContactEmail: "client@test.com",
		ContactPhone: "+1234567890",
		WebsiteURL:   "https://test.com",
		IsActive:     true,
	}
	err := clientRepo.Create(context.Background(), client)
	require.NoError(t, err)

	repo := NewSupplierRepository(pool)

	supplier := &models.Supplier{
		ClientID: client.ID,
		Code:     "SUP001",
		Name:     "Test Supplier",
		Email:    "supplier@test.com",
		IsActive: true,
	}
	err = repo.Create(context.Background(), supplier)
	require.NoError(t, err)

	t.Run("successfully retrieves supplier by ID", func(t *testing.T) {
		found, err := repo.GetByID(context.Background(), client.ID, supplier.ID)
		require.NoError(t, err)
		assert.Equal(t, supplier.ID, found.ID)
		assert.Equal(t, supplier.Code, found.Code)
		assert.Equal(t, supplier.Name, found.Name)
		assert.Equal(t, supplier.Email, found.Email)
	})

	t.Run("returns error for non-existent ID", func(t *testing.T) {
		_, err := repo.GetByID(context.Background(), client.ID, uuid.New())
		assert.Error(t, err)
	})
}

func TestSupplierRepository_GetByCode(t *testing.T) {
	pool := SetupTestDB(t)
	defer CleanupTestDB(t, pool)

	TruncateTable(t, pool, "suppliers")
	TruncateTable(t, pool, "clients")

	clientRepo := NewClientRepository(pool)
	client := &models.Client{
		Name:         "Test Client",
		Slug:         "test-client-code",
		Description:  "Test client",
		ContactEmail: "client@test.com",
		ContactPhone: "+1234567890",
		WebsiteURL:   "https://test.com",
		IsActive:     true,
	}
	err := clientRepo.Create(context.Background(), client)
	require.NoError(t, err)

	repo := NewSupplierRepository(pool)

	supplier := &models.Supplier{
		ClientID: client.ID,
		Code:     "CODE123",
		Name:     "Test Supplier",
		Email:    "supplier@test.com",
		IsActive: true,
	}
	err = repo.Create(context.Background(), supplier)
	require.NoError(t, err)

	t.Run("successfully retrieves supplier by code", func(t *testing.T) {
		found, err := repo.GetByCode(context.Background(), client.ID, "CODE123")
		require.NoError(t, err)
		assert.Equal(t, supplier.ID, found.ID)
		assert.Equal(t, "CODE123", found.Code)
	})

	t.Run("returns error for non-existent code", func(t *testing.T) {
		_, err := repo.GetByCode(context.Background(), client.ID, "NONEXISTENT")
		assert.Error(t, err)
	})
}

func TestSupplierRepository_Update(t *testing.T) {
	pool := SetupTestDB(t)
	defer CleanupTestDB(t, pool)

	TruncateTable(t, pool, "suppliers")
	TruncateTable(t, pool, "clients")

	clientRepo := NewClientRepository(pool)
	client := &models.Client{
		Name:         "Test Client",
		Slug:         "test-client-update",
		Description:  "Test client",
		ContactEmail: "client@test.com",
		ContactPhone: "+1234567890",
		WebsiteURL:   "https://test.com",
		IsActive:     true,
	}
	err := clientRepo.Create(context.Background(), client)
	require.NoError(t, err)

	repo := NewSupplierRepository(pool)

	supplier := &models.Supplier{
		ClientID: client.ID,
		Code:     "SUP001",
		Name:     "Original Name",
		Email:    "original@test.com",
		Phone:    "+1111111111",
		IsActive: true,
	}
	err = repo.Create(context.Background(), supplier)
	require.NoError(t, err)

	t.Run("successfully updates supplier", func(t *testing.T) {
		supplier.Name = "Updated Name"
		supplier.Email = "updated@test.com"
		supplier.Phone = "+2222222222"
		err := repo.Update(context.Background(), supplier)
		require.NoError(t, err)

		found, err := repo.GetByID(context.Background(), client.ID, supplier.ID)
		require.NoError(t, err)
		assert.Equal(t, "Updated Name", found.Name)
		assert.Equal(t, "updated@test.com", found.Email)
		assert.Equal(t, "+2222222222", found.Phone)
	})

	t.Run("returns error when updating non-existent supplier", func(t *testing.T) {
		nonExistent := &models.Supplier{
			ClientID: client.ID,
			ID:       uuid.New(),
			Name:     "Non-existent",
		}
		err := repo.Update(context.Background(), nonExistent)
		assert.Error(t, err)
	})
}

func TestSupplierRepository_Delete(t *testing.T) {
	pool := SetupTestDB(t)
	defer CleanupTestDB(t, pool)

	TruncateTable(t, pool, "suppliers")
	TruncateTable(t, pool, "clients")

	clientRepo := NewClientRepository(pool)
	client := &models.Client{
		Name:         "Test Client",
		Slug:         "test-client-delete",
		Description:  "Test client",
		ContactEmail: "client@test.com",
		ContactPhone: "+1234567890",
		WebsiteURL:   "https://test.com",
		IsActive:     true,
	}
	err := clientRepo.Create(context.Background(), client)
	require.NoError(t, err)

	repo := NewSupplierRepository(pool)

	supplier := &models.Supplier{
		ClientID: client.ID,
		Code:     "SUP001",
		Name:     "To Delete",
		Email:    "delete@test.com",
		IsActive: true,
	}
	err = repo.Create(context.Background(), supplier)
	require.NoError(t, err)

	t.Run("successfully deletes supplier", func(t *testing.T) {
		err := repo.Delete(context.Background(), client.ID, supplier.ID)
		require.NoError(t, err)

		_, err = repo.GetByID(context.Background(), client.ID, supplier.ID)
		assert.Error(t, err)
	})

	t.Run("returns error when deleting non-existent supplier", func(t *testing.T) {
		err := repo.Delete(context.Background(), client.ID, uuid.New())
		assert.Error(t, err)
	})
}

func TestSupplierRepository_ListByClient(t *testing.T) {
	pool := SetupTestDB(t)
	defer CleanupTestDB(t, pool)

	TruncateTable(t, pool, "suppliers")
	TruncateTable(t, pool, "clients")

	clientRepo := NewClientRepository(pool)
	client := &models.Client{
		Name:         "Test Client",
		Slug:         "test-client-list",
		Description:  "Test client",
		ContactEmail: "client@test.com",
		ContactPhone: "+1234567890",
		WebsiteURL:   "https://test.com",
		IsActive:     true,
	}
	err := clientRepo.Create(context.Background(), client)
	require.NoError(t, err)

	repo := NewSupplierRepository(pool)

	// Create multiple suppliers
	for i := 1; i <= 5; i++ {
		supplier := &models.Supplier{
			ClientID: client.ID,
			Code:     fmt.Sprintf("SUP%03d", i),
			Name:     fmt.Sprintf("Supplier %d", i),
			Email:    fmt.Sprintf("supplier%d@test.com", i),
			IsActive: i%2 == 1, // Odd suppliers are active
		}
		err := repo.Create(context.Background(), supplier)
		require.NoError(t, err)
	}

	t.Run("successfully lists suppliers with pagination", func(t *testing.T) {
		suppliers, err := repo.ListByClient(context.Background(), client.ID, 10, 0)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(suppliers), 5)
	})

	t.Run("respects limit parameter", func(t *testing.T) {
		suppliers, err := repo.ListByClient(context.Background(), client.ID, 3, 0)
		require.NoError(t, err)
		assert.LessOrEqual(t, len(suppliers), 3)
	})

	t.Run("respects offset parameter", func(t *testing.T) {
		allSuppliers, err := repo.ListByClient(context.Background(), client.ID, 100, 0)
		require.NoError(t, err)

		if len(allSuppliers) > 1 {
			offsetSuppliers, err := repo.ListByClient(context.Background(), client.ID, 100, 1)
			require.NoError(t, err)
			assert.Less(t, len(offsetSuppliers), len(allSuppliers))
		}
	})
}

func TestSupplierRepository_ListActive(t *testing.T) {
	pool := SetupTestDB(t)
	defer CleanupTestDB(t, pool)

	TruncateTable(t, pool, "suppliers")
	TruncateTable(t, pool, "clients")

	clientRepo := NewClientRepository(pool)
	client := &models.Client{
		Name:         "Test Client",
		Slug:         "test-client-active",
		Description:  "Test client",
		ContactEmail: "client@test.com",
		ContactPhone: "+1234567890",
		WebsiteURL:   "https://test.com",
		IsActive:     true,
	}
	err := clientRepo.Create(context.Background(), client)
	require.NoError(t, err)

	repo := NewSupplierRepository(pool)

	// Create multiple suppliers with different active states
	for i := 1; i <= 5; i++ {
		supplier := &models.Supplier{
			ClientID: client.ID,
			Code:     fmt.Sprintf("SUP-ACTIVE-%03d", i),
			Name:     fmt.Sprintf("Supplier %d", i),
			Email:    fmt.Sprintf("supplier%d@test.com", i),
			IsActive: i <= 3, // First 3 are active
		}
		err := repo.Create(context.Background(), supplier)
		require.NoError(t, err)
	}

	t.Run("successfully lists only active suppliers", func(t *testing.T) {
		suppliers, err := repo.ListActive(context.Background(), client.ID, 10, 0)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(suppliers), 3)

		for _, supplier := range suppliers {
			assert.True(t, supplier.IsActive)
		}
	})

	t.Run("respects limit parameter", func(t *testing.T) {
		suppliers, err := repo.ListActive(context.Background(), client.ID, 2, 0)
		require.NoError(t, err)
		assert.LessOrEqual(t, len(suppliers), 2)

		for _, supplier := range suppliers {
			assert.True(t, supplier.IsActive)
		}
	})
}
