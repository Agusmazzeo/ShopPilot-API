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

// ProductHandler handles product-related HTTP requests
type ProductHandler struct {
	service services.ProductService
}

// NewProductHandler creates a new product handler
func NewProductHandler(service services.ProductService) *ProductHandler {
	return &ProductHandler{
		service: service,
	}
}


// Request DTOs
type CreateProductRequestDTO struct {
	ShopID      string                 `json:"shopId"`
	Code        string                 `json:"code"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Metadata    map[string]interface{} `json:"metadata"`
	IsActive    bool                   `json:"isActive"`
}

type UpdateProductRequestDTO struct {
	Name        *string                `json:"name,omitempty"`
	Description *string                `json:"description,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	IsActive    *bool                  `json:"isActive,omitempty"`
}

type CreateVariantRequestDTO struct {
	SKU              string                 `json:"sku"`
	Name             string                 `json:"name"`
	Price            float64                `json:"price"`
	CompareAtPrice   *float64               `json:"compareAtPrice,omitempty"`
	Cost             *float64               `json:"cost,omitempty"`
	Quantity         int                    `json:"quantity"`
	Weight           *float64               `json:"weight,omitempty"`
	WeightUnit       string                 `json:"weightUnit"`
	RequiresShipping bool                   `json:"requiresShipping"`
	IsDefault        bool                   `json:"isDefault"`
	Attributes       map[string]interface{} `json:"attributes,omitempty"`
	IsActive         bool                   `json:"isActive"`
}

type UpdateVariantRequestDTO struct {
	Name             *string                `json:"name,omitempty"`
	Price            *float64               `json:"price,omitempty"`
	CompareAtPrice   *float64               `json:"compareAtPrice,omitempty"`
	Cost             *float64               `json:"cost,omitempty"`
	Quantity         *int                   `json:"quantity,omitempty"`
	Weight           *float64               `json:"weight,omitempty"`
	WeightUnit       *string                `json:"weightUnit,omitempty"`
	RequiresShipping *bool                  `json:"requiresShipping,omitempty"`
	IsDefault        *bool                  `json:"isDefault,omitempty"`
	Attributes       map[string]interface{} `json:"attributes,omitempty"`
	IsActive         *bool                  `json:"isActive,omitempty"`
}

type AdjustInventoryRequestDTO struct {
	Delta int `json:"delta"`
}

type SetInventoryRequestDTO struct {
	Quantity int `json:"quantity"`
}

type RecordMovementRequestDTO struct {
	ShopID        string  `json:"shopId"`
	MovementType  string  `json:"movementType"`
	Quantity      int     `json:"quantity"`
	ReferenceType string  `json:"referenceType,omitempty"`
	ReferenceID   *string `json:"referenceId,omitempty"`
	Notes         string  `json:"notes,omitempty"`
	PerformedBy   *string `json:"performedBy,omitempty"`
}

type SetInventoryAlertRequestDTO struct {
	ReorderPoint      int  `json:"reorderPoint"`
	ReorderQuantity   int  `json:"reorderQuantity"`
	LowStockThreshold int  `json:"lowStockThreshold"`
	IsEnabled         bool `json:"isEnabled"`
}

// Product handlers

// Create handles POST /api/v1/products
func (h *ProductHandler) Create(w http.ResponseWriter, r *http.Request) {
	clientID, err := utils.GetClientIDFromContext(r.Context())
	if err != nil {
		writeError(w, http.StatusUnauthorized, "NO_CLIENT_CONTEXT", err.Error())
		return
	}

	var dto CreateProductRequestDTO
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	// Validate required fields
	if dto.Code == "" || dto.Name == "" || dto.ShopID == "" {
		writeError(w, http.StatusBadRequest, "MISSING_FIELDS", "Code, name, and shopId are required")
		return
	}

	shopID, err := uuid.Parse(dto.ShopID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_SHOP_ID", "Invalid shop ID format")
		return
	}

	req := &services.CreateProductRequest{
		ClientID:    clientID,
		ShopID:      shopID,
		Code:        dto.Code,
		Name:        dto.Name,
		Description: dto.Description,
		Metadata:    dto.Metadata,
		IsActive:    dto.IsActive,
	}

	product, err := h.service.CreateProduct(r.Context(), req)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "CREATE_FAILED", err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, APIResponse{
		Success: true,
		Data:    product,
	})
}

// Get handles GET /api/v1/products/:id
func (h *ProductHandler) Get(w http.ResponseWriter, r *http.Request) {
	clientID, err := utils.GetClientIDFromContext(r.Context())
	if err != nil {
		writeError(w, http.StatusUnauthorized, "NO_CLIENT_CONTEXT", err.Error())
		return
	}

	vars := mux.Vars(r)
	productID, err := uuid.Parse(vars["id"])
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_PRODUCT_ID", "Invalid product ID format")
		return
	}

	product, err := h.service.GetProduct(r.Context(), clientID, productID)
	if err != nil {
		writeError(w, http.StatusNotFound, "PRODUCT_NOT_FOUND", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, APIResponse{
		Success: true,
		Data:    product,
	})
}

// Update handles PUT /api/v1/products/:id
func (h *ProductHandler) Update(w http.ResponseWriter, r *http.Request) {
	clientID, err := utils.GetClientIDFromContext(r.Context())
	if err != nil {
		writeError(w, http.StatusUnauthorized, "NO_CLIENT_CONTEXT", err.Error())
		return
	}

	vars := mux.Vars(r)
	productID, err := uuid.Parse(vars["id"])
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_PRODUCT_ID", "Invalid product ID format")
		return
	}

	var dto UpdateProductRequestDTO
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	req := &services.UpdateProductRequest{
		Name:        dto.Name,
		Description: dto.Description,
		Metadata:    dto.Metadata,
		IsActive:    dto.IsActive,
	}

	if err := h.service.UpdateProduct(r.Context(), clientID, productID, req); err != nil {
		writeError(w, http.StatusInternalServerError, "UPDATE_FAILED", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, APIResponse{
		Success: true,
		Data:    map[string]string{"message": "Product updated successfully"},
	})
}

// Delete handles DELETE /api/v1/products/:id
func (h *ProductHandler) Delete(w http.ResponseWriter, r *http.Request) {
	clientID, err := utils.GetClientIDFromContext(r.Context())
	if err != nil {
		writeError(w, http.StatusUnauthorized, "NO_CLIENT_CONTEXT", err.Error())
		return
	}

	vars := mux.Vars(r)
	productID, err := uuid.Parse(vars["id"])
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_PRODUCT_ID", "Invalid product ID format")
		return
	}

	if err := h.service.DeleteProduct(r.Context(), clientID, productID); err != nil {
		writeError(w, http.StatusInternalServerError, "DELETE_FAILED", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, APIResponse{
		Success: true,
		Data:    map[string]string{"message": "Product deleted successfully"},
	})
}

// List handles GET /api/v1/products
func (h *ProductHandler) List(w http.ResponseWriter, r *http.Request) {
	clientID, err := utils.GetClientIDFromContext(r.Context())
	if err != nil {
		writeError(w, http.StatusUnauthorized, "NO_CLIENT_CONTEXT", err.Error())
		return
	}

	// Parse query parameters
	page := h.getIntQueryParam(r, "page", 1)
	pageSize := h.getIntQueryParam(r, "page_size", 20)

	// Optional shop filter
	var shopID *uuid.UUID
	if shopIDStr := r.URL.Query().Get("shopId"); shopIDStr != "" {
		sid, err := uuid.Parse(shopIDStr)
		if err != nil {
			writeError(w, http.StatusBadRequest, "INVALID_SHOP_ID", "Invalid shop ID format")
			return
		}
		shopID = &sid
	}

	products, total, err := h.service.ListProducts(r.Context(), clientID, shopID, page, pageSize)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "LIST_FAILED", err.Error())
		return
	}

	meta := &Meta{
		Page:       page,
		PageSize:   pageSize,
		Total:      total,
		TotalPages: (total + pageSize - 1) / pageSize,
	}

	writeJSON(w, http.StatusOK, APIResponse{
		Success: true,
		Data:    products,
		Meta:    meta,
	})
}

// Search handles GET /api/v1/products/search
func (h *ProductHandler) Search(w http.ResponseWriter, r *http.Request) {
	clientID, err := utils.GetClientIDFromContext(r.Context())
	if err != nil {
		writeError(w, http.StatusUnauthorized, "NO_CLIENT_CONTEXT", err.Error())
		return
	}

	query := r.URL.Query().Get("q")
	if query == "" {
		writeError(w, http.StatusBadRequest, "MISSING_QUERY", "Search query parameter 'q' is required")
		return
	}

	page := h.getIntQueryParam(r, "page", 1)
	pageSize := h.getIntQueryParam(r, "page_size", 20)

	products, total, err := h.service.SearchProducts(r.Context(), clientID, query, page, pageSize)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "SEARCH_FAILED", err.Error())
		return
	}

	meta := &Meta{
		Page:       page,
		PageSize:   pageSize,
		Total:      total,
		TotalPages: (total + pageSize - 1) / pageSize,
	}

	writeJSON(w, http.StatusOK, APIResponse{
		Success: true,
		Data:    products,
		Meta:    meta,
	})
}

// Variant handlers

// CreateVariant handles POST /api/v1/products/:productId/variants
func (h *ProductHandler) CreateVariant(w http.ResponseWriter, r *http.Request) {
	clientID, err := utils.GetClientIDFromContext(r.Context())
	if err != nil {
		writeError(w, http.StatusUnauthorized, "NO_CLIENT_CONTEXT", err.Error())
		return
	}

	vars := mux.Vars(r)
	productID, err := uuid.Parse(vars["productId"])
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_PRODUCT_ID", "Invalid product ID format")
		return
	}

	var dto CreateVariantRequestDTO
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	// Validate required fields
	if dto.SKU == "" || dto.Name == "" {
		writeError(w, http.StatusBadRequest, "MISSING_FIELDS", "SKU and name are required")
		return
	}

	req := &services.CreateVariantRequest{
		SKU:              dto.SKU,
		Name:             dto.Name,
		Price:            dto.Price,
		CompareAtPrice:   dto.CompareAtPrice,
		Cost:             dto.Cost,
		Quantity:         dto.Quantity,
		Weight:           dto.Weight,
		WeightUnit:       dto.WeightUnit,
		RequiresShipping: dto.RequiresShipping,
		IsDefault:        dto.IsDefault,
		Attributes:       dto.Attributes,
		IsActive:         dto.IsActive,
	}

	variant, err := h.service.CreateVariant(r.Context(), clientID, productID, req)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "CREATE_FAILED", err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, APIResponse{
		Success: true,
		Data:    variant,
	})
}

// GetVariant handles GET /api/v1/variants/:id
func (h *ProductHandler) GetVariant(w http.ResponseWriter, r *http.Request) {
	clientID, err := utils.GetClientIDFromContext(r.Context())
	if err != nil {
		writeError(w, http.StatusUnauthorized, "NO_CLIENT_CONTEXT", err.Error())
		return
	}

	vars := mux.Vars(r)
	variantID, err := uuid.Parse(vars["id"])
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_VARIANT_ID", "Invalid variant ID format")
		return
	}

	variant, err := h.service.GetVariant(r.Context(), clientID, variantID)
	if err != nil {
		writeError(w, http.StatusNotFound, "VARIANT_NOT_FOUND", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, APIResponse{
		Success: true,
		Data:    variant,
	})
}

// UpdateVariant handles PUT /api/v1/variants/:id
func (h *ProductHandler) UpdateVariant(w http.ResponseWriter, r *http.Request) {
	clientID, err := utils.GetClientIDFromContext(r.Context())
	if err != nil {
		writeError(w, http.StatusUnauthorized, "NO_CLIENT_CONTEXT", err.Error())
		return
	}

	vars := mux.Vars(r)
	variantID, err := uuid.Parse(vars["id"])
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_VARIANT_ID", "Invalid variant ID format")
		return
	}

	var dto UpdateVariantRequestDTO
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	req := &services.UpdateVariantRequest{
		Name:             dto.Name,
		Price:            dto.Price,
		CompareAtPrice:   dto.CompareAtPrice,
		Cost:             dto.Cost,
		Quantity:         dto.Quantity,
		Weight:           dto.Weight,
		WeightUnit:       dto.WeightUnit,
		RequiresShipping: dto.RequiresShipping,
		IsDefault:        dto.IsDefault,
		Attributes:       dto.Attributes,
		IsActive:         dto.IsActive,
	}

	if err := h.service.UpdateVariant(r.Context(), clientID, variantID, req); err != nil {
		writeError(w, http.StatusInternalServerError, "UPDATE_FAILED", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, APIResponse{
		Success: true,
		Data:    map[string]string{"message": "Variant updated successfully"},
	})
}

// DeleteVariant handles DELETE /api/v1/variants/:id
func (h *ProductHandler) DeleteVariant(w http.ResponseWriter, r *http.Request) {
	clientID, err := utils.GetClientIDFromContext(r.Context())
	if err != nil {
		writeError(w, http.StatusUnauthorized, "NO_CLIENT_CONTEXT", err.Error())
		return
	}

	vars := mux.Vars(r)
	variantID, err := uuid.Parse(vars["id"])
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_VARIANT_ID", "Invalid variant ID format")
		return
	}

	if err := h.service.DeleteVariant(r.Context(), clientID, variantID); err != nil {
		writeError(w, http.StatusInternalServerError, "DELETE_FAILED", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, APIResponse{
		Success: true,
		Data:    map[string]string{"message": "Variant deleted successfully"},
	})
}

// ListVariants handles GET /api/v1/products/:productId/variants
func (h *ProductHandler) ListVariants(w http.ResponseWriter, r *http.Request) {
	clientID, err := utils.GetClientIDFromContext(r.Context())
	if err != nil {
		writeError(w, http.StatusUnauthorized, "NO_CLIENT_CONTEXT", err.Error())
		return
	}

	vars := mux.Vars(r)
	productID, err := uuid.Parse(vars["productId"])
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_PRODUCT_ID", "Invalid product ID format")
		return
	}

	variants, err := h.service.ListVariants(r.Context(), clientID, productID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "LIST_FAILED", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, APIResponse{
		Success: true,
		Data:    variants,
	})
}

// Inventory handlers

// AdjustInventory handles POST /api/v1/variants/:id/inventory/adjust
func (h *ProductHandler) AdjustInventory(w http.ResponseWriter, r *http.Request) {
	clientID, err := utils.GetClientIDFromContext(r.Context())
	if err != nil {
		writeError(w, http.StatusUnauthorized, "NO_CLIENT_CONTEXT", err.Error())
		return
	}

	vars := mux.Vars(r)
	variantID, err := uuid.Parse(vars["id"])
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_VARIANT_ID", "Invalid variant ID format")
		return
	}

	var dto AdjustInventoryRequestDTO
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	if err := h.service.AdjustInventory(r.Context(), clientID, variantID, dto.Delta); err != nil {
		writeError(w, http.StatusInternalServerError, "ADJUST_FAILED", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, APIResponse{
		Success: true,
		Data:    map[string]string{"message": "Inventory adjusted successfully"},
	})
}

// SetInventory handles PUT /api/v1/variants/:id/inventory
func (h *ProductHandler) SetInventory(w http.ResponseWriter, r *http.Request) {
	clientID, err := utils.GetClientIDFromContext(r.Context())
	if err != nil {
		writeError(w, http.StatusUnauthorized, "NO_CLIENT_CONTEXT", err.Error())
		return
	}

	vars := mux.Vars(r)
	variantID, err := uuid.Parse(vars["id"])
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_VARIANT_ID", "Invalid variant ID format")
		return
	}

	var dto SetInventoryRequestDTO
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	if err := h.service.SetInventory(r.Context(), clientID, variantID, dto.Quantity); err != nil {
		writeError(w, http.StatusInternalServerError, "SET_FAILED", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, APIResponse{
		Success: true,
		Data:    map[string]string{"message": "Inventory set successfully"},
	})
}

// CheckStock handles GET /api/v1/variants/:id/inventory
func (h *ProductHandler) CheckStock(w http.ResponseWriter, r *http.Request) {
	clientID, err := utils.GetClientIDFromContext(r.Context())
	if err != nil {
		writeError(w, http.StatusUnauthorized, "NO_CLIENT_CONTEXT", err.Error())
		return
	}

	vars := mux.Vars(r)
	variantID, err := uuid.Parse(vars["id"])
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_VARIANT_ID", "Invalid variant ID format")
		return
	}

	quantity, err := h.service.CheckStock(r.Context(), clientID, variantID)
	if err != nil {
		writeError(w, http.StatusNotFound, "STOCK_CHECK_FAILED", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, APIResponse{
		Success: true,
		Data:    map[string]int{"quantity": quantity},
	})
}

// Inventory movement handlers

// GetMovements handles GET /api/v1/variants/:id/movements
func (h *ProductHandler) GetMovements(w http.ResponseWriter, r *http.Request) {
	clientID, err := utils.GetClientIDFromContext(r.Context())
	if err != nil {
		writeError(w, http.StatusUnauthorized, "NO_CLIENT_CONTEXT", err.Error())
		return
	}

	vars := mux.Vars(r)
	variantID, err := uuid.Parse(vars["id"])
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_VARIANT_ID", "Invalid variant ID format")
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

	movements, total, err := h.service.GetMovementHistory(r.Context(), clientID, variantID, page, pageSize)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "GET_MOVEMENTS_FAILED", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, APIResponse{
		Success: true,
		Data: map[string]interface{}{
			"movements":  movements,
			"pagination": map[string]int{"page": page, "pageSize": pageSize, "total": total},
		},
	})
}

// RecordMovement handles POST /api/v1/variants/:id/movements
func (h *ProductHandler) RecordMovement(w http.ResponseWriter, r *http.Request) {
	clientID, err := utils.GetClientIDFromContext(r.Context())
	if err != nil {
		writeError(w, http.StatusUnauthorized, "NO_CLIENT_CONTEXT", err.Error())
		return
	}

	vars := mux.Vars(r)
	variantID, err := uuid.Parse(vars["id"])
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_VARIANT_ID", "Invalid variant ID format")
		return
	}

	var dto RecordMovementRequestDTO
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	// Validate required fields
	if dto.ShopID == "" || dto.MovementType == "" {
		writeError(w, http.StatusBadRequest, "MISSING_FIELDS", "Shop ID and movement type are required")
		return
	}

	shopID, err := uuid.Parse(dto.ShopID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_SHOP_ID", "Invalid shop ID format")
		return
	}

	// Parse optional UUID fields
	var referenceID *uuid.UUID
	if dto.ReferenceID != nil && *dto.ReferenceID != "" {
		refID, err := uuid.Parse(*dto.ReferenceID)
		if err != nil {
			writeError(w, http.StatusBadRequest, "INVALID_REFERENCE_ID", "Invalid reference ID format")
			return
		}
		referenceID = &refID
	}

	var performedBy *uuid.UUID
	if dto.PerformedBy != nil && *dto.PerformedBy != "" {
		perfID, err := uuid.Parse(*dto.PerformedBy)
		if err != nil {
			writeError(w, http.StatusBadRequest, "INVALID_PERFORMED_BY", "Invalid performed by ID format")
			return
		}
		performedBy = &perfID
	}

	req := &services.RecordMovementRequest{
		ClientID:      clientID,
		VariantID:     variantID,
		ShopID:        shopID,
		MovementType:  models.InventoryMovementType(dto.MovementType),
		Quantity:      dto.Quantity,
		ReferenceType: dto.ReferenceType,
		ReferenceID:   referenceID,
		Notes:         dto.Notes,
		PerformedBy:   performedBy,
	}

	movement, err := h.service.RecordMovement(r.Context(), req)
	if err != nil {
		writeError(w, http.StatusBadRequest, "RECORD_MOVEMENT_FAILED", err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, APIResponse{
		Success: true,
		Data:    movement,
	})
}

// Inventory alert handlers

// SetAlert handles PUT /api/v1/variants/:id/alerts
func (h *ProductHandler) SetAlert(w http.ResponseWriter, r *http.Request) {
	clientID, err := utils.GetClientIDFromContext(r.Context())
	if err != nil {
		writeError(w, http.StatusUnauthorized, "NO_CLIENT_CONTEXT", err.Error())
		return
	}

	vars := mux.Vars(r)
	variantID, err := uuid.Parse(vars["id"])
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_VARIANT_ID", "Invalid variant ID format")
		return
	}

	// Get shopID from query parameter
	shopIDStr := r.URL.Query().Get("shopId")
	if shopIDStr == "" {
		writeError(w, http.StatusBadRequest, "MISSING_SHOP_ID", "Shop ID query parameter is required")
		return
	}

	shopID, err := uuid.Parse(shopIDStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_SHOP_ID", "Invalid shop ID format")
		return
	}

	var dto SetInventoryAlertRequestDTO
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	req := &services.SetInventoryAlertRequest{
		ReorderPoint:      dto.ReorderPoint,
		ReorderQuantity:   dto.ReorderQuantity,
		LowStockThreshold: dto.LowStockThreshold,
		IsEnabled:         dto.IsEnabled,
	}

	alert, err := h.service.SetInventoryAlert(r.Context(), clientID, variantID, shopID, req)
	if err != nil {
		writeError(w, http.StatusBadRequest, "SET_ALERT_FAILED", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, APIResponse{
		Success: true,
		Data:    alert,
	})
}

// GetAlert handles GET /api/v1/variants/:id/alerts
func (h *ProductHandler) GetAlert(w http.ResponseWriter, r *http.Request) {
	clientID, err := utils.GetClientIDFromContext(r.Context())
	if err != nil {
		writeError(w, http.StatusUnauthorized, "NO_CLIENT_CONTEXT", err.Error())
		return
	}

	vars := mux.Vars(r)
	variantID, err := uuid.Parse(vars["id"])
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_VARIANT_ID", "Invalid variant ID format")
		return
	}

	// Get shopID from query parameter
	shopIDStr := r.URL.Query().Get("shopId")
	if shopIDStr == "" {
		writeError(w, http.StatusBadRequest, "MISSING_SHOP_ID", "Shop ID query parameter is required")
		return
	}

	shopID, err := uuid.Parse(shopIDStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_SHOP_ID", "Invalid shop ID format")
		return
	}

	alert, err := h.service.GetInventoryAlert(r.Context(), clientID, variantID, shopID)
	if err != nil {
		writeError(w, http.StatusNotFound, "ALERT_NOT_FOUND", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, APIResponse{
		Success: true,
		Data:    alert,
	})
}

// GetLowStock handles GET /api/v1/shops/:shopId/low-stock
func (h *ProductHandler) GetLowStock(w http.ResponseWriter, r *http.Request) {
	clientID, err := utils.GetClientIDFromContext(r.Context())
	if err != nil {
		writeError(w, http.StatusUnauthorized, "NO_CLIENT_CONTEXT", err.Error())
		return
	}

	vars := mux.Vars(r)
	shopID, err := uuid.Parse(vars["shopId"])
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_SHOP_ID", "Invalid shop ID format")
		return
	}

	alerts, err := h.service.CheckLowStockAlerts(r.Context(), clientID, shopID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "LOW_STOCK_CHECK_FAILED", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, APIResponse{
		Success: true,
		Data:    map[string]interface{}{"alerts": alerts},
	})
}

// Helper methods

func (h *ProductHandler) getIntQueryParam(r *http.Request, key string, defaultValue int) int {
	valueStr := r.URL.Query().Get(key)
	if valueStr == "" {
		return defaultValue
	}

	value, err := strconv.Atoi(valueStr)
	if err != nil || value < 1 {
		return defaultValue
	}

	return value
}
