package db

import (
	"time"

	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Connect はセキュアなロギングを使用してデータベース接続を確立します
func Connect(databaseURL string, zapLogger *zap.Logger, dbLogLevel string) *gorm.DB {
	// 機密データをマスクするセキュアなロガーを作成します
	gormLogger := NewSecureLogger(zapLogger, dbLogLevel)

	db, err := gorm.Open(postgres.Open(databaseURL), &gorm.Config{
		Logger: gormLogger,
	})
	if err != nil {
		zapLogger.Fatal("Failed to connect to database", zap.Error(err))
	}

	// コネクションプールを設定します
	sqlDB, err := db.DB()
	if err != nil {
		zapLogger.Fatal("Failed to get database instance", zap.Error(err))
	}

	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)
	sqlDB.SetConnMaxIdleTime(10 * time.Minute)

	zapLogger.Info("Successfully connected to database with connection pool configured")
	return db
}
