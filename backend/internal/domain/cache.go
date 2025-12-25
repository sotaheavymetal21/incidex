package domain

import (
	"context"
	"errors"
	"time"
)

var (
	// ErrNotFound is returned when a cache key is not found
	ErrNotFound = errors.New("cache key not found")
)

// CacheRepository defines the interface for caching operations
type CacheRepository interface {
	// Get retrieves a value from cache
	Get(ctx context.Context, key string) (string, error)

	// Set stores a value in cache with optional TTL (0 = no expiration)
	Set(ctx context.Context, key string, value string, ttl time.Duration) error

	// Delete removes a value from cache
	Delete(ctx context.Context, key string) error

	// DeleteByPattern removes all keys matching the pattern
	DeleteByPattern(ctx context.Context, pattern string) error

	// Exists checks if a key exists in cache
	Exists(ctx context.Context, key string) (bool, error)
}
