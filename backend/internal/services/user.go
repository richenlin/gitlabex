package services

import (
	"fmt"
	"time"

	"github.com/xanzy/go-gitlab"
	"gorm.io/gorm"

	"gitlabex/internal/models"
)

// UserService 用户服务 - 简化版本，专注于GitLab集成
type UserService struct {
	db        *gorm.DB
	gitlabSvc *GitLabService
}

// NewUserService 创建用户服务实例
func NewUserService(db *gorm.DB, gitlabSvc *GitLabService) *UserService {
	return &UserService{
		db:        db,
		gitlabSvc: gitlabSvc,
	}
}

// GetUserByGitLabID 通过GitLab ID获取用户
func (s *UserService) GetUserByGitLabID(gitlabID int) (*models.User, error) {
	var user models.User
	err := s.db.Where("gitlab_id = ?", gitlabID).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// GetUserByID 通过ID获取用户
func (s *UserService) GetUserByID(id uint) (*models.User, error) {
	var user models.User
	err := s.db.First(&user, id).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// SyncUserFromGitLab 从GitLab同步用户信息
func (s *UserService) SyncUserFromGitLab(gitlabUser *gitlab.User) (*models.User, error) {
	return s.gitlabSvc.SyncUser(gitlabUser)
}

// UpdateUserLastSync 更新用户最后同步时间
func (s *UserService) UpdateUserLastSync(userID uint) error {
	return s.db.Model(&models.User{}).Where("id = ?", userID).Update("last_sync_at", time.Now()).Error
}

// GetUserEducationRole 获取用户在指定资源中的教育角色
func (s *UserService) GetUserEducationRole(user *models.User, resourceType string, resourceID int) (models.EducationRole, error) {
	return s.gitlabSvc.GetUserRole(user.GitLabID, resourceType, resourceID)
}

// CheckUserPermission 检查用户权限
func (s *UserService) CheckUserPermission(user *models.User, resourceType string, resourceID int, action string) (bool, error) {
	return s.gitlabSvc.CheckPermission(user.GitLabID, resourceType, resourceID, action)
}

// ListActiveUsers 获取活跃用户列表（最近24小时内同步过的用户）
func (s *UserService) ListActiveUsers() ([]*models.User, error) {
	var users []*models.User
	yesterday := time.Now().Add(-24 * time.Hour)
	err := s.db.Where("last_sync_at > ?", yesterday).Find(&users).Error
	return users, err
}

// GetUserDashboard 获取用户教育仪表板数据
func (s *UserService) GetUserDashboard(user *models.User) (*UserDashboard, error) {
	// 获取GitLab仪表板数据
	gitlabDashboard, err := s.gitlabSvc.GetEducationDashboard(user.GitLabID)
	if err != nil {
		return nil, fmt.Errorf("failed to get GitLab dashboard: %w", err)
	}

	// 转换为用户仪表板格式
	dashboard := &UserDashboard{
		User:           user,
		Groups:         gitlabDashboard.Groups,
		Projects:       gitlabDashboard.Projects,
		AssignedIssues: gitlabDashboard.AssignedIssues,
		AssignedMRs:    gitlabDashboard.AssignedMRs,
		LastSyncAt:     user.LastSyncAt,
	}

	return dashboard, nil
}

// CreateUser 创建用户（仅用于测试，实际用户通过GitLab同步创建）
func (s *UserService) CreateUser(gitlabID int, username, email, name string) (*models.User, error) {
	user := &models.User{
		GitLabID:   gitlabID,
		Username:   username,
		Email:      email,
		Name:       name,
		LastSyncAt: time.Now(),
	}

	if err := s.db.Create(user).Error; err != nil {
		return nil, err
	}

	return user, nil
}

// CreateTestUser 创建测试用户
func (s *UserService) CreateTestUser() (*models.User, error) {
	// 检查测试用户是否已存在
	var existingUser models.User
	if err := s.db.Where("gitlab_id = ?", 1).First(&existingUser).Error; err == nil {
		// 更新最后同步时间
		existingUser.LastSyncAt = time.Now()
		s.db.Save(&existingUser)
		return &existingUser, nil
	}

	// 创建测试用户 - 使用FirstOrCreate确保唯一性
	testUser := &models.User{
		GitLabID:   1,
		Username:   "testuser",
		Email:      "test@example.com",
		Name:       "Test User",
		Avatar:     "https://www.gravatar.com/avatar/default",
		LastSyncAt: time.Now(),
	}

	// 使用FirstOrCreate避免重复创建
	result := s.db.Where("gitlab_id = ?", testUser.GitLabID).FirstOrCreate(testUser)
	if result.Error != nil {
		return nil, result.Error
	}

	return testUser, nil
}

// GetCurrentUser 获取当前用户（简化版本，返回测试用户）
func (s *UserService) GetCurrentUser() (*models.User, error) {
	// 在实际应用中，这里会从JWT token或session中获取用户信息
	// 目前返回测试用户
	return s.CreateTestUser()
}

// UpdateUser 更新用户信息
func (s *UserService) UpdateUser(user *models.User, updates map[string]interface{}) error {
	updates["last_sync_at"] = time.Now()
	return s.db.Model(user).Updates(updates).Error
}

// DeleteUser 删除用户
func (s *UserService) DeleteUser(userID uint) error {
	return s.db.Delete(&models.User{}, userID).Error
}

// UserDashboard 用户仪表板数据结构
type UserDashboard struct {
	User           *models.User           `json:"user"`
	Groups         []*gitlab.Group        `json:"groups"`
	Projects       []*gitlab.Project      `json:"projects"`
	AssignedIssues []*gitlab.Issue        `json:"assigned_issues"`
	AssignedMRs    []*gitlab.MergeRequest `json:"assigned_mrs"`
	LastSyncAt     time.Time              `json:"last_sync_at"`
}

// UserProfile 用户资料
type UserProfile struct {
	ID         uint                 `json:"id"`
	GitLabID   int                  `json:"gitlab_id"`
	Username   string               `json:"username"`
	Email      string               `json:"email"`
	Name       string               `json:"name"`
	Avatar     string               `json:"avatar"`
	Role       models.EducationRole `json:"role"`
	LastSyncAt time.Time            `json:"last_sync_at"`
	IsActive   bool                 `json:"is_active"`
}

// ToProfile 转换为用户资料
func (s *UserService) ToProfile(user *models.User, role models.EducationRole) *UserProfile {
	return &UserProfile{
		ID:         user.ID,
		GitLabID:   user.GitLabID,
		Username:   user.Username,
		Email:      user.Email,
		Name:       user.Name,
		Avatar:     user.Avatar,
		Role:       role,
		LastSyncAt: user.LastSyncAt,
		IsActive:   user.IsActive(),
	}
}
