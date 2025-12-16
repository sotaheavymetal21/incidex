package config

import (
	"os"
)

type Config struct {
	Port           string
	DatabaseURL    string
	RedisURL       string
	MinioEndpoint  string
	MinioAccessKey string
	MinioSecretKey string
	JWTSecret      string
}

func Load() *Config {
	return &Config{
		Port: getEnv("PORT", "8080"),
		// Note: Default values match the settings in .env for local development convenience.
		// In production, these should be strictly set via environment variables.
		DatabaseURL:    getEnv("DATABASE_URL", "postgres://user:password@localhost:5432/incidex?sslmode=disable"),
		RedisURL:       getEnv("REDIS_URL", "localhost:6379"),
		MinioEndpoint:  getEnv("MINIO_ENDPOINT", "localhost:9000"),
		MinioAccessKey: getEnv("MINIO_ACCESS_KEY", "minioadmin"),
		MinioSecretKey: getEnv("MINIO_SECRET_KEY", "minioadmin"),
		JWTSecret:      getEnv("JWT_SECRET", "super-secret-key-change-me-in-production"),
	}
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
