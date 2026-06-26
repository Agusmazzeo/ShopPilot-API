package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"github.com/gorilla/mux"

	"github.com/yourorg/shoppilot/internal/models"
	"github.com/yourorg/shoppilot/internal/services"
)

// PlatformUserHandler handles HTTP requests for platform user operations
type PlatformUserHandler struct {
	service services.PlatformUserService
}

// NewPlatformUserHandler creates a new platform user handler
func NewPlatformUserHandler(service services.PlatformUserService) *PlatformUserHandler {
	return &PlatformUserHandler{
		service: service,
	}
}

// Request DTOs

// CreatePlatformUserRequest represents the request to create a platform user
type CreatePlatformUserRequest struct {
	Email     string `json:"email"`
	Username  string `json:"username"`
	Password  string `json:"password"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Phone     string `json:"phone"`
}

// UpdatePlatformUserRequest represents the request to update a platform user
type UpdatePlatformUserRequest struct {
	Email     *string `json:"email,omitempty"`
	FirstName *string `json:"first_name,omitempty"`
	LastName  *string `json:"last_name,omitempty"`
	Phone     *string `json:"phone,omitempty"`
	AvatarURL *string `json:"avatar_url,omitempty"`
}

// Response DTOs

// PlatformUserResponse represents the response for a platform user
type PlatformUserResponse struct {
	ID        string   `json:"id"`
	Email     string   `json:"email"`
	Username  string   `json:"username"`
	FirstName string   `json:"first_name"`
	LastName  string   `json:"last_name"`
	Phone     string   `json:"phone"`
	AvatarURL *string  `json:"avatar_url,omitempty"`
	Status    string   `json:"status"`
	CreatedAt string   `json:"created_at"`
	UpdatedAt string   `json:"updated_at"`
}

// PlatformUserListResponse represents a paginated list of platform users
type PlatformUserListResponse struct {
	Users      []PlatformUserResponse `json:"users"`
	Page       int                    `json:"page"`
	PageSize   int                    `json:"page_size"`
	TotalCount int                    `json:"total_count"`
}

// PermissionResponse represents a permission
type PermissionResponse struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Resource    string `json:"resource"`
	Action      string `json:"action"`
}

// Helper functions

func (h *PlatformUserHandler) writeSuccess(w http.ResponseWriter, status int, data interface{}) {
	writeJSON(w, status, APIResponse{
		Success: true,
		Data:    data,
	})
}

func mapPlatformUserToResponse(user *models.PlatformUser) PlatformUserResponse {
	return PlatformUserResponse{
		ID:        user.ID.String(),
		Email:     user.Email,
		Username:  user.Username,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Phone:     user.Phone,
		AvatarURL: user.AvatarURL,
		Status:    getStatusName(user.UserStatusID),
		CreatedAt: user.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt: user.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

func getStatusName(statusID int) string {
	statusMap := map[int]string{
		1: "active",
		2: "inactive",
		3: "suspended",
	}
	if name, ok := statusMap[statusID]; ok {
		return name
	}
	return "unknown"
}

// HTTP Handlers

// Create handles POST /api/v1/platform/users
func (h *PlatformUserHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req CreatePlatformUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", "Invalid request body")
		return
	}

	// Convert to service request
	serviceReq := &services.CreatePlatformUserRequest{
		Email:     req.Email,
		Username:  req.Username,
		Password:  req.Password,
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Phone:     req.Phone,
	}

	user, err := h.service.CreateUser(r.Context(), serviceReq)
	if err != nil {
		switch err {
		case services.ErrInvalidEmail:
			writeError(w, http.StatusBadRequest, "invalid_email", err.Error())
		case services.ErrInvalidUsername:
			writeError(w, http.StatusBadRequest, "invalid_username", err.Error())
		case services.ErrPasswordTooShort:
			writeError(w, http.StatusBadRequest, "password_too_short", err.Error())
		default:
			writeError(w, http.StatusInternalServerError, "internal_error", "Failed to create user")
		}
		return
	}

	response := mapPlatformUserToResponse(user)
	h.writeSuccess(w, http.StatusCreated, response)
}

// Get handles GET /api/v1/platform/users/:id
func (h *PlatformUserHandler) Get(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]

	id, err := uuid.Parse(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_id", "Invalid user ID format")
		return
	}

	user, err := h.service.GetUser(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusNotFound, "user_not_found", "User not found")
		return
	}

	response := mapPlatformUserToResponse(user)
	h.writeSuccess(w, http.StatusOK, response)
}

// Update handles PUT /api/v1/platform/users/:id
func (h *PlatformUserHandler) Update(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]

	id, err := uuid.Parse(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_id", "Invalid user ID format")
		return
	}

	var req UpdatePlatformUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", "Invalid request body")
		return
	}

	// Convert to service request
	serviceReq := &services.UpdatePlatformUserRequest{
		Email:     req.Email,
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Phone:     req.Phone,
		AvatarURL: req.AvatarURL,
	}

	if err := h.service.UpdateUser(r.Context(), id, serviceReq); err != nil {
		switch err {
		case services.ErrInvalidEmail:
			writeError(w, http.StatusBadRequest, "invalid_email", err.Error())
		default:
			writeError(w, http.StatusInternalServerError, "internal_error", "Failed to update user")
		}
		return
	}

	h.writeSuccess(w, http.StatusOK, map[string]string{"message": "User updated successfully"})
}

// Delete handles DELETE /api/v1/platform/users/:id
func (h *PlatformUserHandler) Delete(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]

	id, err := uuid.Parse(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_id", "Invalid user ID format")
		return
	}

	if err := h.service.DeleteUser(r.Context(), id); err != nil {
		switch err {
		case services.ErrCannotDeleteSuperAdmin:
			writeError(w, http.StatusBadRequest, "cannot_delete_super_admin", err.Error())
		default:
			writeError(w, http.StatusInternalServerError, "internal_error", "Failed to delete user")
		}
		return
	}

	h.writeSuccess(w, http.StatusOK, map[string]string{"message": "User deleted successfully"})
}

// List handles GET /api/v1/platform/users
func (h *PlatformUserHandler) List(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	pageStr := r.URL.Query().Get("page")
	pageSizeStr := r.URL.Query().Get("page_size")

	page := 1
	pageSize := 10

	if pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	if pageSizeStr != "" {
		if ps, err := strconv.Atoi(pageSizeStr); err == nil && ps > 0 {
			pageSize = ps
		}
	}

	users, totalCount, err := h.service.ListUsers(r.Context(), page, pageSize)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "Failed to list users")
		return
	}

	// Map users to response
	userResponses := make([]PlatformUserResponse, 0, len(users))
	for _, user := range users {
		userResponses = append(userResponses, mapPlatformUserToResponse(user))
	}

	response := PlatformUserListResponse{
		Users:      userResponses,
		Page:       page,
		PageSize:   pageSize,
		TotalCount: totalCount,
	}

	h.writeSuccess(w, http.StatusOK, response)
}

// AssignRole handles POST /api/v1/platform/users/:id/roles
func (h *PlatformUserHandler) AssignRole(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]

	id, err := uuid.Parse(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_id", "Invalid user ID format")
		return
	}

	var req AssignRoleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", "Invalid request body")
		return
	}

	if req.RoleName == "" {
		writeError(w, http.StatusBadRequest, "missing_role_name", "Role name is required")
		return
	}

	if err := h.service.AssignRole(r.Context(), id, req.RoleName); err != nil {
		switch err {
		case services.ErrRoleNotFound:
			writeError(w, http.StatusBadRequest, "role_not_found", err.Error())
		default:
			writeError(w, http.StatusInternalServerError, "internal_error", "Failed to assign role")
		}
		return
	}

	h.writeSuccess(w, http.StatusOK, map[string]string{"message": "Role assigned successfully"})
}

// RemoveRole handles DELETE /api/v1/platform/users/:id/roles/:roleId
func (h *PlatformUserHandler) RemoveRole(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]
	roleIDStr := vars["roleId"]

	id, err := uuid.Parse(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_id", "Invalid user ID format")
		return
	}

	// Parse role ID to get role name
	// In a real implementation, you'd look up the role name from the role ID
	// For now, we'll use a simple mapping
	roleMap := map[string]string{
		"1": "super_admin",
		"2": "platform_admin",
		"3": "support",
	}

	roleName, ok := roleMap[roleIDStr]
	if !ok {
		writeError(w, http.StatusBadRequest, "invalid_role_id", "Invalid role ID")
		return
	}

	if err := h.service.RemoveRole(r.Context(), id, roleName); err != nil {
		switch err {
		case services.ErrCannotRemoveLastSuperAdmin:
			writeError(w, http.StatusBadRequest, "cannot_remove_last_super_admin", err.Error())
		case services.ErrRoleNotFound:
			writeError(w, http.StatusBadRequest, "role_not_found", err.Error())
		default:
			writeError(w, http.StatusInternalServerError, "internal_error", "Failed to remove role")
		}
		return
	}

	h.writeSuccess(w, http.StatusOK, map[string]string{"message": "Role removed successfully"})
}

// GetPermissions handles GET /api/v1/platform/users/:id/permissions
func (h *PlatformUserHandler) GetPermissions(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]

	id, err := uuid.Parse(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_id", "Invalid user ID format")
		return
	}

	permissions, err := h.service.GetUserPermissions(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "Failed to get permissions")
		return
	}

	// Map permissions to response
	permissionResponses := make([]PermissionResponse, 0, len(permissions))
	for _, perm := range permissions {
		permissionResponses = append(permissionResponses, PermissionResponse{
			ID:          perm.ID,
			Name:        perm.Name,
			Description: perm.Description,
			Resource:    perm.Resource,
			Action:      perm.Action,
		})
	}

	h.writeSuccess(w, http.StatusOK, permissionResponses)
}
