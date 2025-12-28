package logger

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var Log *zap.Logger

// InitLogger initializes the global logger based on environment
func InitLogger(env string) error {
	var config zap.Config

	if env == "production" {
		config = zap.NewProductionConfig()
		// In production, log JSON format with INFO level
		config.Level = zap.NewAtomicLevelAt(zapcore.InfoLevel)
	} else {
		config = zap.NewDevelopmentConfig()
		// In development, log human-readable format with DEBUG level
		config.Level = zap.NewAtomicLevelAt(zapcore.DebugLevel)
	}

	// Customize time encoding
	config.EncoderConfig.TimeKey = "timestamp"
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	// Build logger
	logger, err := config.Build()
	if err != nil {
		return err
	}

	Log = logger
	return nil
}

// Sync flushes any buffered log entries
func Sync() {
	if Log != nil {
		_ = Log.Sync()
	}
}

// GetEnv returns the environment from ENV variable, defaults to "development"
func GetEnv() string {
	env := os.Getenv("APP_ENV")
	if env == "" {
		env = "development"
	}
	return env
}
