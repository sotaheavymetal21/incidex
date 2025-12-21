package notification

import (
	"fmt"
	"incidex/internal/domain"
)

// NotificationService は通知を統合管理するサービス
type NotificationService struct {
	emailService *EmailService
	slackService *SlackService
	settingRepo  domain.NotificationSettingRepository
	userRepo     domain.UserRepository
}

// NewNotificationService は新しい通知サービスを作成します
func NewNotificationService(
	settingRepo domain.NotificationSettingRepository,
	userRepo domain.UserRepository,
) *NotificationService {
	return &NotificationService{
		emailService: NewEmailService(),
		slackService: NewSlackService(),
		settingRepo:  settingRepo,
		userRepo:     userRepo,
	}
}

// NotifyIncidentCreated はインシデント作成通知を送信します
func (s *NotificationService) NotifyIncidentCreated(incident *domain.Incident, creator *domain.User) error {
	// 担当者に通知
	if incident.AssigneeID != nil && *incident.AssigneeID != creator.ID {
		if err := s.notifyUser(*incident.AssigneeID, func(setting *domain.NotificationSetting, user *domain.User) error {
			if !setting.NotifyOnIncidentCreated {
				return nil
			}

			// Email通知
			if setting.EmailEnabled {
				if err := s.emailService.SendIncidentCreatedEmail(
					user.Email,
					incident.Title,
					incident.ID,
					string(incident.Severity),
				); err != nil {
					fmt.Printf("Failed to send email: %v\n", err)
				}
			}

			// Slack通知
			if setting.SlackEnabled && setting.SlackWebhook != "" {
				if err := s.slackService.SendIncidentCreatedMessage(
					setting.SlackWebhook,
					incident.Title,
					incident.ID,
					string(incident.Severity),
					creator.Name,
				); err != nil {
					fmt.Printf("Failed to send slack message: %v\n", err)
				}
			}

			return nil
		}); err != nil {
			return err
		}
	}

	return nil
}

// NotifyAssigned は担当者割り当て通知を送信します
func (s *NotificationService) NotifyAssigned(incident *domain.Incident, assignee *domain.User, assignedBy *domain.User) error {
	return s.notifyUser(assignee.ID, func(setting *domain.NotificationSetting, user *domain.User) error {
		if !setting.NotifyOnAssigned {
			return nil
		}

		// Email通知
		if setting.EmailEnabled {
			if err := s.emailService.SendAssignedEmail(
				user.Email,
				incident.Title,
				incident.ID,
				assignedBy.Name,
			); err != nil {
				fmt.Printf("Failed to send email: %v\n", err)
			}
		}

		// Slack通知
		if setting.SlackEnabled && setting.SlackWebhook != "" {
			if err := s.slackService.SendAssignedMessage(
				setting.SlackWebhook,
				incident.Title,
				incident.ID,
				assignee.Name,
				assignedBy.Name,
			); err != nil {
				fmt.Printf("Failed to send slack message: %v\n", err)
			}
		}

		return nil
	})
}

// NotifyComment はコメント追加通知を送信します
func (s *NotificationService) NotifyComment(incident *domain.Incident, commenter *domain.User, comment string) error {
	// 担当者に通知（コメント者本人以外）
	if incident.AssigneeID != nil && *incident.AssigneeID != commenter.ID {
		if err := s.notifyUser(*incident.AssigneeID, func(setting *domain.NotificationSetting, user *domain.User) error {
			if !setting.NotifyOnComment {
				return nil
			}

			// Email通知
			if setting.EmailEnabled {
				if err := s.emailService.SendCommentEmail(
					user.Email,
					incident.Title,
					incident.ID,
					commenter.Name,
					comment,
				); err != nil {
					fmt.Printf("Failed to send email: %v\n", err)
				}
			}

			// Slack通知
			if setting.SlackEnabled && setting.SlackWebhook != "" {
				if err := s.slackService.SendCommentMessage(
					setting.SlackWebhook,
					incident.Title,
					incident.ID,
					commenter.Name,
					comment,
				); err != nil {
					fmt.Printf("Failed to send slack message: %v\n", err)
				}
			}

			return nil
		}); err != nil {
			return err
		}
	}

	// 作成者に通知（コメント者本人と担当者以外）
	if incident.CreatorID != commenter.ID && (incident.AssigneeID == nil || incident.CreatorID != *incident.AssigneeID) {
		if err := s.notifyUser(incident.CreatorID, func(setting *domain.NotificationSetting, user *domain.User) error {
			if !setting.NotifyOnComment {
				return nil
			}

			// Email通知
			if setting.EmailEnabled {
				if err := s.emailService.SendCommentEmail(
					user.Email,
					incident.Title,
					incident.ID,
					commenter.Name,
					comment,
				); err != nil {
					fmt.Printf("Failed to send email: %v\n", err)
				}
			}

			// Slack通知
			if setting.SlackEnabled && setting.SlackWebhook != "" {
				if err := s.slackService.SendCommentMessage(
					setting.SlackWebhook,
					incident.Title,
					incident.ID,
					commenter.Name,
					comment,
				); err != nil {
					fmt.Printf("Failed to send slack message: %v\n", err)
				}
			}

			return nil
		}); err != nil {
			return err
		}
	}

	return nil
}

// NotifyStatusChange はステータス変更通知を送信します
func (s *NotificationService) NotifyStatusChange(incident *domain.Incident, oldStatus, newStatus string) error {
	userIDs := s.getInterestedUsers(incident)

	for _, userID := range userIDs {
		if err := s.notifyUser(userID, func(setting *domain.NotificationSetting, user *domain.User) error {
			if !setting.NotifyOnStatusChange {
				return nil
			}

			// Email通知
			if setting.EmailEnabled {
				if err := s.emailService.SendStatusChangeEmail(
					user.Email,
					incident.Title,
					incident.ID,
					oldStatus,
					newStatus,
				); err != nil {
					fmt.Printf("Failed to send email: %v\n", err)
				}
			}

			// Slack通知
			if setting.SlackEnabled && setting.SlackWebhook != "" {
				if err := s.slackService.SendStatusChangeMessage(
					setting.SlackWebhook,
					incident.Title,
					incident.ID,
					oldStatus,
					newStatus,
				); err != nil {
					fmt.Printf("Failed to send slack message: %v\n", err)
				}
			}

			return nil
		}); err != nil {
			return err
		}
	}

	return nil
}

// NotifyResolved はインシデント解決通知を送信します
func (s *NotificationService) NotifyResolved(incident *domain.Incident, resolver *domain.User) error {
	userIDs := s.getInterestedUsers(incident)

	for _, userID := range userIDs {
		if err := s.notifyUser(userID, func(setting *domain.NotificationSetting, user *domain.User) error {
			if !setting.NotifyOnResolved {
				return nil
			}

			// Email通知
			if setting.EmailEnabled {
				if err := s.emailService.SendResolvedEmail(
					user.Email,
					incident.Title,
					incident.ID,
					resolver.Name,
				); err != nil {
					fmt.Printf("Failed to send email: %v\n", err)
				}
			}

			// Slack通知
			if setting.SlackEnabled && setting.SlackWebhook != "" {
				if err := s.slackService.SendResolvedMessage(
					setting.SlackWebhook,
					incident.Title,
					incident.ID,
					resolver.Name,
				); err != nil {
					fmt.Printf("Failed to send slack message: %v\n", err)
				}
			}

			return nil
		}); err != nil {
			return err
		}
	}

	return nil
}

// notifyUser は指定ユーザーに通知を送信します
func (s *NotificationService) notifyUser(userID uint, fn func(*domain.NotificationSetting, *domain.User) error) error {
	// ユーザー取得
	user, err := s.userRepo.FindByID(nil, userID)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}

	// 通知設定取得
	setting, err := s.settingRepo.GetByUserID(userID)
	if err != nil {
		// 設定がない場合はデフォルト設定を使用
		setting = &domain.NotificationSetting{
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
		}
	}

	return fn(setting, user)
}

// getInterestedUsers はインシデントに関係するユーザーIDのリストを取得します
func (s *NotificationService) getInterestedUsers(incident *domain.Incident) []uint {
	userIDs := []uint{}

	// 作成者
	userIDs = append(userIDs, incident.CreatorID)

	// 担当者
	if incident.AssigneeID != nil {
		found := false
		for _, id := range userIDs {
			if id == *incident.AssigneeID {
				found = true
				break
			}
		}
		if !found {
			userIDs = append(userIDs, *incident.AssigneeID)
		}
	}

	return userIDs
}
