package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/yourorg/shoppilot/internal/models"
	"github.com/yourorg/shoppilot/internal/services"
	"github.com/yourorg/shoppilot/internal/utils"
)

// SupplierHandler handles supplier-related HTTP requests
type SupplierHandler struct {
	service services.SupplierService
}

// NewSupplierHandler creates a new supplier handler
func NewSupplierHandler(service services.SupplierService) *SupplierHandler {
	return &SupplierHandler{
		service: service,
	}
}

// Request DTOs

type CreateSupplierRequestDTO struct {
	Code         string                 `json:"code"`
	Name         string                 `json:"name"`
	Email        string                 `json:"email"`
	Phone        string                 `json:"phone"`
	Address      string                 `json:"address"`
	City         string                 `json:"city"`
	State        string                 `json:"state"`
	PostalCode   string                 `json:"postalCode"`
	Country      string                 `json:"country"`
	TaxID        string                 `json:"taxId"`
	PaymentTerms string                 `json:"paymentTerms"`
	Currency     string                 `json:"currency"`
	Notes        string                 `json:"notes"`
	IsActive     bool                   `json:"isActive"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

type UpdateSupplierRequestDTO struct {
	Name         *string                `json:"name,omitempty"`
	Email        *string                `json:"email,omitempty"`
	Phone        *string                `json:"phone,omitempty"`
	Address      *string                `json:"address,omitempty"`
	City         *string                `json:"city,omitempty"`
	State        *string                `json:"state,omitempty"`
	PostalCode   *string                `json:"postalCode,omitempty"`
	Country      *string                `json:"country,omitempty"`
	TaxID        *string                `json:"taxId,omitempty"`
	PaymentTerms *string                `json:"paymentTerms,omitempty"`
	Currency     *string                `json:"currency,omitempty"`
	Notes        *string                `json:"notes,omitempty"`
	IsActive     *bool                  `json:"isActive,omitempty"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

// Supplier handlers

// Create handles POST /api/v1/suppliers
func (h *SupplierHandler) Create(w http.ResponseWriter, r *http.Request) {
	clientID, err := utils.GetClientIDFromContext(r.Context())
	if err != nil {
		writeError(w, http.StatusUnauthorized, "NO_CLIENT_CONTEXT", err.Error())
		return
	}

	var dto CreateSupplierRequestDTO
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	// Validate required fields
	if dto.Code == "" || dto.Name == "" {
		writeError(w, http.StatusBadRequest, "MISSING_FIELDS", "Code and name are required")
		return
	}

	req := &services.CreateSupplierRequest{
		Code:         dto.Code,
		Name:         dto.Name,
		Email:        dto.Email,
		Phone:        dto.Phone,
		Address:      dto.Address,
		City:         dto.City,
		State:        dto.State,
		PostalCode:   dto.PostalCode,
		Country:      dto.Country,
		TaxID:        dto.TaxID,
		PaymentTerms: dto.PaymentTerms,
		Currency:     dto.Currency,
		Notes:        dto.Notes,
		IsActive:     dto.IsActive,
		Metadata:     dto.Metadata,
	}

	supplier, err := h.service.CreateSupplier(r.Context(), clientID, req)
	if err != nil {
		writeError(w, http.StatusBadRequest, "CREATE_FAILED", err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, APIResponse{
		Success: true,
		Data:    supplier,
	})
}

// Get handles GET /api/v1/suppliers/:id
func (h *SupplierHandler) Get(w http.ResponseWriter, r *http.Request) {
	clientID, err := utils.GetClientIDFromContext(r.Context())
	if err != nil {
		writeError(w, http.StatusUnauthorized, "NO_CLIENT_CONTEXT", err.Error())
		return
	}

	vars := mux.Vars(r)
	supplierID, err := uuid.Parse(vars["id"])
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_SUPPLIER_ID", "Invalid supplier ID format")
		return
	}

	supplier, err := h.service.GetSupplier(r.Context(), clientID, supplierID)
	if err != nil {
		writeError(w, http.StatusNotFound, "SUPPLIER_NOT_FOUND", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, APIResponse{
		Success: true,
		Data:    supplier,
	})
}

// Update handles PUT /api/v1/suppliers/:id
func (h *SupplierHandler) Update(w http.ResponseWriter, r *http.Request) {
	clientID, err := utils.GetClientIDFromContext(r.Context())
	if err != nil {
		writeError(w, http.StatusUnauthorized, "NO_CLIENT_CONTEXT", err.Error())
		return
	}

	vars := mux.Vars(r)
	supplierID, err := uuid.Parse(vars["id"])
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_SUPPLIER_ID", "Invalid supplier ID format")
		return
	}

	var dto UpdateSupplierRequestDTO
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	req := &services.UpdateSupplierRequest{
		Name:         dto.Name,
		Email:        dto.Email,
		Phone:        dto.Phone,
		Address:      dto.Address,
		City:         dto.City,
		State:        dto.State,
		PostalCode:   dto.PostalCode,
		Country:      dto.Country,
		TaxID:        dto.TaxID,
		PaymentTerms: dto.PaymentTerms,
		Currency:     dto.Currency,
		Notes:        dto.Notes,
		IsActive:     dto.IsActive,
		Metadata:     dto.Metadata,
	}

	if err := h.service.UpdateSupplier(r.Context(), clientID, supplierID, req); err != nil {
		writeError(w, http.StatusBadRequest, "UPDATE_FAILED", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, APIResponse{
		Success: true,
		Data:    map[string]string{"message": "Supplier updated successfully"},
	})
}

// Delete handles DELETE /api/v1/suppliers/:id
func (h *SupplierHandler) Delete(w http.ResponseWriter, r *http.Request) {
	clientID, err := utils.GetClientIDFromContext(r.Context())
	if err != nil {
		writeError(w, http.StatusUnauthorized, "NO_CLIENT_CONTEXT", err.Error())
		return
	}

	vars := mux.Vars(r)
	supplierID, err := uuid.Parse(vars["id"])
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_SUPPLIER_ID", "Invalid supplier ID format")
		return
	}

	if err := h.service.DeleteSupplier(r.Context(), clientID, supplierID); err != nil {
		writeError(w, http.StatusBadRequest, "DELETE_FAILED", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, APIResponse{
		Success: true,
		Data:    map[string]string{"message": "Supplier deleted successfully"},
	})
}

// List handles GET /api/v1/suppliers
func (h *SupplierHandler) List(w http.ResponseWriter, r *http.Request) {
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

	// Check if only active suppliers are requested
	activeOnly := r.URL.Query().Get("active") == "true"

	var suppliers []*models.Supplier
	var total int

	if activeOnly {
		suppliers, total, err = h.service.ListActiveSuppliers(r.Context(), clientID, page, pageSize)
	} else {
		suppliers, total, err = h.service.ListSuppliers(r.Context(), clientID, page, pageSize)
	}

	if err != nil {
		writeError(w, http.StatusInternalServerError, "LIST_FAILED", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, APIResponse{
		Success: true,
		Data: map[string]interface{}{
			"suppliers":  suppliers,
			"pagination": map[string]int{"page": page, "pageSize": pageSize, "total": total},
		},
	})
}
