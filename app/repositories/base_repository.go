package repositories

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// BaseRepository provides common database operations
type BaseRepository struct {
	pool *pgxpool.Pool
}

// NewBaseRepository creates a new base repository
func NewBaseRepository(pool *pgxpool.Pool) *BaseRepository {
	return &BaseRepository{pool: pool}
}

// GetPool returns the connection pool
func (r *BaseRepository) GetPool() *pgxpool.Pool {
	return r.pool
}

// WithTransaction executes a function within a database transaction
func (r *BaseRepository) WithTransaction(ctx context.Context, fn func(pgx.Tx) error) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}

	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback(ctx)
			panic(p)
		} else if err != nil {
			_ = tx.Rollback(ctx)
		} else {
			err = tx.Commit(ctx)
		}
	}()

	err = fn(tx)
	return err
}

// Common query helpers

// GetCurrentTimestamp returns the current timestamp for database operations
func GetCurrentTimestamp() time.Time {
	return time.Now().UTC()
}

// NullableString converts a string pointer to a nullable string value
func NullableString(s *string) interface{} {
	if s == nil {
		return nil
	}
	return *s
}

// NullableInt converts an int pointer to a nullable int value
func NullableInt(i *int) interface{} {
	if i == nil {
		return nil
	}
	return *i
}
