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
	// Database logging configuration
	// Options: "silent", "error", "warn", "info"
	// Default: "warn" for production, "info" for development
	DBLogLevel string
	// CORS configuration
	CORSAllowedOrigins []string
	// Initial admin user (created on first startup if no users exist)
	InitialAdminEmail    string
	InitialAdminPassword string
	InitialAdminName     string
	// Frontend URL for email links
	FrontendURL string
	// Auto migration configuration
	AutoMigrate     bool
	MigrationsDir   string
}

func Load() *Config {
	appEnv := getEnv("APP_ENV", "development")

	// Required environment variables - no defaults for security
	cfg := &Config{
		Port:                 getEnv("PORT", "8080"),
		DatabaseURL:          getEnvRequired("DATABASE_URL"),
		RedisURL:             getEnvRequired("REDIS_URL"),
		MinioEndpoint:        getEnvRequired("MINIO_ENDPOINT"),
		MinioAccessKey:       getEnvRequired("MINIO_ACCESS_KEY"),
		MinioSecretKey:       getEnvRequired("MINIO_SECRET_KEY"),
		JWTSecret:            getEnvRequired("JWT_SECRET"),
		AppEnv:               appEnv,
		DBLogLevel:           getEnv("DB_LOG_LEVEL", getDefaultDBLogLevel(appEnv)),
		CORSAllowedOrigins:   parseCORSOrigins(getEnvRequired("CORS_ALLOWED_ORIGINS")),
		InitialAdminEmail:    getEnv("INITIAL_ADMIN_EMAIL", ""),
		InitialAdminPassword: getEnv("INITIAL_ADMIN_PASSWORD", ""),
		InitialAdminName:     getEnv("INITIAL_ADMIN_NAME", ""),
		FrontendURL:          getEnvRequired("FRONTEND_URL"),
		AutoMigrate:          getEnv("AUTO_MIGRATE", "false") == "true",
		MigrationsDir:        getEnv("MIGRATIONS_DIR", "./migrations"),
	}

	// Validate configuration
	validateConfig(cfg)

	return cfg
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

// getEnvRequired returns the value of an environment variable or panics if not set
func getEnvRequired(key string) string {
	value, ok := os.LookupEnv(key)
	if !ok || value == "" {
		panic(fmt.Sprintf("CRITICAL: Required environment variable %s is not set", key))
	}
	return value
}

// parseCORSOrigins parses a comma-separated string into a slice of origins
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

func getDefaultDBLogLevel(env string) string {
	if isProduction(env) {
		return "warn" // Production: only log errors and slow queries
	}
	return "info" // Development: log all queries (with masked sensitive data)
}

// validateConfig validates the configuration and logs warnings/errors
func validateConfig(cfg *Config) {
	errors := []string{}
	warnings := []string{}

	// Validate JWT Secret length
	if len(cfg.JWTSecret) < 32 {
		errors = append(errors, "JWT_SECRET must be at least 32 characters long")
	}

	// Validate Database SSL in production
	if isProduction(cfg.AppEnv) && strings.Contains(cfg.DatabaseURL, "sslmode=disable") {
		warnings = append(warnings, "Database SSL is disabled in production - this is insecure")
	}

	// Validate CORS origins
	if len(cfg.CORSAllowedOrigins) == 0 {
		errors = append(errors, "CORS_ALLOWED_ORIGINS must be configured")
	}

	// Log warnings
	if len(warnings) > 0 {
		log.Println("Configuration warnings:")
		for _, warning := range warnings {
			log.Printf("  - WARNING: %s\n", warning)
		}
	}

	// Panic on critical errors
	if len(errors) > 0 {
		errorMsg := "CRITICAL: Configuration validation failed:\n"
		for _, err := range errors {
			errorMsg += fmt.Sprintf("  - %s\n", err)
		}
		panic(errorMsg)
	}

	log.Printf("Configuration loaded successfully (environment: %s)", cfg.AppEnv)
}
