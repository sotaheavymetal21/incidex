package middleware

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

// RateLimitMiddleware implements rate limiting using Redis
type RateLimitMiddleware struct {
	redisClient *redis.Client
	limit       int           // max requests
	window      time.Duration // time window
}

// NewRateLimitMiddleware creates a new rate limit middleware
func NewRateLimitMiddleware(redisClient *redis.Client, limit int, window time.Duration) *RateLimitMiddleware {
	return &RateLimitMiddleware{
		redisClient: redisClient,
		limit:       limit,
		window:      window,
	}
}

// Handle returns a Gin middleware handler for rate limiting
func (m *RateLimitMiddleware) Handle() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get client identifier (IP address)
		clientIP := c.ClientIP()

		// Create Redis key for this client
		key := fmt.Sprintf("rate_limit:%s:%s", c.Request.URL.Path, clientIP)

		ctx := c.Request.Context()

		// Increment counter
		count, err := m.redisClient.Incr(ctx, key).Result()
		if err != nil {
			// If Redis is down, allow the request (fail open)
			c.Next()
			return
		}

		// Set expiration on first request
		if count == 1 {
			m.redisClient.Expire(ctx, key, m.window)
		}

		// Check if limit exceeded
		if count > int64(m.limit) {
			// Get TTL to inform client when they can retry
			ttl, _ := m.redisClient.TTL(ctx, key).Result()

			c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", m.limit))
			c.Header("X-RateLimit-Remaining", "0")
			c.Header("X-RateLimit-Reset", fmt.Sprintf("%d", time.Now().Add(ttl).Unix()))
			c.Header("Retry-After", fmt.Sprintf("%d", int(ttl.Seconds())))

			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error": "Rate limit exceeded. Please try again later.",
			})
			return
		}

		// Set rate limit headers
		remaining := m.limit - int(count)
		c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", m.limit))
		c.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", remaining))

		c.Next()
	}
}
