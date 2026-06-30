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

func setupInventoryMovementTestData(t *testing.T, pool *pgxpool.Pool) (*models.Client, *models.Shop, *models.Product, *models.ProductVariant) {
	// Create client
	clientRepo := NewClientRepository(pool)
	client := &models.Client{
		Name:         "Test Client",
		Slug:         "test-client-movement",
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
		Slug:       "test-shop-movement",
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

func TestInventoryMovementRepository_Create(t *testing.T) {
	pool := SetupTestDB(t)
	defer CleanupTestDB(t, pool)

	TruncateTable(t, pool, "inventory_movements")
	TruncateTable(t, pool, "product_variants")
	TruncateTable(t, pool, "products")
	TruncateTable(t, pool, "shops")
	TruncateTable(t, pool, "clients")

	client, shop, _, variant := setupInventoryMovementTestData(t, pool)
	repo := NewInventoryMovementRepository(pool)

	t.Run("successfully creates an inventory movement", func(t *testing.T) {
		refID := uuid.New()
		performedBy := uuid.New()
		movement := &models.InventoryMovement{
			ClientID:         client.ID,
			VariantID:        variant.ID,
			ShopID:           shop.ID,
			MovementType:     models.MovementTypeAdjustment,
			Quantity:         10,
			PreviousQuantity: 50,
			NewQuantity:      60,
			ReferenceType:    "manual_adjustment",
			ReferenceID:      &refID,
			Notes:            "Test movement",
			PerformedBy:      &performedBy,
		}

		err := repo.Create(context.Background(), movement)
		require.NoError(t, err)
		assert.NotEqual(t, uuid.Nil, movement.ID)
		assert.NotZero(t, movement.CreatedAt)
	})

	t.Run("creates movement with different types", func(t *testing.T) {
		types := []models.InventoryMovementType{
			models.MovementTypeSale,
			models.MovementTypePurchase,
			models.MovementTypeReturnFromCustomer,
			models.MovementTypeTransfer,
			models.MovementTypeDamaged,
		}

		for _, movType := range types {
			refID := uuid.New()
			movement := &models.InventoryMovement{
				ClientID:         client.ID,
				VariantID:        variant.ID,
				ShopID:           shop.ID,
				MovementType:     movType,
				Quantity:         5,
				PreviousQuantity: 50,
				NewQuantity:      55,
				ReferenceType:    "test",
				ReferenceID:      &refID,
			}

			err := repo.Create(context.Background(), movement)
			require.NoError(t, err)
		}
	})
}

func TestInventoryMovementRepository_GetByID(t *testing.T) {
	pool := SetupTestDB(t)
	defer CleanupTestDB(t, pool)

	TruncateTable(t, pool, "inventory_movements")
	TruncateTable(t, pool, "product_variants")
	TruncateTable(t, pool, "products")
	TruncateTable(t, pool, "shops")
	TruncateTable(t, pool, "clients")

	client, shop, _, variant := setupInventoryMovementTestData(t, pool)
	repo := NewInventoryMovementRepository(pool)

	refID := uuid.New()
	movement := &models.InventoryMovement{
		ClientID:         client.ID,
		VariantID:        variant.ID,
		ShopID:           shop.ID,
		MovementType:     models.MovementTypeAdjustment,
		Quantity:         10,
		PreviousQuantity: 50,
		NewQuantity:      60,
		ReferenceType:    "test",
		ReferenceID:      &refID,
	}
	err := repo.Create(context.Background(), movement)
	require.NoError(t, err)

	t.Run("successfully retrieves movement by ID", func(t *testing.T) {
		found, err := repo.GetByID(context.Background(), client.ID, movement.ID)
		require.NoError(t, err)
		assert.Equal(t, movement.ID, found.ID)
		assert.Equal(t, movement.MovementType, found.MovementType)
		assert.Equal(t, movement.Quantity, found.Quantity)
	})

	t.Run("returns error for non-existent ID", func(t *testing.T) {
		_, err := repo.GetByID(context.Background(), client.ID, uuid.New())
		assert.Error(t, err)
	})
}

func TestInventoryMovementRepository_ListByVariant(t *testing.T) {
	pool := SetupTestDB(t)
	defer CleanupTestDB(t, pool)

	TruncateTable(t, pool, "inventory_movements")
	TruncateTable(t, pool, "product_variants")
	TruncateTable(t, pool, "products")
	TruncateTable(t, pool, "shops")
	TruncateTable(t, pool, "clients")

	client, shop, _, variant := setupInventoryMovementTestData(t, pool)
	repo := NewInventoryMovementRepository(pool)

	// Create multiple movements for this variant
	for i := 1; i <= 5; i++ {
		refID := uuid.New()
		movement := &models.InventoryMovement{
			ClientID:         client.ID,
			VariantID:        variant.ID,
			ShopID:           shop.ID,
			MovementType:     models.MovementTypeAdjustment,
			Quantity:         i * 10,
			PreviousQuantity: 50,
			NewQuantity:      50 + (i * 10),
			ReferenceType:    "test",
			ReferenceID:      &refID,
		}
		err := repo.Create(context.Background(), movement)
		require.NoError(t, err)
	}

	t.Run("successfully lists movements by variant", func(t *testing.T) {
		movements, err := repo.ListByVariant(context.Background(), client.ID, variant.ID, 10, 0)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(movements), 5)
	})

	t.Run("respects limit parameter", func(t *testing.T) {
		movements, err := repo.ListByVariant(context.Background(), client.ID, variant.ID, 3, 0)
		require.NoError(t, err)
		assert.LessOrEqual(t, len(movements), 3)
	})

	t.Run("respects offset parameter", func(t *testing.T) {
		allMovements, err := repo.ListByVariant(context.Background(), client.ID, variant.ID, 100, 0)
		require.NoError(t, err)

		if len(allMovements) > 1 {
			offsetMovements, err := repo.ListByVariant(context.Background(), client.ID, variant.ID, 100, 1)
			require.NoError(t, err)
			assert.Less(t, len(offsetMovements), len(allMovements))
		}
	})
}

func TestInventoryMovementRepository_ListByShop(t *testing.T) {
	pool := SetupTestDB(t)
	defer CleanupTestDB(t, pool)

	TruncateTable(t, pool, "inventory_movements")
	TruncateTable(t, pool, "product_variants")
	TruncateTable(t, pool, "products")
	TruncateTable(t, pool, "shops")
	TruncateTable(t, pool, "clients")

	client, shop, _, variant := setupInventoryMovementTestData(t, pool)
	repo := NewInventoryMovementRepository(pool)

	// Create movements for this shop
	for i := 1; i <= 3; i++ {
		refID := uuid.New()
		movement := &models.InventoryMovement{
			ClientID:         client.ID,
			VariantID:        variant.ID,
			ShopID:           shop.ID,
			MovementType:     models.MovementTypeAdjustment,
			Quantity:         i * 10,
			PreviousQuantity: 50,
			NewQuantity:      50 + (i * 10),
			ReferenceType:    "test",
			ReferenceID:      &refID,
		}
		err := repo.Create(context.Background(), movement)
		require.NoError(t, err)
	}

	t.Run("successfully lists movements by shop", func(t *testing.T) {
		movements, err := repo.ListByShop(context.Background(), client.ID, shop.ID, 10, 0)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(movements), 3)
	})
}

func TestInventoryMovementRepository_ListByType(t *testing.T) {
	pool := SetupTestDB(t)
	defer CleanupTestDB(t, pool)

	TruncateTable(t, pool, "inventory_movements")
	TruncateTable(t, pool, "product_variants")
	TruncateTable(t, pool, "products")
	TruncateTable(t, pool, "shops")
	TruncateTable(t, pool, "clients")

	client, shop, _, variant := setupInventoryMovementTestData(t, pool)
	repo := NewInventoryMovementRepository(pool)

	// Create movements with different types
	for i := 1; i <= 5; i++ {
		movType := models.MovementTypeAdjustment
		if i > 3 {
			movType = models.MovementTypeSale
		}

		refID := uuid.New()
		movement := &models.InventoryMovement{
			ClientID:         client.ID,
			VariantID:        variant.ID,
			ShopID:           shop.ID,
			MovementType:     movType,
			Quantity:         i * 10,
			PreviousQuantity: 50,
			NewQuantity:      50 + (i * 10),
			ReferenceType:    "test",
			ReferenceID:      &refID,
		}
		err := repo.Create(context.Background(), movement)
		require.NoError(t, err)
	}

	t.Run("successfully lists movements by type", func(t *testing.T) {
		adjustments, err := repo.ListByType(context.Background(), client.ID, models.MovementTypeAdjustment, 10, 0)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(adjustments), 3)

		for _, mov := range adjustments {
			assert.Equal(t, models.MovementTypeAdjustment, mov.MovementType)
		}

		sales, err := repo.ListByType(context.Background(), client.ID, models.MovementTypeSale, 10, 0)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(sales), 2)

		for _, mov := range sales {
			assert.Equal(t, models.MovementTypeSale, mov.MovementType)
		}
	})
}

func TestInventoryMovementRepository_ListByReference(t *testing.T) {
	pool := SetupTestDB(t)
	defer CleanupTestDB(t, pool)

	TruncateTable(t, pool, "inventory_movements")
	TruncateTable(t, pool, "product_variants")
	TruncateTable(t, pool, "products")
	TruncateTable(t, pool, "shops")
	TruncateTable(t, pool, "clients")

	client, shop, _, variant := setupInventoryMovementTestData(t, pool)
	repo := NewInventoryMovementRepository(pool)

	// Create a reference ID for a group of movements
	purchaseOrderID := uuid.New()
	salesOrderID := uuid.New()

	// Create movements with different reference types and IDs
	for i := 1; i <= 3; i++ {
		movement := &models.InventoryMovement{
			ClientID:         client.ID,
			VariantID:        variant.ID,
			ShopID:           shop.ID,
			MovementType:     models.MovementTypePurchase,
			Quantity:         i * 10,
			PreviousQuantity: 50,
			NewQuantity:      50 + (i * 10),
			ReferenceType:    "purchase_order",
			ReferenceID:      &purchaseOrderID,
		}
		err := repo.Create(context.Background(), movement)
		require.NoError(t, err)
	}

	for i := 1; i <= 2; i++ {
		movement := &models.InventoryMovement{
			ClientID:         client.ID,
			VariantID:        variant.ID,
			ShopID:           shop.ID,
			MovementType:     models.MovementTypeSale,
			Quantity:         -i * 5,
			PreviousQuantity: 80,
			NewQuantity:      80 - (i * 5),
			ReferenceType:    "sales_order",
			ReferenceID:      &salesOrderID,
		}
		err := repo.Create(context.Background(), movement)
		require.NoError(t, err)
	}

	t.Run("successfully lists movements by reference type and ID", func(t *testing.T) {
		poMovements, err := repo.ListByReference(context.Background(), client.ID, "purchase_order", purchaseOrderID)
		require.NoError(t, err)
		assert.Equal(t, 3, len(poMovements))

		for _, mov := range poMovements {
			assert.Equal(t, "purchase_order", mov.ReferenceType)
			assert.NotNil(t, mov.ReferenceID)
			assert.Equal(t, purchaseOrderID, *mov.ReferenceID)
		}

		soMovements, err := repo.ListByReference(context.Background(), client.ID, "sales_order", salesOrderID)
		require.NoError(t, err)
		assert.Equal(t, 2, len(soMovements))

		for _, mov := range soMovements {
			assert.Equal(t, "sales_order", mov.ReferenceType)
			assert.NotNil(t, mov.ReferenceID)
			assert.Equal(t, salesOrderID, *mov.ReferenceID)
		}
	})

	t.Run("returns empty list for non-existent reference", func(t *testing.T) {
		movements, err := repo.ListByReference(context.Background(), client.ID, "nonexistent", uuid.New())
		require.NoError(t, err)
		assert.Equal(t, 0, len(movements))
	})
}

func TestInventoryMovementRepository_AuditTrail(t *testing.T) {
	pool := SetupTestDB(t)
	defer CleanupTestDB(t, pool)

	TruncateTable(t, pool, "inventory_movements")
	TruncateTable(t, pool, "product_variants")
	TruncateTable(t, pool, "products")
	TruncateTable(t, pool, "shops")
	TruncateTable(t, pool, "clients")

	client, shop, _, variant := setupInventoryMovementTestData(t, pool)
	repo := NewInventoryMovementRepository(pool)

	// Simulate a sequence of inventory operations
	operations := []struct {
		movType  models.InventoryMovementType
		qty      int
		prevQty  int
		newQty   int
		refType  string
		notes    string
	}{
		{models.MovementTypePurchase, 100, 50, 150, "purchase_order", "Received from supplier"},
		{models.MovementTypeSale, -20, 150, 130, "sales_order", "Sold to customer"},
		{models.MovementTypeReturnFromCustomer, 5, 130, 135, "sales_order", "Customer return"},
		{models.MovementTypeAdjustment, -10, 135, 125, "manual_adjustment", "Damaged stock"},
		{models.MovementTypeDamaged, -3, 125, 122, "manual_adjustment", "Damaged stock removed"},
	}

	for _, op := range operations {
		refID := uuid.New()
		movement := &models.InventoryMovement{
			ClientID:         client.ID,
			VariantID:        variant.ID,
			ShopID:           shop.ID,
			MovementType:     op.movType,
			Quantity:         op.qty,
			PreviousQuantity: op.prevQty,
			NewQuantity:      op.newQty,
			ReferenceType:    op.refType,
			ReferenceID:      &refID,
			Notes:            op.notes,
		}
		err := repo.Create(context.Background(), movement)
		require.NoError(t, err)
	}

	t.Run("audit trail shows complete history", func(t *testing.T) {
		movements, err := repo.ListByVariant(context.Background(), client.ID, variant.ID, 100, 0)
		require.NoError(t, err)
		assert.Equal(t, 5, len(movements))

		// Verify the sequence is preserved (ordered by created_at DESC)
		// Most recent should be first
		assert.Equal(t, "Damaged stock removed", movements[0].Notes)
		assert.Equal(t, "Damaged stock", movements[1].Notes)
	})
}
