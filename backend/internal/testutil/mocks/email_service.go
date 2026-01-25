package mocks

import (
	"github.com/stretchr/testify/mock"
)

// MockEmailService は EmailService のモック実装です
type MockEmailService struct {
	mock.Mock
}

func NewMockEmailService() *MockEmailService {
	return &MockEmailService{}
}

func (m *MockEmailService) SendPasswordResetEmail(to, userName, resetToken, frontendURL string) error {
	args := m.Called(to, userName, resetToken, frontendURL)
	return args.Error(0)
}

func (m *MockEmailService) SendIncidentNotification(to, incidentTitle, incidentURL string) error {
	args := m.Called(to, incidentTitle, incidentURL)
	return args.Error(0)
}

func (m *MockEmailService) SendIncidentUpdateNotification(to, incidentTitle, updateMessage, incidentURL string) error {
	args := m.Called(to, incidentTitle, updateMessage, incidentURL)
	return args.Error(0)
}
