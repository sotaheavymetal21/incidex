package cache

import (
	"context"
	"incidex/internal/domain"
	"time"

	"github.com/redis/go-redis/v9"
)

type redisCache struct {
	client *redis.Client
}

// NewRedisCache creates a new Redis cache repository
// If client is nil, returns a no-op cache that always returns cache misses
func NewRedisCache(client *redis.Client) domain.CacheRepository {
	if client == nil {
		return &noOpCache{}
	}
	return &redisCache{client: client}
}

func (r *redisCache) Get(ctx context.Context, key string) (string, error) {
	val, err := r.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return "", domain.ErrNotFound
	}
	return val, err
}

func (r *redisCache) Set(ctx context.Context, key string, value string, ttl time.Duration) error {
	return r.client.Set(ctx, key, value, ttl).Err()
}

func (r *redisCache) Delete(ctx context.Context, key string) error {
	return r.client.Del(ctx, key).Err()
}

func (r *redisCache) DeleteByPattern(ctx context.Context, pattern string) error {
	iter := r.client.Scan(ctx, 0, pattern, 0).Iterator()
	for iter.Next(ctx) {
		if err := r.client.Del(ctx, iter.Val()).Err(); err != nil {
			return err
		}
	}
	return iter.Err()
}

func (r *redisCache) Exists(ctx context.Context, key string) (bool, error) {
	result, err := r.client.Exists(ctx, key).Result()
	return result > 0, err
}

// noOpCache is a no-op implementation that always returns cache misses
// Used when Redis is unavailable to ensure system continues to function
type noOpCache struct{}

func (n *noOpCache) Get(ctx context.Context, key string) (string, error) {
	return "", domain.ErrNotFound
}

func (n *noOpCache) Set(ctx context.Context, key string, value string, ttl time.Duration) error {
	return nil // No-op
}

func (n *noOpCache) Delete(ctx context.Context, key string) error {
	return nil // No-op
}

func (n *noOpCache) DeleteByPattern(ctx context.Context, pattern string) error {
	return nil // No-op
}

func (n *noOpCache) Exists(ctx context.Context, key string) (bool, error) {
	return false, nil
}
