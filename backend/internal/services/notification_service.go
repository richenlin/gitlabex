package services

import (
	"fmt"
	"gitlabex/internal/models"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// NotificationService 通知服务
type NotificationService struct {
	db *gorm.DB
}

// NewNotificationService 创建通知服务
func NewNotificationService(db *gorm.DB) *NotificationService {
	return &NotificationService{db: db}
}

// CreateNotification 创建通知
func (s *NotificationService) CreateNotification(notification *models.Notification) error {
	return s.db.Create(notification).Error
}

// CreateAnnouncement 创建公告并发送给目标用户
func (s *NotificationService) CreateAnnouncement(announcement *models.Announcement, targetUserIDs []uuid.UUID) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		// 保存公告
		if err := tx.Create(announcement).Error; err != nil {
			return err
		}

		// 为每个目标用户创建通知
		for _, userID := range targetUserIDs {
			notification := &models.Notification{
				Type:        models.NotificationTypeAnnouncement,
				Title:       announcement.Title,
				Content:     announcement.Content,
				RecipientID: userID,
				SenderID:    &announcement.AuthorID,
				ActionURL:   fmt.Sprintf("/announcements/%s", announcement.ID),
			}
			if err := tx.Create(notification).Error; err != nil {
				return err
			}
		}

		return nil
	})
}

// GetAnnouncementsWithPagination 分页获取公告列表
func (s *NotificationService) GetAnnouncementsWithPagination(announcements *[]models.Announcement, total *int64, limit, offset int) error {
	// 计算总数
	s.db.Model(&models.Announcement{}).Count(total)

	// 获取分页数据
	return s.db.Where("1=1").
		Where("is_active = true AND valid_from <= ? AND (valid_to IS NULL OR valid_to >= ?)", time.Now(), time.Now()).
		Order("created_at DESC").
		Limit(limit).Offset(offset).
		Find(announcements).Error
}

// GetAnnouncementByID 根据ID获取公告
func (s *NotificationService) GetAnnouncementByID(id uuid.UUID) (*models.Announcement, error) {
	var announcement models.Announcement
	err := s.db.Where("1=1").First(&announcement, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &announcement, nil
}

// GetUserNotifications 获取用户通知列表
func (s *NotificationService) GetUserNotifications(userID uuid.UUID, limit, offset int) ([]models.Notification, int64, error) {
	var notifications []models.Notification
	var total int64

	// 计算总数
	s.db.Model(&models.Notification{}).Where("recipient_id = ?", userID).Count(&total)

	// 获取分页数据
	err := s.db.Preload("Sender").Preload("Project").Preload("Topic").Preload("Homework").Preload("Document").
		Where("recipient_id = ?", userID).
		Order("created_at DESC").
		Limit(limit).Offset(offset).
		Find(&notifications).Error

	return notifications, total, err
}

// GetUnreadNotificationsCount 获取未读通知数量
func (s *NotificationService) GetUnreadNotificationsCount(userID uuid.UUID) (int64, error) {
	var count int64
	err := s.db.Model(&models.Notification{}).
		Where("recipient_id = ? AND is_read = false", userID).
		Count(&count).Error
	return count, err
}

// MarkAsRead 标记通知为已读
func (s *NotificationService) MarkAsRead(notificationID uuid.UUID, userID uuid.UUID) error {
	now := time.Now()
	return s.db.Model(&models.Notification{}).
		Where("id = ? AND recipient_id = ?", notificationID, userID).
		Updates(map[string]interface{}{
			"is_read": true,
			"read_at": now,
		}).Error
}

// MarkAllAsRead 标记所有通知为已读
func (s *NotificationService) MarkAllAsRead(userID uuid.UUID) error {
	now := time.Now()
	return s.db.Model(&models.Notification{}).
		Where("recipient_id = ? AND is_read = false", userID).
		Updates(map[string]interface{}{
			"is_read": true,
			"read_at": now,
		}).Error
}

// CreateProjectNotification 创建课题相关通知
func (s *NotificationService) CreateProjectNotification(projectID uuid.UUID, projectTitle string,
	notificationType models.NotificationType, userIDs []uuid.UUID, senderID uuid.UUID) error {

	var title, content string

	switch notificationType {
	case models.NotificationTypeProjectCreate:
		title = "新课题创建"
		content = fmt.Sprintf("课题 '%s' 已创建", projectTitle)
	case models.NotificationTypeProjectUpdate:
		title = "课题更新"
		content = fmt.Sprintf("课题 '%s' 已更新", projectTitle)
	default:
		return fmt.Errorf("unsupported notification type: %s", notificationType)
	}

	for _, userID := range userIDs {
		if userID == senderID {
			continue // 不给自己发通知
		}

		notification := &models.Notification{
			Type:        notificationType,
			Title:       title,
			Content:     content,
			RecipientID: userID,
			SenderID:    &senderID,
			ProjectID:   &projectID,
			ActionURL:   fmt.Sprintf("/research/%s", projectID),
		}

		if err := s.CreateNotification(notification); err != nil {
			return err
		}
	}

	return nil
}

// CreateTopicNotification 创建话题相关通知
func (s *NotificationService) CreateTopicNotification(topicID uuid.UUID, topicTitle string,
	notificationType models.NotificationType, userIDs []uuid.UUID, senderID uuid.UUID) error {

	var title, content string

	switch notificationType {
	case models.NotificationTypeTopicCreate:
		title = "新话题创建"
		content = fmt.Sprintf("话题 '%s' 已创建", topicTitle)
	case models.NotificationTypeTopicReply:
		title = "话题回复"
		content = fmt.Sprintf("话题 '%s' 有新回复", topicTitle)
	default:
		return fmt.Errorf("unsupported notification type: %s", notificationType)
	}

	for _, userID := range userIDs {
		if userID == senderID {
			continue
		}

		notification := &models.Notification{
			Type:        notificationType,
			Title:       title,
			Content:     content,
			RecipientID: userID,
			SenderID:    &senderID,
			TopicID:     &topicID,
			ActionURL:   fmt.Sprintf("/topics/%s", topicID),
		}

		if err := s.CreateNotification(notification); err != nil {
			return err
		}
	}

	return nil
}

// CreateHomeworkNotification 创建作业相关通知
func (s *NotificationService) CreateHomeworkNotification(homeworkID uuid.UUID, homeworkTitle string,
	notificationType models.NotificationType, userIDs []uuid.UUID, senderID uuid.UUID) error {

	var title, content string

	switch notificationType {
	case models.NotificationTypeHomeworkCreate:
		title = "新作业发布"
		content = fmt.Sprintf("作业 '%s' 已发布", homeworkTitle)
	case models.NotificationTypeHomeworkSubmit:
		title = "作业提交"
		content = fmt.Sprintf("作业 '%s' 有新提交", homeworkTitle)
	case models.NotificationTypeHomeworkGrade:
		title = "作业评分"
		content = fmt.Sprintf("作业 '%s' 已评分", homeworkTitle)
	default:
		return fmt.Errorf("unsupported notification type: %s", notificationType)
	}

	for _, userID := range userIDs {
		if userID == senderID {
			continue
		}

		notification := &models.Notification{
			Type:        notificationType,
			Title:       title,
			Content:     content,
			RecipientID: userID,
			SenderID:    &senderID,
			HomeworkID:  &homeworkID,
			ActionURL:   fmt.Sprintf("/homework/%s", homeworkID),
		}

		if err := s.CreateNotification(notification); err != nil {
			return err
		}
	}

	return nil
}
