package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestRequestID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("generates new request ID when not provided", func(t *testing.T) {
		r := gin.New()
		r.Use(RequestID())
		r.GET("/test", func(c *gin.Context) {
			requestID := GetRequestID(c)
			assert.NotEmpty(t, requestID)
			c.String(http.StatusOK, "ok")
		})

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.NotEmpty(t, w.Header().Get(RequestIDKey))
	})

	t.Run("uses provided request ID from header", func(t *testing.T) {
		r := gin.New()
		r.Use(RequestID())
		r.GET("/test", func(c *gin.Context) {
			requestID := GetRequestID(c)
			assert.Equal(t, "custom-request-id-123", requestID)
			c.String(http.StatusOK, "ok")
		})

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set(RequestIDKey, "custom-request-id-123")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "custom-request-id-123", w.Header().Get(RequestIDKey))
	})

	t.Run("request ID is UUID format when generated", func(t *testing.T) {
		r := gin.New()
		r.Use(RequestID())
		r.GET("/test", func(c *gin.Context) {
			c.String(http.StatusOK, "ok")
		})

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		requestID := w.Header().Get(RequestIDKey)
		// UUID形式: 8-4-4-4-12
		assert.Regexp(t, `^[a-f0-9]{8}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{12}$`, requestID)
	})
}

func TestGetRequestID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("returns empty string when request ID not set", func(t *testing.T) {
		r := gin.New()
		r.GET("/test", func(c *gin.Context) {
			requestID := GetRequestID(c)
			assert.Empty(t, requestID)
			c.String(http.StatusOK, "ok")
		})

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})
}
