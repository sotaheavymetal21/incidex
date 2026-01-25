package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// HealthHandler はヘルスチェック関連の HTTP handler を提供します
type HealthHandler struct {
	db *gorm.DB
}

// NewHealthHandler は新しい HealthHandler を作成します
func NewHealthHandler(db *gorm.DB) *HealthHandler {
	return &HealthHandler{db: db}
}

// Liveness はアプリケーションが動作しているかを返します
func (h *HealthHandler) Liveness(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
	})
}

// Readiness はデータベース接続を含む依存サービスの状態を返します
func (h *HealthHandler) Readiness(c *gin.Context) {
	sqlDB, err := h.db.DB()
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status":   "error",
			"database": "unavailable",
			"error":    "failed to get database connection",
		})
		return
	}

	if err := sqlDB.Ping(); err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status":   "error",
			"database": "unavailable",
			"error":    "database ping failed",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":   "ok",
		"database": "connected",
	})
}
