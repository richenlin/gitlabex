package services

import (
	"fmt"
	"time"

	"gitlabex/backend/internal/models"
	"gitlabex/backend/pkg/gitlab"

	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
)

// UserService 用户服务
type UserService struct {
	db     *gorm.DB
	redis  *redis.Client
	gitlab *gitlab.Client
}

// NewUserService 创建用户服务实例
func NewUserService(db *gorm.DB, redis *redis.Client, gitlab *gitlab.Client) *UserService {
	return &UserService{
		db:     db,
		redis:  redis,
		gitlab: gitlab,
	}
}

// CreateUser 创建用户
func (s *UserService) CreateUser(user *models.User) error {
	if err := s.db.Create(user).Error; err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	// 创建默认用户偏好设置
	preference := &models.UserPreference{
		UserID: user.ID,
	}
	if err := s.db.Create(preference).Error; err != nil {
		// 记录错误但不影响用户创建
		fmt.Printf("Failed to create user preference: %v\n", err)
	}

	return nil
}

// GetUserByID 根据ID获取用户
func (s *UserService) GetUserByID(id uint) (*models.User, error) {
	var user models.User
	if err := s.db.Preload("Teams").Preload("Projects").First(&user, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	return &user, nil
}

// GetUserByGitLabID 根据GitLab ID获取用户
func (s *UserService) GetUserByGitLabID(gitlabID int) (*models.User, error) {
	var user models.User
	if err := s.db.Where("git_lab_id = ?", gitlabID).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	return &user, nil
}

// GetUserByEmail 根据邮箱获取用户
func (s *UserService) GetUserByEmail(email string) (*models.User, error) {
	var user models.User
	if err := s.db.Where("email = ?", email).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	return &user, nil
}

// UpdateUser 更新用户信息
func (s *UserService) UpdateUser(user *models.User) error {
	if err := s.db.Save(user).Error; err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}
	return nil
}

// DeleteUser 删除用户
func (s *UserService) DeleteUser(id uint) error {
	if err := s.db.Delete(&models.User{}, id).Error; err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}
	return nil
}

// ListUsers 获取用户列表
func (s *UserService) ListUsers(page, pageSize int, role *models.UserRole) ([]models.User, int64, error) {
	var users []models.User
	var total int64

	query := s.db.Model(&models.User{})

	if role != nil {
		query = query.Where("role = ?", *role)
	}

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count users: %w", err)
	}

	// 获取分页数据
	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Find(&users).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to get users: %w", err)
	}

	return users, total, nil
}

// SyncUserFromGitLab 从GitLab同步用户信息
func (s *UserService) SyncUserFromGitLab(gitlabID int) (*models.User, error) {
	// 获取GitLab用户信息
	gitlabUser, err := s.gitlab.GetUser(gitlabID)
	if err != nil {
		return nil, fmt.Errorf("failed to get GitLab user: %w", err)
	}

	// 检查用户是否已存在
	existingUser, err := s.GetUserByGitLabID(gitlabID)
	if err == nil {
		// 用户已存在，更新信息
		existingUser.Name = gitlabUser.Name
		existingUser.Email = gitlabUser.Email
		existingUser.Avatar = gitlabUser.AvatarURL
		existingUser.Bio = gitlabUser.Bio
		existingUser.Location = gitlabUser.Location
		existingUser.Website = gitlabUser.WebURL
		existingUser.UpdatedAt = time.Now()

		if err := s.UpdateUser(existingUser); err != nil {
			return nil, err
		}
		return existingUser, nil
	}

	// 创建新用户
	user := &models.User{
		GitLabID: gitlabUser.ID,
		Username: gitlabUser.Username,
		Email:    gitlabUser.Email,
		Name:     gitlabUser.Name,
		Avatar:   gitlabUser.AvatarURL,
		Bio:      gitlabUser.Bio,
		Location: gitlabUser.Location,
		Website:  gitlabUser.WebURL,
		Role:     models.RoleStudent, // 默认角色为学生
		IsActive: true,
	}

	if err := s.CreateUser(user); err != nil {
		return nil, err
	}

	return user, nil
}

// UpdateUserRole 更新用户角色
func (s *UserService) UpdateUserRole(userID uint, role models.UserRole) error {
	if err := s.db.Model(&models.User{}).Where("id = ?", userID).Update("role", role).Error; err != nil {
		return fmt.Errorf("failed to update user role: %w", err)
	}
	return nil
}

// GetUserProfile 获取用户详细资料
func (s *UserService) GetUserProfile(userID uint) (*models.UserProfile, error) {
	user, err := s.GetUserByID(userID)
	if err != nil {
		return nil, err
	}
	return user.ToProfile(), nil
}

// UpdateUserLastLogin 更新用户最后登录时间
func (s *UserService) UpdateUserLastLogin(userID uint) error {
	now := time.Now()
	if err := s.db.Model(&models.User{}).Where("id = ?", userID).Update("last_login", now).Error; err != nil {
		return fmt.Errorf("failed to update last login: %w", err)
	}
	return nil
}

// GetUserPreferences 获取用户偏好设置
func (s *UserService) GetUserPreferences(userID uint) (*models.UserPreference, error) {
	var preference models.UserPreference
	if err := s.db.Where("user_id = ?", userID).First(&preference).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			// 创建默认偏好设置
			preference = models.UserPreference{
				UserID: userID,
			}
			if err := s.db.Create(&preference).Error; err != nil {
				return nil, fmt.Errorf("failed to create user preference: %w", err)
			}
		} else {
			return nil, fmt.Errorf("failed to get user preference: %w", err)
		}
	}
	return &preference, nil
}

// UpdateUserPreferences 更新用户偏好设置
func (s *UserService) UpdateUserPreferences(preference *models.UserPreference) error {
	if err := s.db.Save(preference).Error; err != nil {
		return fmt.Errorf("failed to update user preference: %w", err)
	}
	return nil
}

// CreateUserSession 创建用户会话
func (s *UserService) CreateUserSession(userID uint, token string, expiresAt time.Time) (*models.UserSession, error) {
	session := &models.UserSession{
		UserID:    userID,
		Token:     token,
		ExpiresAt: expiresAt,
	}

	if err := s.db.Create(session).Error; err != nil {
		return nil, fmt.Errorf("failed to create user session: %w", err)
	}

	return session, nil
}

// GetUserSession 获取用户会话
func (s *UserService) GetUserSession(token string) (*models.UserSession, error) {
	var session models.UserSession
	if err := s.db.Preload("User").Where("token = ?", token).First(&session).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("session not found")
		}
		return nil, fmt.Errorf("failed to get user session: %w", err)
	}

	// 检查会话是否过期
	if session.IsExpired() {
		// 删除过期会话
		s.db.Delete(&session)
		return nil, fmt.Errorf("session expired")
	}

	return &session, nil
}

// DeleteUserSession 删除用户会话
func (s *UserService) DeleteUserSession(token string) error {
	if err := s.db.Where("token = ?", token).Delete(&models.UserSession{}).Error; err != nil {
		return fmt.Errorf("failed to delete user session: %w", err)
	}
	return nil
}

// DeleteExpiredSessions 删除过期会话
func (s *UserService) DeleteExpiredSessions() error {
	if err := s.db.Where("expires_at < ?", time.Now()).Delete(&models.UserSession{}).Error; err != nil {
		return fmt.Errorf("failed to delete expired sessions: %w", err)
	}
	return nil
}

// GetUserTeams 获取用户的团队列表
func (s *UserService) GetUserTeams(userID uint) ([]models.Team, error) {
	var teams []models.Team
	if err := s.db.Model(&models.User{ID: userID}).Association("Teams").Find(&teams); err != nil {
		return nil, fmt.Errorf("failed to get user teams: %w", err)
	}
	return teams, nil
}

// GetUserProjects 获取用户的项目列表
func (s *UserService) GetUserProjects(userID uint) ([]models.Project, error) {
	var projects []models.Project
	if err := s.db.Model(&models.User{ID: userID}).Association("Projects").Find(&projects); err != nil {
		return nil, fmt.Errorf("failed to get user projects: %w", err)
	}
	return projects, nil
}
