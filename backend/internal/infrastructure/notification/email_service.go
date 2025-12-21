package notification

import (
	"fmt"
	"net/smtp"
	"os"
)

// EmailService はEmail通知を送信するサービス
type EmailService struct {
	smtpHost     string
	smtpPort     string
	smtpUsername string
	smtpPassword string
	fromAddress  string
}

// NewEmailService は新しいEmailサービスを作成します
func NewEmailService() *EmailService {
	return &EmailService{
		smtpHost:     getEnv("SMTP_HOST", "smtp.gmail.com"),
		smtpPort:     getEnv("SMTP_PORT", "587"),
		smtpUsername: os.Getenv("SMTP_USERNAME"),
		smtpPassword: os.Getenv("SMTP_PASSWORD"),
		fromAddress:  getEnv("SMTP_FROM", os.Getenv("SMTP_USERNAME")),
	}
}

// SendEmail はメールを送信します
func (s *EmailService) SendEmail(to, subject, body string) error {
	if s.smtpUsername == "" || s.smtpPassword == "" {
		// SMTP設定がない場合はログのみ出力（開発環境用）
		fmt.Printf("[EMAIL] To: %s, Subject: %s\n%s\n", to, subject, body)
		return nil
	}

	auth := smtp.PlainAuth("", s.smtpUsername, s.smtpPassword, s.smtpHost)

	msg := []byte(fmt.Sprintf(
		"From: %s\r\n"+
			"To: %s\r\n"+
			"Subject: %s\r\n"+
			"MIME-Version: 1.0\r\n"+
			"Content-Type: text/html; charset=UTF-8\r\n"+
			"\r\n"+
			"%s\r\n",
		s.fromAddress, to, subject, body,
	))

	addr := fmt.Sprintf("%s:%s", s.smtpHost, s.smtpPort)
	return smtp.SendMail(addr, auth, s.fromAddress, []string{to}, msg)
}

// SendIncidentCreatedEmail はインシデント作成通知を送信します
func (s *EmailService) SendIncidentCreatedEmail(to, incidentTitle string, incidentID uint, severity string) error {
	subject := fmt.Sprintf("[Incidex] 新しいインシデントが作成されました: %s", incidentTitle)

	body := fmt.Sprintf(`
		<html>
		<body>
			<h2>新しいインシデントが作成されました</h2>
			<p><strong>タイトル:</strong> %s</p>
			<p><strong>重要度:</strong> %s</p>
			<p><strong>インシデントID:</strong> #%d</p>
			<p><a href="http://localhost:3000/incidents/%d">詳細を見る</a></p>
		</body>
		</html>
	`, incidentTitle, severity, incidentID, incidentID)

	return s.SendEmail(to, subject, body)
}

// SendAssignedEmail は担当者割り当て通知を送信します
func (s *EmailService) SendAssignedEmail(to, incidentTitle string, incidentID uint, assignedBy string) error {
	subject := fmt.Sprintf("[Incidex] インシデントが割り当てられました: %s", incidentTitle)

	body := fmt.Sprintf(`
		<html>
		<body>
			<h2>新しいインシデントが割り当てられました</h2>
			<p><strong>タイトル:</strong> %s</p>
			<p><strong>割り当て者:</strong> %s</p>
			<p><strong>インシデントID:</strong> #%d</p>
			<p><a href="http://localhost:3000/incidents/%d">詳細を見る</a></p>
		</body>
		</html>
	`, incidentTitle, assignedBy, incidentID, incidentID)

	return s.SendEmail(to, subject, body)
}

// SendCommentEmail はコメント追加通知を送信します
func (s *EmailService) SendCommentEmail(to, incidentTitle string, incidentID uint, commenterName, comment string) error {
	subject := fmt.Sprintf("[Incidex] 新しいコメント: %s", incidentTitle)

	body := fmt.Sprintf(`
		<html>
		<body>
			<h2>新しいコメントが追加されました</h2>
			<p><strong>インシデント:</strong> %s</p>
			<p><strong>コメント者:</strong> %s</p>
			<p><strong>コメント:</strong></p>
			<blockquote>%s</blockquote>
			<p><a href="http://localhost:3000/incidents/%d">詳細を見る</a></p>
		</body>
		</html>
	`, incidentTitle, commenterName, comment, incidentID)

	return s.SendEmail(to, subject, body)
}

// SendStatusChangeEmail はステータス変更通知を送信します
func (s *EmailService) SendStatusChangeEmail(to, incidentTitle string, incidentID uint, oldStatus, newStatus string) error {
	subject := fmt.Sprintf("[Incidex] ステータス変更: %s", incidentTitle)

	body := fmt.Sprintf(`
		<html>
		<body>
			<h2>インシデントのステータスが変更されました</h2>
			<p><strong>インシデント:</strong> %s</p>
			<p><strong>変更:</strong> %s → %s</p>
			<p><a href="http://localhost:3000/incidents/%d">詳細を見る</a></p>
		</body>
		</html>
	`, incidentTitle, oldStatus, newStatus, incidentID)

	return s.SendEmail(to, subject, body)
}

// SendResolvedEmail はインシデント解決通知を送信します
func (s *EmailService) SendResolvedEmail(to, incidentTitle string, incidentID uint, resolvedBy string) error {
	subject := fmt.Sprintf("[Incidex] インシデントが解決されました: %s", incidentTitle)

	body := fmt.Sprintf(`
		<html>
		<body>
			<h2>インシデントが解決されました</h2>
			<p><strong>インシデント:</strong> %s</p>
			<p><strong>解決者:</strong> %s</p>
			<p><a href="http://localhost:3000/incidents/%d">詳細を見る</a></p>
		</body>
		</html>
	`, incidentTitle, resolvedBy, incidentID)

	return s.SendEmail(to, subject, body)
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
