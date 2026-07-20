package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/yourorg/shoppilot/internal/services"
	"github.com/yourorg/shoppilot/internal/utils"
)

// CustomerHandler handles customer-related HTTP requests
type CustomerHandler struct {
	service services.CustomerService
}

// NewCustomerHandler creates a new customer handler
func NewCustomerHandler(service services.CustomerService) *CustomerHandler {
	return &CustomerHandler{
		service: service,
	}
}

// Request DTOs

type CreateCustomerRequestDTO struct {
	Code               string                 `json:"code"`
	FirstName          string                 `json:"firstName"`
	LastName           string                 `json:"lastName"`
	Email              string                 `json:"email"`
	Phone              string                 `json:"phone"`
	ShippingAddress    string                 `json:"shippingAddress"`
	ShippingCity       string                 `json:"shippingCity"`
	ShippingState      string                 `json:"shippingState"`
	ShippingPostalCode string                 `json:"shippingPostalCode"`
	ShippingCountry    string                 `json:"shippingCountry"`
	BillingAddress     string                 `json:"billingAddress"`
	BillingCity        string                 `json:"billingCity"`
	BillingState       string                 `json:"billingState"`
	BillingPostalCode  string                 `json:"billingPostalCode"`
	BillingCountry     string                 `json:"billingCountry"`
	TaxID              string                 `json:"taxId"`
	Notes              string                 `json:"notes"`
	IsActive           bool                   `json:"isActive"`
	Metadata           map[string]interface{} `json:"metadata,omitempty"`
}

type UpdateCustomerRequestDTO struct {
	FirstName          *string                `json:"firstName,omitempty"`
	LastName           *string                `json:"lastName,omitempty"`
	Email              *string                `json:"email,omitempty"`
	Phone              *string                `json:"phone,omitempty"`
	ShippingAddress    *string                `json:"shippingAddress,omitempty"`
	ShippingCity       *string                `json:"shippingCity,omitempty"`
	ShippingState      *string                `json:"shippingState,omitempty"`
	ShippingPostalCode *string                `json:"shippingPostalCode,omitempty"`
	ShippingCountry    *string                `json:"shippingCountry,omitempty"`
	BillingAddress     *string                `json:"billingAddress,omitempty"`
	BillingCity        *string                `json:"billingCity,omitempty"`
	BillingState       *string                `json:"billingState,omitempty"`
	BillingPostalCode  *string                `json:"billingPostalCode,omitempty"`
	BillingCountry     *string                `json:"billingCountry,omitempty"`
	TaxID              *string                `json:"taxId,omitempty"`
	Notes              *string                `json:"notes,omitempty"`
	IsActive           *bool                  `json:"isActive,omitempty"`
	Metadata           map[string]interface{} `json:"metadata,omitempty"`
}

// Customer handlers

// Create handles POST /api/v1/customers
func (h *CustomerHandler) Create(w http.ResponseWriter, r *http.Request) {
	clientID, err := utils.GetClientIDFromContext(r.Context())
	if err != nil {
		writeError(w, http.StatusUnauthorized, "NO_CLIENT_CONTEXT", err.Error())
		return
	}

	var dto CreateCustomerRequestDTO
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	// Validate required fields
	if dto.Code == "" || dto.FirstName == "" || dto.LastName == "" {
		writeError(w, http.StatusBadRequest, "MISSING_FIELDS", "Code, first name, and last name are required")
		return
	}

	req := &services.CreateCustomerRequest{
		Code:               dto.Code,
		FirstName:          dto.FirstName,
		LastName:           dto.LastName,
		Email:              dto.Email,
		Phone:              dto.Phone,
		ShippingAddress:    dto.ShippingAddress,
		ShippingCity:       dto.ShippingCity,
		ShippingState:      dto.ShippingState,
		ShippingPostalCode: dto.ShippingPostalCode,
		ShippingCountry:    dto.ShippingCountry,
		BillingAddress:     dto.BillingAddress,
		BillingCity:        dto.BillingCity,
		BillingState:       dto.BillingState,
		BillingPostalCode:  dto.BillingPostalCode,
		BillingCountry:     dto.BillingCountry,
		TaxID:              dto.TaxID,
		Notes:              dto.Notes,
		IsActive:           dto.IsActive,
		Metadata:           dto.Metadata,
	}

	customer, err := h.service.CreateCustomer(r.Context(), clientID, req)
	if err != nil {
		writeError(w, http.StatusBadRequest, "CREATE_FAILED", err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, APIResponse{
		Success: true,
		Data:    customer,
	})
}

// Get handles GET /api/v1/customers/:id
func (h *CustomerHandler) Get(w http.ResponseWriter, r *http.Request) {
	clientID, err := utils.GetClientIDFromContext(r.Context())
	if err != nil {
		writeError(w, http.StatusUnauthorized, "NO_CLIENT_CONTEXT", err.Error())
		return
	}

	vars := mux.Vars(r)
	customerID, err := uuid.Parse(vars["id"])
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_CUSTOMER_ID", "Invalid customer ID format")
		return
	}

	customer, err := h.service.GetCustomer(r.Context(), clientID, customerID)
	if err != nil {
		writeError(w, http.StatusNotFound, "CUSTOMER_NOT_FOUND", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, APIResponse{
		Success: true,
		Data:    customer,
	})
}

// Update handles PUT /api/v1/customers/:id
func (h *CustomerHandler) Update(w http.ResponseWriter, r *http.Request) {
	clientID, err := utils.GetClientIDFromContext(r.Context())
	if err != nil {
		writeError(w, http.StatusUnauthorized, "NO_CLIENT_CONTEXT", err.Error())
		return
	}

	vars := mux.Vars(r)
	customerID, err := uuid.Parse(vars["id"])
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_CUSTOMER_ID", "Invalid customer ID format")
		return
	}

	var dto UpdateCustomerRequestDTO
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	req := &services.UpdateCustomerRequest{
		FirstName:          dto.FirstName,
		LastName:           dto.LastName,
		Email:              dto.Email,
		Phone:              dto.Phone,
		ShippingAddress:    dto.ShippingAddress,
		ShippingCity:       dto.ShippingCity,
		ShippingState:      dto.ShippingState,
		ShippingPostalCode: dto.ShippingPostalCode,
		ShippingCountry:    dto.ShippingCountry,
		BillingAddress:     dto.BillingAddress,
		BillingCity:        dto.BillingCity,
		BillingState:       dto.BillingState,
		BillingPostalCode:  dto.BillingPostalCode,
		BillingCountry:     dto.BillingCountry,
		TaxID:              dto.TaxID,
		Notes:              dto.Notes,
		IsActive:           dto.IsActive,
		Metadata:           dto.Metadata,
	}

	if err := h.service.UpdateCustomer(r.Context(), clientID, customerID, req); err != nil {
		writeError(w, http.StatusBadRequest, "UPDATE_FAILED", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, APIResponse{
		Success: true,
		Data:    map[string]string{"message": "Customer updated successfully"},
	})
}

// Delete handles DELETE /api/v1/customers/:id
func (h *CustomerHandler) Delete(w http.ResponseWriter, r *http.Request) {
	clientID, err := utils.GetClientIDFromContext(r.Context())
	if err != nil {
		writeError(w, http.StatusUnauthorized, "NO_CLIENT_CONTEXT", err.Error())
		return
	}

	vars := mux.Vars(r)
	customerID, err := uuid.Parse(vars["id"])
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_CUSTOMER_ID", "Invalid customer ID format")
		return
	}

	if err := h.service.DeleteCustomer(r.Context(), clientID, customerID); err != nil {
		writeError(w, http.StatusBadRequest, "DELETE_FAILED", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, APIResponse{
		Success: true,
		Data:    map[string]string{"message": "Customer deleted successfully"},
	})
}

// List handles GET /api/v1/customers
func (h *CustomerHandler) List(w http.ResponseWriter, r *http.Request) {
	clientID, err := utils.GetClientIDFromContext(r.Context())
	if err != nil {
		writeError(w, http.StatusUnauthorized, "NO_CLIENT_CONTEXT", err.Error())
		return
	}

	// Parse pagination parameters
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}

	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))
	if pageSize < 1 {
		pageSize = 20
	}

	customers, total, err := h.service.ListCustomers(r.Context(), clientID, page, pageSize)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "LIST_FAILED", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, APIResponse{
		Success: true,
		Data: map[string]interface{}{
			"customers":  customers,
			"pagination": map[string]int{"page": page, "pageSize": pageSize, "total": total},
		},
	})
}

// Search handles GET /api/v1/customers/search
func (h *CustomerHandler) Search(w http.ResponseWriter, r *http.Request) {
	clientID, err := utils.GetClientIDFromContext(r.Context())
	if err != nil {
		writeError(w, http.StatusUnauthorized, "NO_CLIENT_CONTEXT", err.Error())
		return
	}

	// Get search query
	query := r.URL.Query().Get("q")
	if query == "" {
		writeError(w, http.StatusBadRequest, "MISSING_QUERY", "Search query parameter 'q' is required")
		return
	}

	// Parse pagination parameters
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}

	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))
	if pageSize < 1 {
		pageSize = 20
	}

	customers, total, err := h.service.SearchCustomers(r.Context(), clientID, query, page, pageSize)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "SEARCH_FAILED", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, APIResponse{
		Success: true,
		Data: map[string]interface{}{
			"customers":  customers,
			"pagination": map[string]int{"page": page, "pageSize": pageSize, "total": total},
		},
	})
}
