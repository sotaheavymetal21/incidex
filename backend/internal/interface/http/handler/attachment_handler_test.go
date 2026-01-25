package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"incidex/internal/domain"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockAttachmentUsecase は usecase.AttachmentUsecase のモック実装です
type MockAttachmentUsecase struct {
	mock.Mock
}

func NewMockAttachmentUsecase() *MockAttachmentUsecase {
	return &MockAttachmentUsecase{}
}

func (m *MockAttachmentUsecase) UploadAttachment(ctx context.Context, incidentID, userID uint, fileName string, fileSize int64, mimeType string, reader io.Reader) (*domain.Attachment, error) {
	args := m.Called(ctx, incidentID, userID, fileName, fileSize, mimeType, reader)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Attachment), args.Error(1)
}

func (m *MockAttachmentUsecase) GetAttachmentsByIncidentID(ctx context.Context, incidentID uint) ([]*domain.Attachment, error) {
	args := m.Called(ctx, incidentID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Attachment), args.Error(1)
}

func (m *MockAttachmentUsecase) GetAttachment(ctx context.Context, id uint) (*domain.Attachment, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Attachment), args.Error(1)
}

func (m *MockAttachmentUsecase) DownloadAttachment(ctx context.Context, id uint) (io.ReadCloser, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(io.ReadCloser), args.Error(1)
}

func (m *MockAttachmentUsecase) DeleteAttachment(ctx context.Context, id, userID uint, userRole domain.Role) error {
	args := m.Called(ctx, id, userID, userRole)
	return args.Error(0)
}

func TestAttachmentHandler_GetByIncidentID(t *testing.T) {
	t.Parallel()

	t.Run("successfully returns attachments", func(t *testing.T) {
		t.Parallel()

		mockUsecase := NewMockAttachmentUsecase()
		handler := NewAttachmentHandler(mockUsecase)

		attachments := []*domain.Attachment{
			{
				ID:         1,
				IncidentID: 1,
				FileName:   "test.pdf",
				FileSize:   1024,
				MimeType:   "application/pdf",
				CreatedAt:  time.Now(),
			},
		}

		mockUsecase.On("GetAttachmentsByIncidentID", mock.Anything, uint(1)).Return(attachments, nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/incidents/1/attachments", nil)
		c.Params = gin.Params{{Key: "id", Value: "1"}}

		handler.GetByIncidentID(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var response []*domain.Attachment
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Len(t, response, 1)
		assert.Equal(t, "test.pdf", response[0].FileName)
	})

	t.Run("fails with invalid incident ID", func(t *testing.T) {
		t.Parallel()

		mockUsecase := NewMockAttachmentUsecase()
		handler := NewAttachmentHandler(mockUsecase)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/incidents/invalid/attachments", nil)
		c.Params = gin.Params{{Key: "id", Value: "invalid"}}

		handler.GetByIncidentID(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("fails when usecase returns error", func(t *testing.T) {
		t.Parallel()

		mockUsecase := NewMockAttachmentUsecase()
		handler := NewAttachmentHandler(mockUsecase)

		mockUsecase.On("GetAttachmentsByIncidentID", mock.Anything, uint(1)).
			Return(nil, domain.ErrDatabase("database error", nil))

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/incidents/1/attachments", nil)
		c.Params = gin.Params{{Key: "id", Value: "1"}}

		handler.GetByIncidentID(c)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

func TestAttachmentHandler_Download(t *testing.T) {
	t.Parallel()

	t.Run("successfully downloads attachment", func(t *testing.T) {
		t.Parallel()

		mockUsecase := NewMockAttachmentUsecase()
		handler := NewAttachmentHandler(mockUsecase)

		attachment := &domain.Attachment{
			ID:       1,
			FileName: "test.pdf",
			FileSize: 1024,
			MimeType: "application/pdf",
		}

		fileContent := "test file content"
		reader := io.NopCloser(strings.NewReader(fileContent))

		mockUsecase.On("GetAttachment", mock.Anything, uint(1)).Return(attachment, nil)
		mockUsecase.On("DownloadAttachment", mock.Anything, uint(1)).Return(reader, nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/incidents/1/attachments/1/download", nil)
		c.Params = gin.Params{{Key: "attachmentId", Value: "1"}}

		handler.Download(c)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "application/pdf", w.Header().Get("Content-Type"))
		assert.Contains(t, w.Header().Get("Content-Disposition"), "test.pdf")
	})

	t.Run("fails with invalid attachment ID", func(t *testing.T) {
		t.Parallel()

		mockUsecase := NewMockAttachmentUsecase()
		handler := NewAttachmentHandler(mockUsecase)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/incidents/1/attachments/invalid/download", nil)
		c.Params = gin.Params{{Key: "attachmentId", Value: "invalid"}}

		handler.Download(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("fails when attachment not found", func(t *testing.T) {
		t.Parallel()

		mockUsecase := NewMockAttachmentUsecase()
		handler := NewAttachmentHandler(mockUsecase)

		mockUsecase.On("GetAttachment", mock.Anything, uint(999)).
			Return(nil, domain.ErrNotFound("attachment not found"))

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/incidents/1/attachments/999/download", nil)
		c.Params = gin.Params{{Key: "attachmentId", Value: "999"}}

		handler.Download(c)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}

func TestAttachmentHandler_Delete(t *testing.T) {
	t.Parallel()

	t.Run("admin can delete attachment", func(t *testing.T) {
		t.Parallel()

		mockUsecase := NewMockAttachmentUsecase()
		handler := NewAttachmentHandler(mockUsecase)

		mockUsecase.On("DeleteAttachment", mock.Anything, uint(1), uint(1), domain.RoleAdmin).Return(nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodDelete, "/api/incidents/1/attachments/1", nil)
		c.Params = gin.Params{
			{Key: "id", Value: "1"},
			{Key: "attachmentId", Value: "1"},
		}
		c.Set("userID", uint(1))
		c.Set("role", domain.RoleAdmin)

		handler.Delete(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]any
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Contains(t, response["message"], "deleted")
	})

	t.Run("fails without authentication", func(t *testing.T) {
		t.Parallel()

		mockUsecase := NewMockAttachmentUsecase()
		handler := NewAttachmentHandler(mockUsecase)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodDelete, "/api/incidents/1/attachments/1", nil)
		c.Params = gin.Params{
			{Key: "id", Value: "1"},
			{Key: "attachmentId", Value: "1"},
		}
		// No userID set

		handler.Delete(c)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("fails without role", func(t *testing.T) {
		t.Parallel()

		mockUsecase := NewMockAttachmentUsecase()
		handler := NewAttachmentHandler(mockUsecase)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodDelete, "/api/incidents/1/attachments/1", nil)
		c.Params = gin.Params{
			{Key: "id", Value: "1"},
			{Key: "attachmentId", Value: "1"},
		}
		c.Set("userID", uint(1))
		// No role set

		handler.Delete(c)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("fails with invalid incident ID", func(t *testing.T) {
		t.Parallel()

		mockUsecase := NewMockAttachmentUsecase()
		handler := NewAttachmentHandler(mockUsecase)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodDelete, "/api/incidents/invalid/attachments/1", nil)
		c.Params = gin.Params{
			{Key: "id", Value: "invalid"},
			{Key: "attachmentId", Value: "1"},
		}
		c.Set("userID", uint(1))
		c.Set("role", domain.RoleAdmin)

		handler.Delete(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("fails with invalid attachment ID", func(t *testing.T) {
		t.Parallel()

		mockUsecase := NewMockAttachmentUsecase()
		handler := NewAttachmentHandler(mockUsecase)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodDelete, "/api/incidents/1/attachments/invalid", nil)
		c.Params = gin.Params{
			{Key: "id", Value: "1"},
			{Key: "attachmentId", Value: "invalid"},
		}
		c.Set("userID", uint(1))
		c.Set("role", domain.RoleAdmin)

		handler.Delete(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("editor can delete own attachment", func(t *testing.T) {
		t.Parallel()

		mockUsecase := NewMockAttachmentUsecase()
		handler := NewAttachmentHandler(mockUsecase)

		mockUsecase.On("DeleteAttachment", mock.Anything, uint(1), uint(2), domain.RoleEditor).Return(nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodDelete, "/api/incidents/1/attachments/1", nil)
		c.Params = gin.Params{
			{Key: "id", Value: "1"},
			{Key: "attachmentId", Value: "1"},
		}
		c.Set("userID", uint(2))
		c.Set("role", domain.RoleEditor)

		handler.Delete(c)

		assert.Equal(t, http.StatusOK, w.Code)
	})
}

func TestAttachmentHandler_Upload(t *testing.T) {
	t.Parallel()

	t.Run("fails without authentication", func(t *testing.T) {
		t.Parallel()

		mockUsecase := NewMockAttachmentUsecase()
		handler := NewAttachmentHandler(mockUsecase)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/api/incidents/1/attachments", nil)
		c.Params = gin.Params{{Key: "id", Value: "1"}}
		// No userID set

		handler.Upload(c)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("fails with invalid incident ID", func(t *testing.T) {
		t.Parallel()

		mockUsecase := NewMockAttachmentUsecase()
		handler := NewAttachmentHandler(mockUsecase)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/api/incidents/invalid/attachments", nil)
		c.Params = gin.Params{{Key: "id", Value: "invalid"}}
		c.Set("userID", uint(1))

		handler.Upload(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("fails without file", func(t *testing.T) {
		t.Parallel()

		mockUsecase := NewMockAttachmentUsecase()
		handler := NewAttachmentHandler(mockUsecase)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		// Create request without file
		c.Request = httptest.NewRequest(http.MethodPost, "/api/incidents/1/attachments", bytes.NewBuffer(nil))
		c.Request.Header.Set("Content-Type", "multipart/form-data")
		c.Params = gin.Params{{Key: "id", Value: "1"}}
		c.Set("userID", uint(1))

		handler.Upload(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}
