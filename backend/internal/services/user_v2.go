package services

import (
	"fmt"

	"gitlabex/internal/models"

	"gorm.io/gorm"
)

// UserServiceV2 用户管理服务V2
type UserServiceV2 struct {
	db            *gorm.DB
	gitlabService *GitLabService
}

// NewUserServiceV2 创建用户管理服务V2
func NewUserServiceV2(db *gorm.DB, gitlabService *GitLabService) *UserServiceV2 {
	return &UserServiceV2{
		db:            db,
		gitlabService: gitlabService,
	}
}

// UserWithRole 用户信息加角色
type UserWithRole struct {
	*models.User
	RoleName        string `json:"role_name"`
	DynamicRole     string `json:"dynamic_role"`      // 从GitLab动态获取的角色
	DynamicRoleName string `json:"dynamic_role_name"` // 动态角色的中文名称
}

// GetUserByID 根据ID获取用户信息
func (s *UserServiceV2) GetUserByID(userID uint) (*models.User, error) {
	var user models.User
	if err := s.db.First(&user, userID).Error; err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}
	return &user, nil
}

// GetUserWithRole 获取用户信息及角色
func (s *UserServiceV2) GetUserWithRole(userID uint) (*UserWithRole, error) {
	user, err := s.GetUserByID(userID)
	if err != nil {
		return nil, err
	}

	userWithRole := &UserWithRole{
		User:     user,
		RoleName: s.GetRoleName(user.Role),
	}

	// 从GitLab获取动态角色
	if dynamicRole, err := s.GetUserDynamicRole(user.GitLabID); err == nil {
		userWithRole.DynamicRole = dynamicRole
		userWithRole.DynamicRoleName = s.GetDynamicRoleName(dynamicRole)
	}

	return userWithRole, nil
}

// GetRoleName 获取角色名称（静态角色）
func (s *UserServiceV2) GetRoleName(role int) string {
	switch role {
	case 1:
		return "管理员"
	case 2:
		return "教师"
	case 3:
		return "学生"
	case 4:
		return "访客"
	default:
		return "未知"
	}
}

// GetDynamicRoleName 获取动态角色名称
func (s *UserServiceV2) GetDynamicRoleName(role string) string {
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

// GetUserDynamicRole 从GitLab获取用户的动态角色
func (s *UserServiceV2) GetUserDynamicRole(gitlabUserID int) (string, error) {
	// 暂时直接使用数据库中的角色信息
	// 后续可以集成GitLab API来动态获取
	localUser, err := s.GetUserByGitLabID(gitlabUserID)
	if err != nil {
		return "guest", nil
	}

	switch localUser.Role {
	case 1:
		return "admin", nil
	case 2:
		return "teacher", nil
	case 3:
		return "student", nil
	case 4:
		return "guest", nil
	default:
		return "guest", nil
	}
}

// GetUserByGitLabID 根据GitLab ID获取用户
func (s *UserServiceV2) GetUserByGitLabID(gitlabID int) (*models.User, error) {
	var user models.User
	if err := s.db.Where("gitlab_id = ?", gitlabID).First(&user).Error; err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}
	return &user, nil
}

// UpdateUserRole 更新用户角色
func (s *UserServiceV2) UpdateUserRole(userID uint, role int) error {
	return s.db.Model(&models.User{}).Where("id = ?", userID).Update("role", role).Error
}

// GetAllUsers 获取所有用户（分页）
func (s *UserServiceV2) GetAllUsers(page, pageSize int) ([]UserWithRole, int64, error) {
	var users []models.User
	var total int64

	// 计算总数
	if err := s.db.Model(&models.User{}).Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count users: %w", err)
	}

	// 分页查询
	offset := (page - 1) * pageSize
	if err := s.db.Offset(offset).Limit(pageSize).Find(&users).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to get users: %w", err)
	}

	// 转换为带角色的用户信息
	usersWithRole := make([]UserWithRole, len(users))
	for i, user := range users {
		usersWithRole[i] = UserWithRole{
			User:     &user,
			RoleName: s.GetRoleName(user.Role),
		}

		// 获取动态角色
		if dynamicRole, err := s.GetUserDynamicRole(user.GitLabID); err == nil {
			usersWithRole[i].DynamicRole = dynamicRole
			usersWithRole[i].DynamicRoleName = s.GetDynamicRoleName(dynamicRole)
		}
	}

	return usersWithRole, total, nil
}

// SearchUsers 搜索用户
func (s *UserServiceV2) SearchUsers(keyword string, role *int, active *bool, page, pageSize int) ([]UserWithRole, int64, error) {
	query := s.db.Model(&models.User{})

	// 搜索条件
	if keyword != "" {
		query = query.Where("name LIKE ? OR username LIKE ? OR email LIKE ?",
			"%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%")
	}

	if role != nil {
		query = query.Where("role = ?", *role)
	}

	if active != nil {
		query = query.Where("active = ?", *active)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count users: %w", err)
	}

	var users []models.User
	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Find(&users).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to search users: %w", err)
	}

	// 转换为带角色的用户信息
	usersWithRole := make([]UserWithRole, len(users))
	for i, user := range users {
		usersWithRole[i] = UserWithRole{
			User:     &user,
			RoleName: s.GetRoleName(user.Role),
		}

		// 获取动态角色
		if dynamicRole, err := s.GetUserDynamicRole(user.GitLabID); err == nil {
			usersWithRole[i].DynamicRole = dynamicRole
			usersWithRole[i].DynamicRoleName = s.GetDynamicRoleName(dynamicRole)
		}
	}

	return usersWithRole, total, nil
}

// GetUserStats 获取用户统计信息
func (s *UserServiceV2) GetUserStats() (map[string]int, error) {
	stats := make(map[string]int)

	// 总用户数
	var totalUsers int64
	if err := s.db.Model(&models.User{}).Count(&totalUsers).Error; err != nil {
		return nil, fmt.Errorf("failed to count total users: %w", err)
	}
	stats["total_users"] = int(totalUsers)

	// 活跃用户数
	var activeUsers int64
	if err := s.db.Model(&models.User{}).Where("active = true").Count(&activeUsers).Error; err != nil {
		return nil, fmt.Errorf("failed to count active users: %w", err)
	}
	stats["active_users"] = int(activeUsers)

	// 各角色用户数
	var adminCount, teacherCount, studentCount, guestCount int64
	s.db.Model(&models.User{}).Where("role = 1").Count(&adminCount)
	s.db.Model(&models.User{}).Where("role = 2").Count(&teacherCount)
	s.db.Model(&models.User{}).Where("role = 3").Count(&studentCount)
	s.db.Model(&models.User{}).Where("role = 4").Count(&guestCount)

	stats["admin_count"] = int(adminCount)
	stats["teacher_count"] = int(teacherCount)
	stats["student_count"] = int(studentCount)
	stats["guest_count"] = int(guestCount)

	return stats, nil
}

// IsTeacher 检查用户是否是教师
func (s *UserServiceV2) IsTeacher(userID uint) bool {
	user, err := s.GetUserByID(userID)
	if err != nil {
		return false
	}

	// 检查静态角色
	if user.Role == 2 { // 教师
		return true
	}

	// 检查动态角色
	dynamicRole, err := s.GetUserDynamicRole(user.GitLabID)
	if err != nil {
		return false
	}

	return dynamicRole == "teacher" || dynamicRole == "admin"
}

// IsAdmin 检查用户是否是管理员
func (s *UserServiceV2) IsAdmin(userID uint) bool {
	user, err := s.GetUserByID(userID)
	if err != nil {
		return false
	}

	// 检查静态角色
	if user.Role == 1 { // 管理员
		return true
	}

	// 检查动态角色
	dynamicRole, err := s.GetUserDynamicRole(user.GitLabID)
	if err != nil {
		return false
	}

	return dynamicRole == "admin"
}

// GetUserRole 获取用户角色（优先动态角色）
func (s *UserServiceV2) GetUserRole(userID uint) string {
	user, err := s.GetUserByID(userID)
	if err != nil {
		return "guest"
	}

	// 尝试获取动态角色
	if dynamicRole, err := s.GetUserDynamicRole(user.GitLabID); err == nil {
		return dynamicRole
	}

	// 回退到静态角色
	switch user.Role {
	case 1:
		return "admin"
	case 2:
		return "teacher"
	case 3:
		return "student"
	case 4:
		return "guest"
	default:
		return "guest"
	}
}
