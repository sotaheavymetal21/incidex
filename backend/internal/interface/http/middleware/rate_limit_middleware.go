package middleware

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

// RateLimitMiddleware は Redis を使用してレート制限を実装します
type RateLimitMiddleware struct {
	redisClient *redis.Client
	limit       int           // 最大 request 数
	window      time.Duration // 時間ウィンドウ
	// Redis が利用できない場合のフォールバック用インメモリリミッター
	fallbackMu     sync.Mutex
	fallbackCounts map[string]*fallbackEntry
}

type fallbackEntry struct {
	count     int
	expiresAt time.Time
}

// NewRateLimitMiddleware は新しいレート制限 middleware を作成します
func NewRateLimitMiddleware(redisClient *redis.Client, limit int, window time.Duration) *RateLimitMiddleware {
	return &RateLimitMiddleware{
		redisClient:    redisClient,
		limit:          limit,
		window:         window,
		fallbackCounts: make(map[string]*fallbackEntry),
	}
}

// Handle はレート制限用の Gin middleware handler を返します
func (m *RateLimitMiddleware) Handle() gin.HandlerFunc {
	return func(c *gin.Context) {
		// クライアント識別子（IP アドレス）を取得します
		clientIP := c.ClientIP()

		// このクライアント用の Redis キーを作成します
		key := fmt.Sprintf("rate_limit:%s:%s", c.Request.URL.Path, clientIP)

		ctx := c.Request.Context()

		// カウンターをインクリメントします
		count, err := m.redisClient.Incr(ctx, key).Result()
		if err != nil {
			// Redis がダウンしている場合、フォールバック用のインメモリリミッターを使用します（フェイルセーフ）
			if !m.checkFallbackLimit(key) {
				c.Header("Retry-After", "60")
				c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
					"error": "Rate limit exceeded. Please try again later.",
				})
				return
			}
			c.Next()
			return
		}

		// 最初の request で有効期限を設定します
		if count == 1 {
			m.redisClient.Expire(ctx, key, m.window)
		}

		// 制限を超過したかチェックします
		if count > int64(m.limit) {
			// クライアントにリトライ可能なタイミングを通知するため TTL を取得します
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

		// レート制限ヘッダーを設定します
		remaining := m.limit - int(count)
		c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", m.limit))
		c.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", remaining))

		c.Next()
	}
}

// checkFallbackLimit は Redis が利用できない場合にインメモリのレート制限をチェックします
func (m *RateLimitMiddleware) checkFallbackLimit(key string) bool {
	m.fallbackMu.Lock()
	defer m.fallbackMu.Unlock()

	now := time.Now()

	// 期限切れのエントリをクリーンアップします
	for k, entry := range m.fallbackCounts {
		if now.After(entry.expiresAt) {
			delete(m.fallbackCounts, k)
		}
	}

	entry, exists := m.fallbackCounts[key]
	if !exists {
		m.fallbackCounts[key] = &fallbackEntry{
			count:     1,
			expiresAt: now.Add(m.window),
		}
		return true
	}

	if now.After(entry.expiresAt) {
		entry.count = 1
		entry.expiresAt = now.Add(m.window)
		return true
	}

	entry.count++
	return entry.count <= m.limit
}
