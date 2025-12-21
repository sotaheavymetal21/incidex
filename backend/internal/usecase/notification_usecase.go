package usecase

import (
	"errors"
	"incidex/internal/domain"
)

// NotificationUsecase は通知設定のユースケース
type NotificationUsecase struct {
	notificationRepo domain.NotificationSettingRepository
}

// NewNotificationUsecase は新しい通知設定ユースケースを作成します
func NewNotificationUsecase(notificationRepo domain.NotificationSettingRepository) *NotificationUsecase {
	return &NotificationUsecase{
		notificationRepo: notificationRepo,
	}
}

// GetSettingByUserID はユーザーIDで通知設定を取得します
func (u *NotificationUsecase) GetSettingByUserID(userID uint) (*domain.NotificationSetting, error) {
	setting, err := u.notificationRepo.GetByUserID(userID)
	if err != nil {
		// 設定がない場合はデフォルト設定を返す
		return &domain.NotificationSetting{
			UserID:                        userID,
			EmailEnabled:                  true,
			SlackEnabled:                  false,
			NotifyOnIncidentCreated:       true,
			NotifyOnAssigned:              true,
			NotifyOnComment:               true,
			NotifyOnStatusChange:          true,
			NotifyOnSeverityChange:        true,
			NotifyOnResolved:              true,
			NotifyOnEscalation:            true,
		}, nil
	}
	return setting, nil
}

// CreateSetting は新しい通知設定を作成します
func (u *NotificationUsecase) CreateSetting(setting *domain.NotificationSetting) error {
	if setting.UserID == 0 {
		return errors.New("user_id is required")
	}

	// 既存の設定があるかチェック
	existing, _ := u.notificationRepo.GetByUserID(setting.UserID)
	if existing != nil {
		return errors.New("notification setting already exists for this user")
	}

	return u.notificationRepo.Create(setting)
}

// UpdateSetting は通知設定を更新します
func (u *NotificationUsecase) UpdateSetting(userID uint, setting *domain.NotificationSetting) error {
	if userID == 0 {
		return errors.New("user_id is required")
	}

	// 既存の設定を取得
	existing, err := u.notificationRepo.GetByUserID(userID)
	if err != nil {
		// 設定がない場合は新規作成
		setting.UserID = userID
		return u.notificationRepo.Create(setting)
	}

	// IDとUserIDを保持して更新
	setting.ID = existing.ID
	setting.UserID = userID

	return u.notificationRepo.Update(setting)
}

// DeleteSetting は通知設定を削除します
func (u *NotificationUsecase) DeleteSetting(userID uint) error {
	return u.notificationRepo.Delete(userID)
}
