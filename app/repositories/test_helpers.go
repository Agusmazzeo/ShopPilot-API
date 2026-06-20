package repositories

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
)

// SetupTestDB creates a test database connection
func SetupTestDB(t *testing.T) *pgxpool.Pool {
	// Get test database URL from environment
	dbURL := os.Getenv("TEST_DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgresql://shoppilot:shoppilot@localhost:5434/shoppilot?sslmode=disable"
	}

	pool, err := pgxpool.New(context.Background(), dbURL)
	require.NoError(t, err, "Failed to connect to test database")

	// Verify connection
	err = pool.Ping(context.Background())
	require.NoError(t, err, "Failed to ping test database")

	return pool
}

// CleanupTestDB closes the database connection
func CleanupTestDB(t *testing.T, pool *pgxpool.Pool) {
	pool.Close()
}

// TruncateTable truncates a table for test cleanup
func TruncateTable(t *testing.T, pool *pgxpool.Pool, tableName string) {
	_, err := pool.Exec(context.Background(), fmt.Sprintf("TRUNCATE TABLE %s RESTART IDENTITY CASCADE", tableName))
	require.NoError(t, err, "Failed to truncate table %s", tableName)
}

// BeginTestTransaction starts a test transaction
func BeginTestTransaction(t *testing.T, pool *pgxpool.Pool) *pgxpool.Pool {
	// For simplicity, we'll use table truncation instead of transactions
	// In production tests, consider using testcontainers-go for isolated DB instances
	return pool
}
