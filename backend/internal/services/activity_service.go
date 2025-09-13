package services

import (
	"fmt"
	"gitlabex/internal/models"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ActivityService 活动服务
type ActivityService struct {
	db *gorm.DB
}

// NewActivityService 创建活动服务
func NewActivityService(db *gorm.DB) *ActivityService {
	return &ActivityService{
		db: db,
	}
}

// ActivityItem 活动项
type ActivityItem struct {
	ID          uuid.UUID `json:"id"`
	Type        string    `json:"type"`                   // document, topic, homework, comment
	Title       string    `json:"title"`                  // 活动标题
	Description string    `json:"description"`            // 活动描述
	UserName    string    `json:"user_name"`              // 操作用户名
	UserAvatar  string    `json:"user_avatar"`            // 用户头像
	ProjectName string    `json:"project_name,omitempty"` // 所属项目名称
	CreatedAt   time.Time `json:"created_at"`             // 创建时间
	URL         string    `json:"url"`                    // 跳转链接
}

// GetRecentActivities 获取最近活动
func (s *ActivityService) GetRecentActivities(limit int) ([]ActivityItem, error) {
	fmt.Printf("DEBUG: GetRecentActivities called with limit=%d\n", limit)
	var activities []ActivityItem

	// 获取最近的项目创建活动
	projectActivities, err := s.getRecentProjectActivities(limit / 5)
	if err != nil {
		return nil, err
	}
	activities = append(activities, projectActivities...)

	// 获取最近的文档上传活动
	documentActivities, err := s.getRecentDocumentActivities(limit / 5)
	if err != nil {
		return nil, err
	}
	activities = append(activities, documentActivities...)

	// 获取最近的话题讨论活动
	topicActivities, err := s.getRecentTopicActivities(limit / 5)
	if err != nil {
		return nil, err
	}
	activities = append(activities, topicActivities...)

	// 获取最近的作业发布活动
	homeworkActivities, err := s.getRecentHomeworkActivities(limit / 5)
	if err != nil {
		return nil, err
	}
	activities = append(activities, homeworkActivities...)

	// 获取最近的评论活动
	commentActivities, err := s.getRecentCommentActivities(limit / 5)
	if err != nil {
		return nil, err
	}
	activities = append(activities, commentActivities...)

	// 按时间排序并限制数量
	activities = s.sortAndLimitActivities(activities, limit)

	fmt.Printf("DEBUG: Returning %d activities\n", len(activities))
	return activities, nil
}

// getRecentDocumentActivities 获取最近的文档活动
func (s *ActivityService) getRecentDocumentActivities(limit int) ([]ActivityItem, error) {
	var documents []models.Document
	err := s.db.Preload("Project").
		Where("status = ?", models.DocumentStatusApproved).
		Order("created_at DESC").
		Limit(limit).
		Find(&documents).Error

	if err != nil {
		return nil, err
	}

	var activities []ActivityItem
	for _, doc := range documents {
		activity := ActivityItem{
			ID:          doc.ID,
			Type:        "document",
			Title:       "上传了文档",
			Description: doc.Title,
			UserName:    "GitLab用户", // 用户信息需要从GitLab API获取
			UserAvatar:  "",
			ProjectName: doc.Project.Name,
			CreatedAt:   doc.CreatedAt,
			URL:         "/documents/" + doc.ID.String(),
		}
		activities = append(activities, activity)
	}

	return activities, nil
}

// getRecentTopicActivities 获取最近的话题活动
func (s *ActivityService) getRecentTopicActivities(limit int) ([]ActivityItem, error) {
	var topics []models.Topic
	err := s.db.Preload("Project").
		Where("status = ?", "active").
		Order("created_at DESC").
		Limit(limit).
		Find(&topics).Error

	if err != nil {
		return nil, err
	}

	var activities []ActivityItem
	for _, topic := range topics {
		projectName := ""
		if topic.Project.Name != "" {
			projectName = topic.Project.Name
		}

		activity := ActivityItem{
			ID:          topic.ID,
			Type:        "topic",
			Title:       "发布了话题",
			Description: topic.Title,
			UserName:    "GitLab用户", // 用户信息需要从GitLab API获取
			UserAvatar:  "",
			ProjectName: projectName,
			CreatedAt:   topic.CreatedAt,
			URL:         "/topics/" + topic.ID.String(),
		}
		activities = append(activities, activity)
	}

	return activities, nil
}

// getRecentHomeworkActivities 获取最近的作业活动
func (s *ActivityService) getRecentHomeworkActivities(limit int) ([]ActivityItem, error) {
	var homeworks []models.Homework
	err := s.db.Preload("Project").
		Where("status = ?", models.HomeworkStatusPublished).
		Order("created_at DESC").
		Limit(limit).
		Find(&homeworks).Error

	if err != nil {
		return nil, err
	}

	var activities []ActivityItem
	for _, homework := range homeworks {
		activity := ActivityItem{
			ID:          homework.ID,
			Type:        "homework",
			Title:       "发布了作业",
			Description: homework.Title,
			UserName:    "GitLab用户", // 用户信息需要从GitLab API获取
			UserAvatar:  "",
			ProjectName: homework.Project.Name,
			CreatedAt:   homework.CreatedAt,
			URL:         "/homeworks/" + homework.ID.String(),
		}
		activities = append(activities, activity)
	}

	return activities, nil
}

// getRecentCommentActivities 获取最近的评论活动
func (s *ActivityService) getRecentCommentActivities(limit int) ([]ActivityItem, error) {
	var comments []models.Comment
	err := s.db.Preload("Topic").Preload("Topic.Project").
		Order("created_at DESC").
		Limit(limit).
		Find(&comments).Error

	if err != nil {
		return nil, err
	}

	var activities []ActivityItem
	for _, comment := range comments {
		projectName := ""
		if comment.Topic.Project.Name != "" {
			projectName = comment.Topic.Project.Name
		}

		activity := ActivityItem{
			ID:          comment.ID,
			Type:        "comment",
			Title:       "评论了话题",
			Description: comment.Topic.Title,
			UserName:    "GitLab用户", // 用户信息需要从GitLab API获取
			UserAvatar:  "",
			ProjectName: projectName,
			CreatedAt:   comment.CreatedAt,
			URL:         "/topics/" + comment.Topic.ID.String(),
		}
		activities = append(activities, activity)
	}

	return activities, nil
}

// sortAndLimitActivities 排序并限制活动数量
func (s *ActivityService) sortAndLimitActivities(activities []ActivityItem, limit int) []ActivityItem {
	// 按时间降序排序
	for i := 0; i < len(activities)-1; i++ {
		for j := i + 1; j < len(activities); j++ {
			if activities[i].CreatedAt.Before(activities[j].CreatedAt) {
				activities[i], activities[j] = activities[j], activities[i]
			}
		}
	}

	// 限制数量
	if len(activities) > limit {
		activities = activities[:limit]
	}

	return activities
}

// GetUserActivities 获取特定用户的活动
func (s *ActivityService) GetUserActivities(userID uuid.UUID, limit int) ([]ActivityItem, error) {
	var activities []ActivityItem

	// 获取用户的文档活动
	documentActivities, err := s.getUserDocumentActivities(userID, limit/3)
	if err != nil {
		return nil, err
	}
	activities = append(activities, documentActivities...)

	// 获取用户的话题活动
	topicActivities, err := s.getUserTopicActivities(userID, limit/3)
	if err != nil {
		return nil, err
	}
	activities = append(activities, topicActivities...)

	// 获取用户的作业活动
	homeworkActivities, err := s.getUserHomeworkActivities(userID, limit/3)
	if err != nil {
		return nil, err
	}
	activities = append(activities, homeworkActivities...)

	// 排序并限制数量
	activities = s.sortAndLimitActivities(activities, limit)

	return activities, nil
}

// getUserDocumentActivities 获取用户的文档活动
func (s *ActivityService) getUserDocumentActivities(userID uuid.UUID, limit int) ([]ActivityItem, error) {
	var documents []models.Document
	err := s.db.Preload("Project").
		Where("uploader_id = ? AND status = ?", userID, models.DocumentStatusApproved).
		Order("created_at DESC").
		Limit(limit).
		Find(&documents).Error

	if err != nil {
		return nil, err
	}

	var activities []ActivityItem
	for _, doc := range documents {
		activity := ActivityItem{
			ID:          doc.ID,
			Type:        "document",
			Title:       "上传了文档",
			Description: doc.Title,
			UserName:    "GitLab用户", // 用户信息需要从GitLab API获取
			UserAvatar:  "",
			ProjectName: doc.Project.Name,
			CreatedAt:   doc.CreatedAt,
			URL:         "/documents/" + doc.ID.String(),
		}
		activities = append(activities, activity)
	}

	return activities, nil
}

// getUserTopicActivities 获取用户的话题活动
func (s *ActivityService) getUserTopicActivities(userID uuid.UUID, limit int) ([]ActivityItem, error) {
	var topics []models.Topic
	err := s.db.Preload("Project").
		Where("author_id = ? AND status = ?", userID, "active").
		Order("created_at DESC").
		Limit(limit).
		Find(&topics).Error

	if err != nil {
		return nil, err
	}

	var activities []ActivityItem
	for _, topic := range topics {
		projectName := ""
		if topic.Project.Name != "" {
			projectName = topic.Project.Name
		}

		activity := ActivityItem{
			ID:          topic.ID,
			Type:        "topic",
			Title:       "发布了话题",
			Description: topic.Title,
			UserName:    "GitLab用户", // 用户信息需要从GitLab API获取
			UserAvatar:  "",
			ProjectName: projectName,
			CreatedAt:   topic.CreatedAt,
			URL:         "/topics/" + topic.ID.String(),
		}
		activities = append(activities, activity)
	}

	return activities, nil
}

// getUserHomeworkActivities 获取用户的作业活动
func (s *ActivityService) getUserHomeworkActivities(userID uuid.UUID, limit int) ([]ActivityItem, error) {
	var homeworks []models.Homework
	err := s.db.Preload("Project").
		Where("creator_id = ? AND status = ?", userID, models.HomeworkStatusPublished).
		Order("created_at DESC").
		Limit(limit).
		Find(&homeworks).Error

	if err != nil {
		return nil, err
	}

	var activities []ActivityItem
	for _, homework := range homeworks {
		activity := ActivityItem{
			ID:          homework.ID,
			Type:        "homework",
			Title:       "发布了作业",
			Description: homework.Title,
			UserName:    "GitLab用户", // 用户信息需要从GitLab API获取
			UserAvatar:  "",
			ProjectName: homework.Project.Name,
			CreatedAt:   homework.CreatedAt,
			URL:         "/homeworks/" + homework.ID.String(),
		}
		activities = append(activities, activity)
	}

	return activities, nil
}

// getRecentProjectActivities 获取最近的项目创建活动
func (s *ActivityService) getRecentProjectActivities(limit int) ([]ActivityItem, error) {
	var projects []models.ResearchProject
	err := s.db.Where("is_public = ?", true).
		Order("created_at DESC").
		Limit(limit).
		Find(&projects).Error

	if err != nil {
		return nil, err
	}

	// 调试日志
	fmt.Printf("DEBUG: Found %d projects\n", len(projects))

	var activities []ActivityItem
	for _, project := range projects {
		activity := ActivityItem{
			ID:          project.ID,
			Type:        "project",
			Title:       "创建了课题",
			Description: project.Name,
			UserName:    "GitLab用户", // 用户信息需要从GitLab API获取
			UserAvatar:  "",
			ProjectName: project.Name,
			CreatedAt:   project.CreatedAt,
			URL:         "/scenes/" + project.ID.String(),
		}
		activities = append(activities, activity)
	}

	return activities, nil
}
