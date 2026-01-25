package db

import (
	"time"

	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Connect establishes a database connection with secure logging
func Connect(databaseURL string, zapLogger *zap.Logger, isProduction bool) *gorm.DB {
	// Create secure logger that masks sensitive data
	gormLogger := NewSecureLogger(zapLogger, isProduction)

	db, err := gorm.Open(postgres.Open(databaseURL), &gorm.Config{
		Logger: gormLogger,
	})
	if err != nil {
		zapLogger.Fatal("Failed to connect to database", zap.Error(err))
	}

	// Configure connection pool
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
