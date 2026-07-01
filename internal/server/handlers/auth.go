package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/yourorg/shoppilot/internal/services"
)

// AuthHandler handles authentication-related HTTP requests
type AuthHandler struct {
	authService services.AuthServiceI
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(authService services.AuthServiceI) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

// LoginRequest represents the login request payload
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// LoginResponse represents the login response payload
type LoginResponse struct {
	Token string      `json:"token"`
	User  interface{} `json:"user"`
}

// RefreshTokenRequest represents the refresh token request payload
type RefreshTokenRequest struct {
	Token string `json:"token"`
}

// RefreshTokenResponse represents the refresh token response payload
type RefreshTokenResponse struct {
	Token string `json:"token"`
}

// Login handles POST /api/v1/auth/login
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErrorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate required fields
	if req.Email == "" || req.Password == "" {
		writeErrorResponse(w, http.StatusBadRequest, "Email and password are required")
		return
	}

	// Authenticate user through service
	loginResult, err := h.authService.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		switch err {
		case services.ErrInvalidCredentials:
			writeErrorResponse(w, http.StatusUnauthorized, "Invalid credentials")
		case services.ErrUserNotActive:
			writeErrorResponse(w, http.StatusUnauthorized, "User account is not active")
		default:
			writeErrorResponse(w, http.StatusInternalServerError, "Authentication failed")
		}
		return
	}

	response := LoginResponse{
		Token: loginResult.Token,
		User:  loginResult.User,
	}

	writeJSONResponse(w, http.StatusOK, response)
}

// Logout handles POST /api/v1/auth/logout
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	// For now, logout is just a client-side operation
	// In a real implementation, you might invalidate the token on the server side
	response := map[string]string{
		"message": "Logged out successfully",
	}
	writeJSONResponse(w, http.StatusOK, response)
}

// RefreshToken handles POST /api/v1/auth/refresh
func (h *AuthHandler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	var req RefreshTokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErrorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate required fields
	if req.Token == "" {
		writeErrorResponse(w, http.StatusBadRequest, "Token is required")
		return
	}

	// Refresh token through service
	newToken, err := h.authService.RefreshToken(r.Context(), req.Token)
	if err != nil {
		if err == services.ErrInvalidToken {
			writeErrorResponse(w, http.StatusUnauthorized, "Invalid token")
			return
		}
		writeErrorResponse(w, http.StatusInternalServerError, "Token refresh failed")
		return
	}

	response := RefreshTokenResponse{
		Token: newToken,
	}

	writeJSONResponse(w, http.StatusOK, response)
}

// Helper functions
func writeJSONResponse(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}

func writeErrorResponse(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	response := map[string]string{"error": message}
	json.NewEncoder(w).Encode(response)
}
