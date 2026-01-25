package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRateLimitMiddleware_Handle(t *testing.T) {
	t.Parallel()

	t.Run("allows requests under limit", func(t *testing.T) {
		t.Parallel()

		// miniredis を起動します
		s, err := miniredis.Run()
		require.NoError(t, err)
		defer s.Close()

		redisClient := redis.NewClient(&redis.Options{
			Addr: s.Addr(),
		})

		middleware := NewRateLimitMiddleware(redisClient, 10, time.Minute)

		router := gin.New()
		router.Use(middleware.Handle())
		router.GET("/test", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		})

		// request を実行します
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.RemoteAddr = "192.168.1.1:12345"
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "10", rec.Header().Get("X-RateLimit-Limit"))
		assert.Equal(t, "9", rec.Header().Get("X-RateLimit-Remaining"))
	})

	t.Run("blocks requests over limit", func(t *testing.T) {
		t.Parallel()

		s, err := miniredis.Run()
		require.NoError(t, err)
		defer s.Close()

		redisClient := redis.NewClient(&redis.Options{
			Addr: s.Addr(),
		})

		middleware := NewRateLimitMiddleware(redisClient, 3, time.Minute)

		router := gin.New()
		router.Use(middleware.Handle())
		router.GET("/test", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		})

		// 3回の request を実行します（制限内）
		for i := 0; i < 3; i++ {
			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			req.RemoteAddr = "192.168.1.2:12345"
			rec := httptest.NewRecorder()

			router.ServeHTTP(rec, req)
			assert.Equal(t, http.StatusOK, rec.Code)
		}

		// 4回目の request はブロックされるべきです
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.RemoteAddr = "192.168.1.2:12345"
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusTooManyRequests, rec.Code)
		assert.Equal(t, "0", rec.Header().Get("X-RateLimit-Remaining"))
		assert.Contains(t, rec.Body.String(), "Rate limit exceeded")
	})

	t.Run("different IPs have separate limits", func(t *testing.T) {
		t.Parallel()

		s, err := miniredis.Run()
		require.NoError(t, err)
		defer s.Close()

		redisClient := redis.NewClient(&redis.Options{
			Addr: s.Addr(),
		})

		middleware := NewRateLimitMiddleware(redisClient, 2, time.Minute)

		router := gin.New()
		router.Use(middleware.Handle())
		router.GET("/test", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		})

		// IP 1 の制限を使い切ります
		for i := 0; i < 2; i++ {
			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			req.RemoteAddr = "192.168.1.10:12345"
			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)
		}

		// IP 1 はブロックされるべきです
		req1 := httptest.NewRequest(http.MethodGet, "/test", nil)
		req1.RemoteAddr = "192.168.1.10:12345"
		rec1 := httptest.NewRecorder()
		router.ServeHTTP(rec1, req1)
		assert.Equal(t, http.StatusTooManyRequests, rec1.Code)

		// IP 2 はまだ許可されるべきです
		req2 := httptest.NewRequest(http.MethodGet, "/test", nil)
		req2.RemoteAddr = "192.168.1.20:12345"
		rec2 := httptest.NewRecorder()
		router.ServeHTTP(rec2, req2)
		assert.Equal(t, http.StatusOK, rec2.Code)
	})

	t.Run("different paths have separate limits", func(t *testing.T) {
		t.Parallel()

		s, err := miniredis.Run()
		require.NoError(t, err)
		defer s.Close()

		redisClient := redis.NewClient(&redis.Options{
			Addr: s.Addr(),
		})

		middleware := NewRateLimitMiddleware(redisClient, 1, time.Minute)

		router := gin.New()
		router.Use(middleware.Handle())
		router.GET("/path1", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		})
		router.GET("/path2", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		})

		// /path1 で制限を使用します
		req1 := httptest.NewRequest(http.MethodGet, "/path1", nil)
		req1.RemoteAddr = "192.168.1.30:12345"
		rec1 := httptest.NewRecorder()
		router.ServeHTTP(rec1, req1)
		assert.Equal(t, http.StatusOK, rec1.Code)

		// /path1 への2回目の request はブロックされるべきです
		req1b := httptest.NewRequest(http.MethodGet, "/path1", nil)
		req1b.RemoteAddr = "192.168.1.30:12345"
		rec1b := httptest.NewRecorder()
		router.ServeHTTP(rec1b, req1b)
		assert.Equal(t, http.StatusTooManyRequests, rec1b.Code)

		// /path2 はまだ許可されるべきです
		req2 := httptest.NewRequest(http.MethodGet, "/path2", nil)
		req2.RemoteAddr = "192.168.1.30:12345"
		rec2 := httptest.NewRecorder()
		router.ServeHTTP(rec2, req2)
		assert.Equal(t, http.StatusOK, rec2.Code)
	})

	t.Run("sets retry-after header", func(t *testing.T) {
		t.Parallel()

		s, err := miniredis.Run()
		require.NoError(t, err)
		defer s.Close()

		redisClient := redis.NewClient(&redis.Options{
			Addr: s.Addr(),
		})

		middleware := NewRateLimitMiddleware(redisClient, 1, time.Minute)

		router := gin.New()
		router.Use(middleware.Handle())
		router.GET("/test", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		})

		// 制限を使用します
		req1 := httptest.NewRequest(http.MethodGet, "/test", nil)
		req1.RemoteAddr = "192.168.1.40:12345"
		rec1 := httptest.NewRecorder()
		router.ServeHTTP(rec1, req1)

		// レート制限をトリガーします
		req2 := httptest.NewRequest(http.MethodGet, "/test", nil)
		req2.RemoteAddr = "192.168.1.40:12345"
		rec2 := httptest.NewRecorder()
		router.ServeHTTP(rec2, req2)

		assert.Equal(t, http.StatusTooManyRequests, rec2.Code)
		assert.NotEmpty(t, rec2.Header().Get("Retry-After"))
	})
}

func TestRateLimitMiddleware_FallbackLimit(t *testing.T) {
	t.Parallel()

	t.Run("uses fallback when redis is unavailable", func(t *testing.T) {
		t.Parallel()

		// Redis がダウンしている状態をシミュレートするため無効なアドレスでクライアントを作成します
		redisClient := redis.NewClient(&redis.Options{
			Addr: "invalid:6379",
		})

		middleware := NewRateLimitMiddleware(redisClient, 2, time.Minute)

		router := gin.New()
		router.Use(middleware.Handle())
		router.GET("/test", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		})

		// 最初の request はフォールバックを使用してパスするべきです
		req1 := httptest.NewRequest(http.MethodGet, "/test", nil)
		req1.RemoteAddr = "192.168.1.50:12345"
		rec1 := httptest.NewRecorder()
		router.ServeHTTP(rec1, req1)
		assert.Equal(t, http.StatusOK, rec1.Code)

		// 2回目の request もパスするべきです（フォールバック制限内）
		req2 := httptest.NewRequest(http.MethodGet, "/test", nil)
		req2.RemoteAddr = "192.168.1.50:12345"
		rec2 := httptest.NewRecorder()
		router.ServeHTTP(rec2, req2)
		assert.Equal(t, http.StatusOK, rec2.Code)

		// 3回目の request はフォールバックによってブロックされるべきです
		req3 := httptest.NewRequest(http.MethodGet, "/test", nil)
		req3.RemoteAddr = "192.168.1.50:12345"
		rec3 := httptest.NewRecorder()
		router.ServeHTTP(rec3, req3)
		assert.Equal(t, http.StatusTooManyRequests, rec3.Code)
	})
}

func TestCheckFallbackLimit(t *testing.T) {
	t.Parallel()

	t.Run("allows requests under fallback limit", func(t *testing.T) {
		t.Parallel()

		middleware := &RateLimitMiddleware{
			limit:          3,
			window:         time.Minute,
			fallbackCounts: make(map[string]*fallbackEntry),
		}

		// 最初の request はパスするべきです
		assert.True(t, middleware.checkFallbackLimit("test-key-1"))
		// 2回目の request はパスするべきです
		assert.True(t, middleware.checkFallbackLimit("test-key-1"))
		// 3回目の request はパスするべきです
		assert.True(t, middleware.checkFallbackLimit("test-key-1"))
		// 4回目の request は失敗するべきです
		assert.False(t, middleware.checkFallbackLimit("test-key-1"))
	})

	t.Run("resets after window expires", func(t *testing.T) {
		t.Parallel()

		middleware := &RateLimitMiddleware{
			limit:          1,
			window:         100 * time.Millisecond,
			fallbackCounts: make(map[string]*fallbackEntry),
		}

		// 最初の request はパスします
		assert.True(t, middleware.checkFallbackLimit("test-key-2"))
		// 2回目の request は失敗します（制限超過）
		assert.False(t, middleware.checkFallbackLimit("test-key-2"))

		// ウィンドウが期限切れになるまで待機します
		time.Sleep(150 * time.Millisecond)

		// 再び許可されるべきです
		assert.True(t, middleware.checkFallbackLimit("test-key-2"))
	})

	t.Run("different keys have separate limits", func(t *testing.T) {
		t.Parallel()

		middleware := &RateLimitMiddleware{
			limit:          1,
			window:         time.Minute,
			fallbackCounts: make(map[string]*fallbackEntry),
		}

		// キー 1 は制限を使用します
		assert.True(t, middleware.checkFallbackLimit("key-a"))
		assert.False(t, middleware.checkFallbackLimit("key-a"))

		// キー 2 はまだ許可されるべきです
		assert.True(t, middleware.checkFallbackLimit("key-b"))
	})
}
