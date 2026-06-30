package repositories

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yourorg/shoppilot/internal/models"
)

func setupInventoryAlertTestData(t *testing.T, pool *pgxpool.Pool) (*models.Client, *models.Shop, *models.Product, *models.ProductVariant) {
	// Create client
	clientRepo := NewClientRepository(pool)
	client := &models.Client{
		Name:         "Test Client",
		Slug:         "test-client-alert",
		Description:  "Test client",
		ContactEmail: "client@test.com",
		ContactPhone: "+1234567890",
		WebsiteURL:   "https://test.com",
		IsActive:     true,
	}
	err := clientRepo.Create(context.Background(), client)
	require.NoError(t, err)

	// Create shop
	shopRepo := NewShopRepository(pool)
	shop := &models.Shop{
		ClientID:   client.ID,
		Name:       "Test Shop",
		Slug:       "test-shop-alert",
		Address:    "123 Test St",
		City:       "Test City",
		State:      "TS",
		PostalCode: "12345",
		Country:    "USA",
		Phone:      "+1234567890",
		IsActive:   true,
	}
	err = shopRepo.Create(context.Background(), shop)
	require.NoError(t, err)

	// Create product
	productRepo := NewProductRepository(pool)
	product := &models.Product{
		ClientID:    client.ID,
		ShopID:      shop.ID,
		Name:        "Test Product",
		Code:        "PROD001",
		Description: "Test product",
		IsActive:    true,
	}
	err = productRepo.Create(context.Background(), product)
	require.NoError(t, err, "failed to create product")

	// Create variant
	variant := &models.ProductVariant{
		ClientID:  client.ID,
		ProductID: product.ID,
		ShopID:    shop.ID,
		SKU:       "SKU001",
		Name:      "Test Variant",
		Price:     100.00,
		Quantity:  50,
		IsActive:  true,
	}
	err = productRepo.CreateVariant(context.Background(), variant)
	require.NoError(t, err)

	return client, shop, product, variant
}

func TestInventoryAlertRepository_Create(t *testing.T) {
	pool := SetupTestDB(t)
	defer CleanupTestDB(t, pool)

	TruncateTable(t, pool, "inventory_alerts")
	TruncateTable(t, pool, "product_variants")
	TruncateTable(t, pool, "products")
	TruncateTable(t, pool, "shops")
	TruncateTable(t, pool, "clients")

	client, shop, _, variant := setupInventoryAlertTestData(t, pool)
	repo := NewInventoryAlertRepository(pool)

	t.Run("successfully creates an inventory alert", func(t *testing.T) {
		alert := &models.InventoryAlert{
			ClientID:          client.ID,
			VariantID:         variant.ID,
			ShopID:            shop.ID,
			LowStockThreshold: 10,
			IsEnabled:         true,
		}

		err := repo.Create(context.Background(), alert)
		require.NoError(t, err)
		assert.NotEqual(t, uuid.Nil, alert.ID)
		assert.NotZero(t, alert.CreatedAt)
		assert.NotZero(t, alert.UpdatedAt)
	})

	t.Run("fails with duplicate variant/shop combination", func(t *testing.T) {
		TruncateTable(t, pool, "inventory_alerts")

		alert1 := &models.InventoryAlert{
			ClientID:          client.ID,
			VariantID:         variant.ID,
			ShopID:            shop.ID,
			LowStockThreshold: 10,
			IsEnabled:         true,
		}
		err := repo.Create(context.Background(), alert1)
		require.NoError(t, err)

		alert2 := &models.InventoryAlert{
			ClientID:          client.ID,
			VariantID:         variant.ID,
			ShopID:            shop.ID,
			LowStockThreshold: 20,
			IsEnabled:         true,
		}
		err = repo.Create(context.Background(), alert2)
		assert.Error(t, err)
	})
}

func TestInventoryAlertRepository_GetByID(t *testing.T) {
	pool := SetupTestDB(t)
	defer CleanupTestDB(t, pool)

	TruncateTable(t, pool, "inventory_alerts")
	TruncateTable(t, pool, "product_variants")
	TruncateTable(t, pool, "products")
	TruncateTable(t, pool, "shops")
	TruncateTable(t, pool, "clients")

	client, shop, _, variant := setupInventoryAlertTestData(t, pool)
	repo := NewInventoryAlertRepository(pool)

	alert := &models.InventoryAlert{
		ClientID:          client.ID,
		VariantID:         variant.ID,
		ShopID:            shop.ID,
		LowStockThreshold: 10,
		IsEnabled:         true,
	}
	err := repo.Create(context.Background(), alert)
	require.NoError(t, err)

	t.Run("successfully retrieves alert by ID", func(t *testing.T) {
		found, err := repo.GetByID(context.Background(), client.ID, alert.ID)
		require.NoError(t, err)
		assert.Equal(t, alert.ID, found.ID)
		assert.Equal(t, alert.VariantID, found.VariantID)
		assert.Equal(t, alert.ShopID, found.ShopID)
		assert.Equal(t, alert.LowStockThreshold, found.LowStockThreshold)
	})

	t.Run("returns error for non-existent ID", func(t *testing.T) {
		_, err := repo.GetByID(context.Background(), client.ID, uuid.New())
		assert.Error(t, err)
	})
}

func TestInventoryAlertRepository_GetByVariant(t *testing.T) {
	pool := SetupTestDB(t)
	defer CleanupTestDB(t, pool)

	TruncateTable(t, pool, "inventory_alerts")
	TruncateTable(t, pool, "product_variants")
	TruncateTable(t, pool, "products")
	TruncateTable(t, pool, "shops")
	TruncateTable(t, pool, "clients")

	client, shop, _, variant := setupInventoryAlertTestData(t, pool)
	repo := NewInventoryAlertRepository(pool)

	alert := &models.InventoryAlert{
		ClientID:          client.ID,
		VariantID:         variant.ID,
		ShopID:            shop.ID,
		LowStockThreshold: 10,
		IsEnabled:         true,
	}
	err := repo.Create(context.Background(), alert)
	require.NoError(t, err)

	t.Run("successfully retrieves alert by variant", func(t *testing.T) {
		found, err := repo.GetByVariant(context.Background(), client.ID, variant.ID, shop.ID)
		require.NoError(t, err)
		assert.Equal(t, alert.ID, found.ID)
		assert.Equal(t, variant.ID, found.VariantID)
		assert.Equal(t, shop.ID, found.ShopID)
	})

	t.Run("returns error for non-existent variant/shop combination", func(t *testing.T) {
		_, err := repo.GetByVariant(context.Background(), client.ID, uuid.New(), shop.ID)
		assert.Error(t, err)
	})
}

func TestInventoryAlertRepository_Update(t *testing.T) {
	pool := SetupTestDB(t)
	defer CleanupTestDB(t, pool)

	TruncateTable(t, pool, "inventory_alerts")
	TruncateTable(t, pool, "product_variants")
	TruncateTable(t, pool, "products")
	TruncateTable(t, pool, "shops")
	TruncateTable(t, pool, "clients")

	client, shop, _, variant := setupInventoryAlertTestData(t, pool)
	repo := NewInventoryAlertRepository(pool)

	alert := &models.InventoryAlert{
		ClientID:          client.ID,
		VariantID:         variant.ID,
		ShopID:            shop.ID,
		LowStockThreshold: 10,
		IsEnabled:         true,
	}
	err := repo.Create(context.Background(), alert)
	require.NoError(t, err)

	t.Run("successfully updates alert threshold", func(t *testing.T) {
		alert.LowStockThreshold = 20
		alert.IsEnabled = false
		err := repo.Update(context.Background(), alert)
		require.NoError(t, err)

		found, err := repo.GetByID(context.Background(), client.ID, alert.ID)
		require.NoError(t, err)
		assert.Equal(t, 20, found.LowStockThreshold)
		assert.False(t, found.IsEnabled)
	})

	t.Run("returns error when updating non-existent alert", func(t *testing.T) {
		nonExistent := &models.InventoryAlert{
			ClientID:          client.ID,
			ID:                uuid.New(),
			VariantID:         variant.ID,
			ShopID:            shop.ID,
			LowStockThreshold: 15,
			IsEnabled:         true,
		}
		err := repo.Update(context.Background(), nonExistent)
		assert.Error(t, err)
	})
}

func TestInventoryAlertRepository_Delete(t *testing.T) {
	pool := SetupTestDB(t)
	defer CleanupTestDB(t, pool)

	TruncateTable(t, pool, "inventory_alerts")
	TruncateTable(t, pool, "product_variants")
	TruncateTable(t, pool, "products")
	TruncateTable(t, pool, "shops")
	TruncateTable(t, pool, "clients")

	client, shop, _, variant := setupInventoryAlertTestData(t, pool)
	repo := NewInventoryAlertRepository(pool)

	alert := &models.InventoryAlert{
		ClientID:          client.ID,
		VariantID:         variant.ID,
		ShopID:            shop.ID,
		LowStockThreshold: 10,
		IsEnabled:         true,
	}
	err := repo.Create(context.Background(), alert)
	require.NoError(t, err)

	t.Run("successfully deletes alert", func(t *testing.T) {
		err := repo.Delete(context.Background(), client.ID, alert.ID)
		require.NoError(t, err)

		_, err = repo.GetByID(context.Background(), client.ID, alert.ID)
		assert.Error(t, err)
	})

	t.Run("returns error when deleting non-existent alert", func(t *testing.T) {
		err := repo.Delete(context.Background(), client.ID, uuid.New())
		assert.Error(t, err)
	})
}

func TestInventoryAlertRepository_ListByShop(t *testing.T) {
	pool := SetupTestDB(t)
	defer CleanupTestDB(t, pool)

	TruncateTable(t, pool, "inventory_alerts")
	TruncateTable(t, pool, "product_variants")
	TruncateTable(t, pool, "products")
	TruncateTable(t, pool, "shops")
	TruncateTable(t, pool, "clients")

	client, shop, product, _ := setupInventoryAlertTestData(t, pool)
	repo := NewInventoryAlertRepository(pool)
	productRepo := NewProductRepository(pool)

	// Create multiple variants and alerts
	for i := 1; i <= 5; i++ {
		variant := &models.ProductVariant{
			ClientID:  client.ID,
			ProductID: product.ID,
			ShopID:    shop.ID,
			SKU:       uuid.New().String(),
			Name:      "Variant " + uuid.New().String(),
			Price:     100.00,
			Quantity:  50,
			IsActive:  true,
		}
		err := productRepo.CreateVariant(context.Background(), variant)
		require.NoError(t, err)

		alert := &models.InventoryAlert{
			ClientID:          client.ID,
			VariantID:         variant.ID,
			ShopID:            shop.ID,
			LowStockThreshold: i * 10,
			IsEnabled:         i%2 == 1,
		}
		err = repo.Create(context.Background(), alert)
		require.NoError(t, err)
	}

	t.Run("successfully lists alerts by shop", func(t *testing.T) {
		alerts, err := repo.ListByShop(context.Background(), client.ID, shop.ID, 10, 0)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(alerts), 5)
	})

	t.Run("respects limit parameter", func(t *testing.T) {
		alerts, err := repo.ListByShop(context.Background(), client.ID, shop.ID, 3, 0)
		require.NoError(t, err)
		assert.LessOrEqual(t, len(alerts), 3)
	})

	t.Run("respects offset parameter", func(t *testing.T) {
		allAlerts, err := repo.ListByShop(context.Background(), client.ID, shop.ID, 100, 0)
		require.NoError(t, err)

		if len(allAlerts) > 1 {
			offsetAlerts, err := repo.ListByShop(context.Background(), client.ID, shop.ID, 100, 1)
			require.NoError(t, err)
			assert.Less(t, len(offsetAlerts), len(allAlerts))
		}
	})
}

func TestInventoryAlertRepository_ListTriggered(t *testing.T) {
	pool := SetupTestDB(t)
	defer CleanupTestDB(t, pool)

	TruncateTable(t, pool, "inventory_alerts")
	TruncateTable(t, pool, "product_variants")
	TruncateTable(t, pool, "products")
	TruncateTable(t, pool, "shops")
	TruncateTable(t, pool, "clients")

	client, shop, product, _ := setupInventoryAlertTestData(t, pool)
	repo := NewInventoryAlertRepository(pool)
	productRepo := NewProductRepository(pool)

	// Create variants with different quantities
	testCases := []struct {
		quantity  int
		threshold int
		enabled   bool
		triggered bool // Should be in triggered list?
	}{
		{5, 10, true, true},   // quantity <= threshold, enabled -> TRIGGERED
		{15, 10, true, false}, // quantity > threshold -> NOT triggered
		{5, 10, false, false}, // quantity <= threshold but disabled -> NOT triggered
		{8, 10, true, true},   // quantity <= threshold, enabled -> TRIGGERED
		{12, 15, true, true},  // quantity <= threshold, enabled -> TRIGGERED
	}

	for i, tc := range testCases {
		variant := &models.ProductVariant{
			ClientID:  client.ID,
			ProductID: product.ID,
			ShopID:    shop.ID,
			SKU:       uuid.New().String(),
			Name:      "Variant " + uuid.New().String(),
			Price:     100.00,
			Quantity:  tc.quantity,
			IsActive:  true,
		}
		err := productRepo.CreateVariant(context.Background(), variant)
		require.NoError(t, err, "Failed to create variant %d", i)

		alert := &models.InventoryAlert{
			ClientID:          client.ID,
			VariantID:         variant.ID,
			ShopID:            shop.ID,
			LowStockThreshold: tc.threshold,
			IsEnabled:         tc.enabled,
		}
		err = repo.Create(context.Background(), alert)
		require.NoError(t, err, "Failed to create alert %d", i)
	}

	t.Run("successfully lists only triggered alerts", func(t *testing.T) {
		alerts, err := repo.ListTriggered(context.Background(), client.ID)
		require.NoError(t, err)

		// Should return exactly 3 alerts (cases where triggered=true)
		assert.Equal(t, 3, len(alerts))

		// Verify all returned alerts are enabled
		for _, alert := range alerts {
			assert.True(t, alert.IsEnabled, "Triggered alert should be enabled")
		}
	})
}
