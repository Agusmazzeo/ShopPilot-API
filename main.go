package main

import (
	"fmt"
	"log"

	"github.com/yourorg/shoppilot/app/config"
	"github.com/yourorg/shoppilot/app/redis"
	"github.com/yourorg/shoppilot/app/repositories"
	"github.com/yourorg/shoppilot/internal/server"
	_ "github.com/joho/godotenv/autoload"
)

func main() {
	fmt.Println("ShopPilot API starting...")

	// Load configuration
	cfg := config.Load()
	if err := cfg.Validate(); err != nil {
		log.Fatalf("Configuration validation failed: %v", err)
	}

	log.Printf("Configuration loaded successfully")

	// Initialize repository manager
	repoManager, err := repositories.NewRepositoryManager(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize repository manager: %v", err)
	}
	defer repoManager.Close()

	log.Printf("Database connection established successfully")

	// Initialize Redis client
	redisClient, err := redis.NewClient(&cfg.Redis)
	if err != nil {
		log.Fatalf("Failed to initialize Redis client: %v", err)
	}
	defer redisClient.Close()

	log.Printf("Redis connection established successfully")

	// Initialize and start server
	srv := server.New(cfg, repoManager, redisClient)

	log.Printf("Starting HTTP server on %s:%d", cfg.Server.Host, cfg.Server.Port)

	if err := srv.Start(); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
