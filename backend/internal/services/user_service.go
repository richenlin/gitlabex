package services

import (
	"fmt"
	"gitlabex/internal/config"
	"gitlabex/internal/models"
)

// UserService GitLab用户服务 - 完全基于GitLab API，无本地存储
type UserService struct {
	gitlabService *GitLabService
	config        *config.Config
}

// NewUserService 创建用户服务实例
func NewUserService(gitlabService *GitLabService, cfg *config.Config) *UserService {
	return &UserService{
		gitlabService: gitlabService,
		config:        cfg,
	}
}

// GetUserByGitLabID 根据GitLab ID获取用户信息
func (s *UserService) GetUserByGitLabID(accessToken string, gitlabID int64) (*models.GitLabUser, error) {
	// 直接从GitLab API获取用户信息
	gitlabUser, err := s.gitlabService.GetUser(accessToken)
	if err != nil {
		return nil, fmt.Errorf("获取GitLab用户信息失败: %v", err)
	}

	// 验证用户ID是否匹配
	if gitlabUser.ID != gitlabID {
		return nil, fmt.Errorf("用户ID不匹配")
	}

	// 转换为统一的GitLabUser结构
	user := &models.GitLabUser{
		ID:        gitlabUser.ID,
		Username:  gitlabUser.Username,
		Email:     gitlabUser.Email,
		Name:      gitlabUser.Name,
		AvatarURL: gitlabUser.Avatar,
		IsAdmin:   false, // 需要通过其他API检查管理员权限
	}

	return user, nil
}

// GetUserByUsername 根据用户名获取用户信息
func (s *UserService) GetUserByUsername(accessToken, username string) (*models.GitLabUser, error) {
	// GitLab API不直接支持按用户名查找，需要先获取当前用户信息
	gitlabUser, err := s.gitlabService.GetUser(accessToken)
	if err != nil {
		return nil, fmt.Errorf("获取GitLab用户信息失败: %v", err)
	}

	// 检查用户名是否匹配
	if gitlabUser.Username != username {
		return nil, fmt.Errorf("用户名不匹配")
	}

	user := &models.GitLabUser{
		ID:        gitlabUser.ID,
		Username:  gitlabUser.Username,
		Email:     gitlabUser.Email,
		Name:      gitlabUser.Name,
		AvatarURL: gitlabUser.Avatar,
		IsAdmin:   false,
	}

	return user, nil
}

// GetCurrentUser 获取当前用户信息
func (s *UserService) GetCurrentUser(accessToken string) (*models.GitLabUser, error) {
	gitlabUser, err := s.gitlabService.GetUser(accessToken)
	if err != nil {
		return nil, fmt.Errorf("获取GitLab用户信息失败: %v", err)
	}

	user := &models.GitLabUser{
		ID:        gitlabUser.ID,
		Username:  gitlabUser.Username,
		Email:     gitlabUser.Email,
		Name:      gitlabUser.Name,
		AvatarURL: gitlabUser.Avatar,
		IsAdmin:   false,
	}

	return user, nil
}

// GetProjectMembers 获取项目成员列表
func (s *UserService) GetProjectMembers(accessToken string, projectID int64) ([]*models.GitLabUser, error) {
	// 通过GitLab API获取项目成员
	// 这里需要在GitLabService中实现GetProjectMembers方法
	return nil, fmt.Errorf("获取项目成员功能待实现")
}

// GetUserProjectRole 获取用户在项目中的角色
func (s *UserService) GetUserProjectRole(accessToken string, projectID, userID int64) (models.GitLabRole, error) {
	// 通过GitLab API获取用户在项目中的权限级别
	// 这里需要在GitLabService中实现相应方法
	return models.GitLabGuest, fmt.Errorf("获取用户项目角色功能待实现")
}

// CheckProjectPermission 检查用户是否有项目权限
func (s *UserService) CheckProjectPermission(accessToken string, projectID int64, requiredRole models.GitLabRole) (bool, error) {
	// 验证用户对项目的访问权限
	err := s.gitlabService.ValidateRepositoryAccess(accessToken, projectID)
	if err != nil {
		return false, err
	}

	// 获取用户在项目中的角色
	// 这里需要实现具体的权限检查逻辑
	return true, nil
}
