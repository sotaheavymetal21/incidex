package middleware

import (
	"os"
	"strings"

	"github.com/gin-gonic/gin"
)

// SecurityHeaders adds security headers to all responses
func SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check if running in production
		appEnv := os.Getenv("APP_ENV")
		isProduction := strings.ToLower(appEnv) == "production" || strings.ToLower(appEnv) == "prod"
		// Prevent clickjacking attacks
		c.Header("X-Frame-Options", "DENY")

		// Prevent MIME type sniffing
		c.Header("X-Content-Type-Options", "nosniff")

		// Enable XSS protection
		c.Header("X-XSS-Protection", "1; mode=block")

		// Content Security Policy - restrictive default
		// Adjust this based on your frontend needs
		csp := "default-src 'self'; " +
			"script-src 'self' 'unsafe-inline' 'unsafe-eval'; " +
			"style-src 'self' 'unsafe-inline'; " +
			"img-src 'self' data: https:; " +
			"font-src 'self' data:; " +
			"connect-src 'self'; " +
			"frame-ancestors 'none';"
		c.Header("Content-Security-Policy", csp)

		// Referrer Policy
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")

		// HSTS - only enable in production with HTTPS
		if isProduction {
			c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains; preload")
		}

		// Additional security headers
		c.Header("X-Permitted-Cross-Domain-Policies", "none")
		c.Header("Permissions-Policy", "geolocation=(), microphone=(), camera=()")

		c.Next()
	}
}
