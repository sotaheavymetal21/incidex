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
}

// Insecure default values - only for local development
const (
	defaultJWTSecret      = "super-secret-key-change-me-in-production"
	defaultDatabaseURL    = "postgres://user:password@localhost:5432/incidex?sslmode=disable"
	defaultMinioAccessKey = "minioadmin"
	defaultMinioSecretKey = "minioadmin"
)

func Load() *Config {
	cfg := &Config{
		Port:           getEnv("PORT", "8080"),
		DatabaseURL:    getEnv("DATABASE_URL", defaultDatabaseURL),
		RedisURL:       getEnv("REDIS_URL", "localhost:6379"),
		MinioEndpoint:  getEnv("MINIO_ENDPOINT", "localhost:9000"),
		MinioAccessKey: getEnv("MINIO_ACCESS_KEY", defaultMinioAccessKey),
		MinioSecretKey: getEnv("MINIO_SECRET_KEY", defaultMinioSecretKey),
		JWTSecret:      getEnv("JWT_SECRET", defaultJWTSecret),
		AppEnv:         getEnv("APP_ENV", "development"),
	}

	// Validate configuration for production environment
	if isProduction(cfg.AppEnv) {
		validateProductionConfig(cfg)
	} else if isDevelopment(cfg.AppEnv) {
		logDevelopmentWarnings(cfg)
	}

	return cfg
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func isProduction(env string) bool {
	return strings.ToLower(env) == "production" || strings.ToLower(env) == "prod"
}

func isDevelopment(env string) bool {
	return strings.ToLower(env) == "development" || strings.ToLower(env) == "dev"
}

func validateProductionConfig(cfg *Config) {
	errors := []string{}

	// Check JWT Secret
	if cfg.JWTSecret == defaultJWTSecret || len(cfg.JWTSecret) < 32 {
		errors = append(errors, "JWT_SECRET must be set to a strong secret (minimum 32 characters) in production")
	}

	// Check Database URL
	if cfg.DatabaseURL == defaultDatabaseURL {
		errors = append(errors, "DATABASE_URL must be properly configured in production")
	}
	if strings.Contains(cfg.DatabaseURL, "sslmode=disable") {
		log.Println("WARNING: Database SSL is disabled. This is insecure for production.")
	}

	// Check MinIO credentials
	if cfg.MinioAccessKey == defaultMinioAccessKey {
		errors = append(errors, "MINIO_ACCESS_KEY must be changed from default value in production")
	}
	if cfg.MinioSecretKey == defaultMinioSecretKey {
		errors = append(errors, "MINIO_SECRET_KEY must be changed from default value in production")
	}

	// If there are any critical errors, panic to prevent startup
	if len(errors) > 0 {
		errorMsg := "CRITICAL SECURITY ERROR: Production environment detected with insecure configuration:\n"
		for _, err := range errors {
			errorMsg += fmt.Sprintf("  - %s\n", err)
		}
		errorMsg += "\nPlease set proper environment variables before running in production."
		panic(errorMsg)
	}

	log.Println("INFO: Production configuration validated successfully")
}

func logDevelopmentWarnings(cfg *Config) {
	warnings := []string{}

	if cfg.JWTSecret == defaultJWTSecret {
		warnings = append(warnings, "Using default JWT_SECRET (not suitable for production)")
	}
	if cfg.DatabaseURL == defaultDatabaseURL {
		warnings = append(warnings, "Using default DATABASE_URL (not suitable for production)")
	}
	if cfg.MinioAccessKey == defaultMinioAccessKey {
		warnings = append(warnings, "Using default MINIO credentials (not suitable for production)")
	}

	if len(warnings) > 0 {
		log.Println("DEVELOPMENT MODE: The following default values are being used:")
		for _, warning := range warnings {
			log.Printf("  - %s\n", warning)
		}
		log.Println("These defaults are convenient for local development but must be changed for production.")
	}
}
