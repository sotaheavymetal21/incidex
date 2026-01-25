package testutil

import (
	"incidex/internal/domain"
	"time"
)

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
		IncidentID:    incidentID,
		AuthorID:      authorID,
		RootCause:     "Test root cause",
		ImpactAnalysis: "Test impact analysis",
		WhatWentWell:  "Test what went well",
		WhatWentWrong: "Test what went wrong",
		LessonsLearned: "Test lessons learned",
		Status:        domain.PMStatusDraft,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
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
