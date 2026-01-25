package notification

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

// SlackService はSlack通知を送信するサービス
type SlackService struct {
	frontendURL string
}

// NewSlackService は新しいSlackサービスを作成します
func NewSlackService(frontendURL string) *SlackService {
	return &SlackService{
		frontendURL: frontendURL,
	}
}

// SlackMessage はSlackメッセージの構造体
type SlackMessage struct {
	Text        string       `json:"text,omitempty"`
	Blocks      []SlackBlock `json:"blocks,omitempty"`
	Attachments []Attachment `json:"attachments,omitempty"`
}

// SlackBlock はSlackブロックの構造体
type SlackBlock struct {
	Type string         `json:"type"`
	Text *SlackText     `json:"text,omitempty"`
	Fields []SlackText  `json:"fields,omitempty"`
}

// SlackText はSlackテキストの構造体
type SlackText struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// Attachment はSlack添付ファイルの構造体
type Attachment struct {
	Color  string `json:"color,omitempty"`
	Text   string `json:"text,omitempty"`
	Footer string `json:"footer,omitempty"`
}

// SendMessage はSlackにメッセージを送信します
func (s *SlackService) SendMessage(webhookURL string, message SlackMessage) error {
	if webhookURL == "" {
		// Webhook URLがない場合はログのみ出力（開発環境用）
		fmt.Printf("[SLACK] %s\n", message.Text)
		return nil
	}

	jsonData, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("Slackメッセージのマーシャルに失敗しました: %w", err)
	}

	resp, err := http.Post(webhookURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("Slackメッセージの送信に失敗しました: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Slackが非OKステータスを返しました: %d", resp.StatusCode)
	}

	return nil
}

// SendIncidentCreatedMessage はインシデント作成通知を送信します
func (s *SlackService) SendIncidentCreatedMessage(webhookURL, incidentTitle string, incidentID uint, severity, creatorName string) error {
	color := getSeverityColor(severity)

	message := SlackMessage{
		Text: fmt.Sprintf("🚨 新しいインシデントが作成されました: %s", incidentTitle),
		Blocks: []SlackBlock{
			{
				Type: "section",
				Text: &SlackText{
					Type: "mrkdwn",
					Text: fmt.Sprintf("*🚨 新しいインシデントが作成されました*\n*<%s|#%d %s>*",
						fmt.Sprintf("%s/incidents/%d", s.frontendURL, incidentID),
						incidentID,
						incidentTitle),
				},
			},
			{
				Type: "section",
				Fields: []SlackText{
					{Type: "mrkdwn", Text: fmt.Sprintf("*重要度:*\n%s", getSeverityEmoji(severity))},
					{Type: "mrkdwn", Text: fmt.Sprintf("*作成者:*\n%s", creatorName)},
				},
			},
		},
		Attachments: []Attachment{
			{
				Color:  color,
				Footer: "Incidex - Incident Management System",
			},
		},
	}

	return s.SendMessage(webhookURL, message)
}

// SendAssignedMessage は担当者割り当て通知を送信します
func (s *SlackService) SendAssignedMessage(webhookURL, incidentTitle string, incidentID uint, assigneeName, assignedBy string) error {
	message := SlackMessage{
		Text: fmt.Sprintf("👤 インシデントが割り当てられました: %s", incidentTitle),
		Blocks: []SlackBlock{
			{
				Type: "section",
				Text: &SlackText{
					Type: "mrkdwn",
					Text: fmt.Sprintf("*👤 インシデントが割り当てられました*\n*<%s|#%d %s>*",
						fmt.Sprintf("%s/incidents/%d", s.frontendURL, incidentID),
						incidentID,
						incidentTitle),
				},
			},
			{
				Type: "section",
				Fields: []SlackText{
					{Type: "mrkdwn", Text: fmt.Sprintf("*担当者:*\n%s", assigneeName)},
					{Type: "mrkdwn", Text: fmt.Sprintf("*割り当て者:*\n%s", assignedBy)},
				},
			},
		},
		Attachments: []Attachment{
			{
				Color:  "#36a64f",
				Footer: "Incidex - Incident Management System",
			},
		},
	}

	return s.SendMessage(webhookURL, message)
}

// SendCommentMessage はコメント追加通知を送信します
func (s *SlackService) SendCommentMessage(webhookURL, incidentTitle string, incidentID uint, commenterName, comment string) error {
	message := SlackMessage{
		Text: fmt.Sprintf("💬 新しいコメント: %s", incidentTitle),
		Blocks: []SlackBlock{
			{
				Type: "section",
				Text: &SlackText{
					Type: "mrkdwn",
					Text: fmt.Sprintf("*💬 新しいコメントが追加されました*\n*<%s|#%d %s>*",
						fmt.Sprintf("%s/incidents/%d", s.frontendURL, incidentID),
						incidentID,
						incidentTitle),
				},
			},
			{
				Type: "section",
				Text: &SlackText{
					Type: "mrkdwn",
					Text: fmt.Sprintf("*%s:*\n> %s", commenterName, comment),
				},
			},
		},
		Attachments: []Attachment{
			{
				Color:  "#3AA3E3",
				Footer: "Incidex - Incident Management System",
			},
		},
	}

	return s.SendMessage(webhookURL, message)
}

// SendStatusChangeMessage はステータス変更通知を送信します
func (s *SlackService) SendStatusChangeMessage(webhookURL, incidentTitle string, incidentID uint, oldStatus, newStatus string) error {
	message := SlackMessage{
		Text: fmt.Sprintf("🔄 ステータス変更: %s", incidentTitle),
		Blocks: []SlackBlock{
			{
				Type: "section",
				Text: &SlackText{
					Type: "mrkdwn",
					Text: fmt.Sprintf("*🔄 ステータスが変更されました*\n*<%s|#%d %s>*",
						fmt.Sprintf("%s/incidents/%d", s.frontendURL, incidentID),
						incidentID,
						incidentTitle),
				},
			},
			{
				Type: "section",
				Text: &SlackText{
					Type: "mrkdwn",
					Text: fmt.Sprintf("*変更:* %s → %s", getStatusText(oldStatus), getStatusText(newStatus)),
				},
			},
		},
		Attachments: []Attachment{
			{
				Color:  "#FFA500",
				Footer: "Incidex - Incident Management System",
			},
		},
	}

	return s.SendMessage(webhookURL, message)
}

// SendResolvedMessage はインシデント解決通知を送信します
func (s *SlackService) SendResolvedMessage(webhookURL, incidentTitle string, incidentID uint, resolvedBy string) error {
	message := SlackMessage{
		Text: fmt.Sprintf("✅ インシデントが解決されました: %s", incidentTitle),
		Blocks: []SlackBlock{
			{
				Type: "section",
				Text: &SlackText{
					Type: "mrkdwn",
					Text: fmt.Sprintf("*✅ インシデントが解決されました*\n*<%s|#%d %s>*",
						fmt.Sprintf("%s/incidents/%d", s.frontendURL, incidentID),
						incidentID,
						incidentTitle),
				},
			},
			{
				Type: "section",
				Fields: []SlackText{
					{Type: "mrkdwn", Text: fmt.Sprintf("*解決者:*\n%s", resolvedBy)},
				},
			},
		},
		Attachments: []Attachment{
			{
				Color:  "#36a64f",
				Footer: "Incidex - Incident Management System",
			},
		},
	}

	return s.SendMessage(webhookURL, message)
}

// SendSeverityChangeMessage は重要度変更通知を送信します
func (s *SlackService) SendSeverityChangeMessage(webhookURL, incidentTitle string, incidentID uint, oldSeverity, newSeverity string) error {
	newColor := getSeverityColor(newSeverity)

	message := SlackMessage{
		Text: fmt.Sprintf("⚠️ 重要度変更: %s", incidentTitle),
		Blocks: []SlackBlock{
			{
				Type: "section",
				Text: &SlackText{
					Type: "mrkdwn",
					Text: fmt.Sprintf("*⚠️ 重要度が変更されました*\n*<%s|#%d %s>*",
						fmt.Sprintf("%s/incidents/%d", s.frontendURL, incidentID),
						incidentID,
						incidentTitle),
				},
			},
			{
				Type: "section",
				Text: &SlackText{
					Type: "mrkdwn",
					Text: fmt.Sprintf("*変更:* %s → %s", getSeverityEmoji(oldSeverity), getSeverityEmoji(newSeverity)),
				},
			},
		},
		Attachments: []Attachment{
			{
				Color:  newColor,
				Footer: "Incidex - Incident Management System",
			},
		},
	}

	return s.SendMessage(webhookURL, message)
}

func getSeverityColor(severity string) string {
	switch severity {
	case "critical":
		return "#FF0000"
	case "high":
		return "#FF6B6B"
	case "medium":
		return "#FFA500"
	case "low":
		return "#4CAF50"
	default:
		return "#808080"
	}
}

func getSeverityEmoji(severity string) string {
	switch severity {
	case "critical":
		return "🔴 致命的"
	case "high":
		return "🟠 高"
	case "medium":
		return "🟡 中"
	case "low":
		return "🟢 低"
	default:
		return "⚪ 不明"
	}
}

func getStatusText(status string) string {
	switch status {
	case "open":
		return "未対応"
	case "investigating":
		return "調査中"
	case "resolved":
		return "解決済み"
	case "closed":
		return "完了"
	default:
		return status
	}
}
