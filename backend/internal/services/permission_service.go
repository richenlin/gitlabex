package services

import (
	"errors"
	"gitlabex/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// PermissionService 权限管理服务
type PermissionService struct {
	db *gorm.DB
}

// NewPermissionService 创建权限服务
func NewPermissionService(db *gorm.DB) *PermissionService {
	return &PermissionService{
		db: db,
	}
}

// ProjectPermission 项目权限类型
type ProjectPermission string

const (
	ProjectPermissionView   ProjectPermission = "view"
	ProjectPermissionEdit   ProjectPermission = "edit"
	ProjectPermissionDelete ProjectPermission = "delete"
	ProjectPermissionManage ProjectPermission = "manage"
	ProjectPermissionCreate ProjectPermission = "create"
	ProjectPermissionUpload ProjectPermission = "upload"
	ProjectPermissionAdmin  ProjectPermission = "admin"
)

// CheckProjectPermission 检查用户在项目中的权限
func (s *PermissionService) CheckProjectPermission(userID, projectID uuid.UUID, permission ProjectPermission) (bool, error) {
	// 获取用户信息
	var user models.User
	if err := s.db.First(&user, "id = ?", userID).Error; err != nil {
		return false, err
	}

	// 管理员拥有所有权限
	if user.EduRole >= models.EduRoleAdmin {
		return true, nil
	}

	// 获取项目信息
	var project models.ResearchProject
	if err := s.db.First(&project, "id = ?", projectID).Error; err != nil {
		return false, err
	}

	// 项目创建者拥有所有权限
	if project.CreatorID == userID {
		return true, nil
	}

	// 公开项目，所有人都可以查看
	if project.IsPublic && permission == ProjectPermissionView {
		return true, nil
	}

	// 获取用户在项目中的角色
	var member models.ProjectMember
	err := s.db.Where("project_id = ? AND user_id = ?", projectID, userID).First(&member).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil // 用户不是项目成员
		}
		return false, err
	}

	// 根据用户角色和权限类型判断是否允许
	return s.hasPermissionForRole(member.Role, permission), nil
}

// hasPermissionForRole 根据角色判断是否有特定权限
func (s *PermissionService) hasPermissionForRole(role models.ProjectRole, permission ProjectPermission) bool {
	// 项目所有者拥有所有权限
	if role == models.ProjectRoleOwner {
		return true
	}

	// 维护者权限
	if role == models.ProjectRoleMaintainer {
		switch permission {
		case ProjectPermissionView, ProjectPermissionEdit, ProjectPermissionCreate, ProjectPermissionUpload:
			return true
		default:
			return false
		}
	}

	// 开发者权限
	if role == models.ProjectRoleDeveloper {
		switch permission {
		case ProjectPermissionView, ProjectPermissionEdit, ProjectPermissionCreate, ProjectPermissionUpload:
			return true
		default:
			return false
		}
	}

	// 报告者权限（只读权限）
	if role == models.ProjectRoleReporter {
		switch permission {
		case ProjectPermissionView:
			return true
		default:
			return false
		}
	}

	return false
}

// GetUserProjectRole 获取用户在项目中的角色
func (s *PermissionService) GetUserProjectRole(userID, projectID uuid.UUID) (models.ProjectRole, error) {
	var member models.ProjectMember
	err := s.db.Where("project_id = ? AND user_id = ?", projectID, userID).First(&member).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return models.ProjectRoleReporter, nil // 默认角色
		}
		return models.ProjectRoleReporter, err
	}
	return member.Role, nil
}

// GetUserProjectPermissions 获取用户在项目中的所有权限
func (s *PermissionService) GetUserProjectPermissions(userID, projectID uuid.UUID) ([]ProjectPermission, error) {
	permissions := []ProjectPermission{}

	allPermissions := []ProjectPermission{
		ProjectPermissionView,
		ProjectPermissionEdit,
		ProjectPermissionDelete,
		ProjectPermissionManage,
		ProjectPermissionCreate,
		ProjectPermissionUpload,
		ProjectPermissionAdmin,
	}

	for _, permission := range allPermissions {
		hasPerm, err := s.CheckProjectPermission(userID, projectID, permission)
		if err != nil {
			return nil, err
		}
		if hasPerm {
			permissions = append(permissions, permission)
		}
	}

	return permissions, nil
}

// CheckDocumentPermission 检查用户对文档的权限
func (s *PermissionService) CheckDocumentPermission(userID, documentID uuid.UUID, permission ProjectPermission) (bool, error) {
	// 获取文档信息
	var document models.Document
	if err := s.db.First(&document, "id = ?", documentID).Error; err != nil {
		return false, err
	}

	// 检查文档所属项目的权限
	return s.CheckProjectPermission(userID, document.ProjectID, permission)
}

// CheckTopicPermission 检查用户对话题的权限
func (s *PermissionService) CheckTopicPermission(userID, topicID uuid.UUID, permission ProjectPermission) (bool, error) {
	// 获取话题信息
	var topic models.Topic
	if err := s.db.First(&topic, "id = ?", topicID).Error; err != nil {
		return false, err
	}

	// 如果是公开话题，所有人都可以查看
	if topic.ProjectID == nil {
		return permission == ProjectPermissionView, nil
	}

	// 检查话题所属项目的权限
	return s.CheckProjectPermission(userID, *topic.ProjectID, permission)
}

// CheckHomeworkPermission 检查用户对作业的权限
func (s *PermissionService) CheckHomeworkPermission(userID, homeworkID uuid.UUID, permission ProjectPermission) (bool, error) {
	// 获取作业信息
	var homework models.Homework
	if err := s.db.First(&homework, "id = ?", homeworkID).Error; err != nil {
		return false, err
	}

	// 检查作业所属项目的权限
	return s.CheckProjectPermission(userID, homework.ProjectID, permission)
}

// IsProjectOwner 检查用户是否为项目所有者
func (s *PermissionService) IsProjectOwner(userID, projectID uuid.UUID) (bool, error) {
	var project models.ResearchProject
	err := s.db.Select("creator_id").First(&project, "id = ?", projectID).Error
	if err != nil {
		return false, err
	}
	return project.CreatorID == userID, nil
}

// CanManageProject 检查用户是否可以管理项目
func (s *PermissionService) CanManageProject(userID, projectID uuid.UUID) (bool, error) {
	// 管理员或项目所有者
	isOwner, err := s.IsProjectOwner(userID, projectID)
	if err != nil {
		return false, err
	}
	if isOwner {
		return true, nil
	}

	// 获取用户信息
	var user models.User
	if err := s.db.First(&user, "id = ?", userID).Error; err != nil {
		return false, err
	}

	// 管理员可以管理所有项目
	return user.EduRole >= models.EduRoleAdmin, nil
}

// GetProjectMembersWithPermissions 获取项目成员及其权限
func (s *PermissionService) GetProjectMembersWithPermissions(projectID uuid.UUID) ([]map[string]interface{}, error) {
	var members []models.ProjectMember
	err := s.db.Preload("User").Where("project_id = ?", projectID).Find(&members).Error
	if err != nil {
		return nil, err
	}

	result := []map[string]interface{}{}
	for _, member := range members {
		permissions, err := s.GetUserProjectPermissions(member.UserID, projectID)
		if err != nil {
			return nil, err
		}

		memberInfo := map[string]interface{}{
			"id":          member.ID,
			"user":        member.User,
			"role":        member.Role,
			"permissions": permissions,
			"joined_at":   member.CreatedAt,
		}
		result = append(result, memberInfo)
	}

	return result, nil
}
