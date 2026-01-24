package persistence

import (
	"context"
	"incidex/internal/domain"

	"gorm.io/gorm"
)

type userLLMSettingRepository struct {
	db *gorm.DB
}

// NewUserLLMSettingRepository creates a new UserLLMSetting repository
func NewUserLLMSettingRepository(db *gorm.DB) domain.UserLLMSettingRepository {
	return &userLLMSettingRepository{db: db}
}

func (r *userLLMSettingRepository) Create(ctx context.Context, setting *domain.UserLLMSetting) error {
	return r.db.WithContext(ctx).Create(setting).Error
}

func (r *userLLMSettingRepository) Update(ctx context.Context, setting *domain.UserLLMSetting) error {
	return r.db.WithContext(ctx).Save(setting).Error
}

func (r *userLLMSettingRepository) Delete(ctx context.Context, userID uint) error {
	return r.db.WithContext(ctx).Where("user_id = ?", userID).Delete(&domain.UserLLMSetting{}).Error
}

func (r *userLLMSettingRepository) FindByUserID(ctx context.Context, userID uint) (*domain.UserLLMSetting, error) {
	var setting domain.UserLLMSetting
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).First(&setting).Error
	if err != nil {
		return nil, err
	}
	return &setting, nil
}

func (r *userLLMSettingRepository) FindAll(ctx context.Context) ([]*domain.UserLLMSetting, error) {
	var settings []*domain.UserLLMSetting
	err := r.db.WithContext(ctx).Find(&settings).Error
	if err != nil {
		return nil, err
	}
	return settings, nil
}
