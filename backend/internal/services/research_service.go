package services

import (
	"gitlabex/internal/models"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ResearchService 研究课题服务
type ResearchService struct {
	db *gorm.DB
}

// NewResearchService 创建研究课题服务
func NewResearchService(db *gorm.DB, gitlabService *GitLabService) *ResearchService {
	return &ResearchService{
		db: db,
	}
}

// CreateResearchProject 创建研究课题
func (s *ResearchService) CreateResearchProject(project *models.ResearchProject) error {
	return s.db.Create(project).Error
}

// GetResearchProjectByID 根据ID获取研究课题
func (s *ResearchService) GetResearchProjectByID(id uuid.UUID) (*models.ResearchProject, error) {
	var project models.ResearchProject
	err := s.db.Preload("Creator").Preload("Members.User").First(&project, "id = ?", id).Error
	return &project, err
}

// GetAllProjects 获取所有项目
func (s *ResearchService) GetAllProjects(limit, offset int, isPublic, includePrivate bool) ([]models.ResearchProject, int64, error) {
	var projects []models.ResearchProject
	var total int64

	query := s.db.Model(&models.ResearchProject{})

	if !includePrivate {
		query = query.Where("is_public = ?", true)
	} else if isPublic {
		query = query.Where("is_public = ?", true)
	}

	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = query.Preload("Creator").
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&projects).Error

	return projects, total, err
}

// GetUserAccessibleProjects 获取用户可访问的项目
func (s *ResearchService) GetUserAccessibleProjects(userID uuid.UUID, limit, offset int) ([]models.ResearchProject, int64, error) {
	var projects []models.ResearchProject
	var total int64

	// 获取用户参与的项目ID
	var projectIDs []uuid.UUID
	s.db.Model(&models.ProjectMember{}).
		Where("user_id = ?", userID).
		Pluck("project_id", &projectIDs)

	// 获取公开项目或用户参与的项目
	query := s.db.Model(&models.ResearchProject{}).
		Where("is_public = ? OR id IN ?", true, projectIDs)

	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = query.Preload("Creator").
		Preload("Members.User").
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&projects).Error

	return projects, total, err
}

// GetUserProjects 获取用户创建的项目
func (s *ResearchService) GetUserProjects(userID uuid.UUID, limit, offset int) ([]models.ResearchProject, int64, error) {
	var projects []models.ResearchProject
	var total int64

	err := s.db.Model(&models.ResearchProject{}).
		Where("creator_id = ?", userID).
		Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = s.db.Preload("Creator").
		Preload("Members.User").
		Where("creator_id = ?", userID).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&projects).Error

	return projects, total, err
}

// UpdateResearchProject 更新研究课题
func (s *ResearchService) UpdateResearchProject(id uuid.UUID, updates map[string]interface{}) error {
	return s.db.Model(&models.ResearchProject{}).Where("id = ?", id).Updates(updates).Error
}

// DeleteResearchProject 删除研究课题
func (s *ResearchService) DeleteResearchProject(id uuid.UUID) error {
	return s.db.Delete(&models.ResearchProject{}, "id = ?", id).Error
}

// AddProjectMember 添加项目成员
func (s *ResearchService) AddProjectMember(projectID, userID uuid.UUID, role models.ProjectRole) error {
	member := models.ProjectMember{
		ProjectID: projectID,
		UserID:    userID,
		Role:      role,
	}
	return s.db.Create(&member).Error
}

// RemoveProjectMember 移除项目成员
func (s *ResearchService) RemoveProjectMember(projectID, userID uuid.UUID) error {
	return s.db.Where("project_id = ? AND user_id = ?", projectID, userID).
		Delete(&models.ProjectMember{}).Error
}

// GetProjectMembers 获取项目成员列表
func (s *ResearchService) GetProjectMembers(projectID uuid.UUID) ([]models.ProjectMember, error) {
	var members []models.ProjectMember
	err := s.db.Preload("User").Where("project_id = ?", projectID).Find(&members).Error
	return members, err
}

// IsProjectMember 检查用户是否为项目成员
func (s *ResearchService) IsProjectMember(projectID, userID uuid.UUID) (bool, error) {
	var count int64
	err := s.db.Model(&models.ProjectMember{}).
		Where("project_id = ? AND user_id = ?", projectID, userID).
		Count(&count).Error
	return count > 0, err
}

// IsProjectOwner 检查用户是否为项目所有者
func (s *ResearchService) IsProjectOwner(projectID, userID uuid.UUID) (bool, error) {
	var project models.ResearchProject
	err := s.db.Select("creator_id").First(&project, "id = ?", projectID).Error
	if err != nil {
		return false, err
	}
	return project.CreatorID == userID, nil
}

// GetProjectMemberRole 获取用户在项目中的角色
func (s *ResearchService) GetProjectMemberRole(projectID, userID uuid.UUID) (models.ProjectRole, error) {
	var member models.ProjectMember
	err := s.db.Where("project_id = ? AND user_id = ?", projectID, userID).First(&member).Error
	if err != nil {
		return models.ProjectRoleReporter, err
	}
	return member.Role, nil
}

// GetProjectHomework 获取课题相关作业
func (s *ResearchService) GetProjectHomework(projectID uuid.UUID) ([]models.Homework, error) {
	var homeworks []models.Homework
	err := s.db.Where("project_id = ?", projectID).Order("created_at DESC").Find(&homeworks).Error
	return homeworks, err
}

// CreateProjectHomework 创建课题作业
func (s *ResearchService) CreateProjectHomework(homework *models.Homework) error {
	return s.db.Create(homework).Error
}

// GetProjectStats 获取项目统计信息
func (s *ResearchService) GetProjectStats(projectID uuid.UUID) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// 成员数量
	var memberCount int64
	s.db.Model(&models.ProjectMember{}).Where("project_id = ?", projectID).Count(&memberCount)
	stats["member_count"] = memberCount

	// 作业数量
	var homeworkCount int64
	s.db.Model(&models.Homework{}).Where("project_id = ?", projectID).Count(&homeworkCount)
	stats["homework_count"] = homeworkCount

	// 话题数量
	var topicCount int64
	s.db.Model(&models.Topic{}).Where("project_id = ?", projectID).Count(&topicCount)
	stats["topic_count"] = topicCount

	// 文档数量
	var documentCount int64
	s.db.Model(&models.Document{}).Where("project_id = ?", projectID).Count(&documentCount)
	stats["document_count"] = documentCount

	return stats, nil
}

// SearchProjects 搜索项目
func (s *ResearchService) SearchProjects(keyword string, limit, offset int) ([]models.ResearchProject, int64, error) {
	var projects []models.ResearchProject
	var total int64

	query := s.db.Model(&models.ResearchProject{}).
		Where("title LIKE ? OR description LIKE ?", "%"+keyword+"%", "%"+keyword+"%")

	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = query.Preload("Creator").
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&projects).Error

	return projects, total, err
}

// GetHotProjects 获取热门项目
func (s *ResearchService) GetHotProjects(limit int) ([]models.ResearchProject, error) {
	var projects []models.ResearchProject
	err := s.db.Preload("Creator").
		Where("is_public = ?", true).
		Order("view_count DESC, created_at DESC").
		Limit(limit).
		Find(&projects).Error
	return projects, err
}

// IncrementViewCount 增加项目浏览次数
func (s *ResearchService) IncrementViewCount(projectID uuid.UUID) error {
	return s.db.Model(&models.ResearchProject{}).
		Where("id = ?", projectID).
		UpdateColumn("view_count", gorm.Expr("view_count + 1")).Error
}

// GetProjectActivity 获取项目活跃度统计
func (s *ResearchService) GetProjectActivity(projectID uuid.UUID, days int) (map[string]interface{}, error) {
	activity := make(map[string]interface{})

	startDate := time.Now().AddDate(0, 0, -days)

	// 最近话题
	var recentTopics int64
	s.db.Model(&models.Topic{}).
		Where("project_id = ? AND created_at >= ?", projectID, startDate).
		Count(&recentTopics)
	activity["recent_topics"] = recentTopics

	// 最近作业
	var recentHomework int64
	s.db.Model(&models.Homework{}).
		Where("project_id = ? AND created_at >= ?", projectID, startDate).
		Count(&recentHomework)
	activity["recent_homework"] = recentHomework

	// 最近文档
	var recentDocuments int64
	s.db.Model(&models.Document{}).
		Where("project_id = ? AND created_at >= ?", projectID, startDate).
		Count(&recentDocuments)
	activity["recent_documents"] = recentDocuments

	return activity, nil
}
