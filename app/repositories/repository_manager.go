package repositories

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/yourorg/shoppilot/app/config"
)

// RepositoryManager manages all repositories and database connection
type RepositoryManager struct {
	pool *pgxpool.Pool

	PlatformUsers PlatformUserRepository
	Clients       ClientRepository
	ClientUsers   ClientUserRepository
	Shops         ShopRepository
	Products      ProductRepository
}

// NewRepositoryManager creates a new repository manager with database connection
func NewRepositoryManager(cfg *config.Config) (*RepositoryManager, error) {
	// Create connection pool
	pool, err := pgxpool.New(context.Background(), cfg.Database.ConnectionString)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	// Verify connection
	if err := pool.Ping(context.Background()); err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	rm := &RepositoryManager{
		pool:          pool,
		PlatformUsers: NewPlatformUserRepository(pool),
		Clients:       NewClientRepository(pool),
		ClientUsers:   NewClientUserRepository(pool),
		Shops:         NewShopRepository(pool),
		Products:      NewProductRepository(pool),
	}

	return rm, nil
}

// GetPool returns the underlying connection pool
func (rm *RepositoryManager) GetPool() *pgxpool.Pool {
	return rm.pool
}

// Close closes all database connections
func (rm *RepositoryManager) Close() {
	if rm.pool != nil {
		rm.pool.Close()
	}
}

// Health checks if the database connection is healthy
func (rm *RepositoryManager) Health(ctx context.Context) error {
	return rm.pool.Ping(ctx)
}
