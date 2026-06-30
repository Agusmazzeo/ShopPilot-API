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

func TestPurchaseOrderService_CreatePurchaseOrder_Success(t *testing.T) {
	fakePORepo := &fakes.FakePurchaseOrderRepository{}
	fakeProductRepo := &fakes.FakeProductRepository{}
	fakeMovementRepo := &fakes.FakeInventoryMovementRepository{}

	service := NewPurchaseOrderService(fakePORepo, fakeProductRepo, fakeMovementRepo)

	clientID := uuid.New()
	supplierID := uuid.New()
	shopID := uuid.New()

	// Configure fake - PO number doesn't exist
	fakePORepo.GetByPONumberReturns(nil, errors.New("not found"))
	fakePORepo.CreateReturns(nil)
	fakePORepo.CreateItemReturns(nil)

	req := &CreatePurchaseOrderRequest{
		SupplierID:  supplierID,
		ShopID:      shopID,
		PONumber:    "PO-12345",
		OrderDate:   time.Now(),
		TotalAmount: 1000.00,
		Items: []CreatePurchaseOrderItemRequest{
			{
				VariantID:       uuid.New(),
				QuantityOrdered: 10,
				UnitPrice:       50.00,
				TotalPrice:      500.00,
			},
		},
	}

	po, err := service.CreatePurchaseOrder(context.Background(), clientID, req)
	require.NoError(t, err)
	assert.NotNil(t, po)
	assert.Equal(t, "PO-12345", po.PONumber)
	assert.Equal(t, models.POStatusDraft, po.Status)

	// Verify Create was called
	assert.Equal(t, 1, fakePORepo.CreateCallCount())
	assert.Equal(t, 1, fakePORepo.CreateItemCallCount())
}

func TestPurchaseOrderService_CreatePurchaseOrder_AutoGeneratePONumber(t *testing.T) {
	fakePORepo := &fakes.FakePurchaseOrderRepository{}
	fakeProductRepo := &fakes.FakeProductRepository{}
	fakeMovementRepo := &fakes.FakeInventoryMovementRepository{}

	service := NewPurchaseOrderService(fakePORepo, fakeProductRepo, fakeMovementRepo)

	clientID := uuid.New()

	// Configure fake - no PO number provided
	fakePORepo.GetByPONumberReturns(nil, errors.New("not found"))
	fakePORepo.CreateReturns(nil)

	req := &CreatePurchaseOrderRequest{
		SupplierID:  uuid.New(),
		ShopID:      uuid.New(),
		PONumber:    "", // Empty - should auto-generate
		OrderDate:   time.Now(),
		TotalAmount: 1000.00,
	}

	po, err := service.CreatePurchaseOrder(context.Background(), clientID, req)
	require.NoError(t, err)
	assert.NotNil(t, po)
	assert.NotEmpty(t, po.PONumber)
	assert.Contains(t, po.PONumber, "PO-")
}

func TestPurchaseOrderService_CreatePurchaseOrder_DuplicatePONumber(t *testing.T) {
	fakePORepo := &fakes.FakePurchaseOrderRepository{}
	fakeProductRepo := &fakes.FakeProductRepository{}
	fakeMovementRepo := &fakes.FakeInventoryMovementRepository{}

	service := NewPurchaseOrderService(fakePORepo, fakeProductRepo, fakeMovementRepo)

	clientID := uuid.New()
	existingPO := &models.PurchaseOrder{
		ID:       uuid.New(),
		ClientID: clientID,
		PONumber: "PO-12345",
	}

	// Configure fake - PO number already exists
	fakePORepo.GetByPONumberReturns(existingPO, nil)

	req := &CreatePurchaseOrderRequest{
		SupplierID:  uuid.New(),
		ShopID:      uuid.New(),
		PONumber:    "PO-12345",
		OrderDate:   time.Now(),
		TotalAmount: 1000.00,
	}

	po, err := service.CreatePurchaseOrder(context.Background(), clientID, req)
	require.Error(t, err)
	assert.Nil(t, po)
	assert.Contains(t, err.Error(), "already exists")

	// Create should not be called
	assert.Equal(t, 0, fakePORepo.CreateCallCount())
}

func TestPurchaseOrderService_SubmitPurchaseOrder(t *testing.T) {
	fakePORepo := &fakes.FakePurchaseOrderRepository{}
	fakeProductRepo := &fakes.FakeProductRepository{}
	fakeMovementRepo := &fakes.FakeInventoryMovementRepository{}

	service := NewPurchaseOrderService(fakePORepo, fakeProductRepo, fakeMovementRepo)

	clientID := uuid.New()
	poID := uuid.New()

	draftPO := &models.PurchaseOrder{
		ID:       poID,
		ClientID: clientID,
		Status:   models.POStatusDraft,
	}

	fakePORepo.GetByIDReturns(draftPO, nil)
	fakePORepo.UpdateReturns(nil)

	err := service.SubmitPurchaseOrder(context.Background(), clientID, poID)
	require.NoError(t, err)

	// Verify Update was called
	assert.Equal(t, 1, fakePORepo.UpdateCallCount())

	// Verify status was changed to submitted
	_, updatedPO := fakePORepo.UpdateArgsForCall(0)
	assert.Equal(t, models.POStatusSubmitted, updatedPO.Status)
}

func TestPurchaseOrderService_SubmitPurchaseOrder_OnlyDraft(t *testing.T) {
	fakePORepo := &fakes.FakePurchaseOrderRepository{}
	fakeProductRepo := &fakes.FakeProductRepository{}
	fakeMovementRepo := &fakes.FakeInventoryMovementRepository{}

	service := NewPurchaseOrderService(fakePORepo, fakeProductRepo, fakeMovementRepo)

	clientID := uuid.New()
	poID := uuid.New()

	submittedPO := &models.PurchaseOrder{
		ID:       poID,
		ClientID: clientID,
		Status:   models.POStatusSubmitted,
	}

	fakePORepo.GetByIDReturns(submittedPO, nil)

	err := service.SubmitPurchaseOrder(context.Background(), clientID, poID)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "only draft")

	// Update should not be called
	assert.Equal(t, 0, fakePORepo.UpdateCallCount())
}

func TestPurchaseOrderService_CancelPurchaseOrder(t *testing.T) {
	fakePORepo := &fakes.FakePurchaseOrderRepository{}
	fakeProductRepo := &fakes.FakeProductRepository{}
	fakeMovementRepo := &fakes.FakeInventoryMovementRepository{}

	service := NewPurchaseOrderService(fakePORepo, fakeProductRepo, fakeMovementRepo)

	clientID := uuid.New()
	poID := uuid.New()

	submittedPO := &models.PurchaseOrder{
		ID:       poID,
		ClientID: clientID,
		Status:   models.POStatusSubmitted,
	}

	fakePORepo.GetByIDReturns(submittedPO, nil)
	fakePORepo.UpdateReturns(nil)

	err := service.CancelPurchaseOrder(context.Background(), clientID, poID)
	require.NoError(t, err)

	// Verify Update was called
	assert.Equal(t, 1, fakePORepo.UpdateCallCount())

	// Verify status was changed to cancelled
	_, updatedPO := fakePORepo.UpdateArgsForCall(0)
	assert.Equal(t, models.POStatusCancelled, updatedPO.Status)
}

func TestPurchaseOrderService_AddItem(t *testing.T) {
	fakePORepo := &fakes.FakePurchaseOrderRepository{}
	fakeProductRepo := &fakes.FakeProductRepository{}
	fakeMovementRepo := &fakes.FakeInventoryMovementRepository{}

	service := NewPurchaseOrderService(fakePORepo, fakeProductRepo, fakeMovementRepo)

	clientID := uuid.New()
	poID := uuid.New()

	draftPO := &models.PurchaseOrder{
		ID:       poID,
		ClientID: clientID,
		Status:   models.POStatusDraft,
	}

	fakePORepo.GetByIDReturns(draftPO, nil)
	fakePORepo.CreateItemReturns(nil)

	req := &AddPurchaseOrderItemRequest{
		VariantID:       uuid.New(),
		QuantityOrdered: 10,
		UnitPrice:       50.00,
	}

	item, err := service.AddItem(context.Background(), clientID, poID, req)
	require.NoError(t, err)
	assert.NotNil(t, item)
	assert.Equal(t, 1, fakePORepo.CreateItemCallCount())
}

func TestPurchaseOrderService_AddItem_OnlyDraft(t *testing.T) {
	fakePORepo := &fakes.FakePurchaseOrderRepository{}
	fakeProductRepo := &fakes.FakeProductRepository{}
	fakeMovementRepo := &fakes.FakeInventoryMovementRepository{}

	service := NewPurchaseOrderService(fakePORepo, fakeProductRepo, fakeMovementRepo)

	clientID := uuid.New()
	poID := uuid.New()

	submittedPO := &models.PurchaseOrder{
		ID:       poID,
		ClientID: clientID,
		Status:   models.POStatusSubmitted,
	}

	fakePORepo.GetByIDReturns(submittedPO, nil)

	req := &AddPurchaseOrderItemRequest{
		VariantID:       uuid.New(),
		QuantityOrdered: 10,
		UnitPrice:       50.00,
	}

	item, err := service.AddItem(context.Background(), clientID, poID, req)
	require.Error(t, err)
	assert.Nil(t, item)
	assert.Contains(t, err.Error(), "draft")

	// CreateItem should not be called
	assert.Equal(t, 0, fakePORepo.CreateItemCallCount())
}

func TestPurchaseOrderService_RemoveItem(t *testing.T) {
	fakePORepo := &fakes.FakePurchaseOrderRepository{}
	fakeProductRepo := &fakes.FakeProductRepository{}
	fakeMovementRepo := &fakes.FakeInventoryMovementRepository{}

	service := NewPurchaseOrderService(fakePORepo, fakeProductRepo, fakeMovementRepo)

	clientID := uuid.New()
	poID := uuid.New()
	itemID := uuid.New()

	draftPO := &models.PurchaseOrder{
		ID:       poID,
		ClientID: clientID,
		Status:   models.POStatusDraft,
	}

	fakePORepo.GetByIDReturns(draftPO, nil)
	fakePORepo.DeleteItemReturns(nil)

	err := service.RemoveItem(context.Background(), clientID, poID, itemID)
	require.NoError(t, err)
	assert.Equal(t, 1, fakePORepo.DeleteItemCallCount())
}

func TestPurchaseOrderService_ReceiveItems_Success(t *testing.T) {
	fakePORepo := &fakes.FakePurchaseOrderRepository{}
	fakeProductRepo := &fakes.FakeProductRepository{}
	fakeMovementRepo := &fakes.FakeInventoryMovementRepository{}

	service := NewPurchaseOrderService(fakePORepo, fakeProductRepo, fakeMovementRepo)

	clientID := uuid.New()
	poID := uuid.New()
	shopID := uuid.New()
	itemID := uuid.New()
	variantID := uuid.New()

	submittedPO := &models.PurchaseOrder{
		ID:       poID,
		ClientID: clientID,
		ShopID:   shopID,
		PONumber: "PO-12345",
		Status:   models.POStatusSubmitted,
	}

	item := &models.PurchaseOrderItem{
		ID:               itemID,
		ClientID:         clientID,
		PurchaseOrderID:  poID,
		VariantID:        variantID,
		QuantityOrdered:  10,
		QuantityReceived: 0,
	}

	variant := &models.ProductVariant{
		ID:       variantID,
		ClientID: clientID,
		Quantity: 50,
	}

	allItems := []*models.PurchaseOrderItem{item}

	// Configure fakes
	fakePORepo.GetByIDReturns(submittedPO, nil)
	fakePORepo.GetItemReturns(item, nil)
	fakePORepo.ReceiveItemReturns(nil)
	fakePORepo.ListItemsByPOReturns(allItems, nil)
	fakePORepo.UpdateReturns(nil)
	fakeProductRepo.GetVariantByIDReturns(variant, nil)
	fakeProductRepo.UpdateInventoryReturns(nil)
	fakeMovementRepo.CreateReturns(nil)

	receiveReqs := []ReceiveItemRequest{
		{
			ItemID:           itemID,
			QuantityReceived: 5,
		},
	}

	err := service.ReceiveItems(context.Background(), clientID, poID, receiveReqs)
	require.NoError(t, err)

	// Verify all operations were called in order
	assert.Equal(t, 1, fakePORepo.ReceiveItemCallCount())
	assert.Equal(t, 1, fakeProductRepo.UpdateInventoryCallCount())
	assert.Equal(t, 1, fakeMovementRepo.CreateCallCount())

	// Verify inventory was updated correctly
	_, _, _, newQty := fakeProductRepo.UpdateInventoryArgsForCall(0)
	assert.Equal(t, 55, newQty) // 50 + 5

	// Verify movement was created with correct values
	_, movement := fakeMovementRepo.CreateArgsForCall(0)
	assert.Equal(t, models.MovementTypePurchase, movement.MovementType)
	assert.Equal(t, 5, movement.Quantity)
	assert.Equal(t, 50, movement.PreviousQuantity)
	assert.Equal(t, 55, movement.NewQuantity)
}

func TestPurchaseOrderService_ReceiveItems_PartiallyReceived(t *testing.T) {
	fakePORepo := &fakes.FakePurchaseOrderRepository{}
	fakeProductRepo := &fakes.FakeProductRepository{}
	fakeMovementRepo := &fakes.FakeInventoryMovementRepository{}

	service := NewPurchaseOrderService(fakePORepo, fakeProductRepo, fakeMovementRepo)

	clientID := uuid.New()
	poID := uuid.New()
	shopID := uuid.New()
	itemID := uuid.New()
	variantID := uuid.New()

	submittedPO := &models.PurchaseOrder{
		ID:       poID,
		ClientID: clientID,
		ShopID:   shopID,
		PONumber: "PO-12345",
		Status:   models.POStatusSubmitted,
	}

	item := &models.PurchaseOrderItem{
		ID:               itemID,
		ClientID:         clientID,
		PurchaseOrderID:  poID,
		VariantID:        variantID,
		QuantityOrdered:  10,
		QuantityReceived: 5, // Partially received
	}

	variant := &models.ProductVariant{
		ID:       variantID,
		ClientID: clientID,
		Quantity: 50,
	}

	allItems := []*models.PurchaseOrderItem{item}

	// Configure fakes
	fakePORepo.GetByIDReturns(submittedPO, nil)
	fakePORepo.GetItemReturns(item, nil)
	fakePORepo.ReceiveItemReturns(nil)
	fakePORepo.ListItemsByPOReturns(allItems, nil)
	fakePORepo.UpdateReturns(nil)
	fakeProductRepo.GetVariantByIDReturns(variant, nil)
	fakeProductRepo.UpdateInventoryReturns(nil)
	fakeMovementRepo.CreateReturns(nil)

	receiveReqs := []ReceiveItemRequest{
		{
			ItemID:           itemID,
			QuantityReceived: 3,
		},
	}

	err := service.ReceiveItems(context.Background(), clientID, poID, receiveReqs)
	require.NoError(t, err)

	// Verify PO status was updated to partially received
	assert.Equal(t, 1, fakePORepo.UpdateCallCount())
	_, updatedPO := fakePORepo.UpdateArgsForCall(0)
	assert.Equal(t, models.POStatusPartiallyReceived, updatedPO.Status)
}

func TestPurchaseOrderService_ReceiveItems_FullyReceived(t *testing.T) {
	fakePORepo := &fakes.FakePurchaseOrderRepository{}
	fakeProductRepo := &fakes.FakeProductRepository{}
	fakeMovementRepo := &fakes.FakeInventoryMovementRepository{}

	service := NewPurchaseOrderService(fakePORepo, fakeProductRepo, fakeMovementRepo)

	clientID := uuid.New()
	poID := uuid.New()
	shopID := uuid.New()
	itemID := uuid.New()
	variantID := uuid.New()

	submittedPO := &models.PurchaseOrder{
		ID:       poID,
		ClientID: clientID,
		ShopID:   shopID,
		PONumber: "PO-12345",
		Status:   models.POStatusSubmitted,
	}

	item := &models.PurchaseOrderItem{
		ID:               itemID,
		ClientID:         clientID,
		PurchaseOrderID:  poID,
		VariantID:        variantID,
		QuantityOrdered:  10,
		QuantityReceived: 10, // Fully received
	}

	variant := &models.ProductVariant{
		ID:       variantID,
		ClientID: clientID,
		Quantity: 50,
	}

	allItems := []*models.PurchaseOrderItem{item}

	// Configure fakes
	fakePORepo.GetByIDReturns(submittedPO, nil)
	fakePORepo.GetItemReturns(item, nil)
	fakePORepo.ReceiveItemReturns(nil)
	fakePORepo.ListItemsByPOReturns(allItems, nil)
	fakePORepo.UpdateReturns(nil)
	fakeProductRepo.GetVariantByIDReturns(variant, nil)
	fakeProductRepo.UpdateInventoryReturns(nil)
	fakeMovementRepo.CreateReturns(nil)

	receiveReqs := []ReceiveItemRequest{
		{
			ItemID:           itemID,
			QuantityReceived: 10,
		},
	}

	err := service.ReceiveItems(context.Background(), clientID, poID, receiveReqs)
	require.NoError(t, err)

	// Verify PO status was updated to received
	assert.Equal(t, 1, fakePORepo.UpdateCallCount())
	_, updatedPO := fakePORepo.UpdateArgsForCall(0)
	assert.Equal(t, models.POStatusReceived, updatedPO.Status)
}

func TestPurchaseOrderService_ListPurchaseOrders(t *testing.T) {
	fakePORepo := &fakes.FakePurchaseOrderRepository{}
	fakeProductRepo := &fakes.FakeProductRepository{}
	fakeMovementRepo := &fakes.FakeInventoryMovementRepository{}

	service := NewPurchaseOrderService(fakePORepo, fakeProductRepo, fakeMovementRepo)

	clientID := uuid.New()

	expectedPOs := []*models.PurchaseOrder{
		{ID: uuid.New(), ClientID: clientID, PONumber: "PO-001"},
		{ID: uuid.New(), ClientID: clientID, PONumber: "PO-002"},
	}

	fakePORepo.ListByClientReturns(expectedPOs, nil)

	pos, total, err := service.ListPurchaseOrders(context.Background(), clientID, nil, 1, 20)
	require.NoError(t, err)
	assert.Equal(t, expectedPOs, pos)
	assert.Equal(t, 2, total)
	assert.Equal(t, 1, fakePORepo.ListByClientCallCount())
}

func TestPurchaseOrderService_ListPurchaseOrders_FilterBySupplier(t *testing.T) {
	fakePORepo := &fakes.FakePurchaseOrderRepository{}
	fakeProductRepo := &fakes.FakeProductRepository{}
	fakeMovementRepo := &fakes.FakeInventoryMovementRepository{}

	service := NewPurchaseOrderService(fakePORepo, fakeProductRepo, fakeMovementRepo)

	clientID := uuid.New()
	supplierID := uuid.New()

	expectedPOs := []*models.PurchaseOrder{
		{ID: uuid.New(), ClientID: clientID, SupplierID: supplierID, PONumber: "PO-001"},
	}

	fakePORepo.ListBySupplierReturns(expectedPOs, nil)

	filters := &PurchaseOrderFilters{
		SupplierID: &supplierID,
	}

	pos, total, err := service.ListPurchaseOrders(context.Background(), clientID, filters, 1, 20)
	require.NoError(t, err)
	assert.Equal(t, expectedPOs, pos)
	assert.Equal(t, 1, total)
	assert.Equal(t, 1, fakePORepo.ListBySupplierCallCount())
}
