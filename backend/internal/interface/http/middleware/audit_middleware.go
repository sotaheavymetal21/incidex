package middleware

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"incidex/internal/domain"
	"incidex/internal/pkg/logger"
	"incidex/internal/pkg/sanitizer"
)

type AuditMiddleware struct {
	auditLogRepo domain.AuditLogRepository
	userRepo     domain.UserRepository
}

func NewAuditMiddleware(auditLogRepo domain.AuditLogRepository, userRepo domain.UserRepository) *AuditMiddleware {
	return &AuditMiddleware{
		auditLogRepo: auditLogRepo,
		userRepo:     userRepo,
	}
}

// Log は監査目的で API 呼び出しを記録する middleware です
func (m *AuditMiddleware) Log() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 特定のパスとメソッドでは監査ログをスキップします
		if shouldSkipAudit(c.Request.URL.Path, c.Request.Method) {
			c.Next()
			return
		}

		// POST/PUT/DELETE の request body を詳細用にキャプチャします
		var requestBody string
		if c.Request.Method == "POST" || c.Request.Method == "PUT" || c.Request.Method == "DELETE" {
			bodyBytes, _ := io.ReadAll(c.Request.Body)
			c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
			requestBody = string(bodyBytes)
			// 機密データ（パスワード、token など）をサニタイズします
			requestBody = sanitizer.SanitizeJSON(requestBody)
		}

		// request を処理します
		c.Next()

		// request 処理後に監査ログを作成します
		go func() {
			log := &domain.AuditLog{
				Method:     c.Request.Method,
				Path:       c.Request.URL.Path,
				IPAddress:  sanitizer.SanitizeIP(c.ClientIP()), // GDPR 準拠のため IP をマスクします
				UserAgent:  c.Request.UserAgent(),
				StatusCode: c.Writer.Status(),
				CreatedAt:  time.Now(),
			}

			// context からユーザー情報を取得します（利用可能な場合）
			if userIDVal, exists := c.Get("userID"); exists {
				if userID, ok := userIDVal.(uint); ok {
					log.UserID = &userID

					// ユーザー詳細を取得します（goroutine 内なので background context を使用します）
					ctx := context.Background()
					user, err := m.userRepo.FindByID(ctx, userID)
					if err == nil && user != nil {
						log.UserName = user.Name
						log.UserEmail = user.Email
					}
				}
			}

			// アクションとリソースを判定します
			action, resourceType, resourceID := determineActionAndResource(c)
			log.Action = action
			log.ResourceType = resourceType
			log.ResourceID = resourceID

			// 詳細を追加します
			details := make(map[string]interface{})
			if requestBody != "" && len(requestBody) < 1000 {
				details["request_body"] = requestBody
			}
			if len(details) > 0 {
				detailsJSON, _ := json.Marshal(details)
				log.Details = string(detailsJSON)
			}

			// 監査ログを保存します
			// goroutine 内なので background context を使用します
			ctx := context.Background()
			if err := m.auditLogRepo.Create(ctx, log); err != nil {
				logger.Log.Error("Failed to create audit log",
					zap.Error(err),
					zap.String("method", log.Method),
					zap.String("path", log.Path),
				)
			}
		}()
	}
}

func shouldSkipAudit(path string, method string) bool {
	// GET と OPTIONS request をスキップします
	// GET: 読み取り専用操作は監査不要です
	// OPTIONS: CORS プリフライト request はブラウザ自動化であり、ユーザーアクションではありません
	if method == "GET" || method == "OPTIONS" {
		return true
	}

	// メソッドに関係なく特定のエンドポイントをスキップします
	skipPaths := []string{
		"/api/health",
		"/api/stats",      // 統計情報は読み取り専用です
		"/api/export",     // エクスポート操作は読み取り専用です
		"/api/audit-logs", // 監査ログのクエリは監査しません
	}

	for _, skip := range skipPaths {
		if strings.HasPrefix(path, skip) {
			return true
		}
	}

	return false
}

func determineActionAndResource(c *gin.Context) (domain.AuditAction, string, *uint) {
	method := c.Request.Method
	path := c.Request.URL.Path

	var action domain.AuditAction
	var resourceType string
	var resourceID *uint

	// HTTP メソッドとパスに基づいてアクションを判定します
	switch method {
	case "POST":
		if strings.Contains(path, "/login") {
			action = domain.AuditActionLogin
		} else if strings.Contains(path, "/register") {
			action = domain.AuditActionCreate
			resourceType = "auth"
		} else if strings.Contains(path, "/summarize") {
			action = domain.AuditActionUpdate
		} else if strings.Contains(path, "/assign") {
			action = domain.AuditActionUpdate
		} else if strings.Contains(path, "/publish") {
			action = domain.AuditActionUpdate
		} else if strings.Contains(path, "/unpublish") {
			action = domain.AuditActionUpdate
		} else if strings.Contains(path, "/ai-suggestion") {
			action = domain.AuditActionCreate
		} else {
			action = domain.AuditActionCreate
		}
	case "GET":
		action = domain.AuditActionRead
	case "PUT", "PATCH":
		action = domain.AuditActionUpdate
	case "DELETE":
		action = domain.AuditActionDelete
	}

	// パスからリソースタイプを判定します（より具体的なパスを先にチェックします）
	if strings.Contains(path, "/attachments") {
		resourceType = "attachment"
	} else if strings.Contains(path, "/comments") {
		resourceType = "comment"
	} else if strings.Contains(path, "/timeline") {
		resourceType = "timeline"
	} else if strings.Contains(path, "/activities") {
		resourceType = "activity"
	} else if strings.Contains(path, "/post-mortems") {
		resourceType = "post_mortem"
	} else if strings.Contains(path, "/action-items") {
		resourceType = "action_item"
	} else if strings.Contains(path, "/incidents") {
		resourceType = "incident"
	} else if strings.Contains(path, "/users") {
		resourceType = "user"
	} else if strings.Contains(path, "/tags") {
		resourceType = "tag"
	} else if strings.Contains(path, "/templates") {
		resourceType = "template"
	} else if strings.Contains(path, "/notifications") {
		resourceType = "notification"
	} else if strings.Contains(path, "/reports") {
		resourceType = "report"
	} else if strings.Contains(path, "/stats") {
		resourceType = "stats"
	} else if strings.Contains(path, "/export") {
		resourceType = "export"
	} else if strings.Contains(path, "/audit-logs") {
		resourceType = "audit_log"
	} else if strings.Contains(path, "/auth") {
		resourceType = "auth"
	}

	// パスパラメータからリソース ID を抽出しようとします
	if idParam := c.Param("id"); idParam != "" {
		// 可能であれば ID をパースします
		if parsedID, err := strconv.ParseUint(idParam, 10, 32); err == nil {
			id := uint(parsedID)
			resourceID = &id
		}
	}

	// ネストされたルートの場合、インシデント ID または他の親 ID を取得しようとします
	if resourceID == nil {
		// パスセグメント内のインシデント ID をチェックします
		pathParts := strings.Split(path, "/")
		for i, part := range pathParts {
			if part == "incidents" && i+1 < len(pathParts) {
				if parsedID, err := strconv.ParseUint(pathParts[i+1], 10, 32); err == nil {
					id := uint(parsedID)
					resourceID = &id
					break
				}
			}
		}
	}

	return action, resourceType, resourceID
}

// 注意: sanitizeSensitiveData 関数は削除されました
// 包括的なサニタイズには sanitizer.SanitizeJSON を使用してください
