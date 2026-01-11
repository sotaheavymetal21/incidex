package domain

import (
	"context"
	"time"
)

// LLMProviderType はLLMプロバイダーの種類
type LLMProviderType string

const (
	LLMProviderOpenAI      LLMProviderType = "openai"
	LLMProviderAzureOpenAI LLMProviderType = "azure-openai"
	LLMProviderClaude      LLMProviderType = "claude"
	LLMProviderOllama      LLMProviderType = "ollama"
	LLMProviderCustom      LLMProviderType = "custom"
)

// UserLLMSetting はユーザーごとのLLM設定
// SECURITY: API keys are NOT stored on the server
// Users manage API keys in their browser (localStorage)
type UserLLMSetting struct {
	ID           uint            `gorm:"primaryKey" json:"id"`
	UserID       uint            `gorm:"not null;uniqueIndex" json:"user_id"`
	User         *User           `gorm:"foreignKey:UserID" json:"user,omitempty"`
	ProviderType LLMProviderType `gorm:"not null" json:"provider_type"`
	Endpoint     string          `gorm:"type:varchar(500)" json:"endpoint,omitempty"` // Custom endpoint
	Model        string          `gorm:"not null" json:"model"`
	Enabled      bool            `gorm:"default:true" json:"enabled"`
	CreatedAt    time.Time       `json:"created_at"`
	UpdatedAt    time.Time       `json:"updated_at"`
}

// LLMRuntimeConfig はリクエスト時にフロントエンドから送信される実行時設定
// APIキーを含むが、サーバーには保存しない
type LLMRuntimeConfig struct {
	ProviderType LLMProviderType `json:"provider_type"`
	APIKey       string          `json:"api_key"`        // Only in memory, never persisted
	Endpoint     string          `json:"endpoint"`
	Model        string          `json:"model"`
}

// UserLLMSettingRepository はLLM設定のリポジトリインターフェース
type UserLLMSettingRepository interface {
	Create(ctx context.Context, setting *UserLLMSetting) error
	Update(ctx context.Context, setting *UserLLMSetting) error
	Delete(ctx context.Context, userID uint) error
	FindByUserID(ctx context.Context, userID uint) (*UserLLMSetting, error)
	FindAll(ctx context.Context) ([]*UserLLMSetting, error)
}

// IsValidProvider はプロバイダータイプが有効かチェック
func IsValidProvider(provider LLMProviderType) bool {
	switch provider {
	case LLMProviderOpenAI, LLMProviderAzureOpenAI, LLMProviderClaude, LLMProviderOllama, LLMProviderCustom:
		return true
	default:
		return false
	}
}

// GetDefaultEndpoint はプロバイダーごとのデフォルトエンドポイントを返す
// Azure OpenAIとCustomプロバイダーはエンドポイントが必須のため、空文字列を返す
func GetDefaultEndpoint(provider LLMProviderType) string {
	switch provider {
	case LLMProviderOpenAI:
		return "https://api.openai.com/v1"
	case LLMProviderClaude:
		return "https://api.anthropic.com/v1"
	case LLMProviderOllama:
		return "http://localhost:11434/v1"
	case LLMProviderAzureOpenAI, LLMProviderCustom:
		// これらのプロバイダーはエンドポイントが必須のため、デフォルトは返さない
		return ""
	default:
		return ""
	}
}
