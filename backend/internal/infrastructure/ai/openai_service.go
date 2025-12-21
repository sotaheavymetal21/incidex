package ai

import (
	"context"
	"fmt"
	"incidex/internal/domain"
	"os"
	"strings"

	"github.com/sashabaranov/go-openai"
)

// OpenAIService はOpenAI APIを使用したAIサービス
type OpenAIService struct {
	client *openai.Client
	model  string
}

// NewOpenAIService は新しいOpenAIサービスを作成します
func NewOpenAIService() *OpenAIService {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		// API Keyがない場合は nil を返す（開発環境用）
		fmt.Println("[AI] OpenAI API Key not configured. AI summary will be disabled.")
		return nil
	}

	client := openai.NewClient(apiKey)
	model := os.Getenv("OPENAI_MODEL")
	if model == "" {
		model = "gpt-4-turbo-preview" // デフォルトモデル
	}

	return &OpenAIService{
		client: client,
		model:  model,
	}
}

// GenerateIncidentSummary はインシデントの説明から要約を生成します
func (s *OpenAIService) GenerateIncidentSummary(title, description, severity, impactScope string) (string, error) {
	if s == nil || s.client == nil {
		// API Keyが設定されていない場合は空文字列を返す
		return "", nil
	}

	prompt := fmt.Sprintf(`以下のインシデント情報を簡潔に要約してください。要約は2-3文で、重要なポイントのみを含めてください。

タイトル: %s
重要度: %s
影響範囲: %s
詳細:
%s

要約:`, title, severity, impactScope, description)

	resp, err := s.client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: s.model,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleSystem,
					Content: "あなたはインシデント管理のエキスパートです。インシデントの詳細から簡潔で分かりやすい要約を作成します。",
				},
				{
					Role:    openai.ChatMessageRoleUser,
					Content: prompt,
				},
			},
			MaxTokens:   200,
			Temperature: 0.3, // より一貫性のある出力のために低めに設定
		},
	)

	if err != nil {
		return "", fmt.Errorf("failed to generate summary: %w", err)
	}

	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("no summary generated")
	}

	summary := resp.Choices[0].Message.Content
	return summary, nil
}

// GenerateRootCauseAnalysis は根本原因分析を生成します（将来の拡張用）
func (s *OpenAIService) GenerateRootCauseAnalysis(incidentDetails, resolution string) (string, error) {
	if s == nil || s.client == nil {
		return "", nil
	}

	prompt := fmt.Sprintf(`以下のインシデント情報と解決方法から、根本原因分析（RCA）を生成してください。

インシデント詳細:
%s

解決方法:
%s

根本原因分析:
1. 発生原因
2. 影響範囲
3. 再発防止策

を含めて作成してください。`, incidentDetails, resolution)

	resp, err := s.client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: s.model,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleSystem,
					Content: "あなたはインシデント管理と根本原因分析のエキスパートです。",
				},
				{
					Role:    openai.ChatMessageRoleUser,
					Content: prompt,
				},
			},
			MaxTokens:   500,
			Temperature: 0.3,
		},
	)

	if err != nil {
		return "", fmt.Errorf("failed to generate RCA: %w", err)
	}

	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("no RCA generated")
	}

	rca := resp.Choices[0].Message.Content
	return rca, nil
}

// GeneratePostMortemRootCauseSuggestion はインシデント情報とタイムラインから根本原因の候補を提案します
func (s *OpenAIService) GeneratePostMortemRootCauseSuggestion(
	incidentTitle, incidentDescription string,
	timeline []*domain.IncidentActivity,
) (string, error) {
	if s == nil || s.client == nil {
		return "", nil
	}

	// タイムラインをテキスト化（最大100件）
	var timelineText strings.Builder
	maxTimeline := 100
	if len(timeline) > maxTimeline {
		timeline = timeline[:maxTimeline]
	}

	for _, activity := range timeline {
		timelineText.WriteString(fmt.Sprintf("- [%s] %s",
			activity.CreatedAt.Format("2006-01-02 15:04:05"),
			activity.ActivityType))
		if activity.Comment != "" {
			timelineText.WriteString(fmt.Sprintf(": %s", activity.Comment))
		}
		if activity.OldValue != "" && activity.NewValue != "" {
			timelineText.WriteString(fmt.Sprintf(" (%s → %s)", activity.OldValue, activity.NewValue))
		}
		timelineText.WriteString("\n")
	}

	prompt := fmt.Sprintf(`以下のインシデント情報とタイムラインから、根本原因の候補を3-5つ提案してください。
優先度が高いと考えられる順に、以下の形式で記載してください:

1. [根本原因候補1]: 説明
2. [根本原因候補2]: 説明
3. [根本原因候補3]: 説明

インシデントタイトル: %s

インシデント詳細:
%s

タイムライン:
%s

根本原因候補:`, incidentTitle, incidentDescription, timelineText.String())

	resp, err := s.client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: s.model,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleSystem,
					Content: "あなたはインシデント管理と根本原因分析のエキスパートです。インシデントのタイムラインと詳細から、考えられる根本原因を優先度順に提案します。",
				},
				{
					Role:    openai.ChatMessageRoleUser,
					Content: prompt,
				},
			},
			MaxTokens:   800,
			Temperature: 0.3,
		},
	)

	if err != nil {
		return "", fmt.Errorf("failed to generate root cause suggestion: %w", err)
	}

	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("no root cause suggestion generated")
	}

	suggestion := resp.Choices[0].Message.Content
	return suggestion, nil
}
