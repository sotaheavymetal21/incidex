package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"incidex/internal/domain"
	"incidex/internal/testutil"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockTagUsecase は usecase.TagUsecase のモック実装です
type MockTagUsecase struct {
	mock.Mock
}

func NewMockTagUsecase() *MockTagUsecase {
	return &MockTagUsecase{}
}

func (m *MockTagUsecase) CreateTag(ctx context.Context, name, color string) (*domain.Tag, error) {
	args := m.Called(ctx, name, color)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Tag), args.Error(1)
}

func (m *MockTagUsecase) GetAllTags(ctx context.Context) ([]*domain.Tag, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Tag), args.Error(1)
}

func (m *MockTagUsecase) GetTagByID(ctx context.Context, id uint) (*domain.Tag, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Tag), args.Error(1)
}

func (m *MockTagUsecase) UpdateTag(ctx context.Context, id uint, name, color string) (*domain.Tag, error) {
	args := m.Called(ctx, id, name, color)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Tag), args.Error(1)
}

func (m *MockTagUsecase) DeleteTag(ctx context.Context, id uint) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func TestTagHandler_Create(t *testing.T) {
	t.Parallel()

	t.Run("successfully creates tag", func(t *testing.T) {
		t.Parallel()

		tagUsecase := NewMockTagUsecase()
		handler := NewTagHandler(tagUsecase)

		createdTag := testutil.NewTestTag(func(tag *domain.Tag) {
			tag.Name = "Production"
			tag.Color = "#ff0000"
		})

		tagUsecase.On("CreateTag", mock.Anything, "Production", "#ff0000").Return(createdTag, nil)

		reqBody := CreateTagRequest{Name: "Production", Color: "#ff0000"}
		body, _ := json.Marshal(reqBody)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/tags", bytes.NewBuffer(body))
		c.Request.Header.Set("Content-Type", "application/json")

		handler.Create(c)

		assert.Equal(t, http.StatusCreated, w.Code)

		var response domain.Tag
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "Production", response.Name)

		tagUsecase.AssertExpectations(t)
	})

	t.Run("fails with missing name", func(t *testing.T) {
		t.Parallel()

		tagUsecase := NewMockTagUsecase()
		handler := NewTagHandler(tagUsecase)

		reqBody := map[string]string{"color": "#ff0000"}
		body, _ := json.Marshal(reqBody)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/tags", bytes.NewBuffer(body))
		c.Request.Header.Set("Content-Type", "application/json")

		handler.Create(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		tagUsecase.AssertNotCalled(t, "CreateTag")
	})

	t.Run("fails when usecase returns error", func(t *testing.T) {
		t.Parallel()

		tagUsecase := NewMockTagUsecase()
		handler := NewTagHandler(tagUsecase)

		tagUsecase.On("CreateTag", mock.Anything, "Production", "#ff0000").
			Return(nil, domain.ErrConflict("Tag already exists"))

		reqBody := CreateTagRequest{Name: "Production", Color: "#ff0000"}
		body, _ := json.Marshal(reqBody)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/tags", bytes.NewBuffer(body))
		c.Request.Header.Set("Content-Type", "application/json")

		handler.Create(c)

		assert.Equal(t, http.StatusConflict, w.Code)
		tagUsecase.AssertExpectations(t)
	})
}

func TestTagHandler_GetAll(t *testing.T) {
	t.Parallel()

	t.Run("successfully returns all tags", func(t *testing.T) {
		t.Parallel()

		tagUsecase := NewMockTagUsecase()
		handler := NewTagHandler(tagUsecase)

		tags := []*domain.Tag{
			testutil.NewTestTag(func(t *domain.Tag) { t.ID = 1; t.Name = "Production" }),
			testutil.NewTestTag(func(t *domain.Tag) { t.ID = 2; t.Name = "Database" }),
		}

		tagUsecase.On("GetAllTags", mock.Anything).Return(tags, nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/tags", nil)

		handler.GetAll(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var response []*domain.Tag
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Len(t, response, 2)

		tagUsecase.AssertExpectations(t)
	})

	t.Run("returns empty list when no tags", func(t *testing.T) {
		t.Parallel()

		tagUsecase := NewMockTagUsecase()
		handler := NewTagHandler(tagUsecase)

		tagUsecase.On("GetAllTags", mock.Anything).Return([]*domain.Tag{}, nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/tags", nil)

		handler.GetAll(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var response []*domain.Tag
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Empty(t, response)

		tagUsecase.AssertExpectations(t)
	})

	t.Run("fails when usecase returns error", func(t *testing.T) {
		t.Parallel()

		tagUsecase := NewMockTagUsecase()
		handler := NewTagHandler(tagUsecase)

		tagUsecase.On("GetAllTags", mock.Anything).Return(nil, domain.ErrDatabase("database error", nil))

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/tags", nil)

		handler.GetAll(c)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		tagUsecase.AssertExpectations(t)
	})
}

func TestTagHandler_Update(t *testing.T) {
	t.Parallel()

	t.Run("successfully updates tag", func(t *testing.T) {
		t.Parallel()

		tagUsecase := NewMockTagUsecase()
		handler := NewTagHandler(tagUsecase)

		updatedTag := testutil.NewTestTag(func(tag *domain.Tag) {
			tag.ID = 1
			tag.Name = "Updated Tag"
			tag.Color = "#00ff00"
		})

		tagUsecase.On("UpdateTag", mock.Anything, uint(1), "Updated Tag", "#00ff00").Return(updatedTag, nil)

		reqBody := UpdateTagRequest{Name: "Updated Tag", Color: "#00ff00"}
		body, _ := json.Marshal(reqBody)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPut, "/api/v1/tags/1", bytes.NewBuffer(body))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Params = gin.Params{{Key: "id", Value: "1"}}

		handler.Update(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var response domain.Tag
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "Updated Tag", response.Name)

		tagUsecase.AssertExpectations(t)
	})

	t.Run("fails with invalid ID", func(t *testing.T) {
		t.Parallel()

		tagUsecase := NewMockTagUsecase()
		handler := NewTagHandler(tagUsecase)

		reqBody := UpdateTagRequest{Name: "Updated Tag", Color: "#00ff00"}
		body, _ := json.Marshal(reqBody)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPut, "/api/v1/tags/invalid", bytes.NewBuffer(body))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Params = gin.Params{{Key: "id", Value: "invalid"}}

		handler.Update(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		tagUsecase.AssertNotCalled(t, "UpdateTag")
	})

	t.Run("fails when tag not found", func(t *testing.T) {
		t.Parallel()

		tagUsecase := NewMockTagUsecase()
		handler := NewTagHandler(tagUsecase)

		tagUsecase.On("UpdateTag", mock.Anything, uint(999), "Updated Tag", "#00ff00").
			Return(nil, domain.ErrNotFound("Tag"))

		reqBody := UpdateTagRequest{Name: "Updated Tag", Color: "#00ff00"}
		body, _ := json.Marshal(reqBody)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPut, "/api/v1/tags/999", bytes.NewBuffer(body))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Params = gin.Params{{Key: "id", Value: "999"}}

		handler.Update(c)

		assert.Equal(t, http.StatusNotFound, w.Code)
		tagUsecase.AssertExpectations(t)
	})

	t.Run("fails with missing name", func(t *testing.T) {
		t.Parallel()

		tagUsecase := NewMockTagUsecase()
		handler := NewTagHandler(tagUsecase)

		reqBody := map[string]string{"color": "#00ff00"}
		body, _ := json.Marshal(reqBody)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPut, "/api/v1/tags/1", bytes.NewBuffer(body))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Params = gin.Params{{Key: "id", Value: "1"}}

		handler.Update(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		tagUsecase.AssertNotCalled(t, "UpdateTag")
	})
}

func TestTagHandler_Delete(t *testing.T) {
	t.Parallel()

	t.Run("successfully deletes tag", func(t *testing.T) {
		t.Parallel()

		tagUsecase := NewMockTagUsecase()
		handler := NewTagHandler(tagUsecase)

		tagUsecase.On("DeleteTag", mock.Anything, uint(1)).Return(nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodDelete, "/api/v1/tags/1", nil)
		c.Params = gin.Params{{Key: "id", Value: "1"}}

		handler.Delete(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Contains(t, response["message"], "deleted")

		tagUsecase.AssertExpectations(t)
	})

	t.Run("fails with invalid ID", func(t *testing.T) {
		t.Parallel()

		tagUsecase := NewMockTagUsecase()
		handler := NewTagHandler(tagUsecase)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodDelete, "/api/v1/tags/invalid", nil)
		c.Params = gin.Params{{Key: "id", Value: "invalid"}}

		handler.Delete(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		tagUsecase.AssertNotCalled(t, "DeleteTag")
	})

	t.Run("fails when tag not found", func(t *testing.T) {
		t.Parallel()

		tagUsecase := NewMockTagUsecase()
		handler := NewTagHandler(tagUsecase)

		tagUsecase.On("DeleteTag", mock.Anything, uint(999)).Return(domain.ErrNotFound("Tag"))

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodDelete, "/api/v1/tags/999", nil)
		c.Params = gin.Params{{Key: "id", Value: "999"}}

		handler.Delete(c)

		assert.Equal(t, http.StatusNotFound, w.Code)
		tagUsecase.AssertExpectations(t)
	})
}
