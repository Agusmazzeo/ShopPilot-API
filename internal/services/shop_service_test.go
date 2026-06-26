package services

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/yourorg/shoppilot/internal/models"
)

func TestCreateShop(t *testing.T) {
	ctx := context.Background()
	clientID := uuid.New()
	shopID := uuid.New()

	t.Run("success", func(t *testing.T) {
		mockShopRepo := new(MockShopRepository)
		mockClientRepo := new(MockClientRepository)
		service := NewShopService(mockShopRepo, mockClientRepo)

		client := &models.Client{
			ID:   clientID,
			Name: "Test Client",
			Slug: "test-client",
		}

		req := &CreateShopRequest{
			Name:        "Coffee Shop",
			Description: "A great coffee shop",
			WebpageURL:  "https://coffeeshop.com",
			Address:     "123 Main St",
			City:        "New York",
			State:       "NY",
			Country:     "USA",
			PostalCode:  "10001",
			Phone:       "555-1234",
			Email:       "info@coffeeshop.com",
		}

		mockClientRepo.On("GetByID", ctx, clientID).Return(client, nil)
		mockShopRepo.On("GetBySlug", ctx, clientID, "coffee-shop").Return(nil, fmt.Errorf("not found"))
		mockShopRepo.On("Create", ctx, mock.MatchedBy(func(shop *models.Shop) bool {
			return shop.ClientID == clientID &&
				shop.Name == "Coffee Shop" &&
				shop.Slug == "coffee-shop" &&
				shop.Description == "A great coffee shop" &&
				shop.IsActive == true
		})).Return(nil).Run(func(args mock.Arguments) {
			shop := args.Get(1).(*models.Shop)
			shop.ID = shopID
		})

		result, err := service.CreateShop(ctx, clientID, req)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, clientID, result.ClientID)
		assert.Equal(t, "Coffee Shop", result.Name)
		assert.Equal(t, "coffee-shop", result.Slug)
		assert.True(t, result.IsActive)
		mockClientRepo.AssertExpectations(t)
		mockShopRepo.AssertExpectations(t)
	})

	t.Run("client not found", func(t *testing.T) {
		mockShopRepo := new(MockShopRepository)
		mockClientRepo := new(MockClientRepository)
		service := NewShopService(mockShopRepo, mockClientRepo)

		req := &CreateShopRequest{
			Name: "Coffee Shop",
		}

		mockClientRepo.On("GetByID", ctx, clientID).Return(nil, fmt.Errorf("not found"))

		result, err := service.CreateShop(ctx, clientID, req)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "client validation failed")
		mockClientRepo.AssertExpectations(t)
	})

	t.Run("slug uniqueness - generates unique slug when duplicate exists", func(t *testing.T) {
		mockShopRepo := new(MockShopRepository)
		mockClientRepo := new(MockClientRepository)
		service := NewShopService(mockShopRepo, mockClientRepo)

		client := &models.Client{
			ID:   clientID,
			Name: "Test Client",
			Slug: "test-client",
		}

		existingShop := &models.Shop{
			ID:       uuid.New(),
			ClientID: clientID,
			Name:     "Coffee Shop",
			Slug:     "coffee-shop",
		}

		req := &CreateShopRequest{
			Name: "Coffee Shop",
		}

		mockClientRepo.On("GetByID", ctx, clientID).Return(client, nil)
		mockShopRepo.On("GetBySlug", ctx, clientID, "coffee-shop").Return(existingShop, nil)
		mockShopRepo.On("GetBySlug", ctx, clientID, "coffee-shop-1").Return(nil, fmt.Errorf("not found"))
		mockShopRepo.On("Create", ctx, mock.MatchedBy(func(shop *models.Shop) bool {
			return shop.Slug == "coffee-shop-1"
		})).Return(nil)

		result, err := service.CreateShop(ctx, clientID, req)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "coffee-shop-1", result.Slug)
		mockClientRepo.AssertExpectations(t)
		mockShopRepo.AssertExpectations(t)
	})
}

func TestGetShop(t *testing.T) {
	ctx := context.Background()
	clientID := uuid.New()
	shopID := uuid.New()

	t.Run("success", func(t *testing.T) {
		mockShopRepo := new(MockShopRepository)
		mockClientRepo := new(MockClientRepository)
		service := NewShopService(mockShopRepo, mockClientRepo)

		expectedShop := &models.Shop{
			ID:       shopID,
			ClientID: clientID,
			Name:     "Coffee Shop",
			Slug:     "coffee-shop",
		}

		mockShopRepo.On("GetByID", ctx, clientID, shopID).Return(expectedShop, nil)

		result, err := service.GetShop(ctx, clientID, shopID)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, shopID, result.ID)
		assert.Equal(t, "Coffee Shop", result.Name)
		mockShopRepo.AssertExpectations(t)
	})

	t.Run("shop not found", func(t *testing.T) {
		mockShopRepo := new(MockShopRepository)
		mockClientRepo := new(MockClientRepository)
		service := NewShopService(mockShopRepo, mockClientRepo)

		mockShopRepo.On("GetByID", ctx, clientID, shopID).Return(nil, fmt.Errorf("not found"))

		result, err := service.GetShop(ctx, clientID, shopID)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "failed to get shop")
		mockShopRepo.AssertExpectations(t)
	})
}

func TestGetShopBySlug(t *testing.T) {
	ctx := context.Background()
	clientID := uuid.New()
	shopID := uuid.New()

	t.Run("success", func(t *testing.T) {
		mockShopRepo := new(MockShopRepository)
		mockClientRepo := new(MockClientRepository)
		service := NewShopService(mockShopRepo, mockClientRepo)

		expectedShop := &models.Shop{
			ID:       shopID,
			ClientID: clientID,
			Name:     "Coffee Shop",
			Slug:     "coffee-shop",
		}

		mockShopRepo.On("GetBySlug", ctx, clientID, "coffee-shop").Return(expectedShop, nil)

		result, err := service.GetShopBySlug(ctx, clientID, "coffee-shop")

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "coffee-shop", result.Slug)
		mockShopRepo.AssertExpectations(t)
	})

	t.Run("shop not found", func(t *testing.T) {
		mockShopRepo := new(MockShopRepository)
		mockClientRepo := new(MockClientRepository)
		service := NewShopService(mockShopRepo, mockClientRepo)

		mockShopRepo.On("GetBySlug", ctx, clientID, "nonexistent").Return(nil, fmt.Errorf("not found"))

		result, err := service.GetShopBySlug(ctx, clientID, "nonexistent")

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "failed to get shop by slug")
		mockShopRepo.AssertExpectations(t)
	})
}

func TestUpdateShop(t *testing.T) {
	ctx := context.Background()
	clientID := uuid.New()
	shopID := uuid.New()

	t.Run("success - basic update", func(t *testing.T) {
		mockShopRepo := new(MockShopRepository)
		mockClientRepo := new(MockClientRepository)
		service := NewShopService(mockShopRepo, mockClientRepo)

		existingShop := &models.Shop{
			ID:       shopID,
			ClientID: clientID,
			Name:     "Coffee Shop",
			Slug:     "coffee-shop",
			IsActive: true,
		}

		newDescription := "Updated description"
		newCity := "Boston"
		req := &UpdateShopRequest{
			Description: &newDescription,
			City:        &newCity,
		}

		mockShopRepo.On("GetByID", ctx, clientID, shopID).Return(existingShop, nil)
		mockShopRepo.On("Update", ctx, mock.MatchedBy(func(shop *models.Shop) bool {
			return shop.ID == shopID &&
				shop.Description == "Updated description" &&
				shop.City == "Boston"
		})).Return(nil)

		err := service.UpdateShop(ctx, clientID, shopID, req)

		assert.NoError(t, err)
		mockShopRepo.AssertExpectations(t)
	})

	t.Run("name change regenerates slug", func(t *testing.T) {
		mockShopRepo := new(MockShopRepository)
		mockClientRepo := new(MockClientRepository)
		service := NewShopService(mockShopRepo, mockClientRepo)

		existingShop := &models.Shop{
			ID:       shopID,
			ClientID: clientID,
			Name:     "Coffee Shop",
			Slug:     "coffee-shop",
		}

		newName := "Tea House"
		req := &UpdateShopRequest{
			Name: &newName,
		}

		mockShopRepo.On("GetByID", ctx, clientID, shopID).Return(existingShop, nil)
		mockShopRepo.On("GetBySlug", ctx, clientID, "tea-house").Return(nil, fmt.Errorf("not found"))
		mockShopRepo.On("Update", ctx, mock.MatchedBy(func(shop *models.Shop) bool {
			return shop.Name == "Tea House" && shop.Slug == "tea-house"
		})).Return(nil)

		err := service.UpdateShop(ctx, clientID, shopID, req)

		assert.NoError(t, err)
		mockShopRepo.AssertExpectations(t)
	})

	t.Run("name change with duplicate slug generates unique slug", func(t *testing.T) {
		mockShopRepo := new(MockShopRepository)
		mockClientRepo := new(MockClientRepository)
		service := NewShopService(mockShopRepo, mockClientRepo)

		existingShop := &models.Shop{
			ID:       shopID,
			ClientID: clientID,
			Name:     "Coffee Shop",
			Slug:     "coffee-shop",
		}

		otherShop := &models.Shop{
			ID:       uuid.New(),
			ClientID: clientID,
			Name:     "Tea House",
			Slug:     "tea-house",
		}

		newName := "Tea House"
		req := &UpdateShopRequest{
			Name: &newName,
		}

		mockShopRepo.On("GetByID", ctx, clientID, shopID).Return(existingShop, nil)
		mockShopRepo.On("GetBySlug", ctx, clientID, "tea-house").Return(otherShop, nil)
		mockShopRepo.On("GetBySlug", ctx, clientID, "tea-house-1").Return(nil, fmt.Errorf("not found"))
		mockShopRepo.On("Update", ctx, mock.MatchedBy(func(shop *models.Shop) bool {
			return shop.Name == "Tea House" && shop.Slug == "tea-house-1"
		})).Return(nil)

		err := service.UpdateShop(ctx, clientID, shopID, req)

		assert.NoError(t, err)
		mockShopRepo.AssertExpectations(t)
	})

	t.Run("shop not found", func(t *testing.T) {
		mockShopRepo := new(MockShopRepository)
		mockClientRepo := new(MockClientRepository)
		service := NewShopService(mockShopRepo, mockClientRepo)

		newName := "New Name"
		req := &UpdateShopRequest{
			Name: &newName,
		}

		mockShopRepo.On("GetByID", ctx, clientID, shopID).Return(nil, fmt.Errorf("not found"))

		err := service.UpdateShop(ctx, clientID, shopID, req)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to get shop")
		mockShopRepo.AssertExpectations(t)
	})

	t.Run("update all fields", func(t *testing.T) {
		mockShopRepo := new(MockShopRepository)
		mockClientRepo := new(MockClientRepository)
		service := NewShopService(mockShopRepo, mockClientRepo)

		existingShop := &models.Shop{
			ID:       shopID,
			ClientID: clientID,
			Name:     "Coffee Shop",
			Slug:     "coffee-shop",
		}

		newName := "Updated Shop"
		newDesc := "New Description"
		newWebpage := "https://newwebpage.com"
		newAddress := "456 Elm St"
		newCity := "Boston"
		newState := "MA"
		newCountry := "USA"
		newPostal := "02101"
		newPhone := "555-9999"
		newEmail := "new@shop.com"
		logoURL := "https://logo.com/logo.png"
		isActive := false

		req := &UpdateShopRequest{
			Name:        &newName,
			Description: &newDesc,
			WebpageURL:  &newWebpage,
			Address:     &newAddress,
			City:        &newCity,
			State:       &newState,
			Country:     &newCountry,
			PostalCode:  &newPostal,
			Phone:       &newPhone,
			Email:       &newEmail,
			LogoURL:     &logoURL,
			IsActive:    &isActive,
		}

		mockShopRepo.On("GetByID", ctx, clientID, shopID).Return(existingShop, nil)
		mockShopRepo.On("GetBySlug", ctx, clientID, "updated-shop").Return(nil, fmt.Errorf("not found"))
		mockShopRepo.On("Update", ctx, mock.MatchedBy(func(shop *models.Shop) bool {
			return shop.Name == newName &&
				shop.Slug == "updated-shop" &&
				shop.Description == newDesc &&
				shop.WebpageURL == newWebpage &&
				shop.Address == newAddress &&
				shop.City == newCity &&
				shop.State == newState &&
				shop.Country == newCountry &&
				shop.PostalCode == newPostal &&
				shop.Phone == newPhone &&
				shop.Email == newEmail &&
				shop.LogoURL != nil &&
				*shop.LogoURL == logoURL &&
				shop.IsActive == false
		})).Return(nil)

		err := service.UpdateShop(ctx, clientID, shopID, req)

		assert.NoError(t, err)
		mockShopRepo.AssertExpectations(t)
	})
}

func TestDeleteShop(t *testing.T) {
	ctx := context.Background()
	clientID := uuid.New()
	shopID := uuid.New()

	t.Run("success", func(t *testing.T) {
		mockShopRepo := new(MockShopRepository)
		mockClientRepo := new(MockClientRepository)
		service := NewShopService(mockShopRepo, mockClientRepo)

		existingShop := &models.Shop{
			ID:       shopID,
			ClientID: clientID,
			Name:     "Coffee Shop",
			Slug:     "coffee-shop",
		}

		mockShopRepo.On("GetByID", ctx, clientID, shopID).Return(existingShop, nil)
		mockShopRepo.On("Delete", ctx, clientID, shopID).Return(nil)

		err := service.DeleteShop(ctx, clientID, shopID)

		assert.NoError(t, err)
		mockShopRepo.AssertExpectations(t)
	})

	t.Run("shop not found", func(t *testing.T) {
		mockShopRepo := new(MockShopRepository)
		mockClientRepo := new(MockClientRepository)
		service := NewShopService(mockShopRepo, mockClientRepo)

		mockShopRepo.On("GetByID", ctx, clientID, shopID).Return(nil, fmt.Errorf("not found"))

		err := service.DeleteShop(ctx, clientID, shopID)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to get shop")
		mockShopRepo.AssertExpectations(t)
	})
}

func TestListShops(t *testing.T) {
	ctx := context.Background()
	clientID := uuid.New()

	t.Run("success with pagination", func(t *testing.T) {
		mockShopRepo := new(MockShopRepository)
		mockClientRepo := new(MockClientRepository)
		service := NewShopService(mockShopRepo, mockClientRepo)

		expectedShops := []*models.Shop{
			{
				ID:       uuid.New(),
				ClientID: clientID,
				Name:     "Shop 1",
				Slug:     "shop-1",
			},
			{
				ID:       uuid.New(),
				ClientID: clientID,
				Name:     "Shop 2",
				Slug:     "shop-2",
			},
		}

		page := 1
		pageSize := 20
		offset := 0

		mockShopRepo.On("ListByClient", ctx, clientID, pageSize, offset).Return(expectedShops, nil)

		result, total, err := service.ListShops(ctx, clientID, page, pageSize)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result, 2)
		assert.Equal(t, 0, total)
		mockShopRepo.AssertExpectations(t)
	})

	t.Run("success with page 2", func(t *testing.T) {
		mockShopRepo := new(MockShopRepository)
		mockClientRepo := new(MockClientRepository)
		service := NewShopService(mockShopRepo, mockClientRepo)

		expectedShops := []*models.Shop{
			{
				ID:       uuid.New(),
				ClientID: clientID,
				Name:     "Shop 3",
				Slug:     "shop-3",
			},
		}

		page := 2
		pageSize := 20
		offset := 20

		mockShopRepo.On("ListByClient", ctx, clientID, pageSize, offset).Return(expectedShops, nil)

		result, total, err := service.ListShops(ctx, clientID, page, pageSize)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result, 1)
		assert.Equal(t, 0, total)
		mockShopRepo.AssertExpectations(t)
	})

	t.Run("default page size when invalid", func(t *testing.T) {
		mockShopRepo := new(MockShopRepository)
		mockClientRepo := new(MockClientRepository)
		service := NewShopService(mockShopRepo, mockClientRepo)

		page := 1
		pageSize := 0 // Invalid
		defaultPageSize := 20
		offset := 0

		mockShopRepo.On("ListByClient", ctx, clientID, defaultPageSize, offset).Return([]*models.Shop{}, nil)

		result, total, err := service.ListShops(ctx, clientID, page, pageSize)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, 0, total)
		mockShopRepo.AssertExpectations(t)
	})

	t.Run("page size capped at 100", func(t *testing.T) {
		mockShopRepo := new(MockShopRepository)
		mockClientRepo := new(MockClientRepository)
		service := NewShopService(mockShopRepo, mockClientRepo)

		page := 1
		pageSize := 150 // Too large
		maxPageSize := 20 // Will be capped to default
		offset := 0

		mockShopRepo.On("ListByClient", ctx, clientID, maxPageSize, offset).Return([]*models.Shop{}, nil)

		result, total, err := service.ListShops(ctx, clientID, page, pageSize)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, 0, total)
		mockShopRepo.AssertExpectations(t)
	})

	t.Run("default page to 1 when invalid", func(t *testing.T) {
		mockShopRepo := new(MockShopRepository)
		mockClientRepo := new(MockClientRepository)
		service := NewShopService(mockShopRepo, mockClientRepo)

		page := 0 // Invalid
		pageSize := 20
		offset := 0 // Page 1

		mockShopRepo.On("ListByClient", ctx, clientID, pageSize, offset).Return([]*models.Shop{}, nil)

		result, total, err := service.ListShops(ctx, clientID, page, pageSize)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, 0, total)
		mockShopRepo.AssertExpectations(t)
	})
}

func TestRemoveUserFromShop(t *testing.T) {
	ctx := context.Background()
	shopID := uuid.New()
	clientUserRoleID := 123

	t.Run("success", func(t *testing.T) {
		mockShopRepo := new(MockShopRepository)
		mockClientRepo := new(MockClientRepository)
		service := NewShopService(mockShopRepo, mockClientRepo)

		mockShopRepo.On("RemoveUser", ctx, shopID, clientUserRoleID).Return(nil)

		err := service.RemoveUserFromShop(ctx, shopID, clientUserRoleID)

		assert.NoError(t, err)
		mockShopRepo.AssertExpectations(t)
	})

	t.Run("repository error", func(t *testing.T) {
		mockShopRepo := new(MockShopRepository)
		mockClientRepo := new(MockClientRepository)
		service := NewShopService(mockShopRepo, mockClientRepo)

		mockShopRepo.On("RemoveUser", ctx, shopID, clientUserRoleID).Return(fmt.Errorf("database error"))

		err := service.RemoveUserFromShop(ctx, shopID, clientUserRoleID)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to remove user from shop")
		mockShopRepo.AssertExpectations(t)
	})
}

func TestGetShopUsers(t *testing.T) {
	ctx := context.Background()
	shopID := uuid.New()
	clientID := uuid.New()

	t.Run("success", func(t *testing.T) {
		mockShopRepo := new(MockShopRepository)
		mockClientRepo := new(MockClientRepository)
		service := NewShopService(mockShopRepo, mockClientRepo)

		expectedUsers := []*models.ShopUser{
			{
				ID:               1,
				ClientID:         clientID,
				ShopID:           shopID,
				ClientUserRoleID: 10,
			},
			{
				ID:               2,
				ClientID:         clientID,
				ShopID:           shopID,
				ClientUserRoleID: 11,
			},
		}

		mockShopRepo.On("GetShopUsers", ctx, shopID).Return(expectedUsers, nil)

		result, err := service.GetShopUsers(ctx, shopID)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result, 2)
		assert.Equal(t, shopID, result[0].ShopID)
		assert.Equal(t, shopID, result[1].ShopID)
		mockShopRepo.AssertExpectations(t)
	})

	t.Run("empty list", func(t *testing.T) {
		mockShopRepo := new(MockShopRepository)
		mockClientRepo := new(MockClientRepository)
		service := NewShopService(mockShopRepo, mockClientRepo)

		mockShopRepo.On("GetShopUsers", ctx, shopID).Return([]*models.ShopUser{}, nil)

		result, err := service.GetShopUsers(ctx, shopID)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result, 0)
		mockShopRepo.AssertExpectations(t)
	})

	t.Run("repository error", func(t *testing.T) {
		mockShopRepo := new(MockShopRepository)
		mockClientRepo := new(MockClientRepository)
		service := NewShopService(mockShopRepo, mockClientRepo)

		mockShopRepo.On("GetShopUsers", ctx, shopID).Return(nil, fmt.Errorf("database error"))

		result, err := service.GetShopUsers(ctx, shopID)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "failed to get shop users")
		mockShopRepo.AssertExpectations(t)
	})
}
