package middleware

import (
	"os"
	"strings"

	"github.com/gin-gonic/gin"
)

// SecurityHeaders はすべての response にセキュリティヘッダーを追加します
func SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 本番環境で実行中かチェックします
		appEnv := os.Getenv("APP_ENV")
		isProduction := strings.ToLower(appEnv) == "production" || strings.ToLower(appEnv) == "prod"
		// クリックジャッキング攻撃を防止します
		c.Header("X-Frame-Options", "DENY")

		// MIME タイプスニッフィングを防止します
		c.Header("X-Content-Type-Options", "nosniff")

		// XSS 保護を有効にします
		c.Header("X-XSS-Protection", "1; mode=block")

		// Content Security Policy
		// 注意: 'unsafe-inline' は Next.js の styled-jsx とインラインスタイルに必要です
		// 'unsafe-eval' はセキュリティのため削除されました
		csp := "default-src 'self'; " +
			"script-src 'self' 'unsafe-inline'; " +
			"style-src 'self' 'unsafe-inline'; " +
			"img-src 'self' data: https:; " +
			"font-src 'self' data:; " +
			"connect-src 'self'; " +
			"frame-ancestors 'none'; " +
			"base-uri 'self'; " +
			"form-action 'self';"
		c.Header("Content-Security-Policy", csp)

		// Referrer ポリシー
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")

		// HSTS - HTTPS を使用する本番環境でのみ有効にします
		if isProduction {
			c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains; preload")
		}

		// 追加のセキュリティヘッダー
		c.Header("X-Permitted-Cross-Domain-Policies", "none")
		c.Header("Permissions-Policy", "geolocation=(), microphone=(), camera=()")

		c.Next()
	}
}
