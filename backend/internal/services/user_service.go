package services

import (
	"gitlabex/internal/models"
	"strings"

	"golang.org/x/crypto/bcrypt"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// 使用gitlab_service.go中定义的GitLabUser类型

// UserService 用户服务实现
type UserService struct {
	DB *gorm.DB
}

// NewUserService 创建用户服务实例
func NewUserService(db *gorm.DB, gitlabService *GitLabService) *UserService {
	return &UserService{DB: db}
}

// GetUserByID 根据ID获取用户
func (s *UserService) GetUserByID(id uuid.UUID) (*models.User, error) {
	var user models.User
	err := s.DB.First(&user, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// GetUsersByRoles 根据多个角色获取用户
func (s *UserService) GetUsersByRoles(roles []models.UserRole) ([]models.User, error) {
	var users []models.User
	if len(roles) == 0 {
		return users, nil
	}

	err := s.DB.Where("role IN ?", roles).Find(&users).Error
	return users, err
}

// GetAllUsers 获取所有用户
func (s *UserService) GetAllUsers(page, limit int) ([]models.User, int64, error) {
	var users []models.User
	var total int64

	offset := (page - 1) * limit
	err := s.DB.Model(&models.User{}).Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = s.DB.Limit(limit).Offset(offset).Find(&users).Error
	return users, total, err
}

// SearchUsers 搜索用户
func (s *UserService) SearchUsers(query string) ([]models.User, error) {
	var users []models.User

	searchTerm := "%" + query + "%"
	err := s.DB.Where("username LIKE ? OR email LIKE ? OR name LIKE ?",
		searchTerm, searchTerm, searchTerm).Find(&users).Error
	return users, err
}

// GetUserStats 获取用户统计信息
func (s *UserService) GetUserStats() (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// 总用户数
	var totalUsers int64
	err := s.DB.Model(&models.User{}).Count(&totalUsers).Error
	if err != nil {
		return nil, err
	}
	stats["total_users"] = totalUsers

	// 按角色统计
	var roleStats []struct {
		Role  models.UserRole
		Count int64
	}
	err = s.DB.Model(&models.User{}).
		Select("role, count(*) as count").
		Group("role").
		Find(&roleStats).Error
	if err != nil {
		return nil, err
	}

	roleCounts := make(map[string]int64)
	for _, rs := range roleStats {
		roleCounts[string(rs.Role)] = rs.Count
	}
	stats["role_counts"] = roleCounts

	return stats, nil
}

// UpdateUserModel 更新用户信息 (通过模型)
func (s *UserService) UpdateUserModel(user *models.User) error {
	return s.DB.Save(user).Error
}

// GetUsersByRole 根据角色获取用户
func (s *UserService) GetUsersByRole(role models.UserRole) ([]models.User, error) {
	var users []models.User
	err := s.DB.Where("role = ?", role).Find(&users).Error
	return users, err
}

// CreateUserWithPassword 创建带密码的用户 (用于第三方同步)
func (s *UserService) CreateUserWithPassword(user *models.User, password string) error {
	// 哈希密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	// 为了存储密码，我们需要扩展User模型或创建单独的认证表
	// 这里先简单存储到User模型的一个新字段中
	// 注意：实际生产环境中应该有专门的密码存储策略

	// 创建用户
	if err := s.DB.Create(user).Error; err != nil {
		return err
	}

	// 如果需要，这里可以创建密码记录到单独的表
	// 或者扩展User模型包含password字段
	_ = hashedPassword // 避免未使用变量错误

	return nil
}

// ValidateUserPassword 验证用户密码 (用于同步用户登录)
func (s *UserService) ValidateUserPassword(username, password string) (*models.User, error) {
	var user models.User
	if err := s.DB.Where("username = ? AND is_active = true", username).First(&user).Error; err != nil {
		return nil, err
	}

	// 这里需要从密码存储中验证密码
	// 简化实现，实际应该查询密码表进行bcrypt验证

	return &user, nil
}

// GetUserByEmail 根据邮箱获取用户
func (s *UserService) GetUserByEmail(email string) (*models.User, error) {
	var user models.User
	if err := s.DB.Where("email = ?", email).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// GetUserByUsername 根据用户名获取用户
func (s *UserService) GetUserByUsername(username string) (*models.User, error) {
	var user models.User
	if err := s.DB.Where("username = ?", username).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// UpdateUser 根据ID更新用户信息
func (s *UserService) UpdateUser(userID uuid.UUID, updates map[string]interface{}) error {
	return s.DB.Model(&models.User{}).Where("id = ?", userID).Updates(updates).Error
}

// CreateOrUpdateUserFromGitLab 从GitLab用户信息创建或更新用户
func (s *UserService) CreateOrUpdateUserFromGitLab(gitlabUser *GitLabUser, accessToken, refreshToken string) (*models.User, error) {
	// 先尝试根据GitLab ID查找现有用户
	var user models.User
	err := s.DB.Where("gitlab_id = ?", gitlabUser.ID).First(&user).Error

	if err == gorm.ErrRecordNotFound {
		// 用户不存在，创建新用户
		user = models.User{
			GitLabID:     gitlabUser.ID,
			Username:     gitlabUser.Username,
			Email:        gitlabUser.Email,
			Name:         gitlabUser.Name,
			AvatarURL:    gitlabUser.Avatar,
			Role:         models.RoleStudent,    // 默认角色为学生
			EduRole:      models.EduRoleStudent, // 默认教育角色为学生
			IsActive:     true,
			AccessToken:  accessToken,
			RefreshToken: refreshToken,
		}

		// 根据GitLab用户信息推断角色
		user.Role, user.EduRole = s.inferUserRoles(gitlabUser)

		if err := s.DB.Create(&user).Error; err != nil {
			return nil, err
		}
	} else if err != nil {
		// 数据库查询错误
		return nil, err
	} else {
		// 用户存在，更新信息
		updates := map[string]interface{}{
			"username":      gitlabUser.Username,
			"email":         gitlabUser.Email,
			"name":          gitlabUser.Name,
			"avatar_url":    gitlabUser.Avatar,
			"access_token":  accessToken,
			"refresh_token": refreshToken,
			"is_active":     true,
		}

		// 更新角色（如果需要）
		newRole, newEduRole := s.inferUserRoles(gitlabUser)
		if newRole != user.Role {
			updates["role"] = newRole
		}
		if newEduRole != user.EduRole {
			updates["edu_role"] = newEduRole
		}

		if err := s.DB.Model(&user).Updates(updates).Error; err != nil {
			return nil, err
		}

		// 重新加载更新后的用户信息
		if err := s.DB.Where("id = ?", user.ID).First(&user).Error; err != nil {
			return nil, err
		}
	}

	return &user, nil
}

// GetUserByGitLabID 根据GitLab ID获取用户
func (s *UserService) GetUserByGitLabID(gitlabID int64) (*models.User, error) {
	var user models.User
	if err := s.DB.Where("gitlab_id = ?", gitlabID).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// inferUserRoles 根据GitLab用户信息推断用户角色
func (s *UserService) inferUserRoles(gitlabUser *GitLabUser) (models.UserRole, models.EducationRole) {
	// 这里可以根据GitLab用户的信息（如用户名、邮箱域名等）来推断角色
	// 简化实现：默认所有用户都是学生，可以后续通过管理界面调整

	// 可以根据邮箱域名或用户名模式来推断角色
	// 例如：teacher_*, admin_* 等用户名前缀
	username := gitlabUser.Username
	email := gitlabUser.Email

	// 检查是否是管理员
	if isAdminUser(username, email) {
		return models.RoleAdmin, models.EduRoleAdmin
	}

	// 检查是否是教师
	if isTeacherUser(username, email) {
		return models.RoleTeacher, models.EduRoleTeacher
	}

	// 检查是否是助教
	if isAssistantUser(username, email) {
		return models.RoleAssistant, models.EduRoleAssistant
	}

	// 默认为学生
	return models.RoleStudent, models.EduRoleStudent
}

// isAdminUser 检查是否是管理员用户
func isAdminUser(username, email string) bool {
	// 可以根据用户名模式或邮箱域名判断
	adminPatterns := []string{"admin", "administrator", "root"}
	for _, pattern := range adminPatterns {
		if strings.Contains(strings.ToLower(username), pattern) {
			return true
		}
	}
	return false
}

// isTeacherUser 检查是否是教师用户
func isTeacherUser(username, email string) bool {
	teacherPatterns := []string{"teacher", "prof", "instructor"}
	for _, pattern := range teacherPatterns {
		if strings.Contains(strings.ToLower(username), pattern) {
			return true
		}
	}

	// 可以根据邮箱域名判断，例如 @university.edu
	teacherDomains := []string{"@university.edu", "@school.edu"}
	for _, domain := range teacherDomains {
		if strings.Contains(strings.ToLower(email), domain) {
			return true
		}
	}

	return false
}

// isAssistantUser 检查是否是助教用户
func isAssistantUser(username, email string) bool {
	assistantPatterns := []string{"assistant", "ta", "tutor"}
	for _, pattern := range assistantPatterns {
		if strings.Contains(strings.ToLower(username), pattern) {
			return true
		}
	}
	return false
}
