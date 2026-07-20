package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/yourorg/shoppilot/internal/models"
	"github.com/yourorg/shoppilot/internal/services"
	"github.com/yourorg/shoppilot/internal/utils"
)

// SalesOrderHandler handles sales order-related HTTP requests
type SalesOrderHandler struct {
	service services.SalesOrderService
}

// NewSalesOrderHandler creates a new sales order handler
func NewSalesOrderHandler(service services.SalesOrderService) *SalesOrderHandler {
	return &SalesOrderHandler{
		service: service,
	}
}

// Request DTOs

type CreateSalesOrderRequestDTO struct {
	CustomerID  string                              `json:"customerId"`
	ShopID      string                              `json:"shopId"`
	OrderNumber string                              `json:"orderNumber,omitempty"`
	OrderDate   time.Time                           `json:"orderDate"`
	Subtotal    float64                             `json:"subtotal"`
	TaxAmount   float64                             `json:"taxAmount"`
	TotalAmount float64                             `json:"totalAmount"`
	Notes       string                              `json:"notes,omitempty"`
	Metadata    map[string]interface{}              `json:"metadata,omitempty"`
	Items       []CreateSalesOrderItemRequestDTO    `json:"items,omitempty"`
}

type CreateSalesOrderItemRequestDTO struct {
	VariantID       string  `json:"variantId"`
	QuantityOrdered int     `json:"quantityOrdered"`
	UnitPrice       float64 `json:"unitPrice"`
	TotalPrice      float64 `json:"totalPrice"`
	Notes           string  `json:"notes,omitempty"`
}

type UpdateSalesOrderRequestDTO struct {
	Notes    *string                `json:"notes,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

type AddSalesOrderItemRequestDTO struct {
	VariantID       string  `json:"variantId"`
	QuantityOrdered int     `json:"quantityOrdered"`
	UnitPrice       float64 `json:"unitPrice"`
	Notes           string  `json:"notes,omitempty"`
}

type FulfillItemRequestDTO struct {
	ItemID            string `json:"itemId"`
	QuantityFulfilled int    `json:"quantityFulfilled"`
}

// Sales Order handlers

// Create handles POST /api/v1/clients/:clientId/sales-orders
func (h *SalesOrderHandler) Create(w http.ResponseWriter, r *http.Request) {
	clientID, err := utils.GetClientIDFromContext(r.Context())
	if err != nil {
		writeError(w, http.StatusUnauthorized, "NO_CLIENT_CONTEXT", err.Error())
		return
	}

	var dto CreateSalesOrderRequestDTO
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	// Validate required fields
	if dto.CustomerID == "" || dto.ShopID == "" {
		writeError(w, http.StatusBadRequest, "MISSING_FIELDS", "Customer ID and shop ID are required")
		return
	}

	customerID, err := uuid.Parse(dto.CustomerID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_CUSTOMER_ID", "Invalid customer ID format")
		return
	}

	shopID, err := uuid.Parse(dto.ShopID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_SHOP_ID", "Invalid shop ID format")
		return
	}

	// Convert items
	items := make([]services.CreateSalesOrderItemRequest, len(dto.Items))
	for i, itemDTO := range dto.Items {
		variantID, err := uuid.Parse(itemDTO.VariantID)
		if err != nil {
			writeError(w, http.StatusBadRequest, "INVALID_VARIANT_ID", "Invalid variant ID format in items")
			return
		}

		items[i] = services.CreateSalesOrderItemRequest{
			VariantID:       variantID,
			QuantityOrdered: itemDTO.QuantityOrdered,
			UnitPrice:       itemDTO.UnitPrice,
			TotalPrice:      itemDTO.TotalPrice,
			Notes:           itemDTO.Notes,
		}
	}

	req := &services.CreateSalesOrderRequest{
		CustomerID:  customerID,
		ShopID:      shopID,
		OrderNumber: dto.OrderNumber,
		OrderDate:   dto.OrderDate,
		Subtotal:    dto.Subtotal,
		TaxAmount:   dto.TaxAmount,
		TotalAmount: dto.TotalAmount,
		Notes:       dto.Notes,
		Metadata:    dto.Metadata,
		Items:       items,
	}

	so, err := h.service.CreateSalesOrder(r.Context(), clientID, req)
	if err != nil {
		writeError(w, http.StatusBadRequest, "CREATE_FAILED", err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, APIResponse{
		Success: true,
		Data:    so,
	})
}

// Get handles GET /api/v1/clients/:clientId/sales-orders/:id
func (h *SalesOrderHandler) Get(w http.ResponseWriter, r *http.Request) {
	clientID, err := utils.GetClientIDFromContext(r.Context())
	if err != nil {
		writeError(w, http.StatusUnauthorized, "NO_CLIENT_CONTEXT", err.Error())
		return
	}

	vars := mux.Vars(r)
	soID, err := uuid.Parse(vars["id"])
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_SO_ID", "Invalid sales order ID format")
		return
	}

	so, err := h.service.GetSalesOrder(r.Context(), clientID, soID)
	if err != nil {
		writeError(w, http.StatusNotFound, "SO_NOT_FOUND", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, APIResponse{
		Success: true,
		Data:    so,
	})
}

// Update handles PUT /api/v1/clients/:clientId/sales-orders/:id
func (h *SalesOrderHandler) Update(w http.ResponseWriter, r *http.Request) {
	clientID, err := utils.GetClientIDFromContext(r.Context())
	if err != nil {
		writeError(w, http.StatusUnauthorized, "NO_CLIENT_CONTEXT", err.Error())
		return
	}

	vars := mux.Vars(r)
	soID, err := uuid.Parse(vars["id"])
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_SO_ID", "Invalid sales order ID format")
		return
	}

	var dto UpdateSalesOrderRequestDTO
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	req := &services.UpdateSalesOrderRequest{
		Notes:    dto.Notes,
		Metadata: dto.Metadata,
	}

	if err := h.service.UpdateSalesOrder(r.Context(), clientID, soID, req); err != nil {
		writeError(w, http.StatusBadRequest, "UPDATE_FAILED", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, APIResponse{
		Success: true,
		Data:    map[string]string{"message": "Sales order updated successfully"},
	})
}

// Delete handles DELETE /api/v1/clients/:clientId/sales-orders/:id
func (h *SalesOrderHandler) Delete(w http.ResponseWriter, r *http.Request) {
	clientID, err := utils.GetClientIDFromContext(r.Context())
	if err != nil {
		writeError(w, http.StatusUnauthorized, "NO_CLIENT_CONTEXT", err.Error())
		return
	}

	vars := mux.Vars(r)
	soID, err := uuid.Parse(vars["id"])
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_SO_ID", "Invalid sales order ID format")
		return
	}

	if err := h.service.DeleteSalesOrder(r.Context(), clientID, soID); err != nil {
		writeError(w, http.StatusBadRequest, "DELETE_FAILED", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, APIResponse{
		Success: true,
		Data:    map[string]string{"message": "Sales order deleted successfully"},
	})
}

// List handles GET /api/v1/clients/:clientId/sales-orders
func (h *SalesOrderHandler) List(w http.ResponseWriter, r *http.Request) {
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

	// Parse filters
	filters := &services.SalesOrderFilters{}

	if customerIDStr := r.URL.Query().Get("customerId"); customerIDStr != "" {
		customerID, err := uuid.Parse(customerIDStr)
		if err != nil {
			writeError(w, http.StatusBadRequest, "INVALID_CUSTOMER_ID", "Invalid customer ID format")
			return
		}
		filters.CustomerID = &customerID
	}

	if shopIDStr := r.URL.Query().Get("shopId"); shopIDStr != "" {
		shopID, err := uuid.Parse(shopIDStr)
		if err != nil {
			writeError(w, http.StatusBadRequest, "INVALID_SHOP_ID", "Invalid shop ID format")
			return
		}
		filters.ShopID = &shopID
	}

	if statusStr := r.URL.Query().Get("status"); statusStr != "" {
		status := models.SalesOrderStatus(statusStr)
		filters.Status = &status
	}

	salesOrders, total, err := h.service.ListSalesOrders(r.Context(), clientID, filters, page, pageSize)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "LIST_FAILED", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, APIResponse{
		Success: true,
		Data:    salesOrders,
		Meta: &Meta{
			Page:       page,
			PageSize:   pageSize,
			Total:      total,
			TotalPages: (total + pageSize - 1) / pageSize,
		},
	})
}

// Confirm handles POST /api/v1/clients/:clientId/sales-orders/:id/confirm
func (h *SalesOrderHandler) Confirm(w http.ResponseWriter, r *http.Request) {
	clientID, err := utils.GetClientIDFromContext(r.Context())
	if err != nil {
		writeError(w, http.StatusUnauthorized, "NO_CLIENT_CONTEXT", err.Error())
		return
	}

	vars := mux.Vars(r)
	soID, err := uuid.Parse(vars["id"])
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_SO_ID", "Invalid sales order ID format")
		return
	}

	if err := h.service.ConfirmSalesOrder(r.Context(), clientID, soID); err != nil {
		writeError(w, http.StatusBadRequest, "CONFIRM_FAILED", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, APIResponse{
		Success: true,
		Data:    map[string]string{"message": "Sales order confirmed successfully"},
	})
}

// Cancel handles POST /api/v1/clients/:clientId/sales-orders/:id/cancel
func (h *SalesOrderHandler) Cancel(w http.ResponseWriter, r *http.Request) {
	clientID, err := utils.GetClientIDFromContext(r.Context())
	if err != nil {
		writeError(w, http.StatusUnauthorized, "NO_CLIENT_CONTEXT", err.Error())
		return
	}

	vars := mux.Vars(r)
	soID, err := uuid.Parse(vars["id"])
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_SO_ID", "Invalid sales order ID format")
		return
	}

	if err := h.service.CancelSalesOrder(r.Context(), clientID, soID); err != nil {
		writeError(w, http.StatusBadRequest, "CANCEL_FAILED", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, APIResponse{
		Success: true,
		Data:    map[string]string{"message": "Sales order cancelled successfully"},
	})
}

// AddItem handles POST /api/v1/clients/:clientId/sales-orders/:id/items
func (h *SalesOrderHandler) AddItem(w http.ResponseWriter, r *http.Request) {
	clientID, err := utils.GetClientIDFromContext(r.Context())
	if err != nil {
		writeError(w, http.StatusUnauthorized, "NO_CLIENT_CONTEXT", err.Error())
		return
	}

	vars := mux.Vars(r)
	soID, err := uuid.Parse(vars["id"])
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_SO_ID", "Invalid sales order ID format")
		return
	}

	var dto AddSalesOrderItemRequestDTO
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	// Validate required fields
	if dto.VariantID == "" {
		writeError(w, http.StatusBadRequest, "MISSING_FIELDS", "Variant ID is required")
		return
	}

	variantID, err := uuid.Parse(dto.VariantID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_VARIANT_ID", "Invalid variant ID format")
		return
	}

	req := &services.AddSalesOrderItemRequest{
		VariantID:       variantID,
		QuantityOrdered: dto.QuantityOrdered,
		UnitPrice:       dto.UnitPrice,
		Notes:           dto.Notes,
	}

	item, err := h.service.AddItem(r.Context(), clientID, soID, req)
	if err != nil {
		writeError(w, http.StatusBadRequest, "ADD_ITEM_FAILED", err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, APIResponse{
		Success: true,
		Data:    item,
	})
}

// RemoveItem handles DELETE /api/v1/clients/:clientId/sales-orders/:id/items/:itemId
func (h *SalesOrderHandler) RemoveItem(w http.ResponseWriter, r *http.Request) {
	clientID, err := utils.GetClientIDFromContext(r.Context())
	if err != nil {
		writeError(w, http.StatusUnauthorized, "NO_CLIENT_CONTEXT", err.Error())
		return
	}

	vars := mux.Vars(r)
	soID, err := uuid.Parse(vars["id"])
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_SO_ID", "Invalid sales order ID format")
		return
	}

	itemID, err := uuid.Parse(vars["itemId"])
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ITEM_ID", "Invalid item ID format")
		return
	}

	if err := h.service.RemoveItem(r.Context(), clientID, soID, itemID); err != nil {
		writeError(w, http.StatusBadRequest, "REMOVE_ITEM_FAILED", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, APIResponse{
		Success: true,
		Data:    map[string]string{"message": "Item removed successfully"},
	})
}

// ListItems handles GET /api/v1/clients/:clientId/sales-orders/:id/items
func (h *SalesOrderHandler) ListItems(w http.ResponseWriter, r *http.Request) {
	clientID, err := utils.GetClientIDFromContext(r.Context())
	if err != nil {
		writeError(w, http.StatusUnauthorized, "NO_CLIENT_CONTEXT", err.Error())
		return
	}

	vars := mux.Vars(r)
	soID, err := uuid.Parse(vars["id"])
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_SO_ID", "Invalid sales order ID format")
		return
	}

	items, err := h.service.ListItems(r.Context(), clientID, soID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "LIST_ITEMS_FAILED", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, APIResponse{
		Success: true,
		Data:    items,
	})
}

// Fulfill handles POST /api/v1/clients/:clientId/sales-orders/:id/fulfill
func (h *SalesOrderHandler) Fulfill(w http.ResponseWriter, r *http.Request) {
	clientID, err := utils.GetClientIDFromContext(r.Context())
	if err != nil {
		writeError(w, http.StatusUnauthorized, "NO_CLIENT_CONTEXT", err.Error())
		return
	}

	vars := mux.Vars(r)
	soID, err := uuid.Parse(vars["id"])
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_SO_ID", "Invalid sales order ID format")
		return
	}

	var dtoItems []FulfillItemRequestDTO
	if err := json.NewDecoder(r.Body).Decode(&dtoItems); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	// Convert DTOs to service request format
	items := make([]services.FulfillItemRequest, len(dtoItems))
	for i, dto := range dtoItems {
		itemID, err := uuid.Parse(dto.ItemID)
		if err != nil {
			writeError(w, http.StatusBadRequest, "INVALID_ITEM_ID", "Invalid item ID format")
			return
		}

		items[i] = services.FulfillItemRequest{
			ItemID:            itemID,
			QuantityFulfilled: dto.QuantityFulfilled,
		}
	}

	if err := h.service.FulfillItems(r.Context(), clientID, soID, items); err != nil {
		writeError(w, http.StatusBadRequest, "FULFILL_FAILED", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, APIResponse{
		Success: true,
		Data:    map[string]string{"message": "Items fulfilled successfully"},
	})
}
