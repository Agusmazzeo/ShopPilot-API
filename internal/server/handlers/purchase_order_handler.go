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

// PurchaseOrderHandler handles purchase order-related HTTP requests
type PurchaseOrderHandler struct {
	service services.PurchaseOrderService
}

// NewPurchaseOrderHandler creates a new purchase order handler
func NewPurchaseOrderHandler(service services.PurchaseOrderService) *PurchaseOrderHandler {
	return &PurchaseOrderHandler{
		service: service,
	}
}

// Request DTOs

type CreatePurchaseOrderRequestDTO struct {
	SupplierID   string                                `json:"supplierId"`
	ShopID       string                                `json:"shopId"`
	PONumber     string                                `json:"poNumber,omitempty"`
	OrderDate    time.Time                             `json:"orderDate"`
	ExpectedDate *time.Time                            `json:"expectedDate,omitempty"`
	Notes        string                                `json:"notes,omitempty"`
	TotalAmount  float64                               `json:"totalAmount"`
	Metadata     map[string]interface{}                `json:"metadata,omitempty"`
	Items        []CreatePurchaseOrderItemRequestDTO   `json:"items,omitempty"`
}

type CreatePurchaseOrderItemRequestDTO struct {
	VariantID       string  `json:"variantId"`
	QuantityOrdered int     `json:"quantityOrdered"`
	UnitPrice       float64 `json:"unitPrice"`
	TotalPrice      float64 `json:"totalPrice"`
	Notes           string  `json:"notes,omitempty"`
}

type UpdatePurchaseOrderRequestDTO struct {
	ExpectedDate *time.Time             `json:"expectedDate,omitempty"`
	Notes        *string                `json:"notes,omitempty"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

type AddPurchaseOrderItemRequestDTO struct {
	VariantID       string  `json:"variantId"`
	QuantityOrdered int     `json:"quantityOrdered"`
	UnitPrice       float64 `json:"unitPrice"`
	Notes           string  `json:"notes,omitempty"`
}

type ReceiveItemRequestDTO struct {
	ItemID           string `json:"itemId"`
	QuantityReceived int    `json:"quantityReceived"`
}

// Purchase Order handlers

// Create handles POST /api/v1/clients/:clientId/purchase-orders
func (h *PurchaseOrderHandler) Create(w http.ResponseWriter, r *http.Request) {
	clientID, err := utils.GetClientIDFromContext(r.Context())
	if err != nil {
		writeError(w, http.StatusUnauthorized, "NO_CLIENT_CONTEXT", err.Error())
		return
	}

	var dto CreatePurchaseOrderRequestDTO
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	// Validate required fields
	if dto.SupplierID == "" || dto.ShopID == "" {
		writeError(w, http.StatusBadRequest, "MISSING_FIELDS", "Supplier ID and shop ID are required")
		return
	}

	supplierID, err := uuid.Parse(dto.SupplierID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_SUPPLIER_ID", "Invalid supplier ID format")
		return
	}

	shopID, err := uuid.Parse(dto.ShopID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_SHOP_ID", "Invalid shop ID format")
		return
	}

	// Convert items
	items := make([]services.CreatePurchaseOrderItemRequest, len(dto.Items))
	for i, itemDTO := range dto.Items {
		variantID, err := uuid.Parse(itemDTO.VariantID)
		if err != nil {
			writeError(w, http.StatusBadRequest, "INVALID_VARIANT_ID", "Invalid variant ID format in items")
			return
		}

		items[i] = services.CreatePurchaseOrderItemRequest{
			VariantID:       variantID,
			QuantityOrdered: itemDTO.QuantityOrdered,
			UnitPrice:       itemDTO.UnitPrice,
			TotalPrice:      itemDTO.TotalPrice,
			Notes:           itemDTO.Notes,
		}
	}

	req := &services.CreatePurchaseOrderRequest{
		SupplierID:   supplierID,
		ShopID:       shopID,
		PONumber:     dto.PONumber,
		OrderDate:    dto.OrderDate,
		ExpectedDate: dto.ExpectedDate,
		Notes:        dto.Notes,
		TotalAmount:  dto.TotalAmount,
		Metadata:     dto.Metadata,
		Items:        items,
	}

	po, err := h.service.CreatePurchaseOrder(r.Context(), clientID, req)
	if err != nil {
		writeError(w, http.StatusBadRequest, "CREATE_FAILED", err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, APIResponse{
		Success: true,
		Data:    po,
	})
}

// Get handles GET /api/v1/clients/:clientId/purchase-orders/:id
func (h *PurchaseOrderHandler) Get(w http.ResponseWriter, r *http.Request) {
	clientID, err := utils.GetClientIDFromContext(r.Context())
	if err != nil {
		writeError(w, http.StatusUnauthorized, "NO_CLIENT_CONTEXT", err.Error())
		return
	}

	vars := mux.Vars(r)
	poID, err := uuid.Parse(vars["id"])
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_PO_ID", "Invalid purchase order ID format")
		return
	}

	po, err := h.service.GetPurchaseOrder(r.Context(), clientID, poID)
	if err != nil {
		writeError(w, http.StatusNotFound, "PO_NOT_FOUND", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, APIResponse{
		Success: true,
		Data:    po,
	})
}

// Update handles PUT /api/v1/clients/:clientId/purchase-orders/:id
func (h *PurchaseOrderHandler) Update(w http.ResponseWriter, r *http.Request) {
	clientID, err := utils.GetClientIDFromContext(r.Context())
	if err != nil {
		writeError(w, http.StatusUnauthorized, "NO_CLIENT_CONTEXT", err.Error())
		return
	}

	vars := mux.Vars(r)
	poID, err := uuid.Parse(vars["id"])
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_PO_ID", "Invalid purchase order ID format")
		return
	}

	var dto UpdatePurchaseOrderRequestDTO
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	req := &services.UpdatePurchaseOrderRequest{
		ExpectedDate: dto.ExpectedDate,
		Notes:        dto.Notes,
		Metadata:     dto.Metadata,
	}

	if err := h.service.UpdatePurchaseOrder(r.Context(), clientID, poID, req); err != nil {
		writeError(w, http.StatusBadRequest, "UPDATE_FAILED", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, APIResponse{
		Success: true,
		Data:    map[string]string{"message": "Purchase order updated successfully"},
	})
}

// Delete handles DELETE /api/v1/clients/:clientId/purchase-orders/:id
func (h *PurchaseOrderHandler) Delete(w http.ResponseWriter, r *http.Request) {
	clientID, err := utils.GetClientIDFromContext(r.Context())
	if err != nil {
		writeError(w, http.StatusUnauthorized, "NO_CLIENT_CONTEXT", err.Error())
		return
	}

	vars := mux.Vars(r)
	poID, err := uuid.Parse(vars["id"])
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_PO_ID", "Invalid purchase order ID format")
		return
	}

	if err := h.service.DeletePurchaseOrder(r.Context(), clientID, poID); err != nil {
		writeError(w, http.StatusBadRequest, "DELETE_FAILED", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, APIResponse{
		Success: true,
		Data:    map[string]string{"message": "Purchase order deleted successfully"},
	})
}

// List handles GET /api/v1/clients/:clientId/purchase-orders
func (h *PurchaseOrderHandler) List(w http.ResponseWriter, r *http.Request) {
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
	filters := &services.PurchaseOrderFilters{}

	if supplierIDStr := r.URL.Query().Get("supplierId"); supplierIDStr != "" {
		supplierID, err := uuid.Parse(supplierIDStr)
		if err != nil {
			writeError(w, http.StatusBadRequest, "INVALID_SUPPLIER_ID", "Invalid supplier ID format")
			return
		}
		filters.SupplierID = &supplierID
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
		status := models.PurchaseOrderStatus(statusStr)
		filters.Status = &status
	}

	purchaseOrders, total, err := h.service.ListPurchaseOrders(r.Context(), clientID, filters, page, pageSize)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "LIST_FAILED", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, APIResponse{
		Success: true,
		Data:    purchaseOrders,
		Meta: &Meta{
			Page:       page,
			PageSize:   pageSize,
			Total:      total,
			TotalPages: (total + pageSize - 1) / pageSize,
		},
	})
}

// Submit handles POST /api/v1/clients/:clientId/purchase-orders/:id/submit
func (h *PurchaseOrderHandler) Submit(w http.ResponseWriter, r *http.Request) {
	clientID, err := utils.GetClientIDFromContext(r.Context())
	if err != nil {
		writeError(w, http.StatusUnauthorized, "NO_CLIENT_CONTEXT", err.Error())
		return
	}

	vars := mux.Vars(r)
	poID, err := uuid.Parse(vars["id"])
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_PO_ID", "Invalid purchase order ID format")
		return
	}

	if err := h.service.SubmitPurchaseOrder(r.Context(), clientID, poID); err != nil {
		writeError(w, http.StatusBadRequest, "SUBMIT_FAILED", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, APIResponse{
		Success: true,
		Data:    map[string]string{"message": "Purchase order submitted successfully"},
	})
}

// Cancel handles POST /api/v1/clients/:clientId/purchase-orders/:id/cancel
func (h *PurchaseOrderHandler) Cancel(w http.ResponseWriter, r *http.Request) {
	clientID, err := utils.GetClientIDFromContext(r.Context())
	if err != nil {
		writeError(w, http.StatusUnauthorized, "NO_CLIENT_CONTEXT", err.Error())
		return
	}

	vars := mux.Vars(r)
	poID, err := uuid.Parse(vars["id"])
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_PO_ID", "Invalid purchase order ID format")
		return
	}

	if err := h.service.CancelPurchaseOrder(r.Context(), clientID, poID); err != nil {
		writeError(w, http.StatusBadRequest, "CANCEL_FAILED", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, APIResponse{
		Success: true,
		Data:    map[string]string{"message": "Purchase order cancelled successfully"},
	})
}

// AddItem handles POST /api/v1/clients/:clientId/purchase-orders/:id/items
func (h *PurchaseOrderHandler) AddItem(w http.ResponseWriter, r *http.Request) {
	clientID, err := utils.GetClientIDFromContext(r.Context())
	if err != nil {
		writeError(w, http.StatusUnauthorized, "NO_CLIENT_CONTEXT", err.Error())
		return
	}

	vars := mux.Vars(r)
	poID, err := uuid.Parse(vars["id"])
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_PO_ID", "Invalid purchase order ID format")
		return
	}

	var dto AddPurchaseOrderItemRequestDTO
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

	req := &services.AddPurchaseOrderItemRequest{
		VariantID:       variantID,
		QuantityOrdered: dto.QuantityOrdered,
		UnitPrice:       dto.UnitPrice,
		Notes:           dto.Notes,
	}

	item, err := h.service.AddItem(r.Context(), clientID, poID, req)
	if err != nil {
		writeError(w, http.StatusBadRequest, "ADD_ITEM_FAILED", err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, APIResponse{
		Success: true,
		Data:    item,
	})
}

// RemoveItem handles DELETE /api/v1/clients/:clientId/purchase-orders/:id/items/:itemId
func (h *PurchaseOrderHandler) RemoveItem(w http.ResponseWriter, r *http.Request) {
	clientID, err := utils.GetClientIDFromContext(r.Context())
	if err != nil {
		writeError(w, http.StatusUnauthorized, "NO_CLIENT_CONTEXT", err.Error())
		return
	}

	vars := mux.Vars(r)
	poID, err := uuid.Parse(vars["id"])
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_PO_ID", "Invalid purchase order ID format")
		return
	}

	itemID, err := uuid.Parse(vars["itemId"])
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ITEM_ID", "Invalid item ID format")
		return
	}

	if err := h.service.RemoveItem(r.Context(), clientID, poID, itemID); err != nil {
		writeError(w, http.StatusBadRequest, "REMOVE_ITEM_FAILED", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, APIResponse{
		Success: true,
		Data:    map[string]string{"message": "Item removed successfully"},
	})
}

// ListItems handles GET /api/v1/clients/:clientId/purchase-orders/:id/items
func (h *PurchaseOrderHandler) ListItems(w http.ResponseWriter, r *http.Request) {
	clientID, err := utils.GetClientIDFromContext(r.Context())
	if err != nil {
		writeError(w, http.StatusUnauthorized, "NO_CLIENT_CONTEXT", err.Error())
		return
	}

	vars := mux.Vars(r)
	poID, err := uuid.Parse(vars["id"])
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_PO_ID", "Invalid purchase order ID format")
		return
	}

	items, err := h.service.ListItems(r.Context(), clientID, poID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "LIST_ITEMS_FAILED", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, APIResponse{
		Success: true,
		Data:    items,
	})
}

// Receive handles POST /api/v1/clients/:clientId/purchase-orders/:id/receive
func (h *PurchaseOrderHandler) Receive(w http.ResponseWriter, r *http.Request) {
	clientID, err := utils.GetClientIDFromContext(r.Context())
	if err != nil {
		writeError(w, http.StatusUnauthorized, "NO_CLIENT_CONTEXT", err.Error())
		return
	}

	vars := mux.Vars(r)
	poID, err := uuid.Parse(vars["id"])
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_PO_ID", "Invalid purchase order ID format")
		return
	}

	var dtoItems []ReceiveItemRequestDTO
	if err := json.NewDecoder(r.Body).Decode(&dtoItems); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	// Convert DTOs to service request format
	items := make([]services.ReceiveItemRequest, len(dtoItems))
	for i, dto := range dtoItems {
		itemID, err := uuid.Parse(dto.ItemID)
		if err != nil {
			writeError(w, http.StatusBadRequest, "INVALID_ITEM_ID", "Invalid item ID format")
			return
		}

		items[i] = services.ReceiveItemRequest{
			ItemID:           itemID,
			QuantityReceived: dto.QuantityReceived,
		}
	}

	if err := h.service.ReceiveItems(r.Context(), clientID, poID, items); err != nil {
		writeError(w, http.StatusBadRequest, "RECEIVE_FAILED", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, APIResponse{
		Success: true,
		Data:    map[string]string{"message": "Items received successfully"},
	})
}
