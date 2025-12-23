package middleware

import (
	"bytes"
	"encoding/json"
	"incidex/internal/domain"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
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

// AuditLog middleware records API calls for auditing purposes
func (m *AuditMiddleware) Log() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip audit logging for certain paths
		if shouldSkipAudit(c.Request.URL.Path) {
			c.Next()
			return
		}

		// Capture request body for POST/PUT/DELETE (for details)
		var requestBody string
		if c.Request.Method == "POST" || c.Request.Method == "PUT" || c.Request.Method == "DELETE" {
			bodyBytes, _ := io.ReadAll(c.Request.Body)
			c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
			requestBody = string(bodyBytes)
			// Sanitize sensitive data (passwords, tokens, etc.)
			requestBody = sanitizeSensitiveData(requestBody)
		}

		// Process request
		c.Next()

		// Create audit log after request is processed
		go func() {
			log := &domain.AuditLog{
				Method:     c.Request.Method,
				Path:       c.Request.URL.Path,
				IPAddress:  c.ClientIP(),
				UserAgent:  c.Request.UserAgent(),
				StatusCode: c.Writer.Status(),
				CreatedAt:  time.Now(),
			}

			// Get user info from context (if available)
			if userIDVal, exists := c.Get("userID"); exists {
				if userID, ok := userIDVal.(float64); ok {
					uid := uint(userID)
					log.UserID = &uid

					// Fetch user details
					user, err := m.userRepo.FindByID(c.Request.Context(), uid)
					if err == nil && user != nil {
						log.UserName = user.Name
						log.UserEmail = user.Email
					}
				}
			}

			// Determine action and resource
			action, resourceType, resourceID := determineActionAndResource(c)
			log.Action = action
			log.ResourceType = resourceType
			log.ResourceID = resourceID

			// Add details
			details := make(map[string]interface{})
			if requestBody != "" && len(requestBody) < 1000 {
				details["request_body"] = requestBody
			}
			if len(details) > 0 {
				detailsJSON, _ := json.Marshal(details)
				log.Details = string(detailsJSON)
			}

			// Save audit log (ignore errors to not affect main request)
			_ = m.auditLogRepo.Create(c.Request.Context(), log)
		}()
	}
}

func shouldSkipAudit(path string) bool {
	skipPaths := []string{
		"/api/health",
		"/api/protected",
		"/api/stats/dashboard",
		"/api/stats/sla",
	}

	for _, skip := range skipPaths {
		if path == skip {
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

	// Determine action based on HTTP method
	switch method {
	case "POST":
		if strings.Contains(path, "/login") {
			action = domain.AuditActionLogin
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

	// Determine resource type from path
	if strings.Contains(path, "/incidents") {
		resourceType = "incident"
	} else if strings.Contains(path, "/users") {
		resourceType = "user"
	} else if strings.Contains(path, "/tags") {
		resourceType = "tag"
	} else if strings.Contains(path, "/templates") {
		resourceType = "template"
	} else if strings.Contains(path, "/post-mortems") {
		resourceType = "post_mortem"
	} else if strings.Contains(path, "/action-items") {
		resourceType = "action_item"
	} else if strings.Contains(path, "/auth") {
		resourceType = "auth"
	}

	// Try to extract resource ID from path parameter
	if idParam := c.Param("id"); idParam != "" {
		// Parse ID if possible
		if parsedID, err := strconv.ParseUint(idParam, 10, 32); err == nil {
			id := uint(parsedID)
			resourceID = &id
		}
	}

	return action, resourceType, resourceID
}

func sanitizeSensitiveData(body string) string {
	// Remove password fields
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(body), &data); err != nil {
		return body
	}

	// Sanitize passwords
	if _, exists := data["password"]; exists {
		data["password"] = "***REDACTED***"
	}
	if _, exists := data["old_password"]; exists {
		data["old_password"] = "***REDACTED***"
	}
	if _, exists := data["new_password"]; exists {
		data["new_password"] = "***REDACTED***"
	}

	sanitized, _ := json.Marshal(data)
	return string(sanitized)
}
