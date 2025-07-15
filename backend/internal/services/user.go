package services

import (
	"fmt"
	"strings"

	"gitlabex/internal/config"
	"gitlabex/internal/models"

	"gorm.io/gorm"
)

// UserService 用户管理服务
type UserService struct {
	db                *gorm.DB
	gitlabService     *GitLabService
	permissionService *PermissionService
}

// NewUserService 创建用户管理服务
func NewUserService(db *gorm.DB, gitlabService *GitLabService, permissionService *PermissionService) *UserService {
	return &UserService{
		db:                db,
		gitlabService:     gitlabService,
		permissionService: permissionService,
	}
}

// UserWithRole 用户信息加角色
type UserWithRole struct {
	*models.User
	RoleName        string `json:"role_name"`
	DynamicRole     string `json:"dynamic_role"`      // 从GitLab动态获取的角色
	DynamicRoleName string `json:"dynamic_role_name"` // 动态角色的中文名称
}

// UserProfile 用户资料信息
type UserProfile struct {
	*models.User
	RoleName        string `json:"role_name"`
	DynamicRole     string `json:"dynamic_role"`
	DynamicRoleName string `json:"dynamic_role_name"`
	ProjectCount    int    `json:"project_count"`
	AssignmentCount int    `json:"assignment_count"`
}

// GetUserByID 根据ID获取用户信息
func (s *UserService) GetUserByID(userID uint) (*models.User, error) {
	var user models.User
	if err := s.db.First(&user, userID).Error; err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}
	return &user, nil
}

// GetUserWithRole 获取用户信息及角色
func (s *UserService) GetUserWithRole(userID uint) (*UserWithRole, error) {
	user, err := s.GetUserByID(userID)
	if err != nil {
		return nil, err
	}

	userWithRole := &UserWithRole{
		User:     user,
		RoleName: GetRoleName(user.Role),
	}

	// 从GitLab获取动态角色
	if dynamicRole, err := s.GetUserDynamicRole(user.GitLabID); err == nil {
		userWithRole.DynamicRole = dynamicRole
		userWithRole.DynamicRoleName = GetDynamicRoleName(dynamicRole)
	}

	return userWithRole, nil
}

// GetRoleName 获取角色名称（静态角色）
func GetRoleName(role int) string {
	switch role {
	case int(models.EduRoleAdmin):
		return "管理员"
	case int(models.EduRoleTeacher):
		return "教师"
	case int(models.EduRoleAssistant):
		return "助教"
	case int(models.EduRoleStudent):
		return "学生"
	case int(models.EduRoleGuest):
		return "访客"
	default:
		return "未知"
	}
}

// GetDynamicRoleName 获取动态角色名称
func GetDynamicRoleName(role string) string {
	switch role {
	case "admin":
		return "管理员"
	case "teacher":
		return "教师"
	case "assistant":
		return "助教"
	case "student":
		return "学生"
	case "guest":
		return "访客"
	default:
		return "未知"
	}
}

// GetUserDynamicRole 获取用户在GitLab中的动态角色
func (s *UserService) GetUserDynamicRole(gitlabUserID int) (string, error) {
	// TODO: 调用GitLab API获取用户的实际权限
	// 这里暂时返回基于数据库角色的映射
	var user models.User
	if err := s.db.Where("gitlab_id = ?", gitlabUserID).First(&user).Error; err != nil {
		return "guest", err
	}

	switch user.Role {
	case int(models.EduRoleAdmin):
		return "admin", nil
	case int(models.EduRoleTeacher):
		return "teacher", nil
	case int(models.EduRoleAssistant):
		return "assistant", nil
	case int(models.EduRoleStudent):
		return "student", nil
	default:
		return "guest", nil
	}
}

// GetUserByGitLabID 根据GitLab ID获取用户
func (s *UserService) GetUserByGitLabID(gitlabID int) (*models.User, error) {
	var user models.User
	if err := s.db.Where("gitlab_id = ?", gitlabID).First(&user).Error; err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}
	return &user, nil
}

// UpdateUserRole 更新用户角色
func (s *UserService) UpdateUserRole(userID uint, role int) error {
	return s.db.Model(&models.User{}).Where("id = ?", userID).Update("role", role).Error
}

// GetAllUsers 获取所有用户（分页）
func (s *UserService) GetAllUsers(page, pageSize int) ([]UserWithRole, int64, error) {
	var users []models.User
	var total int64

	// 获取总数
	if err := s.db.Model(&models.User{}).Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count users: %w", err)
	}

	// 分页查询
	offset := (page - 1) * pageSize
	if err := s.db.Offset(offset).Limit(pageSize).Find(&users).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to fetch users: %w", err)
	}

	// 转换为UserWithRole
	var usersWithRole []UserWithRole
	for _, user := range users {
		userWithRole := UserWithRole{
			User:     &user,
			RoleName: GetRoleName(user.Role),
		}

		// 获取动态角色
		if dynamicRole, err := s.GetUserDynamicRole(user.GitLabID); err == nil {
			userWithRole.DynamicRole = dynamicRole
			userWithRole.DynamicRoleName = GetDynamicRoleName(dynamicRole)
		}

		usersWithRole = append(usersWithRole, userWithRole)
	}

	return usersWithRole, total, nil
}

// SearchUsers 搜索用户
func (s *UserService) SearchUsers(keyword string, role *int, active *bool, page, pageSize int) ([]UserWithRole, int64, error) {
	query := s.db.Model(&models.User{})

	// 添加搜索条件
	if keyword != "" {
		keyword = strings.ToLower(keyword)
		query = query.Where("LOWER(username) LIKE ? OR LOWER(name) LIKE ? OR LOWER(email) LIKE ?",
			"%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%")
	}

	if role != nil {
		query = query.Where("role = ?", *role)
	}

	if active != nil {
		query = query.Where("active = ?", *active)
	}

	// 获取总数
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count users: %w", err)
	}

	// 分页查询
	var users []models.User
	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Find(&users).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to search users: %w", err)
	}

	// 转换为UserWithRole
	var usersWithRole []UserWithRole
	for _, user := range users {
		userWithRole := UserWithRole{
			User:     &user,
			RoleName: GetRoleName(user.Role),
		}

		// 获取动态角色
		if dynamicRole, err := s.GetUserDynamicRole(user.GitLabID); err == nil {
			userWithRole.DynamicRole = dynamicRole
			userWithRole.DynamicRoleName = GetDynamicRoleName(dynamicRole)
		}

		usersWithRole = append(usersWithRole, userWithRole)
	}

	return usersWithRole, total, nil
}

// GetUserStats 获取用户统计信息
func (s *UserService) GetUserStats() (map[string]int, error) {
	stats := make(map[string]int)

	// 总用户数
	var totalUsers int64
	if err := s.db.Model(&models.User{}).Count(&totalUsers).Error; err != nil {
		return nil, fmt.Errorf("failed to count total users: %w", err)
	}
	stats["total"] = int(totalUsers)

	// 活跃用户数
	var activeUsers int64
	if err := s.db.Model(&models.User{}).Where("active = true").Count(&activeUsers).Error; err != nil {
		return nil, fmt.Errorf("failed to count active users: %w", err)
	}
	stats["active"] = int(activeUsers)

	// 各角色统计
	roleStats := []struct {
		role int
		key  string
	}{
		{int(models.EduRoleAdmin), "admin"},
		{int(models.EduRoleTeacher), "teacher"},
		{int(models.EduRoleAssistant), "assistant"},
		{int(models.EduRoleStudent), "student"},
		{int(models.EduRoleGuest), "guest"},
	}

	for _, rs := range roleStats {
		var count int64
		if err := s.db.Model(&models.User{}).Where("role = ?", rs.role).Count(&count).Error; err != nil {
			return nil, fmt.Errorf("failed to count users with role %s: %w", rs.key, err)
		}
		stats[rs.key] = int(count)
	}

	return stats, nil
}

// ListActiveUsers 获取活跃用户列表
func (s *UserService) ListActiveUsers() ([]models.User, error) {
	var users []models.User
	if err := s.db.Where("active = true").Find(&users).Error; err != nil {
		return nil, fmt.Errorf("failed to list active users: %w", err)
	}
	return users, nil
}

// IsTeacher 检查用户是否是教师
func (s *UserService) IsTeacher(userID uint) bool {
	var user models.User
	if err := s.db.First(&user, userID).Error; err != nil {
		return false
	}
	return user.Role == int(models.EduRoleTeacher) || user.Role == int(models.EduRoleAdmin)
}

// IsAdmin 检查用户是否是管理员
func (s *UserService) IsAdmin(userID uint) bool {
	var user models.User
	if err := s.db.First(&user, userID).Error; err != nil {
		return false
	}
	return user.Role == int(models.EduRoleAdmin)
}

// CheckUserPermission 检查用户权限（通用方法）
func (s *UserService) CheckUserPermission(userID uint, resourceType string, resourceID uint, permission string) (bool, error) {
	// 获取用户信息
	user, err := s.GetUserByID(userID)
	if err != nil {
		return false, err
	}

	// 根据资源类型检查权限
	switch resourceType {
	case "project":
		return s.permissionService.CanAccessProject(user.ID, resourceID, permission), nil
	case "assignment":
		return s.permissionService.CanAccessAssignment(user.ID, resourceID, permission), nil
	}

	return false, nil
}

// SyncUserFromGitLab 从GitLab同步用户信息
func (s *UserService) SyncUserFromGitLab(gitlabID int) (*models.User, error) {
	// TODO: 实现从GitLab同步用户信息的逻辑
	return s.GetUserByGitLabID(gitlabID)
}

// GetCurrentUser 获取当前用户（用于测试环境）
func (s *UserService) GetCurrentUser() (*models.User, error) {
	// 在生产环境中，这个方法应该从上下文中获取当前用户
	// 这里返回一个默认的管理员用户用于测试
	var user models.User
	if err := s.db.Where("role = ?", int(models.EduRoleAdmin)).First(&user).Error; err != nil {
		return nil, fmt.Errorf("no admin user found: %w", err)
	}
	return &user, nil
}

// GetUserProfile 获取用户资料信息
func (s *UserService) GetUserProfile(userID uint) (*UserProfile, error) {
	user, err := s.GetUserByID(userID)
	if err != nil {
		return nil, err
	}

	// 统计用户相关的课题数量
	var projectCount int64
	s.db.Table("project_members").Where("user_id = ? AND is_active = true", userID).Count(&projectCount)

	// 统计用户相关的作业数量
	var assignmentCount int64
	s.db.Table("assignment_submissions").Where("student_id = ?", userID).Count(&assignmentCount)

	profile := &UserProfile{
		User:            user,
		RoleName:        GetRoleName(user.Role),
		ProjectCount:    int(projectCount),
		AssignmentCount: int(assignmentCount),
	}

	// 获取动态角色
	if dynamicRole, err := s.GetUserDynamicRole(user.GitLabID); err == nil {
		profile.DynamicRole = dynamicRole
		profile.DynamicRoleName = GetDynamicRoleName(dynamicRole)
	}

	return profile, nil
}

// GetUserDashboard 获取用户仪表板信息
func (s *UserService) GetUserDashboard(userID uint) (map[string]interface{}, error) {
	user, err := s.GetUserByID(userID)
	if err != nil {
		return nil, err
	}

	dashboard := map[string]interface{}{
		"user":      user,
		"role_name": GetRoleName(user.Role),
	}

	// 根据角色返回不同的仪表板信息
	switch user.Role {
	case int(models.EduRoleTeacher), int(models.EduRoleAdmin):
		// 教师和管理员看到的统计信息
		var projectCount, studentCount int64
		s.db.Model(&models.Project{}).Where("teacher_id = ?", userID).Count(&projectCount)
		s.db.Table("project_members").Joins("JOIN projects ON projects.id = project_members.project_id").
			Where("projects.teacher_id = ? AND project_members.is_active = true", userID).Count(&studentCount)

		dashboard["project_count"] = projectCount
		dashboard["student_count"] = studentCount

	case int(models.EduRoleStudent):
		// 学生看到的统计信息
		var projectCount, assignmentCount int64
		s.db.Table("project_members").Where("user_id = ? AND is_active = true", userID).Count(&projectCount)
		s.db.Model(&models.AssignmentSubmission{}).Where("student_id = ?", userID).Count(&assignmentCount)

		dashboard["project_count"] = projectCount
		dashboard["assignment_count"] = assignmentCount
	}

	return dashboard, nil
}

// ToProfile 将用户信息转换为UserProfile
func (s *UserService) ToProfile(user *models.User) *UserProfile {
	if user == nil {
		return nil
	}

	// 统计用户相关信息
	var projectCount, assignmentCount int64
	s.db.Table("project_members").Where("user_id = ? AND is_active = true", user.ID).Count(&projectCount)
	s.db.Table("assignment_submissions").Where("student_id = ?", user.ID).Count(&assignmentCount)

	profile := &UserProfile{
		User:            user,
		RoleName:        GetRoleName(user.Role),
		ProjectCount:    int(projectCount),
		AssignmentCount: int(assignmentCount),
	}

	// 获取动态角色
	if dynamicRole, err := s.GetUserDynamicRole(user.GitLabID); err == nil {
		profile.DynamicRole = dynamicRole
		profile.DynamicRoleName = GetDynamicRoleName(dynamicRole)
	}

	return profile
}

// UpdateUserFields 更新用户字段
func (s *UserService) UpdateUserFields(userID uint, updates map[string]interface{}) error {
	// 验证允许更新的字段
	allowedFields := map[string]bool{
		"name":   true,
		"email":  true,
		"avatar": true,
		"role":   true,
		"active": true,
	}

	// 过滤只允许更新的字段
	filteredUpdates := make(map[string]interface{})
	for field, value := range updates {
		if allowedFields[field] {
			filteredUpdates[field] = value
		}
	}

	if len(filteredUpdates) == 0 {
		return fmt.Errorf("no valid fields to update")
	}

	// 验证角色字段
	if role, exists := filteredUpdates["role"]; exists {
		if roleInt, ok := role.(int); ok {
			if !config.IsValidRole(roleInt) {
				return fmt.Errorf("invalid role: %d", roleInt)
			}
		}
	}

	return s.db.Model(&models.User{}).Where("id = ?", userID).Updates(filteredUpdates).Error
}
