package services

import (
	"fmt"
	"net/http"
	"strconv"

	"gitlabex/internal/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// PermissionService 基于GitLab的权限管理服务
type PermissionService struct {
	db            *gorm.DB
	gitlabService *GitLabService
}

// NewPermissionService 创建权限管理服务
func NewPermissionService(db *gorm.DB, gitlabService *GitLabService) *PermissionService {
	return &PermissionService{
		db:            db,
		gitlabService: gitlabService,
	}
}

// GetUserRole 从GitLab获取用户在特定资源上的角色
func (s *PermissionService) GetUserRole(userID uint, resourceType string, resourceID uint) (models.EducationRole, error) {
	// 获取用户信息
	var user models.User
	if err := s.db.First(&user, userID).Error; err != nil {
		return models.EduRoleGuest, fmt.Errorf("user not found: %w", err)
	}

	// 根据资源类型获取GitLab权限
	switch resourceType {
	case "project":
		return s.getUserProjectRole(user.GitLabID, resourceID)
	case "assignment":
		// 作业权限基于所属课题的权限
		return s.getUserAssignmentRole(user.GitLabID, resourceID)
	case "system":
		// 系统级权限基于用户的整体角色
		return s.getUserSystemRole(user.GitLabID)
	default:
		return models.EduRoleGuest, nil
	}
}

// getUserSystemRole 获取用户的系统级角色
func (s *PermissionService) getUserSystemRole(gitlabUserID int) (models.EducationRole, error) {
	// 从GitLab获取用户的全局权限
	// 这里需要调用GitLab API获取用户的全局角色
	// 暂时使用数据库中的静态角色映射
	var user models.User
	if err := s.db.Where("gitlab_id = ?", gitlabUserID).First(&user).Error; err != nil {
		return models.EduRoleGuest, nil
	}

	// 映射静态角色到教育角色
	switch user.Role {
	case 1: // Admin
		return models.EduRoleAdmin, nil
	case 2: // Teacher
		return models.EduRoleTeacher, nil
	case 3: // Student
		return models.EduRoleStudent, nil
	default:
		return models.EduRoleGuest, nil
	}
}

// getUserProjectRole 获取用户在特定课题中的角色
func (s *PermissionService) getUserProjectRole(gitlabUserID int, projectID uint) (models.EducationRole, error) {
	// 获取课题信息
	var project models.Project
	if err := s.db.First(&project, projectID).Error; err != nil {
		return models.EduRoleGuest, fmt.Errorf("project not found: %w", err)
	}

	// 获取用户信息
	var user models.User
	if err := s.db.Where("gitlab_id = ?", gitlabUserID).First(&user).Error; err != nil {
		return models.EduRoleGuest, fmt.Errorf("user not found: %w", err)
	}

	// 检查是否是课题创建者
	if project.TeacherID == user.ID {
		return models.EduRoleTeacher, nil
	}

	// 检查是否是课题成员
	var member models.ProjectMember
	if err := s.db.Where("project_id = ? AND user_id = ? AND is_active = true",
		projectID, user.ID).First(&member).Error; err == nil {
		// 根据成员角色返回相应权限
		if member.Role == "teacher" || member.Role == "maintainer" {
			return models.EduRoleTeacher, nil
		} else if member.Role == "assistant" || member.Role == "developer" {
			return models.EduRoleAssistant, nil
		} else {
			return models.EduRoleStudent, nil
		}
	}

	// 系统管理员可以访问所有课题
	if systemRole, _ := s.getUserSystemRole(gitlabUserID); systemRole == models.EduRoleAdmin {
		return models.EduRoleAdmin, nil
	}

	return models.EduRoleGuest, nil
}

// getUserAssignmentRole 获取用户在特定作业中的角色
func (s *PermissionService) getUserAssignmentRole(gitlabUserID int, assignmentID uint) (models.EducationRole, error) {
	// 获取作业信息
	var assignment models.Assignment
	if err := s.db.Preload("Project").First(&assignment, assignmentID).Error; err != nil {
		return models.EduRoleGuest, fmt.Errorf("assignment not found: %w", err)
	}

	// 作业权限基于所属课题的权限
	return s.getUserProjectRole(gitlabUserID, assignment.ProjectID)
}

// CanAccessProject 检查用户是否能访问课题
func (s *PermissionService) CanAccessProject(userID, projectID uint, permission string) bool {
	role, err := s.GetUserRole(userID, "project", projectID)
	if err != nil {
		return false
	}

	switch permission {
	case "read":
		return role >= models.EduRoleStudent
	case "write":
		return role >= models.EduRoleAssistant
	case "manage":
		return role >= models.EduRoleTeacher
	case "delete":
		return role >= models.EduRoleAdmin
	default:
		return false
	}
}

// CanAccessAssignment 检查用户是否能访问作业
func (s *PermissionService) CanAccessAssignment(userID, assignmentID uint, permission string) bool {
	role, err := s.GetUserRole(userID, "assignment", assignmentID)
	if err != nil {
		return false
	}

	switch permission {
	case "read":
		return role >= models.EduRoleStudent
	case "submit":
		return role >= models.EduRoleStudent
	case "review":
		return role >= models.EduRoleTeacher
	case "manage":
		return role >= models.EduRoleTeacher
	case "delete":
		return role >= models.EduRoleAdmin
	default:
		return false
	}
}

// IsAdmin 检查用户是否是管理员
func (s *PermissionService) IsAdmin(userID uint) bool {
	role, err := s.GetUserRole(userID, "system", 0)
	if err != nil {
		return false
	}
	return role >= models.EduRoleAdmin
}

// IsTeacher 检查用户是否是教师
func (s *PermissionService) IsTeacher(userID uint) bool {
	role, err := s.GetUserRole(userID, "system", 0)
	if err != nil {
		return false
	}
	return role >= models.EduRoleTeacher
}

// IsProjectOwner 检查用户是否是课题所有者
func (s *PermissionService) IsProjectOwner(userID, projectID uint) bool {
	var project models.Project
	if err := s.db.First(&project, projectID).Error; err != nil {
		return false
	}
	return project.TeacherID == userID
}

// RequireAuth 基础认证中间件
func (s *PermissionService) RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Authentication required",
			})
			c.Abort()
			return
		}

		// 获取用户信息
		var user models.User
		if err := s.db.First(&user, userID).Error; err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid user",
			})
			c.Abort()
			return
		}

		c.Set("current_user", &user)
		c.Next()
	}
}

// RequireRole 角色权限中间件
func (s *PermissionService) RequireRole(requiredRole models.EducationRole) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, exists := c.Get("current_user")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Authentication required",
			})
			c.Abort()
			return
		}

		currentUser := user.(*models.User)
		userRole, err := s.GetUserRole(currentUser.ID, "system", 0)
		if err != nil || userRole < requiredRole {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "Insufficient permissions",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireAdmin 管理员权限中间件
func (s *PermissionService) RequireAdmin() gin.HandlerFunc {
	return s.RequireRole(models.EduRoleAdmin)
}

// RequireTeacher 教师权限中间件（教师或管理员）
func (s *PermissionService) RequireTeacher() gin.HandlerFunc {
	return s.RequireRole(models.EduRoleTeacher)
}

// RequireProjectAccess 课题访问权限中间件
func (s *PermissionService) RequireProjectAccess(permission string) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, exists := c.Get("current_user")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Authentication required",
			})
			c.Abort()
			return
		}

		projectIDStr := c.Param("id")
		if projectIDStr == "" {
			projectIDStr = c.Param("project_id")
		}

		projectID, err := strconv.ParseUint(projectIDStr, 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid project ID",
			})
			c.Abort()
			return
		}

		currentUser := user.(*models.User)
		if !s.CanAccessProject(currentUser.ID, uint(projectID), permission) {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "Access denied to this project",
			})
			c.Abort()
			return
		}

		c.Set("project_id", uint(projectID))
		c.Next()
	}
}

// RequireAssignmentAccess 作业访问权限中间件
func (s *PermissionService) RequireAssignmentAccess(permission string) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, exists := c.Get("current_user")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Authentication required",
			})
			c.Abort()
			return
		}

		assignmentIDStr := c.Param("id")
		if assignmentIDStr == "" {
			assignmentIDStr = c.Param("assignment_id")
		}

		assignmentID, err := strconv.ParseUint(assignmentIDStr, 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid assignment ID",
			})
			c.Abort()
			return
		}

		currentUser := user.(*models.User)
		if !s.CanAccessAssignment(currentUser.ID, uint(assignmentID), permission) {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "Access denied to this assignment",
			})
			c.Abort()
			return
		}

		c.Set("assignment_id", uint(assignmentID))
		c.Next()
	}
}

// ===== 讨论相关权限方法 =====

// CanCreateDiscussion 检查用户是否可以在课题中创建讨论
func (s *PermissionService) CanCreateDiscussion(userID, projectID uint) bool {
	return s.CanAccessProject(userID, projectID, "write")
}

// CanViewProject 检查用户是否可以查看课题
func (s *PermissionService) CanViewProject(userID, projectID uint) bool {
	return s.CanAccessProject(userID, projectID, "read")
}

// IsTeacherOrAdmin 检查用户是否是教师或管理员（在特定课题或全局）
func (s *PermissionService) IsTeacherOrAdmin(userID, projectID uint) bool {
	// 检查全局角色
	if s.IsAdmin(userID) || s.IsTeacher(userID) {
		return true
	}

	// 检查课题级权限
	return s.CanAccessProject(userID, projectID, "manage")
}

// CanEditDiscussion 检查用户是否可以编辑讨论
func (s *PermissionService) CanEditDiscussion(userID, discussionID uint) bool {
	// 获取讨论信息
	var discussion models.Discussion
	if err := s.db.First(&discussion, discussionID).Error; err != nil {
		return false
	}

	// 讨论作者可以编辑自己的讨论
	if discussion.AuthorID == userID {
		return true
	}

	// 教师和管理员可以编辑任何讨论
	return s.IsTeacherOrAdmin(userID, discussion.ProjectID)
}

// CanDeleteDiscussion 检查用户是否可以删除讨论
func (s *PermissionService) CanDeleteDiscussion(userID, discussionID uint) bool {
	// 获取讨论信息
	var discussion models.Discussion
	if err := s.db.First(&discussion, discussionID).Error; err != nil {
		return false
	}

	// 讨论作者可以删除自己的讨论
	if discussion.AuthorID == userID {
		return true
	}

	// 教师和管理员可以删除任何讨论
	return s.IsTeacherOrAdmin(userID, discussion.ProjectID)
}

// CanReplyDiscussion 检查用户是否可以回复讨论
func (s *PermissionService) CanReplyDiscussion(userID, projectID uint) bool {
	return s.CanAccessProject(userID, projectID, "read")
}
