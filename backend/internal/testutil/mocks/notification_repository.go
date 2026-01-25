package mocks

import (
	"incidex/internal/domain"

	"github.com/stretchr/testify/mock"
)

// MockNotificationSettingRepository は domain.NotificationSettingRepository のモック実装です
type MockNotificationSettingRepository struct {
	mock.Mock
}

func NewMockNotificationSettingRepository() *MockNotificationSettingRepository {
	return &MockNotificationSettingRepository{}
}

func (m *MockNotificationSettingRepository) Create(setting *domain.NotificationSetting) error {
	args := m.Called(setting)
	return args.Error(0)
}

func (m *MockNotificationSettingRepository) Update(setting *domain.NotificationSetting) error {
	args := m.Called(setting)
	return args.Error(0)
}

func (m *MockNotificationSettingRepository) GetByUserID(userID uint) (*domain.NotificationSetting, error) {
	args := m.Called(userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.NotificationSetting), args.Error(1)
}

func (m *MockNotificationSettingRepository) Delete(userID uint) error {
	args := m.Called(userID)
	return args.Error(0)
}
