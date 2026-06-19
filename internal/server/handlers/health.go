package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/yourorg/shoppilot/app/redis"
	"github.com/yourorg/shoppilot/app/repositories"
)

// HealthHandler handles health check requests
type HealthHandler struct {
	repoManager *repositories.RepositoryManager
	redisClient *redis.Client
}

// NewHealthHandler creates a new health handler
func NewHealthHandler(repoManager *repositories.RepositoryManager, redisClient *redis.Client) *HealthHandler {
	return &HealthHandler{
		repoManager: repoManager,
		redisClient: redisClient,
	}
}

// HealthResponse represents health check response
type HealthResponse struct {
	Status   string            `json:"status"`
	Services map[string]string `json:"services"`
	Time     string            `json:"time"`
}

// Health checks overall system health
func (h *HealthHandler) Health(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()

	response := HealthResponse{
		Status:   "healthy",
		Services: make(map[string]string),
		Time:     time.Now().UTC().Format(time.RFC3339),
	}

	// Check database
	if err := h.repoManager.Health(ctx); err != nil {
		response.Status = "unhealthy"
		response.Services["database"] = "down"
	} else {
		response.Services["database"] = "up"
	}

	// Check Redis
	if err := h.redisClient.Health(ctx); err != nil {
		response.Status = "unhealthy"
		response.Services["redis"] = "down"
	} else {
		response.Services["redis"] = "up"
	}

	statusCode := http.StatusOK
	if response.Status == "unhealthy" {
		statusCode = http.StatusServiceUnavailable
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(response)
}

// Ready checks if the service is ready to accept traffic
func (h *HealthHandler) Ready(w http.ResponseWriter, r *http.Request) {
	h.Health(w, r)
}

// Live checks if the service is alive
func (h *HealthHandler) Live(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status": "alive",
		"time":   time.Now().UTC().Format(time.RFC3339),
	})
}
