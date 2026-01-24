package logger

import (
	"os"
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var Log *zap.Logger

// InitLogger initializes the global logger based on environment
func InitLogger(env string) error {
	var config zap.Config

	isProduction := strings.ToLower(env) == "production" || strings.ToLower(env) == "prod"

	if isProduction {
		config = zap.NewProductionConfig()
		// In production, log JSON format with INFO level by default
		config.Level = zap.NewAtomicLevelAt(zapcore.InfoLevel)
		// Disable stack traces in production for security
		config.DisableStacktrace = true
		// Disable caller information in production to reduce log size
		config.DisableCaller = false // Keep caller for debugging, but can be disabled if needed
	} else {
		config = zap.NewDevelopmentConfig()
		// In development, log human-readable format with DEBUG level by default
		config.Level = zap.NewAtomicLevelAt(zapcore.DebugLevel)
		// Enable stack traces in development
		config.DisableStacktrace = false
	}

	// Override log level from environment variable if specified
	if logLevel := os.Getenv("LOG_LEVEL"); logLevel != "" {
		level := parseLogLevel(logLevel)
		config.Level = zap.NewAtomicLevelAt(level)
	}

	// Configure stack trace behavior from environment variable
	if stackTrace := os.Getenv("LOG_STACKTRACE"); stackTrace != "" {
		config.DisableStacktrace = strings.ToLower(stackTrace) == "false"
	}

	// Customize time encoding
	config.EncoderConfig.TimeKey = "timestamp"
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	// Add severity field for better log parsing
	config.EncoderConfig.LevelKey = "severity"
	config.EncoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder

	// Customize message key
	config.EncoderConfig.MessageKey = "message"

	// Configure output paths
	config.OutputPaths = []string{"stdout"}
	config.ErrorOutputPaths = []string{"stderr"}

	// Build logger
	logger, err := config.Build(
		// Add common fields to all logs
		zap.Fields(
			zap.String("service", "incidex"),
			zap.String("environment", env),
		),
	)
	if err != nil {
		return err
	}

	Log = logger
	return nil
}

// parseLogLevel converts string log level to zapcore.Level
func parseLogLevel(level string) zapcore.Level {
	switch strings.ToLower(level) {
	case "debug":
		return zapcore.DebugLevel
	case "info":
		return zapcore.InfoLevel
	case "warn", "warning":
		return zapcore.WarnLevel
	case "error":
		return zapcore.ErrorLevel
	case "fatal":
		return zapcore.FatalLevel
	case "panic":
		return zapcore.PanicLevel
	default:
		return zapcore.InfoLevel
	}
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
