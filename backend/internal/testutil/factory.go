package testutil

import (
	"incidex/internal/domain"
	"time"
)

// Factory provides methods to create test data

// NewTestUser creates a new test user with default values
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

// NewTestAdmin creates a new test admin user
func NewTestAdmin(overrides ...func(*domain.User)) *domain.User {
	return NewTestUser(append([]func(*domain.User){
		func(u *domain.User) {
			u.Role = domain.RoleAdmin
			u.Email = "admin@example.com"
			u.Name = "Test Admin"
		},
	}, overrides...)...)
}

// NewTestEditor creates a new test editor user
func NewTestEditor(overrides ...func(*domain.User)) *domain.User {
	return NewTestUser(append([]func(*domain.User){
		func(u *domain.User) {
			u.Role = domain.RoleEditor
			u.Email = "editor@example.com"
			u.Name = "Test Editor"
		},
	}, overrides...)...)
}

// NewTestIncident creates a new test incident with default values
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

// NewTestCriticalIncident creates a critical severity test incident
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

// NewTestResolvedIncident creates a resolved test incident
func NewTestResolvedIncident(overrides ...func(*domain.Incident)) *domain.Incident {
	resolvedAt := time.Now()
	return NewTestIncident(append([]func(*domain.Incident){
		func(i *domain.Incident) {
			i.Status = domain.StatusResolved
			i.ResolvedAt = &resolvedAt
		},
	}, overrides...)...)
}

// NewTestTag creates a new test tag with default values
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

// NewTestRefreshToken creates a new test refresh token
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

// NewTestIncidentActivity creates a new test incident activity
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

// NewTestPostMortem creates a new test post mortem
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

// TimePtr returns a pointer to a time.Time value
func TimePtr(t time.Time) *time.Time {
	return &t
}

// UintPtr returns a pointer to a uint value
func UintPtr(u uint) *uint {
	return &u
}

// StringPtr returns a pointer to a string value
func StringPtr(s string) *string {
	return &s
}
