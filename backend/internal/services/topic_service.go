package services

import (
	"gitlabex/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// TopicService 话题服务
type TopicService struct {
	db *gorm.DB
}

// NewTopicService 创建话题服务
func NewTopicService(db *gorm.DB, gitlabService *GitLabService) *TopicService {
	return &TopicService{
		db: db,
	}
}

// GetTopicsByProject 获取项目相关话题
func (s *TopicService) GetTopicsByProject(projectID uuid.UUID, limit, offset int) ([]models.Topic, int64, error) {
	var topics []models.Topic
	var total int64

	err := s.db.Model(&models.Topic{}).
		Where("project_id = ?", projectID).
		Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = s.db.Preload("Project").
		Where("project_id = ?", projectID).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&topics).Error

	return topics, total, err
}

// GetPublicTopics 获取公开话题
func (s *TopicService) GetPublicTopics(limit, offset int) ([]models.Topic, int64, error) {
	var topics []models.Topic
	var total int64

	err := s.db.Model(&models.Topic{}).
		Where("project_id IS NULL").
		Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = s.db.Where("project_id IS NULL").
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&topics).Error

	return topics, total, err
}

// GetTopicByID 根据ID获取话题
func (s *TopicService) GetTopicByID(id uuid.UUID) (*models.Topic, error) {
	var topic models.Topic
	err := s.db.Preload("Project").
		Preload("Likes").
		First(&topic, "id = ?", id).Error
	return &topic, err
}

// CreateTopic 创建话题
func (s *TopicService) CreateTopic(topic *models.Topic) error {
	return s.db.Create(topic).Error
}

// UpdateTopic 更新话题
func (s *TopicService) UpdateTopic(id uuid.UUID, updates map[string]interface{}) error {
	return s.db.Model(&models.Topic{}).Where("id = ?", id).Updates(updates).Error
}

// DeleteTopic 删除话题
func (s *TopicService) DeleteTopic(id uuid.UUID) error {
	return s.db.Delete(&models.Topic{}, "id = ?", id).Error
}

// CreateComment 创建评论
func (s *TopicService) CreateComment(comment *models.Comment) error {
	return s.db.Create(comment).Error
}

// GetCommentsByTopic 获取话题评论
func (s *TopicService) GetCommentsByTopic(topicID uuid.UUID) ([]models.Comment, error) {
	var comments []models.Comment
	err := s.db.Where("topic_id = ?", topicID).
		Order("created_at ASC").
		Find(&comments).Error
	return comments, err
}

// LikeTopic 点赞话题
func (s *TopicService) LikeTopic(like *models.TopicLike) error {
	return s.db.Create(like).Error
}

// UnlikeTopic 取消点赞
func (s *TopicService) UnlikeTopic(userID int64, topicID uuid.UUID) error {
	return s.db.Where("user_id = ? AND topic_id = ?", userID, topicID).
		Delete(&models.TopicLike{}).Error
}

// HasLikedTopic 检查用户是否已点赞话题
func (s *TopicService) HasLikedTopic(userID int64, topicID uuid.UUID) (bool, error) {
	var count int64
	err := s.db.Model(&models.TopicLike{}).
		Where("user_id = ? AND topic_id = ?", userID, topicID).
		Count(&count).Error
	return count > 0, err
}

// UpdateTopicLikesCount 更新话题点赞数
func (s *TopicService) UpdateTopicLikesCount(topicID uuid.UUID) error {
	var count int64
	err := s.db.Model(&models.TopicLike{}).
		Where("topic_id = ?", topicID).
		Count(&count).Error
	if err != nil {
		return err
	}

	return s.db.Model(&models.Topic{}).
		Where("id = ?", topicID).
		UpdateColumn("likes_count", count).Error
}

// SearchTopics 搜索话题
func (s *TopicService) SearchTopics(keyword string, limit, offset int) ([]models.Topic, int64, error) {
	var topics []models.Topic
	var total int64

	query := s.db.Model(&models.Topic{}).
		Where("title LIKE ? OR content LIKE ?", "%"+keyword+"%", "%"+keyword+"%")

	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = query.Preload("Project").
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&topics).Error

	return topics, total, err
}

// GetHotTopics 获取热门话题
func (s *TopicService) GetHotTopics(limit int) ([]models.Topic, error) {
	var topics []models.Topic
	err := s.db.Preload("Project").
		Order("likes_count DESC, created_at DESC").
		Limit(limit).
		Find(&topics).Error
	return topics, err
}

// GetRecentTopics 获取最近话题
func (s *TopicService) GetRecentTopics(limit int) ([]models.Topic, error) {
	var topics []models.Topic
	err := s.db.Preload("Project").
		Order("created_at DESC").
		Limit(limit).
		Find(&topics).Error
	return topics, err
}

// GetUserTopics 获取用户创建的话题
func (s *TopicService) GetUserTopics(userID int64, limit, offset int) ([]models.Topic, int64, error) {
	var topics []models.Topic
	var total int64

	err := s.db.Model(&models.Topic{}).
		Where("author_id = ?", userID).
		Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = s.db.Preload("Project").
		Where("author_id = ?", userID).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&topics).Error

	return topics, total, err
}

// GetTopicStats 获取话题统计信息
func (s *TopicService) GetTopicStats() (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// 总话题数
	var totalTopics int64
	s.db.Model(&models.Topic{}).Count(&totalTopics)
	stats["total_topics"] = totalTopics

	// 今日新话题
	var todayTopics int64
	s.db.Model(&models.Topic{}).
		Where("DATE(created_at) = DATE(NOW())").
		Count(&todayTopics)
	stats["today_topics"] = todayTopics

	// 本周新话题
	var weekTopics int64
	s.db.Model(&models.Topic{}).
		Where("created_at >= DATE_SUB(NOW(), INTERVAL 7 DAY)").
		Count(&weekTopics)
	stats["week_topics"] = weekTopics

	// 本月新话题
	var monthTopics int64
	s.db.Model(&models.Topic{}).
		Where("created_at >= DATE_SUB(NOW(), INTERVAL 30 DAY)").
		Count(&monthTopics)
	stats["month_topics"] = monthTopics

	return stats, nil
}
