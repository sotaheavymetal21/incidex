package logger

import (
	"os"
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var Log *zap.Logger

// InitLogger は環境に基づいてグローバルロガーを初期化します
func InitLogger(env string) error {
	var config zap.Config

	isProduction := strings.ToLower(env) == "production" || strings.ToLower(env) == "prod"

	if isProduction {
		config = zap.NewProductionConfig()
		// 本番環境では、デフォルトで INFO レベルの JSON 形式でログを記録します
		config.Level = zap.NewAtomicLevelAt(zapcore.InfoLevel)
		// 本番環境ではセキュリティのためスタックトレースを無効化します
		config.DisableStacktrace = true
		// 本番環境ではログサイズを削減するため呼び出し元情報を無効化できます（デバッグ用に維持可能）
		config.DisableCaller = false
	} else {
		config = zap.NewDevelopmentConfig()
		// 開発環境では、デフォルトで DEBUG レベルの人間が読める形式でログを記録します
		config.Level = zap.NewAtomicLevelAt(zapcore.DebugLevel)
		// 開発環境ではスタックトレースを有効化します
		config.DisableStacktrace = false
	}

	// 環境変数が指定されている場合はログレベルを上書きします
	if logLevel := os.Getenv("LOG_LEVEL"); logLevel != "" {
		level := parseLogLevel(logLevel)
		config.Level = zap.NewAtomicLevelAt(level)
	}

	// 環境変数からスタックトレースの動作を設定します
	if stackTrace := os.Getenv("LOG_STACKTRACE"); stackTrace != "" {
		config.DisableStacktrace = strings.ToLower(stackTrace) == "false"
	}

	// タイムエンコーディングをカスタマイズします
	config.EncoderConfig.TimeKey = "timestamp"
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	// ログのパースを改善するため severity フィールドを追加します
	config.EncoderConfig.LevelKey = "severity"
	config.EncoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder

	// メッセージキーをカスタマイズします
	config.EncoderConfig.MessageKey = "message"

	// 出力パスを設定します
	config.OutputPaths = []string{"stdout"}
	config.ErrorOutputPaths = []string{"stderr"}

	// ロガーをビルドします
	logger, err := config.Build(
		// すべてのログに共通フィールドを追加します
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

// parseLogLevel は文字列のログレベルを zapcore.Level に変換します
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

// Sync はバッファされたログエントリをフラッシュします
func Sync() {
	if Log != nil {
		_ = Log.Sync()
	}
}

// GetEnv は ENV 変数から環境を返します。デフォルトは "development" です
func GetEnv() string {
	env := os.Getenv("APP_ENV")
	if env == "" {
		env = "development"
	}
	return env
}
