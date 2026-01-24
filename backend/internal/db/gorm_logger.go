package db

import (
	"context"
	"errors"
	"incidex/internal/pkg/sanitizer"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

// SecureLogger implements gorm.Interface with security-focused logging
type SecureLogger struct {
	ZapLogger                 *zap.Logger
	LogLevel                  gormlogger.LogLevel
	SlowThreshold             time.Duration
	IgnoreRecordNotFoundError bool
	MaskParameters            bool // Always mask SQL parameters for security
}

// LogMode sets log mode
func (l *SecureLogger) LogMode(level gormlogger.LogLevel) gormlogger.Interface {
	newlogger := *l
	newlogger.LogLevel = level
	return &newlogger
}

// Info logs info level messages
func (l *SecureLogger) Info(ctx context.Context, msg string, data ...interface{}) {
	if l.LogLevel >= gormlogger.Info {
		l.ZapLogger.Sugar().Infof(msg, data...)
	}
}

// Warn logs warn level messages
func (l *SecureLogger) Warn(ctx context.Context, msg string, data ...interface{}) {
	if l.LogLevel >= gormlogger.Warn {
		l.ZapLogger.Sugar().Warnf(msg, data...)
	}
}

// Error logs error level messages
func (l *SecureLogger) Error(ctx context.Context, msg string, data ...interface{}) {
	if l.LogLevel >= gormlogger.Error {
		l.ZapLogger.Sugar().Errorf(msg, data...)
	}
}

// Trace logs SQL queries with security masking
func (l *SecureLogger) Trace(ctx context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error) {
	if l.LogLevel <= gormlogger.Silent {
		return
	}

	elapsed := time.Since(begin)
	sql, rows := fc()

	// Mask sensitive parameters in SQL
	maskedSQL := l.maskSensitiveData(sql)

	switch {
	case err != nil && l.LogLevel >= gormlogger.Error:
		if l.IgnoreRecordNotFoundError && errors.Is(err, gorm.ErrRecordNotFound) {
			return
		}
		l.ZapLogger.Error("database error",
			zap.Error(err),
			zap.Duration("elapsed", elapsed),
			zap.Int64("rows", rows),
			zap.String("sql", maskedSQL),
		)

	case elapsed > l.SlowThreshold && l.SlowThreshold != 0 && l.LogLevel >= gormlogger.Warn:
		l.ZapLogger.Warn("slow query",
			zap.Duration("elapsed", elapsed),
			zap.Duration("threshold", l.SlowThreshold),
			zap.Int64("rows", rows),
			zap.String("sql", maskedSQL),
		)

	case l.LogLevel >= gormlogger.Info:
		l.ZapLogger.Debug("database query",
			zap.Duration("elapsed", elapsed),
			zap.Int64("rows", rows),
			zap.String("sql", maskedSQL),
		)
	}
}

// maskSensitiveData masks sensitive information in SQL queries
func (l *SecureLogger) maskSensitiveData(sql string) string {
	if !l.MaskParameters {
		return sql
	}

	// Use the centralized sanitizer for comprehensive masking
	return sanitizer.SanitizeSQL(sql)
}

// NewSecureLogger creates a new secure logger instance
func NewSecureLogger(zapLogger *zap.Logger, isProduction bool) gormlogger.Interface {
	config := &SecureLogger{
		ZapLogger:                 zapLogger,
		SlowThreshold:             200 * time.Millisecond,
		IgnoreRecordNotFoundError: true,
		MaskParameters:            true, // Always mask for security
	}

	if isProduction {
		// Production: Only log errors and slow queries
		config.LogLevel = gormlogger.Warn
	} else {
		// Development: Log more details but still mask sensitive data
		config.LogLevel = gormlogger.Info
	}

	return config
}

// ParseLogLevel converts string log level to gorm log level
func ParseLogLevel(level string) gormlogger.LogLevel {
	switch level {
	case "silent":
		return gormlogger.Silent
	case "error":
		return gormlogger.Error
	case "warn":
		return gormlogger.Warn
	case "info":
		return gormlogger.Info
	default:
		return gormlogger.Warn
	}
}

// LogLevelString returns string representation of log level
func LogLevelString(level gormlogger.LogLevel) string {
	switch level {
	case gormlogger.Silent:
		return "silent"
	case gormlogger.Error:
		return "error"
	case gormlogger.Warn:
		return "warn"
	case gormlogger.Info:
		return "info"
	default:
		return "unknown"
	}
}
