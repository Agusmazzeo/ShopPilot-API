package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/yourorg/shoppilot/internal/services"
	"github.com/yourorg/shoppilot/internal/utils"
)

// AuthMiddleware creates a middleware that validates JWT tokens
func AuthMiddleware(authService services.AuthServiceI) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get the Authorization header
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				writeErrorResponse(w, http.StatusUnauthorized, "Authorization header is required")
				return
			}

			// Check if it's a Bearer token
			if !strings.HasPrefix(authHeader, "Bearer ") {
				writeErrorResponse(w, http.StatusUnauthorized, "Authorization header must be a Bearer token")
				return
			}

			// Extract the token
			token := strings.TrimPrefix(authHeader, "Bearer ")
			if token == "" {
				writeErrorResponse(w, http.StatusUnauthorized, "Token is required")
				return
			}

			// Validate the token
			user, err := authService.ValidateToken(r.Context(), token)
			if err != nil {
				writeErrorResponse(w, http.StatusUnauthorized, "Invalid or expired token")
				return
			}

			// Store user information in the context for use in handlers
			ctx := context.WithValue(r.Context(), utils.UserContextKey, user)
			ctx = context.WithValue(ctx, utils.UserIDContextKey, user.ID)
			ctx = context.WithValue(ctx, utils.UserEmailContextKey, user.Email)
			ctx = context.WithValue(ctx, utils.UserRoleContextKey, user.RoleName)
			ctx = context.WithValue(ctx, utils.UserPermissionsContextKey, user.Permissions)

			// Continue to the next handler with updated context
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequirePermission creates a middleware that checks if the user has a specific permission
func RequirePermission(resource, action string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get user from context (should be set by AuthMiddleware)
			user, ok := utils.GetUserFromContext(r.Context())
			if !ok {
				writeErrorResponse(w, http.StatusUnauthorized, "User not found in context")
				return
			}

			// Check if user has the required permission
			hasPermission := false
			for _, permission := range user.Permissions {
				if permission.Resource == resource && permission.Action == action {
					hasPermission = true
					break
				}
			}

			if !hasPermission {
				writeErrorResponse(w, http.StatusForbidden, "Insufficient permissions for "+resource+":"+action)
				return
			}

			// Continue to the next handler
			next.ServeHTTP(w, r)
		})
	}
}

// RequireRole creates a middleware that checks if the user has a specific role
func RequireRole(roleName string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get user role from context (should be set by AuthMiddleware)
			userRole := r.Context().Value(utils.UserRoleContextKey)
			if userRole == nil {
				writeErrorResponse(w, http.StatusUnauthorized, "User role not found in context")
				return
			}

			role, ok := userRole.(string)
			if !ok || role != roleName {
				writeErrorResponse(w, http.StatusForbidden, "Insufficient role permissions")
				return
			}

			// Continue to the next handler
			next.ServeHTTP(w, r)
		})
	}
}

// writeErrorResponse writes an error response in JSON format
func writeErrorResponse(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	response := map[string]string{"error": message}
	_ = json.NewEncoder(w).Encode(response)
}
