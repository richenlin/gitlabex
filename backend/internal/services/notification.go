package services

import (
	"fmt"
	"time"

	"gitlabex/internal/models"

	"gorm.io/gorm"
)

// NotificationService 通知管理服务
type NotificationService struct {
	db                *gorm.DB
	permissionService *PermissionService
}

// NewNotificationService 创建通知管理服务
func NewNotificationService(db *gorm.DB, permissionService *PermissionService) *NotificationService {
	return &NotificationService{
		db:                db,
		permissionService: permissionService,
	}
}

// CreateNotificationRequest 创建通知请求
type CreateNotificationRequest struct {
	UserID     uint   `json:"user_id" binding:"required"`
	Title      string `json:"title" binding:"required"`
	Content    string `json:"content"`
	Type       string `json:"type" binding:"required"`
	TargetType string `json:"target_type"`
	TargetID   uint   `json:"target_id"`
}

// CreateNotification 创建通知
func (s *NotificationService) CreateNotification(notification *models.Notification) error {
	if err := s.db.Create(notification).Error; err != nil {
		return fmt.Errorf("failed to create notification: %w", err)
	}
	return nil
}

// CreateNotificationFromRequest 从请求创建通知
func (s *NotificationService) CreateNotificationFromRequest(req *CreateNotificationRequest) (*models.Notification, error) {
	notification := &models.Notification{
		UserID:     req.UserID,
		Title:      req.Title,
		Content:    req.Content,
		Type:       req.Type,
		TargetType: req.TargetType,
		TargetID:   req.TargetID,
		Read:       false,
	}

	if err := s.CreateNotification(notification); err != nil {
		return nil, err
	}

	// 预加载关联数据
	if err := s.db.Preload("User").First(notification, notification.ID).Error; err != nil {
		return nil, fmt.Errorf("failed to load notification: %w", err)
	}

	return notification, nil
}

// GetNotificationsByUser 获取用户的通知列表
func (s *NotificationService) GetNotificationsByUser(userID uint, page, pageSize int) ([]models.Notification, int64, error) {
	var notifications []models.Notification
	var total int64

	// 计算总数
	if err := s.db.Model(&models.Notification{}).Where("user_id = ?", userID).Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count notifications: %w", err)
	}

	// 分页查询
	offset := (page - 1) * pageSize
	err := s.db.Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(pageSize).
		Offset(offset).
		Find(&notifications).Error

	if err != nil {
		return nil, 0, fmt.Errorf("failed to get notifications: %w", err)
	}

	return notifications, total, nil
}

// GetUnreadNotifications 获取未读通知
func (s *NotificationService) GetUnreadNotifications(userID uint) ([]models.Notification, error) {
	var notifications []models.Notification
	err := s.db.Where("user_id = ? AND read = false", userID).
		Order("created_at DESC").
		Find(&notifications).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get unread notifications: %w", err)
	}

	return notifications, nil
}

// GetUnreadCount 获取未读通知数量
func (s *NotificationService) GetUnreadCount(userID uint) (int, error) {
	var count int64
	err := s.db.Model(&models.Notification{}).
		Where("user_id = ? AND read = false", userID).
		Count(&count).Error

	if err != nil {
		return 0, fmt.Errorf("failed to count unread notifications: %w", err)
	}

	return int(count), nil
}

// MarkAsRead 标记通知为已读
func (s *NotificationService) MarkAsRead(notificationID uint, userID uint) error {
	var notification models.Notification
	if err := s.db.Where("id = ? AND user_id = ?", notificationID, userID).First(&notification).Error; err != nil {
		return fmt.Errorf("notification not found: %w", err)
	}

	notification.MarkAsRead()
	if err := s.db.Save(&notification).Error; err != nil {
		return fmt.Errorf("failed to mark notification as read: %w", err)
	}

	return nil
}

// MarkAllAsRead 标记所有通知为已读
func (s *NotificationService) MarkAllAsRead(userID uint) error {
	now := time.Now()
	result := s.db.Model(&models.Notification{}).
		Where("user_id = ? AND read = false", userID).
		Updates(map[string]interface{}{
			"read":    true,
			"read_at": &now,
		})

	if result.Error != nil {
		return fmt.Errorf("failed to mark all notifications as read: %w", result.Error)
	}

	return nil
}

// DeleteNotification 删除通知
func (s *NotificationService) DeleteNotification(notificationID uint, userID uint) error {
	result := s.db.Where("id = ? AND user_id = ?", notificationID, userID).Delete(&models.Notification{})
	if result.Error != nil {
		return fmt.Errorf("failed to delete notification: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("notification not found")
	}

	return nil
}

// DeleteAllNotifications 删除所有通知
func (s *NotificationService) DeleteAllNotifications(userID uint) error {
	if err := s.db.Where("user_id = ?", userID).Delete(&models.Notification{}).Error; err != nil {
		return fmt.Errorf("failed to delete all notifications: %w", err)
	}

	return nil
}

// NotifyAssignmentSubmitted 通知作业提交
func (s *NotificationService) NotifyAssignmentSubmitted(submission *models.AssignmentSubmission) error {
	// 获取作业和老师信息
	var assignment models.Assignment
	if err := s.db.Preload("Teacher").Preload("Project").First(&assignment, submission.AssignmentID).Error; err != nil {
		return fmt.Errorf("failed to get assignment: %w", err)
	}

	// 获取学生信息
	var student models.User
	if err := s.db.First(&student, submission.StudentID).Error; err != nil {
		return fmt.Errorf("failed to get student: %w", err)
	}

	// 创建通知
	notification := &models.Notification{
		UserID:     assignment.TeacherID,
		Title:      "新作业提交",
		Content:    fmt.Sprintf("学生 %s 提交了作业「%s」", student.Name, assignment.Title),
		Type:       models.NotificationTypeAssignmentSubmitted,
		TargetType: "assignment",
		TargetID:   assignment.ID,
	}

	return s.CreateNotification(notification)
}

// NotifyAssignmentReviewed 通知作业已评审
func (s *NotificationService) NotifyAssignmentReviewed(review *models.Review) error {
	// 获取提交信息
	var submission models.AssignmentSubmission
	if err := s.db.Preload("Assignment").Preload("Student").First(&submission, review.SubmissionID).Error; err != nil {
		return fmt.Errorf("failed to get submission: %w", err)
	}

	// 创建通知
	notification := &models.Notification{
		UserID:     submission.StudentID,
		Title:      "作业评审完成",
		Content:    fmt.Sprintf("您的作业「%s」已完成评审，得分：%d", submission.Assignment.Title, review.Score),
		Type:       models.NotificationTypeAssignmentReviewed,
		TargetType: "assignment",
		TargetID:   submission.AssignmentID,
	}

	return s.CreateNotification(notification)
}

// NotifyAssignmentCreated 通知作业创建
func (s *NotificationService) NotifyAssignmentCreated(assignment *models.Assignment) error {
	// 获取课题成员
	var members []models.ProjectMember
	if err := s.db.Where("project_id = ? AND status = 'active'", assignment.ProjectID).Find(&members).Error; err != nil {
		return fmt.Errorf("failed to get project members: %w", err)
	}

	// 为每个成员创建通知
	for _, member := range members {
		notification := &models.Notification{
			UserID:     member.StudentID,
			Title:      "新作业发布",
			Content:    fmt.Sprintf("课题中发布了新作业「%s」，截止时间：%s", assignment.Title, assignment.DueDate.Format("2006-01-02 15:04")),
			Type:       models.NotificationTypeAssignmentCreated,
			TargetType: "assignment",
			TargetID:   assignment.ID,
		}

		if err := s.CreateNotification(notification); err != nil {
			// 记录错误但不中断流程
			continue
		}
	}

	return nil
}

// NotifyProjectJoined 通知课题加入
func (s *NotificationService) NotifyProjectJoined(projectID uint, studentID uint) error {
	// 获取课题信息
	var project models.Project
	if err := s.db.First(&project, projectID).Error; err != nil {
		return fmt.Errorf("failed to get project: %w", err)
	}

	// 获取学生信息
	var student models.User
	if err := s.db.First(&student, studentID).Error; err != nil {
		return fmt.Errorf("failed to get student: %w", err)
	}

	// 通知老师
	notification := &models.Notification{
		UserID:     project.TeacherID,
		Title:      "新成员加入课题",
		Content:    fmt.Sprintf("学生 %s 加入了课题「%s」", student.Name, project.Name),
		Type:       models.NotificationTypeProjectJoined,
		TargetType: "project",
		TargetID:   projectID,
	}

	return s.CreateNotification(notification)
}

// NotifyClassJoined 通知班级加入
func (s *NotificationService) NotifyClassJoined(classID uint, studentID uint) error {
	// 获取班级信息
	var class models.Class
	if err := s.db.First(&class, classID).Error; err != nil {
		return fmt.Errorf("failed to get class: %w", err)
	}

	// 获取学生信息
	var student models.User
	if err := s.db.First(&student, studentID).Error; err != nil {
		return fmt.Errorf("failed to get student: %w", err)
	}

	// 通知老师
	notification := &models.Notification{
		UserID:     class.TeacherID,
		Title:      "新学生加入班级",
		Content:    fmt.Sprintf("学生 %s 加入了班级「%s」", student.Name, class.Name),
		Type:       models.NotificationTypeClassJoined,
		TargetType: "class",
		TargetID:   classID,
	}

	return s.CreateNotification(notification)
}

// NotifyProjectCreated 通知课题创建
func (s *NotificationService) NotifyProjectCreated(project *models.Project) error {
	// 如果课题关联了班级，通知班级成员
	if project.ClassID != 0 {
		var members []models.ClassMember
		if err := s.db.Where("class_id = ? AND status = 'active'", project.ClassID).Find(&members).Error; err != nil {
			return fmt.Errorf("failed to get class members: %w", err)
		}

		// 为每个班级成员创建通知
		for _, member := range members {
			notification := &models.Notification{
				UserID:     member.StudentID,
				Title:      "新课题发布",
				Content:    fmt.Sprintf("班级中发布了新课题「%s」", project.Name),
				Type:       models.NotificationTypeProjectCreated,
				TargetType: "project",
				TargetID:   project.ID,
			}

			if err := s.CreateNotification(notification); err != nil {
				// 记录错误但不中断流程
				continue
			}
		}
	}

	return nil
}

// GetNotificationsByType 根据类型获取通知
func (s *NotificationService) GetNotificationsByType(userID uint, notificationType string) ([]models.Notification, error) {
	var notifications []models.Notification
	err := s.db.Where("user_id = ? AND type = ?", userID, notificationType).
		Order("created_at DESC").
		Find(&notifications).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get notifications by type: %w", err)
	}

	return notifications, nil
}

// GetNotificationStats 获取通知统计
func (s *NotificationService) GetNotificationStats(userID uint) (*NotificationStats, error) {
	stats := &NotificationStats{}

	// 总通知数
	var totalCount int64
	if err := s.db.Model(&models.Notification{}).Where("user_id = ?", userID).Count(&totalCount).Error; err != nil {
		return nil, fmt.Errorf("failed to count total notifications: %w", err)
	}
	stats.TotalCount = int(totalCount)

	// 未读通知数
	unreadCount, err := s.GetUnreadCount(userID)
	if err != nil {
		return nil, err
	}
	stats.UnreadCount = unreadCount

	// 今日通知数
	today := time.Now().Format("2006-01-02")
	var todayCount int64
	if err := s.db.Model(&models.Notification{}).
		Where("user_id = ? AND DATE(created_at) = ?", userID, today).
		Count(&todayCount).Error; err != nil {
		return nil, fmt.Errorf("failed to count today's notifications: %w", err)
	}
	stats.TodayCount = int(todayCount)

	// 按类型统计
	var typeStats []NotificationTypeStat
	if err := s.db.Model(&models.Notification{}).
		Select("type, COUNT(*) as count").
		Where("user_id = ?", userID).
		Group("type").
		Scan(&typeStats).Error; err != nil {
		return nil, fmt.Errorf("failed to get type stats: %w", err)
	}
	stats.TypeStats = typeStats

	return stats, nil
}

// NotificationStats 通知统计
type NotificationStats struct {
	TotalCount  int                    `json:"total_count"`
	UnreadCount int                    `json:"unread_count"`
	TodayCount  int                    `json:"today_count"`
	TypeStats   []NotificationTypeStat `json:"type_stats"`
}

// NotificationTypeStat 通知类型统计
type NotificationTypeStat struct {
	Type  string `json:"type"`
	Count int    `json:"count"`
}

// CleanupOldNotifications 清理旧通知
func (s *NotificationService) CleanupOldNotifications(days int) error {
	cutoffDate := time.Now().AddDate(0, 0, -days)

	result := s.db.Where("created_at < ?", cutoffDate).Delete(&models.Notification{})
	if result.Error != nil {
		return fmt.Errorf("failed to cleanup old notifications: %w", result.Error)
	}

	return nil
}
