package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yourorg/shoppilot/internal/models"
	"github.com/yourorg/shoppilot/internal/server/handlers/fakes"
)

func TestSupplierHandler_Create_Success(t *testing.T) {
	fakeService := &fakes.FakeSupplierService{}
	handler := NewSupplierHandler(fakeService)

	clientID := uuid.New()
	supplierID := uuid.New()

	// Configure fake
	fakeService.CreateSupplierReturns(&models.Supplier{
		ID:       supplierID,
		ClientID: clientID,
		Code:     "SUP001",
		Name:     "Test Supplier",
		Email:    "test@supplier.com",
		IsActive: true,
	}, nil)

	// Create request body
	reqBody := CreateSupplierRequestDTO{
		Code:     "SUP001",
		Name:     "Test Supplier",
		Email:    "test@supplier.com",
		IsActive: true,
	}
	bodyBytes, _ := json.Marshal(reqBody)

	// Create HTTP request
	req := httptest.NewRequest(http.MethodPost, "/api/v1/clients/"+clientID.String()+"/suppliers", bytes.NewReader(bodyBytes))
	req = mux.SetURLVars(req, map[string]string{"clientId": clientID.String()})
	rec := httptest.NewRecorder()

	// Execute handler
	handler.Create(rec, req)

	// Verify response
	assert.Equal(t, http.StatusCreated, rec.Code)
	assert.Equal(t, 1, fakeService.CreateSupplierCallCount())

	// Verify response body
	var response APIResponse
	err := json.NewDecoder(rec.Body).Decode(&response)
	require.NoError(t, err)
	assert.True(t, response.Success)
}

func TestSupplierHandler_Create_InvalidJSON(t *testing.T) {
	fakeService := &fakes.FakeSupplierService{}
	handler := NewSupplierHandler(fakeService)

	clientID := uuid.New()

	// Create invalid JSON
	req := httptest.NewRequest(http.MethodPost, "/api/v1/clients/"+clientID.String()+"/suppliers", bytes.NewReader([]byte("invalid json")))
	req = mux.SetURLVars(req, map[string]string{"clientId": clientID.String()})
	rec := httptest.NewRecorder()

	handler.Create(rec, req)

	// Verify 400 status
	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Equal(t, 0, fakeService.CreateSupplierCallCount())
}

func TestSupplierHandler_Create_MissingFields(t *testing.T) {
	fakeService := &fakes.FakeSupplierService{}
	handler := NewSupplierHandler(fakeService)

	clientID := uuid.New()

	// Create request with missing required fields
	reqBody := CreateSupplierRequestDTO{
		// Missing Code and Name
		Email: "test@supplier.com",
	}
	bodyBytes, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/clients/"+clientID.String()+"/suppliers", bytes.NewReader(bodyBytes))
	req = mux.SetURLVars(req, map[string]string{"clientId": clientID.String()})
	rec := httptest.NewRecorder()

	handler.Create(rec, req)

	// Verify 400 status
	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Equal(t, 0, fakeService.CreateSupplierCallCount())
}

func TestSupplierHandler_Create_ValidationError(t *testing.T) {
	fakeService := &fakes.FakeSupplierService{}
	handler := NewSupplierHandler(fakeService)

	clientID := uuid.New()

	// Configure fake to return validation error
	fakeService.CreateSupplierReturns(nil, errors.New("invalid email format"))

	reqBody := CreateSupplierRequestDTO{
		Code:  "SUP001",
		Name:  "Test Supplier",
		Email: "invalid-email",
	}
	bodyBytes, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/clients/"+clientID.String()+"/suppliers", bytes.NewReader(bodyBytes))
	req = mux.SetURLVars(req, map[string]string{"clientId": clientID.String()})
	rec := httptest.NewRecorder()

	handler.Create(rec, req)

	// Verify 400 status
	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Equal(t, 1, fakeService.CreateSupplierCallCount())
}

func TestSupplierHandler_Get_Success(t *testing.T) {
	fakeService := &fakes.FakeSupplierService{}
	handler := NewSupplierHandler(fakeService)

	clientID := uuid.New()
	supplierID := uuid.New()

	// Configure fake
	fakeService.GetSupplierReturns(&models.Supplier{
		ID:       supplierID,
		ClientID: clientID,
		Code:     "SUP001",
		Name:     "Test Supplier",
	}, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/clients/"+clientID.String()+"/suppliers/"+supplierID.String(), nil)
	req = mux.SetURLVars(req, map[string]string{
		"clientId": clientID.String(),
		"id":       supplierID.String(),
	})
	rec := httptest.NewRecorder()

	handler.Get(rec, req)

	// Verify response
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, 1, fakeService.GetSupplierCallCount())
}

func TestSupplierHandler_Get_NotFound(t *testing.T) {
	fakeService := &fakes.FakeSupplierService{}
	handler := NewSupplierHandler(fakeService)

	clientID := uuid.New()
	supplierID := uuid.New()

	// Configure fake to return error
	fakeService.GetSupplierReturns(nil, errors.New("supplier not found"))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/clients/"+clientID.String()+"/suppliers/"+supplierID.String(), nil)
	req = mux.SetURLVars(req, map[string]string{
		"clientId": clientID.String(),
		"id":       supplierID.String(),
	})
	rec := httptest.NewRecorder()

	handler.Get(rec, req)

	// Verify 404 status
	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestSupplierHandler_Update_Success(t *testing.T) {
	fakeService := &fakes.FakeSupplierService{}
	handler := NewSupplierHandler(fakeService)

	clientID := uuid.New()
	supplierID := uuid.New()

	// Configure fake
	fakeService.UpdateSupplierReturns(nil)

	newName := "Updated Supplier"
	reqBody := UpdateSupplierRequestDTO{
		Name: &newName,
	}
	bodyBytes, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPut, "/api/v1/clients/"+clientID.String()+"/suppliers/"+supplierID.String(), bytes.NewReader(bodyBytes))
	req = mux.SetURLVars(req, map[string]string{
		"clientId": clientID.String(),
		"id":       supplierID.String(),
	})
	rec := httptest.NewRecorder()

	handler.Update(rec, req)

	// Verify response
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, 1, fakeService.UpdateSupplierCallCount())
}

func TestSupplierHandler_Update_InvalidJSON(t *testing.T) {
	fakeService := &fakes.FakeSupplierService{}
	handler := NewSupplierHandler(fakeService)

	clientID := uuid.New()
	supplierID := uuid.New()

	req := httptest.NewRequest(http.MethodPut, "/api/v1/clients/"+clientID.String()+"/suppliers/"+supplierID.String(), bytes.NewReader([]byte("invalid json")))
	req = mux.SetURLVars(req, map[string]string{
		"clientId": clientID.String(),
		"id":       supplierID.String(),
	})
	rec := httptest.NewRecorder()

	handler.Update(rec, req)

	// Verify 400 status
	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Equal(t, 0, fakeService.UpdateSupplierCallCount())
}

func TestSupplierHandler_Delete_Success(t *testing.T) {
	fakeService := &fakes.FakeSupplierService{}
	handler := NewSupplierHandler(fakeService)

	clientID := uuid.New()
	supplierID := uuid.New()

	// Configure fake
	fakeService.DeleteSupplierReturns(nil)

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/clients/"+clientID.String()+"/suppliers/"+supplierID.String(), nil)
	req = mux.SetURLVars(req, map[string]string{
		"clientId": clientID.String(),
		"id":       supplierID.String(),
	})
	rec := httptest.NewRecorder()

	handler.Delete(rec, req)

	// Verify response
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, 1, fakeService.DeleteSupplierCallCount())
}

func TestSupplierHandler_Delete_WithActivePurchaseOrders(t *testing.T) {
	fakeService := &fakes.FakeSupplierService{}
	handler := NewSupplierHandler(fakeService)

	clientID := uuid.New()
	supplierID := uuid.New()

	// Configure fake to return error
	fakeService.DeleteSupplierReturns(errors.New("violates foreign key constraint"))

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/clients/"+clientID.String()+"/suppliers/"+supplierID.String(), nil)
	req = mux.SetURLVars(req, map[string]string{
		"clientId": clientID.String(),
		"id":       supplierID.String(),
	})
	rec := httptest.NewRecorder()

	handler.Delete(rec, req)

	// Verify 400 status
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestSupplierHandler_List_Success(t *testing.T) {
	fakeService := &fakes.FakeSupplierService{}
	handler := NewSupplierHandler(fakeService)

	clientID := uuid.New()

	// Configure fake
	suppliers := []*models.Supplier{
		{ID: uuid.New(), ClientID: clientID, Code: "SUP001", Name: "Supplier 1"},
		{ID: uuid.New(), ClientID: clientID, Code: "SUP002", Name: "Supplier 2"},
	}
	fakeService.ListSuppliersReturns(suppliers, 2, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/clients/"+clientID.String()+"/suppliers", nil)
	req = mux.SetURLVars(req, map[string]string{"clientId": clientID.String()})
	rec := httptest.NewRecorder()

	handler.List(rec, req)

	// Verify response
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, 1, fakeService.ListSuppliersCallCount())
}

func TestSupplierHandler_List_ActiveOnly(t *testing.T) {
	fakeService := &fakes.FakeSupplierService{}
	handler := NewSupplierHandler(fakeService)

	clientID := uuid.New()

	// Configure fake
	activeSuppliers := []*models.Supplier{
		{ID: uuid.New(), ClientID: clientID, Code: "SUP001", Name: "Active Supplier", IsActive: true},
	}
	fakeService.ListActiveSuppliersReturns(activeSuppliers, 1, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/clients/"+clientID.String()+"/suppliers?active=true", nil)
	req = mux.SetURLVars(req, map[string]string{"clientId": clientID.String()})
	rec := httptest.NewRecorder()

	handler.List(rec, req)

	// Verify response
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, 1, fakeService.ListActiveSuppliersCallCount())
	assert.Equal(t, 0, fakeService.ListSuppliersCallCount())
}
