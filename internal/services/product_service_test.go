package services

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/yourorg/shoppilot/internal/models"
	"github.com/yourorg/shoppilot/internal/services/fakes"
)

// MockProductRepository is a mock implementation of ProductRepository
type MockProductRepository struct {
	mock.Mock
}

func (m *MockProductRepository) Create(ctx context.Context, product *models.Product) error {
	args := m.Called(ctx, product)
	return args.Error(0)
}

func (m *MockProductRepository) GetByID(ctx context.Context, clientID uuid.UUID, productID uuid.UUID) (*models.Product, error) {
	args := m.Called(ctx, clientID, productID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Product), args.Error(1)
}

func (m *MockProductRepository) GetByCode(ctx context.Context, clientID uuid.UUID, code string) (*models.Product, error) {
	args := m.Called(ctx, clientID, code)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Product), args.Error(1)
}

func (m *MockProductRepository) Update(ctx context.Context, product *models.Product) error {
	args := m.Called(ctx, product)
	return args.Error(0)
}

func (m *MockProductRepository) Delete(ctx context.Context, clientID uuid.UUID, productID uuid.UUID) error {
	args := m.Called(ctx, clientID, productID)
	return args.Error(0)
}

func (m *MockProductRepository) ListByShop(ctx context.Context, clientID uuid.UUID, shopID uuid.UUID, limit, offset int) ([]*models.Product, error) {
	args := m.Called(ctx, clientID, shopID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Product), args.Error(1)
}

func (m *MockProductRepository) ListByClient(ctx context.Context, clientID uuid.UUID, limit, offset int) ([]*models.Product, error) {
	args := m.Called(ctx, clientID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Product), args.Error(1)
}

func (m *MockProductRepository) Search(ctx context.Context, clientID uuid.UUID, query string, limit, offset int) ([]*models.Product, error) {
	args := m.Called(ctx, clientID, query, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Product), args.Error(1)
}

func (m *MockProductRepository) CreateVariant(ctx context.Context, variant *models.ProductVariant) error {
	args := m.Called(ctx, variant)
	return args.Error(0)
}

func (m *MockProductRepository) GetVariantByID(ctx context.Context, clientID uuid.UUID, variantID uuid.UUID) (*models.ProductVariant, error) {
	args := m.Called(ctx, clientID, variantID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.ProductVariant), args.Error(1)
}

func (m *MockProductRepository) GetVariantBySKU(ctx context.Context, clientID uuid.UUID, sku string) (*models.ProductVariant, error) {
	args := m.Called(ctx, clientID, sku)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.ProductVariant), args.Error(1)
}

func (m *MockProductRepository) UpdateVariant(ctx context.Context, variant *models.ProductVariant) error {
	args := m.Called(ctx, variant)
	return args.Error(0)
}

func (m *MockProductRepository) DeleteVariant(ctx context.Context, clientID uuid.UUID, variantID uuid.UUID) error {
	args := m.Called(ctx, clientID, variantID)
	return args.Error(0)
}

func (m *MockProductRepository) ListVariantsByProduct(ctx context.Context, clientID uuid.UUID, productID uuid.UUID) ([]*models.ProductVariant, error) {
	args := m.Called(ctx, clientID, productID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.ProductVariant), args.Error(1)
}

func (m *MockProductRepository) GetDefaultVariant(ctx context.Context, clientID uuid.UUID, productID uuid.UUID) (*models.ProductVariant, error) {
	args := m.Called(ctx, clientID, productID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.ProductVariant), args.Error(1)
}

func (m *MockProductRepository) UpdateInventory(ctx context.Context, clientID uuid.UUID, variantID uuid.UUID, quantity int) error {
	args := m.Called(ctx, clientID, variantID, quantity)
	return args.Error(0)
}

// Test CreateProduct

func TestProductService_CreateProduct_Success(t *testing.T) {
	mockRepo := new(MockProductRepository)
	service := NewProductService(mockRepo, nil, nil)

	clientID := uuid.New()
	shopID := uuid.New()

	req := &CreateProductRequest{
		ClientID:    clientID,
		ShopID:      shopID,
		Code:        "PROD-001",
		Name:        "Test Product",
		Description: "Test Description",
		IsActive:    true,
	}

	// Mock: Code doesn't exist (unique check)
	mockRepo.On("GetByCode", mock.Anything, clientID, "PROD-001").
		Return(nil, fmt.Errorf("not found"))

	// Mock: Create succeeds
	mockRepo.On("Create", mock.Anything, mock.AnythingOfType("*models.Product")).
		Return(nil)

	product, err := service.CreateProduct(context.Background(), req)

	assert.NoError(t, err)
	assert.NotNil(t, product)
	assert.Equal(t, "PROD-001", product.Code)
	assert.Equal(t, "Test Product", product.Name)
	mockRepo.AssertExpectations(t)
}

func TestProductService_CreateProduct_DuplicateCode(t *testing.T) {
	mockRepo := new(MockProductRepository)
	service := NewProductService(mockRepo, nil, nil)

	clientID := uuid.New()
	shopID := uuid.New()

	req := &CreateProductRequest{
		ClientID: clientID,
		ShopID:   shopID,
		Code:     "PROD-001",
		Name:     "Test Product",
		IsActive: true,
	}

	// Mock: Code already exists
	existingProduct := &models.Product{
		ID:       uuid.New(),
		ClientID: clientID,
		Code:     "PROD-001",
	}
	mockRepo.On("GetByCode", mock.Anything, clientID, "PROD-001").
		Return(existingProduct, nil)

	product, err := service.CreateProduct(context.Background(), req)

	assert.Error(t, err)
	assert.Nil(t, product)
	assert.Contains(t, err.Error(), "already exists")
	mockRepo.AssertExpectations(t)
}

// Test GetProduct

func TestProductService_GetProduct_Success(t *testing.T) {
	mockRepo := new(MockProductRepository)
	service := NewProductService(mockRepo, nil, nil)

	clientID := uuid.New()
	productID := uuid.New()

	expectedProduct := &models.Product{
		ID:       productID,
		ClientID: clientID,
		Code:     "PROD-001",
		Name:     "Test Product",
	}

	mockRepo.On("GetByID", mock.Anything, clientID, productID).
		Return(expectedProduct, nil)

	product, err := service.GetProduct(context.Background(), clientID, productID)

	assert.NoError(t, err)
	assert.NotNil(t, product)
	assert.Equal(t, productID, product.ID)
	mockRepo.AssertExpectations(t)
}

// Test UpdateProduct

func TestProductService_UpdateProduct_Success(t *testing.T) {
	mockRepo := new(MockProductRepository)
	service := NewProductService(mockRepo, nil, nil)

	clientID := uuid.New()
	productID := uuid.New()

	existingProduct := &models.Product{
		ID:       productID,
		ClientID: clientID,
		Code:     "PROD-001",
		Name:     "Old Name",
	}

	newName := "New Name"
	newDesc := "New Description"
	req := &UpdateProductRequest{
		Name:        &newName,
		Description: &newDesc,
	}

	mockRepo.On("GetByID", mock.Anything, clientID, productID).
		Return(existingProduct, nil)

	mockRepo.On("Update", mock.Anything, mock.AnythingOfType("*models.Product")).
		Return(nil)

	err := service.UpdateProduct(context.Background(), clientID, productID, req)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

// Test DeleteProduct

func TestProductService_DeleteProduct_Success(t *testing.T) {
	mockRepo := new(MockProductRepository)
	service := NewProductService(mockRepo, nil, nil)

	clientID := uuid.New()
	productID := uuid.New()

	mockRepo.On("Delete", mock.Anything, clientID, productID).
		Return(nil)

	err := service.DeleteProduct(context.Background(), clientID, productID)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

// Test ListProducts

func TestProductService_ListProducts_ByShop(t *testing.T) {
	mockRepo := new(MockProductRepository)
	service := NewProductService(mockRepo, nil, nil)

	clientID := uuid.New()
	shopID := uuid.New()

	expectedProducts := []*models.Product{
		{ID: uuid.New(), ClientID: clientID, ShopID: shopID, Name: "Product 1"},
		{ID: uuid.New(), ClientID: clientID, ShopID: shopID, Name: "Product 2"},
	}

	mockRepo.On("ListByShop", mock.Anything, clientID, shopID, 20, 0).
		Return(expectedProducts, nil)

	products, total, err := service.ListProducts(context.Background(), clientID, &shopID, 1, 20)

	assert.NoError(t, err)
	assert.Equal(t, 2, len(products))
	assert.Equal(t, 2, total)
	mockRepo.AssertExpectations(t)
}

func TestProductService_ListProducts_ByClient(t *testing.T) {
	mockRepo := new(MockProductRepository)
	service := NewProductService(mockRepo, nil, nil)

	clientID := uuid.New()

	expectedProducts := []*models.Product{
		{ID: uuid.New(), ClientID: clientID, Name: "Product 1"},
		{ID: uuid.New(), ClientID: clientID, Name: "Product 2"},
	}

	mockRepo.On("ListByClient", mock.Anything, clientID, 20, 0).
		Return(expectedProducts, nil)

	products, total, err := service.ListProducts(context.Background(), clientID, nil, 1, 20)

	assert.NoError(t, err)
	assert.Equal(t, 2, len(products))
	assert.Equal(t, 2, total)
	mockRepo.AssertExpectations(t)
}

// Test SearchProducts

func TestProductService_SearchProducts_Success(t *testing.T) {
	mockRepo := new(MockProductRepository)
	service := NewProductService(mockRepo, nil, nil)

	clientID := uuid.New()
	query := "test"

	expectedProducts := []*models.Product{
		{ID: uuid.New(), ClientID: clientID, Name: "Test Product"},
	}

	mockRepo.On("Search", mock.Anything, clientID, query, 20, 0).
		Return(expectedProducts, nil)

	products, total, err := service.SearchProducts(context.Background(), clientID, query, 1, 20)

	assert.NoError(t, err)
	assert.Equal(t, 1, len(products))
	assert.Equal(t, 1, total)
	mockRepo.AssertExpectations(t)
}

// Test CreateVariant

func TestProductService_CreateVariant_Success(t *testing.T) {
	mockRepo := new(MockProductRepository)
	service := NewProductService(mockRepo, nil, nil)

	clientID := uuid.New()
	shopID := uuid.New()
	productID := uuid.New()

	product := &models.Product{
		ID:       productID,
		ClientID: clientID,
		ShopID:   shopID,
	}

	req := &CreateVariantRequest{
		SKU:       "SKU-001",
		Name:      "Variant 1",
		Price:     99.99,
		Quantity:  10,
		IsDefault: false,
		IsActive:  true,
	}

	// Mock: Product exists
	mockRepo.On("GetByID", mock.Anything, clientID, productID).
		Return(product, nil)

	// Mock: SKU doesn't exist (unique check)
	mockRepo.On("GetVariantBySKU", mock.Anything, clientID, "SKU-001").
		Return(nil, fmt.Errorf("not found"))

	// Mock: Create succeeds
	mockRepo.On("CreateVariant", mock.Anything, mock.AnythingOfType("*models.ProductVariant")).
		Return(nil)

	variant, err := service.CreateVariant(context.Background(), clientID, productID, req)

	assert.NoError(t, err)
	assert.NotNil(t, variant)
	assert.Equal(t, "SKU-001", variant.SKU)
	mockRepo.AssertExpectations(t)
}

func TestProductService_CreateVariant_DuplicateSKU(t *testing.T) {
	mockRepo := new(MockProductRepository)
	service := NewProductService(mockRepo, nil, nil)

	clientID := uuid.New()
	productID := uuid.New()

	product := &models.Product{
		ID:       productID,
		ClientID: clientID,
	}

	req := &CreateVariantRequest{
		SKU:      "SKU-001",
		Name:     "Variant 1",
		Price:    99.99,
		Quantity: 10,
	}

	// Mock: Product exists
	mockRepo.On("GetByID", mock.Anything, clientID, productID).
		Return(product, nil)

	// Mock: SKU already exists
	existingVariant := &models.ProductVariant{
		ID:       uuid.New(),
		ClientID: clientID,
		SKU:      "SKU-001",
	}
	mockRepo.On("GetVariantBySKU", mock.Anything, clientID, "SKU-001").
		Return(existingVariant, nil)

	variant, err := service.CreateVariant(context.Background(), clientID, productID, req)

	assert.Error(t, err)
	assert.Nil(t, variant)
	assert.Contains(t, err.Error(), "already exists")
	mockRepo.AssertExpectations(t)
}

func TestProductService_CreateVariant_WithDefaultUnsetsPrevious(t *testing.T) {
	mockRepo := new(MockProductRepository)
	service := NewProductService(mockRepo, nil, nil)

	clientID := uuid.New()
	shopID := uuid.New()
	productID := uuid.New()

	product := &models.Product{
		ID:       productID,
		ClientID: clientID,
		ShopID:   shopID,
	}

	req := &CreateVariantRequest{
		SKU:       "SKU-002",
		Name:      "Variant 2",
		Price:     99.99,
		Quantity:  10,
		IsDefault: true,
		IsActive:  true,
	}

	// Existing default variant
	existingDefault := &models.ProductVariant{
		ID:        uuid.New(),
		ClientID:  clientID,
		ProductID: productID,
		SKU:       "SKU-001",
		IsDefault: true,
	}

	// Mock: Product exists
	mockRepo.On("GetByID", mock.Anything, clientID, productID).
		Return(product, nil)

	// Mock: SKU doesn't exist
	mockRepo.On("GetVariantBySKU", mock.Anything, clientID, "SKU-002").
		Return(nil, fmt.Errorf("not found"))

	// Mock: Default variant exists
	mockRepo.On("GetDefaultVariant", mock.Anything, clientID, productID).
		Return(existingDefault, nil)

	// Mock: Update existing default to unset is_default
	mockRepo.On("UpdateVariant", mock.Anything, existingDefault).
		Return(nil)

	// Mock: Create new variant
	mockRepo.On("CreateVariant", mock.Anything, mock.AnythingOfType("*models.ProductVariant")).
		Return(nil)

	variant, err := service.CreateVariant(context.Background(), clientID, productID, req)

	assert.NoError(t, err)
	assert.NotNil(t, variant)
	assert.True(t, variant.IsDefault)
	mockRepo.AssertExpectations(t)
}

// Test UpdateVariant

func TestProductService_UpdateVariant_Success(t *testing.T) {
	mockRepo := new(MockProductRepository)
	service := NewProductService(mockRepo, nil, nil)

	clientID := uuid.New()
	variantID := uuid.New()

	existingVariant := &models.ProductVariant{
		ID:       variantID,
		ClientID: clientID,
		SKU:      "SKU-001",
		Name:     "Old Name",
		Price:    50.0,
		Quantity: 10,
	}

	newName := "New Name"
	newPrice := 75.0
	req := &UpdateVariantRequest{
		Name:  &newName,
		Price: &newPrice,
	}

	mockRepo.On("GetVariantByID", mock.Anything, clientID, variantID).
		Return(existingVariant, nil)

	mockRepo.On("UpdateVariant", mock.Anything, mock.AnythingOfType("*models.ProductVariant")).
		Return(nil)

	err := service.UpdateVariant(context.Background(), clientID, variantID, req)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestProductService_UpdateVariant_NegativeInventory(t *testing.T) {
	mockRepo := new(MockProductRepository)
	service := NewProductService(mockRepo, nil, nil)

	clientID := uuid.New()
	variantID := uuid.New()

	existingVariant := &models.ProductVariant{
		ID:       variantID,
		ClientID: clientID,
		Quantity: 10,
	}

	negativeQty := -5
	req := &UpdateVariantRequest{
		Quantity: &negativeQty,
	}

	mockRepo.On("GetVariantByID", mock.Anything, clientID, variantID).
		Return(existingVariant, nil)

	err := service.UpdateVariant(context.Background(), clientID, variantID, req)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot be negative")
	mockRepo.AssertExpectations(t)
}

func TestProductService_UpdateVariant_SetDefaultUnsetsPrevious(t *testing.T) {
	mockRepo := new(MockProductRepository)
	service := NewProductService(mockRepo, nil, nil)

	clientID := uuid.New()
	productID := uuid.New()
	variantID := uuid.New()

	existingVariant := &models.ProductVariant{
		ID:        variantID,
		ClientID:  clientID,
		ProductID: productID,
		SKU:       "SKU-002",
		IsDefault: false,
	}

	existingDefault := &models.ProductVariant{
		ID:        uuid.New(),
		ClientID:  clientID,
		ProductID: productID,
		SKU:       "SKU-001",
		IsDefault: true,
	}

	isDefault := true
	req := &UpdateVariantRequest{
		IsDefault: &isDefault,
	}

	mockRepo.On("GetVariantByID", mock.Anything, clientID, variantID).
		Return(existingVariant, nil)

	mockRepo.On("GetDefaultVariant", mock.Anything, clientID, productID).
		Return(existingDefault, nil)

	mockRepo.On("UpdateVariant", mock.Anything, existingDefault).
		Return(nil)

	mockRepo.On("UpdateVariant", mock.Anything, existingVariant).
		Return(nil)

	err := service.UpdateVariant(context.Background(), clientID, variantID, req)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

// Test DeleteVariant

func TestProductService_DeleteVariant_Success(t *testing.T) {
	mockRepo := new(MockProductRepository)
	service := NewProductService(mockRepo, nil, nil)

	clientID := uuid.New()
	productID := uuid.New()
	variantID := uuid.New()

	variant := &models.ProductVariant{
		ID:        variantID,
		ClientID:  clientID,
		ProductID: productID,
	}

	// Mock: Variant exists
	mockRepo.On("GetVariantByID", mock.Anything, clientID, variantID).
		Return(variant, nil)

	// Mock: Product has multiple variants (>1)
	variants := []*models.ProductVariant{
		{ID: variantID},
		{ID: uuid.New()},
	}
	mockRepo.On("ListVariantsByProduct", mock.Anything, clientID, productID).
		Return(variants, nil)

	// Mock: Delete succeeds
	mockRepo.On("DeleteVariant", mock.Anything, clientID, variantID).
		Return(nil)

	err := service.DeleteVariant(context.Background(), clientID, variantID)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestProductService_DeleteVariant_LastVariant(t *testing.T) {
	mockRepo := new(MockProductRepository)
	service := NewProductService(mockRepo, nil, nil)

	clientID := uuid.New()
	productID := uuid.New()
	variantID := uuid.New()

	variant := &models.ProductVariant{
		ID:        variantID,
		ClientID:  clientID,
		ProductID: productID,
	}

	// Mock: Variant exists
	mockRepo.On("GetVariantByID", mock.Anything, clientID, variantID).
		Return(variant, nil)

	// Mock: Product has only one variant
	variants := []*models.ProductVariant{
		{ID: variantID},
	}
	mockRepo.On("ListVariantsByProduct", mock.Anything, clientID, productID).
		Return(variants, nil)

	err := service.DeleteVariant(context.Background(), clientID, variantID)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "at least one variant")
	mockRepo.AssertExpectations(t)
}

// Test ListVariants

func TestProductService_ListVariants_Success(t *testing.T) {
	mockRepo := new(MockProductRepository)
	service := NewProductService(mockRepo, nil, nil)

	clientID := uuid.New()
	productID := uuid.New()

	product := &models.Product{
		ID:       productID,
		ClientID: clientID,
	}

	expectedVariants := []*models.ProductVariant{
		{ID: uuid.New(), ProductID: productID, SKU: "SKU-001"},
		{ID: uuid.New(), ProductID: productID, SKU: "SKU-002"},
	}

	mockRepo.On("GetByID", mock.Anything, clientID, productID).
		Return(product, nil)

	mockRepo.On("ListVariantsByProduct", mock.Anything, clientID, productID).
		Return(expectedVariants, nil)

	variants, err := service.ListVariants(context.Background(), clientID, productID)

	assert.NoError(t, err)
	assert.Equal(t, 2, len(variants))
	mockRepo.AssertExpectations(t)
}

// Test AdjustInventory

func TestProductService_AdjustInventory_Increase(t *testing.T) {
	mockRepo := new(MockProductRepository)
	fakeMovementRepo := &fakes.FakeInventoryMovementRepository{}
	service := NewProductService(mockRepo, fakeMovementRepo, nil)

	clientID := uuid.New()
	variantID := uuid.New()
	shopID := uuid.New()

	variant := &models.ProductVariant{
		ID:       variantID,
		ClientID: clientID,
		ShopID:   shopID,
		Quantity: 10,
	}

	mockRepo.On("GetVariantByID", mock.Anything, clientID, variantID).
		Return(variant, nil)

	mockRepo.On("UpdateInventory", mock.Anything, clientID, variantID, 15).
		Return(nil)

	fakeMovementRepo.CreateReturns(nil)

	err := service.AdjustInventory(context.Background(), clientID, variantID, 5)

	assert.NoError(t, err)
	assert.Equal(t, 1, fakeMovementRepo.CreateCallCount())
	mockRepo.AssertExpectations(t)
}

func TestProductService_AdjustInventory_Decrease(t *testing.T) {
	mockRepo := new(MockProductRepository)
	fakeMovementRepo := &fakes.FakeInventoryMovementRepository{}
	service := NewProductService(mockRepo, fakeMovementRepo, nil)

	clientID := uuid.New()
	variantID := uuid.New()
	shopID := uuid.New()

	variant := &models.ProductVariant{
		ID:       variantID,
		ClientID: clientID,
		ShopID:   shopID,
		Quantity: 10,
	}

	mockRepo.On("GetVariantByID", mock.Anything, clientID, variantID).
		Return(variant, nil)

	mockRepo.On("UpdateInventory", mock.Anything, clientID, variantID, 5).
		Return(nil)

	fakeMovementRepo.CreateReturns(nil)

	err := service.AdjustInventory(context.Background(), clientID, variantID, -5)

	assert.NoError(t, err)
	assert.Equal(t, 1, fakeMovementRepo.CreateCallCount())
	mockRepo.AssertExpectations(t)
}

func TestProductService_AdjustInventory_Negative(t *testing.T) {
	mockRepo := new(MockProductRepository)
	service := NewProductService(mockRepo, nil, nil)

	clientID := uuid.New()
	variantID := uuid.New()

	variant := &models.ProductVariant{
		ID:       variantID,
		ClientID: clientID,
		Quantity: 5,
	}

	mockRepo.On("GetVariantByID", mock.Anything, clientID, variantID).
		Return(variant, nil)

	err := service.AdjustInventory(context.Background(), clientID, variantID, -10)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "insufficient inventory")
	mockRepo.AssertExpectations(t)
}

// Test SetInventory

func TestProductService_SetInventory_Success(t *testing.T) {
	mockRepo := new(MockProductRepository)
	fakeMovementRepo := &fakes.FakeInventoryMovementRepository{}
	service := NewProductService(mockRepo, fakeMovementRepo, nil)

	clientID := uuid.New()
	variantID := uuid.New()
	shopID := uuid.New()

	variant := &models.ProductVariant{
		ID:       variantID,
		ClientID: clientID,
		ShopID:   shopID,
		Quantity: 50,
	}

	mockRepo.On("GetVariantByID", mock.Anything, clientID, variantID).
		Return(variant, nil)
	mockRepo.On("UpdateInventory", mock.Anything, clientID, variantID, 100).
		Return(nil)
	fakeMovementRepo.CreateReturns(nil)

	err := service.SetInventory(context.Background(), clientID, variantID, 100)

	assert.NoError(t, err)
	assert.Equal(t, 1, fakeMovementRepo.CreateCallCount())
	mockRepo.AssertExpectations(t)
}

func TestProductService_SetInventory_Negative(t *testing.T) {
	mockRepo := new(MockProductRepository)
	service := NewProductService(mockRepo, nil, nil)

	clientID := uuid.New()
	variantID := uuid.New()

	err := service.SetInventory(context.Background(), clientID, variantID, -10)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot be negative")
	mockRepo.AssertExpectations(t)
}

// Test CheckStock

func TestProductService_CheckStock_Success(t *testing.T) {
	mockRepo := new(MockProductRepository)
	service := NewProductService(mockRepo, nil, nil)

	clientID := uuid.New()
	variantID := uuid.New()

	variant := &models.ProductVariant{
		ID:       variantID,
		ClientID: clientID,
		Quantity: 42,
	}

	mockRepo.On("GetVariantByID", mock.Anything, clientID, variantID).
		Return(variant, nil)

	quantity, err := service.CheckStock(context.Background(), clientID, variantID)

	assert.NoError(t, err)
	assert.Equal(t, 42, quantity)
	mockRepo.AssertExpectations(t)
}
