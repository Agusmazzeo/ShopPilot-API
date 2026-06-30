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

func setupSalesOrderTestData(t *testing.T, pool *pgxpool.Pool) (*models.Client, *models.Shop, *models.Customer, *models.Product, *models.ProductVariant) {
	// Create client
	clientRepo := NewClientRepository(pool)
	client := &models.Client{
		Name:         "Test Client",
		Slug:         "test-client-so",
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
		Slug:       "test-shop-so",
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

	// Create customer
	customerRepo := NewCustomerRepository(pool)
	customer := &models.Customer{
		ClientID:  client.ID,
		Code:      "CUST001",
		FirstName: "John",
		LastName:  "Doe",
		Email:     "john@test.com",
		IsActive:  true,
	}
	err = customerRepo.Create(context.Background(), customer)
	require.NoError(t, err)

	// Create product
	productRepo := NewProductRepository(pool)
	product := &models.Product{
		ClientID:    client.ID,
		ShopID:      shop.ID,
		Name:        "Test Product",
		Code:        "PROD001",
		Description: "Test product for SO",
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

	return client, shop, customer, product, variant
}

func TestSalesOrderRepository_Create(t *testing.T) {
	pool := SetupTestDB(t)
	defer CleanupTestDB(t, pool)

	TruncateTable(t, pool, "sales_orders")
	TruncateTable(t, pool, "customers")
	TruncateTable(t, pool, "product_variants")
	TruncateTable(t, pool, "products")
	TruncateTable(t, pool, "shops")
	TruncateTable(t, pool, "clients")

	client, shop, customer, _, _ := setupSalesOrderTestData(t, pool)
	repo := NewSalesOrderRepository(pool)

	t.Run("successfully creates a sales order", func(t *testing.T) {
		so := &models.SalesOrder{
			ClientID:     client.ID,
			CustomerID:   customer.ID,
			ShopID:       shop.ID,
			OrderNumber:  "SO-001",
			Status:       models.SOStatusPending,
			OrderDate:    time.Now(),
			Subtotal:     900.00,
			TaxAmount:    90.00,
			ShippingAmount: 10.00,
			TotalAmount:  1000.00,
			Currency:     "USD",
		}

		err := repo.Create(context.Background(), so)
		require.NoError(t, err)
		assert.NotEqual(t, uuid.Nil, so.ID)
		assert.NotZero(t, so.CreatedAt)
		assert.NotZero(t, so.UpdatedAt)
	})

	t.Run("fails with duplicate order number", func(t *testing.T) {
		so1 := &models.SalesOrder{
			ClientID:    client.ID,
			CustomerID:  customer.ID,
			ShopID:      shop.ID,
			OrderNumber: "SO-UNIQUE",
			Status:      models.SOStatusPending,
			OrderDate:   time.Now(),
		}
		err := repo.Create(context.Background(), so1)
		require.NoError(t, err)

		so2 := &models.SalesOrder{
			ClientID:    client.ID,
			CustomerID:  customer.ID,
			ShopID:      shop.ID,
			OrderNumber: "SO-UNIQUE", // Same order number
			Status:      models.SOStatusPending,
			OrderDate:   time.Now(),
		}
		err = repo.Create(context.Background(), so2)
		assert.Error(t, err)
	})
}

func TestSalesOrderRepository_GetByID(t *testing.T) {
	pool := SetupTestDB(t)
	defer CleanupTestDB(t, pool)

	TruncateTable(t, pool, "sales_orders")
	TruncateTable(t, pool, "customers")
	TruncateTable(t, pool, "shops")
	TruncateTable(t, pool, "clients")

	client, shop, customer, _, _ := setupSalesOrderTestData(t, pool)
	repo := NewSalesOrderRepository(pool)

	so := &models.SalesOrder{
		ClientID:    client.ID,
		CustomerID:  customer.ID,
		ShopID:      shop.ID,
		OrderNumber: "SO-001",
		Status:      models.SOStatusPending,
		OrderDate:   time.Now(),
	}
	err := repo.Create(context.Background(), so)
	require.NoError(t, err)

	t.Run("successfully retrieves sales order by ID", func(t *testing.T) {
		found, err := repo.GetByID(context.Background(), client.ID, so.ID)
		require.NoError(t, err)
		assert.Equal(t, so.ID, found.ID)
		assert.Equal(t, so.OrderNumber, found.OrderNumber)
		assert.Equal(t, so.Status, found.Status)
	})

	t.Run("returns error for non-existent ID", func(t *testing.T) {
		_, err := repo.GetByID(context.Background(), client.ID, uuid.New())
		assert.Error(t, err)
	})
}

func TestSalesOrderRepository_GetByOrderNumber(t *testing.T) {
	pool := SetupTestDB(t)
	defer CleanupTestDB(t, pool)

	TruncateTable(t, pool, "sales_orders")
	TruncateTable(t, pool, "customers")
	TruncateTable(t, pool, "shops")
	TruncateTable(t, pool, "clients")

	client, shop, customer, _, _ := setupSalesOrderTestData(t, pool)
	repo := NewSalesOrderRepository(pool)

	so := &models.SalesOrder{
		ClientID:    client.ID,
		CustomerID:  customer.ID,
		ShopID:      shop.ID,
		OrderNumber: "SO-SEARCH",
		Status:      models.SOStatusPending,
		OrderDate:   time.Now(),
	}
	err := repo.Create(context.Background(), so)
	require.NoError(t, err)

	t.Run("successfully retrieves sales order by order number", func(t *testing.T) {
		found, err := repo.GetByOrderNumber(context.Background(), client.ID, "SO-SEARCH")
		require.NoError(t, err)
		assert.Equal(t, so.ID, found.ID)
		assert.Equal(t, "SO-SEARCH", found.OrderNumber)
	})

	t.Run("returns error for non-existent order number", func(t *testing.T) {
		_, err := repo.GetByOrderNumber(context.Background(), client.ID, "NONEXISTENT")
		assert.Error(t, err)
	})
}

func TestSalesOrderRepository_Update(t *testing.T) {
	pool := SetupTestDB(t)
	defer CleanupTestDB(t, pool)

	TruncateTable(t, pool, "sales_orders")
	TruncateTable(t, pool, "customers")
	TruncateTable(t, pool, "shops")
	TruncateTable(t, pool, "clients")

	client, shop, customer, _, _ := setupSalesOrderTestData(t, pool)
	repo := NewSalesOrderRepository(pool)

	so := &models.SalesOrder{
		ClientID:    client.ID,
		CustomerID:  customer.ID,
		ShopID:      shop.ID,
		OrderNumber: "SO-001",
		Status:      models.SOStatusPending,
		OrderDate:   time.Now(),
		TotalAmount: 1000.00,
	}
	err := repo.Create(context.Background(), so)
	require.NoError(t, err)

	t.Run("successfully updates sales order", func(t *testing.T) {
		so.Status = models.SOStatusConfirmed
		so.TotalAmount = 1200.00
		so.Notes = "Updated notes"

		err := repo.Update(context.Background(), so)
		require.NoError(t, err)

		found, err := repo.GetByID(context.Background(), client.ID, so.ID)
		require.NoError(t, err)
		assert.Equal(t, models.SOStatusConfirmed, found.Status)
		assert.Equal(t, 1200.00, found.TotalAmount)
		assert.Equal(t, "Updated notes", found.Notes)
	})
}

func TestSalesOrderRepository_Delete(t *testing.T) {
	pool := SetupTestDB(t)
	defer CleanupTestDB(t, pool)

	TruncateTable(t, pool, "sales_orders")
	TruncateTable(t, pool, "customers")
	TruncateTable(t, pool, "shops")
	TruncateTable(t, pool, "clients")

	client, shop, customer, _, _ := setupSalesOrderTestData(t, pool)
	repo := NewSalesOrderRepository(pool)

	so := &models.SalesOrder{
		ClientID:    client.ID,
		CustomerID:  customer.ID,
		ShopID:      shop.ID,
		OrderNumber: "SO-DELETE",
		Status:      models.SOStatusPending,
		OrderDate:   time.Now(),
	}
	err := repo.Create(context.Background(), so)
	require.NoError(t, err)

	t.Run("successfully deletes sales order", func(t *testing.T) {
		err := repo.Delete(context.Background(), client.ID, so.ID)
		require.NoError(t, err)

		_, err = repo.GetByID(context.Background(), client.ID, so.ID)
		assert.Error(t, err)
	})
}

func TestSalesOrderRepository_ListByCustomer(t *testing.T) {
	pool := SetupTestDB(t)
	defer CleanupTestDB(t, pool)

	TruncateTable(t, pool, "sales_orders")
	TruncateTable(t, pool, "customers")
	TruncateTable(t, pool, "shops")
	TruncateTable(t, pool, "clients")

	client, shop, customer, _, _ := setupSalesOrderTestData(t, pool)
	repo := NewSalesOrderRepository(pool)

	// Create multiple orders for this customer
	for i := 1; i <= 3; i++ {
		so := &models.SalesOrder{
			ClientID:    client.ID,
			CustomerID:  customer.ID,
			ShopID:      shop.ID,
			OrderNumber: fmt.Sprintf("SO-%03d", i),
			Status:      models.SOStatusPending,
			OrderDate:   time.Now(),
		}
		err := repo.Create(context.Background(), so)
		require.NoError(t, err)
	}

	t.Run("successfully lists sales orders by customer", func(t *testing.T) {
		orders, err := repo.ListByCustomer(context.Background(), client.ID, customer.ID, 10, 0)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(orders), 3)
	})
}

func TestSalesOrderRepository_ListByStatus(t *testing.T) {
	pool := SetupTestDB(t)
	defer CleanupTestDB(t, pool)

	TruncateTable(t, pool, "sales_orders")
	TruncateTable(t, pool, "customers")
	TruncateTable(t, pool, "shops")
	TruncateTable(t, pool, "clients")

	client, shop, customer, _, _ := setupSalesOrderTestData(t, pool)
	repo := NewSalesOrderRepository(pool)

	// Create orders with different statuses
	for i := 1; i <= 5; i++ {
		status := models.SOStatusPending
		if i > 3 {
			status = models.SOStatusConfirmed
		}

		so := &models.SalesOrder{
			ClientID:    client.ID,
			CustomerID:  customer.ID,
			ShopID:      shop.ID,
			OrderNumber: fmt.Sprintf("SO-STATUS-%03d", i),
			Status:      status,
			OrderDate:   time.Now(),
		}
		err := repo.Create(context.Background(), so)
		require.NoError(t, err)
	}

	t.Run("successfully lists sales orders by status", func(t *testing.T) {
		pendingOrders, err := repo.ListByStatus(context.Background(), client.ID, models.SOStatusPending, 10, 0)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(pendingOrders), 3)

		for _, so := range pendingOrders {
			assert.Equal(t, models.SOStatusPending, so.Status)
		}
	})
}

func TestSalesOrderRepository_CreateItem(t *testing.T) {
	pool := SetupTestDB(t)
	defer CleanupTestDB(t, pool)

	TruncateTable(t, pool, "sales_order_items")
	TruncateTable(t, pool, "sales_orders")
	TruncateTable(t, pool, "customers")
	TruncateTable(t, pool, "product_variants")
	TruncateTable(t, pool, "products")
	TruncateTable(t, pool, "shops")
	TruncateTable(t, pool, "clients")

	client, shop, customer, _, variant := setupSalesOrderTestData(t, pool)
	repo := NewSalesOrderRepository(pool)

	so := &models.SalesOrder{
		ClientID:    client.ID,
		CustomerID:  customer.ID,
		ShopID:      shop.ID,
		OrderNumber: "SO-ITEMS",
		Status:      models.SOStatusPending,
		OrderDate:   time.Now(),
	}
	err := repo.Create(context.Background(), so)
	require.NoError(t, err)

	t.Run("successfully creates sales order item", func(t *testing.T) {
		item := &models.SalesOrderItem{
			ClientID:      client.ID,
			SalesOrderID:  so.ID,
			VariantID:     variant.ID,
			QuantityOrdered: 10,
			UnitPrice:     100.00,
			TaxRate:       0.08,
			DiscountAmount: 50.00,
			TotalPrice:    950.00,
			Notes:         "Test item",
		}

		err := repo.CreateItem(context.Background(), item)
		require.NoError(t, err)
		assert.NotEqual(t, uuid.Nil, item.ID)
		assert.Equal(t, 0, item.QuantityFulfilled)
		assert.NotZero(t, item.CreatedAt)
	})
}

func TestSalesOrderRepository_GetItem(t *testing.T) {
	pool := SetupTestDB(t)
	defer CleanupTestDB(t, pool)

	TruncateTable(t, pool, "sales_order_items")
	TruncateTable(t, pool, "sales_orders")
	TruncateTable(t, pool, "customers")
	TruncateTable(t, pool, "product_variants")
	TruncateTable(t, pool, "products")
	TruncateTable(t, pool, "shops")
	TruncateTable(t, pool, "clients")

	client, shop, customer, _, variant := setupSalesOrderTestData(t, pool)
	repo := NewSalesOrderRepository(pool)

	so := &models.SalesOrder{
		ClientID:    client.ID,
		CustomerID:  customer.ID,
		ShopID:      shop.ID,
		OrderNumber: "SO-001",
		Status:      models.SOStatusPending,
		OrderDate:   time.Now(),
	}
	err := repo.Create(context.Background(), so)
	require.NoError(t, err)

	item := &models.SalesOrderItem{
		ClientID:      client.ID,
		SalesOrderID:  so.ID,
		VariantID:     variant.ID,
		QuantityOrdered: 20,
		UnitPrice:     75.00,
		TotalPrice:    1500.00,
	}
	err = repo.CreateItem(context.Background(), item)
	require.NoError(t, err)

	t.Run("successfully retrieves sales order item", func(t *testing.T) {
		found, err := repo.GetItem(context.Background(), client.ID, item.ID)
		require.NoError(t, err)
		assert.Equal(t, item.ID, found.ID)
		assert.Equal(t, item.QuantityOrdered, found.QuantityOrdered)
		assert.Equal(t, item.UnitPrice, found.UnitPrice)
	})
}

func TestSalesOrderRepository_ListItemsBySO(t *testing.T) {
	pool := SetupTestDB(t)
	defer CleanupTestDB(t, pool)

	TruncateTable(t, pool, "sales_order_items")
	TruncateTable(t, pool, "sales_orders")
	TruncateTable(t, pool, "customers")
	TruncateTable(t, pool, "product_variants")
	TruncateTable(t, pool, "products")
	TruncateTable(t, pool, "shops")
	TruncateTable(t, pool, "clients")

	client, shop, customer, _, variant := setupSalesOrderTestData(t, pool)
	repo := NewSalesOrderRepository(pool)

	so := &models.SalesOrder{
		ClientID:    client.ID,
		CustomerID:  customer.ID,
		ShopID:      shop.ID,
		OrderNumber: "SO-LIST",
		Status:      models.SOStatusPending,
		OrderDate:   time.Now(),
	}
	err := repo.Create(context.Background(), so)
	require.NoError(t, err)

	// Create multiple items
	for i := 1; i <= 3; i++ {
		item := &models.SalesOrderItem{
			ClientID:      client.ID,
			SalesOrderID:  so.ID,
			VariantID:     variant.ID,
			QuantityOrdered: i * 10,
			UnitPrice:     100.00,
			TotalPrice:    float64(i * 1000),
		}
		err := repo.CreateItem(context.Background(), item)
		require.NoError(t, err)
	}

	t.Run("successfully lists all items for sales order", func(t *testing.T) {
		items, err := repo.ListItemsBySO(context.Background(), client.ID, so.ID)
		require.NoError(t, err)
		assert.Equal(t, 3, len(items))
	})
}

func TestSalesOrderRepository_FulfillItem(t *testing.T) {
	pool := SetupTestDB(t)
	defer CleanupTestDB(t, pool)

	TruncateTable(t, pool, "sales_order_items")
	TruncateTable(t, pool, "sales_orders")
	TruncateTable(t, pool, "customers")
	TruncateTable(t, pool, "product_variants")
	TruncateTable(t, pool, "products")
	TruncateTable(t, pool, "shops")
	TruncateTable(t, pool, "clients")

	client, shop, customer, _, variant := setupSalesOrderTestData(t, pool)
	repo := NewSalesOrderRepository(pool)

	so := &models.SalesOrder{
		ClientID:    client.ID,
		CustomerID:  customer.ID,
		ShopID:      shop.ID,
		OrderNumber: "SO-FULFILL",
		Status:      models.SOStatusConfirmed,
		OrderDate:   time.Now(),
	}
	err := repo.Create(context.Background(), so)
	require.NoError(t, err)

	item := &models.SalesOrderItem{
		ClientID:      client.ID,
		SalesOrderID:  so.ID,
		VariantID:     variant.ID,
		QuantityOrdered: 100,
		UnitPrice:     10.00,
		TotalPrice:    1000.00,
	}
	err = repo.CreateItem(context.Background(), item)
	require.NoError(t, err)

	t.Run("successfully fulfills items", func(t *testing.T) {
		err := repo.FulfillItem(context.Background(), client.ID, item.ID, 50)
		require.NoError(t, err)

		found, err := repo.GetItem(context.Background(), client.ID, item.ID)
		require.NoError(t, err)
		assert.Equal(t, 50, found.QuantityFulfilled)
	})

	t.Run("successfully fulfills additional items", func(t *testing.T) {
		err := repo.FulfillItem(context.Background(), client.ID, item.ID, 30)
		require.NoError(t, err)

		found, err := repo.GetItem(context.Background(), client.ID, item.ID)
		require.NoError(t, err)
		assert.Equal(t, 80, found.QuantityFulfilled) // 50 + 30
	})

	t.Run("fails when exceeding ordered quantity", func(t *testing.T) {
		err := repo.FulfillItem(context.Background(), client.ID, item.ID, 50)
		assert.Error(t, err) // Already fulfilled 80, can't fulfill 50 more (total would be 130, but ordered only 100)
	})
}
