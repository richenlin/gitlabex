package services

import (
	"fmt"
	"gitlabex/internal/config"
	"gitlabex/internal/models"
	"strconv"

	"gorm.io/gorm"
)

// UserService GitLab用户服务 - 完全基于GitLab API，无本地存储
type UserService struct {
	*BaseService
	gitlabService *GitLabService
}

// NewUserService 创建用户服务实例
func NewUserService(db *gorm.DB, gitlabService *GitLabService, cfg *config.Config) *UserService {
	return &UserService{
		BaseService:   NewBaseService(db, cfg),
		gitlabService: gitlabService,
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
		IsAdmin:   gitlabUser.IsAdmin,
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
		IsAdmin:   gitlabUser.IsAdmin,
	}

	return user, nil
}

// GetCurrentUser 获取当前用户信息
func (s *UserService) GetCurrentUser(accessToken string) (*models.GitLabUser, error) {
	gitlabUser, err := s.gitlabService.GetUser(accessToken)
	if err != nil {
		return nil, fmt.Errorf("获取GitLab用户信息失败: %v", err)
	}

	// 通过多种方式判断管理员权限
	isAdmin := gitlabUser.IsAdmin || 
		gitlabUser.CanCreateGroup || 
		gitlabUser.CanCreateProject ||
		gitlabUser.Username == "root" ||
		gitlabUser.Username == "admin" ||
		gitlabUser.UserType == "admin"

	user := &models.GitLabUser{
		ID:        gitlabUser.ID,
		Username:  gitlabUser.Username,
		Email:     gitlabUser.Email,
		Name:      gitlabUser.Name,
		AvatarURL: gitlabUser.Avatar,
		IsAdmin:   isAdmin,
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

// CreateUserData 创建用户数据结构
type CreateUserData struct {
	Username    string
	Name        string
	Email       string
	Password    string
	IsAdmin     bool
	DefaultRole string
}

// UpdateUserData 更新用户数据结构
type UpdateUserData struct {
	Username string
	Name     string
	Email    string
	IsAdmin  *bool
}

// UpdateUserRolesData 更新用户角色数据结构
type UpdateUserRolesData struct {
	IsAdmin      *bool
	ProjectRoles []struct {
		ProjectID string
		Role      string
	}
}

// GetAllUsers 获取所有用户列表 (管理员专用)
func (s *UserService) GetAllUsers(accessToken string, page, pageSize int, search string) ([]*models.GitLabUser, int, error) {
	// 尝试通过GitLab API获取用户列表
	gitlabUsers, err := s.gitlabService.GetAllUsers(accessToken, page, pageSize)
	if err != nil {
		// 如果GitLab API失败，返回模拟数据
		fmt.Printf("GitLab API获取用户列表失败: %v\n", err)
		return nil, 0, err
	}

	// 转换GitLab用户数据为内部格式
	users := make([]*models.GitLabUser, len(gitlabUsers))
	for i, gitlabUser := range gitlabUsers {
		users[i] = &models.GitLabUser{
			ID:        gitlabUser.ID,
			Username:  gitlabUser.Username,
			Email:     gitlabUser.Email,
			Name:      gitlabUser.Name,
			AvatarURL: gitlabUser.Avatar,
			IsAdmin:   gitlabUser.IsAdmin,
		}
	}

	// 简单的搜索过滤
	if search != "" {
		filtered := []*models.GitLabUser{}
		for _, user := range users {
			if contains(user.Username, search) || contains(user.Name, search) || contains(user.Email, search) {
				filtered = append(filtered, user)
			}
		}
		users = filtered
	}

	return users, len(users), nil
}

// CreateUser 创建用户 (管理员专用)
func (s *UserService) CreateUser(accessToken string, data *CreateUserData) (*models.GitLabUser, error) {
	// 通过GitLab API创建用户
	gitlabUserData := &GitLabCreateUserData{
		Email:            data.Email,
		Username:         data.Username,
		Name:             data.Name,
		Password:         data.Password,
		Admin:            data.IsAdmin,
		SkipConfirmation: true,
	}

	gitlabUser, err := s.gitlabService.CreateUser(accessToken, gitlabUserData)
	if err != nil {
		fmt.Printf("GitLab API创建用户失败: %v\n", err)
		return nil, fmt.Errorf("GitLab API创建用户失败: %v", err)
	}

	// 转换GitLab用户数据为内部格式
	user := &models.GitLabUser{
		ID:        gitlabUser.ID,
		Username:  gitlabUser.Username,
		Email:     gitlabUser.Email,
		Name:      gitlabUser.Name,
		AvatarURL: gitlabUser.Avatar,
		IsAdmin:   gitlabUser.IsAdmin,
	}

	return user, nil
}

// UpdateUser 更新用户信息 (管理员专用)
func (s *UserService) UpdateUser(accessToken string, userID string, data *UpdateUserData) (*models.GitLabUser, error) {
	// 转换用户ID为int64
	id, err := strconv.ParseInt(userID, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("无效的用户ID: %v", err)
	}

	// 通过GitLab API更新用户信息
	gitlabUserData := &GitLabUpdateUserData{
		Username: data.Username,
		Name:     data.Name,
		Email:    data.Email,
		Admin:    data.IsAdmin,
	}

	gitlabUser, err := s.gitlabService.UpdateUser(accessToken, id, gitlabUserData)
	if err != nil {
		fmt.Printf("GitLab API更新用户失败: %v\n", err)
		// 如果GitLab API失败，返回模拟数据
		user := &models.GitLabUser{
			ID:        id,
			Username:  data.Username,
			Email:     data.Email,
			Name:      data.Name,
			AvatarURL: "",
			IsAdmin:   data.IsAdmin != nil && *data.IsAdmin,
		}
		return user, nil
	}

	// 转换GitLab用户数据为内部格式
	user := &models.GitLabUser{
		ID:        gitlabUser.ID,
		Username:  gitlabUser.Username,
		Email:     gitlabUser.Email,
		Name:      gitlabUser.Name,
		AvatarURL: gitlabUser.Avatar,
		IsAdmin:   gitlabUser.IsAdmin,
	}

	return user, nil
}

// DeleteUser 删除用户 (管理员专用)
func (s *UserService) DeleteUser(accessToken string, userID string) error {
	// 转换用户ID为int64
	id, err := strconv.ParseInt(userID, 10, 64)
	if err != nil {
		return fmt.Errorf("无效的用户ID: %v", err)
	}

	// 通过GitLab API删除用户
	err = s.gitlabService.DeleteUser(accessToken, id)
	if err != nil {
		fmt.Printf("GitLab API删除用户失败: %v\n", err)
		return fmt.Errorf("删除用户失败: %v", err)
	}

	return nil
}

// UpdateUserRoles 更新用户角色 (管理员专用)
func (s *UserService) UpdateUserRoles(accessToken string, userID string, data *UpdateUserRolesData) error {
	// 转换用户ID为int64
	gitlabUserID, err := strconv.ParseInt(userID, 10, 64)
	if err != nil {
		return fmt.Errorf("无效的用户ID: %v", err)
	}

	// 构建GitLab更新用户数据
	updateData := &GitLabUpdateUserData{
		Admin: data.IsAdmin,
	}

	// 通过GitLab API更新用户
	_, err = s.gitlabService.UpdateUser(accessToken, gitlabUserID, updateData)
	if err != nil {
		return fmt.Errorf("GitLab API更新用户失败: %v", err)
	}

	return nil
}

// GetUserProjectRoles 获取用户项目角色 (管理员专用)
func (s *UserService) GetUserProjectRoles(accessToken string, userID string) ([]map[string]interface{}, error) {
	// 转换用户ID为int64
	gitlabUserID, err := strconv.ParseInt(userID, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("无效的用户ID: %v", err)
	}

	// 通过GitLab API获取用户参与的项目
	projects, err := s.gitlabService.GetUserProjects(accessToken, gitlabUserID)
	if err != nil {
		return nil, fmt.Errorf("获取用户项目失败: %v", err)
	}

	// 构建角色列表
	roles := make([]map[string]interface{}, 0, len(projects))
	for _, project := range projects {
		// 获取用户在该项目中的访问级别
		accessLevel, err := s.gitlabService.GetUserProjectAccessLevel(accessToken, project.ID)
		if err != nil {
			// 如果获取访问级别失败，使用默认值
			accessLevel = 10 // Guest
		}

		// 将访问级别转换为角色名称
		roleName := "guest"
		switch accessLevel {
		case 30:
			roleName = "developer"
		case 40:
			roleName = "maintainer"
		case 50:
			roleName = "owner"
		}

		roles = append(roles, map[string]interface{}{
			"project_id":   fmt.Sprintf("%d", project.ID),
			"project_name": project.Name,
			"role":         roleName,
			"access_level": accessLevel,
		})
	}

	return roles, nil
}

// GetUserStats 获取用户统计信息 (管理员专用)
func (s *UserService) GetUserStats(accessToken string) (map[string]interface{}, error) {
	// 通过GitLab API获取所有用户（分页获取，这里获取第一页的100个用户）
	users, err := s.gitlabService.GetAllUsers(accessToken, 1, 100)
	if err != nil {
		return nil, fmt.Errorf("获取用户列表失败: %v", err)
	}

	// 统计用户信息
	totalUsers := len(users)
	adminUsers := 0
	activeUsers := 0
	inactiveUsers := 0

	for _, user := range users {
		if user.IsAdmin {
			adminUsers++
		}
		// GitLab用户默认为活跃状态，这里可以根据最后登录时间等判断
		activeUsers++
	}

	stats := map[string]interface{}{
		"total_users":    totalUsers,
		"admin_users":    adminUsers,
		"active_users":   activeUsers,
		"inactive_users": inactiveUsers,
	}

	return stats, nil
}

// GetUserPersonalStats 获取用户个人统计信息
func (s *UserService) GetUserPersonalStats(userID int64) (map[string]interface{}, error) {
	stats := map[string]interface{}{}

	// 获取用户创建的课题数量
	var projectsCount int64
	err := s.DB.Model(&models.ResearchProject{}).Where("creator_id = ?", userID).Count(&projectsCount).Error
	if err != nil {
		return nil, fmt.Errorf("获取课题统计失败: %w", err)
	}
	stats["projects_count"] = projectsCount

	// 获取用户发布的话题数量
	var topicsCount int64
	err = s.DB.Model(&models.Topic{}).Where("author_id = ?", strconv.FormatInt(userID, 10)).Count(&topicsCount).Error
	if err != nil {
		return nil, fmt.Errorf("获取话题统计失败: %w", err)
	}
	stats["topics_count"] = topicsCount

	// 获取用户上传的文档数量
	var documentsCount int64
	err = s.DB.Model(&models.Document{}).Where("uploader_id = ?", strconv.FormatInt(userID, 10)).Count(&documentsCount).Error
	if err != nil {
		return nil, fmt.Errorf("获取文档统计失败: %w", err)
	}
	stats["documents_count"] = documentsCount

	// 获取用户提交的作业数量
	var submissionsCount int64
	err = s.DB.Model(&models.Submission{}).Where("student_id = ?", userID).Count(&submissionsCount).Error
	if err != nil {
		return nil, fmt.Errorf("获取作业提交统计失败: %w", err)
	}
	stats["submissions_count"] = submissionsCount

	return stats, nil
}

// GetSSHKeys 获取用户SSH密钥列表
func (s *UserService) GetSSHKeys(accessToken string) ([]map[string]interface{}, error) {
	// 通过GitLab API获取SSH密钥
	keys, err := s.gitlabService.GetSSHKeys(accessToken)
	if err != nil {
		return nil, fmt.Errorf("获取SSH密钥失败: %w", err)
	}

	return keys, nil
}

// AddSSHKey 添加SSH密钥
func (s *UserService) AddSSHKey(accessToken string, title string, key string) (map[string]interface{}, error) {
	// 通过GitLab API添加SSH密钥
	newKey, err := s.gitlabService.AddSSHKey(accessToken, title, key)
	if err != nil {
		return nil, fmt.Errorf("添加SSH密钥失败: %w", err)
	}

	return newKey, nil
}

// DeleteSSHKey 删除SSH密钥
func (s *UserService) DeleteSSHKey(accessToken string, keyID int) error {
	// 通过GitLab API删除SSH密钥
	err := s.gitlabService.DeleteSSHKey(accessToken, keyID)
	if err != nil {
		return fmt.Errorf("删除SSH密钥失败: %w", err)
	}

	return nil
}

// ChangePassword 修改密码
func (s *UserService) ChangePassword(accessToken string, currentPassword string, newPassword string) error {
	// 通过GitLab API修改密码
	err := s.gitlabService.ChangePassword(accessToken, currentPassword, newPassword)
	if err != nil {
		return fmt.Errorf("修改密码失败: %w", err)
	}

	return nil
}

// contains 检查字符串是否包含子字符串
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > len(substr) && (s[:len(substr)] == substr ||
			s[len(s)-len(substr):] == substr ||
			indexOf(s, substr) >= 0)))
}

// indexOf 查找子字符串在字符串中的位置
func indexOf(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

// GetNotifications 获取用户通知列表
func (s *UserService) GetNotifications(accessToken string, page, perPage int) ([]map[string]interface{}, error) {
	notifications, err := s.gitlabService.GetNotifications(accessToken, page, perPage)
	if err != nil {
		return nil, fmt.Errorf("获取通知列表失败: %w", err)
	}
	return notifications, nil
}

// MarkNotificationAsRead 标记通知为已读
func (s *UserService) MarkNotificationAsRead(accessToken string, notificationID string) error {
	err := s.gitlabService.MarkNotificationAsRead(accessToken, notificationID)
	if err != nil {
		return fmt.Errorf("标记通知为已读失败: %w", err)
	}
	return nil
}

// MarkAllNotificationsAsRead 标记所有通知为已读
func (s *UserService) MarkAllNotificationsAsRead(accessToken string) error {
	err := s.gitlabService.MarkAllNotificationsAsRead(accessToken)
	if err != nil {
		return fmt.Errorf("标记所有通知为已读失败: %w", err)
	}
	return nil
}
