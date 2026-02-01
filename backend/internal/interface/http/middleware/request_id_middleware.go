package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// RequestIDKey はコンテキストに格納するRequest IDのキーです
const RequestIDKey = "X-Request-ID"

// RequestID はすべてのリクエストに一意のRequest IDを付与するミドルウェアです
// クライアントからX-Request-IDヘッダーが提供された場合はそれを使用し、
// なければ新しいUUIDを生成します
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader(RequestIDKey)
		if requestID == "" {
			requestID = uuid.New().String()
		}

		// コンテキストに保存
		c.Set(RequestIDKey, requestID)

		// レスポンスヘッダーに追加
		c.Header(RequestIDKey, requestID)

		c.Next()
	}
}

// GetRequestID はコンテキストからRequest IDを取得します
func GetRequestID(c *gin.Context) string {
	if requestID, exists := c.Get(RequestIDKey); exists {
		if id, ok := requestID.(string); ok {
			return id
		}
	}
	return ""
}
