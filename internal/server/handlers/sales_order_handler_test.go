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

func TestSalesOrderHandler_Create_Success(t *testing.T) {
	fakeService := &fakes.FakeSalesOrderService{}
	handler := NewSalesOrderHandler(fakeService)

	clientID := uuid.New()
	customerID := uuid.New()
	shopID := uuid.New()
	soID := uuid.New()
	variantID := uuid.New()

	// Configure fake
	fakeService.CreateSalesOrderReturns(&models.SalesOrder{
		ID:          soID,
		ClientID:    clientID,
		CustomerID:  customerID,
		ShopID:      shopID,
		OrderNumber: "SO-123456",
		Status:      models.SOStatusPending,
	}, nil)

	// Create request body
	reqBody := CreateSalesOrderRequestDTO{
		CustomerID:  customerID.String(),
		ShopID:      shopID.String(),
		OrderNumber: "SO-123456",
		OrderDate:   time.Now(),
		Subtotal:    900.00,
		TaxAmount:   100.00,
		TotalAmount: 1000.00,
		Items: []CreateSalesOrderItemRequestDTO{
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
	req := httptest.NewRequest(http.MethodPost, "/api/v1/clients/"+clientID.String()+"/sales-orders", bytes.NewReader(bodyBytes))
	req = mux.SetURLVars(req, map[string]string{"clientId": clientID.String()})
	rec := httptest.NewRecorder()

	// Execute handler
	handler.Create(rec, req)

	// Verify response
	assert.Equal(t, http.StatusCreated, rec.Code)
	assert.Equal(t, 1, fakeService.CreateSalesOrderCallCount())

	// Verify response body
	var response APIResponse
	err := json.NewDecoder(rec.Body).Decode(&response)
	require.NoError(t, err)
	assert.True(t, response.Success)
}

func TestSalesOrderHandler_Create_InvalidJSON(t *testing.T) {
	fakeService := &fakes.FakeSalesOrderService{}
	handler := NewSalesOrderHandler(fakeService)

	clientID := uuid.New()

	req := httptest.NewRequest(http.MethodPost, "/api/v1/clients/"+clientID.String()+"/sales-orders", bytes.NewReader([]byte("invalid json")))
	req = mux.SetURLVars(req, map[string]string{"clientId": clientID.String()})
	rec := httptest.NewRecorder()

	handler.Create(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Equal(t, 0, fakeService.CreateSalesOrderCallCount())
}

func TestSalesOrderHandler_Create_MissingFields(t *testing.T) {
	fakeService := &fakes.FakeSalesOrderService{}
	handler := NewSalesOrderHandler(fakeService)

	clientID := uuid.New()

	reqBody := CreateSalesOrderRequestDTO{
		// Missing CustomerID and ShopID
		OrderDate: time.Now(),
	}
	bodyBytes, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/clients/"+clientID.String()+"/sales-orders", bytes.NewReader(bodyBytes))
	req = mux.SetURLVars(req, map[string]string{"clientId": clientID.String()})
	rec := httptest.NewRecorder()

	handler.Create(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Equal(t, 0, fakeService.CreateSalesOrderCallCount())
}

func TestSalesOrderHandler_Get_Success(t *testing.T) {
	fakeService := &fakes.FakeSalesOrderService{}
	handler := NewSalesOrderHandler(fakeService)

	clientID := uuid.New()
	soID := uuid.New()

	fakeService.GetSalesOrderReturns(&models.SalesOrder{
		ID:          soID,
		ClientID:    clientID,
		OrderNumber: "SO-123456",
		Status:      models.SOStatusPending,
	}, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/clients/"+clientID.String()+"/sales-orders/"+soID.String(), nil)
	req = mux.SetURLVars(req, map[string]string{
		"clientId": clientID.String(),
		"id":       soID.String(),
	})
	rec := httptest.NewRecorder()

	handler.Get(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, 1, fakeService.GetSalesOrderCallCount())
}

func TestSalesOrderHandler_Get_NotFound(t *testing.T) {
	fakeService := &fakes.FakeSalesOrderService{}
	handler := NewSalesOrderHandler(fakeService)

	clientID := uuid.New()
	soID := uuid.New()

	fakeService.GetSalesOrderReturns(nil, errors.New("sales order not found"))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/clients/"+clientID.String()+"/sales-orders/"+soID.String(), nil)
	req = mux.SetURLVars(req, map[string]string{
		"clientId": clientID.String(),
		"id":       soID.String(),
	})
	rec := httptest.NewRecorder()

	handler.Get(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestSalesOrderHandler_Update_Success(t *testing.T) {
	fakeService := &fakes.FakeSalesOrderService{}
	handler := NewSalesOrderHandler(fakeService)

	clientID := uuid.New()
	soID := uuid.New()

	fakeService.UpdateSalesOrderReturns(nil)

	newNotes := "Updated notes"
	reqBody := UpdateSalesOrderRequestDTO{
		Notes: &newNotes,
	}
	bodyBytes, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPut, "/api/v1/clients/"+clientID.String()+"/sales-orders/"+soID.String(), bytes.NewReader(bodyBytes))
	req = mux.SetURLVars(req, map[string]string{
		"clientId": clientID.String(),
		"id":       soID.String(),
	})
	rec := httptest.NewRecorder()

	handler.Update(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, 1, fakeService.UpdateSalesOrderCallCount())
}

func TestSalesOrderHandler_Update_InvalidJSON(t *testing.T) {
	fakeService := &fakes.FakeSalesOrderService{}
	handler := NewSalesOrderHandler(fakeService)

	clientID := uuid.New()
	soID := uuid.New()

	req := httptest.NewRequest(http.MethodPut, "/api/v1/clients/"+clientID.String()+"/sales-orders/"+soID.String(), bytes.NewReader([]byte("invalid json")))
	req = mux.SetURLVars(req, map[string]string{
		"clientId": clientID.String(),
		"id":       soID.String(),
	})
	rec := httptest.NewRecorder()

	handler.Update(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Equal(t, 0, fakeService.UpdateSalesOrderCallCount())
}

func TestSalesOrderHandler_Delete_Success(t *testing.T) {
	fakeService := &fakes.FakeSalesOrderService{}
	handler := NewSalesOrderHandler(fakeService)

	clientID := uuid.New()
	soID := uuid.New()

	fakeService.DeleteSalesOrderReturns(nil)

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/clients/"+clientID.String()+"/sales-orders/"+soID.String(), nil)
	req = mux.SetURLVars(req, map[string]string{
		"clientId": clientID.String(),
		"id":       soID.String(),
	})
	rec := httptest.NewRecorder()

	handler.Delete(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, 1, fakeService.DeleteSalesOrderCallCount())
}

func TestSalesOrderHandler_List_Success(t *testing.T) {
	fakeService := &fakes.FakeSalesOrderService{}
	handler := NewSalesOrderHandler(fakeService)

	clientID := uuid.New()

	salesOrders := []*models.SalesOrder{
		{ID: uuid.New(), ClientID: clientID, OrderNumber: "SO-001", Status: models.SOStatusPending},
		{ID: uuid.New(), ClientID: clientID, OrderNumber: "SO-002", Status: models.SOStatusConfirmed},
	}
	fakeService.ListSalesOrdersReturns(salesOrders, 2, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/clients/"+clientID.String()+"/sales-orders", nil)
	req = mux.SetURLVars(req, map[string]string{"clientId": clientID.String()})
	rec := httptest.NewRecorder()

	handler.List(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, 1, fakeService.ListSalesOrdersCallCount())
}

func TestSalesOrderHandler_List_WithFilters(t *testing.T) {
	fakeService := &fakes.FakeSalesOrderService{}
	handler := NewSalesOrderHandler(fakeService)

	clientID := uuid.New()
	customerID := uuid.New()

	salesOrders := []*models.SalesOrder{
		{ID: uuid.New(), ClientID: clientID, CustomerID: customerID, OrderNumber: "SO-001"},
	}
	fakeService.ListSalesOrdersReturns(salesOrders, 1, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/clients/"+clientID.String()+"/sales-orders?customerId="+customerID.String(), nil)
	req = mux.SetURLVars(req, map[string]string{"clientId": clientID.String()})
	rec := httptest.NewRecorder()

	handler.List(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, 1, fakeService.ListSalesOrdersCallCount())

	// Verify filters were passed
	_, _, filters, _, _ := fakeService.ListSalesOrdersArgsForCall(0)
	assert.NotNil(t, filters.CustomerID)
	assert.Equal(t, customerID, *filters.CustomerID)
}

func TestSalesOrderHandler_Confirm_Success(t *testing.T) {
	fakeService := &fakes.FakeSalesOrderService{}
	handler := NewSalesOrderHandler(fakeService)

	clientID := uuid.New()
	soID := uuid.New()

	fakeService.ConfirmSalesOrderReturns(nil)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/clients/"+clientID.String()+"/sales-orders/"+soID.String()+"/confirm", nil)
	req = mux.SetURLVars(req, map[string]string{
		"clientId": clientID.String(),
		"id":       soID.String(),
	})
	rec := httptest.NewRecorder()

	handler.Confirm(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, 1, fakeService.ConfirmSalesOrderCallCount())
}

func TestSalesOrderHandler_Confirm_InvalidStatus(t *testing.T) {
	fakeService := &fakes.FakeSalesOrderService{}
	handler := NewSalesOrderHandler(fakeService)

	clientID := uuid.New()
	soID := uuid.New()

	fakeService.ConfirmSalesOrderReturns(errors.New("only pending sales orders can be confirmed"))

	req := httptest.NewRequest(http.MethodPost, "/api/v1/clients/"+clientID.String()+"/sales-orders/"+soID.String()+"/confirm", nil)
	req = mux.SetURLVars(req, map[string]string{
		"clientId": clientID.String(),
		"id":       soID.String(),
	})
	rec := httptest.NewRecorder()

	handler.Confirm(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestSalesOrderHandler_Cancel_Success(t *testing.T) {
	fakeService := &fakes.FakeSalesOrderService{}
	handler := NewSalesOrderHandler(fakeService)

	clientID := uuid.New()
	soID := uuid.New()

	fakeService.CancelSalesOrderReturns(nil)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/clients/"+clientID.String()+"/sales-orders/"+soID.String()+"/cancel", nil)
	req = mux.SetURLVars(req, map[string]string{
		"clientId": clientID.String(),
		"id":       soID.String(),
	})
	rec := httptest.NewRecorder()

	handler.Cancel(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, 1, fakeService.CancelSalesOrderCallCount())
}

func TestSalesOrderHandler_AddItem_Success(t *testing.T) {
	fakeService := &fakes.FakeSalesOrderService{}
	handler := NewSalesOrderHandler(fakeService)

	clientID := uuid.New()
	soID := uuid.New()
	variantID := uuid.New()
	itemID := uuid.New()

	fakeService.AddItemReturns(&models.SalesOrderItem{
		ID:              itemID,
		ClientID:        clientID,
		SalesOrderID:    soID,
		VariantID:       variantID,
		QuantityOrdered: 10,
	}, nil)

	reqBody := AddSalesOrderItemRequestDTO{
		VariantID:       variantID.String(),
		QuantityOrdered: 10,
		UnitPrice:       100.00,
	}
	bodyBytes, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/clients/"+clientID.String()+"/sales-orders/"+soID.String()+"/items", bytes.NewReader(bodyBytes))
	req = mux.SetURLVars(req, map[string]string{
		"clientId": clientID.String(),
		"id":       soID.String(),
	})
	rec := httptest.NewRecorder()

	handler.AddItem(rec, req)

	assert.Equal(t, http.StatusCreated, rec.Code)
	assert.Equal(t, 1, fakeService.AddItemCallCount())
}

func TestSalesOrderHandler_AddItem_MissingVariantID(t *testing.T) {
	fakeService := &fakes.FakeSalesOrderService{}
	handler := NewSalesOrderHandler(fakeService)

	clientID := uuid.New()
	soID := uuid.New()

	reqBody := AddSalesOrderItemRequestDTO{
		// Missing VariantID
		QuantityOrdered: 10,
		UnitPrice:       100.00,
	}
	bodyBytes, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/clients/"+clientID.String()+"/sales-orders/"+soID.String()+"/items", bytes.NewReader(bodyBytes))
	req = mux.SetURLVars(req, map[string]string{
		"clientId": clientID.String(),
		"id":       soID.String(),
	})
	rec := httptest.NewRecorder()

	handler.AddItem(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Equal(t, 0, fakeService.AddItemCallCount())
}

func TestSalesOrderHandler_RemoveItem_Success(t *testing.T) {
	fakeService := &fakes.FakeSalesOrderService{}
	handler := NewSalesOrderHandler(fakeService)

	clientID := uuid.New()
	soID := uuid.New()
	itemID := uuid.New()

	fakeService.RemoveItemReturns(nil)

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/clients/"+clientID.String()+"/sales-orders/"+soID.String()+"/items/"+itemID.String(), nil)
	req = mux.SetURLVars(req, map[string]string{
		"clientId": clientID.String(),
		"id":       soID.String(),
		"itemId":   itemID.String(),
	})
	rec := httptest.NewRecorder()

	handler.RemoveItem(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, 1, fakeService.RemoveItemCallCount())
}

func TestSalesOrderHandler_ListItems_Success(t *testing.T) {
	fakeService := &fakes.FakeSalesOrderService{}
	handler := NewSalesOrderHandler(fakeService)

	clientID := uuid.New()
	soID := uuid.New()

	items := []*models.SalesOrderItem{
		{ID: uuid.New(), ClientID: clientID, SalesOrderID: soID, QuantityOrdered: 10},
		{ID: uuid.New(), ClientID: clientID, SalesOrderID: soID, QuantityOrdered: 20},
	}
	fakeService.ListItemsReturns(items, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/clients/"+clientID.String()+"/sales-orders/"+soID.String()+"/items", nil)
	req = mux.SetURLVars(req, map[string]string{
		"clientId": clientID.String(),
		"id":       soID.String(),
	})
	rec := httptest.NewRecorder()

	handler.ListItems(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, 1, fakeService.ListItemsCallCount())
}

func TestSalesOrderHandler_Fulfill_Success(t *testing.T) {
	fakeService := &fakes.FakeSalesOrderService{}
	handler := NewSalesOrderHandler(fakeService)

	clientID := uuid.New()
	soID := uuid.New()
	itemID := uuid.New()

	fakeService.FulfillItemsReturns(nil)

	reqBody := []FulfillItemRequestDTO{
		{
			ItemID:            itemID.String(),
			QuantityFulfilled: 10,
		},
	}
	bodyBytes, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/clients/"+clientID.String()+"/sales-orders/"+soID.String()+"/fulfill", bytes.NewReader(bodyBytes))
	req = mux.SetURLVars(req, map[string]string{
		"clientId": clientID.String(),
		"id":       soID.String(),
	})
	rec := httptest.NewRecorder()

	handler.Fulfill(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, 1, fakeService.FulfillItemsCallCount())

	// Verify items were passed correctly
	_, _, _, items := fakeService.FulfillItemsArgsForCall(0)
	assert.Len(t, items, 1)
	assert.Equal(t, itemID, items[0].ItemID)
	assert.Equal(t, 10, items[0].QuantityFulfilled)
}

func TestSalesOrderHandler_Fulfill_InvalidJSON(t *testing.T) {
	fakeService := &fakes.FakeSalesOrderService{}
	handler := NewSalesOrderHandler(fakeService)

	clientID := uuid.New()
	soID := uuid.New()

	req := httptest.NewRequest(http.MethodPost, "/api/v1/clients/"+clientID.String()+"/sales-orders/"+soID.String()+"/fulfill", bytes.NewReader([]byte("invalid json")))
	req = mux.SetURLVars(req, map[string]string{
		"clientId": clientID.String(),
		"id":       soID.String(),
	})
	rec := httptest.NewRecorder()

	handler.Fulfill(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Equal(t, 0, fakeService.FulfillItemsCallCount())
}

func TestSalesOrderHandler_Fulfill_ServiceError(t *testing.T) {
	fakeService := &fakes.FakeSalesOrderService{}
	handler := NewSalesOrderHandler(fakeService)

	clientID := uuid.New()
	soID := uuid.New()
	itemID := uuid.New()

	fakeService.FulfillItemsReturns(errors.New("insufficient inventory"))

	reqBody := []FulfillItemRequestDTO{
		{
			ItemID:            itemID.String(),
			QuantityFulfilled: 100,
		},
	}
	bodyBytes, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/clients/"+clientID.String()+"/sales-orders/"+soID.String()+"/fulfill", bytes.NewReader(bodyBytes))
	req = mux.SetURLVars(req, map[string]string{
		"clientId": clientID.String(),
		"id":       soID.String(),
	})
	rec := httptest.NewRecorder()

	handler.Fulfill(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}
