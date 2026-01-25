package db

import (
	"context"
	"log"

	"github.com/redis/go-redis/v9"
)

// ConnectRedis は Redis クライアントを初期化します
func ConnectRedis(redisURL string) *redis.Client {
	opts, err := redis.ParseURL(redisURL)
	if err != nil {
		log.Fatalf("Failed to parse Redis URL: %v", err)
	}

	client := redis.NewClient(opts)

	// 接続をテストします
	ctx := context.Background()
	if err := client.Ping(ctx).Err(); err != nil {
		log.Printf("Warning: Failed to connect to Redis: %v", err)
		log.Printf("Redis caching will be disabled")
		return nil
	}

	log.Println("Successfully connected to Redis")
	return client
}
