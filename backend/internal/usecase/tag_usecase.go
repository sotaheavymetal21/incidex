package usecase

import (
	"context"
	"incidex/internal/domain"
)

type TagUsecase interface {
	CreateTag(ctx context.Context, name, color string) (*domain.Tag, error)
	GetAllTags(ctx context.Context) ([]*domain.Tag, error)
	GetTagByID(ctx context.Context, id uint) (*domain.Tag, error)
	UpdateTag(ctx context.Context, id uint, name, color string) (*domain.Tag, error)
	DeleteTag(ctx context.Context, id uint) error
}

type tagUsecase struct {
	tagRepo domain.TagRepository
}

func NewTagUsecase(tagRepo domain.TagRepository) TagUsecase {
	return &tagUsecase{
		tagRepo: tagRepo,
	}
}

func (u *tagUsecase) CreateTag(ctx context.Context, name, color string) (*domain.Tag, error) {
	tag := &domain.Tag{
		Name:  name,
		Color: color,
	}
	if err := u.tagRepo.Create(tag); err != nil {
		return nil, err
	}
	return tag, nil
}

func (u *tagUsecase) GetAllTags(ctx context.Context) ([]*domain.Tag, error) {
	return u.tagRepo.FindAll()
}

func (u *tagUsecase) GetTagByID(ctx context.Context, id uint) (*domain.Tag, error) {
	return u.tagRepo.FindByID(id)
}

func (u *tagUsecase) UpdateTag(ctx context.Context, id uint, name, color string) (*domain.Tag, error) {
	tag, err := u.tagRepo.FindByID(id)
	if err != nil {
		return nil, err
	}

	tag.Name = name
	tag.Color = color

	if err := u.tagRepo.Update(tag); err != nil {
		return nil, err
	}
	return tag, nil
}

func (u *tagUsecase) DeleteTag(ctx context.Context, id uint) error {
	return u.tagRepo.Delete(id)
}
