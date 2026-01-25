package usecase

import (
	"context"
	"errors"
	"testing"

	"incidex/internal/domain"
	"incidex/internal/testutil"
	"incidex/internal/testutil/mocks"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestTagUsecase_CreateTag(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("successfully creates tag", func(t *testing.T) {
		t.Parallel()

		tagRepo := mocks.NewMockTagRepository()
		usecase := NewTagUsecase(tagRepo)

		tagRepo.On("Create", ctx, mock.AnythingOfType("*domain.Tag")).Return(nil)

		tag, err := usecase.CreateTag(ctx, "Production", "#ff0000")

		require.NoError(t, err)
		assert.NotNil(t, tag)
		assert.Equal(t, "Production", tag.Name)
		assert.Equal(t, "#ff0000", tag.Color)

		tagRepo.AssertExpectations(t)
	})

	t.Run("fails when repository returns error", func(t *testing.T) {
		t.Parallel()

		tagRepo := mocks.NewMockTagRepository()
		usecase := NewTagUsecase(tagRepo)

		tagRepo.On("Create", ctx, mock.AnythingOfType("*domain.Tag")).Return(errors.New("database error"))

		tag, err := usecase.CreateTag(ctx, "Production", "#ff0000")

		require.Error(t, err)
		assert.Nil(t, tag)

		domainErr, ok := domain.AsDomainError(err)
		require.True(t, ok)
		assert.Equal(t, domain.ErrCodeDatabaseError, domainErr.Code)

		tagRepo.AssertExpectations(t)
	})
}

func TestTagUsecase_GetAllTags(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("successfully returns all tags", func(t *testing.T) {
		t.Parallel()

		tagRepo := mocks.NewMockTagRepository()
		usecase := NewTagUsecase(tagRepo)

		expectedTags := []*domain.Tag{
			testutil.NewTestTag(func(t *domain.Tag) { t.ID = 1; t.Name = "Production" }),
			testutil.NewTestTag(func(t *domain.Tag) { t.ID = 2; t.Name = "Database" }),
		}

		tagRepo.On("FindAll", ctx).Return(expectedTags, nil)

		tags, err := usecase.GetAllTags(ctx)

		require.NoError(t, err)
		assert.Len(t, tags, 2)
		assert.Equal(t, "Production", tags[0].Name)
		assert.Equal(t, "Database", tags[1].Name)

		tagRepo.AssertExpectations(t)
	})

	t.Run("returns empty list when no tags exist", func(t *testing.T) {
		t.Parallel()

		tagRepo := mocks.NewMockTagRepository()
		usecase := NewTagUsecase(tagRepo)

		tagRepo.On("FindAll", ctx).Return([]*domain.Tag{}, nil)

		tags, err := usecase.GetAllTags(ctx)

		require.NoError(t, err)
		assert.Empty(t, tags)

		tagRepo.AssertExpectations(t)
	})

	t.Run("fails when repository returns error", func(t *testing.T) {
		t.Parallel()

		tagRepo := mocks.NewMockTagRepository()
		usecase := NewTagUsecase(tagRepo)

		tagRepo.On("FindAll", ctx).Return(nil, errors.New("database error"))

		tags, err := usecase.GetAllTags(ctx)

		require.Error(t, err)
		assert.Nil(t, tags)

		domainErr, ok := domain.AsDomainError(err)
		require.True(t, ok)
		assert.Equal(t, domain.ErrCodeDatabaseError, domainErr.Code)

		tagRepo.AssertExpectations(t)
	})
}

func TestTagUsecase_GetTagByID(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("successfully returns tag by ID", func(t *testing.T) {
		t.Parallel()

		tagRepo := mocks.NewMockTagRepository()
		usecase := NewTagUsecase(tagRepo)

		expectedTag := testutil.NewTestTag(func(t *domain.Tag) { t.ID = 1; t.Name = "Production" })
		tagRepo.On("FindByID", ctx, uint(1)).Return(expectedTag, nil)

		tag, err := usecase.GetTagByID(ctx, 1)

		require.NoError(t, err)
		assert.NotNil(t, tag)
		assert.Equal(t, uint(1), tag.ID)
		assert.Equal(t, "Production", tag.Name)

		tagRepo.AssertExpectations(t)
	})

	t.Run("fails when tag not found", func(t *testing.T) {
		t.Parallel()

		tagRepo := mocks.NewMockTagRepository()
		usecase := NewTagUsecase(tagRepo)

		tagRepo.On("FindByID", ctx, uint(999)).Return(nil, nil)

		tag, err := usecase.GetTagByID(ctx, 999)

		require.Error(t, err)
		assert.Nil(t, tag)

		domainErr, ok := domain.AsDomainError(err)
		require.True(t, ok)
		assert.Equal(t, domain.ErrCodeNotFound, domainErr.Code)

		tagRepo.AssertExpectations(t)
	})

	t.Run("fails when repository returns error", func(t *testing.T) {
		t.Parallel()

		tagRepo := mocks.NewMockTagRepository()
		usecase := NewTagUsecase(tagRepo)

		tagRepo.On("FindByID", ctx, uint(1)).Return(nil, errors.New("database error"))

		tag, err := usecase.GetTagByID(ctx, 1)

		require.Error(t, err)
		assert.Nil(t, tag)

		domainErr, ok := domain.AsDomainError(err)
		require.True(t, ok)
		assert.Equal(t, domain.ErrCodeDatabaseError, domainErr.Code)

		tagRepo.AssertExpectations(t)
	})
}

func TestTagUsecase_UpdateTag(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("successfully updates tag", func(t *testing.T) {
		t.Parallel()

		tagRepo := mocks.NewMockTagRepository()
		usecase := NewTagUsecase(tagRepo)

		existingTag := testutil.NewTestTag(func(t *domain.Tag) { t.ID = 1; t.Name = "OldName" })
		tagRepo.On("FindByID", ctx, uint(1)).Return(existingTag, nil)
		tagRepo.On("Update", ctx, mock.AnythingOfType("*domain.Tag")).Return(nil)

		tag, err := usecase.UpdateTag(ctx, 1, "NewName", "#00ff00")

		require.NoError(t, err)
		assert.NotNil(t, tag)
		assert.Equal(t, "NewName", tag.Name)
		assert.Equal(t, "#00ff00", tag.Color)

		tagRepo.AssertExpectations(t)
	})

	t.Run("fails when tag not found", func(t *testing.T) {
		t.Parallel()

		tagRepo := mocks.NewMockTagRepository()
		usecase := NewTagUsecase(tagRepo)

		tagRepo.On("FindByID", ctx, uint(999)).Return(nil, nil)

		tag, err := usecase.UpdateTag(ctx, 999, "NewName", "#00ff00")

		require.Error(t, err)
		assert.Nil(t, tag)

		domainErr, ok := domain.AsDomainError(err)
		require.True(t, ok)
		assert.Equal(t, domain.ErrCodeNotFound, domainErr.Code)

		tagRepo.AssertExpectations(t)
	})

	t.Run("fails when update returns error", func(t *testing.T) {
		t.Parallel()

		tagRepo := mocks.NewMockTagRepository()
		usecase := NewTagUsecase(tagRepo)

		existingTag := testutil.NewTestTag(func(t *domain.Tag) { t.ID = 1 })
		tagRepo.On("FindByID", ctx, uint(1)).Return(existingTag, nil)
		tagRepo.On("Update", ctx, mock.AnythingOfType("*domain.Tag")).Return(errors.New("database error"))

		tag, err := usecase.UpdateTag(ctx, 1, "NewName", "#00ff00")

		require.Error(t, err)
		assert.Nil(t, tag)

		domainErr, ok := domain.AsDomainError(err)
		require.True(t, ok)
		assert.Equal(t, domain.ErrCodeDatabaseError, domainErr.Code)

		tagRepo.AssertExpectations(t)
	})
}

func TestTagUsecase_DeleteTag(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("successfully deletes tag", func(t *testing.T) {
		t.Parallel()

		tagRepo := mocks.NewMockTagRepository()
		usecase := NewTagUsecase(tagRepo)

		existingTag := testutil.NewTestTag(func(t *domain.Tag) { t.ID = 1 })
		tagRepo.On("FindByID", ctx, uint(1)).Return(existingTag, nil)
		tagRepo.On("Delete", ctx, uint(1)).Return(nil)

		err := usecase.DeleteTag(ctx, 1)

		require.NoError(t, err)

		tagRepo.AssertExpectations(t)
	})

	t.Run("fails when tag not found", func(t *testing.T) {
		t.Parallel()

		tagRepo := mocks.NewMockTagRepository()
		usecase := NewTagUsecase(tagRepo)

		tagRepo.On("FindByID", ctx, uint(999)).Return(nil, nil)

		err := usecase.DeleteTag(ctx, 999)

		require.Error(t, err)

		domainErr, ok := domain.AsDomainError(err)
		require.True(t, ok)
		assert.Equal(t, domain.ErrCodeNotFound, domainErr.Code)

		tagRepo.AssertExpectations(t)
	})

	t.Run("fails when delete returns error", func(t *testing.T) {
		t.Parallel()

		tagRepo := mocks.NewMockTagRepository()
		usecase := NewTagUsecase(tagRepo)

		existingTag := testutil.NewTestTag(func(t *domain.Tag) { t.ID = 1 })
		tagRepo.On("FindByID", ctx, uint(1)).Return(existingTag, nil)
		tagRepo.On("Delete", ctx, uint(1)).Return(errors.New("database error"))

		err := usecase.DeleteTag(ctx, 1)

		require.Error(t, err)

		domainErr, ok := domain.AsDomainError(err)
		require.True(t, ok)
		assert.Equal(t, domain.ErrCodeDatabaseError, domainErr.Code)

		tagRepo.AssertExpectations(t)
	})
}
