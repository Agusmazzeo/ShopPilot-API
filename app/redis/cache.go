package redis

import (
	"context"
	"encoding/json"
	"time"
)

// Cache provides caching operations
type Cache struct {
	client *Client
}

// NewCache creates a new cache instance
func NewCache(client *Client) *Cache {
	return &Cache{client: client}
}

// Set stores a value in cache with expiration
func (c *Cache) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return c.client.rdb.Set(ctx, key, data, expiration).Err()
}

// Get retrieves a value from cache
func (c *Cache) Get(ctx context.Context, key string, dest interface{}) error {
	data, err := c.client.rdb.Get(ctx, key).Bytes()
	if err != nil {
		return err
	}
	return json.Unmarshal(data, dest)
}

// Delete removes a key from cache
func (c *Cache) Delete(ctx context.Context, key string) error {
	return c.client.rdb.Del(ctx, key).Err()
}

// Exists checks if a key exists
func (c *Cache) Exists(ctx context.Context, key string) (bool, error) {
	count, err := c.client.rdb.Exists(ctx, key).Result()
	return count > 0, err
}

// SetString stores a string value
func (c *Cache) SetString(ctx context.Context, key, value string, expiration time.Duration) error {
	return c.client.rdb.Set(ctx, key, value, expiration).Err()
}

// GetString retrieves a string value
func (c *Cache) GetString(ctx context.Context, key string) (string, error) {
	return c.client.rdb.Get(ctx, key).Result()
}
