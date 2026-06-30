package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yourorg/shoppilot/internal/models"
	"github.com/yourorg/shoppilot/internal/server/handlers/fakes"
)

func TestPurchaseOrderHandler_Create_Success(t *testing.T) {
	fakeService := &fakes.FakePurchaseOrderService{}
	handler := NewPurchaseOrderHandler(fakeService)

	clientID := uuid.New()
	supplierID := uuid.New()
	shopID := uuid.New()
	poID := uuid.New()
	variantID := uuid.New()

	// Configure fake
	fakeService.CreatePurchaseOrderReturns(&models.PurchaseOrder{
		ID:         poID,
		ClientID:   clientID,
		SupplierID: supplierID,
		ShopID:     shopID,
		PONumber:   "PO-123456",
		Status:     models.POStatusDraft,
	}, nil)

	// Create request body
	reqBody := CreatePurchaseOrderRequestDTO{
		SupplierID:   supplierID.String(),
		ShopID:       shopID.String(),
		PONumber:     "PO-123456",
		OrderDate:    time.Now(),
		TotalAmount:  1000.00,
		Items: []CreatePurchaseOrderItemRequestDTO{
			{
				VariantID:       variantID.String(),
				QuantityOrdered: 10,
				UnitPrice:       100.00,
				TotalPrice:      1000.00,
			},
		},
	}
	bodyBytes, _ := json.Marshal(reqBody)

	// Create HTTP request
	req := httptest.NewRequest(http.MethodPost, "/api/v1/clients/"+clientID.String()+"/purchase-orders", bytes.NewReader(bodyBytes))
	req = mux.SetURLVars(req, map[string]string{"clientId": clientID.String()})
	rec := httptest.NewRecorder()

	// Execute handler
	handler.Create(rec, req)

	// Verify response
	assert.Equal(t, http.StatusCreated, rec.Code)
	assert.Equal(t, 1, fakeService.CreatePurchaseOrderCallCount())

	// Verify response body
	var response APIResponse
	err := json.NewDecoder(rec.Body).Decode(&response)
	require.NoError(t, err)
	assert.True(t, response.Success)
}

func TestPurchaseOrderHandler_Create_InvalidJSON(t *testing.T) {
	fakeService := &fakes.FakePurchaseOrderService{}
	handler := NewPurchaseOrderHandler(fakeService)

	clientID := uuid.New()

	req := httptest.NewRequest(http.MethodPost, "/api/v1/clients/"+clientID.String()+"/purchase-orders", bytes.NewReader([]byte("invalid json")))
	req = mux.SetURLVars(req, map[string]string{"clientId": clientID.String()})
	rec := httptest.NewRecorder()

	handler.Create(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Equal(t, 0, fakeService.CreatePurchaseOrderCallCount())
}

func TestPurchaseOrderHandler_Create_MissingFields(t *testing.T) {
	fakeService := &fakes.FakePurchaseOrderService{}
	handler := NewPurchaseOrderHandler(fakeService)

	clientID := uuid.New()

	reqBody := CreatePurchaseOrderRequestDTO{
		// Missing SupplierID and ShopID
		OrderDate: time.Now(),
	}
	bodyBytes, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/clients/"+clientID.String()+"/purchase-orders", bytes.NewReader(bodyBytes))
	req = mux.SetURLVars(req, map[string]string{"clientId": clientID.String()})
	rec := httptest.NewRecorder()

	handler.Create(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Equal(t, 0, fakeService.CreatePurchaseOrderCallCount())
}

func TestPurchaseOrderHandler_Get_Success(t *testing.T) {
	fakeService := &fakes.FakePurchaseOrderService{}
	handler := NewPurchaseOrderHandler(fakeService)

	clientID := uuid.New()
	poID := uuid.New()

	fakeService.GetPurchaseOrderReturns(&models.PurchaseOrder{
		ID:       poID,
		ClientID: clientID,
		PONumber: "PO-123456",
		Status:   models.POStatusDraft,
	}, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/clients/"+clientID.String()+"/purchase-orders/"+poID.String(), nil)
	req = mux.SetURLVars(req, map[string]string{
		"clientId": clientID.String(),
		"id":       poID.String(),
	})
	rec := httptest.NewRecorder()

	handler.Get(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, 1, fakeService.GetPurchaseOrderCallCount())
}

func TestPurchaseOrderHandler_Get_NotFound(t *testing.T) {
	fakeService := &fakes.FakePurchaseOrderService{}
	handler := NewPurchaseOrderHandler(fakeService)

	clientID := uuid.New()
	poID := uuid.New()

	fakeService.GetPurchaseOrderReturns(nil, errors.New("purchase order not found"))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/clients/"+clientID.String()+"/purchase-orders/"+poID.String(), nil)
	req = mux.SetURLVars(req, map[string]string{
		"clientId": clientID.String(),
		"id":       poID.String(),
	})
	rec := httptest.NewRecorder()

	handler.Get(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestPurchaseOrderHandler_Update_Success(t *testing.T) {
	fakeService := &fakes.FakePurchaseOrderService{}
	handler := NewPurchaseOrderHandler(fakeService)

	clientID := uuid.New()
	poID := uuid.New()

	fakeService.UpdatePurchaseOrderReturns(nil)

	newNotes := "Updated notes"
	reqBody := UpdatePurchaseOrderRequestDTO{
		Notes: &newNotes,
	}
	bodyBytes, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPut, "/api/v1/clients/"+clientID.String()+"/purchase-orders/"+poID.String(), bytes.NewReader(bodyBytes))
	req = mux.SetURLVars(req, map[string]string{
		"clientId": clientID.String(),
		"id":       poID.String(),
	})
	rec := httptest.NewRecorder()

	handler.Update(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, 1, fakeService.UpdatePurchaseOrderCallCount())
}

func TestPurchaseOrderHandler_Update_InvalidJSON(t *testing.T) {
	fakeService := &fakes.FakePurchaseOrderService{}
	handler := NewPurchaseOrderHandler(fakeService)

	clientID := uuid.New()
	poID := uuid.New()

	req := httptest.NewRequest(http.MethodPut, "/api/v1/clients/"+clientID.String()+"/purchase-orders/"+poID.String(), bytes.NewReader([]byte("invalid json")))
	req = mux.SetURLVars(req, map[string]string{
		"clientId": clientID.String(),
		"id":       poID.String(),
	})
	rec := httptest.NewRecorder()

	handler.Update(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Equal(t, 0, fakeService.UpdatePurchaseOrderCallCount())
}

func TestPurchaseOrderHandler_Delete_Success(t *testing.T) {
	fakeService := &fakes.FakePurchaseOrderService{}
	handler := NewPurchaseOrderHandler(fakeService)

	clientID := uuid.New()
	poID := uuid.New()

	fakeService.DeletePurchaseOrderReturns(nil)

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/clients/"+clientID.String()+"/purchase-orders/"+poID.String(), nil)
	req = mux.SetURLVars(req, map[string]string{
		"clientId": clientID.String(),
		"id":       poID.String(),
	})
	rec := httptest.NewRecorder()

	handler.Delete(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, 1, fakeService.DeletePurchaseOrderCallCount())
}

func TestPurchaseOrderHandler_List_Success(t *testing.T) {
	fakeService := &fakes.FakePurchaseOrderService{}
	handler := NewPurchaseOrderHandler(fakeService)

	clientID := uuid.New()

	purchaseOrders := []*models.PurchaseOrder{
		{ID: uuid.New(), ClientID: clientID, PONumber: "PO-001", Status: models.POStatusDraft},
		{ID: uuid.New(), ClientID: clientID, PONumber: "PO-002", Status: models.POStatusSubmitted},
	}
	fakeService.ListPurchaseOrdersReturns(purchaseOrders, 2, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/clients/"+clientID.String()+"/purchase-orders", nil)
	req = mux.SetURLVars(req, map[string]string{"clientId": clientID.String()})
	rec := httptest.NewRecorder()

	handler.List(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, 1, fakeService.ListPurchaseOrdersCallCount())
}

func TestPurchaseOrderHandler_List_WithFilters(t *testing.T) {
	fakeService := &fakes.FakePurchaseOrderService{}
	handler := NewPurchaseOrderHandler(fakeService)

	clientID := uuid.New()
	supplierID := uuid.New()

	purchaseOrders := []*models.PurchaseOrder{
		{ID: uuid.New(), ClientID: clientID, SupplierID: supplierID, PONumber: "PO-001"},
	}
	fakeService.ListPurchaseOrdersReturns(purchaseOrders, 1, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/clients/"+clientID.String()+"/purchase-orders?supplierId="+supplierID.String(), nil)
	req = mux.SetURLVars(req, map[string]string{"clientId": clientID.String()})
	rec := httptest.NewRecorder()

	handler.List(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, 1, fakeService.ListPurchaseOrdersCallCount())

	// Verify filters were passed
	_, _, filters, _, _ := fakeService.ListPurchaseOrdersArgsForCall(0)
	assert.NotNil(t, filters.SupplierID)
	assert.Equal(t, supplierID, *filters.SupplierID)
}

func TestPurchaseOrderHandler_Submit_Success(t *testing.T) {
	fakeService := &fakes.FakePurchaseOrderService{}
	handler := NewPurchaseOrderHandler(fakeService)

	clientID := uuid.New()
	poID := uuid.New()

	fakeService.SubmitPurchaseOrderReturns(nil)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/clients/"+clientID.String()+"/purchase-orders/"+poID.String()+"/submit", nil)
	req = mux.SetURLVars(req, map[string]string{
		"clientId": clientID.String(),
		"id":       poID.String(),
	})
	rec := httptest.NewRecorder()

	handler.Submit(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, 1, fakeService.SubmitPurchaseOrderCallCount())
}

func TestPurchaseOrderHandler_Submit_InvalidStatus(t *testing.T) {
	fakeService := &fakes.FakePurchaseOrderService{}
	handler := NewPurchaseOrderHandler(fakeService)

	clientID := uuid.New()
	poID := uuid.New()

	fakeService.SubmitPurchaseOrderReturns(errors.New("only draft purchase orders can be submitted"))

	req := httptest.NewRequest(http.MethodPost, "/api/v1/clients/"+clientID.String()+"/purchase-orders/"+poID.String()+"/submit", nil)
	req = mux.SetURLVars(req, map[string]string{
		"clientId": clientID.String(),
		"id":       poID.String(),
	})
	rec := httptest.NewRecorder()

	handler.Submit(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestPurchaseOrderHandler_Cancel_Success(t *testing.T) {
	fakeService := &fakes.FakePurchaseOrderService{}
	handler := NewPurchaseOrderHandler(fakeService)

	clientID := uuid.New()
	poID := uuid.New()

	fakeService.CancelPurchaseOrderReturns(nil)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/clients/"+clientID.String()+"/purchase-orders/"+poID.String()+"/cancel", nil)
	req = mux.SetURLVars(req, map[string]string{
		"clientId": clientID.String(),
		"id":       poID.String(),
	})
	rec := httptest.NewRecorder()

	handler.Cancel(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, 1, fakeService.CancelPurchaseOrderCallCount())
}

func TestPurchaseOrderHandler_AddItem_Success(t *testing.T) {
	fakeService := &fakes.FakePurchaseOrderService{}
	handler := NewPurchaseOrderHandler(fakeService)

	clientID := uuid.New()
	poID := uuid.New()
	variantID := uuid.New()
	itemID := uuid.New()

	fakeService.AddItemReturns(&models.PurchaseOrderItem{
		ID:              itemID,
		ClientID:        clientID,
		PurchaseOrderID: poID,
		VariantID:       variantID,
		QuantityOrdered: 10,
	}, nil)

	reqBody := AddPurchaseOrderItemRequestDTO{
		VariantID:       variantID.String(),
		QuantityOrdered: 10,
		UnitPrice:       100.00,
	}
	bodyBytes, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/clients/"+clientID.String()+"/purchase-orders/"+poID.String()+"/items", bytes.NewReader(bodyBytes))
	req = mux.SetURLVars(req, map[string]string{
		"clientId": clientID.String(),
		"id":       poID.String(),
	})
	rec := httptest.NewRecorder()

	handler.AddItem(rec, req)

	assert.Equal(t, http.StatusCreated, rec.Code)
	assert.Equal(t, 1, fakeService.AddItemCallCount())
}

func TestPurchaseOrderHandler_AddItem_MissingVariantID(t *testing.T) {
	fakeService := &fakes.FakePurchaseOrderService{}
	handler := NewPurchaseOrderHandler(fakeService)

	clientID := uuid.New()
	poID := uuid.New()

	reqBody := AddPurchaseOrderItemRequestDTO{
		// Missing VariantID
		QuantityOrdered: 10,
		UnitPrice:       100.00,
	}
	bodyBytes, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/clients/"+clientID.String()+"/purchase-orders/"+poID.String()+"/items", bytes.NewReader(bodyBytes))
	req = mux.SetURLVars(req, map[string]string{
		"clientId": clientID.String(),
		"id":       poID.String(),
	})
	rec := httptest.NewRecorder()

	handler.AddItem(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Equal(t, 0, fakeService.AddItemCallCount())
}

func TestPurchaseOrderHandler_RemoveItem_Success(t *testing.T) {
	fakeService := &fakes.FakePurchaseOrderService{}
	handler := NewPurchaseOrderHandler(fakeService)

	clientID := uuid.New()
	poID := uuid.New()
	itemID := uuid.New()

	fakeService.RemoveItemReturns(nil)

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/clients/"+clientID.String()+"/purchase-orders/"+poID.String()+"/items/"+itemID.String(), nil)
	req = mux.SetURLVars(req, map[string]string{
		"clientId": clientID.String(),
		"id":       poID.String(),
		"itemId":   itemID.String(),
	})
	rec := httptest.NewRecorder()

	handler.RemoveItem(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, 1, fakeService.RemoveItemCallCount())
}

func TestPurchaseOrderHandler_ListItems_Success(t *testing.T) {
	fakeService := &fakes.FakePurchaseOrderService{}
	handler := NewPurchaseOrderHandler(fakeService)

	clientID := uuid.New()
	poID := uuid.New()

	items := []*models.PurchaseOrderItem{
		{ID: uuid.New(), ClientID: clientID, PurchaseOrderID: poID, QuantityOrdered: 10},
		{ID: uuid.New(), ClientID: clientID, PurchaseOrderID: poID, QuantityOrdered: 20},
	}
	fakeService.ListItemsReturns(items, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/clients/"+clientID.String()+"/purchase-orders/"+poID.String()+"/items", nil)
	req = mux.SetURLVars(req, map[string]string{
		"clientId": clientID.String(),
		"id":       poID.String(),
	})
	rec := httptest.NewRecorder()

	handler.ListItems(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, 1, fakeService.ListItemsCallCount())
}

func TestPurchaseOrderHandler_Receive_Success(t *testing.T) {
	fakeService := &fakes.FakePurchaseOrderService{}
	handler := NewPurchaseOrderHandler(fakeService)

	clientID := uuid.New()
	poID := uuid.New()
	itemID := uuid.New()

	fakeService.ReceiveItemsReturns(nil)

	reqBody := []ReceiveItemRequestDTO{
		{
			ItemID:           itemID.String(),
			QuantityReceived: 10,
		},
	}
	bodyBytes, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/clients/"+clientID.String()+"/purchase-orders/"+poID.String()+"/receive", bytes.NewReader(bodyBytes))
	req = mux.SetURLVars(req, map[string]string{
		"clientId": clientID.String(),
		"id":       poID.String(),
	})
	rec := httptest.NewRecorder()

	handler.Receive(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, 1, fakeService.ReceiveItemsCallCount())

	// Verify items were passed correctly
	_, _, _, items := fakeService.ReceiveItemsArgsForCall(0)
	assert.Len(t, items, 1)
	assert.Equal(t, itemID, items[0].ItemID)
	assert.Equal(t, 10, items[0].QuantityReceived)
}

func TestPurchaseOrderHandler_Receive_InvalidJSON(t *testing.T) {
	fakeService := &fakes.FakePurchaseOrderService{}
	handler := NewPurchaseOrderHandler(fakeService)

	clientID := uuid.New()
	poID := uuid.New()

	req := httptest.NewRequest(http.MethodPost, "/api/v1/clients/"+clientID.String()+"/purchase-orders/"+poID.String()+"/receive", bytes.NewReader([]byte("invalid json")))
	req = mux.SetURLVars(req, map[string]string{
		"clientId": clientID.String(),
		"id":       poID.String(),
	})
	rec := httptest.NewRecorder()

	handler.Receive(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Equal(t, 0, fakeService.ReceiveItemsCallCount())
}

func TestPurchaseOrderHandler_Receive_ServiceError(t *testing.T) {
	fakeService := &fakes.FakePurchaseOrderService{}
	handler := NewPurchaseOrderHandler(fakeService)

	clientID := uuid.New()
	poID := uuid.New()
	itemID := uuid.New()

	fakeService.ReceiveItemsReturns(errors.New("insufficient inventory"))

	reqBody := []ReceiveItemRequestDTO{
		{
			ItemID:           itemID.String(),
			QuantityReceived: 100,
		},
	}
	bodyBytes, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/clients/"+clientID.String()+"/purchase-orders/"+poID.String()+"/receive", bytes.NewReader(bodyBytes))
	req = mux.SetURLVars(req, map[string]string{
		"clientId": clientID.String(),
		"id":       poID.String(),
	})
	rec := httptest.NewRecorder()

	handler.Receive(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}
