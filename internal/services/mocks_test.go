package services

import (
	"context"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/yourorg/shoppilot/internal/models"
)

// MockClientRepository is a mock implementation of ClientRepository
type MockClientRepository struct {
	mock.Mock
}

func (m *MockClientRepository) Create(ctx context.Context, client *models.Client) error {
	args := m.Called(ctx, client)
	return args.Error(0)
}

func (m *MockClientRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Client, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Client), args.Error(1)
}

func (m *MockClientRepository) GetBySlug(ctx context.Context, slug string) (*models.Client, error) {
	args := m.Called(ctx, slug)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Client), args.Error(1)
}

func (m *MockClientRepository) Update(ctx context.Context, client *models.Client) error {
	args := m.Called(ctx, client)
	return args.Error(0)
}

func (m *MockClientRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockClientRepository) List(ctx context.Context, limit, offset int) ([]*models.Client, error) {
	args := m.Called(ctx, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Client), args.Error(1)
}

func (m *MockClientRepository) ListActive(ctx context.Context, limit, offset int) ([]*models.Client, error) {
	args := m.Called(ctx, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Client), args.Error(1)
}

// MockShopRepository is a mock implementation of ShopRepository
type MockShopRepository struct {
	mock.Mock
}

func (m *MockShopRepository) Create(ctx context.Context, shop *models.Shop) error {
	args := m.Called(ctx, shop)
	return args.Error(0)
}

func (m *MockShopRepository) GetByID(ctx context.Context, clientID uuid.UUID, shopID uuid.UUID) (*models.Shop, error) {
	args := m.Called(ctx, clientID, shopID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Shop), args.Error(1)
}

func (m *MockShopRepository) GetBySlug(ctx context.Context, clientID uuid.UUID, slug string) (*models.Shop, error) {
	args := m.Called(ctx, clientID, slug)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Shop), args.Error(1)
}

func (m *MockShopRepository) Update(ctx context.Context, shop *models.Shop) error {
	args := m.Called(ctx, shop)
	return args.Error(0)
}

func (m *MockShopRepository) Delete(ctx context.Context, clientID uuid.UUID, shopID uuid.UUID) error {
	args := m.Called(ctx, clientID, shopID)
	return args.Error(0)
}

func (m *MockShopRepository) ListByClient(ctx context.Context, clientID uuid.UUID, limit, offset int) ([]*models.Shop, error) {
	args := m.Called(ctx, clientID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Shop), args.Error(1)
}

func (m *MockShopRepository) ListActiveByClient(ctx context.Context, clientID uuid.UUID, limit, offset int) ([]*models.Shop, error) {
	args := m.Called(ctx, clientID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Shop), args.Error(1)
}

func (m *MockShopRepository) AssignUser(ctx context.Context, shopID uuid.UUID, clientUserRoleID int) error {
	args := m.Called(ctx, shopID, clientUserRoleID)
	return args.Error(0)
}

func (m *MockShopRepository) RemoveUser(ctx context.Context, shopID uuid.UUID, clientUserRoleID int) error {
	args := m.Called(ctx, shopID, clientUserRoleID)
	return args.Error(0)
}

func (m *MockShopRepository) GetShopUsers(ctx context.Context, shopID uuid.UUID) ([]*models.ShopUser, error) {
	args := m.Called(ctx, shopID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.ShopUser), args.Error(1)
}

// MockClientUserRepository is a mock implementation of ClientUserRepository
type MockClientUserRepository struct {
	mock.Mock
}

func (m *MockClientUserRepository) Create(ctx context.Context, user *models.ClientUser) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockClientUserRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.ClientUser, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.ClientUser), args.Error(1)
}

func (m *MockClientUserRepository) GetByEmail(ctx context.Context, clientID uuid.UUID, email string) (*models.ClientUser, error) {
	args := m.Called(ctx, clientID, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.ClientUser), args.Error(1)
}

func (m *MockClientUserRepository) GetByUsername(ctx context.Context, clientID uuid.UUID, username string) (*models.ClientUser, error) {
	args := m.Called(ctx, clientID, username)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.ClientUser), args.Error(1)
}

func (m *MockClientUserRepository) Update(ctx context.Context, user *models.ClientUser) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockClientUserRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockClientUserRepository) ListByClient(ctx context.Context, clientID uuid.UUID, limit, offset int) ([]*models.ClientUser, error) {
	args := m.Called(ctx, clientID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.ClientUser), args.Error(1)
}

func (m *MockClientUserRepository) AssignRole(ctx context.Context, userID uuid.UUID, roleID int) error {
	args := m.Called(ctx, userID, roleID)
	return args.Error(0)
}

func (m *MockClientUserRepository) RemoveRole(ctx context.Context, userID uuid.UUID, roleID int) error {
	args := m.Called(ctx, userID, roleID)
	return args.Error(0)
}

func (m *MockClientUserRepository) GetUserRoles(ctx context.Context, userID uuid.UUID) ([]*models.ClientRole, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.ClientRole), args.Error(1)
}

func (m *MockClientUserRepository) GetUserPermissions(ctx context.Context, userID uuid.UUID) ([]*models.ClientPermission, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.ClientPermission), args.Error(1)
}
