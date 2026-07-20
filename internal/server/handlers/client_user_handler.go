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

// ClientUserHandler handles HTTP requests for client users
type ClientUserHandler struct {
	service services.ClientUserService
}

// NewClientUserHandler creates a new client user handler
func NewClientUserHandler(service services.ClientUserService) *ClientUserHandler {
	return &ClientUserHandler{
		service: service,
	}
}

// CreateClientUserRequest represents the request body for creating a client user
type CreateClientUserRequest struct {
	Email        string  `json:"email"`
	Username     string  `json:"username"`
	Password     string  `json:"password"`
	FirstName    string  `json:"first_name"`
	LastName     string  `json:"last_name"`
	Phone        string  `json:"phone"`
	AvatarURL    *string `json:"avatar_url,omitempty"`
	UserStatusID int     `json:"user_status_id"`
}

// UpdateClientUserRequest represents the request body for updating a client user
type UpdateClientUserRequest struct {
	Email        *string `json:"email,omitempty"`
	Username     *string `json:"username,omitempty"`
	FirstName    *string `json:"first_name,omitempty"`
	LastName     *string `json:"last_name,omitempty"`
	Phone        *string `json:"phone,omitempty"`
	AvatarURL    *string `json:"avatar_url,omitempty"`
	UserStatusID *int    `json:"user_status_id,omitempty"`
}


// ClientUserResponse represents a client user in API responses
type ClientUserResponse struct {
	ID           string  `json:"id"`
	ClientID     string  `json:"client_id"`
	Email        string  `json:"email"`
	Username     string  `json:"username"`
	FirstName    string  `json:"first_name"`
	LastName     string  `json:"last_name"`
	Phone        string  `json:"phone"`
	AvatarURL    *string `json:"avatar_url,omitempty"`
	UserStatusID int     `json:"user_status_id"`
	CreatedAt    string  `json:"created_at"`
	UpdatedAt    string  `json:"updated_at"`
}

// Create handles POST /api/v1/users
func (h *ClientUserHandler) Create(w http.ResponseWriter, r *http.Request) {
	clientID, err := utils.GetClientIDFromContext(r.Context())
	if err != nil {
		writeError(w, http.StatusUnauthorized, "NO_CLIENT_CONTEXT", err.Error())
		return
	}

	var req CreateClientUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	// Convert to service request
	serviceReq := &services.CreateClientUserRequest{
		Email:        req.Email,
		Username:     req.Username,
		Password:     req.Password,
		FirstName:    req.FirstName,
		LastName:     req.LastName,
		Phone:        req.Phone,
		AvatarURL:    req.AvatarURL,
		UserStatusID: req.UserStatusID,
	}

	user, err := h.service.CreateUser(r.Context(), clientID, serviceReq)
	if err != nil {
		writeError(w, http.StatusBadRequest, "CREATE_FAILED", err.Error())
		return
	}

	response := ClientUserResponse{
		ID:           user.ID.String(),
		ClientID:     user.ClientID.String(),
		Email:        user.Email,
		Username:     user.Username,
		FirstName:    user.FirstName,
		LastName:     user.LastName,
		Phone:        user.Phone,
		AvatarURL:    user.AvatarURL,
		UserStatusID: user.UserStatusID,
		CreatedAt:    user.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:    user.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	writeJSON(w, http.StatusCreated, APIResponse{
		Success: true,
		Data:    response,
	})
}

// Get handles GET /api/v1/users/:id
func (h *ClientUserHandler) Get(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID, err := uuid.Parse(vars["id"])
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_USER_ID", "Invalid user ID format")
		return
	}

	user, err := h.service.GetUser(r.Context(), userID)
	if err != nil {
		writeError(w, http.StatusNotFound, "USER_NOT_FOUND", err.Error())
		return
	}

	response := ClientUserResponse{
		ID:           user.ID.String(),
		ClientID:     user.ClientID.String(),
		Email:        user.Email,
		Username:     user.Username,
		FirstName:    user.FirstName,
		LastName:     user.LastName,
		Phone:        user.Phone,
		AvatarURL:    user.AvatarURL,
		UserStatusID: user.UserStatusID,
		CreatedAt:    user.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:    user.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	writeJSON(w, http.StatusOK, APIResponse{
		Success: true,
		Data:    response,
	})
}

// Update handles PUT /api/v1/users/:id
func (h *ClientUserHandler) Update(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID, err := uuid.Parse(vars["id"])
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_USER_ID", "Invalid user ID format")
		return
	}

	var req UpdateClientUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	// Convert to service request
	serviceReq := &services.UpdateClientUserRequest{
		Email:        req.Email,
		Username:     req.Username,
		FirstName:    req.FirstName,
		LastName:     req.LastName,
		Phone:        req.Phone,
		AvatarURL:    req.AvatarURL,
		UserStatusID: req.UserStatusID,
	}

	if err := h.service.UpdateUser(r.Context(), userID, serviceReq); err != nil {
		writeError(w, http.StatusBadRequest, "UPDATE_FAILED", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, APIResponse{
		Success: true,
		Data:    map[string]string{"message": "User updated successfully"},
	})
}

// Delete handles DELETE /api/v1/users/:id
func (h *ClientUserHandler) Delete(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID, err := uuid.Parse(vars["id"])
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_USER_ID", "Invalid user ID format")
		return
	}

	if err := h.service.DeleteUser(r.Context(), userID); err != nil {
		writeError(w, http.StatusBadRequest, "DELETE_FAILED", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, APIResponse{
		Success: true,
		Data:    map[string]string{"message": "User deleted successfully"},
	})
}

// List handles GET /api/v1/users
func (h *ClientUserHandler) List(w http.ResponseWriter, r *http.Request) {
	clientID, err := utils.GetClientIDFromContext(r.Context())
	if err != nil {
		writeError(w, http.StatusUnauthorized, "NO_CLIENT_CONTEXT", err.Error())
		return
	}

	// Parse pagination parameters
	page := 1
	pageSize := 20

	if pageStr := r.URL.Query().Get("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	if pageSizeStr := r.URL.Query().Get("page_size"); pageSizeStr != "" {
		if ps, err := strconv.Atoi(pageSizeStr); err == nil && ps > 0 && ps <= 100 {
			pageSize = ps
		}
	}

	users, total, err := h.service.ListUsers(r.Context(), clientID, page, pageSize)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "LIST_FAILED", err.Error())
		return
	}

	// Convert to response format
	userResponses := make([]ClientUserResponse, 0, len(users))
	for _, user := range users {
		userResponses = append(userResponses, ClientUserResponse{
			ID:           user.ID.String(),
			ClientID:     user.ClientID.String(),
			Email:        user.Email,
			Username:     user.Username,
			FirstName:    user.FirstName,
			LastName:     user.LastName,
			Phone:        user.Phone,
			AvatarURL:    user.AvatarURL,
			UserStatusID: user.UserStatusID,
			CreatedAt:    user.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			UpdatedAt:    user.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		})
	}

	totalPages := (total + pageSize - 1) / pageSize

	writeJSON(w, http.StatusOK, APIResponse{
		Success: true,
		Data:    userResponses,
		Meta: &Meta{
			Page:       page,
			PageSize:   pageSize,
			Total:      total,
			TotalPages: totalPages,
		},
	})
}

// AssignRole handles POST /api/v1/users/:id/roles
func (h *ClientUserHandler) AssignRole(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID, err := uuid.Parse(vars["id"])
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_USER_ID", "Invalid user ID format")
		return
	}

	var req AssignRoleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	if err := h.service.AssignRole(r.Context(), userID, req.RoleName); err != nil {
		writeError(w, http.StatusBadRequest, "ASSIGN_ROLE_FAILED", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, APIResponse{
		Success: true,
		Data:    map[string]string{"message": "Role assigned successfully"},
	})
}

// RemoveRole handles DELETE /api/v1/users/:id/roles/:roleId
func (h *ClientUserHandler) RemoveRole(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID, err := uuid.Parse(vars["id"])
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_USER_ID", "Invalid user ID format")
		return
	}

	// Note: roleId in the URL is actually the role name based on the service interface
	// In a real implementation, you might want to use role ID or name consistently
	roleName := vars["roleId"]

	if err := h.service.RemoveRole(r.Context(), userID, roleName); err != nil {
		writeError(w, http.StatusBadRequest, "REMOVE_ROLE_FAILED", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, APIResponse{
		Success: true,
		Data:    map[string]string{"message": "Role removed successfully"},
	})
}

