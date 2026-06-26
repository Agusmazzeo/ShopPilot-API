package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/yourorg/shoppilot/internal/services"
)

// ShopHandler handles HTTP requests for shop operations
type ShopHandler struct {
	service services.ShopService
}

// NewShopHandler creates a new shop handler
func NewShopHandler(service services.ShopService) *ShopHandler {
	return &ShopHandler{
		service: service,
	}
}

// AssignUserRequest represents the request to assign a user to a shop
type AssignUserRequest struct {
	ClientUserID uuid.UUID `json:"clientUserId" validate:"required"`
	RoleName     string    `json:"roleName" validate:"required"`
}

// Create handles POST /api/v1/clients/:clientId/shops
func (h *ShopHandler) Create(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	clientID, err := uuid.Parse(vars["clientId"])
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_CLIENT_ID", "Invalid client ID format")
		return
	}

	var req services.CreateShopRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	shop, err := h.service.CreateShop(r.Context(), clientID, &req)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "CREATE_FAILED", err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, APIResponse{
		Success: true,
		Data:    shop,
	})
}

// Get handles GET /api/v1/clients/:clientId/shops/:id
func (h *ShopHandler) Get(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	clientID, err := uuid.Parse(vars["clientId"])
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_CLIENT_ID", "Invalid client ID format")
		return
	}

	shopID, err := uuid.Parse(vars["id"])
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_SHOP_ID", "Invalid shop ID format")
		return
	}

	shop, err := h.service.GetShop(r.Context(), clientID, shopID)
	if err != nil {
		writeError(w, http.StatusNotFound, "SHOP_NOT_FOUND", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, APIResponse{
		Success: true,
		Data:    shop,
	})
}

// GetBySlug handles GET /api/v1/clients/:clientId/shops/slug/:slug
func (h *ShopHandler) GetBySlug(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	clientID, err := uuid.Parse(vars["clientId"])
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_CLIENT_ID", "Invalid client ID format")
		return
	}

	slug := vars["slug"]
	if slug == "" {
		writeError(w, http.StatusBadRequest, "INVALID_SLUG", "Slug is required")
		return
	}

	shop, err := h.service.GetShopBySlug(r.Context(), clientID, slug)
	if err != nil {
		writeError(w, http.StatusNotFound, "SHOP_NOT_FOUND", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, APIResponse{
		Success: true,
		Data:    shop,
	})
}

// Update handles PUT /api/v1/clients/:clientId/shops/:id
func (h *ShopHandler) Update(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	clientID, err := uuid.Parse(vars["clientId"])
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_CLIENT_ID", "Invalid client ID format")
		return
	}

	shopID, err := uuid.Parse(vars["id"])
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_SHOP_ID", "Invalid shop ID format")
		return
	}

	var req services.UpdateShopRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	if err := h.service.UpdateShop(r.Context(), clientID, shopID, &req); err != nil {
		writeError(w, http.StatusInternalServerError, "UPDATE_FAILED", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, APIResponse{
		Success: true,
		Data:    map[string]string{"message": "Shop updated successfully"},
	})
}

// Delete handles DELETE /api/v1/clients/:clientId/shops/:id
func (h *ShopHandler) Delete(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	clientID, err := uuid.Parse(vars["clientId"])
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_CLIENT_ID", "Invalid client ID format")
		return
	}

	shopID, err := uuid.Parse(vars["id"])
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_SHOP_ID", "Invalid shop ID format")
		return
	}

	if err := h.service.DeleteShop(r.Context(), clientID, shopID); err != nil {
		writeError(w, http.StatusInternalServerError, "DELETE_FAILED", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, APIResponse{
		Success: true,
		Data:    map[string]string{"message": "Shop deleted successfully"},
	})
}

// List handles GET /api/v1/clients/:clientId/shops
func (h *ShopHandler) List(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	clientID, err := uuid.Parse(vars["clientId"])
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_CLIENT_ID", "Invalid client ID format")
		return
	}

	// Parse pagination parameters
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}

	pageSize, _ := strconv.Atoi(r.URL.Query().Get("pageSize"))
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	shops, total, err := h.service.ListShops(r.Context(), clientID, page, pageSize)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "LIST_FAILED", err.Error())
		return
	}

	// Calculate total pages
	totalPages := 0
	if total > 0 {
		totalPages = (total + pageSize - 1) / pageSize
	}

	writeJSON(w, http.StatusOK, APIResponse{
		Success: true,
		Data:    shops,
		Meta: &Meta{
			Page:       page,
			PageSize:   pageSize,
			Total:      total,
			TotalPages: totalPages,
		},
	})
}

// AssignUser handles POST /api/v1/shops/:id/users
func (h *ShopHandler) AssignUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	shopID, err := uuid.Parse(vars["id"])
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_SHOP_ID", "Invalid shop ID format")
		return
	}

	var req AssignUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	if err := h.service.AssignUserToShop(r.Context(), shopID, req.ClientUserID, req.RoleName); err != nil {
		writeError(w, http.StatusInternalServerError, "ASSIGN_FAILED", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, APIResponse{
		Success: true,
		Data:    map[string]string{"message": "User assigned to shop successfully"},
	})
}

// RemoveUser handles DELETE /api/v1/shops/:id/users/:userRoleId
func (h *ShopHandler) RemoveUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	shopID, err := uuid.Parse(vars["id"])
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_SHOP_ID", "Invalid shop ID format")
		return
	}

	userRoleID, err := strconv.Atoi(vars["userRoleId"])
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_USER_ROLE_ID", "Invalid user role ID format")
		return
	}

	if err := h.service.RemoveUserFromShop(r.Context(), shopID, userRoleID); err != nil {
		writeError(w, http.StatusInternalServerError, "REMOVE_FAILED", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, APIResponse{
		Success: true,
		Data:    map[string]string{"message": "User removed from shop successfully"},
	})
}

// GetUsers handles GET /api/v1/shops/:id/users
func (h *ShopHandler) GetUsers(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	shopID, err := uuid.Parse(vars["id"])
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_SHOP_ID", "Invalid shop ID format")
		return
	}

	shopUsers, err := h.service.GetShopUsers(r.Context(), shopID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "GET_USERS_FAILED", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, APIResponse{
		Success: true,
		Data:    shopUsers,
	})
}

