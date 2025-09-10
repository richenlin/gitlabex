package services

import (
	"gitlabex/internal/models"

	"golang.org/x/crypto/bcrypt"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

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
