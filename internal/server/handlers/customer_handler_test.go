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

func TestCustomerHandler_Create_Success(t *testing.T) {
	fakeService := &fakes.FakeCustomerService{}
	handler := NewCustomerHandler(fakeService)

	clientID := uuid.New()
	customerID := uuid.New()

	// Configure fake
	fakeService.CreateCustomerReturns(&models.Customer{
		ID:        customerID,
		ClientID:  clientID,
		Code:      "CUST001",
		FirstName: "John",
		LastName:  "Doe",
		Email:     "john.doe@example.com",
		IsActive:  true,
	}, nil)

	// Create request body
	reqBody := CreateCustomerRequestDTO{
		Code:      "CUST001",
		FirstName: "John",
		LastName:  "Doe",
		Email:     "john.doe@example.com",
		IsActive:  true,
	}
	bodyBytes, _ := json.Marshal(reqBody)

	// Create HTTP request
	req := httptest.NewRequest(http.MethodPost, "/api/v1/clients/"+clientID.String()+"/customers", bytes.NewReader(bodyBytes))
	req = mux.SetURLVars(req, map[string]string{"clientId": clientID.String()})
	rec := httptest.NewRecorder()

	// Execute handler
	handler.Create(rec, req)

	// Verify response
	assert.Equal(t, http.StatusCreated, rec.Code)
	assert.Equal(t, 1, fakeService.CreateCustomerCallCount())

	// Verify response body
	var response APIResponse
	err := json.NewDecoder(rec.Body).Decode(&response)
	require.NoError(t, err)
	assert.True(t, response.Success)
}

func TestCustomerHandler_Create_InvalidJSON(t *testing.T) {
	fakeService := &fakes.FakeCustomerService{}
	handler := NewCustomerHandler(fakeService)

	clientID := uuid.New()

	// Create invalid JSON
	req := httptest.NewRequest(http.MethodPost, "/api/v1/clients/"+clientID.String()+"/customers", bytes.NewReader([]byte("invalid json")))
	req = mux.SetURLVars(req, map[string]string{"clientId": clientID.String()})
	rec := httptest.NewRecorder()

	handler.Create(rec, req)

	// Verify 400 status
	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Equal(t, 0, fakeService.CreateCustomerCallCount())
}

func TestCustomerHandler_Create_MissingFields(t *testing.T) {
	fakeService := &fakes.FakeCustomerService{}
	handler := NewCustomerHandler(fakeService)

	clientID := uuid.New()

	// Create request with missing required fields
	reqBody := CreateCustomerRequestDTO{
		// Missing Code, FirstName, LastName
		Email: "john.doe@example.com",
	}
	bodyBytes, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/clients/"+clientID.String()+"/customers", bytes.NewReader(bodyBytes))
	req = mux.SetURLVars(req, map[string]string{"clientId": clientID.String()})
	rec := httptest.NewRecorder()

	handler.Create(rec, req)

	// Verify 400 status
	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Equal(t, 0, fakeService.CreateCustomerCallCount())
}

func TestCustomerHandler_Create_ValidationError(t *testing.T) {
	fakeService := &fakes.FakeCustomerService{}
	handler := NewCustomerHandler(fakeService)

	clientID := uuid.New()

	// Configure fake to return validation error
	fakeService.CreateCustomerReturns(nil, errors.New("invalid email format"))

	reqBody := CreateCustomerRequestDTO{
		Code:      "CUST001",
		FirstName: "John",
		LastName:  "Doe",
		Email:     "invalid-email",
	}
	bodyBytes, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/clients/"+clientID.String()+"/customers", bytes.NewReader(bodyBytes))
	req = mux.SetURLVars(req, map[string]string{"clientId": clientID.String()})
	rec := httptest.NewRecorder()

	handler.Create(rec, req)

	// Verify 400 status
	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Equal(t, 1, fakeService.CreateCustomerCallCount())
}

func TestCustomerHandler_Get_Success(t *testing.T) {
	fakeService := &fakes.FakeCustomerService{}
	handler := NewCustomerHandler(fakeService)

	clientID := uuid.New()
	customerID := uuid.New()

	// Configure fake
	fakeService.GetCustomerReturns(&models.Customer{
		ID:        customerID,
		ClientID:  clientID,
		Code:      "CUST001",
		FirstName: "John",
		LastName:  "Doe",
	}, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/clients/"+clientID.String()+"/customers/"+customerID.String(), nil)
	req = mux.SetURLVars(req, map[string]string{
		"clientId": clientID.String(),
		"id":       customerID.String(),
	})
	rec := httptest.NewRecorder()

	handler.Get(rec, req)

	// Verify response
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, 1, fakeService.GetCustomerCallCount())
}

func TestCustomerHandler_Get_NotFound(t *testing.T) {
	fakeService := &fakes.FakeCustomerService{}
	handler := NewCustomerHandler(fakeService)

	clientID := uuid.New()
	customerID := uuid.New()

	// Configure fake to return error
	fakeService.GetCustomerReturns(nil, errors.New("customer not found"))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/clients/"+clientID.String()+"/customers/"+customerID.String(), nil)
	req = mux.SetURLVars(req, map[string]string{
		"clientId": clientID.String(),
		"id":       customerID.String(),
	})
	rec := httptest.NewRecorder()

	handler.Get(rec, req)

	// Verify 404 status
	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestCustomerHandler_Update_Success(t *testing.T) {
	fakeService := &fakes.FakeCustomerService{}
	handler := NewCustomerHandler(fakeService)

	clientID := uuid.New()
	customerID := uuid.New()

	// Configure fake
	fakeService.UpdateCustomerReturns(nil)

	newFirstName := "Jane"
	reqBody := UpdateCustomerRequestDTO{
		FirstName: &newFirstName,
	}
	bodyBytes, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPut, "/api/v1/clients/"+clientID.String()+"/customers/"+customerID.String(), bytes.NewReader(bodyBytes))
	req = mux.SetURLVars(req, map[string]string{
		"clientId": clientID.String(),
		"id":       customerID.String(),
	})
	rec := httptest.NewRecorder()

	handler.Update(rec, req)

	// Verify response
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, 1, fakeService.UpdateCustomerCallCount())
}

func TestCustomerHandler_Update_InvalidJSON(t *testing.T) {
	fakeService := &fakes.FakeCustomerService{}
	handler := NewCustomerHandler(fakeService)

	clientID := uuid.New()
	customerID := uuid.New()

	req := httptest.NewRequest(http.MethodPut, "/api/v1/clients/"+clientID.String()+"/customers/"+customerID.String(), bytes.NewReader([]byte("invalid json")))
	req = mux.SetURLVars(req, map[string]string{
		"clientId": clientID.String(),
		"id":       customerID.String(),
	})
	rec := httptest.NewRecorder()

	handler.Update(rec, req)

	// Verify 400 status
	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Equal(t, 0, fakeService.UpdateCustomerCallCount())
}

func TestCustomerHandler_Delete_Success(t *testing.T) {
	fakeService := &fakes.FakeCustomerService{}
	handler := NewCustomerHandler(fakeService)

	clientID := uuid.New()
	customerID := uuid.New()

	// Configure fake
	fakeService.DeleteCustomerReturns(nil)

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/clients/"+clientID.String()+"/customers/"+customerID.String(), nil)
	req = mux.SetURLVars(req, map[string]string{
		"clientId": clientID.String(),
		"id":       customerID.String(),
	})
	rec := httptest.NewRecorder()

	handler.Delete(rec, req)

	// Verify response
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, 1, fakeService.DeleteCustomerCallCount())
}

func TestCustomerHandler_Delete_WithActiveSalesOrders(t *testing.T) {
	fakeService := &fakes.FakeCustomerService{}
	handler := NewCustomerHandler(fakeService)

	clientID := uuid.New()
	customerID := uuid.New()

	// Configure fake to return error
	fakeService.DeleteCustomerReturns(errors.New("violates foreign key constraint"))

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/clients/"+clientID.String()+"/customers/"+customerID.String(), nil)
	req = mux.SetURLVars(req, map[string]string{
		"clientId": clientID.String(),
		"id":       customerID.String(),
	})
	rec := httptest.NewRecorder()

	handler.Delete(rec, req)

	// Verify 400 status
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestCustomerHandler_List_Success(t *testing.T) {
	fakeService := &fakes.FakeCustomerService{}
	handler := NewCustomerHandler(fakeService)

	clientID := uuid.New()

	// Configure fake
	customers := []*models.Customer{
		{ID: uuid.New(), ClientID: clientID, Code: "CUST001", FirstName: "John", LastName: "Doe"},
		{ID: uuid.New(), ClientID: clientID, Code: "CUST002", FirstName: "Jane", LastName: "Smith"},
	}
	fakeService.ListCustomersReturns(customers, 2, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/clients/"+clientID.String()+"/customers", nil)
	req = mux.SetURLVars(req, map[string]string{"clientId": clientID.String()})
	rec := httptest.NewRecorder()

	handler.List(rec, req)

	// Verify response
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, 1, fakeService.ListCustomersCallCount())
}

func TestCustomerHandler_Search_Success(t *testing.T) {
	fakeService := &fakes.FakeCustomerService{}
	handler := NewCustomerHandler(fakeService)

	clientID := uuid.New()

	// Configure fake
	customers := []*models.Customer{
		{ID: uuid.New(), ClientID: clientID, Code: "CUST001", FirstName: "John", LastName: "Doe"},
	}
	fakeService.SearchCustomersReturns(customers, 1, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/clients/"+clientID.String()+"/customers/search?q=john", nil)
	req = mux.SetURLVars(req, map[string]string{"clientId": clientID.String()})
	rec := httptest.NewRecorder()

	handler.Search(rec, req)

	// Verify response
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, 1, fakeService.SearchCustomersCallCount())
}

func TestCustomerHandler_Search_MissingQuery(t *testing.T) {
	fakeService := &fakes.FakeCustomerService{}
	handler := NewCustomerHandler(fakeService)

	clientID := uuid.New()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/clients/"+clientID.String()+"/customers/search", nil)
	req = mux.SetURLVars(req, map[string]string{"clientId": clientID.String()})
	rec := httptest.NewRecorder()

	handler.Search(rec, req)

	// Verify 400 status
	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Equal(t, 0, fakeService.SearchCustomersCallCount())
}
