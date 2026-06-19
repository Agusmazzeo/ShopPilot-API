package main

import (
	"fmt"
	"log"
	"os"

	"github.com/yourorg/shoppilot/app/config"
	_ "github.com/joho/godotenv/autoload"
)

func main() {
	fmt.Println("ShopPilot API starting...")

	// Load configuration
	cfg := config.Load()
	if err := cfg.Validate(); err != nil {
		log.Fatalf("Configuration validation failed: %v", err)
	}

	log.Printf("Server configured to run on %s:%d", cfg.Server.Host, cfg.Server.Port)
	log.Printf("Database: %s", cfg.Database.ConnectionString)
	log.Printf("Redis: %s:%d", cfg.Redis.Host, cfg.Redis.Port)

	// TODO: Initialize database
	// TODO: Initialize Redis
	// TODO: Initialize services
	// TODO: Start HTTP server

	os.Exit(0)
}
