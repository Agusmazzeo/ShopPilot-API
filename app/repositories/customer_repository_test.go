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

func TestCustomerRepository_Create(t *testing.T) {
	pool := SetupTestDB(t)
	defer CleanupTestDB(t, pool)

	TruncateTable(t, pool, "customers")
	TruncateTable(t, pool, "clients")

	clientRepo := NewClientRepository(pool)
	client := &models.Client{
		Name:         "Test Client",
		Slug:         "test-client-customers",
		Description:  "Test client for customer tests",
		ContactEmail: "client@test.com",
		ContactPhone: "+1234567890",
		WebsiteURL:   "https://test.com",
		IsActive:     true,
	}
	err := clientRepo.Create(context.Background(), client)
	require.NoError(t, err)

	repo := NewCustomerRepository(pool)

	t.Run("successfully creates a customer", func(t *testing.T) {
		customer := &models.Customer{
			ClientID:           client.ID,
			Code:               "CUST001",
			FirstName:          "John",
			LastName:           "Doe",
			Email:              "john.doe@test.com",
			Phone:              "+1234567890",
			ShippingAddress:    "123 Shipping St",
			ShippingCity:       "Shipping City",
			ShippingState:      "SC",
			ShippingPostalCode: "12345",
			ShippingCountry:    "USA",
			BillingAddress:     "456 Billing Ave",
			BillingCity:        "Billing City",
			BillingState:       "BC",
			BillingPostalCode:  "67890",
			BillingCountry:     "USA",
			TaxID:              "TAX123",
			Notes:              "Test customer notes",
			IsActive:           true,
		}

		err := repo.Create(context.Background(), customer)
		require.NoError(t, err)
		assert.NotEqual(t, uuid.Nil, customer.ID)
		assert.NotZero(t, customer.CreatedAt)
		assert.NotZero(t, customer.UpdatedAt)
	})

	t.Run("fails with duplicate code", func(t *testing.T) {
		customer1 := &models.Customer{
			ClientID:  client.ID,
			Code:      "UNIQUE001",
			FirstName: "Jane",
			LastName:  "Smith",
			Email:     "jane@test.com",
			IsActive:  true,
		}
		err := repo.Create(context.Background(), customer1)
		require.NoError(t, err)

		customer2 := &models.Customer{
			ClientID:  client.ID,
			Code:      "UNIQUE001", // Same code
			FirstName: "Bob",
			LastName:  "Jones",
			Email:     "bob@test.com",
			IsActive:  true,
		}
		err = repo.Create(context.Background(), customer2)
		assert.Error(t, err)
	})
}

func TestCustomerRepository_GetByID(t *testing.T) {
	pool := SetupTestDB(t)
	defer CleanupTestDB(t, pool)

	TruncateTable(t, pool, "customers")
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

	repo := NewCustomerRepository(pool)

	customer := &models.Customer{
		ClientID:  client.ID,
		Code:      "CUST001",
		FirstName: "John",
		LastName:  "Doe",
		Email:     "john@test.com",
		IsActive:  true,
	}
	err = repo.Create(context.Background(), customer)
	require.NoError(t, err)

	t.Run("successfully retrieves customer by ID", func(t *testing.T) {
		found, err := repo.GetByID(context.Background(), client.ID, customer.ID)
		require.NoError(t, err)
		assert.Equal(t, customer.ID, found.ID)
		assert.Equal(t, customer.Code, found.Code)
		assert.Equal(t, customer.FirstName, found.FirstName)
		assert.Equal(t, customer.LastName, found.LastName)
	})

	t.Run("returns error for non-existent ID", func(t *testing.T) {
		_, err := repo.GetByID(context.Background(), client.ID, uuid.New())
		assert.Error(t, err)
	})
}

func TestCustomerRepository_GetByCode(t *testing.T) {
	pool := SetupTestDB(t)
	defer CleanupTestDB(t, pool)

	TruncateTable(t, pool, "customers")
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

	repo := NewCustomerRepository(pool)

	customer := &models.Customer{
		ClientID:  client.ID,
		Code:      "CODE123",
		FirstName: "Jane",
		LastName:  "Smith",
		Email:     "jane@test.com",
		IsActive:  true,
	}
	err = repo.Create(context.Background(), customer)
	require.NoError(t, err)

	t.Run("successfully retrieves customer by code", func(t *testing.T) {
		found, err := repo.GetByCode(context.Background(), client.ID, "CODE123")
		require.NoError(t, err)
		assert.Equal(t, customer.ID, found.ID)
		assert.Equal(t, "CODE123", found.Code)
	})

	t.Run("returns error for non-existent code", func(t *testing.T) {
		_, err := repo.GetByCode(context.Background(), client.ID, "NONEXISTENT")
		assert.Error(t, err)
	})
}

func TestCustomerRepository_GetByEmail(t *testing.T) {
	pool := SetupTestDB(t)
	defer CleanupTestDB(t, pool)

	TruncateTable(t, pool, "customers")
	TruncateTable(t, pool, "clients")

	clientRepo := NewClientRepository(pool)
	client := &models.Client{
		Name:         "Test Client",
		Slug:         "test-client-email",
		Description:  "Test client",
		ContactEmail: "client@test.com",
		ContactPhone: "+1234567890",
		WebsiteURL:   "https://test.com",
		IsActive:     true,
	}
	err := clientRepo.Create(context.Background(), client)
	require.NoError(t, err)

	repo := NewCustomerRepository(pool)

	customer := &models.Customer{
		ClientID:  client.ID,
		Code:      "CUST001",
		FirstName: "Alice",
		LastName:  "Johnson",
		Email:     "alice.johnson@test.com",
		IsActive:  true,
	}
	err = repo.Create(context.Background(), customer)
	require.NoError(t, err)

	t.Run("successfully retrieves customer by email", func(t *testing.T) {
		found, err := repo.GetByEmail(context.Background(), client.ID, "alice.johnson@test.com")
		require.NoError(t, err)
		assert.Equal(t, customer.ID, found.ID)
		assert.Equal(t, "alice.johnson@test.com", found.Email)
	})

	t.Run("returns error for non-existent email", func(t *testing.T) {
		_, err := repo.GetByEmail(context.Background(), client.ID, "nonexistent@test.com")
		assert.Error(t, err)
	})
}

func TestCustomerRepository_Update(t *testing.T) {
	pool := SetupTestDB(t)
	defer CleanupTestDB(t, pool)

	TruncateTable(t, pool, "customers")
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

	repo := NewCustomerRepository(pool)

	customer := &models.Customer{
		ClientID:  client.ID,
		Code:      "CUST001",
		FirstName: "Original",
		LastName:  "Name",
		Email:     "original@test.com",
		Phone:     "+1111111111",
		IsActive:  true,
	}
	err = repo.Create(context.Background(), customer)
	require.NoError(t, err)

	t.Run("successfully updates customer", func(t *testing.T) {
		customer.FirstName = "Updated"
		customer.LastName = "Customer"
		customer.Email = "updated@test.com"
		customer.Phone = "+2222222222"
		err := repo.Update(context.Background(), customer)
		require.NoError(t, err)

		found, err := repo.GetByID(context.Background(), client.ID, customer.ID)
		require.NoError(t, err)
		assert.Equal(t, "Updated", found.FirstName)
		assert.Equal(t, "Customer", found.LastName)
		assert.Equal(t, "updated@test.com", found.Email)
		assert.Equal(t, "+2222222222", found.Phone)
	})

	t.Run("returns error when updating non-existent customer", func(t *testing.T) {
		nonExistent := &models.Customer{
			ClientID:  client.ID,
			ID:        uuid.New(),
			FirstName: "Non",
			LastName:  "Existent",
		}
		err := repo.Update(context.Background(), nonExistent)
		assert.Error(t, err)
	})
}

func TestCustomerRepository_Delete(t *testing.T) {
	pool := SetupTestDB(t)
	defer CleanupTestDB(t, pool)

	TruncateTable(t, pool, "customers")
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

	repo := NewCustomerRepository(pool)

	customer := &models.Customer{
		ClientID:  client.ID,
		Code:      "CUST001",
		FirstName: "To",
		LastName:  "Delete",
		Email:     "delete@test.com",
		IsActive:  true,
	}
	err = repo.Create(context.Background(), customer)
	require.NoError(t, err)

	t.Run("successfully deletes customer", func(t *testing.T) {
		err := repo.Delete(context.Background(), client.ID, customer.ID)
		require.NoError(t, err)

		_, err = repo.GetByID(context.Background(), client.ID, customer.ID)
		assert.Error(t, err)
	})

	t.Run("returns error when deleting non-existent customer", func(t *testing.T) {
		err := repo.Delete(context.Background(), client.ID, uuid.New())
		assert.Error(t, err)
	})
}

func TestCustomerRepository_ListByClient(t *testing.T) {
	pool := SetupTestDB(t)
	defer CleanupTestDB(t, pool)

	TruncateTable(t, pool, "customers")
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

	repo := NewCustomerRepository(pool)

	// Create multiple customers
	for i := 1; i <= 5; i++ {
		customer := &models.Customer{
			ClientID:  client.ID,
			Code:      fmt.Sprintf("CUST%03d", i),
			FirstName: fmt.Sprintf("Customer%d", i),
			LastName:  fmt.Sprintf("Last%d", i),
			Email:     fmt.Sprintf("customer%d@test.com", i),
			IsActive:  true,
		}
		err := repo.Create(context.Background(), customer)
		require.NoError(t, err)
	}

	t.Run("successfully lists customers with pagination", func(t *testing.T) {
		customers, err := repo.ListByClient(context.Background(), client.ID, 10, 0)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(customers), 5)
	})

	t.Run("respects limit parameter", func(t *testing.T) {
		customers, err := repo.ListByClient(context.Background(), client.ID, 3, 0)
		require.NoError(t, err)
		assert.LessOrEqual(t, len(customers), 3)
	})

	t.Run("respects offset parameter", func(t *testing.T) {
		allCustomers, err := repo.ListByClient(context.Background(), client.ID, 100, 0)
		require.NoError(t, err)

		if len(allCustomers) > 1 {
			offsetCustomers, err := repo.ListByClient(context.Background(), client.ID, 100, 1)
			require.NoError(t, err)
			assert.Less(t, len(offsetCustomers), len(allCustomers))
		}
	})
}

func TestCustomerRepository_Search(t *testing.T) {
	pool := SetupTestDB(t)
	defer CleanupTestDB(t, pool)

	TruncateTable(t, pool, "customers")
	TruncateTable(t, pool, "clients")

	clientRepo := NewClientRepository(pool)
	client := &models.Client{
		Name:         "Test Client",
		Slug:         "test-client-search",
		Description:  "Test client",
		ContactEmail: "client@test.com",
		ContactPhone: "+1234567890",
		WebsiteURL:   "https://test.com",
		IsActive:     true,
	}
	err := clientRepo.Create(context.Background(), client)
	require.NoError(t, err)

	repo := NewCustomerRepository(pool)

	// Create customers with various names and emails
	customers := []*models.Customer{
		{ClientID: client.ID, Code: "C001", FirstName: "John", LastName: "Doe", Email: "john.doe@test.com", IsActive: true},
		{ClientID: client.ID, Code: "C002", FirstName: "Jane", LastName: "Smith", Email: "jane.smith@test.com", IsActive: true},
		{ClientID: client.ID, Code: "C003", FirstName: "Bob", LastName: "Johnson", Email: "bob.johnson@test.com", IsActive: true},
		{ClientID: client.ID, Code: "C004", FirstName: "Alice", LastName: "Williams", Email: "alice@example.com", IsActive: true},
	}

	for _, c := range customers {
		err := repo.Create(context.Background(), c)
		require.NoError(t, err)
	}

	t.Run("search by first name", func(t *testing.T) {
		results, err := repo.Search(context.Background(), client.ID, "John", 10, 0)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(results), 1)

		found := false
		for _, r := range results {
			if r.FirstName == "John" {
				found = true
				break
			}
		}
		assert.True(t, found)
	})

	t.Run("search by last name", func(t *testing.T) {
		results, err := repo.Search(context.Background(), client.ID, "Smith", 10, 0)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(results), 1)

		found := false
		for _, r := range results {
			if r.LastName == "Smith" {
				found = true
				break
			}
		}
		assert.True(t, found)
	})

	t.Run("search by email", func(t *testing.T) {
		results, err := repo.Search(context.Background(), client.ID, "jane.smith", 10, 0)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(results), 1)
	})

	t.Run("search by code", func(t *testing.T) {
		results, err := repo.Search(context.Background(), client.ID, "C003", 10, 0)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(results), 1)

		found := false
		for _, r := range results {
			if r.Code == "C003" {
				found = true
				break
			}
		}
		assert.True(t, found)
	})

	t.Run("search with no matches", func(t *testing.T) {
		results, err := repo.Search(context.Background(), client.ID, "NonExistentSearchTerm", 10, 0)
		require.NoError(t, err)
		assert.Equal(t, 0, len(results))
	})
}
