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

// ClientHandler handles HTTP requests for client operations
type ClientHandler struct {
	service services.ClientService
}

// NewClientHandler creates a new client handler instance
func NewClientHandler(service services.ClientService) *ClientHandler {
	return &ClientHandler{
		service: service,
	}
}

// ClientResponse represents the API response for a client
type ClientResponse struct {
	ID           uuid.UUID `json:"id"`
	Name         string    `json:"name"`
	Slug         string    `json:"slug"`
	Description  string    `json:"description,omitempty"`
	ContactEmail string    `json:"contact_email"`
	ContactPhone string    `json:"contact_phone,omitempty"`
	WebsiteURL   string    `json:"website_url,omitempty"`
	LogoURL      *string   `json:"logo_url,omitempty"`
	IsActive     bool      `json:"is_active"`
	CreatedAt    string    `json:"created_at"`
	UpdatedAt    string    `json:"updated_at"`
}

// toClientResponse converts a models.Client to a ClientResponse
func toClientResponse(client *models.Client) *ClientResponse {
	return &ClientResponse{
		ID:           client.ID,
		Name:         client.Name,
		Slug:         client.Slug,
		Description:  client.Description,
		ContactEmail: client.ContactEmail,
		ContactPhone: client.ContactPhone,
		WebsiteURL:   client.WebsiteURL,
		LogoURL:      client.LogoURL,
		IsActive:     client.IsActive,
		CreatedAt:    client.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:    client.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

// Create handles POST /api/v1/clients
func (h *ClientHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req services.CreateClientRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	client, err := h.service.CreateClient(r.Context(), &req)
	if err != nil {
		// Check for specific error types
		if err.Error() == "client with slug already exists" {
			writeError(w, http.StatusConflict, "SLUG_EXISTS", err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, "CREATE_FAILED", err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, APIResponse{
		Success: true,
		Data:    toClientResponse(client),
	})
}

// Get handles GET /api/v1/clients/:id
func (h *ClientHandler) Get(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]

	id, err := uuid.Parse(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "Invalid client ID format")
		return
	}

	client, err := h.service.GetClient(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "Client not found")
		return
	}

	writeJSON(w, http.StatusOK, APIResponse{
		Success: true,
		Data:    toClientResponse(client),
	})
}

// GetBySlug handles GET /api/v1/clients/slug/:slug
func (h *ClientHandler) GetBySlug(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	slug := vars["slug"]

	if slug == "" {
		writeError(w, http.StatusBadRequest, "INVALID_SLUG", "Slug is required")
		return
	}

	client, err := h.service.GetClientBySlug(r.Context(), slug)
	if err != nil {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "Client not found")
		return
	}

	writeJSON(w, http.StatusOK, APIResponse{
		Success: true,
		Data:    toClientResponse(client),
	})
}

// Update handles PUT /api/v1/clients/:id
func (h *ClientHandler) Update(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]

	id, err := uuid.Parse(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "Invalid client ID format")
		return
	}

	var req services.UpdateClientRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	if err := h.service.UpdateClient(r.Context(), id, &req); err != nil {
		// Check for specific error types
		if err.Error() == "client not found" {
			writeError(w, http.StatusNotFound, "NOT_FOUND", err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, "UPDATE_FAILED", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, APIResponse{
		Success: true,
		Data:    map[string]string{"message": "Client updated successfully"},
	})
}

// Delete handles DELETE /api/v1/clients/:id
func (h *ClientHandler) Delete(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]

	id, err := uuid.Parse(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "Invalid client ID format")
		return
	}

	if err := h.service.DeleteClient(r.Context(), id); err != nil {
		if err.Error() == "client not found" {
			writeError(w, http.StatusNotFound, "NOT_FOUND", err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, "DELETE_FAILED", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, APIResponse{
		Success: true,
		Data:    map[string]string{"message": "Client deleted successfully"},
	})
}

// List handles GET /api/v1/clients
func (h *ClientHandler) List(w http.ResponseWriter, r *http.Request) {
	// Parse pagination parameters
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

	clients, total, err := h.service.ListClients(r.Context(), page, pageSize)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "LIST_FAILED", err.Error())
		return
	}

	// Convert to response format
	clientResponses := make([]*ClientResponse, len(clients))
	for i, client := range clients {
		clientResponses[i] = toClientResponse(client)
	}

	// Calculate total pages
	totalPages := (total + pageSize - 1) / pageSize

	writeJSON(w, http.StatusOK, APIResponse{
		Success: true,
		Data:    clientResponses,
		Meta: &Meta{
			Page:       page,
			PageSize:   pageSize,
			Total:      total,
			TotalPages: totalPages,
		},
	})
}

// Activate handles POST /api/v1/clients/:id/activate
func (h *ClientHandler) Activate(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]

	id, err := uuid.Parse(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "Invalid client ID format")
		return
	}

	if err := h.service.ActivateClient(r.Context(), id); err != nil {
		if err.Error() == "client not found" {
			writeError(w, http.StatusNotFound, "NOT_FOUND", err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, "ACTIVATE_FAILED", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, APIResponse{
		Success: true,
		Data:    map[string]string{"message": "Client activated successfully"},
	})
}

// Deactivate handles POST /api/v1/clients/:id/deactivate
func (h *ClientHandler) Deactivate(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]

	id, err := uuid.Parse(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "Invalid client ID format")
		return
	}

	if err := h.service.DeactivateClient(r.Context(), id); err != nil {
		if err.Error() == "client not found" {
			writeError(w, http.StatusNotFound, "NOT_FOUND", err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, "DEACTIVATE_FAILED", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, APIResponse{
		Success: true,
		Data:    map[string]string{"message": "Client deactivated successfully"},
	})
}
