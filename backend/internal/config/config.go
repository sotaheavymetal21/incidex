package config

import (
	"fmt"
	"log"
	"os"
	"strings"
)

type Config struct {
	Port           string
	DatabaseURL    string
	RedisURL       string
	MinioEndpoint  string
	MinioAccessKey string
	MinioSecretKey string
	JWTSecret      string
	AppEnv         string
	// ロギング設定
	// LOG_LEVEL: アプリケーションログレベル (debug, info, warn, error, fatal)
	// デフォルト: 本番環境では "info"、開発環境では "debug"
	LogLevel string
	// LOG_STACKTRACE: スタックトレースの有効化/無効化 (true/false)
	// デフォルト: 本番環境では false、開発環境では true
	LogStacktrace bool
	// データベースロギング設定
	// DB_LOG_LEVEL: データベースクエリログレベル (silent, error, warn, info)
	// デフォルト: 本番環境では "warn"、開発環境では "info"
	DBLogLevel string
	// CORS 設定
	CORSAllowedOrigins []string
	// 初期管理者ユーザー（ユーザーが存在しない場合に初回起動時に作成されます）
	InitialAdminEmail    string
	InitialAdminPassword string
	InitialAdminName     string
	// メールリンク用のフロントエンド URL
	FrontendURL string
	// 自動マイグレーション設定
	AutoMigrate   bool
	MigrationsDir string
}

func Load() *Config {
	appEnv := getEnv("APP_ENV", "development")

	// 必須環境変数 - セキュリティのためデフォルト値なし
	cfg := &Config{
		Port:                 getEnv("PORT", "8080"),
		DatabaseURL:          getEnvRequired("DATABASE_URL"),
		RedisURL:             getEnvRequired("REDIS_URL"),
		MinioEndpoint:        getEnvRequired("MINIO_ENDPOINT"),
		MinioAccessKey:       getEnvRequired("MINIO_ACCESS_KEY"),
		MinioSecretKey:       getEnvRequired("MINIO_SECRET_KEY"),
		JWTSecret:            getEnvRequired("JWT_SECRET"),
		AppEnv:               appEnv,
		LogLevel:             getEnv("LOG_LEVEL", getDefaultLogLevel(appEnv)),
		LogStacktrace:        getEnv("LOG_STACKTRACE", getDefaultLogStacktrace(appEnv)) == "true",
		DBLogLevel:           getEnv("DB_LOG_LEVEL", getDefaultDBLogLevel(appEnv)),
		CORSAllowedOrigins:   parseCORSOrigins(getEnvRequired("CORS_ALLOWED_ORIGINS")),
		InitialAdminEmail:    getEnv("INITIAL_ADMIN_EMAIL", ""),
		InitialAdminPassword: getEnv("INITIAL_ADMIN_PASSWORD", ""),
		InitialAdminName:     getEnv("INITIAL_ADMIN_NAME", ""),
		FrontendURL:          getEnvRequired("FRONTEND_URL"),
		AutoMigrate:          getEnv("AUTO_MIGRATE", "false") == "true",
		MigrationsDir:        getEnv("MIGRATIONS_DIR", "./migrations"),
	}

	// 設定をバリデーションします
	validateConfig(cfg)

	return cfg
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

// getEnvRequired は環境変数の値を返します。設定されていない場合は panic します
func getEnvRequired(key string) string {
	value, ok := os.LookupEnv(key)
	if !ok || value == "" {
		panic(fmt.Sprintf("CRITICAL: Required environment variable %s is not set", key))
	}
	return value
}

// parseCORSOrigins はカンマ区切りの文字列をオリジンのスライスにパースします
func parseCORSOrigins(origins string) []string {
	if origins == "" {
		return []string{}
	}

	parts := strings.Split(origins, ",")
	result := make([]string, 0, len(parts))

	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}

	return result
}

func isProduction(env string) bool {
	return strings.ToLower(env) == "production" || strings.ToLower(env) == "prod"
}

func isDevelopment(env string) bool {
	return strings.ToLower(env) == "development" || strings.ToLower(env) == "dev"
}

func getDefaultLogLevel(env string) string {
	if isProduction(env) {
		return "info" // 本番環境: info レベルのログ
	}
	return "debug" // 開発環境: debug レベルのログ
}

func getDefaultLogStacktrace(env string) string {
	if isProduction(env) {
		return "false" // 本番環境: セキュリティのためスタックトレースを無効化
	}
	return "true" // 開発環境: デバッグのためスタックトレースを有効化
}

func getDefaultDBLogLevel(env string) string {
	// すべての環境でデフォルトは "warn"
	// error とスロークエリ（>200ms）のみをログに記録します
	// すべてのクエリを表示するには DB_LOG_LEVEL=info を設定します（開発/デバッグ専用）
	return "warn"
}

// validateConfig は設定をバリデーションし、警告/エラーをログに記録します
func validateConfig(cfg *Config) {
	errors := []string{}
	warnings := []string{}

	// JWT Secret の長さをバリデーションします
	if len(cfg.JWTSecret) < 32 {
		errors = append(errors, "JWT_SECRET must be at least 32 characters long")
	}

	// 本番環境でのデータベース SSL をバリデーションします
	if isProduction(cfg.AppEnv) && strings.Contains(cfg.DatabaseURL, "sslmode=disable") {
		warnings = append(warnings, "Database SSL is disabled in production - this is insecure")
	}

	// CORS オリジンをバリデーションします
	if len(cfg.CORSAllowedOrigins) == 0 {
		errors = append(errors, "CORS_ALLOWED_ORIGINS must be configured")
	}

	// 警告をログに記録します
	if len(warnings) > 0 {
		log.Println("Configuration warnings:")
		for _, warning := range warnings {
			log.Printf("  - WARNING: %s\n", warning)
		}
	}

	// 重大なエラーの場合は panic します
	if len(errors) > 0 {
		errorMsg := "CRITICAL: Configuration validation failed:\n"
		for _, err := range errors {
			errorMsg += fmt.Sprintf("  - %s\n", err)
		}
		panic(errorMsg)
	}

	log.Printf("Configuration loaded successfully (environment: %s)", cfg.AppEnv)
}
