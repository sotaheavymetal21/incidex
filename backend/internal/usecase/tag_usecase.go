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
	if err := u.tagRepo.Create(ctx, tag); err != nil {
		return nil, domain.ErrDatabase("Failed to create tag", err)
	}
	return tag, nil
}

func (u *tagUsecase) GetAllTags(ctx context.Context) ([]*domain.Tag, error) {
	tags, err := u.tagRepo.FindAll(ctx)
	if err != nil {
		return nil, domain.ErrDatabase("Failed to fetch tags", err)
	}
	return tags, nil
}

func (u *tagUsecase) GetTagByID(ctx context.Context, id uint) (*domain.Tag, error) {
	tag, err := u.tagRepo.FindByID(ctx, id)
	if err != nil {
		return nil, domain.ErrDatabase("Failed to fetch tag", err)
	}
	if tag == nil {
		return nil, domain.ErrNotFound("Tag")
	}
	return tag, nil
}

func (u *tagUsecase) UpdateTag(ctx context.Context, id uint, name, color string) (*domain.Tag, error) {
	tag, err := u.tagRepo.FindByID(ctx, id)
	if err != nil {
		return nil, domain.ErrDatabase("Failed to fetch tag", err)
	}
	if tag == nil {
		return nil, domain.ErrNotFound("Tag")
	}

	tag.Name = name
	tag.Color = color

	if err := u.tagRepo.Update(ctx, tag); err != nil {
		return nil, domain.ErrDatabase("Failed to update tag", err)
	}
	return tag, nil
}

func (u *tagUsecase) DeleteTag(ctx context.Context, id uint) error {
	tag, err := u.tagRepo.FindByID(ctx, id)
	if err != nil {
		return domain.ErrDatabase("Failed to fetch tag", err)
	}
	if tag == nil {
		return domain.ErrNotFound("Tag")
	}

	if err := u.tagRepo.Delete(ctx, id); err != nil {
		return domain.ErrDatabase("Failed to delete tag", err)
	}
	return nil
}
