package services

import (
	"fmt"
	"net/http"
	"strconv"

	"gitlabex/internal/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// PermissionService 权限管理服务
type PermissionService struct {
	db *gorm.DB
}

// NewPermissionService 创建权限管理服务
func NewPermissionService(db *gorm.DB) *PermissionService {
	return &PermissionService{
		db: db,
	}
}

// 角色常量
const (
	RoleAdmin   = 1 // 管理员
	RoleTeacher = 2 // 老师
	RoleStudent = 3 // 学生
	RoleGuest   = 4 // 访客
)

// 权限常量
const (
	PermissionRead   = "read"
	PermissionWrite  = "write"
	PermissionDelete = "delete"
	PermissionManage = "manage"
)

// RequireAuth 基础认证中间件
func (s *PermissionService) RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// TODO: 临时屏蔽OAuth检查，方便测试项目其他功能
		// 需要在OAuth配置完成后重新启用认证检查
		// 统一在permission.go中处理所有权限相关的认证逻辑

		fmt.Printf("DEBUG: RequireAuth called for path: %s\n", c.Request.URL.Path)

		if s.isPermissionBypassEnabled() {
			fmt.Printf("DEBUG: Permission bypass enabled, creating test user\n")
			testUser := s.createTestUser()
			c.Set("current_user", testUser)
			c.Next()
			return
		}

		// 正常的权限认证逻辑
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

		// 检查访客权限 - 访客无法进入系统
		if user.Role == RoleGuest {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "Guest access is not allowed",
			})
			c.Abort()
			return
		}

		c.Set("current_user", &user)
		c.Next()
	}
}

// isPermissionBypassEnabled 检查是否启用权限绕过（临时测试功能）
func (s *PermissionService) isPermissionBypassEnabled() bool {
	// TODO: 临时返回true以便测试，后续需要改为false或通过配置控制
	return true
}

// createTestUser 创建测试用户（临时测试功能）
func (s *PermissionService) createTestUser() *models.User {
	// TODO: 临时测试用户，后续需要移除
	testUser := &models.User{
		ID:       1,
		GitLabID: 1,
		Username: "testuser",
		Email:    "test@example.com",
		Name:     "Test User",
		Role:     2, // 学生角色
		Active:   true,
	}

	// 确保测试用户存在于数据库中
	var existingUser models.User
	if err := s.db.Where("id = ?", testUser.ID).First(&existingUser).Error; err != nil {
		// 用户不存在，创建测试用户
		if err := s.db.Create(testUser).Error; err != nil {
			// 如果创建失败，返回内存中的用户对象
			return testUser
		}
	}

	return testUser
}

// RequireRole 角色权限中间件
func (s *PermissionService) RequireRole(roles ...int) gin.HandlerFunc {
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

		// 检查用户角色
		hasPermission := false
		for _, role := range roles {
			if currentUser.Role == role {
				hasPermission = true
				break
			}
		}

		if !hasPermission {
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
	return s.RequireRole(RoleAdmin)
}

// RequireTeacher 老师权限中间件（老师或管理员）
func (s *PermissionService) RequireTeacher() gin.HandlerFunc {
	return s.RequireRole(RoleAdmin, RoleTeacher)
}

// CanAccessClass 检查用户是否能访问班级
func (s *PermissionService) CanAccessClass(user *models.User, classID uint, permission string) bool {
	// 管理员可以访问所有班级
	if user.Role == RoleAdmin {
		return true
	}

	var class models.Class
	if err := s.db.First(&class, classID).Error; err != nil {
		return false
	}

	// 老师可以访问自己创建的班级
	if user.Role == RoleTeacher && class.TeacherID == user.ID {
		return true
	}

	// 学生只能查看自己加入的班级
	if user.Role == RoleStudent && permission == PermissionRead {
		var member models.ClassMember
		err := s.db.Where("class_id = ? AND student_id = ? AND status = 'active'",
			classID, user.ID).First(&member).Error
		return err == nil
	}

	return false
}

// CanAccessProject 检查用户是否能访问课题
func (s *PermissionService) CanAccessProject(user *models.User, projectID uint, permission string) bool {
	// 管理员可以访问所有课题
	if user.Role == RoleAdmin {
		return true
	}

	var project models.Project
	if err := s.db.First(&project, projectID).Error; err != nil {
		return false
	}

	// 老师可以访问自己创建的课题
	if user.Role == RoleTeacher && project.TeacherID == user.ID {
		return true
	}

	// 学生只能查看自己参加的课题
	if user.Role == RoleStudent && permission == PermissionRead {
		var member models.ProjectMember
		err := s.db.Where("project_id = ? AND student_id = ? AND status = 'active'",
			projectID, user.ID).First(&member).Error
		return err == nil
	}

	return false
}

// CanAccessAssignment 检查用户是否能访问作业
func (s *PermissionService) CanAccessAssignment(user *models.User, assignmentID uint, permission string) bool {
	// 管理员可以访问所有作业
	if user.Role == RoleAdmin {
		return true
	}

	var assignment models.Assignment
	if err := s.db.Preload("Project").First(&assignment, assignmentID).Error; err != nil {
		return false
	}

	// 老师可以访问自己创建的作业
	if user.Role == RoleTeacher && assignment.TeacherID == user.ID {
		return true
	}

	// 学生只能查看自己课题中的作业
	if user.Role == RoleStudent && permission == PermissionRead {
		var member models.ProjectMember
		err := s.db.Where("project_id = ? AND student_id = ? AND status = 'active'",
			assignment.ProjectID, user.ID).First(&member).Error
		return err == nil
	}

	return false
}

// RequireClassAccess 班级访问权限中间件
func (s *PermissionService) RequireClassAccess(permission string) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, exists := c.Get("current_user")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Authentication required",
			})
			c.Abort()
			return
		}

		classIDStr := c.Param("id")
		if classIDStr == "" {
			classIDStr = c.Param("class_id")
		}

		classID, err := strconv.ParseUint(classIDStr, 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid class ID",
			})
			c.Abort()
			return
		}

		currentUser := user.(*models.User)
		if !s.CanAccessClass(currentUser, uint(classID), permission) {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "Access denied to this class",
			})
			c.Abort()
			return
		}

		c.Set("class_id", uint(classID))
		c.Next()
	}
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
		if !s.CanAccessProject(currentUser, uint(projectID), permission) {
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
		if !s.CanAccessAssignment(currentUser, uint(assignmentID), permission) {
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

// GetUserRole 获取用户角色名称
func GetUserRoleName(role int) string {
	switch role {
	case RoleAdmin:
		return "admin"
	case RoleTeacher:
		return "teacher"
	case RoleStudent:
		return "student"
	case RoleGuest:
		return "guest"
	default:
		return "unknown"
	}
}

// IsValidRole 检查角色是否有效
func IsValidRole(role int) bool {
	return role >= RoleAdmin && role <= RoleGuest
}

// CanCreateDiscussion 检查用户是否能创建讨论
func (s *PermissionService) CanCreateDiscussion(userID uint, projectID uint) bool {
	var user models.User
	if err := s.db.First(&user, userID).Error; err != nil {
		return false
	}
	return s.CanAccessProject(&user, projectID, PermissionWrite)
}

// CanViewProject 检查用户是否能查看项目
func (s *PermissionService) CanViewProject(userID uint, projectID uint) bool {
	var user models.User
	if err := s.db.First(&user, userID).Error; err != nil {
		return false
	}
	return s.CanAccessProject(&user, projectID, PermissionRead)
}

// IsTeacherOrAdmin 检查用户是否为教师或管理员
func (s *PermissionService) IsTeacherOrAdmin(userID uint, projectID uint) bool {
	var user models.User
	if err := s.db.First(&user, userID).Error; err != nil {
		return false
	}
	return user.Role == RoleAdmin || user.Role == RoleTeacher
}

// CanEditDiscussion 检查用户是否能编辑讨论
func (s *PermissionService) CanEditDiscussion(userID uint, discussionID uint) bool {
	var user models.User
	if err := s.db.First(&user, userID).Error; err != nil {
		return false
	}

	// 管理员可以编辑所有讨论
	if user.Role == RoleAdmin {
		return true
	}

	// 获取讨论信息
	var discussion models.Discussion
	if err := s.db.First(&discussion, discussionID).Error; err != nil {
		return false
	}

	// 作者可以编辑自己的讨论
	if discussion.AuthorID == userID {
		return true
	}

	// 项目的教师可以编辑讨论
	if user.Role == RoleTeacher {
		var project models.Project
		if err := s.db.First(&project, discussion.ProjectID).Error; err != nil {
			return false
		}
		return project.TeacherID == userID
	}

	return false
}

// CanDeleteDiscussion 检查用户是否能删除讨论
func (s *PermissionService) CanDeleteDiscussion(userID uint, discussionID uint) bool {
	var user models.User
	if err := s.db.First(&user, userID).Error; err != nil {
		return false
	}

	// 管理员可以删除所有讨论
	if user.Role == RoleAdmin {
		return true
	}

	// 获取讨论信息
	var discussion models.Discussion
	if err := s.db.First(&discussion, discussionID).Error; err != nil {
		return false
	}

	// 作者可以删除自己的讨论
	if discussion.AuthorID == userID {
		return true
	}

	// 项目的教师可以删除讨论
	if user.Role == RoleTeacher {
		var project models.Project
		if err := s.db.First(&project, discussion.ProjectID).Error; err != nil {
			return false
		}
		return project.TeacherID == userID
	}

	return false
}

// CanReplyDiscussion 检查用户是否能回复讨论
func (s *PermissionService) CanReplyDiscussion(userID uint, projectID uint) bool {
	var user models.User
	if err := s.db.First(&user, userID).Error; err != nil {
		return false
	}
	return s.CanAccessProject(&user, projectID, PermissionRead)
}
