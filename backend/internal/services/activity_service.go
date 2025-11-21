package services

import (
	"fmt"
	"gitlabex/internal/dto"
	"gitlabex/internal/models"

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

// GetRecentActivities 获取最近活动
func (s *ActivityService) GetRecentActivities(limit int) ([]dto.ActivityItem, error) {
	fmt.Printf("DEBUG: GetRecentActivities called with limit=%d\n", limit)
	var activities []dto.ActivityItem

	// 获取最近的项目创建活动
	projectActivities, err := s.getRecentProjectActivities(limit / 5)
	if err != nil {
		fmt.Printf("DEBUG: Error getting project activities: %v\n", err)
		// 不返回错误，继续获取其他类型的活动
	}
	activities = append(activities, projectActivities...)

	// 获取最近的文档上传活动
	documentActivities, err := s.getRecentDocumentActivities(limit / 5)
	if err != nil {
		fmt.Printf("DEBUG: Error getting document activities: %v\n", err)
		// 不返回错误，继续获取其他类型的活动
	}
	activities = append(activities, documentActivities...)

	// 获取最近的话题讨论活动
	topicActivities, err := s.getRecentTopicActivities(limit / 5)
	if err != nil {
		fmt.Printf("DEBUG: Error getting topic activities: %v\n", err)
		// 不返回错误，继续获取其他类型的活动
	}
	activities = append(activities, topicActivities...)

	// 获取最近的作业发布活动
	homeworkActivities, err := s.getRecentHomeworkActivities(limit / 5)
	if err != nil {
		fmt.Printf("DEBUG: Error getting homework activities: %v\n", err)
		// 不返回错误，继续获取其他类型的活动
	}
	activities = append(activities, homeworkActivities...)

	// 获取最近的评论活动
	commentActivities, err := s.getRecentCommentActivities(limit / 5)
	if err != nil {
		fmt.Printf("DEBUG: Error getting comment activities: %v\n", err)
		// 不返回错误，继续获取其他类型的活动
	}
	activities = append(activities, commentActivities...)

	// 按时间排序并限制数量
	activities = s.sortAndLimitActivities(activities, limit)

	fmt.Printf("DEBUG: Returning %d activities\n", len(activities))
	return activities, nil
}

// getRecentDocumentActivities 获取最近的文档活动
func (s *ActivityService) getRecentDocumentActivities(limit int) ([]dto.ActivityItem, error) {
	fmt.Printf("DEBUG: Getting document activities with limit=%d\n", limit)

	var documents []models.Document
	err := s.db.Preload("Project").
		Where("status = ?", models.DocumentStatusApproved).
		Order("created_at DESC").
		Limit(limit).
		Find(&documents).Error

	if err != nil {
		fmt.Printf("DEBUG: Error querying documents: %v\n", err)
		return nil, err
	}

	fmt.Printf("DEBUG: Found %d documents\n", len(documents))

	var activities []dto.ActivityItem
	for _, doc := range documents {
		projectName := ""
		// 安全地获取项目名称，如果Project关联失败则使用空字符串
		defer func() {
			if r := recover(); r != nil {
				fmt.Printf("DEBUG: Recovered from panic accessing doc.Project.Name: %v\n", r)
			}
		}()

		if doc.Project.Name != "" {
			projectName = doc.Project.Name
		}

		activity := dto.ActivityItem{
			ID:          doc.ID,
			Type:        "document",
			Title:       "上传了文档",
			Description: doc.Title,
			UserName:    "GitLab用户", // 用户信息需要从GitLab API获取
			UserAvatar:  "",
			ProjectName: projectName,
			CreatedAt:   doc.CreatedAt,
			URL:         "/documents/" + doc.ID.String(),
		}
		activities = append(activities, activity)
	}

	fmt.Printf("DEBUG: Returning %d document activities\n", len(activities))
	return activities, nil
}

// getRecentTopicActivities 获取最近的话题活动
func (s *ActivityService) getRecentTopicActivities(limit int) ([]dto.ActivityItem, error) {
	fmt.Printf("DEBUG: Getting topic activities with limit=%d\n", limit)

	var topics []models.Topic
	err := s.db.Preload("Project").
		Where("status = ?", "active").
		Order("created_at DESC").
		Limit(limit).
		Find(&topics).Error

	if err != nil {
		fmt.Printf("DEBUG: Error querying topics: %v\n", err)
		return nil, err
	}

	fmt.Printf("DEBUG: Found %d topics\n", len(topics))

	var activities []dto.ActivityItem
	for _, topic := range topics {
		projectName := ""
		defer func() {
			if r := recover(); r != nil {
				fmt.Printf("DEBUG: Recovered from panic accessing topic.Project.Name: %v\n", r)
			}
		}()

		if topic.Project.Name != "" {
			projectName = topic.Project.Name
		}

		activity := dto.ActivityItem{
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

	fmt.Printf("DEBUG: Returning %d topic activities\n", len(activities))
	return activities, nil
}

// getRecentHomeworkActivities 获取最近的作业活动
func (s *ActivityService) getRecentHomeworkActivities(limit int) ([]dto.ActivityItem, error) {
	fmt.Printf("DEBUG: Getting homework activities with limit=%d\n", limit)

	var homeworks []models.Homework
	err := s.db.Preload("Project").
		Where("status = ?", models.HomeworkStatusPublished).
		Order("created_at DESC").
		Limit(limit).
		Find(&homeworks).Error

	if err != nil {
		fmt.Printf("DEBUG: Error querying homeworks: %v\n", err)
		return nil, err
	}

	fmt.Printf("DEBUG: Found %d homeworks\n", len(homeworks))

	var activities []dto.ActivityItem
	for _, homework := range homeworks {
		projectName := ""
		defer func() {
			if r := recover(); r != nil {
				fmt.Printf("DEBUG: Recovered from panic accessing homework.Project.Name: %v\n", r)
			}
		}()

		if homework.Project.Name != "" {
			projectName = homework.Project.Name
		}

		activity := dto.ActivityItem{
			ID:          homework.ID,
			Type:        "homework",
			Title:       "发布了作业",
			Description: homework.Title,
			UserName:    "GitLab用户", // 用户信息需要从GitLab API获取
			UserAvatar:  "",
			ProjectName: projectName,
			CreatedAt:   homework.CreatedAt,
			URL:         "/homeworks/" + homework.ID.String(),
		}
		activities = append(activities, activity)
	}

	fmt.Printf("DEBUG: Returning %d homework activities\n", len(activities))
	return activities, nil
}

// getRecentCommentActivities 获取最近的评论活动
func (s *ActivityService) getRecentCommentActivities(limit int) ([]dto.ActivityItem, error) {
	fmt.Printf("DEBUG: Getting comment activities with limit=%d\n", limit)

	var comments []models.Comment
	err := s.db.Preload("Topic").Preload("Topic.Project").
		Order("created_at DESC").
		Limit(limit).
		Find(&comments).Error

	if err != nil {
		fmt.Printf("DEBUG: Error querying comments: %v\n", err)
		return nil, err
	}

	fmt.Printf("DEBUG: Found %d comments\n", len(comments))

	var activities []dto.ActivityItem
	for _, comment := range comments {
		projectName := ""
		topicTitle := ""
		topicID := ""

		defer func() {
			if r := recover(); r != nil {
				fmt.Printf("DEBUG: Recovered from panic accessing comment.Topic: %v\n", r)
			}
		}()

		if comment.Topic.Project.Name != "" {
			projectName = comment.Topic.Project.Name
		}

		topicTitle = comment.Topic.Title
		topicID = comment.Topic.ID.String()

		activity := dto.ActivityItem{
			ID:          comment.ID,
			Type:        "comment",
			Title:       "评论了话题",
			Description: topicTitle,
			UserName:    "GitLab用户", // 用户信息需要从GitLab API获取
			UserAvatar:  "",
			ProjectName: projectName,
			CreatedAt:   comment.CreatedAt,
			URL:         "/topics/" + topicID,
		}
		activities = append(activities, activity)
	}

	fmt.Printf("DEBUG: Returning %d comment activities\n", len(activities))
	return activities, nil
}

// sortAndLimitActivities 排序并限制活动数量
func (s *ActivityService) sortAndLimitActivities(activities []dto.ActivityItem, limit int) []dto.ActivityItem {
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
func (s *ActivityService) GetUserActivities(userID int64, limit int) ([]dto.ActivityItem, error) {
	var activities []dto.ActivityItem

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
func (s *ActivityService) getUserDocumentActivities(userID int64, limit int) ([]dto.ActivityItem, error) {
	var documents []models.Document
	err := s.db.Preload("Project").
		Where("uploader_id = ? AND status = ?", userID, models.DocumentStatusApproved).
		Order("created_at DESC").
		Limit(limit).
		Find(&documents).Error

	if err != nil {
		return nil, err
	}

	var activities []dto.ActivityItem
	for _, doc := range documents {
		projectName := ""
		defer func() {
			if r := recover(); r != nil {
				fmt.Printf("DEBUG: Recovered from panic accessing doc.Project.Name: %v\n", r)
			}
		}()

		if doc.Project.Name != "" {
			projectName = doc.Project.Name
		}

		activity := dto.ActivityItem{
			ID:          doc.ID,
			Type:        "document",
			Title:       "上传了文档",
			Description: doc.Title,
			UserName:    "GitLab用户", // 用户信息需要从GitLab API获取
			UserAvatar:  "",
			ProjectName: projectName,
			CreatedAt:   doc.CreatedAt,
			URL:         "/documents/" + doc.ID.String(),
		}
		activities = append(activities, activity)
	}

	return activities, nil
}

// getUserTopicActivities 获取用户的话题活动
func (s *ActivityService) getUserTopicActivities(userID int64, limit int) ([]dto.ActivityItem, error) {
	var topics []models.Topic
	err := s.db.Preload("Project").
		Where("author_id = ? AND status = ?", fmt.Sprintf("%d", userID), "active").
		Order("created_at DESC").
		Limit(limit).
		Find(&topics).Error

	if err != nil {
		return nil, err
	}

	var activities []dto.ActivityItem
	for _, topic := range topics {
		projectName := ""
		defer func() {
			if r := recover(); r != nil {
				fmt.Printf("DEBUG: Recovered from panic accessing topic.Project.Name: %v\n", r)
			}
		}()

		if topic.Project.Name != "" {
			projectName = topic.Project.Name
		}

		activity := dto.ActivityItem{
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
func (s *ActivityService) getUserHomeworkActivities(userID int64, limit int) ([]dto.ActivityItem, error) {
	var homeworks []models.Homework
	err := s.db.Preload("Project").
		Where("creator_id = ? AND status = ?", userID, models.HomeworkStatusPublished).
		Order("created_at DESC").
		Limit(limit).
		Find(&homeworks).Error

	if err != nil {
		return nil, err
	}

	var activities []dto.ActivityItem
	for _, homework := range homeworks {
		projectName := ""
		defer func() {
			if r := recover(); r != nil {
				fmt.Printf("DEBUG: Recovered from panic accessing homework.Project.Name: %v\n", r)
			}
		}()

		if homework.Project.Name != "" {
			projectName = homework.Project.Name
		}

		activity := dto.ActivityItem{
			ID:          homework.ID,
			Type:        "homework",
			Title:       "发布了作业",
			Description: homework.Title,
			UserName:    "GitLab用户", // 用户信息需要从GitLab API获取
			UserAvatar:  "",
			ProjectName: projectName,
			CreatedAt:   homework.CreatedAt,
			URL:         "/homeworks/" + homework.ID.String(),
		}
		activities = append(activities, activity)
	}

	return activities, nil
}

// getRecentProjectActivities 获取最近的项目创建活动
func (s *ActivityService) getRecentProjectActivities(limit int) ([]dto.ActivityItem, error) {
	fmt.Printf("DEBUG: Getting project activities with limit=%d\n", limit)

	var projects []models.ResearchProject
	err := s.db.Where("is_public = ?", true).
		Order("created_at DESC").
		Limit(limit).
		Find(&projects).Error

	if err != nil {
		fmt.Printf("DEBUG: Error querying projects: %v\n", err)
		return nil, err
	}

	// 调试日志
	fmt.Printf("DEBUG: Found %d projects\n", len(projects))

	var activities []dto.ActivityItem
	for _, project := range projects {
		activity := dto.ActivityItem{
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

	fmt.Printf("DEBUG: Returning %d project activities\n", len(activities))
	return activities, nil
}
