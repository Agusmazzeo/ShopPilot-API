package config

import (
	"fmt"
	"os"
	"strconv"
)

// Config holds all configuration for the application
type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Redis    RedisConfig
	JWT      JWTConfig
	Frontend FrontendConfig
	Storage  StorageConfig
}

// ServerConfig holds HTTP server configuration
type ServerConfig struct {
	Host string
	Port int
}

// DatabaseConfig holds PostgreSQL configuration
type DatabaseConfig struct {
	ConnectionString string
}

// RedisConfig holds Redis configuration
type RedisConfig struct {
	Host     string
	Port     int
	Password string
	DB       int
	UseTLS   bool
}

// JWTConfig holds JWT authentication configuration
type JWTConfig struct {
	Secret     string
	Expiration int // in hours
}

// FrontendConfig holds frontend application settings
type FrontendConfig struct {
	BaseURL string
}

// StorageConfig holds file storage configuration
type StorageConfig struct {
	Provider      string // "local" or "s3"
	LocalBasePath string
	S3Bucket      string
	S3Region      string
	S3AccessKey   string
	S3SecretKey   string
}

// Load reads configuration from environment variables
func Load() *Config {
	return &Config{
		Server: ServerConfig{
			Host: getEnv("SERVER_HOST", "0.0.0.0"),
			Port: getEnvAsInt("SERVER_PORT", 8080),
		},
		Database: DatabaseConfig{
			ConnectionString: getEnv("DATABASE_URL", ""),
		},
		Redis: RedisConfig{
			Host:     getEnv("REDIS_HOST", "localhost"),
			Port:     getEnvAsInt("REDIS_PORT", 6379),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       getEnvAsInt("REDIS_DB", 0),
			UseTLS:   getEnvAsBool("REDIS_USE_TLS", false),
		},
		JWT: JWTConfig{
			Secret:     getEnv("JWT_SECRET", "change-me-in-production"),
			Expiration: getEnvAsInt("JWT_EXPIRATION_HOURS", 24),
		},
		Frontend: FrontendConfig{
			BaseURL: getEnv("FRONTEND_URL", "http://localhost:5173"),
		},
		Storage: StorageConfig{
			Provider:      getEnv("STORAGE_PROVIDER", "local"),
			LocalBasePath: getEnv("STORAGE_LOCAL_PATH", "./uploads"),
			S3Bucket:      getEnv("STORAGE_S3_BUCKET", ""),
			S3Region:      getEnv("STORAGE_S3_REGION", "us-east-1"),
			S3AccessKey:   getEnv("STORAGE_S3_ACCESS_KEY", ""),
			S3SecretKey:   getEnv("STORAGE_S3_SECRET_KEY", ""),
		},
	}
}

// getEnv reads an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvAsInt reads an environment variable as integer or returns default
func getEnvAsInt(key string, defaultValue int) int {
	valueStr := getEnv(key, "")
	if valueStr == "" {
		return defaultValue
	}
	if value, err := strconv.Atoi(valueStr); err == nil {
		return value
	}
	return defaultValue
}

// getEnvAsBool reads an environment variable as boolean or returns default
func getEnvAsBool(key string, defaultValue bool) bool {
	valueStr := getEnv(key, "")
	if valueStr == "" {
		return defaultValue
	}
	if value, err := strconv.ParseBool(valueStr); err == nil {
		return value
	}
	return defaultValue
}

// Validate checks if all required configuration is present
func (c *Config) Validate() error {
	if c.Database.ConnectionString == "" {
		return fmt.Errorf("DATABASE_URL is required")
	}
	// In production, JWT_SECRET must be set to a secure value
	if c.JWT.Secret == "change-me-in-production" || c.JWT.Secret == "" {
		// Only allow default in development
		env := os.Getenv("APP_ENV")
		if env == "production" || env == "prod" {
			return fmt.Errorf("JWT_SECRET must be set to a secure value in production")
		}
		fmt.Println("WARNING: Using default JWT secret. Set JWT_SECRET in production!")
	}
	// Validate minimum secret length for security
	if len(c.JWT.Secret) < 32 {
		fmt.Println("WARNING: JWT_SECRET should be at least 32 characters for security")
	}
	return nil
}
