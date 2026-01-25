package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"incidex/internal/domain"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func init() {
	gin.SetMode(gin.TestMode)
}

const testSecretKey = "test-secret-key-for-testing-12345"

func createTestToken(userID uint, role string, expiresAt time.Time, secret string) string {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": float64(userID),
		"role":    role,
		"exp":     expiresAt.Unix(),
	})
	tokenString, _ := token.SignedString([]byte(secret))
	return tokenString
}

func TestJWTMiddleware_Handle(t *testing.T) {
	t.Parallel()

	middleware := NewJWTMiddleware(testSecretKey)

	t.Run("valid token sets context", func(t *testing.T) {
		t.Parallel()

		router := gin.New()
		router.Use(middleware.Handle())
		router.GET("/test", func(c *gin.Context) {
			userID, _ := c.Get("userID")
			role, _ := c.Get("role")
			c.JSON(http.StatusOK, gin.H{"user_id": userID, "role": role})
		})

		validToken := createTestToken(1, "admin", time.Now().Add(time.Hour), testSecretKey)

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("Authorization", "Bearer "+validToken)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("missing authorization header returns 401", func(t *testing.T) {
		t.Parallel()

		router := gin.New()
		router.Use(middleware.Handle())
		router.GET("/test", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		})

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusUnauthorized, rec.Code)
		assert.Contains(t, rec.Body.String(), "Authorization header is required")
	})

	t.Run("non-bearer token returns 401", func(t *testing.T) {
		t.Parallel()

		router := gin.New()
		router.Use(middleware.Handle())
		router.GET("/test", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		})

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("Authorization", "Basic some-token")
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusUnauthorized, rec.Code)
		assert.Contains(t, rec.Body.String(), "Invalid token format")
	})

	t.Run("invalid signature returns 401", func(t *testing.T) {
		t.Parallel()

		router := gin.New()
		router.Use(middleware.Handle())
		router.GET("/test", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		})

		invalidToken := createTestToken(1, "admin", time.Now().Add(time.Hour), "wrong-secret-key")

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("Authorization", "Bearer "+invalidToken)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusUnauthorized, rec.Code)
		assert.Contains(t, rec.Body.String(), "Invalid token")
	})

	t.Run("expired token returns 401", func(t *testing.T) {
		t.Parallel()

		router := gin.New()
		router.Use(middleware.Handle())
		router.GET("/test", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		})

		expiredToken := createTestToken(1, "admin", time.Now().Add(-time.Hour), testSecretKey)

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("Authorization", "Bearer "+expiredToken)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusUnauthorized, rec.Code)
		assert.Contains(t, rec.Body.String(), "Invalid token")
	})

	t.Run("malformed token returns 401", func(t *testing.T) {
		t.Parallel()

		router := gin.New()
		router.Use(middleware.Handle())
		router.GET("/test", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		})

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("Authorization", "Bearer not-a-valid-jwt-token")
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusUnauthorized, rec.Code)
	})

	t.Run("token without user_id returns 401", func(t *testing.T) {
		t.Parallel()

		router := gin.New()
		router.Use(middleware.Handle())
		router.GET("/test", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		})

		// user_id なしの token を作成します
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"role": "admin",
			"exp":  time.Now().Add(time.Hour).Unix(),
		})
		tokenString, _ := token.SignedString([]byte(testSecretKey))

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("Authorization", "Bearer "+tokenString)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusUnauthorized, rec.Code)
		assert.Contains(t, rec.Body.String(), "Invalid user_id")
	})

	t.Run("token without role returns 401", func(t *testing.T) {
		t.Parallel()

		router := gin.New()
		router.Use(middleware.Handle())
		router.GET("/test", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		})

		// role なしの token を作成します
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"user_id": float64(1),
			"exp":     time.Now().Add(time.Hour).Unix(),
		})
		tokenString, _ := token.SignedString([]byte(testSecretKey))

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("Authorization", "Bearer "+tokenString)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusUnauthorized, rec.Code)
		assert.Contains(t, rec.Body.String(), "Missing or invalid role")
	})

	t.Run("token with invalid role returns 401", func(t *testing.T) {
		t.Parallel()

		router := gin.New()
		router.Use(middleware.Handle())
		router.GET("/test", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		})

		invalidRoleToken := createTestToken(1, "superadmin", time.Now().Add(time.Hour), testSecretKey)

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("Authorization", "Bearer "+invalidRoleToken)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusUnauthorized, rec.Code)
		assert.Contains(t, rec.Body.String(), "Invalid role value")
	})
}

func TestRequireRole(t *testing.T) {
	t.Parallel()

	jwtMiddleware := NewJWTMiddleware(testSecretKey)

	t.Run("allowed role passes", func(t *testing.T) {
		t.Parallel()

		router := gin.New()
		router.Use(jwtMiddleware.Handle())
		router.Use(RequireRole(domain.RoleAdmin, domain.RoleEditor))
		router.GET("/test", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		})

		adminToken := createTestToken(1, "admin", time.Now().Add(time.Hour), testSecretKey)

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("Authorization", "Bearer "+adminToken)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("non-allowed role returns 403", func(t *testing.T) {
		t.Parallel()

		router := gin.New()
		router.Use(jwtMiddleware.Handle())
		router.Use(RequireRole(domain.RoleAdmin))
		router.GET("/test", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		})

		viewerToken := createTestToken(1, "viewer", time.Now().Add(time.Hour), testSecretKey)

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("Authorization", "Bearer "+viewerToken)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusForbidden, rec.Code)
		assert.Contains(t, rec.Body.String(), "Insufficient permissions")
	})

	t.Run("missing role in context returns 403", func(t *testing.T) {
		t.Parallel()

		router := gin.New()
		// role がない状態をシミュレートするため JWT middleware をスキップします
		router.Use(RequireRole(domain.RoleAdmin))
		router.GET("/test", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		})

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusForbidden, rec.Code)
		assert.Contains(t, rec.Body.String(), "Role not found")
	})
}

func TestRequireAdmin(t *testing.T) {
	t.Parallel()

	jwtMiddleware := NewJWTMiddleware(testSecretKey)

	t.Run("admin passes", func(t *testing.T) {
		t.Parallel()

		router := gin.New()
		router.Use(jwtMiddleware.Handle())
		router.Use(RequireAdmin())
		router.GET("/test", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		})

		adminToken := createTestToken(1, "admin", time.Now().Add(time.Hour), testSecretKey)

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("Authorization", "Bearer "+adminToken)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("editor fails", func(t *testing.T) {
		t.Parallel()

		router := gin.New()
		router.Use(jwtMiddleware.Handle())
		router.Use(RequireAdmin())
		router.GET("/test", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		})

		editorToken := createTestToken(1, "editor", time.Now().Add(time.Hour), testSecretKey)

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("Authorization", "Bearer "+editorToken)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusForbidden, rec.Code)
	})
}

func TestRequireEditorOrAdmin(t *testing.T) {
	t.Parallel()

	jwtMiddleware := NewJWTMiddleware(testSecretKey)

	tests := []struct {
		name       string
		role       string
		wantStatus int
	}{
		{"admin passes", "admin", http.StatusOK},
		{"editor passes", "editor", http.StatusOK},
		{"viewer fails", "viewer", http.StatusForbidden},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			router := gin.New()
			router.Use(jwtMiddleware.Handle())
			router.Use(RequireEditorOrAdmin())
			router.GET("/test", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"status": "ok"})
			})

			token := createTestToken(1, tt.role, time.Now().Add(time.Hour), testSecretKey)

			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			req.Header.Set("Authorization", "Bearer "+token)
			rec := httptest.NewRecorder()

			router.ServeHTTP(rec, req)

			require.Equal(t, tt.wantStatus, rec.Code)
		})
	}
}
