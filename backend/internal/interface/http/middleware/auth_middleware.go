package middleware

import (
	"fmt"
	"incidex/internal/domain"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

type JWTMiddleware struct {
	secretKey []byte
}

func NewJWTMiddleware(secretKey string) *JWTMiddleware {
	return &JWTMiddleware{secretKey: []byte(secretKey)}
}

func (m *JWTMiddleware) Handle() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorization header is required"})
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token format"})
			return
		}

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return m.secretKey, nil
		})

		if err != nil || !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			return
		}

		if claims, ok := token.Claims.(jwt.MapClaims); ok {
			// user_id のバリデーションと抽出
			userIDFloat, ok := claims["user_id"].(float64)
			if !ok {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid user_id in token"})
				return
			}
			c.Set("userID", uint(userIDFloat))

			// role のバリデーションと抽出
			roleStr, ok := claims["role"].(string)
			if !ok {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Missing or invalid role in token"})
				return
			}

			// role の値が有効か検証します
			userRole := domain.Role(roleStr)
			switch userRole {
			case domain.RoleAdmin, domain.RoleEditor, domain.RoleViewer:
				c.Set("role", userRole)
			default:
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid role value in token"})
				return
			}
		}

		c.Next()
	}
}

// RequireRole はユーザーが必要な role のいずれかを持っているかチェックする middleware を返します
func RequireRole(allowedRoles ...domain.Role) gin.HandlerFunc {
	return func(c *gin.Context) {
		roleValue, exists := c.Get("role")
		if !exists {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Role not found in context"})
			return
		}

		userRole, ok := roleValue.(domain.Role)
		if !ok {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Invalid role format"})
			return
		}

		// ユーザーの role が許可された role に含まれているかチェックします
		for _, allowedRole := range allowedRoles {
			if userRole == allowedRole {
				c.Next()
				return
			}
		}

		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions"})
	}
}

// RequireAdmin は RequireRole(domain.RoleAdmin) の省略形です
func RequireAdmin() gin.HandlerFunc {
	return RequireRole(domain.RoleAdmin)
}

// RequireEditorOrAdmin はユーザーが editor または admin であることを要求します
func RequireEditorOrAdmin() gin.HandlerFunc {
	return RequireRole(domain.RoleAdmin, domain.RoleEditor)
}
