package server

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/yourorg/shoppilot/app/config"
	"github.com/yourorg/shoppilot/app/redis"
	"github.com/yourorg/shoppilot/app/repositories"
)

// Server represents the HTTP server
type Server struct {
	cfg         *config.Config
	repoManager *repositories.RepositoryManager
	redisClient *redis.Client
	httpServer  *http.Server
}

// New creates a new server instance
func New(cfg *config.Config, repoManager *repositories.RepositoryManager, redisClient *redis.Client) *Server {
	return &Server{
		cfg:         cfg,
		repoManager: repoManager,
		redisClient: redisClient,
	}
}

// Start starts the HTTP server
func (s *Server) Start() error {
	// Create router with all routes and middleware
	router := s.setupRouter()

	// Create HTTP server
	addr := fmt.Sprintf("%s:%d", s.cfg.Server.Host, s.cfg.Server.Port)
	s.httpServer = &http.Server{
		Addr:         addr,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Channel to listen for errors coming from the listener
	serverErrors := make(chan error, 1)

	// Start the server
	go func() {
		log.Printf("Server listening on %s", addr)
		serverErrors <- s.httpServer.ListenAndServe()
	}()

	// Channel to listen for interrupt or terminate signals
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	// Blocking select waiting for either error or shutdown signal
	select {
	case err := <-serverErrors:
		return fmt.Errorf("server error: %w", err)

	case sig := <-shutdown:
		log.Printf("Received signal: %v. Starting graceful shutdown...", sig)

		// Give outstanding requests 5 seconds to complete
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := s.httpServer.Shutdown(ctx); err != nil {
			s.httpServer.Close()
			return fmt.Errorf("could not gracefully shutdown server: %w", err)
		}

		log.Println("Server stopped gracefully")
	}

	return nil
}
