package mocks

import (
	"incidex/internal/domain"

	"github.com/stretchr/testify/mock"
)

// MockAttachmentRepository は domain.AttachmentRepository のモック実装です
type MockAttachmentRepository struct {
	mock.Mock
}

func NewMockAttachmentRepository() *MockAttachmentRepository {
	return &MockAttachmentRepository{}
}

func (m *MockAttachmentRepository) Create(attachment *domain.Attachment) error {
	args := m.Called(attachment)
	return args.Error(0)
}

func (m *MockAttachmentRepository) FindByID(id uint) (*domain.Attachment, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Attachment), args.Error(1)
}

func (m *MockAttachmentRepository) FindByIncidentID(incidentID uint) ([]*domain.Attachment, error) {
	args := m.Called(incidentID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Attachment), args.Error(1)
}

func (m *MockAttachmentRepository) Delete(id uint) error {
	args := m.Called(id)
	return args.Error(0)
}
