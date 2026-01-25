package domain

import (
	"context"
	"time"
)

// CacheRepository はキャッシュ操作のインターフェースを定義します
type CacheRepository interface {
	// Get はキャッシュから値を取得します
	Get(ctx context.Context, key string) (string, error)

	// Set は値をキャッシュに保存します（TTL 0 = 有効期限なし）
	Set(ctx context.Context, key string, value string, ttl time.Duration) error

	// Delete はキャッシュから値を削除します
	Delete(ctx context.Context, key string) error

	// DeleteByPattern はパターンに一致するすべてのキーを削除します
	DeleteByPattern(ctx context.Context, pattern string) error

	// Exists はキーがキャッシュに存在するかを確認します
	Exists(ctx context.Context, key string) (bool, error)
}
