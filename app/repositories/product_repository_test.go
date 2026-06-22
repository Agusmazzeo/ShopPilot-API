package repositories

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yourorg/shoppilot/app/models"
)

func TestProductRepository_Create(t *testing.T) {
	pool := SetupTestDB(t)
	defer CleanupTestDB(t, pool)

	TruncateTable(t, pool, "products")

	repo := NewProductRepository(pool)

	t.Run("successfully creates a product", func(t *testing.T) {
		product := &models.Product{
			ClientID:         1,
			ShopID:           1,
			Name:             "Test Product",
			Slug:             "test-product",
			Price:            29.99,
			WeightUnit:       "kg",
			RequiresShipping: true,
			IsActive:         true,
		}

		err := repo.Create(context.Background(), product)
		require.NoError(t, err)
		assert.Greater(t, product.ID, 0)
		assert.NotZero(t, product.CreatedAt)
		assert.NotZero(t, product.UpdatedAt)
	})

	t.Run("fails with duplicate slug within same client and shop", func(t *testing.T) {
		product1 := &models.Product{
			ClientID:         1,
			ShopID:           1,
			Name:             "Product One",
			Slug:             "unique-product",
			Price:            19.99,
			WeightUnit:       "kg",
			RequiresShipping: true,
			IsActive:         true,
		}
		err := repo.Create(context.Background(), product1)
		require.NoError(t, err)

		product2 := &models.Product{
			ClientID:         1,
			ShopID:           1,
			Name:             "Product Two",
			Slug:             "unique-product", // Same slug, same client+shop
			Price:            29.99,
			WeightUnit:       "kg",
			RequiresShipping: true,
			IsActive:         true,
		}
		err = repo.Create(context.Background(), product2)
		assert.Error(t, err)
	})

	t.Run("allows duplicate slug across different shops", func(t *testing.T) {
		product1 := &models.Product{
			ClientID:         1,
			ShopID:           1,
			Name:             "Product One",
			Slug:             "shared-product",
			Price:            19.99,
			WeightUnit:       "kg",
			RequiresShipping: true,
			IsActive:         true,
		}
		err := repo.Create(context.Background(), product1)
		require.NoError(t, err)

		product2 := &models.Product{
			ClientID:         1,
			ShopID:           2, // Different shop
			Name:             "Product Two",
			Slug:             "shared-product", // Same slug
			Price:            29.99,
			WeightUnit:       "kg",
			RequiresShipping: true,
			IsActive:         true,
		}
		err = repo.Create(context.Background(), product2)
		require.NoError(t, err)
		assert.NotEqual(t, product1.ID, product2.ID)
	})

	t.Run("successfully creates product with category", func(t *testing.T) {
		categoryID := 1
		product := &models.Product{
			ClientID:         1,
			ShopID:           1,
			CategoryID:       &categoryID,
			Name:             "Categorized Product",
			Slug:             "categorized-product",
			Price:            39.99,
			WeightUnit:       "kg",
			RequiresShipping: true,
			IsActive:         true,
		}
		err := repo.Create(context.Background(), product)
		require.NoError(t, err)
		assert.Equal(t, categoryID, *product.CategoryID)
	})
}

func TestProductRepository_GetByID(t *testing.T) {
	pool := SetupTestDB(t)
	defer CleanupTestDB(t, pool)

	TruncateTable(t, pool, "products")

	repo := NewProductRepository(pool)

	product := &models.Product{
		ClientID:         1,
		ShopID:           1,
		Name:             "Test Product",
		Slug:             "test-product",
		Price:            29.99,
		WeightUnit:       "kg",
		RequiresShipping: true,
		IsActive:         true,
	}
	err := repo.Create(context.Background(), product)
	require.NoError(t, err)

	t.Run("successfully retrieves product by ID", func(t *testing.T) {
		found, err := repo.GetByID(context.Background(), product.ID)
		require.NoError(t, err)
		assert.Equal(t, product.ID, found.ID)
		assert.Equal(t, product.Name, found.Name)
		assert.Equal(t, product.Slug, found.Slug)
		assert.Equal(t, product.Price, found.Price)
	})

	t.Run("returns error for non-existent ID", func(t *testing.T) {
		_, err := repo.GetByID(context.Background(), 99999)
		assert.Error(t, err)
	})
}

func TestProductRepository_GetBySlug(t *testing.T) {
	pool := SetupTestDB(t)
	defer CleanupTestDB(t, pool)

	TruncateTable(t, pool, "products")

	repo := NewProductRepository(pool)

	product := &models.Product{
		ClientID:         1,
		ShopID:           1,
		Name:             "Test Product",
		Slug:             "test-product",
		Price:            29.99,
		WeightUnit:       "kg",
		RequiresShipping: true,
		IsActive:         true,
	}
	err := repo.Create(context.Background(), product)
	require.NoError(t, err)

	t.Run("successfully retrieves product by slug, client, and shop", func(t *testing.T) {
		found, err := repo.GetBySlug(context.Background(), 1, 1, "test-product")
		require.NoError(t, err)
		assert.Equal(t, product.ID, found.ID)
		assert.Equal(t, "test-product", found.Slug)
	})

	t.Run("returns error for non-existent slug", func(t *testing.T) {
		_, err := repo.GetBySlug(context.Background(), 1, 1, "nonexistent")
		assert.Error(t, err)
	})

	t.Run("returns error for wrong client", func(t *testing.T) {
		_, err := repo.GetBySlug(context.Background(), 2, 1, "test-product")
		assert.Error(t, err)
	})

	t.Run("returns error for wrong shop", func(t *testing.T) {
		_, err := repo.GetBySlug(context.Background(), 1, 2, "test-product")
		assert.Error(t, err)
	})
}

func TestProductRepository_ListByShopID(t *testing.T) {
	pool := SetupTestDB(t)
	defer CleanupTestDB(t, pool)

	TruncateTable(t, pool, "products")

	repo := NewProductRepository(pool)

	// Create products for shop 1
	for i := 1; i <= 3; i++ {
		product := &models.Product{
			ClientID:         1,
			ShopID:           1,
			Name:             fmt.Sprintf("Product %d", i),
			Slug:             fmt.Sprintf("product-%d", i),
			Price:            float64(i) * 10.0,
			WeightUnit:       "kg",
			RequiresShipping: true,
			IsActive:         true,
		}
		err := repo.Create(context.Background(), product)
		require.NoError(t, err)
	}

	// Create products for shop 2
	for i := 1; i <= 2; i++ {
		product := &models.Product{
			ClientID:         1,
			ShopID:           2,
			Name:             fmt.Sprintf("Shop2 Product %d", i),
			Slug:             fmt.Sprintf("shop2-product-%d", i),
			Price:            float64(i) * 15.0,
			WeightUnit:       "kg",
			RequiresShipping: true,
			IsActive:         true,
		}
		err := repo.Create(context.Background(), product)
		require.NoError(t, err)
	}

	t.Run("successfully lists products for shop 1", func(t *testing.T) {
		products, err := repo.ListByShopID(context.Background(), 1)
		require.NoError(t, err)
		assert.Len(t, products, 3)
	})

	t.Run("successfully lists products for shop 2", func(t *testing.T) {
		products, err := repo.ListByShopID(context.Background(), 2)
		require.NoError(t, err)
		assert.Len(t, products, 2)
	})

	t.Run("returns empty list for shop with no products", func(t *testing.T) {
		products, err := repo.ListByShopID(context.Background(), 99)
		require.NoError(t, err)
		assert.Empty(t, products)
	})
}

func TestProductRepository_ListByCategoryID(t *testing.T) {
	pool := SetupTestDB(t)
	defer CleanupTestDB(t, pool)

	TruncateTable(t, pool, "products")

	repo := NewProductRepository(pool)

	categoryID1 := 1
	categoryID2 := 2

	// Create products for category 1
	for i := 1; i <= 2; i++ {
		product := &models.Product{
			ClientID:         1,
			ShopID:           1,
			CategoryID:       &categoryID1,
			Name:             fmt.Sprintf("Cat1 Product %d", i),
			Slug:             fmt.Sprintf("cat1-product-%d", i),
			Price:            float64(i) * 10.0,
			WeightUnit:       "kg",
			RequiresShipping: true,
			IsActive:         true,
		}
		err := repo.Create(context.Background(), product)
		require.NoError(t, err)
	}

	// Create products for category 2
	product := &models.Product{
		ClientID:         1,
		ShopID:           1,
		CategoryID:       &categoryID2,
		Name:             "Cat2 Product",
		Slug:             "cat2-product",
		Price:            25.0,
		WeightUnit:       "kg",
		RequiresShipping: true,
		IsActive:         true,
	}
	err := repo.Create(context.Background(), product)
	require.NoError(t, err)

	t.Run("successfully lists products for category 1", func(t *testing.T) {
		products, err := repo.ListByCategoryID(context.Background(), categoryID1)
		require.NoError(t, err)
		assert.Len(t, products, 2)
	})

	t.Run("successfully lists products for category 2", func(t *testing.T) {
		products, err := repo.ListByCategoryID(context.Background(), categoryID2)
		require.NoError(t, err)
		assert.Len(t, products, 1)
	})

	t.Run("returns empty list for category with no products", func(t *testing.T) {
		products, err := repo.ListByCategoryID(context.Background(), 99)
		require.NoError(t, err)
		assert.Empty(t, products)
	})
}

func TestProductRepository_ListByClientID(t *testing.T) {
	pool := SetupTestDB(t)
	defer CleanupTestDB(t, pool)

	TruncateTable(t, pool, "products")

	repo := NewProductRepository(pool)

	// Create products for client 1, shop 1
	for i := 1; i <= 2; i++ {
		product := &models.Product{
			ClientID:         1,
			ShopID:           1,
			Name:             fmt.Sprintf("C1S1 Product %d", i),
			Slug:             fmt.Sprintf("c1s1-product-%d", i),
			Price:            float64(i) * 10.0,
			WeightUnit:       "kg",
			RequiresShipping: true,
			IsActive:         true,
		}
		err := repo.Create(context.Background(), product)
		require.NoError(t, err)
	}

	// Create products for client 1, shop 2
	product := &models.Product{
		ClientID:         1,
		ShopID:           2,
		Name:             "C1S2 Product",
		Slug:             "c1s2-product",
		Price:            25.0,
		WeightUnit:       "kg",
		RequiresShipping: true,
		IsActive:         true,
	}
	err := repo.Create(context.Background(), product)
	require.NoError(t, err)

	// Create products for client 2
	product2 := &models.Product{
		ClientID:         2,
		ShopID:           1,
		Name:             "C2 Product",
		Slug:             "c2-product",
		Price:            30.0,
		WeightUnit:       "kg",
		RequiresShipping: true,
		IsActive:         true,
	}
	err = repo.Create(context.Background(), product2)
	require.NoError(t, err)

	t.Run("successfully lists all products for client 1", func(t *testing.T) {
		products, err := repo.ListByClientID(context.Background(), 1)
		require.NoError(t, err)
		assert.Len(t, products, 3)
	})

	t.Run("successfully lists products for client 2", func(t *testing.T) {
		products, err := repo.ListByClientID(context.Background(), 2)
		require.NoError(t, err)
		assert.Len(t, products, 1)
	})

	t.Run("returns empty list for client with no products", func(t *testing.T) {
		products, err := repo.ListByClientID(context.Background(), 99)
		require.NoError(t, err)
		assert.Empty(t, products)
	})
}

func TestProductRepository_Update(t *testing.T) {
	pool := SetupTestDB(t)
	defer CleanupTestDB(t, pool)

	TruncateTable(t, pool, "products")

	repo := NewProductRepository(pool)

	product := &models.Product{
		ClientID:         1,
		ShopID:           1,
		Name:             "Original Name",
		Slug:             "original",
		Price:            29.99,
		WeightUnit:       "kg",
		RequiresShipping: true,
		IsActive:         true,
	}
	err := repo.Create(context.Background(), product)
	require.NoError(t, err)

	t.Run("successfully updates product", func(t *testing.T) {
		product.Name = "Updated Name"
		product.Price = 39.99
		newWeight := 2.5
		product.Weight = &newWeight
		err := repo.Update(context.Background(), product)
		require.NoError(t, err)

		// Verify update
		found, err := repo.GetByID(context.Background(), product.ID)
		require.NoError(t, err)
		assert.Equal(t, "Updated Name", found.Name)
		assert.Equal(t, 39.99, found.Price)
		assert.Equal(t, 2.5, *found.Weight)
	})

	t.Run("returns error for non-existent product", func(t *testing.T) {
		nonExistent := &models.Product{
			ID:               99999,
			Name:             "Test",
			Price:            19.99,
			WeightUnit:       "kg",
			RequiresShipping: true,
			IsActive:         true,
		}
		err := repo.Update(context.Background(), nonExistent)
		assert.Error(t, err)
	})
}

func TestProductRepository_Delete(t *testing.T) {
	pool := SetupTestDB(t)
	defer CleanupTestDB(t, pool)

	TruncateTable(t, pool, "products")

	repo := NewProductRepository(pool)

	product := &models.Product{
		ClientID:         1,
		ShopID:           1,
		Name:             "To Delete",
		Slug:             "to-delete",
		Price:            29.99,
		WeightUnit:       "kg",
		RequiresShipping: true,
		IsActive:         true,
	}
	err := repo.Create(context.Background(), product)
	require.NoError(t, err)

	t.Run("successfully deletes product", func(t *testing.T) {
		err := repo.Delete(context.Background(), product.ID)
		require.NoError(t, err)

		// Verify deletion
		_, err = repo.GetByID(context.Background(), product.ID)
		assert.Error(t, err)
	})

	t.Run("returns error for non-existent product", func(t *testing.T) {
		err := repo.Delete(context.Background(), 99999)
		assert.Error(t, err)
	})
}
