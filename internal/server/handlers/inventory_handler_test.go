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

func TestInventoryHandler_GetMovements_Success(t *testing.T) {
	fakeService := &fakes.FakeProductService{}
	handler := NewProductHandler(fakeService)

	clientID := uuid.New()
	variantID := uuid.New()

	movements := []*models.InventoryMovement{
		{
			ID:               uuid.New(),
			ClientID:         clientID,
			VariantID:        variantID,
			MovementType:     models.MovementTypePurchase,
			Quantity:         10,
			PreviousQuantity: 0,
			NewQuantity:      10,
			CreatedAt:     time.Now(),
		},
		{
			ID:               uuid.New(),
			ClientID:         clientID,
			VariantID:        variantID,
			MovementType:     models.MovementTypeSale,
			Quantity:         -5,
			PreviousQuantity: 10,
			NewQuantity:      5,
			CreatedAt:     time.Now(),
		},
	}
	fakeService.GetMovementHistoryReturns(movements, 2, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/clients/"+clientID.String()+"/variants/"+variantID.String()+"/movements", nil)
	req = mux.SetURLVars(req, map[string]string{
		"clientId": clientID.String(),
		"id":       variantID.String(),
	})
	rec := httptest.NewRecorder()

	handler.GetMovements(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, 1, fakeService.GetMovementHistoryCallCount())

	// Verify response body
	var response APIResponse
	err := json.NewDecoder(rec.Body).Decode(&response)
	require.NoError(t, err)
	assert.True(t, response.Success)
}

func TestInventoryHandler_GetMovements_WithPagination(t *testing.T) {
	fakeService := &fakes.FakeProductService{}
	handler := NewProductHandler(fakeService)

	clientID := uuid.New()
	variantID := uuid.New()

	fakeService.GetMovementHistoryReturns([]*models.InventoryMovement{}, 0, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/clients/"+clientID.String()+"/variants/"+variantID.String()+"/movements?page=2&pageSize=10", nil)
	req = mux.SetURLVars(req, map[string]string{
		"clientId": clientID.String(),
		"id":       variantID.String(),
	})
	rec := httptest.NewRecorder()

	handler.GetMovements(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	// Verify pagination was passed correctly
	_, _, _, page, pageSize := fakeService.GetMovementHistoryArgsForCall(0)
	assert.Equal(t, 2, page)
	assert.Equal(t, 10, pageSize)
}

func TestInventoryHandler_RecordMovement_Success(t *testing.T) {
	fakeService := &fakes.FakeProductService{}
	handler := NewProductHandler(fakeService)

	clientID := uuid.New()
	variantID := uuid.New()
	shopID := uuid.New()
	movementID := uuid.New()

	fakeService.RecordMovementReturns(&models.InventoryMovement{
		ID:               movementID,
		ClientID:         clientID,
		VariantID:        variantID,
		ShopID:           shopID,
		MovementType:     models.MovementTypeAdjustment,
		Quantity:         10,
		PreviousQuantity: 0,
		NewQuantity:      10,
		CreatedAt:     time.Now(),
	}, nil)

	reqBody := RecordMovementRequestDTO{
		ShopID:       shopID.String(),
		MovementType: string(models.MovementTypeAdjustment),
		Quantity:     10,
		Notes:        "Manual adjustment",
	}
	bodyBytes, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/clients/"+clientID.String()+"/variants/"+variantID.String()+"/movements", bytes.NewReader(bodyBytes))
	req = mux.SetURLVars(req, map[string]string{
		"clientId": clientID.String(),
		"id":       variantID.String(),
	})
	rec := httptest.NewRecorder()

	handler.RecordMovement(rec, req)

	assert.Equal(t, http.StatusCreated, rec.Code)
	assert.Equal(t, 1, fakeService.RecordMovementCallCount())
}

func TestInventoryHandler_RecordMovement_InvalidJSON(t *testing.T) {
	fakeService := &fakes.FakeProductService{}
	handler := NewProductHandler(fakeService)

	clientID := uuid.New()
	variantID := uuid.New()

	req := httptest.NewRequest(http.MethodPost, "/api/v1/clients/"+clientID.String()+"/variants/"+variantID.String()+"/movements", bytes.NewReader([]byte("invalid json")))
	req = mux.SetURLVars(req, map[string]string{
		"clientId": clientID.String(),
		"id":       variantID.String(),
	})
	rec := httptest.NewRecorder()

	handler.RecordMovement(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Equal(t, 0, fakeService.RecordMovementCallCount())
}

func TestInventoryHandler_RecordMovement_MissingFields(t *testing.T) {
	fakeService := &fakes.FakeProductService{}
	handler := NewProductHandler(fakeService)

	clientID := uuid.New()
	variantID := uuid.New()

	reqBody := RecordMovementRequestDTO{
		// Missing ShopID and MovementType
		Quantity: 10,
	}
	bodyBytes, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/clients/"+clientID.String()+"/variants/"+variantID.String()+"/movements", bytes.NewReader(bodyBytes))
	req = mux.SetURLVars(req, map[string]string{
		"clientId": clientID.String(),
		"id":       variantID.String(),
	})
	rec := httptest.NewRecorder()

	handler.RecordMovement(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Equal(t, 0, fakeService.RecordMovementCallCount())
}

func TestInventoryHandler_SetAlert_Success(t *testing.T) {
	fakeService := &fakes.FakeProductService{}
	handler := NewProductHandler(fakeService)

	clientID := uuid.New()
	variantID := uuid.New()
	shopID := uuid.New()
	alertID := uuid.New()

	fakeService.SetInventoryAlertReturns(&models.InventoryAlert{
		ID:                alertID,
		ClientID:          clientID,
		VariantID:         variantID,
		ShopID:            shopID,
		ReorderPoint:      10,
		ReorderQuantity:   50,
		LowStockThreshold: 5,
		IsEnabled:         true,
	}, nil)

	reqBody := SetInventoryAlertRequestDTO{
		ReorderPoint:      10,
		ReorderQuantity:   50,
		LowStockThreshold: 5,
		IsEnabled:         true,
	}
	bodyBytes, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPut, "/api/v1/clients/"+clientID.String()+"/variants/"+variantID.String()+"/alerts?shopId="+shopID.String(), bytes.NewReader(bodyBytes))
	req = mux.SetURLVars(req, map[string]string{
		"clientId": clientID.String(),
		"id":       variantID.String(),
	})
	rec := httptest.NewRecorder()

	handler.SetAlert(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, 1, fakeService.SetInventoryAlertCallCount())
}

func TestInventoryHandler_SetAlert_MissingShopID(t *testing.T) {
	fakeService := &fakes.FakeProductService{}
	handler := NewProductHandler(fakeService)

	clientID := uuid.New()
	variantID := uuid.New()

	reqBody := SetInventoryAlertRequestDTO{
		ReorderPoint:      10,
		ReorderQuantity:   50,
		LowStockThreshold: 5,
		IsEnabled:         true,
	}
	bodyBytes, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPut, "/api/v1/clients/"+clientID.String()+"/variants/"+variantID.String()+"/alerts", bytes.NewReader(bodyBytes))
	req = mux.SetURLVars(req, map[string]string{
		"clientId": clientID.String(),
		"id":       variantID.String(),
	})
	rec := httptest.NewRecorder()

	handler.SetAlert(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Equal(t, 0, fakeService.SetInventoryAlertCallCount())
}

func TestInventoryHandler_GetAlert_Success(t *testing.T) {
	fakeService := &fakes.FakeProductService{}
	handler := NewProductHandler(fakeService)

	clientID := uuid.New()
	variantID := uuid.New()
	shopID := uuid.New()

	fakeService.GetInventoryAlertReturns(&models.InventoryAlert{
		ID:                uuid.New(),
		ClientID:          clientID,
		VariantID:         variantID,
		ShopID:            shopID,
		ReorderPoint:      10,
		ReorderQuantity:   50,
		LowStockThreshold: 5,
		IsEnabled:         true,
	}, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/clients/"+clientID.String()+"/variants/"+variantID.String()+"/alerts?shopId="+shopID.String(), nil)
	req = mux.SetURLVars(req, map[string]string{
		"clientId": clientID.String(),
		"id":       variantID.String(),
	})
	rec := httptest.NewRecorder()

	handler.GetAlert(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, 1, fakeService.GetInventoryAlertCallCount())
}

func TestInventoryHandler_GetAlert_NotFound(t *testing.T) {
	fakeService := &fakes.FakeProductService{}
	handler := NewProductHandler(fakeService)

	clientID := uuid.New()
	variantID := uuid.New()
	shopID := uuid.New()

	fakeService.GetInventoryAlertReturns(nil, errors.New("alert not found"))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/clients/"+clientID.String()+"/variants/"+variantID.String()+"/alerts?shopId="+shopID.String(), nil)
	req = mux.SetURLVars(req, map[string]string{
		"clientId": clientID.String(),
		"id":       variantID.String(),
	})
	rec := httptest.NewRecorder()

	handler.GetAlert(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestInventoryHandler_GetLowStock_Success(t *testing.T) {
	fakeService := &fakes.FakeProductService{}
	handler := NewProductHandler(fakeService)

	clientID := uuid.New()
	shopID := uuid.New()

	alerts := []*models.InventoryAlert{
		{
			ID:                uuid.New(),
			ClientID:          clientID,
			ShopID:            shopID,
			ReorderPoint:      10,
			LowStockThreshold: 5,
			IsEnabled:         true,
		},
		{
			ID:                uuid.New(),
			ClientID:          clientID,
			ShopID:            shopID,
			ReorderPoint:      20,
			LowStockThreshold: 10,
			IsEnabled:         true,
		},
	}
	fakeService.CheckLowStockAlertsReturns(alerts, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/clients/"+clientID.String()+"/shops/"+shopID.String()+"/low-stock", nil)
	req = mux.SetURLVars(req, map[string]string{
		"clientId": clientID.String(),
		"shopId":   shopID.String(),
	})
	rec := httptest.NewRecorder()

	handler.GetLowStock(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, 1, fakeService.CheckLowStockAlertsCallCount())

	// Verify response contains alerts
	var response APIResponse
	err := json.NewDecoder(rec.Body).Decode(&response)
	require.NoError(t, err)
	assert.True(t, response.Success)
}

func TestInventoryHandler_GetLowStock_ServiceError(t *testing.T) {
	fakeService := &fakes.FakeProductService{}
	handler := NewProductHandler(fakeService)

	clientID := uuid.New()
	shopID := uuid.New()

	fakeService.CheckLowStockAlertsReturns(nil, errors.New("database error"))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/clients/"+clientID.String()+"/shops/"+shopID.String()+"/low-stock", nil)
	req = mux.SetURLVars(req, map[string]string{
		"clientId": clientID.String(),
		"shopId":   shopID.String(),
	})
	rec := httptest.NewRecorder()

	handler.GetLowStock(rec, req)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}
