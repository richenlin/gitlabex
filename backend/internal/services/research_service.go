package services

import (
	"gitlabex/internal/models"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ResearchService 研究课题服务
type ResearchService struct {
	*BaseService
}

// NewResearchService 创建研究课题服务
func NewResearchService(db *gorm.DB, gitlabService *GitLabService) *ResearchService {
	return &ResearchService{
		BaseService: NewBaseService(db, gitlabService.Config),
	}
}

// CreateResearchProject 创建研究课题
func (s *ResearchService) CreateResearchProject(project *models.ResearchProject) error {
	return s.DB.Create(project).Error
}

// GetResearchProjectByID 根据ID获取研究课题
func (s *ResearchService) GetResearchProjectByID(id uuid.UUID) (*models.ResearchProject, error) {
	var project models.ResearchProject
	// 移除了Creator和Members的预加载，因为这些信息现在从GitLab API获取
	err := s.DB.First(&project, "id = ?", id).Error
	return &project, err
}

// GetAllProjects 获取所有项目
func (s *ResearchService) GetAllProjects(limit, offset int, isPublic, includePrivate bool) ([]models.ResearchProject, int64, error) {
	var projects []models.ResearchProject
	var total int64

	query := s.DB.Model(&models.ResearchProject{})

	if !includePrivate {
		query = query.Where("is_public = ?", true)
	} else if isPublic {
		query = query.Where("is_public = ?", true)
	}

	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = query.Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&projects).Error

	return projects, total, err
}

// GetUserAccessibleProjectsByGitLabID 获取用户可访问的项目 (使用GitLab用户ID)
func (s *ResearchService) GetUserAccessibleProjectsByGitLabID(gitlabUserID int64, limit, offset int) ([]models.ResearchProject, int64, error) {
	var projects []models.ResearchProject
	var total int64

	// 注意：由于移除了本地成员管理，这里只返回公开项目和用户创建的项目
	// 具体的项目访问权限由GitLab API控制
	query := s.DB.Model(&models.ResearchProject{}).
		Where("is_public = ? OR creator_id = ?", true, gitlabUserID)

	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = query.Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&projects).Error

	return projects, total, err
}

// GetUserProjectsByGitLabID 获取用户创建的项目 (使用GitLab用户ID)
func (s *ResearchService) GetUserProjectsByGitLabID(gitlabUserID int64, limit, offset int) ([]models.ResearchProject, int64, error) {
	var projects []models.ResearchProject
	var total int64

	err := s.DB.Model(&models.ResearchProject{}).
		Where("creator_id = ?", gitlabUserID).
		Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = s.DB.Where("creator_id = ?", gitlabUserID).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&projects).Error

	return projects, total, err
}

// UpdateResearchProject 更新研究课题
func (s *ResearchService) UpdateResearchProject(id uuid.UUID, updates map[string]interface{}) error {
	return s.DB.Model(&models.ResearchProject{}).Where("id = ?", id).Updates(updates).Error
}

// DeleteResearchProject 删除研究课题
func (s *ResearchService) DeleteResearchProject(id uuid.UUID) error {
	return s.DB.Delete(&models.ResearchProject{}, "id = ?", id).Error
}

// 注意：本地成员管理方法已移除，成员管理完全使用GitLab API

// IsProjectOwnerByGitLabID 检查GitLab用户是否为项目所有者
func (s *ResearchService) IsProjectOwnerByGitLabID(projectID uuid.UUID, gitlabUserID int64) (bool, error) {
	var project models.ResearchProject
	err := s.DB.Select("creator_id").First(&project, "id = ?", projectID).Error
	if err != nil {
		return false, err
	}
	return project.CreatorID == gitlabUserID, nil
}

// 注意：GetProjectMemberRole已移除，角色信息从GitLab API获取

// GetProjectHomework 获取课题相关作业
func (s *ResearchService) GetProjectHomework(projectID uuid.UUID) ([]models.Homework, error) {
	var homeworks []models.Homework
	err := s.DB.Where("project_id = ?", projectID).Order("created_at DESC").Find(&homeworks).Error
	return homeworks, err
}

// CreateProjectHomework 创建课题作业
func (s *ResearchService) CreateProjectHomework(homework *models.Homework) error {
	return s.DB.Create(homework).Error
}

// GetProjectStats 获取项目统计信息
func (s *ResearchService) GetProjectStats(projectID uuid.UUID) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// 注意：成员数量统计已移除，成员信息从GitLab API获取

	// 作业数量
	var homeworkCount int64
	s.DB.Model(&models.Homework{}).Where("project_id = ?", projectID).Count(&homeworkCount)
	stats["homework_count"] = homeworkCount

	// 话题数量
	var topicCount int64
	s.DB.Model(&models.Topic{}).Where("project_id = ?", projectID).Count(&topicCount)
	stats["topic_count"] = topicCount

	// 文档数量
	var documentCount int64
	s.DB.Model(&models.Document{}).Where("project_id = ?", projectID).Count(&documentCount)
	stats["document_count"] = documentCount

	return stats, nil
}

// SearchProjects 搜索项目
func (s *ResearchService) SearchProjects(keyword string, limit, offset int) ([]models.ResearchProject, int64, error) {
	var projects []models.ResearchProject
	var total int64

	query := s.DB.Model(&models.ResearchProject{}).
		Where("title LIKE ? OR description LIKE ?", "%"+keyword+"%", "%"+keyword+"%")

	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = query.Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&projects).Error

	return projects, total, err
}

// GetHotProjects 获取热门项目
func (s *ResearchService) GetHotProjects(limit int) ([]models.ResearchProject, error) {
	var projects []models.ResearchProject
	err := s.DB.Where("is_public = ?", true).
		Order("created_at DESC").
		Limit(limit).
		Find(&projects).Error
	return projects, err
}

// GetProjectActivity 获取项目活跃度统计
func (s *ResearchService) GetProjectActivity(projectID uuid.UUID, days int) (map[string]interface{}, error) {
	activity := make(map[string]interface{})

	startDate := time.Now().AddDate(0, 0, -days)

	// 最近话题
	var recentTopics int64
	s.DB.Model(&models.Topic{}).
		Where("project_id = ? AND created_at >= ?", projectID, startDate).
		Count(&recentTopics)
	activity["recent_topics"] = recentTopics

	// 最近作业
	var recentHomework int64
	s.DB.Model(&models.Homework{}).
		Where("project_id = ? AND created_at >= ?", projectID, startDate).
		Count(&recentHomework)
	activity["recent_homework"] = recentHomework

	// 最近文档
	var recentDocuments int64
	s.DB.Model(&models.Document{}).
		Where("project_id = ? AND created_at >= ?", projectID, startDate).
		Count(&recentDocuments)
	activity["recent_documents"] = recentDocuments

	return activity, nil
}
