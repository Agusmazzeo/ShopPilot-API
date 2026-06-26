package repositories

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yourorg/shoppilot/internal/models"
)

func TestProductRepository_Create(t *testing.T) {
	pool := SetupTestDB(t)
	defer CleanupTestDB(t, pool)

	TruncateTable(t, pool, "product_variants")
	TruncateTable(t, pool, "products")
	TruncateTable(t, pool, "shops")
	TruncateTable(t, pool, "clients")

	repo := NewProductRepository(pool)

	// Create test client
	clientID := uuid.New()
	_, err := pool.Exec(context.Background(),
		`INSERT INTO clients (id, name, slug, is_active) VALUES ($1, $2, $3, $4)`,
		clientID, "Test Client", "test-client", true)
	require.NoError(t, err)

	// Create test shop
	shopID := uuid.New()
	_, err = pool.Exec(context.Background(),
		`INSERT INTO shops (id, client_id, name, slug, webpage_url, is_active) VALUES ($1, $2, $3, $4, $5, $6)`,
		shopID, clientID, "Test Shop", "test-shop", "https://test.com", true)
	require.NoError(t, err)

	t.Run("successfully creates a product", func(t *testing.T) {
		product := &models.Product{
			ClientID:    clientID,
			ShopID:      shopID,
			Code:        "PROD-001",
			Name:        "Test Product",
			Description: "Test Description",
			Metadata:    map[string]interface{}{"color": "blue"},
			IsActive:    true,
		}

		err := repo.Create(context.Background(), product)
		require.NoError(t, err)
		assert.NotEqual(t, uuid.Nil, product.ID)
		assert.NotZero(t, product.CreatedAt)
		assert.NotZero(t, product.UpdatedAt)
	})

	t.Run("fails with duplicate code within same client", func(t *testing.T) {
		product1 := &models.Product{
			ClientID:    clientID,
			ShopID:      shopID,
			Code:        "UNIQUE-CODE",
			Name:        "Product One",
			Description: "Description",
			IsActive:    true,
		}
		err := repo.Create(context.Background(), product1)
		require.NoError(t, err)

		product2 := &models.Product{
			ClientID:    clientID,
			ShopID:      shopID,
			Code:        "UNIQUE-CODE", // Same code, same client
			Name:        "Product Two",
			Description: "Description",
			IsActive:    true,
		}
		err = repo.Create(context.Background(), product2)
		assert.Error(t, err)
	})

	t.Run("allows duplicate code across different clients", func(t *testing.T) {
		client1 := uuid.New()
		_, err1 := pool.Exec(context.Background(),
			`INSERT INTO clients (id, name, slug, is_active) VALUES ($1, $2, $3, $4)`,
			client1, "Client 1", "client-1", true)
		require.NoError(t, err1)

		shop1 := uuid.New()
		_, err1 = pool.Exec(context.Background(),
			`INSERT INTO shops (id, client_id, name, slug, webpage_url, is_active) VALUES ($1, $2, $3, $4, $5, $6)`,
			shop1, client1, "Shop 1", "shop-1-unique", "https://shop1.com", true)
		require.NoError(t, err1)

		client2 := uuid.New()
		_, err2 := pool.Exec(context.Background(),
			`INSERT INTO clients (id, name, slug, is_active) VALUES ($1, $2, $3, $4)`,
			client2, "Client 2", "client-2", true)
		require.NoError(t, err2)

		shop2 := uuid.New()
		_, err2 = pool.Exec(context.Background(),
			`INSERT INTO shops (id, client_id, name, slug, webpage_url, is_active) VALUES ($1, $2, $3, $4, $5, $6)`,
			shop2, client2, "Shop 2", "shop-2-unique", "https://shop2.com", true)
		require.NoError(t, err2)

		product1 := &models.Product{
			ClientID:    client1,
			ShopID:      shop1,
			Code:        "SHARED-CODE",
			Name:        "Product One",
			Description: "Description",
			IsActive:    true,
		}
		err := repo.Create(context.Background(), product1)
		require.NoError(t, err)

		product2 := &models.Product{
			ClientID:    client2,
			ShopID:      shop2,
			Code:        "SHARED-CODE", // Same code, different client
			Name:        "Product Two",
			Description: "Description",
			IsActive:    true,
		}
		err = repo.Create(context.Background(), product2)
		require.NoError(t, err)
		assert.NotEqual(t, product1.ID, product2.ID)
	})
}

func TestProductRepository_GetByID(t *testing.T) {
	pool := SetupTestDB(t)
	defer CleanupTestDB(t, pool)

	TruncateTable(t, pool, "product_variants")
	TruncateTable(t, pool, "products")
	TruncateTable(t, pool, "shops")
	TruncateTable(t, pool, "clients")

	repo := NewProductRepository(pool)

	// Create test client
	clientID := uuid.New()
	_, err := pool.Exec(context.Background(),
		`INSERT INTO clients (id, name, slug, is_active) VALUES ($1, $2, $3, $4)`,
		clientID, "Test Client", "test-client", true)
	require.NoError(t, err)

	// Create test shop
	shopID := uuid.New()
	_, err = pool.Exec(context.Background(),
		`INSERT INTO shops (id, client_id, name, slug, webpage_url, is_active) VALUES ($1, $2, $3, $4, $5, $6)`,
		shopID, clientID, "Test Shop", "test-shop", "https://test.com", true)
	require.NoError(t, err)

	product := &models.Product{
		ClientID:    clientID,
		ShopID:      shopID,
		Code:        "PROD-GET",
		Name:        "Test Product",
		Description: "Description",
		IsActive:    true,
	}
	err = repo.Create(context.Background(), product)
	require.NoError(t, err)

	t.Run("successfully retrieves product by composite key", func(t *testing.T) {
		found, err := repo.GetByID(context.Background(), clientID, product.ID)
		require.NoError(t, err)
		assert.Equal(t, product.ID, found.ID)
		assert.Equal(t, product.ClientID, found.ClientID)
		assert.Equal(t, product.Code, found.Code)
		assert.Equal(t, product.Name, found.Name)
	})

	t.Run("returns error for non-existent ID", func(t *testing.T) {
		_, err := repo.GetByID(context.Background(), clientID, uuid.New())
		assert.Error(t, err)
	})

	t.Run("returns error for wrong client", func(t *testing.T) {
		wrongClient := uuid.New()
		_, err := repo.GetByID(context.Background(), wrongClient, product.ID)
		assert.Error(t, err)
	})
}

func TestProductRepository_GetByCode(t *testing.T) {
	pool := SetupTestDB(t)
	defer CleanupTestDB(t, pool)

	TruncateTable(t, pool, "product_variants")
	TruncateTable(t, pool, "products")
	TruncateTable(t, pool, "shops")
	TruncateTable(t, pool, "clients")

	repo := NewProductRepository(pool)

	// Create test client
	clientID := uuid.New()
	_, err := pool.Exec(context.Background(),
		`INSERT INTO clients (id, name, slug, is_active) VALUES ($1, $2, $3, $4)`,
		clientID, "Test Client", "test-client", true)
	require.NoError(t, err)

	// Create test shop
	shopID := uuid.New()
	_, err = pool.Exec(context.Background(),
		`INSERT INTO shops (id, client_id, name, slug, webpage_url, is_active) VALUES ($1, $2, $3, $4, $5, $6)`,
		shopID, clientID, "Test Shop", "test-shop", "https://test.com", true)
	require.NoError(t, err)

	product := &models.Product{
		ClientID:    clientID,
		ShopID:      shopID,
		Code:        "CODE-123",
		Name:        "Test Product",
		Description: "Description",
		IsActive:    true,
	}
	err = repo.Create(context.Background(), product)
	require.NoError(t, err)

	t.Run("successfully retrieves product by code", func(t *testing.T) {
		found, err := repo.GetByCode(context.Background(), clientID, "CODE-123")
		require.NoError(t, err)
		assert.Equal(t, product.ID, found.ID)
		assert.Equal(t, "CODE-123", found.Code)
	})

	t.Run("returns error for non-existent code", func(t *testing.T) {
		_, err := repo.GetByCode(context.Background(), clientID, "NONEXISTENT")
		assert.Error(t, err)
	})

	t.Run("returns error for wrong client", func(t *testing.T) {
		wrongClient := uuid.New()
		_, err := repo.GetByCode(context.Background(), wrongClient, "CODE-123")
		assert.Error(t, err)
	})
}

func TestProductRepository_Update(t *testing.T) {
	pool := SetupTestDB(t)
	defer CleanupTestDB(t, pool)

	TruncateTable(t, pool, "product_variants")
	TruncateTable(t, pool, "products")
	TruncateTable(t, pool, "shops")
	TruncateTable(t, pool, "clients")

	repo := NewProductRepository(pool)

	// Create test client
	clientID := uuid.New()
	_, err := pool.Exec(context.Background(),
		`INSERT INTO clients (id, name, slug, is_active) VALUES ($1, $2, $3, $4)`,
		clientID, "Test Client", "test-client", true)
	require.NoError(t, err)

	// Create test shop
	shopID := uuid.New()
	_, err = pool.Exec(context.Background(),
		`INSERT INTO shops (id, client_id, name, slug, webpage_url, is_active) VALUES ($1, $2, $3, $4, $5, $6)`,
		shopID, clientID, "Test Shop", "test-shop", "https://test.com", true)
	require.NoError(t, err)

	product := &models.Product{
		ClientID:    clientID,
		ShopID:      shopID,
		Code:        "UPD-001",
		Name:        "Original Name",
		Description: "Original Description",
		IsActive:    true,
	}
	err = repo.Create(context.Background(), product)
	require.NoError(t, err)

	t.Run("successfully updates product", func(t *testing.T) {
		product.Name = "Updated Name"
		product.Description = "Updated Description"
		product.Metadata = map[string]interface{}{"updated": true}

		err := repo.Update(context.Background(), product)
		require.NoError(t, err)

		// Verify update
		found, err := repo.GetByID(context.Background(), clientID, product.ID)
		require.NoError(t, err)
		assert.Equal(t, "Updated Name", found.Name)
		assert.Equal(t, "Updated Description", found.Description)
	})

	t.Run("returns error for non-existent product", func(t *testing.T) {
		nonExistent := &models.Product{
			ClientID: clientID,
			ID:       uuid.New(),
			Name:     "Test",
			IsActive: true,
		}
		err := repo.Update(context.Background(), nonExistent)
		assert.Error(t, err)
	})
}

func TestProductRepository_Delete(t *testing.T) {
	pool := SetupTestDB(t)
	defer CleanupTestDB(t, pool)

	TruncateTable(t, pool, "product_variants")
	TruncateTable(t, pool, "products")
	TruncateTable(t, pool, "shops")
	TruncateTable(t, pool, "clients")

	repo := NewProductRepository(pool)

	// Create test client
	clientID := uuid.New()
	_, err := pool.Exec(context.Background(),
		`INSERT INTO clients (id, name, slug, is_active) VALUES ($1, $2, $3, $4)`,
		clientID, "Test Client", "test-client", true)
	require.NoError(t, err)

	// Create test shop
	shopID := uuid.New()
	_, err = pool.Exec(context.Background(),
		`INSERT INTO shops (id, client_id, name, slug, webpage_url, is_active) VALUES ($1, $2, $3, $4, $5, $6)`,
		shopID, clientID, "Test Shop", "test-shop", "https://test.com", true)
	require.NoError(t, err)

	product := &models.Product{
		ClientID:    clientID,
		ShopID:      shopID,
		Code:        "DEL-001",
		Name:        "To Delete",
		Description: "Description",
		IsActive:    true,
	}
	err = repo.Create(context.Background(), product)
	require.NoError(t, err)

	t.Run("successfully deletes product", func(t *testing.T) {
		err := repo.Delete(context.Background(), clientID, product.ID)
		require.NoError(t, err)

		// Verify deletion
		_, err = repo.GetByID(context.Background(), clientID, product.ID)
		assert.Error(t, err)
	})

	t.Run("returns error for non-existent product", func(t *testing.T) {
		err := repo.Delete(context.Background(), clientID, uuid.New())
		assert.Error(t, err)
	})
}

func TestProductRepository_ListByShop(t *testing.T) {
	pool := SetupTestDB(t)
	defer CleanupTestDB(t, pool)

	TruncateTable(t, pool, "product_variants")
	TruncateTable(t, pool, "products")
	TruncateTable(t, pool, "shops")
	TruncateTable(t, pool, "clients")

	repo := NewProductRepository(pool)

	// Create test client
	clientID := uuid.New()
	_, err := pool.Exec(context.Background(),
		`INSERT INTO clients (id, name, slug, is_active) VALUES ($1, $2, $3, $4)`,
		clientID, "Test Client", "test-client", true)
	require.NoError(t, err)

	// Create test shops
	shop1 := uuid.New()
	_, err = pool.Exec(context.Background(),
		`INSERT INTO shops (id, client_id, name, slug, webpage_url, is_active) VALUES ($1, $2, $3, $4, $5, $6)`,
		shop1, clientID, "Shop 1", "shop-1", "https://shop1.com", true)
	require.NoError(t, err)

	shop2 := uuid.New()
	_, err = pool.Exec(context.Background(),
		`INSERT INTO shops (id, client_id, name, slug, webpage_url, is_active) VALUES ($1, $2, $3, $4, $5, $6)`,
		shop2, clientID, "Shop 2", "shop-2", "https://shop2.com", true)
	require.NoError(t, err)

	// Create products for shop 1
	for i := 1; i <= 3; i++ {
		product := &models.Product{
			ClientID:    clientID,
			ShopID:      shop1,
			Code:        uuid.New().String(),
			Name:        "Shop1 Product",
			Description: "Description",
			IsActive:    true,
		}
		err := repo.Create(context.Background(), product)
		require.NoError(t, err)
	}

	// Create products for shop 2
	for i := 1; i <= 2; i++ {
		product := &models.Product{
			ClientID:    clientID,
			ShopID:      shop2,
			Code:        uuid.New().String(),
			Name:        "Shop2 Product",
			Description: "Description",
			IsActive:    true,
		}
		err := repo.Create(context.Background(), product)
		require.NoError(t, err)
	}

	t.Run("successfully lists products for shop 1", func(t *testing.T) {
		products, err := repo.ListByShop(context.Background(), clientID, shop1, 10, 0)
		require.NoError(t, err)
		assert.Len(t, products, 3)
	})

	t.Run("successfully lists products for shop 2", func(t *testing.T) {
		products, err := repo.ListByShop(context.Background(), clientID, shop2, 10, 0)
		require.NoError(t, err)
		assert.Len(t, products, 2)
	})

	t.Run("respects limit and offset", func(t *testing.T) {
		products, err := repo.ListByShop(context.Background(), clientID, shop1, 2, 0)
		require.NoError(t, err)
		assert.Len(t, products, 2)

		products, err = repo.ListByShop(context.Background(), clientID, shop1, 2, 2)
		require.NoError(t, err)
		assert.Len(t, products, 1)
	})
}

func TestProductRepository_ListByClient(t *testing.T) {
	pool := SetupTestDB(t)
	defer CleanupTestDB(t, pool)

	TruncateTable(t, pool, "product_variants")
	TruncateTable(t, pool, "products")
	TruncateTable(t, pool, "shops")
	TruncateTable(t, pool, "clients")

	repo := NewProductRepository(pool)

	// Create test clients
	client1 := uuid.New()
	_, err := pool.Exec(context.Background(),
		`INSERT INTO clients (id, name, slug, is_active) VALUES ($1, $2, $3, $4)`,
		client1, "Client 1", "client-1", true)
	require.NoError(t, err)

	client2 := uuid.New()
	_, err = pool.Exec(context.Background(),
		`INSERT INTO clients (id, name, slug, is_active) VALUES ($1, $2, $3, $4)`,
		client2, "Client 2", "client-2", true)
	require.NoError(t, err)

	// Create test shops
	shop1 := uuid.New()
	_, err = pool.Exec(context.Background(),
		`INSERT INTO shops (id, client_id, name, slug, webpage_url, is_active) VALUES ($1, $2, $3, $4, $5, $6)`,
		shop1, client1, "Shop 1", "shop-1-listby", "https://shop1.com", true)
	require.NoError(t, err)

	shop2 := uuid.New()
	_, err = pool.Exec(context.Background(),
		`INSERT INTO shops (id, client_id, name, slug, webpage_url, is_active) VALUES ($1, $2, $3, $4, $5, $6)`,
		shop2, client2, "Shop 2", "shop-2-listby", "https://shop2.com", true)
	require.NoError(t, err)

	// Create products for client 1
	for i := 1; i <= 4; i++ {
		product := &models.Product{
			ClientID:    client1,
			ShopID:      shop1,
			Code:        uuid.New().String(),
			Name:        "Client1 Product",
			Description: "Description",
			IsActive:    true,
		}
		err := repo.Create(context.Background(), product)
		require.NoError(t, err)
	}

	// Create products for client 2
	for i := 1; i <= 2; i++ {
		product := &models.Product{
			ClientID:    client2,
			ShopID:      shop2,
			Code:        uuid.New().String(),
			Name:        "Client2 Product",
			Description: "Description",
			IsActive:    true,
		}
		err := repo.Create(context.Background(), product)
		require.NoError(t, err)
	}

	t.Run("successfully lists all products for client", func(t *testing.T) {
		products, err := repo.ListByClient(context.Background(), client1, 10, 0)
		require.NoError(t, err)
		assert.Len(t, products, 4)
	})

	t.Run("respects limit and offset", func(t *testing.T) {
		products, err := repo.ListByClient(context.Background(), client1, 2, 0)
		require.NoError(t, err)
		assert.Len(t, products, 2)

		products, err = repo.ListByClient(context.Background(), client1, 2, 2)
		require.NoError(t, err)
		assert.Len(t, products, 2)
	})
}

func TestProductRepository_Search(t *testing.T) {
	pool := SetupTestDB(t)
	defer CleanupTestDB(t, pool)

	TruncateTable(t, pool, "product_variants")
	TruncateTable(t, pool, "products")
	TruncateTable(t, pool, "shops")
	TruncateTable(t, pool, "clients")

	repo := NewProductRepository(pool)

	// Create test client
	clientID := uuid.New()
	_, err := pool.Exec(context.Background(),
		`INSERT INTO clients (id, name, slug, is_active) VALUES ($1, $2, $3, $4)`,
		clientID, "Test Client", "test-client", true)
	require.NoError(t, err)

	// Create test shop
	shopID := uuid.New()
	_, err = pool.Exec(context.Background(),
		`INSERT INTO shops (id, client_id, name, slug, webpage_url, is_active) VALUES ($1, $2, $3, $4, $5, $6)`,
		shopID, clientID, "Test Shop", "test-shop", "https://test.com", true)
	require.NoError(t, err)

	// Create test products
	products := []*models.Product{
		{
			ClientID:    clientID,
			ShopID:      shopID,
			Code:        "LAPTOP-001",
			Name:        "Dell Laptop",
			Description: "High performance laptop",
			IsActive:    true,
		},
		{
			ClientID:    clientID,
			ShopID:      shopID,
			Code:        "MOUSE-001",
			Name:        "Wireless Mouse",
			Description: "Ergonomic mouse for laptop",
			IsActive:    true,
		},
		{
			ClientID:    clientID,
			ShopID:      shopID,
			Code:        "DESK-001",
			Name:        "Standing Desk",
			Description: "Adjustable height desk",
			IsActive:    true,
		},
	}

	for _, p := range products {
		err := repo.Create(context.Background(), p)
		require.NoError(t, err)
	}

	t.Run("searches by name", func(t *testing.T) {
		results, err := repo.Search(context.Background(), clientID, "Dell", 10, 0)
		require.NoError(t, err)
		assert.Len(t, results, 1)
		assert.Equal(t, "Dell Laptop", results[0].Name)
	})

	t.Run("searches by description", func(t *testing.T) {
		results, err := repo.Search(context.Background(), clientID, "laptop", 10, 0)
		require.NoError(t, err)
		assert.Len(t, results, 2) // Both laptop and mouse have "laptop" in description
	})

	t.Run("searches by code", func(t *testing.T) {
		results, err := repo.Search(context.Background(), clientID, "MOUSE", 10, 0)
		require.NoError(t, err)
		assert.Len(t, results, 1)
		assert.Equal(t, "MOUSE-001", results[0].Code)
	})

	t.Run("returns empty for no matches", func(t *testing.T) {
		results, err := repo.Search(context.Background(), clientID, "nonexistent", 10, 0)
		require.NoError(t, err)
		assert.Empty(t, results)
	})
}

func TestProductRepository_CreateVariant(t *testing.T) {
	pool := SetupTestDB(t)
	defer CleanupTestDB(t, pool)

	TruncateTable(t, pool, "product_variants")
	TruncateTable(t, pool, "products")
	TruncateTable(t, pool, "shops")
	TruncateTable(t, pool, "clients")

	repo := NewProductRepository(pool)

	// Create test client
	clientID := uuid.New()
	_, err := pool.Exec(context.Background(),
		`INSERT INTO clients (id, name, slug, is_active) VALUES ($1, $2, $3, $4)`,
		clientID, "Test Client", "test-client", true)
	require.NoError(t, err)

	// Create test shop
	shopID := uuid.New()
	_, err = pool.Exec(context.Background(),
		`INSERT INTO shops (id, client_id, name, slug, webpage_url, is_active) VALUES ($1, $2, $3, $4, $5, $6)`,
		shopID, clientID, "Test Shop", "test-shop", "https://test.com", true)
	require.NoError(t, err)

	product := &models.Product{
		ClientID:    clientID,
		ShopID:      shopID,
		Code:        "PROD-VAR",
		Name:        "Test Product",
		Description: "Description",
		IsActive:    true,
	}
	err = repo.Create(context.Background(), product)
	require.NoError(t, err)

	t.Run("successfully creates a variant", func(t *testing.T) {
		variant := &models.ProductVariant{
			ClientID:         clientID,
			ShopID:           shopID,
			ProductID:        product.ID,
			SKU:              "VAR-001",
			Name:             "Default Variant",
			Price:            99.99,
			Quantity:         100,
			RequiresShipping: true,
			IsDefault:        true,
			IsActive:         true,
		}

		err := repo.CreateVariant(context.Background(), variant)
		require.NoError(t, err)
		assert.NotEqual(t, uuid.Nil, variant.ID)
		assert.NotZero(t, variant.CreatedAt)
		assert.NotZero(t, variant.UpdatedAt)
	})

	t.Run("fails with duplicate SKU within same client", func(t *testing.T) {
		variant1 := &models.ProductVariant{
			ClientID:  clientID,
			ShopID:    shopID,
			ProductID: product.ID,
			SKU:       "UNIQUE-SKU",
			Name:      "Variant One",
			Price:     50.0,
			Quantity:  10,
			IsActive:  true,
		}
		err := repo.CreateVariant(context.Background(), variant1)
		require.NoError(t, err)

		variant2 := &models.ProductVariant{
			ClientID:  clientID,
			ShopID:    shopID,
			ProductID: product.ID,
			SKU:       "UNIQUE-SKU", // Same SKU, same client
			Name:      "Variant Two",
			Price:     60.0,
			Quantity:  20,
			IsActive:  true,
		}
		err = repo.CreateVariant(context.Background(), variant2)
		assert.Error(t, err)
	})
}

func TestProductRepository_GetVariantByID(t *testing.T) {
	pool := SetupTestDB(t)
	defer CleanupTestDB(t, pool)

	TruncateTable(t, pool, "product_variants")
	TruncateTable(t, pool, "products")
	TruncateTable(t, pool, "shops")
	TruncateTable(t, pool, "clients")

	repo := NewProductRepository(pool)

	// Create test client
	clientID := uuid.New()
	_, err := pool.Exec(context.Background(),
		`INSERT INTO clients (id, name, slug, is_active) VALUES ($1, $2, $3, $4)`,
		clientID, "Test Client", "test-client", true)
	require.NoError(t, err)

	// Create test shop
	shopID := uuid.New()
	_, err = pool.Exec(context.Background(),
		`INSERT INTO shops (id, client_id, name, slug, webpage_url, is_active) VALUES ($1, $2, $3, $4, $5, $6)`,
		shopID, clientID, "Test Shop", "test-shop", "https://test.com", true)
	require.NoError(t, err)

	product := &models.Product{
		ClientID:    clientID,
		ShopID:      shopID,
		Code:        "PROD-001",
		Name:        "Test Product",
		Description: "Description",
		IsActive:    true,
	}
	err = repo.Create(context.Background(), product)
	require.NoError(t, err)

	variant := &models.ProductVariant{
		ClientID:  clientID,
		ShopID:    shopID,
		ProductID: product.ID,
		SKU:       "VAR-GET",
		Name:      "Test Variant",
		Price:     49.99,
		Quantity:  50,
		IsActive:  true,
	}
	err = repo.CreateVariant(context.Background(), variant)
	require.NoError(t, err)

	t.Run("successfully retrieves variant by composite key", func(t *testing.T) {
		found, err := repo.GetVariantByID(context.Background(), clientID, variant.ID)
		require.NoError(t, err)
		assert.Equal(t, variant.ID, found.ID)
		assert.Equal(t, variant.SKU, found.SKU)
		assert.Equal(t, variant.Price, found.Price)
	})

	t.Run("returns error for non-existent ID", func(t *testing.T) {
		_, err := repo.GetVariantByID(context.Background(), clientID, uuid.New())
		assert.Error(t, err)
	})

	t.Run("returns error for wrong client", func(t *testing.T) {
		wrongClient := uuid.New()
		_, err := repo.GetVariantByID(context.Background(), wrongClient, variant.ID)
		assert.Error(t, err)
	})
}

func TestProductRepository_GetVariantBySKU(t *testing.T) {
	pool := SetupTestDB(t)
	defer CleanupTestDB(t, pool)

	TruncateTable(t, pool, "product_variants")
	TruncateTable(t, pool, "products")
	TruncateTable(t, pool, "shops")
	TruncateTable(t, pool, "clients")

	repo := NewProductRepository(pool)

	// Create test client
	clientID := uuid.New()
	_, err := pool.Exec(context.Background(),
		`INSERT INTO clients (id, name, slug, is_active) VALUES ($1, $2, $3, $4)`,
		clientID, "Test Client", "test-client", true)
	require.NoError(t, err)

	// Create test shop
	shopID := uuid.New()
	_, err = pool.Exec(context.Background(),
		`INSERT INTO shops (id, client_id, name, slug, webpage_url, is_active) VALUES ($1, $2, $3, $4, $5, $6)`,
		shopID, clientID, "Test Shop", "test-shop", "https://test.com", true)
	require.NoError(t, err)

	product := &models.Product{
		ClientID:    clientID,
		ShopID:      shopID,
		Code:        "PROD-001",
		Name:        "Test Product",
		Description: "Description",
		IsActive:    true,
	}
	err = repo.Create(context.Background(), product)
	require.NoError(t, err)

	variant := &models.ProductVariant{
		ClientID:  clientID,
		ShopID:    shopID,
		ProductID: product.ID,
		SKU:       "SKU-123",
		Name:      "Test Variant",
		Price:     29.99,
		Quantity:  75,
		IsActive:  true,
	}
	err = repo.CreateVariant(context.Background(), variant)
	require.NoError(t, err)

	t.Run("successfully retrieves variant by SKU", func(t *testing.T) {
		found, err := repo.GetVariantBySKU(context.Background(), clientID, "SKU-123")
		require.NoError(t, err)
		assert.Equal(t, variant.ID, found.ID)
		assert.Equal(t, "SKU-123", found.SKU)
	})

	t.Run("returns error for non-existent SKU", func(t *testing.T) {
		_, err := repo.GetVariantBySKU(context.Background(), clientID, "NONEXISTENT")
		assert.Error(t, err)
	})

	t.Run("returns error for wrong client", func(t *testing.T) {
		wrongClient := uuid.New()
		_, err := repo.GetVariantBySKU(context.Background(), wrongClient, "SKU-123")
		assert.Error(t, err)
	})
}

func TestProductRepository_UpdateVariant(t *testing.T) {
	pool := SetupTestDB(t)
	defer CleanupTestDB(t, pool)

	TruncateTable(t, pool, "product_variants")
	TruncateTable(t, pool, "products")
	TruncateTable(t, pool, "shops")
	TruncateTable(t, pool, "clients")

	repo := NewProductRepository(pool)

	// Create test client
	clientID := uuid.New()
	_, err := pool.Exec(context.Background(),
		`INSERT INTO clients (id, name, slug, is_active) VALUES ($1, $2, $3, $4)`,
		clientID, "Test Client", "test-client", true)
	require.NoError(t, err)

	// Create test shop
	shopID := uuid.New()
	_, err = pool.Exec(context.Background(),
		`INSERT INTO shops (id, client_id, name, slug, webpage_url, is_active) VALUES ($1, $2, $3, $4, $5, $6)`,
		shopID, clientID, "Test Shop", "test-shop", "https://test.com", true)
	require.NoError(t, err)

	product := &models.Product{
		ClientID:    clientID,
		ShopID:      shopID,
		Code:        "PROD-001",
		Name:        "Test Product",
		Description: "Description",
		IsActive:    true,
	}
	err = repo.Create(context.Background(), product)
	require.NoError(t, err)

	variant := &models.ProductVariant{
		ClientID:  clientID,
		ShopID:    shopID,
		ProductID: product.ID,
		SKU:       "UPD-VAR",
		Name:      "Original Name",
		Price:     100.0,
		Quantity:  50,
		IsActive:  true,
	}
	err = repo.CreateVariant(context.Background(), variant)
	require.NoError(t, err)

	t.Run("successfully updates variant", func(t *testing.T) {
		variant.Name = "Updated Name"
		variant.Price = 150.0
		variant.Quantity = 75

		err := repo.UpdateVariant(context.Background(), variant)
		require.NoError(t, err)

		// Verify update
		found, err := repo.GetVariantByID(context.Background(), clientID, variant.ID)
		require.NoError(t, err)
		assert.Equal(t, "Updated Name", found.Name)
		assert.Equal(t, 150.0, found.Price)
		assert.Equal(t, 75, found.Quantity)
	})

	t.Run("returns error for non-existent variant", func(t *testing.T) {
		nonExistent := &models.ProductVariant{
			ClientID: clientID,
			ID:       uuid.New(),
			Name:     "Test",
			Price:    50.0,
			IsActive: true,
		}
		err := repo.UpdateVariant(context.Background(), nonExistent)
		assert.Error(t, err)
	})
}

func TestProductRepository_DeleteVariant(t *testing.T) {
	pool := SetupTestDB(t)
	defer CleanupTestDB(t, pool)

	TruncateTable(t, pool, "product_variants")
	TruncateTable(t, pool, "products")
	TruncateTable(t, pool, "shops")
	TruncateTable(t, pool, "clients")

	repo := NewProductRepository(pool)

	// Create test client
	clientID := uuid.New()
	_, err := pool.Exec(context.Background(),
		`INSERT INTO clients (id, name, slug, is_active) VALUES ($1, $2, $3, $4)`,
		clientID, "Test Client", "test-client", true)
	require.NoError(t, err)

	// Create test shop
	shopID := uuid.New()
	_, err = pool.Exec(context.Background(),
		`INSERT INTO shops (id, client_id, name, slug, webpage_url, is_active) VALUES ($1, $2, $3, $4, $5, $6)`,
		shopID, clientID, "Test Shop", "test-shop", "https://test.com", true)
	require.NoError(t, err)

	product := &models.Product{
		ClientID:    clientID,
		ShopID:      shopID,
		Code:        "PROD-001",
		Name:        "Test Product",
		Description: "Description",
		IsActive:    true,
	}
	err = repo.Create(context.Background(), product)
	require.NoError(t, err)

	variant := &models.ProductVariant{
		ClientID:  clientID,
		ShopID:    shopID,
		ProductID: product.ID,
		SKU:       "DEL-VAR",
		Name:      "To Delete",
		Price:     50.0,
		Quantity:  10,
		IsActive:  true,
	}
	err = repo.CreateVariant(context.Background(), variant)
	require.NoError(t, err)

	t.Run("successfully deletes variant", func(t *testing.T) {
		err := repo.DeleteVariant(context.Background(), clientID, variant.ID)
		require.NoError(t, err)

		// Verify deletion
		_, err = repo.GetVariantByID(context.Background(), clientID, variant.ID)
		assert.Error(t, err)
	})

	t.Run("returns error for non-existent variant", func(t *testing.T) {
		err := repo.DeleteVariant(context.Background(), clientID, uuid.New())
		assert.Error(t, err)
	})
}

func TestProductRepository_ListVariantsByProduct(t *testing.T) {
	pool := SetupTestDB(t)
	defer CleanupTestDB(t, pool)

	TruncateTable(t, pool, "product_variants")
	TruncateTable(t, pool, "products")
	TruncateTable(t, pool, "shops")
	TruncateTable(t, pool, "clients")

	repo := NewProductRepository(pool)

	// Create test client
	clientID := uuid.New()
	_, err := pool.Exec(context.Background(),
		`INSERT INTO clients (id, name, slug, is_active) VALUES ($1, $2, $3, $4)`,
		clientID, "Test Client", "test-client", true)
	require.NoError(t, err)

	// Create test shop
	shopID := uuid.New()
	_, err = pool.Exec(context.Background(),
		`INSERT INTO shops (id, client_id, name, slug, webpage_url, is_active) VALUES ($1, $2, $3, $4, $5, $6)`,
		shopID, clientID, "Test Shop", "test-shop", "https://test.com", true)
	require.NoError(t, err)

	product := &models.Product{
		ClientID:    clientID,
		ShopID:      shopID,
		Code:        "PROD-001",
		Name:        "Test Product",
		Description: "Description",
		IsActive:    true,
	}
	err = repo.Create(context.Background(), product)
	require.NoError(t, err)

	// Create multiple variants
	for i := 1; i <= 3; i++ {
		variant := &models.ProductVariant{
			ClientID:  clientID,
			ShopID:    shopID,
			ProductID: product.ID,
			SKU:       uuid.New().String(),
			Name:      "Variant",
			Price:     float64(i * 10),
			Quantity:  i * 5,
			IsDefault: i == 1, // First one is default
			IsActive:  true,
		}
		err := repo.CreateVariant(context.Background(), variant)
		require.NoError(t, err)
	}

	t.Run("successfully lists all variants for product", func(t *testing.T) {
		variants, err := repo.ListVariantsByProduct(context.Background(), clientID, product.ID)
		require.NoError(t, err)
		assert.Len(t, variants, 3)
		// Default variant should be first
		assert.True(t, variants[0].IsDefault)
	})

	t.Run("returns empty list for product with no variants", func(t *testing.T) {
		newProduct := &models.Product{
			ClientID:    clientID,
			ShopID:      shopID,
			Code:        "NO-VARIANTS",
			Name:        "No Variants Product",
			Description: "Description",
			IsActive:    true,
		}
		err := repo.Create(context.Background(), newProduct)
		require.NoError(t, err)

		variants, err := repo.ListVariantsByProduct(context.Background(), clientID, newProduct.ID)
		require.NoError(t, err)
		assert.Empty(t, variants)
	})
}

func TestProductRepository_GetDefaultVariant(t *testing.T) {
	pool := SetupTestDB(t)
	defer CleanupTestDB(t, pool)

	TruncateTable(t, pool, "product_variants")
	TruncateTable(t, pool, "products")
	TruncateTable(t, pool, "shops")
	TruncateTable(t, pool, "clients")

	repo := NewProductRepository(pool)

	// Create test client
	clientID := uuid.New()
	_, err := pool.Exec(context.Background(),
		`INSERT INTO clients (id, name, slug, is_active) VALUES ($1, $2, $3, $4)`,
		clientID, "Test Client", "test-client", true)
	require.NoError(t, err)

	// Create test shop
	shopID := uuid.New()
	_, err = pool.Exec(context.Background(),
		`INSERT INTO shops (id, client_id, name, slug, webpage_url, is_active) VALUES ($1, $2, $3, $4, $5, $6)`,
		shopID, clientID, "Test Shop", "test-shop", "https://test.com", true)
	require.NoError(t, err)

	product := &models.Product{
		ClientID:    clientID,
		ShopID:      shopID,
		Code:        "PROD-001",
		Name:        "Test Product",
		Description: "Description",
		IsActive:    true,
	}
	err = repo.Create(context.Background(), product)
	require.NoError(t, err)

	// Create non-default variant
	variant1 := &models.ProductVariant{
		ClientID:  clientID,
		ShopID:    shopID,
		ProductID: product.ID,
		SKU:       "VAR-1",
		Name:      "Variant 1",
		Price:     50.0,
		Quantity:  10,
		IsDefault: false,
		IsActive:  true,
	}
	err = repo.CreateVariant(context.Background(), variant1)
	require.NoError(t, err)

	// Create default variant
	defaultVariant := &models.ProductVariant{
		ClientID:  clientID,
		ShopID:    shopID,
		ProductID: product.ID,
		SKU:       "VAR-DEFAULT",
		Name:      "Default Variant",
		Price:     100.0,
		Quantity:  20,
		IsDefault: true,
		IsActive:  true,
	}
	err = repo.CreateVariant(context.Background(), defaultVariant)
	require.NoError(t, err)

	t.Run("successfully retrieves default variant", func(t *testing.T) {
		found, err := repo.GetDefaultVariant(context.Background(), clientID, product.ID)
		require.NoError(t, err)
		assert.Equal(t, defaultVariant.ID, found.ID)
		assert.True(t, found.IsDefault)
		assert.Equal(t, "VAR-DEFAULT", found.SKU)
	})

	t.Run("returns error when no default variant exists", func(t *testing.T) {
		newProduct := &models.Product{
			ClientID:    clientID,
			ShopID:      shopID,
			Code:        "NO-DEFAULT",
			Name:        "No Default Product",
			Description: "Description",
			IsActive:    true,
		}
		err := repo.Create(context.Background(), newProduct)
		require.NoError(t, err)

		_, err = repo.GetDefaultVariant(context.Background(), clientID, newProduct.ID)
		assert.Error(t, err)
	})
}

func TestProductRepository_UpdateInventory(t *testing.T) {
	pool := SetupTestDB(t)
	defer CleanupTestDB(t, pool)

	TruncateTable(t, pool, "product_variants")
	TruncateTable(t, pool, "products")
	TruncateTable(t, pool, "shops")
	TruncateTable(t, pool, "clients")

	repo := NewProductRepository(pool)

	// Create test client
	clientID := uuid.New()
	_, err := pool.Exec(context.Background(),
		`INSERT INTO clients (id, name, slug, is_active) VALUES ($1, $2, $3, $4)`,
		clientID, "Test Client", "test-client", true)
	require.NoError(t, err)

	// Create test shop
	shopID := uuid.New()
	_, err = pool.Exec(context.Background(),
		`INSERT INTO shops (id, client_id, name, slug, webpage_url, is_active) VALUES ($1, $2, $3, $4, $5, $6)`,
		shopID, clientID, "Test Shop", "test-shop", "https://test.com", true)
	require.NoError(t, err)

	product := &models.Product{
		ClientID:    clientID,
		ShopID:      shopID,
		Code:        "PROD-001",
		Name:        "Test Product",
		Description: "Description",
		IsActive:    true,
	}
	err = repo.Create(context.Background(), product)
	require.NoError(t, err)

	variant := &models.ProductVariant{
		ClientID:  clientID,
		ShopID:    shopID,
		ProductID: product.ID,
		SKU:       "INV-VAR",
		Name:      "Inventory Variant",
		Price:     75.0,
		Quantity:  100,
		IsActive:  true,
	}
	err = repo.CreateVariant(context.Background(), variant)
	require.NoError(t, err)

	t.Run("successfully updates inventory", func(t *testing.T) {
		err := repo.UpdateInventory(context.Background(), clientID, variant.ID, 50)
		require.NoError(t, err)

		// Verify update
		found, err := repo.GetVariantByID(context.Background(), clientID, variant.ID)
		require.NoError(t, err)
		assert.Equal(t, 50, found.Quantity)
	})

	t.Run("can set inventory to zero", func(t *testing.T) {
		err := repo.UpdateInventory(context.Background(), clientID, variant.ID, 0)
		require.NoError(t, err)

		found, err := repo.GetVariantByID(context.Background(), clientID, variant.ID)
		require.NoError(t, err)
		assert.Equal(t, 0, found.Quantity)
	})

	t.Run("returns error for non-existent variant", func(t *testing.T) {
		err := repo.UpdateInventory(context.Background(), clientID, uuid.New(), 100)
		assert.Error(t, err)
	})
}
