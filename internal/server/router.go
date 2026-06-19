package server

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/rs/cors"
	"github.com/yourorg/shoppilot/internal/server/handlers"
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

	// API version 1
	_ = router.PathPrefix("/api/v1").Subrouter()

	// TODO: Add authenticated routes
	// auth := apiV1.PathPrefix("/auth").Subrouter()
	// auth.HandleFunc("/login", authHandler.Login).Methods("POST")
	// auth.HandleFunc("/register", authHandler.Register).Methods("POST")

	// TODO: Add protected routes
	// protected := apiV1.PathPrefix("").Subrouter()
	// protected.Use(s.authMiddleware)
	// protected.HandleFunc("/products", productHandler.List).Methods("GET")
	// protected.HandleFunc("/products", productHandler.Create).Methods("POST")

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
