package services

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yourorg/shoppilot/internal/models"
	"github.com/yourorg/shoppilot/internal/services/fakes"
)

func TestSalesOrderService_CreateSalesOrder_Success(t *testing.T) {
	fakeSORepo := &fakes.FakeSalesOrderRepository{}
	fakeProductRepo := &fakes.FakeProductRepository{}
	fakeMovementRepo := &fakes.FakeInventoryMovementRepository{}

	service := NewSalesOrderService(fakeSORepo, fakeProductRepo, fakeMovementRepo)

	clientID := uuid.New()
	customerID := uuid.New()
	shopID := uuid.New()

	// Configure fake - order number doesn't exist
	fakeSORepo.GetByOrderNumberReturns(nil, errors.New("not found"))
	fakeSORepo.CreateReturns(nil)
	fakeSORepo.CreateItemReturns(nil)

	req := &CreateSalesOrderRequest{
		CustomerID:  customerID,
		ShopID:      shopID,
		OrderNumber: "SO-12345",
		OrderDate:   time.Now(),
		Subtotal:    900.00,
		TaxAmount:   100.00,
		TotalAmount: 1000.00,
		Items: []CreateSalesOrderItemRequest{
			{
				VariantID:       uuid.New(),
				QuantityOrdered: 10,
				UnitPrice:       50.00,
				TotalPrice:      500.00,
			},
		},
	}

	so, err := service.CreateSalesOrder(context.Background(), clientID, req)
	require.NoError(t, err)
	assert.NotNil(t, so)
	assert.Equal(t, "SO-12345", so.OrderNumber)
	assert.Equal(t, models.SOStatusPending, so.Status)

	// Verify Create was called
	assert.Equal(t, 1, fakeSORepo.CreateCallCount())
	assert.Equal(t, 1, fakeSORepo.CreateItemCallCount())
}

func TestSalesOrderService_CreateSalesOrder_AutoGenerateOrderNumber(t *testing.T) {
	fakeSORepo := &fakes.FakeSalesOrderRepository{}
	fakeProductRepo := &fakes.FakeProductRepository{}
	fakeMovementRepo := &fakes.FakeInventoryMovementRepository{}

	service := NewSalesOrderService(fakeSORepo, fakeProductRepo, fakeMovementRepo)

	clientID := uuid.New()

	// Configure fake - no order number provided
	fakeSORepo.GetByOrderNumberReturns(nil, errors.New("not found"))
	fakeSORepo.CreateReturns(nil)

	req := &CreateSalesOrderRequest{
		CustomerID:  uuid.New(),
		ShopID:      uuid.New(),
		OrderNumber: "", // Empty - should auto-generate
		OrderDate:   time.Now(),
		TotalAmount: 1000.00,
	}

	so, err := service.CreateSalesOrder(context.Background(), clientID, req)
	require.NoError(t, err)
	assert.NotNil(t, so)
	assert.NotEmpty(t, so.OrderNumber)
	assert.Contains(t, so.OrderNumber, "SO-")
}

func TestSalesOrderService_CreateSalesOrder_DuplicateOrderNumber(t *testing.T) {
	fakeSORepo := &fakes.FakeSalesOrderRepository{}
	fakeProductRepo := &fakes.FakeProductRepository{}
	fakeMovementRepo := &fakes.FakeInventoryMovementRepository{}

	service := NewSalesOrderService(fakeSORepo, fakeProductRepo, fakeMovementRepo)

	clientID := uuid.New()
	existingSO := &models.SalesOrder{
		ID:          uuid.New(),
		ClientID:    clientID,
		OrderNumber: "SO-12345",
	}

	// Configure fake - order number already exists
	fakeSORepo.GetByOrderNumberReturns(existingSO, nil)

	req := &CreateSalesOrderRequest{
		CustomerID:  uuid.New(),
		ShopID:      uuid.New(),
		OrderNumber: "SO-12345",
		OrderDate:   time.Now(),
		TotalAmount: 1000.00,
	}

	so, err := service.CreateSalesOrder(context.Background(), clientID, req)
	require.Error(t, err)
	assert.Nil(t, so)
	assert.Contains(t, err.Error(), "already exists")

	// Create should not be called
	assert.Equal(t, 0, fakeSORepo.CreateCallCount())
}

func TestSalesOrderService_ConfirmSalesOrder_Success(t *testing.T) {
	fakeSORepo := &fakes.FakeSalesOrderRepository{}
	fakeProductRepo := &fakes.FakeProductRepository{}
	fakeMovementRepo := &fakes.FakeInventoryMovementRepository{}

	service := NewSalesOrderService(fakeSORepo, fakeProductRepo, fakeMovementRepo)

	clientID := uuid.New()
	soID := uuid.New()
	variantID := uuid.New()

	pendingSO := &models.SalesOrder{
		ID:       soID,
		ClientID: clientID,
		Status:   models.SOStatusPending,
	}

	item := &models.SalesOrderItem{
		ID:              uuid.New(),
		VariantID:       variantID,
		QuantityOrdered: 10,
	}

	variant := &models.ProductVariant{
		ID:       variantID,
		Quantity: 50, // Sufficient inventory
	}

	fakeSORepo.GetByIDReturns(pendingSO, nil)
	fakeSORepo.ListItemsBySOReturns([]*models.SalesOrderItem{item}, nil)
	fakeSORepo.UpdateReturns(nil)
	fakeProductRepo.GetVariantByIDReturns(variant, nil)

	err := service.ConfirmSalesOrder(context.Background(), clientID, soID)
	require.NoError(t, err)

	// Verify Update was called
	assert.Equal(t, 1, fakeSORepo.UpdateCallCount())

	// Verify status was changed to confirmed
	_, updatedSO := fakeSORepo.UpdateArgsForCall(0)
	assert.Equal(t, models.SOStatusConfirmed, updatedSO.Status)
}

func TestSalesOrderService_ConfirmSalesOrder_InsufficientInventory(t *testing.T) {
	fakeSORepo := &fakes.FakeSalesOrderRepository{}
	fakeProductRepo := &fakes.FakeProductRepository{}
	fakeMovementRepo := &fakes.FakeInventoryMovementRepository{}

	service := NewSalesOrderService(fakeSORepo, fakeProductRepo, fakeMovementRepo)

	clientID := uuid.New()
	soID := uuid.New()
	variantID := uuid.New()

	pendingSO := &models.SalesOrder{
		ID:       soID,
		ClientID: clientID,
		Status:   models.SOStatusPending,
	}

	item := &models.SalesOrderItem{
		ID:              uuid.New(),
		VariantID:       variantID,
		QuantityOrdered: 100, // Requires 100
	}

	variant := &models.ProductVariant{
		ID:       variantID,
		Quantity: 50, // Only 50 available
	}

	fakeSORepo.GetByIDReturns(pendingSO, nil)
	fakeSORepo.ListItemsBySOReturns([]*models.SalesOrderItem{item}, nil)
	fakeProductRepo.GetVariantByIDReturns(variant, nil)

	err := service.ConfirmSalesOrder(context.Background(), clientID, soID)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "insufficient inventory")

	// Update should not be called
	assert.Equal(t, 0, fakeSORepo.UpdateCallCount())
}

func TestSalesOrderService_ConfirmSalesOrder_OnlyPending(t *testing.T) {
	fakeSORepo := &fakes.FakeSalesOrderRepository{}
	fakeProductRepo := &fakes.FakeProductRepository{}
	fakeMovementRepo := &fakes.FakeInventoryMovementRepository{}

	service := NewSalesOrderService(fakeSORepo, fakeProductRepo, fakeMovementRepo)

	clientID := uuid.New()
	soID := uuid.New()

	confirmedSO := &models.SalesOrder{
		ID:       soID,
		ClientID: clientID,
		Status:   models.SOStatusConfirmed,
	}

	fakeSORepo.GetByIDReturns(confirmedSO, nil)

	err := service.ConfirmSalesOrder(context.Background(), clientID, soID)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "only pending")

	// Update should not be called
	assert.Equal(t, 0, fakeSORepo.UpdateCallCount())
}

func TestSalesOrderService_CancelSalesOrder(t *testing.T) {
	fakeSORepo := &fakes.FakeSalesOrderRepository{}
	fakeProductRepo := &fakes.FakeProductRepository{}
	fakeMovementRepo := &fakes.FakeInventoryMovementRepository{}

	service := NewSalesOrderService(fakeSORepo, fakeProductRepo, fakeMovementRepo)

	clientID := uuid.New()
	soID := uuid.New()

	confirmedSO := &models.SalesOrder{
		ID:       soID,
		ClientID: clientID,
		Status:   models.SOStatusConfirmed,
	}

	fakeSORepo.GetByIDReturns(confirmedSO, nil)
	fakeSORepo.UpdateReturns(nil)

	err := service.CancelSalesOrder(context.Background(), clientID, soID)
	require.NoError(t, err)

	// Verify Update was called
	assert.Equal(t, 1, fakeSORepo.UpdateCallCount())

	// Verify status was changed to cancelled
	_, updatedSO := fakeSORepo.UpdateArgsForCall(0)
	assert.Equal(t, models.SOStatusCancelled, updatedSO.Status)
}

func TestSalesOrderService_AddItem(t *testing.T) {
	fakeSORepo := &fakes.FakeSalesOrderRepository{}
	fakeProductRepo := &fakes.FakeProductRepository{}
	fakeMovementRepo := &fakes.FakeInventoryMovementRepository{}

	service := NewSalesOrderService(fakeSORepo, fakeProductRepo, fakeMovementRepo)

	clientID := uuid.New()
	soID := uuid.New()

	pendingSO := &models.SalesOrder{
		ID:       soID,
		ClientID: clientID,
		Status:   models.SOStatusPending,
	}

	fakeSORepo.GetByIDReturns(pendingSO, nil)
	fakeSORepo.CreateItemReturns(nil)

	req := &AddSalesOrderItemRequest{
		VariantID:       uuid.New(),
		QuantityOrdered: 10,
		UnitPrice:       50.00,
	}

	item, err := service.AddItem(context.Background(), clientID, soID, req)
	require.NoError(t, err)
	assert.NotNil(t, item)
	assert.Equal(t, 1, fakeSORepo.CreateItemCallCount())
}

func TestSalesOrderService_AddItem_OnlyPending(t *testing.T) {
	fakeSORepo := &fakes.FakeSalesOrderRepository{}
	fakeProductRepo := &fakes.FakeProductRepository{}
	fakeMovementRepo := &fakes.FakeInventoryMovementRepository{}

	service := NewSalesOrderService(fakeSORepo, fakeProductRepo, fakeMovementRepo)

	clientID := uuid.New()
	soID := uuid.New()

	confirmedSO := &models.SalesOrder{
		ID:       soID,
		ClientID: clientID,
		Status:   models.SOStatusConfirmed,
	}

	fakeSORepo.GetByIDReturns(confirmedSO, nil)

	req := &AddSalesOrderItemRequest{
		VariantID:       uuid.New(),
		QuantityOrdered: 10,
		UnitPrice:       50.00,
	}

	item, err := service.AddItem(context.Background(), clientID, soID, req)
	require.Error(t, err)
	assert.Nil(t, item)
	assert.Contains(t, err.Error(), "pending")

	// CreateItem should not be called
	assert.Equal(t, 0, fakeSORepo.CreateItemCallCount())
}

func TestSalesOrderService_RemoveItem(t *testing.T) {
	fakeSORepo := &fakes.FakeSalesOrderRepository{}
	fakeProductRepo := &fakes.FakeProductRepository{}
	fakeMovementRepo := &fakes.FakeInventoryMovementRepository{}

	service := NewSalesOrderService(fakeSORepo, fakeProductRepo, fakeMovementRepo)

	clientID := uuid.New()
	soID := uuid.New()
	itemID := uuid.New()

	pendingSO := &models.SalesOrder{
		ID:       soID,
		ClientID: clientID,
		Status:   models.SOStatusPending,
	}

	fakeSORepo.GetByIDReturns(pendingSO, nil)
	fakeSORepo.DeleteItemReturns(nil)

	err := service.RemoveItem(context.Background(), clientID, soID, itemID)
	require.NoError(t, err)
	assert.Equal(t, 1, fakeSORepo.DeleteItemCallCount())
}

func TestSalesOrderService_FulfillItems_Success(t *testing.T) {
	fakeSORepo := &fakes.FakeSalesOrderRepository{}
	fakeProductRepo := &fakes.FakeProductRepository{}
	fakeMovementRepo := &fakes.FakeInventoryMovementRepository{}

	service := NewSalesOrderService(fakeSORepo, fakeProductRepo, fakeMovementRepo)

	clientID := uuid.New()
	soID := uuid.New()
	shopID := uuid.New()
	itemID := uuid.New()
	variantID := uuid.New()

	confirmedSO := &models.SalesOrder{
		ID:          soID,
		ClientID:    clientID,
		ShopID:      shopID,
		OrderNumber: "SO-12345",
		Status:      models.SOStatusConfirmed,
	}

	item := &models.SalesOrderItem{
		ID:                itemID,
		ClientID:          clientID,
		SalesOrderID:      soID,
		VariantID:         variantID,
		QuantityOrdered:   10,
		QuantityFulfilled: 0,
	}

	variant := &models.ProductVariant{
		ID:       variantID,
		ClientID: clientID,
		Quantity: 50,
	}

	allItems := []*models.SalesOrderItem{item}

	// Configure fakes
	fakeSORepo.GetByIDReturns(confirmedSO, nil)
	fakeSORepo.GetItemReturns(item, nil)
	fakeSORepo.FulfillItemReturns(nil)
	fakeSORepo.ListItemsBySOReturns(allItems, nil)
	fakeSORepo.UpdateReturns(nil)
	fakeProductRepo.GetVariantByIDReturns(variant, nil)
	fakeProductRepo.UpdateInventoryReturns(nil)
	fakeMovementRepo.CreateReturns(nil)

	fulfillReqs := []FulfillItemRequest{
		{
			ItemID:            itemID,
			QuantityFulfilled: 5,
		},
	}

	err := service.FulfillItems(context.Background(), clientID, soID, fulfillReqs)
	require.NoError(t, err)

	// Verify all operations were called in order
	assert.Equal(t, 1, fakeSORepo.FulfillItemCallCount())
	assert.Equal(t, 1, fakeProductRepo.UpdateInventoryCallCount())
	assert.Equal(t, 1, fakeMovementRepo.CreateCallCount())

	// Verify inventory was updated correctly (decreased)
	_, _, _, newQty := fakeProductRepo.UpdateInventoryArgsForCall(0)
	assert.Equal(t, 45, newQty) // 50 - 5

	// Verify movement was created with correct values
	_, movement := fakeMovementRepo.CreateArgsForCall(0)
	assert.Equal(t, models.MovementTypeSale, movement.MovementType)
	assert.Equal(t, -5, movement.Quantity) // Negative for sales
	assert.Equal(t, 50, movement.PreviousQuantity)
	assert.Equal(t, 45, movement.NewQuantity)
}

func TestSalesOrderService_FulfillItems_InsufficientInventory(t *testing.T) {
	fakeSORepo := &fakes.FakeSalesOrderRepository{}
	fakeProductRepo := &fakes.FakeProductRepository{}
	fakeMovementRepo := &fakes.FakeInventoryMovementRepository{}

	service := NewSalesOrderService(fakeSORepo, fakeProductRepo, fakeMovementRepo)

	clientID := uuid.New()
	soID := uuid.New()
	itemID := uuid.New()
	variantID := uuid.New()

	confirmedSO := &models.SalesOrder{
		ID:       soID,
		ClientID: clientID,
		Status:   models.SOStatusConfirmed,
	}

	item := &models.SalesOrderItem{
		ID:                itemID,
		SalesOrderID:      soID,
		VariantID:         variantID,
		QuantityOrdered:   10,
		QuantityFulfilled: 0,
	}

	variant := &models.ProductVariant{
		ID:       variantID,
		Quantity: 3, // Only 3 available
	}

	fakeSORepo.GetByIDReturns(confirmedSO, nil)
	fakeSORepo.GetItemReturns(item, nil)
	fakeProductRepo.GetVariantByIDReturns(variant, nil)

	fulfillReqs := []FulfillItemRequest{
		{
			ItemID:            itemID,
			QuantityFulfilled: 5, // Trying to fulfill 5
		},
	}

	err := service.FulfillItems(context.Background(), clientID, soID, fulfillReqs)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "insufficient inventory")

	// No updates should happen
	assert.Equal(t, 0, fakeSORepo.FulfillItemCallCount())
	assert.Equal(t, 0, fakeProductRepo.UpdateInventoryCallCount())
	assert.Equal(t, 0, fakeMovementRepo.CreateCallCount())
}

func TestSalesOrderService_FulfillItems_ExceedsOrdered(t *testing.T) {
	fakeSORepo := &fakes.FakeSalesOrderRepository{}
	fakeProductRepo := &fakes.FakeProductRepository{}
	fakeMovementRepo := &fakes.FakeInventoryMovementRepository{}

	service := NewSalesOrderService(fakeSORepo, fakeProductRepo, fakeMovementRepo)

	clientID := uuid.New()
	soID := uuid.New()
	itemID := uuid.New()
	variantID := uuid.New()

	confirmedSO := &models.SalesOrder{
		ID:       soID,
		ClientID: clientID,
		Status:   models.SOStatusConfirmed,
	}

	item := &models.SalesOrderItem{
		ID:                itemID,
		SalesOrderID:      soID,
		VariantID:         variantID,
		QuantityOrdered:   10,
		QuantityFulfilled: 8, // Already fulfilled 8
	}

	variant := &models.ProductVariant{
		ID:       variantID,
		Quantity: 50,
	}

	fakeSORepo.GetByIDReturns(confirmedSO, nil)
	fakeSORepo.GetItemReturns(item, nil)
	fakeProductRepo.GetVariantByIDReturns(variant, nil)

	fulfillReqs := []FulfillItemRequest{
		{
			ItemID:            itemID,
			QuantityFulfilled: 5, // Would exceed ordered (8+5 > 10)
		},
	}

	err := service.FulfillItems(context.Background(), clientID, soID, fulfillReqs)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "cannot fulfill more than ordered")

	// No updates should happen
	assert.Equal(t, 0, fakeSORepo.FulfillItemCallCount())
}

func TestSalesOrderService_FulfillItems_PartiallyFulfilled(t *testing.T) {
	fakeSORepo := &fakes.FakeSalesOrderRepository{}
	fakeProductRepo := &fakes.FakeProductRepository{}
	fakeMovementRepo := &fakes.FakeInventoryMovementRepository{}

	service := NewSalesOrderService(fakeSORepo, fakeProductRepo, fakeMovementRepo)

	clientID := uuid.New()
	soID := uuid.New()
	shopID := uuid.New()
	itemID := uuid.New()
	variantID := uuid.New()

	confirmedSO := &models.SalesOrder{
		ID:          soID,
		ClientID:    clientID,
		ShopID:      shopID,
		OrderNumber: "SO-12345",
		Status:      models.SOStatusConfirmed,
	}

	item := &models.SalesOrderItem{
		ID:                itemID,
		ClientID:          clientID,
		SalesOrderID:      soID,
		VariantID:         variantID,
		QuantityOrdered:   10,
		QuantityFulfilled: 5, // Partially fulfilled
	}

	variant := &models.ProductVariant{
		ID:       variantID,
		ClientID: clientID,
		Quantity: 50,
	}

	allItems := []*models.SalesOrderItem{item}

	// Configure fakes
	fakeSORepo.GetByIDReturns(confirmedSO, nil)
	fakeSORepo.GetItemReturns(item, nil)
	fakeSORepo.FulfillItemReturns(nil)
	fakeSORepo.ListItemsBySOReturns(allItems, nil)
	fakeSORepo.UpdateReturns(nil)
	fakeProductRepo.GetVariantByIDReturns(variant, nil)
	fakeProductRepo.UpdateInventoryReturns(nil)
	fakeMovementRepo.CreateReturns(nil)

	fulfillReqs := []FulfillItemRequest{
		{
			ItemID:            itemID,
			QuantityFulfilled: 3,
		},
	}

	err := service.FulfillItems(context.Background(), clientID, soID, fulfillReqs)
	require.NoError(t, err)

	// Verify SO status was updated to partially fulfilled
	assert.Equal(t, 1, fakeSORepo.UpdateCallCount())
	_, updatedSO := fakeSORepo.UpdateArgsForCall(0)
	assert.Equal(t, models.SOStatusPartiallyFulfilled, updatedSO.Status)
}

func TestSalesOrderService_FulfillItems_FullyFulfilled(t *testing.T) {
	fakeSORepo := &fakes.FakeSalesOrderRepository{}
	fakeProductRepo := &fakes.FakeProductRepository{}
	fakeMovementRepo := &fakes.FakeInventoryMovementRepository{}

	service := NewSalesOrderService(fakeSORepo, fakeProductRepo, fakeMovementRepo)

	clientID := uuid.New()
	soID := uuid.New()
	shopID := uuid.New()
	itemID := uuid.New()
	variantID := uuid.New()

	confirmedSO := &models.SalesOrder{
		ID:          soID,
		ClientID:    clientID,
		ShopID:      shopID,
		OrderNumber: "SO-12345",
		Status:      models.SOStatusConfirmed,
	}

	item := &models.SalesOrderItem{
		ID:                itemID,
		ClientID:          clientID,
		SalesOrderID:      soID,
		VariantID:         variantID,
		QuantityOrdered:   10,
		QuantityFulfilled: 0, // Not yet fulfilled, will be after this operation
	}

	variant := &models.ProductVariant{
		ID:       variantID,
		ClientID: clientID,
		Quantity: 50,
	}

	// After fulfillment, the item will have QuantityFulfilled: 10
	fulfilledItem := &models.SalesOrderItem{
		ID:                itemID,
		ClientID:          clientID,
		SalesOrderID:      soID,
		VariantID:         variantID,
		QuantityOrdered:   10,
		QuantityFulfilled: 10, // After the FulfillItem call
	}
	allItems := []*models.SalesOrderItem{fulfilledItem}

	// Configure fakes
	fakeSORepo.GetByIDReturns(confirmedSO, nil)
	fakeSORepo.GetItemReturns(item, nil)
	fakeSORepo.FulfillItemReturns(nil)
	fakeSORepo.ListItemsBySOReturns(allItems, nil)
	fakeSORepo.UpdateReturns(nil)
	fakeProductRepo.GetVariantByIDReturns(variant, nil)
	fakeProductRepo.UpdateInventoryReturns(nil)
	fakeMovementRepo.CreateReturns(nil)

	fulfillReqs := []FulfillItemRequest{
		{
			ItemID:            itemID,
			QuantityFulfilled: 10,
		},
	}

	err := service.FulfillItems(context.Background(), clientID, soID, fulfillReqs)
	require.NoError(t, err)

	// Verify SO status was updated to fulfilled
	assert.Equal(t, 1, fakeSORepo.UpdateCallCount())
	_, updatedSO := fakeSORepo.UpdateArgsForCall(0)
	assert.Equal(t, models.SOStatusFulfilled, updatedSO.Status)
}

func TestSalesOrderService_ListSalesOrders(t *testing.T) {
	fakeSORepo := &fakes.FakeSalesOrderRepository{}
	fakeProductRepo := &fakes.FakeProductRepository{}
	fakeMovementRepo := &fakes.FakeInventoryMovementRepository{}

	service := NewSalesOrderService(fakeSORepo, fakeProductRepo, fakeMovementRepo)

	clientID := uuid.New()

	expectedSOs := []*models.SalesOrder{
		{ID: uuid.New(), ClientID: clientID, OrderNumber: "SO-001"},
		{ID: uuid.New(), ClientID: clientID, OrderNumber: "SO-002"},
	}

	fakeSORepo.ListByClientReturns(expectedSOs, nil)

	sos, total, err := service.ListSalesOrders(context.Background(), clientID, nil, 1, 20)
	require.NoError(t, err)
	assert.Equal(t, expectedSOs, sos)
	assert.Equal(t, 2, total)
	assert.Equal(t, 1, fakeSORepo.ListByClientCallCount())
}

func TestSalesOrderService_ListSalesOrders_FilterByCustomer(t *testing.T) {
	fakeSORepo := &fakes.FakeSalesOrderRepository{}
	fakeProductRepo := &fakes.FakeProductRepository{}
	fakeMovementRepo := &fakes.FakeInventoryMovementRepository{}

	service := NewSalesOrderService(fakeSORepo, fakeProductRepo, fakeMovementRepo)

	clientID := uuid.New()
	customerID := uuid.New()

	expectedSOs := []*models.SalesOrder{
		{ID: uuid.New(), ClientID: clientID, CustomerID: customerID, OrderNumber: "SO-001"},
	}

	fakeSORepo.ListByCustomerReturns(expectedSOs, nil)

	filters := &SalesOrderFilters{
		CustomerID: &customerID,
	}

	sos, total, err := service.ListSalesOrders(context.Background(), clientID, filters, 1, 20)
	require.NoError(t, err)
	assert.Equal(t, expectedSOs, sos)
	assert.Equal(t, 1, total)
	assert.Equal(t, 1, fakeSORepo.ListByCustomerCallCount())
}
