package server

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/rs/cors"
	"github.com/yourorg/shoppilot/internal/server/handlers"
	"github.com/yourorg/shoppilot/internal/services"
)

// setupRouter configures all routes and middleware
func (s *Server) setupRouter() http.Handler {
	router := mux.NewRouter()

	// Apply global middleware
	router.Use(s.loggingMiddleware)
	router.Use(s.recoveryMiddleware)

	// Health check endpoints
	healthHandler := handlers.NewHealthHandler(s.repoManager, s.redisClient)
	router.HandleFunc("/health", healthHandler.Health).Methods("GET")
	router.HandleFunc("/health/ready", healthHandler.Ready).Methods("GET")
	router.HandleFunc("/health/live", healthHandler.Live).Methods("GET")

	// Initialize services
	platformUserService := services.NewPlatformUserService(s.repoManager.PlatformUsers)
	clientService := services.NewClientService(s.repoManager.Clients)
	clientUserService := services.NewClientUserService(s.repoManager.ClientUsers)
	shopService := services.NewShopService(s.repoManager.Shops, s.repoManager.Clients)
	productService := services.NewProductService(s.repoManager.Products)

	// Initialize handlers
	platformUserHandler := handlers.NewPlatformUserHandler(platformUserService)
	clientHandler := handlers.NewClientHandler(clientService)
	clientUserHandler := handlers.NewClientUserHandler(clientUserService)
	shopHandler := handlers.NewShopHandler(shopService)
	productHandler := handlers.NewProductHandler(productService)

	// API version 1
	apiV1 := router.PathPrefix("/api/v1").Subrouter()

	// Platform routes (admin only)
	platform := apiV1.PathPrefix("/platform").Subrouter()
	platform.HandleFunc("/users", platformUserHandler.Create).Methods("POST")
	platform.HandleFunc("/users/{id}", platformUserHandler.Get).Methods("GET")
	platform.HandleFunc("/users/{id}", platformUserHandler.Update).Methods("PUT")
	platform.HandleFunc("/users/{id}", platformUserHandler.Delete).Methods("DELETE")
	platform.HandleFunc("/users", platformUserHandler.List).Methods("GET")
	platform.HandleFunc("/users/{id}/roles", platformUserHandler.AssignRole).Methods("POST")
	platform.HandleFunc("/users/{id}/roles/{roleId}", platformUserHandler.RemoveRole).Methods("DELETE")
	platform.HandleFunc("/users/{id}/permissions", platformUserHandler.GetPermissions).Methods("GET")

	// Client routes
	apiV1.HandleFunc("/clients", clientHandler.Create).Methods("POST")
	apiV1.HandleFunc("/clients/{id}", clientHandler.Get).Methods("GET")
	apiV1.HandleFunc("/clients/slug/{slug}", clientHandler.GetBySlug).Methods("GET")
	apiV1.HandleFunc("/clients/{id}", clientHandler.Update).Methods("PUT")
	apiV1.HandleFunc("/clients/{id}", clientHandler.Delete).Methods("DELETE")
	apiV1.HandleFunc("/clients", clientHandler.List).Methods("GET")
	apiV1.HandleFunc("/clients/{id}/activate", clientHandler.Activate).Methods("POST")
	apiV1.HandleFunc("/clients/{id}/deactivate", clientHandler.Deactivate).Methods("POST")

	// Client user routes
	apiV1.HandleFunc("/clients/{clientId}/users", clientUserHandler.Create).Methods("POST")
	apiV1.HandleFunc("/clients/{clientId}/users/{id}", clientUserHandler.Get).Methods("GET")
	apiV1.HandleFunc("/clients/{clientId}/users/{id}", clientUserHandler.Update).Methods("PUT")
	apiV1.HandleFunc("/clients/{clientId}/users/{id}", clientUserHandler.Delete).Methods("DELETE")
	apiV1.HandleFunc("/clients/{clientId}/users", clientUserHandler.List).Methods("GET")
	apiV1.HandleFunc("/clients/{clientId}/users/{id}/roles", clientUserHandler.AssignRole).Methods("POST")
	apiV1.HandleFunc("/clients/{clientId}/users/{id}/roles/{roleId}", clientUserHandler.RemoveRole).Methods("DELETE")

	// Shop routes
	apiV1.HandleFunc("/clients/{clientId}/shops", shopHandler.Create).Methods("POST")
	apiV1.HandleFunc("/clients/{clientId}/shops/{id}", shopHandler.Get).Methods("GET")
	apiV1.HandleFunc("/clients/{clientId}/shops/slug/{slug}", shopHandler.GetBySlug).Methods("GET")
	apiV1.HandleFunc("/clients/{clientId}/shops/{id}", shopHandler.Update).Methods("PUT")
	apiV1.HandleFunc("/clients/{clientId}/shops/{id}", shopHandler.Delete).Methods("DELETE")
	apiV1.HandleFunc("/clients/{clientId}/shops", shopHandler.List).Methods("GET")
	apiV1.HandleFunc("/shops/{id}/users", shopHandler.AssignUser).Methods("POST")
	apiV1.HandleFunc("/shops/{id}/users/{userRoleId}", shopHandler.RemoveUser).Methods("DELETE")
	apiV1.HandleFunc("/shops/{id}/users", shopHandler.GetUsers).Methods("GET")

	// Product routes
	apiV1.HandleFunc("/clients/{clientId}/products", productHandler.Create).Methods("POST")
	apiV1.HandleFunc("/clients/{clientId}/products/{id}", productHandler.Get).Methods("GET")
	apiV1.HandleFunc("/clients/{clientId}/products/{id}", productHandler.Update).Methods("PUT")
	apiV1.HandleFunc("/clients/{clientId}/products/{id}", productHandler.Delete).Methods("DELETE")
	apiV1.HandleFunc("/clients/{clientId}/products", productHandler.List).Methods("GET")
	apiV1.HandleFunc("/clients/{clientId}/products/search", productHandler.Search).Methods("GET")

	// Product variant routes
	apiV1.HandleFunc("/clients/{clientId}/products/{productId}/variants", productHandler.CreateVariant).Methods("POST")
	apiV1.HandleFunc("/clients/{clientId}/variants/{id}", productHandler.GetVariant).Methods("GET")
	apiV1.HandleFunc("/clients/{clientId}/variants/{id}", productHandler.UpdateVariant).Methods("PUT")
	apiV1.HandleFunc("/clients/{clientId}/variants/{id}", productHandler.DeleteVariant).Methods("DELETE")
	apiV1.HandleFunc("/clients/{clientId}/products/{productId}/variants", productHandler.ListVariants).Methods("GET")

	// Inventory routes
	apiV1.HandleFunc("/clients/{clientId}/variants/{id}/inventory/adjust", productHandler.AdjustInventory).Methods("POST")
	apiV1.HandleFunc("/clients/{clientId}/variants/{id}/inventory", productHandler.SetInventory).Methods("PUT")
	apiV1.HandleFunc("/clients/{clientId}/variants/{id}/inventory", productHandler.CheckStock).Methods("GET")

	// Root endpoint
	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"message":"ShopPilot API","version":"0.1.0"}`))
	}).Methods("GET")

	// CORS configuration
	corsHandler := cors.New(cors.Options{
		AllowedOrigins: []string{s.cfg.Frontend.BaseURL},
		AllowedMethods: []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders: []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders: []string{"Link"},
		AllowCredentials: true,
		MaxAge: 300,
	})

	return corsHandler.Handler(router)
}
