package server

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/rs/cors"
	"github.com/yourorg/shoppilot/internal/server/handlers"
	"github.com/yourorg/shoppilot/internal/server/middleware"
	"github.com/yourorg/shoppilot/internal/services"
)

// setupRouter configures all routes and middleware
func (s *Server) setupRouter() http.Handler {
	router := mux.NewRouter()

	// Apply global middleware
	router.Use(s.loggingMiddleware)
	router.Use(s.recoveryMiddleware)

	// Health check endpoints (public - no auth required)
	healthHandler := handlers.NewHealthHandler(s.repoManager, s.redisClient)
	router.HandleFunc("/health", healthHandler.Health).Methods("GET")
	router.HandleFunc("/health/ready", healthHandler.Ready).Methods("GET")
	router.HandleFunc("/health/live", healthHandler.Live).Methods("GET")

	// Initialize auth handler
	authHandler := handlers.NewAuthHandler(s.authService)

	// Create auth middleware
	authMiddleware := middleware.AuthMiddleware(s.authService)

	// Helper function to chain auth + permission middleware
	requirePermission := func(resource, action string, handler http.HandlerFunc) http.Handler {
		return authMiddleware(middleware.RequirePermission(resource, action)(http.HandlerFunc(handler)))
	}

	// Authentication routes (public - no auth required)
	router.HandleFunc("/api/v1/auth/login", authHandler.Login).Methods("POST")
	router.HandleFunc("/api/v1/auth/logout", authHandler.Logout).Methods("POST")
	router.HandleFunc("/api/v1/auth/refresh", authHandler.RefreshToken).Methods("POST")

	// Initialize services
	platformUserService := services.NewPlatformUserService(s.repoManager.PlatformUsers)
	clientService := services.NewClientService(s.repoManager.Clients)
	clientUserService := services.NewClientUserService(s.repoManager.ClientUsers)
	shopService := services.NewShopService(s.repoManager.Shops, s.repoManager.Clients)
	productService := services.NewProductService(
		s.repoManager.Products,
		s.repoManager.InventoryMovements,
		s.repoManager.InventoryAlerts,
	)
	supplierService := services.NewSupplierService(s.repoManager.Suppliers)
	customerService := services.NewCustomerService(s.repoManager.Customers)
	purchaseOrderService := services.NewPurchaseOrderService(
		s.repoManager.PurchaseOrders,
		s.repoManager.Products,
		s.repoManager.InventoryMovements,
	)
	salesOrderService := services.NewSalesOrderService(
		s.repoManager.SalesOrders,
		s.repoManager.Products,
		s.repoManager.InventoryMovements,
	)

	// Initialize handlers
	platformUserHandler := handlers.NewPlatformUserHandler(platformUserService)
	clientHandler := handlers.NewClientHandler(clientService)
	clientUserHandler := handlers.NewClientUserHandler(clientUserService)
	shopHandler := handlers.NewShopHandler(shopService)
	productHandler := handlers.NewProductHandler(productService)
	supplierHandler := handlers.NewSupplierHandler(supplierService)
	customerHandler := handlers.NewCustomerHandler(customerService)
	purchaseOrderHandler := handlers.NewPurchaseOrderHandler(purchaseOrderService)
	salesOrderHandler := handlers.NewSalesOrderHandler(salesOrderService)

	// API version 1
	apiV1 := router.PathPrefix("/api/v1").Subrouter()

	// Platform routes (admin only - require platform_users permission)
	platform := apiV1.PathPrefix("/platform").Subrouter()
	platform.Handle("/users", requirePermission("platform_users", "create", platformUserHandler.Create)).Methods("POST")
	platform.Handle("/users/{id}", requirePermission("platform_users", "read", platformUserHandler.Get)).Methods("GET")
	platform.Handle("/users/{id}", requirePermission("platform_users", "update", platformUserHandler.Update)).Methods("PUT")
	platform.Handle("/users/{id}", requirePermission("platform_users", "delete", platformUserHandler.Delete)).Methods("DELETE")
	platform.Handle("/users", requirePermission("platform_users", "read", platformUserHandler.List)).Methods("GET")
	platform.Handle("/users/{id}/roles", requirePermission("platform_users", "update", platformUserHandler.AssignRole)).Methods("POST")
	platform.Handle("/users/{id}/roles/{roleId}", requirePermission("platform_users", "update", platformUserHandler.RemoveRole)).Methods("DELETE")
	platform.Handle("/users/{id}/permissions", requirePermission("platform_users", "read", platformUserHandler.GetPermissions)).Methods("GET")

	// Client routes (require clients permission)
	apiV1.Handle("/clients", requirePermission("clients", "create", clientHandler.Create)).Methods("POST")
	apiV1.Handle("/clients/{id}", requirePermission("clients", "read", clientHandler.Get)).Methods("GET")
	apiV1.Handle("/clients/slug/{slug}", requirePermission("clients", "read", clientHandler.GetBySlug)).Methods("GET")
	apiV1.Handle("/clients/{id}", requirePermission("clients", "update", clientHandler.Update)).Methods("PUT")
	apiV1.Handle("/clients/{id}", requirePermission("clients", "delete", clientHandler.Delete)).Methods("DELETE")
	apiV1.Handle("/clients", requirePermission("clients", "read", clientHandler.List)).Methods("GET")
	apiV1.Handle("/clients/{id}/activate", requirePermission("clients", "update", clientHandler.Activate)).Methods("POST")
	apiV1.Handle("/clients/{id}/deactivate", requirePermission("clients", "update", clientHandler.Deactivate)).Methods("POST")

	// Client user routes (require client_users permission)
	apiV1.Handle("/clients/{clientId}/users", requirePermission("client_users", "create", clientUserHandler.Create)).Methods("POST")
	apiV1.Handle("/clients/{clientId}/users/{id}", requirePermission("client_users", "read", clientUserHandler.Get)).Methods("GET")
	apiV1.Handle("/clients/{clientId}/users/{id}", requirePermission("client_users", "update", clientUserHandler.Update)).Methods("PUT")
	apiV1.Handle("/clients/{clientId}/users/{id}", requirePermission("client_users", "delete", clientUserHandler.Delete)).Methods("DELETE")
	apiV1.Handle("/clients/{clientId}/users", requirePermission("client_users", "read", clientUserHandler.List)).Methods("GET")
	apiV1.Handle("/clients/{clientId}/users/{id}/roles", requirePermission("client_users", "update", clientUserHandler.AssignRole)).Methods("POST")
	apiV1.Handle("/clients/{clientId}/users/{id}/roles/{roleId}", requirePermission("client_users", "update", clientUserHandler.RemoveRole)).Methods("DELETE")

	// Shop routes (require shops permission)
	apiV1.Handle("/clients/{clientId}/shops", requirePermission("shops", "create", shopHandler.Create)).Methods("POST")
	apiV1.Handle("/clients/{clientId}/shops/{id}", requirePermission("shops", "read", shopHandler.Get)).Methods("GET")
	apiV1.Handle("/clients/{clientId}/shops/slug/{slug}", requirePermission("shops", "read", shopHandler.GetBySlug)).Methods("GET")
	apiV1.Handle("/clients/{clientId}/shops/{id}", requirePermission("shops", "update", shopHandler.Update)).Methods("PUT")
	apiV1.Handle("/clients/{clientId}/shops/{id}", requirePermission("shops", "delete", shopHandler.Delete)).Methods("DELETE")
	apiV1.Handle("/clients/{clientId}/shops", requirePermission("shops", "read", shopHandler.List)).Methods("GET")
	apiV1.Handle("/shops/{id}/users", requirePermission("shops", "update", shopHandler.AssignUser)).Methods("POST")
	apiV1.Handle("/shops/{id}/users/{userRoleId}", requirePermission("shops", "update", shopHandler.RemoveUser)).Methods("DELETE")
	apiV1.Handle("/shops/{id}/users", requirePermission("shops", "read", shopHandler.GetUsers)).Methods("GET")

	// Product routes (require products permission)
	apiV1.Handle("/clients/{clientId}/products", requirePermission("products", "create", productHandler.Create)).Methods("POST")
	apiV1.Handle("/clients/{clientId}/products/{id}", requirePermission("products", "read", productHandler.Get)).Methods("GET")
	apiV1.Handle("/clients/{clientId}/products/{id}", requirePermission("products", "update", productHandler.Update)).Methods("PUT")
	apiV1.Handle("/clients/{clientId}/products/{id}", requirePermission("products", "delete", productHandler.Delete)).Methods("DELETE")
	apiV1.Handle("/clients/{clientId}/products", requirePermission("products", "read", productHandler.List)).Methods("GET")
	apiV1.Handle("/clients/{clientId}/products/search", requirePermission("products", "read", productHandler.Search)).Methods("GET")

	// Product variant routes (require products permission)
	apiV1.Handle("/clients/{clientId}/products/{productId}/variants", requirePermission("products", "create", productHandler.CreateVariant)).Methods("POST")
	apiV1.Handle("/clients/{clientId}/variants/{id}", requirePermission("products", "read", productHandler.GetVariant)).Methods("GET")
	apiV1.Handle("/clients/{clientId}/variants/{id}", requirePermission("products", "update", productHandler.UpdateVariant)).Methods("PUT")
	apiV1.Handle("/clients/{clientId}/variants/{id}", requirePermission("products", "delete", productHandler.DeleteVariant)).Methods("DELETE")
	apiV1.Handle("/clients/{clientId}/products/{productId}/variants", requirePermission("products", "read", productHandler.ListVariants)).Methods("GET")

	// Inventory routes (require inventory permission)
	apiV1.Handle("/clients/{clientId}/variants/{id}/inventory/adjust", requirePermission("inventory", "update", productHandler.AdjustInventory)).Methods("POST")
	apiV1.Handle("/clients/{clientId}/variants/{id}/inventory", requirePermission("inventory", "update", productHandler.SetInventory)).Methods("PUT")
	apiV1.Handle("/clients/{clientId}/variants/{id}/inventory", requirePermission("inventory", "read", productHandler.CheckStock)).Methods("GET")
	apiV1.Handle("/clients/{clientId}/variants/{id}/movements", requirePermission("inventory", "read", productHandler.GetMovements)).Methods("GET")
	apiV1.Handle("/clients/{clientId}/variants/{id}/movements", requirePermission("inventory", "create", productHandler.RecordMovement)).Methods("POST")
	apiV1.Handle("/clients/{clientId}/variants/{id}/alerts", requirePermission("inventory", "update", productHandler.SetAlert)).Methods("PUT")
	apiV1.Handle("/clients/{clientId}/variants/{id}/alerts", requirePermission("inventory", "read", productHandler.GetAlert)).Methods("GET")
	apiV1.Handle("/clients/{clientId}/shops/{shopId}/low-stock", requirePermission("inventory", "read", productHandler.GetLowStock)).Methods("GET")

	// Supplier routes (require suppliers permission)
	apiV1.Handle("/clients/{clientId}/suppliers", requirePermission("suppliers", "create", supplierHandler.Create)).Methods("POST")
	apiV1.Handle("/clients/{clientId}/suppliers/{id}", requirePermission("suppliers", "read", supplierHandler.Get)).Methods("GET")
	apiV1.Handle("/clients/{clientId}/suppliers/{id}", requirePermission("suppliers", "update", supplierHandler.Update)).Methods("PUT")
	apiV1.Handle("/clients/{clientId}/suppliers/{id}", requirePermission("suppliers", "delete", supplierHandler.Delete)).Methods("DELETE")
	apiV1.Handle("/clients/{clientId}/suppliers", requirePermission("suppliers", "read", supplierHandler.List)).Methods("GET")

	// Customer routes (require customers permission)
	apiV1.Handle("/clients/{clientId}/customers", requirePermission("customers", "create", customerHandler.Create)).Methods("POST")
	apiV1.Handle("/clients/{clientId}/customers/{id}", requirePermission("customers", "read", customerHandler.Get)).Methods("GET")
	apiV1.Handle("/clients/{clientId}/customers/{id}", requirePermission("customers", "update", customerHandler.Update)).Methods("PUT")
	apiV1.Handle("/clients/{clientId}/customers/{id}", requirePermission("customers", "delete", customerHandler.Delete)).Methods("DELETE")
	apiV1.Handle("/clients/{clientId}/customers", requirePermission("customers", "read", customerHandler.List)).Methods("GET")
	apiV1.Handle("/clients/{clientId}/customers/search", requirePermission("customers", "read", customerHandler.Search)).Methods("GET")

	// Purchase Order routes (require purchase_orders permission)
	apiV1.Handle("/clients/{clientId}/purchase-orders", requirePermission("purchase_orders", "create", purchaseOrderHandler.Create)).Methods("POST")
	apiV1.Handle("/clients/{clientId}/purchase-orders/{id}", requirePermission("purchase_orders", "read", purchaseOrderHandler.Get)).Methods("GET")
	apiV1.Handle("/clients/{clientId}/purchase-orders/{id}", requirePermission("purchase_orders", "update", purchaseOrderHandler.Update)).Methods("PUT")
	apiV1.Handle("/clients/{clientId}/purchase-orders/{id}", requirePermission("purchase_orders", "delete", purchaseOrderHandler.Delete)).Methods("DELETE")
	apiV1.Handle("/clients/{clientId}/purchase-orders", requirePermission("purchase_orders", "read", purchaseOrderHandler.List)).Methods("GET")
	apiV1.Handle("/clients/{clientId}/purchase-orders/{id}/submit", requirePermission("purchase_orders", "update", purchaseOrderHandler.Submit)).Methods("POST")
	apiV1.Handle("/clients/{clientId}/purchase-orders/{id}/cancel", requirePermission("purchase_orders", "update", purchaseOrderHandler.Cancel)).Methods("POST")
	apiV1.Handle("/clients/{clientId}/purchase-orders/{id}/items", requirePermission("purchase_orders", "update", purchaseOrderHandler.AddItem)).Methods("POST")
	apiV1.Handle("/clients/{clientId}/purchase-orders/{id}/items/{itemId}", requirePermission("purchase_orders", "update", purchaseOrderHandler.RemoveItem)).Methods("DELETE")
	apiV1.Handle("/clients/{clientId}/purchase-orders/{id}/items", requirePermission("purchase_orders", "read", purchaseOrderHandler.ListItems)).Methods("GET")
	apiV1.Handle("/clients/{clientId}/purchase-orders/{id}/receive", requirePermission("purchase_orders", "update", purchaseOrderHandler.Receive)).Methods("POST")

	// Sales Order routes (require sales_orders permission)
	apiV1.Handle("/clients/{clientId}/sales-orders", requirePermission("sales_orders", "create", salesOrderHandler.Create)).Methods("POST")
	apiV1.Handle("/clients/{clientId}/sales-orders/{id}", requirePermission("sales_orders", "read", salesOrderHandler.Get)).Methods("GET")
	apiV1.Handle("/clients/{clientId}/sales-orders/{id}", requirePermission("sales_orders", "update", salesOrderHandler.Update)).Methods("PUT")
	apiV1.Handle("/clients/{clientId}/sales-orders/{id}", requirePermission("sales_orders", "delete", salesOrderHandler.Delete)).Methods("DELETE")
	apiV1.Handle("/clients/{clientId}/sales-orders", requirePermission("sales_orders", "read", salesOrderHandler.List)).Methods("GET")
	apiV1.Handle("/clients/{clientId}/sales-orders/{id}/confirm", requirePermission("sales_orders", "update", salesOrderHandler.Confirm)).Methods("POST")
	apiV1.Handle("/clients/{clientId}/sales-orders/{id}/cancel", requirePermission("sales_orders", "update", salesOrderHandler.Cancel)).Methods("POST")
	apiV1.Handle("/clients/{clientId}/sales-orders/{id}/items", requirePermission("sales_orders", "update", salesOrderHandler.AddItem)).Methods("POST")
	apiV1.Handle("/clients/{clientId}/sales-orders/{id}/items/{itemId}", requirePermission("sales_orders", "update", salesOrderHandler.RemoveItem)).Methods("DELETE")
	apiV1.Handle("/clients/{clientId}/sales-orders/{id}/items", requirePermission("sales_orders", "read", salesOrderHandler.ListItems)).Methods("GET")
	apiV1.Handle("/clients/{clientId}/sales-orders/{id}/fulfill", requirePermission("sales_orders", "update", salesOrderHandler.Fulfill)).Methods("POST")

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
