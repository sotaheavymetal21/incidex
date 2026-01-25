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

// SecureLogger はセキュリティに重点を置いたロギングを実装した gorm.Interface です
type SecureLogger struct {
	ZapLogger                 *zap.Logger
	LogLevel                  gormlogger.LogLevel
	SlowThreshold             time.Duration
	IgnoreRecordNotFoundError bool
	MaskParameters            bool // セキュリティのため常に SQL パラメータをマスクします
}

// LogMode はログモードを設定します
func (l *SecureLogger) LogMode(level gormlogger.LogLevel) gormlogger.Interface {
	newlogger := *l
	newlogger.LogLevel = level
	return &newlogger
}

// Info は info レベルのメッセージをログに記録します
func (l *SecureLogger) Info(ctx context.Context, msg string, data ...interface{}) {
	if l.LogLevel >= gormlogger.Info {
		l.ZapLogger.Sugar().Infof(msg, data...)
	}
}

// Warn は warn レベルのメッセージをログに記録します
func (l *SecureLogger) Warn(ctx context.Context, msg string, data ...interface{}) {
	if l.LogLevel >= gormlogger.Warn {
		l.ZapLogger.Sugar().Warnf(msg, data...)
	}
}

// Error は error レベルのメッセージをログに記録します
func (l *SecureLogger) Error(ctx context.Context, msg string, data ...interface{}) {
	if l.LogLevel >= gormlogger.Error {
		l.ZapLogger.Sugar().Errorf(msg, data...)
	}
}

// Trace はセキュリティマスキングを適用して SQL クエリをログに記録します
func (l *SecureLogger) Trace(ctx context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error) {
	if l.LogLevel <= gormlogger.Silent {
		return
	}

	elapsed := time.Since(begin)
	sql, rows := fc()

	// SQL 内の機密パラメータをマスクします
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

// maskSensitiveData は SQL クエリ内の機密情報をマスクします
func (l *SecureLogger) maskSensitiveData(sql string) string {
	if !l.MaskParameters {
		return sql
	}

	// 包括的なマスキングのために一元化されたサニタイザーを使用します
	return sanitizer.SanitizeSQL(sql)
}

// NewSecureLogger は新しいセキュアロガーインスタンスを作成します
func NewSecureLogger(zapLogger *zap.Logger, dbLogLevel string) gormlogger.Interface {
	config := &SecureLogger{
		ZapLogger:                 zapLogger,
		SlowThreshold:             200 * time.Millisecond,
		IgnoreRecordNotFoundError: true,
		MaskParameters:            true, // セキュリティのため常にマスクします
		LogLevel:                  ParseLogLevel(dbLogLevel),
	}

	return config
}

// ParseLogLevel は文字列のログレベルを GORM のログレベルに変換します
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

// LogLevelString はログレベルの文字列表現を返します
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
