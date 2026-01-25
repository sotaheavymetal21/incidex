package db

import (
	"context"

	"gorm.io/gorm"
)

// TxFunc はトランザクションを受け取り error を返す関数型です
type TxFunc func(*gorm.DB) error

// WithTransaction はデータベーストランザクション内で関数を実行します
// 関数が error を返した場合、トランザクションはロールバックされます
// それ以外の場合、トランザクションはコミットされます
func WithTransaction(ctx context.Context, db *gorm.DB, fn TxFunc) error {
	tx := db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return tx.Error
	}

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r) // ロールバック後に panic を再スローします
		}
	}()

	if err := fn(tx); err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}
