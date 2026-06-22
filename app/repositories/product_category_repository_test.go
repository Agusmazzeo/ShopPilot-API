package repositories

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yourorg/shoppilot/app/models"
)

func TestProductCategoryRepository_Create(t *testing.T) {
	pool := SetupTestDB(t)
	defer CleanupTestDB(t, pool)

	TruncateTable(t, pool, "product_categories")

	repo := NewProductCategoryRepository(pool)

	t.Run("successfully creates a product category", func(t *testing.T) {
		category := &models.ProductCategory{
			ClientID:     1,
			ShopID:       1,
			Name:         "Electronics",
			Slug:         "electronics",
			DisplayOrder: 1,
			IsActive:     true,
		}

		err := repo.Create(context.Background(), category)
		require.NoError(t, err)
		assert.Greater(t, category.ID, 0)
		assert.NotZero(t, category.CreatedAt)
		assert.NotZero(t, category.UpdatedAt)
	})

	t.Run("fails with duplicate slug within same client and shop", func(t *testing.T) {
		category1 := &models.ProductCategory{
			ClientID:     1,
			ShopID:       1,
			Name:         "Category One",
			Slug:         "unique-category",
			DisplayOrder: 1,
			IsActive:     true,
		}
		err := repo.Create(context.Background(), category1)
		require.NoError(t, err)

		category2 := &models.ProductCategory{
			ClientID:     1,
			ShopID:       1,
			Name:         "Category Two",
			Slug:         "unique-category", // Same slug, same client+shop
			DisplayOrder: 2,
			IsActive:     true,
		}
		err = repo.Create(context.Background(), category2)
		assert.Error(t, err)
	})

	t.Run("allows duplicate slug across different shops", func(t *testing.T) {
		category1 := &models.ProductCategory{
			ClientID:     1,
			ShopID:       1,
			Name:         "Category One",
			Slug:         "shared-category",
			DisplayOrder: 1,
			IsActive:     true,
		}
		err := repo.Create(context.Background(), category1)
		require.NoError(t, err)

		category2 := &models.ProductCategory{
			ClientID:     1,
			ShopID:       2, // Different shop
			Name:         "Category Two",
			Slug:         "shared-category", // Same slug
			DisplayOrder: 1,
			IsActive:     true,
		}
		err = repo.Create(context.Background(), category2)
		require.NoError(t, err)
		assert.NotEqual(t, category1.ID, category2.ID)
	})

	t.Run("successfully creates hierarchical categories", func(t *testing.T) {
		parent := &models.ProductCategory{
			ClientID:     1,
			ShopID:       1,
			Name:         "Parent Category",
			Slug:         "parent",
			DisplayOrder: 1,
			IsActive:     true,
		}
		err := repo.Create(context.Background(), parent)
		require.NoError(t, err)

		child := &models.ProductCategory{
			ClientID:     1,
			ShopID:       1,
			Name:         "Child Category",
			Slug:         "child",
			ParentID:     &parent.ID,
			DisplayOrder: 1,
			IsActive:     true,
		}
		err = repo.Create(context.Background(), child)
		require.NoError(t, err)
		assert.Equal(t, parent.ID, *child.ParentID)
	})
}

func TestProductCategoryRepository_GetByID(t *testing.T) {
	pool := SetupTestDB(t)
	defer CleanupTestDB(t, pool)

	TruncateTable(t, pool, "product_categories")

	repo := NewProductCategoryRepository(pool)

	category := &models.ProductCategory{
		ClientID:     1,
		ShopID:       1,
		Name:         "Test Category",
		Slug:         "test-category",
		DisplayOrder: 1,
		IsActive:     true,
	}
	err := repo.Create(context.Background(), category)
	require.NoError(t, err)

	t.Run("successfully retrieves category by ID", func(t *testing.T) {
		found, err := repo.GetByID(context.Background(), category.ID)
		require.NoError(t, err)
		assert.Equal(t, category.ID, found.ID)
		assert.Equal(t, category.Name, found.Name)
		assert.Equal(t, category.Slug, found.Slug)
	})

	t.Run("returns error for non-existent ID", func(t *testing.T) {
		_, err := repo.GetByID(context.Background(), 99999)
		assert.Error(t, err)
	})
}

func TestProductCategoryRepository_ListByShopID(t *testing.T) {
	pool := SetupTestDB(t)
	defer CleanupTestDB(t, pool)

	TruncateTable(t, pool, "product_categories")

	repo := NewProductCategoryRepository(pool)

	// Create categories for shop 1
	for i := 1; i <= 3; i++ {
		category := &models.ProductCategory{
			ClientID:     1,
			ShopID:       1,
			Name:         fmt.Sprintf("Category %d", i),
			Slug:         fmt.Sprintf("category-%d", i),
			DisplayOrder: i,
			IsActive:     true,
		}
		err := repo.Create(context.Background(), category)
		require.NoError(t, err)
	}

	// Create categories for shop 2
	for i := 1; i <= 2; i++ {
		category := &models.ProductCategory{
			ClientID:     1,
			ShopID:       2,
			Name:         fmt.Sprintf("Shop2 Category %d", i),
			Slug:         fmt.Sprintf("shop2-category-%d", i),
			DisplayOrder: i,
			IsActive:     true,
		}
		err := repo.Create(context.Background(), category)
		require.NoError(t, err)
	}

	t.Run("successfully lists categories for shop 1", func(t *testing.T) {
		categories, err := repo.ListByShopID(context.Background(), 1)
		require.NoError(t, err)
		assert.Len(t, categories, 3)
		// Verify order
		assert.Equal(t, 1, categories[0].DisplayOrder)
		assert.Equal(t, 2, categories[1].DisplayOrder)
		assert.Equal(t, 3, categories[2].DisplayOrder)
	})

	t.Run("successfully lists categories for shop 2", func(t *testing.T) {
		categories, err := repo.ListByShopID(context.Background(), 2)
		require.NoError(t, err)
		assert.Len(t, categories, 2)
	})

	t.Run("returns empty list for shop with no categories", func(t *testing.T) {
		categories, err := repo.ListByShopID(context.Background(), 99)
		require.NoError(t, err)
		assert.Empty(t, categories)
	})
}

func TestProductCategoryRepository_ListByClientID(t *testing.T) {
	pool := SetupTestDB(t)
	defer CleanupTestDB(t, pool)

	TruncateTable(t, pool, "product_categories")

	repo := NewProductCategoryRepository(pool)

	// Create categories for client 1, shop 1
	for i := 1; i <= 2; i++ {
		category := &models.ProductCategory{
			ClientID:     1,
			ShopID:       1,
			Name:         fmt.Sprintf("C1S1 Category %d", i),
			Slug:         fmt.Sprintf("c1s1-cat-%d", i),
			DisplayOrder: i,
			IsActive:     true,
		}
		err := repo.Create(context.Background(), category)
		require.NoError(t, err)
	}

	// Create categories for client 1, shop 2
	category := &models.ProductCategory{
		ClientID:     1,
		ShopID:       2,
		Name:         "C1S2 Category",
		Slug:         "c1s2-cat",
		DisplayOrder: 1,
		IsActive:     true,
	}
	err := repo.Create(context.Background(), category)
	require.NoError(t, err)

	// Create categories for client 2
	category2 := &models.ProductCategory{
		ClientID:     2,
		ShopID:       1,
		Name:         "C2 Category",
		Slug:         "c2-cat",
		DisplayOrder: 1,
		IsActive:     true,
	}
	err = repo.Create(context.Background(), category2)
	require.NoError(t, err)

	t.Run("successfully lists all categories for client 1", func(t *testing.T) {
		categories, err := repo.ListByClientID(context.Background(), 1)
		require.NoError(t, err)
		assert.Len(t, categories, 3)
	})

	t.Run("successfully lists categories for client 2", func(t *testing.T) {
		categories, err := repo.ListByClientID(context.Background(), 2)
		require.NoError(t, err)
		assert.Len(t, categories, 1)
	})

	t.Run("returns empty list for client with no categories", func(t *testing.T) {
		categories, err := repo.ListByClientID(context.Background(), 99)
		require.NoError(t, err)
		assert.Empty(t, categories)
	})
}

func TestProductCategoryRepository_Update(t *testing.T) {
	pool := SetupTestDB(t)
	defer CleanupTestDB(t, pool)

	TruncateTable(t, pool, "product_categories")

	repo := NewProductCategoryRepository(pool)

	category := &models.ProductCategory{
		ClientID:     1,
		ShopID:       1,
		Name:         "Original Name",
		Slug:         "original",
		DisplayOrder: 1,
		IsActive:     true,
	}
	err := repo.Create(context.Background(), category)
	require.NoError(t, err)

	t.Run("successfully updates category", func(t *testing.T) {
		category.Name = "Updated Name"
		category.DisplayOrder = 5
		err := repo.Update(context.Background(), category)
		require.NoError(t, err)

		// Verify update
		found, err := repo.GetByID(context.Background(), category.ID)
		require.NoError(t, err)
		assert.Equal(t, "Updated Name", found.Name)
		assert.Equal(t, 5, found.DisplayOrder)
	})

	t.Run("returns error for non-existent category", func(t *testing.T) {
		nonExistent := &models.ProductCategory{
			ID:           99999,
			Name:         "Test",
			DisplayOrder: 1,
			IsActive:     true,
		}
		err := repo.Update(context.Background(), nonExistent)
		assert.Error(t, err)
	})
}

func TestProductCategoryRepository_Delete(t *testing.T) {
	pool := SetupTestDB(t)
	defer CleanupTestDB(t, pool)

	TruncateTable(t, pool, "product_categories")

	repo := NewProductCategoryRepository(pool)

	category := &models.ProductCategory{
		ClientID:     1,
		ShopID:       1,
		Name:         "To Delete",
		Slug:         "to-delete",
		DisplayOrder: 1,
		IsActive:     true,
	}
	err := repo.Create(context.Background(), category)
	require.NoError(t, err)

	t.Run("successfully deletes category", func(t *testing.T) {
		err := repo.Delete(context.Background(), category.ID)
		require.NoError(t, err)

		// Verify deletion
		_, err = repo.GetByID(context.Background(), category.ID)
		assert.Error(t, err)
	})

	t.Run("returns error for non-existent category", func(t *testing.T) {
		err := repo.Delete(context.Background(), 99999)
		assert.Error(t, err)
	})
}
