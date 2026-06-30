package repositories

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yourorg/shoppilot/internal/models"
)

func setupPurchaseOrderTestData(t *testing.T, pool *pgxpool.Pool) (*models.Client, *models.Shop, *models.Supplier, *models.Product, *models.ProductVariant) {
	// Create client
	clientRepo := NewClientRepository(pool)
	client := &models.Client{
		Name:         "Test Client",
		Slug:         "test-client-po",
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
		Slug:       "test-shop-po",
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

	// Create supplier
	supplierRepo := NewSupplierRepository(pool)
	supplier := &models.Supplier{
		ClientID: client.ID,
		Code:     "SUP001",
		Name:     "Test Supplier",
		Email:    "supplier@test.com",
		IsActive: true,
	}
	err = supplierRepo.Create(context.Background(), supplier)
	require.NoError(t, err)

	// Create product
	productRepo := NewProductRepository(pool)
	product := &models.Product{
		ClientID:    client.ID,
		ShopID:      shop.ID,
		Name:        "Test Product",
		Code:        "PROD001",
		Description: "Test product for PO",
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

	return client, shop, supplier, product, variant
}

func TestPurchaseOrderRepository_Create(t *testing.T) {
	pool := SetupTestDB(t)
	defer CleanupTestDB(t, pool)

	TruncateTable(t, pool, "purchase_orders")
	TruncateTable(t, pool, "suppliers")
	TruncateTable(t, pool, "product_variants")
	TruncateTable(t, pool, "products")
	TruncateTable(t, pool, "shops")
	TruncateTable(t, pool, "clients")

	client, shop, supplier, _, _ := setupPurchaseOrderTestData(t, pool)
	repo := NewPurchaseOrderRepository(pool)

	t.Run("successfully creates a purchase order", func(t *testing.T) {
		po := &models.PurchaseOrder{
			ClientID:             client.ID,
			SupplierID:           supplier.ID,
			ShopID:               shop.ID,
			PONumber:             "PO-001",
			Status:               models.POStatusDraft,
			OrderDate:            time.Now(),
			ExpectedDeliveryDate: nil,
			TotalAmount:          1000.00,
			Currency:             "USD",
			Notes:                "Test PO",
		}

		err := repo.Create(context.Background(), po)
		require.NoError(t, err)
		assert.NotEqual(t, uuid.Nil, po.ID)
		assert.NotZero(t, po.CreatedAt)
		assert.NotZero(t, po.UpdatedAt)
	})

	t.Run("fails with duplicate PO number", func(t *testing.T) {
		po1 := &models.PurchaseOrder{
			ClientID:   client.ID,
			SupplierID: supplier.ID,
			ShopID:     shop.ID,
			PONumber:   "PO-UNIQUE",
			Status:     models.POStatusDraft,
			OrderDate:  time.Now(),
		}
		err := repo.Create(context.Background(), po1)
		require.NoError(t, err)

		po2 := &models.PurchaseOrder{
			ClientID:   client.ID,
			SupplierID: supplier.ID,
			ShopID:     shop.ID,
			PONumber:   "PO-UNIQUE", // Same PO number
			Status:     models.POStatusDraft,
			OrderDate:  time.Now(),
		}
		err = repo.Create(context.Background(), po2)
		assert.Error(t, err)
	})
}

func TestPurchaseOrderRepository_GetByID(t *testing.T) {
	pool := SetupTestDB(t)
	defer CleanupTestDB(t, pool)

	TruncateTable(t, pool, "purchase_orders")
	TruncateTable(t, pool, "suppliers")
	TruncateTable(t, pool, "shops")
	TruncateTable(t, pool, "clients")

	client, shop, supplier, _, _ := setupPurchaseOrderTestData(t, pool)
	repo := NewPurchaseOrderRepository(pool)

	po := &models.PurchaseOrder{
		ClientID:   client.ID,
		SupplierID: supplier.ID,
		ShopID:     shop.ID,
		PONumber:   "PO-001",
		Status:     models.POStatusDraft,
		OrderDate:  time.Now(),
	}
	err := repo.Create(context.Background(), po)
	require.NoError(t, err)

	t.Run("successfully retrieves purchase order by ID", func(t *testing.T) {
		found, err := repo.GetByID(context.Background(), client.ID, po.ID)
		require.NoError(t, err)
		assert.Equal(t, po.ID, found.ID)
		assert.Equal(t, po.PONumber, found.PONumber)
		assert.Equal(t, po.Status, found.Status)
	})

	t.Run("returns error for non-existent ID", func(t *testing.T) {
		_, err := repo.GetByID(context.Background(), client.ID, uuid.New())
		assert.Error(t, err)
	})
}

func TestPurchaseOrderRepository_GetByPONumber(t *testing.T) {
	pool := SetupTestDB(t)
	defer CleanupTestDB(t, pool)

	TruncateTable(t, pool, "purchase_orders")
	TruncateTable(t, pool, "suppliers")
	TruncateTable(t, pool, "shops")
	TruncateTable(t, pool, "clients")

	client, shop, supplier, _, _ := setupPurchaseOrderTestData(t, pool)
	repo := NewPurchaseOrderRepository(pool)

	po := &models.PurchaseOrder{
		ClientID:   client.ID,
		SupplierID: supplier.ID,
		ShopID:     shop.ID,
		PONumber:   "PO-SEARCH",
		Status:     models.POStatusDraft,
		OrderDate:  time.Now(),
	}
	err := repo.Create(context.Background(), po)
	require.NoError(t, err)

	t.Run("successfully retrieves purchase order by PO number", func(t *testing.T) {
		found, err := repo.GetByPONumber(context.Background(), client.ID, "PO-SEARCH")
		require.NoError(t, err)
		assert.Equal(t, po.ID, found.ID)
		assert.Equal(t, "PO-SEARCH", found.PONumber)
	})

	t.Run("returns error for non-existent PO number", func(t *testing.T) {
		_, err := repo.GetByPONumber(context.Background(), client.ID, "NONEXISTENT")
		assert.Error(t, err)
	})
}

func TestPurchaseOrderRepository_Update(t *testing.T) {
	pool := SetupTestDB(t)
	defer CleanupTestDB(t, pool)

	TruncateTable(t, pool, "purchase_orders")
	TruncateTable(t, pool, "suppliers")
	TruncateTable(t, pool, "shops")
	TruncateTable(t, pool, "clients")

	client, shop, supplier, _, _ := setupPurchaseOrderTestData(t, pool)
	repo := NewPurchaseOrderRepository(pool)

	po := &models.PurchaseOrder{
		ClientID:    client.ID,
		SupplierID:  supplier.ID,
		ShopID:      shop.ID,
		PONumber:    "PO-001",
		Status:      models.POStatusDraft,
		OrderDate:   time.Now(),
		TotalAmount: 1000.00,
	}
	err := repo.Create(context.Background(), po)
	require.NoError(t, err)

	t.Run("successfully updates purchase order", func(t *testing.T) {
		po.Status = models.POStatusSubmitted
		po.TotalAmount = 1500.00
		po.Notes = "Updated notes"

		err := repo.Update(context.Background(), po)
		require.NoError(t, err)

		found, err := repo.GetByID(context.Background(), client.ID, po.ID)
		require.NoError(t, err)
		assert.Equal(t, models.POStatusSubmitted, found.Status)
		assert.Equal(t, 1500.00, found.TotalAmount)
		assert.Equal(t, "Updated notes", found.Notes)
	})
}

func TestPurchaseOrderRepository_Delete(t *testing.T) {
	pool := SetupTestDB(t)
	defer CleanupTestDB(t, pool)

	TruncateTable(t, pool, "purchase_orders")
	TruncateTable(t, pool, "suppliers")
	TruncateTable(t, pool, "shops")
	TruncateTable(t, pool, "clients")

	client, shop, supplier, _, _ := setupPurchaseOrderTestData(t, pool)
	repo := NewPurchaseOrderRepository(pool)

	po := &models.PurchaseOrder{
		ClientID:   client.ID,
		SupplierID: supplier.ID,
		ShopID:     shop.ID,
		PONumber:   "PO-DELETE",
		Status:     models.POStatusDraft,
		OrderDate:  time.Now(),
	}
	err := repo.Create(context.Background(), po)
	require.NoError(t, err)

	t.Run("successfully deletes purchase order", func(t *testing.T) {
		err := repo.Delete(context.Background(), client.ID, po.ID)
		require.NoError(t, err)

		_, err = repo.GetByID(context.Background(), client.ID, po.ID)
		assert.Error(t, err)
	})
}

func TestPurchaseOrderRepository_ListBySupplier(t *testing.T) {
	pool := SetupTestDB(t)
	defer CleanupTestDB(t, pool)

	TruncateTable(t, pool, "purchase_orders")
	TruncateTable(t, pool, "suppliers")
	TruncateTable(t, pool, "shops")
	TruncateTable(t, pool, "clients")

	client, shop, supplier, _, _ := setupPurchaseOrderTestData(t, pool)
	repo := NewPurchaseOrderRepository(pool)

	// Create multiple POs for this supplier
	for i := 1; i <= 3; i++ {
		po := &models.PurchaseOrder{
			ClientID:   client.ID,
			SupplierID: supplier.ID,
			ShopID:     shop.ID,
			PONumber:   fmt.Sprintf("PO-%03d", i),
			Status:     models.POStatusDraft,
			OrderDate:  time.Now(),
		}
		err := repo.Create(context.Background(), po)
		require.NoError(t, err)
	}

	t.Run("successfully lists purchase orders by supplier", func(t *testing.T) {
		orders, err := repo.ListBySupplier(context.Background(), client.ID, supplier.ID, 10, 0)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(orders), 3)
	})
}

func TestPurchaseOrderRepository_ListByStatus(t *testing.T) {
	pool := SetupTestDB(t)
	defer CleanupTestDB(t, pool)

	TruncateTable(t, pool, "purchase_orders")
	TruncateTable(t, pool, "suppliers")
	TruncateTable(t, pool, "shops")
	TruncateTable(t, pool, "clients")

	client, shop, supplier, _, _ := setupPurchaseOrderTestData(t, pool)
	repo := NewPurchaseOrderRepository(pool)

	// Create POs with different statuses
	for i := 1; i <= 5; i++ {
		status := models.POStatusDraft
		if i > 3 {
			status = models.POStatusSubmitted
		}

		po := &models.PurchaseOrder{
			ClientID:   client.ID,
			SupplierID: supplier.ID,
			ShopID:     shop.ID,
			PONumber:   fmt.Sprintf("PO-STATUS-%03d", i),
			Status:     status,
			OrderDate:  time.Now(),
		}
		err := repo.Create(context.Background(), po)
		require.NoError(t, err)
	}

	t.Run("successfully lists purchase orders by status", func(t *testing.T) {
		draftOrders, err := repo.ListByStatus(context.Background(), client.ID, models.POStatusDraft, 10, 0)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(draftOrders), 3)

		for _, po := range draftOrders {
			assert.Equal(t, models.POStatusDraft, po.Status)
		}
	})
}

func TestPurchaseOrderRepository_CreateItem(t *testing.T) {
	pool := SetupTestDB(t)
	defer CleanupTestDB(t, pool)

	TruncateTable(t, pool, "purchase_order_items")
	TruncateTable(t, pool, "purchase_orders")
	TruncateTable(t, pool, "suppliers")
	TruncateTable(t, pool, "product_variants")
	TruncateTable(t, pool, "products")
	TruncateTable(t, pool, "shops")
	TruncateTable(t, pool, "clients")

	client, shop, supplier, _, variant := setupPurchaseOrderTestData(t, pool)
	repo := NewPurchaseOrderRepository(pool)

	po := &models.PurchaseOrder{
		ClientID:   client.ID,
		SupplierID: supplier.ID,
		ShopID:     shop.ID,
		PONumber:   "PO-ITEMS",
		Status:     models.POStatusDraft,
		OrderDate:  time.Now(),
	}
	err := repo.Create(context.Background(), po)
	require.NoError(t, err)

	t.Run("successfully creates purchase order item", func(t *testing.T) {
		item := &models.PurchaseOrderItem{
			ClientID:        client.ID,
			PurchaseOrderID: po.ID,
			VariantID:       variant.ID,
			QuantityOrdered: 10,
			UnitCost:        50.00,
			TotalCost:       500.00,
			Notes:           "Test item",
		}

		err := repo.CreateItem(context.Background(), item)
		require.NoError(t, err)
		assert.NotEqual(t, uuid.Nil, item.ID)
		assert.Equal(t, 0, item.QuantityReceived)
		assert.NotZero(t, item.CreatedAt)
	})
}

func TestPurchaseOrderRepository_GetItem(t *testing.T) {
	pool := SetupTestDB(t)
	defer CleanupTestDB(t, pool)

	TruncateTable(t, pool, "purchase_order_items")
	TruncateTable(t, pool, "purchase_orders")
	TruncateTable(t, pool, "suppliers")
	TruncateTable(t, pool, "product_variants")
	TruncateTable(t, pool, "products")
	TruncateTable(t, pool, "shops")
	TruncateTable(t, pool, "clients")

	client, shop, supplier, _, variant := setupPurchaseOrderTestData(t, pool)
	repo := NewPurchaseOrderRepository(pool)

	po := &models.PurchaseOrder{
		ClientID:   client.ID,
		SupplierID: supplier.ID,
		ShopID:     shop.ID,
		PONumber:   "PO-001",
		Status:     models.POStatusDraft,
		OrderDate:  time.Now(),
	}
	err := repo.Create(context.Background(), po)
	require.NoError(t, err)

	item := &models.PurchaseOrderItem{
		ClientID:        client.ID,
		PurchaseOrderID: po.ID,
		VariantID:       variant.ID,
		QuantityOrdered: 20,
		UnitCost:        75.00,
		TotalCost:       1500.00,
	}
	err = repo.CreateItem(context.Background(), item)
	require.NoError(t, err)

	t.Run("successfully retrieves purchase order item", func(t *testing.T) {
		found, err := repo.GetItem(context.Background(), client.ID, item.ID)
		require.NoError(t, err)
		assert.Equal(t, item.ID, found.ID)
		assert.Equal(t, item.QuantityOrdered, found.QuantityOrdered)
		assert.Equal(t, item.UnitCost, found.UnitCost)
	})
}

func TestPurchaseOrderRepository_ListItemsByPO(t *testing.T) {
	pool := SetupTestDB(t)
	defer CleanupTestDB(t, pool)

	TruncateTable(t, pool, "purchase_order_items")
	TruncateTable(t, pool, "purchase_orders")
	TruncateTable(t, pool, "suppliers")
	TruncateTable(t, pool, "product_variants")
	TruncateTable(t, pool, "products")
	TruncateTable(t, pool, "shops")
	TruncateTable(t, pool, "clients")

	client, shop, supplier, _, variant := setupPurchaseOrderTestData(t, pool)
	repo := NewPurchaseOrderRepository(pool)

	po := &models.PurchaseOrder{
		ClientID:   client.ID,
		SupplierID: supplier.ID,
		ShopID:     shop.ID,
		PONumber:   "PO-LIST",
		Status:     models.POStatusDraft,
		OrderDate:  time.Now(),
	}
	err := repo.Create(context.Background(), po)
	require.NoError(t, err)

	// Create multiple items
	for i := 1; i <= 3; i++ {
		item := &models.PurchaseOrderItem{
			ClientID:        client.ID,
			PurchaseOrderID: po.ID,
			VariantID:       variant.ID,
			QuantityOrdered: i * 10,
			UnitCost:        50.00,
			TotalCost:       float64(i * 10 * 50),
		}
		err := repo.CreateItem(context.Background(), item)
		require.NoError(t, err)
	}

	t.Run("successfully lists all items for purchase order", func(t *testing.T) {
		items, err := repo.ListItemsByPO(context.Background(), client.ID, po.ID)
		require.NoError(t, err)
		assert.Equal(t, 3, len(items))
	})
}

func TestPurchaseOrderRepository_ReceiveItem(t *testing.T) {
	pool := SetupTestDB(t)
	defer CleanupTestDB(t, pool)

	TruncateTable(t, pool, "purchase_order_items")
	TruncateTable(t, pool, "purchase_orders")
	TruncateTable(t, pool, "suppliers")
	TruncateTable(t, pool, "product_variants")
	TruncateTable(t, pool, "products")
	TruncateTable(t, pool, "shops")
	TruncateTable(t, pool, "clients")

	client, shop, supplier, _, variant := setupPurchaseOrderTestData(t, pool)
	repo := NewPurchaseOrderRepository(pool)

	po := &models.PurchaseOrder{
		ClientID:   client.ID,
		SupplierID: supplier.ID,
		ShopID:     shop.ID,
		PONumber:   "PO-RECEIVE",
		Status:     models.POStatusSubmitted,
		OrderDate:  time.Now(),
	}
	err := repo.Create(context.Background(), po)
	require.NoError(t, err)

	item := &models.PurchaseOrderItem{
		ClientID:        client.ID,
		PurchaseOrderID: po.ID,
		VariantID:       variant.ID,
		QuantityOrdered: 100,
		UnitCost:        10.00,
		TotalCost:       1000.00,
	}
	err = repo.CreateItem(context.Background(), item)
	require.NoError(t, err)

	t.Run("successfully receives items", func(t *testing.T) {
		err := repo.ReceiveItem(context.Background(), client.ID, item.ID, 50)
		require.NoError(t, err)

		found, err := repo.GetItem(context.Background(), client.ID, item.ID)
		require.NoError(t, err)
		assert.Equal(t, 50, found.QuantityReceived)
	})

	t.Run("successfully receives additional items", func(t *testing.T) {
		err := repo.ReceiveItem(context.Background(), client.ID, item.ID, 30)
		require.NoError(t, err)

		found, err := repo.GetItem(context.Background(), client.ID, item.ID)
		require.NoError(t, err)
		assert.Equal(t, 80, found.QuantityReceived) // 50 + 30
	})

	t.Run("fails when exceeding ordered quantity", func(t *testing.T) {
		err := repo.ReceiveItem(context.Background(), client.ID, item.ID, 50)
		assert.Error(t, err) // Already received 80, can't receive 50 more (total would be 130, but ordered only 100)
	})
}
