package main

import (
	"fmt"
	"log"
	"os"

	"github.com/yourorg/shoppilot/app/config"
	"github.com/yourorg/shoppilot/app/repositories"
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

	// TODO: Initialize Redis
	// TODO: Initialize services
	// TODO: Start HTTP server

	os.Exit(0)
}
