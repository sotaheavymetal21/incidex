package testutil

import (
	"incidex/internal/domain"
	"incidex/internal/pkg/logger"
	"sync"
	"time"
)

var initLoggerOnce sync.Once

// InitTestLogger はテスト用のロガーを初期化します（一度だけ実行）
func InitTestLogger() {
	initLoggerOnce.Do(func() {
		_ = logger.InitLogger("test")
	})
}

// Factory はテストデータを作成するためのメソッドを提供します

// NewTestUser はデフォルト値を持つ新しいテストユーザーを作成します
func NewTestUser(overrides ...func(*domain.User)) *domain.User {
	user := &domain.User{
		ID:           1,
		Email:        "test@example.com",
		PasswordHash: "$2a$10$VkpvXhQrHnS/4gZi3Y2GC.CXX3RuPBRSyJsYhBsY2E2rX.n2.ZYfS", // "TestPassword1!"
		Name:         "Test User",
		Role:         domain.RoleViewer,
		IsActive:     true,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	for _, override := range overrides {
		override(user)
	}

	return user
}

// NewTestAdmin は新しいテスト管理者ユーザーを作成します
func NewTestAdmin(overrides ...func(*domain.User)) *domain.User {
	return NewTestUser(append([]func(*domain.User){
		func(u *domain.User) {
			u.Role = domain.RoleAdmin
			u.Email = "admin@example.com"
			u.Name = "Test Admin"
		},
	}, overrides...)...)
}

// NewTestEditor は新しいテスト編集者ユーザーを作成します
func NewTestEditor(overrides ...func(*domain.User)) *domain.User {
	return NewTestUser(append([]func(*domain.User){
		func(u *domain.User) {
			u.Role = domain.RoleEditor
			u.Email = "editor@example.com"
			u.Name = "Test Editor"
		},
	}, overrides...)...)
}

// NewTestIncident はデフォルト値を持つ新しいテストインシデントを作成します
func NewTestIncident(overrides ...func(*domain.Incident)) *domain.Incident {
	detectedAt := time.Now().Add(-1 * time.Hour)
	slaHours := domain.GetDefaultSLAHours(domain.SeverityMedium)
	slaDeadline := detectedAt.Add(time.Duration(slaHours) * time.Hour)

	incident := &domain.Incident{
		ID:                       1,
		Title:                    "Test Incident",
		Description:              "This is a test incident description",
		Severity:                 domain.SeverityMedium,
		Status:                   domain.StatusOpen,
		ImpactScope:              "Test environment",
		DetectedAt:               detectedAt,
		CreatorID:                1,
		SLATargetResolutionHours: slaHours,
		SLADeadline:              &slaDeadline,
		CreatedAt:                time.Now(),
		UpdatedAt:                time.Now(),
	}

	for _, override := range overrides {
		override(incident)
	}

	return incident
}

// NewTestCriticalIncident はクリティカル重要度のテストインシデントを作成します
func NewTestCriticalIncident(overrides ...func(*domain.Incident)) *domain.Incident {
	return NewTestIncident(append([]func(*domain.Incident){
		func(i *domain.Incident) {
			i.Severity = domain.SeverityCritical
			i.SLATargetResolutionHours = domain.GetDefaultSLAHours(domain.SeverityCritical)
			slaDeadline := i.DetectedAt.Add(time.Duration(i.SLATargetResolutionHours) * time.Hour)
			i.SLADeadline = &slaDeadline
		},
	}, overrides...)...)
}

// NewTestResolvedIncident は解決済みのテストインシデントを作成します
func NewTestResolvedIncident(overrides ...func(*domain.Incident)) *domain.Incident {
	resolvedAt := time.Now()
	return NewTestIncident(append([]func(*domain.Incident){
		func(i *domain.Incident) {
			i.Status = domain.StatusResolved
			i.ResolvedAt = &resolvedAt
		},
	}, overrides...)...)
}

// NewTestTag はデフォルト値を持つ新しいテストタグを作成します
func NewTestTag(overrides ...func(*domain.Tag)) *domain.Tag {
	tag := &domain.Tag{
		ID:        1,
		Name:      "Test Tag",
		Color:     "#10b981",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	for _, override := range overrides {
		override(tag)
	}

	return tag
}

// NewTestRefreshToken は新しいテスト refresh token を作成します
func NewTestRefreshToken(userID uint, overrides ...func(*domain.RefreshToken)) *domain.RefreshToken {
	token := &domain.RefreshToken{
		ID:        1,
		Token:     "test-refresh-token-12345",
		UserID:    userID,
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
		CreatedAt: time.Now(),
	}

	for _, override := range overrides {
		override(token)
	}

	return token
}

// NewTestIncidentActivity は新しいテストインシデントアクティビティを作成します
func NewTestIncidentActivity(incidentID, userID uint, overrides ...func(*domain.IncidentActivity)) *domain.IncidentActivity {
	activity := &domain.IncidentActivity{
		ID:           1,
		IncidentID:   incidentID,
		UserID:       userID,
		ActivityType: domain.ActivityTypeComment,
		Comment:      "Test comment",
		CreatedAt:    time.Now(),
	}

	for _, override := range overrides {
		override(activity)
	}

	return activity
}

// NewTestPostMortem は新しいテストポストモーテムを作成します
func NewTestPostMortem(incidentID, authorID uint, overrides ...func(*domain.PostMortem)) *domain.PostMortem {
	pm := &domain.PostMortem{
		ID:             1,
		IncidentID:     incidentID,
		AuthorID:       authorID,
		RootCause:      "Test root cause",
		ImpactAnalysis: "Test impact analysis",
		WhatWentWell:   "Test what went well",
		WhatWentWrong:  "Test what went wrong",
		LessonsLearned: "Test lessons learned",
		Status:         domain.PMStatusDraft,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	for _, override := range overrides {
		override(pm)
	}

	return pm
}

// NewTestPasswordResetToken は新しいテストパスワードリセットトークンを作成します
func NewTestPasswordResetToken(userID uint, overrides ...func(*domain.PasswordResetToken)) *domain.PasswordResetToken {
	token := &domain.PasswordResetToken{
		ID:        1,
		UserID:    userID,
		Token:     "test-reset-token-0123456789abcdef0123456789abcdef0123456789abcdef",
		ExpiresAt: time.Now().Add(domain.PasswordResetTokenExpiration),
		CreatedAt: time.Now(),
	}

	for _, override := range overrides {
		override(token)
	}

	return token
}

// TimePtr は time.Time 値へのポインタを返します
func TimePtr(t time.Time) *time.Time {
	return &t
}

// UintPtr は uint 値へのポインタを返します
func UintPtr(u uint) *uint {
	return &u
}

// StringPtr は string 値へのポインタを返します
func StringPtr(s string) *string {
	return &s
}

// NewTestNotificationSetting は新しいテスト通知設定を作成します
func NewTestNotificationSetting(userID uint, overrides ...func(*domain.NotificationSetting)) *domain.NotificationSetting {
	setting := &domain.NotificationSetting{
		ID:                      1,
		UserID:                  userID,
		EmailEnabled:            true,
		SlackEnabled:            false,
		SlackWebhook:            "",
		NotifyOnIncidentCreated: true,
		NotifyOnAssigned:        true,
		NotifyOnComment:         true,
		NotifyOnStatusChange:    true,
		NotifyOnSeverityChange:  true,
		NotifyOnResolved:        true,
		NotifyOnEscalation:      true,
		CreatedAt:               time.Now(),
		UpdatedAt:               time.Now(),
	}

	for _, override := range overrides {
		override(setting)
	}

	return setting
}

// NewTestActionItem は新しいテストアクションアイテムを作成します
func NewTestActionItem(postMortemID uint, overrides ...func(*domain.ActionItem)) *domain.ActionItem {
	dueDate := time.Now().Add(7 * 24 * time.Hour)
	item := &domain.ActionItem{
		ID:           1,
		PostMortemID: postMortemID,
		Title:        "Test Action Item",
		Description:  "Test action item description",
		AssigneeID:   nil,
		Priority:     domain.PriorityMedium,
		Status:       domain.ActionStatusPending,
		DueDate:      &dueDate,
		RelatedLinks: "[]",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	for _, override := range overrides {
		override(item)
	}

	return item
}

// NewTestAttachment は新しいテスト添付ファイルを作成します
func NewTestAttachment(incidentID, userID uint, overrides ...func(*domain.Attachment)) *domain.Attachment {
	attachment := &domain.Attachment{
		ID:         1,
		IncidentID: incidentID,
		UserID:     userID,
		FileName:   "test-file.pdf",
		FileSize:   1024,
		MimeType:   "application/pdf",
		StorageKey: "incidents/1/test-uuid.pdf",
		CreatedAt:  time.Now(),
	}

	for _, override := range overrides {
		override(attachment)
	}

	return attachment
}

// NewTestAuditLog は新しいテスト監査ログを作成します
func NewTestAuditLog(overrides ...func(*domain.AuditLog)) *domain.AuditLog {
	userID := uint(1)
	resourceID := uint(1)
	log := &domain.AuditLog{
		ID:           1,
		UserID:       &userID,
		UserName:     "Test User",
		UserEmail:    "test@example.com",
		Action:       domain.AuditActionCreate,
		ResourceType: "incident",
		ResourceID:   &resourceID,
		Method:       "POST",
		Path:         "/api/v1/incidents",
		IPAddress:    "127.0.0.1",
		UserAgent:    "Test Agent",
		StatusCode:   201,
		Details:      "{}",
		CreatedAt:    time.Now(),
	}

	for _, override := range overrides {
		override(log)
	}

	return log
}

// NewTestMonthlyReport は新しいテスト月次レポートを作成します
func NewTestMonthlyReport(overrides ...func(*domain.MonthlyReport)) *domain.MonthlyReport {
	now := time.Now()
	startDate := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	endDate := startDate.AddDate(0, 1, 0).Add(-time.Second)

	report := &domain.MonthlyReport{
		Period: domain.ReportPeriod{
			StartDate: startDate,
			EndDate:   endDate,
			Month:     int(now.Month()),
			Year:      now.Year(),
		},
		Summary: domain.IncidentSummary{
			TotalIncidents:    10,
			NewIncidents:      5,
			ResolvedIncidents: 3,
			OpenIncidents:     7,
			CriticalIncidents: 2,
		},
		SeverityBreakdown: map[string]int{
			"critical": 2,
			"high":     3,
			"medium":   3,
			"low":      2,
		},
		StatusBreakdown: map[string]int{
			"open":          4,
			"investigating": 3,
			"resolved":      3,
		},
		DailyTrend: []domain.DailyIncidentCount{
			{Date: startDate, Count: 2},
			{Date: startDate.AddDate(0, 0, 1), Count: 3},
		},
		TopTags: []domain.TagStatistic{
			{TagID: 1, TagName: "Production", Count: 5},
			{TagID: 2, TagName: "Database", Count: 3},
		},
		PerformanceMetrics: domain.PerformanceMetrics{
			AverageResolutionTime: 24.5,
		},
	}

	for _, override := range overrides {
		override(report)
	}

	return report
}
