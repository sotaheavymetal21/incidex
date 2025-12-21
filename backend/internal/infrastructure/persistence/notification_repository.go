package persistence

import (
	"incidex/internal/domain"

	"gorm.io/gorm"
)

type notificationSettingRepository struct {
	db *gorm.DB
}

// NewNotificationSettingRepository は新しい通知設定リポジトリを作成します
func NewNotificationSettingRepository(db *gorm.DB) domain.NotificationSettingRepository {
	return &notificationSettingRepository{db: db}
}

func (r *notificationSettingRepository) Create(setting *domain.NotificationSetting) error {
	return r.db.Create(setting).Error
}

func (r *notificationSettingRepository) Update(setting *domain.NotificationSetting) error {
	return r.db.Save(setting).Error
}

func (r *notificationSettingRepository) GetByUserID(userID uint) (*domain.NotificationSetting, error) {
	var setting domain.NotificationSetting
	err := r.db.Where("user_id = ?", userID).First(&setting).Error
	if err != nil {
		return nil, err
	}
	return &setting, nil
}

func (r *notificationSettingRepository) Delete(userID uint) error {
	return r.db.Where("user_id = ?", userID).Delete(&domain.NotificationSetting{}).Error
}
