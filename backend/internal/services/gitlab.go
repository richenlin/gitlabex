package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/xanzy/go-gitlab"
	"gorm.io/gorm"

	"gitlabex/internal/config"
	"gitlabex/internal/models"
)

// GitLabService GitLab服务 - 所有GitLab API的统一封装
type GitLabService struct {
	client *gitlab.Client
	cache  *redis.Client
	db     *gorm.DB
	config *config.Config
}

// NewGitLabService 创建GitLab服务实例
func NewGitLabService(cfg *config.Config, cache *redis.Client, db *gorm.DB) (*GitLabService, error) {
	// 创建GitLab客户端
	client, err := gitlab.NewClient("", gitlab.WithBaseURL(cfg.GitLab.GetBaseURL()))
	if err != nil {
		return nil, fmt.Errorf("failed to create GitLab client: %w", err)
	}

	// 如果有token，设置认证
	if cfg.GitLab.Token != "" {
		client, err = gitlab.NewClient(cfg.GitLab.Token, gitlab.WithBaseURL(cfg.GitLab.GetBaseURL()))
		if err != nil {
			return nil, fmt.Errorf("failed to create GitLab client with token: %w", err)
		}
	}

	return &GitLabService{
		client: client,
		cache:  cache,
		db:     db,
		config: cfg,
	}, nil
}

// SetAccessToken 设置GitLab访问令牌
func (s *GitLabService) SetAccessToken(token string) error {
	// 重新创建客户端with新的token
	client, err := gitlab.NewClient(token, gitlab.WithBaseURL(s.config.GitLab.GetBaseURL()))
	if err != nil {
		return fmt.Errorf("failed to create GitLab client with new token: %w", err)
	}
	s.client = client
	return nil
}

// GetCurrentUser 获取当前用户信息
func (s *GitLabService) GetCurrentUser() (*gitlab.User, error) {
	user, _, err := s.client.Users.CurrentUser()
	return user, err
}

// SyncUser 同步GitLab用户到本地数据库
func (s *GitLabService) SyncUser(gitlabUser *gitlab.User) (*models.User, error) {
	var user models.User

	// 查找或创建用户
	err := s.db.Where("gitlab_id = ?", gitlabUser.ID).First(&user).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}

	// 更新用户信息
	user.GitLabID = gitlabUser.ID
	user.Username = gitlabUser.Username
	user.Email = gitlabUser.Email
	user.Name = gitlabUser.Name
	user.Avatar = gitlabUser.AvatarURL
	user.LastSyncAt = time.Now()

	// 保存用户
	if err := s.db.Save(&user).Error; err != nil {
		return nil, err
	}

	return &user, nil
}

// GetUserRole 获取用户在指定Group/Project中的教育角色
func (s *GitLabService) GetUserRole(userID int, resourceType string, resourceID int) (models.EducationRole, error) {
	cacheKey := fmt.Sprintf("gitlab:role:%s:%d:user:%d", resourceType, resourceID, userID)

	// 尝试从缓存获取
	if cached, err := s.cache.Get(s.cache.Context(), cacheKey).Result(); err == nil {
		var role models.EducationRole
		if err := json.Unmarshal([]byte(cached), &role); err == nil {
			return role, nil
		}
	}

	var accessLevel gitlab.AccessLevelValue
	var err error

	switch resourceType {
	case "group":
		member, _, err := s.client.GroupMembers.GetGroupMember(resourceID, userID)
		if err != nil {
			return models.EduRoleGuest, err
		}
		accessLevel = member.AccessLevel
	case "project":
		member, _, err := s.client.ProjectMembers.GetProjectMember(resourceID, userID)
		if err != nil {
			return models.EduRoleGuest, err
		}
		accessLevel = member.AccessLevel
	default:
		return models.EduRoleGuest, fmt.Errorf("unsupported resource type: %s", resourceType)
	}

	if err != nil {
		return models.EduRoleGuest, err
	}

	// 映射GitLab权限到教育角色
	role := s.mapGitLabAccessLevel(accessLevel)

	// 缓存结果（5分钟）
	roleBytes, _ := json.Marshal(role)
	s.cache.Set(s.cache.Context(), cacheKey, roleBytes, 5*time.Minute)

	return role, nil
}

// CheckPermission 检查用户权限
func (s *GitLabService) CheckPermission(userID int, resourceType string, resourceID int, action string) (bool, error) {
	role, err := s.GetUserRole(userID, resourceType, resourceID)
	if err != nil {
		return false, err
	}

	return s.hasPermissionForAction(role, action), nil
}

// GetEducationDashboard 获取教育仪表板数据
func (s *GitLabService) GetEducationDashboard(userID int) (*EducationDashboard, error) {
	var (
		groups   []*gitlab.Group
		projects []*gitlab.Project
		issues   []*gitlab.Issue
		mrs      []*gitlab.MergeRequest
		wg       sync.WaitGroup
		mu       sync.Mutex
		errs     []error
	)

	wg.Add(4)

	// 并发获取用户组
	go func() {
		defer wg.Done()
		if userGroups, _, err := s.client.Groups.ListGroups(&gitlab.ListGroupsOptions{
			MinAccessLevel: gitlab.AccessLevel(gitlab.ReporterPermissions),
		}); err != nil {
			mu.Lock()
			errs = append(errs, err)
			mu.Unlock()
		} else {
			groups = userGroups
		}
	}()

	// 并发获取用户项目
	go func() {
		defer wg.Done()
		if userProjects, _, err := s.client.Projects.ListUserProjects(userID, &gitlab.ListProjectsOptions{
			MinAccessLevel: gitlab.AccessLevel(gitlab.ReporterPermissions),
		}); err != nil {
			mu.Lock()
			errs = append(errs, err)
			mu.Unlock()
		} else {
			projects = userProjects
		}
	}()

	// 并发获取分配的Issues
	go func() {
		defer wg.Done()
		if assignedIssues, _, err := s.client.Issues.ListIssues(&gitlab.ListIssuesOptions{
			State: gitlab.Ptr("opened"),
		}); err != nil {
			mu.Lock()
			errs = append(errs, err)
			mu.Unlock()
		} else {
			issues = assignedIssues
		}
	}()

	// 并发获取分配的MR
	go func() {
		defer wg.Done()
		if assignedMRs, _, err := s.client.MergeRequests.ListMergeRequests(&gitlab.ListMergeRequestsOptions{
			State: gitlab.Ptr("opened"),
		}); err != nil {
			mu.Lock()
			errs = append(errs, err)
			mu.Unlock()
		} else {
			mrs = assignedMRs
		}
	}()

	wg.Wait()

	if len(errs) > 0 {
		return nil, fmt.Errorf("failed to load dashboard data: %v", errs)
	}

	return &EducationDashboard{
		Groups:         groups,
		Projects:       projects,
		AssignedIssues: issues,
		AssignedMRs:    mrs,
	}, nil
}

// GetWikiPage 获取Wiki页面
func (s *GitLabService) GetWikiPage(projectID int, slug string) (*gitlab.Wiki, error) {
	wiki, _, err := s.client.Wikis.GetWikiPage(projectID, slug, &gitlab.GetWikiPageOptions{})
	return wiki, err
}

// CreateWikiPage 创建Wiki页面
func (s *GitLabService) CreateWikiPage(projectID int, title, content string) (*gitlab.Wiki, error) {
	format := gitlab.WikiFormatValue("markdown")
	wiki, _, err := s.client.Wikis.CreateWikiPage(projectID, &gitlab.CreateWikiPageOptions{
		Title:   gitlab.String(title),
		Content: gitlab.String(content),
		Format:  &format,
	})
	return wiki, err
}

// UploadFile 上传文件到GitLab项目
func (s *GitLabService) UploadFile(projectID int, filename string, content []byte) (*gitlab.ProjectFile, error) {
	file, _, err := s.client.Projects.UploadFile(projectID, bytes.NewReader(content), filename)
	return file, err
}

// CreateIssue 创建Issue
func (s *GitLabService) CreateIssue(projectID int, title, description string, labels []string, dueDate *time.Time) (*gitlab.Issue, error) {
	labelOptions := gitlab.LabelOptions(labels)
	opts := &gitlab.CreateIssueOptions{
		Title:       gitlab.String(title),
		Description: gitlab.String(description),
		Labels:      &labelOptions,
	}
	if dueDate != nil {
		opts.DueDate = (*gitlab.ISOTime)(dueDate)
	}

	issue, _, err := s.client.Issues.CreateIssue(projectID, opts)
	return issue, err
}

// CreateMergeRequest 创建合并请求
func (s *GitLabService) CreateMergeRequest(projectID int, title, description, sourceBranch, targetBranch string, assigneeID int, labels []string) (*gitlab.MergeRequest, error) {
	labelOptions := gitlab.LabelOptions(labels)
	mr, _, err := s.client.MergeRequests.CreateMergeRequest(projectID, &gitlab.CreateMergeRequestOptions{
		Title:        gitlab.String(title),
		Description:  gitlab.String(description),
		SourceBranch: gitlab.String(sourceBranch),
		TargetBranch: gitlab.String(targetBranch),
		AssigneeIDs:  &[]int{assigneeID},
		Labels:       &labelOptions,
	})
	return mr, err
}

// CreateGroup 创建Group
func (s *GitLabService) CreateGroup(name, path, description string, parentID *int) (*gitlab.Group, error) {
	opts := &gitlab.CreateGroupOptions{
		Name:        gitlab.String(name),
		Path:        gitlab.String(path),
		Description: gitlab.String(description),
		Visibility:  gitlab.Visibility(gitlab.PrivateVisibility),
	}
	if parentID != nil {
		opts.ParentID = parentID
	}

	group, _, err := s.client.Groups.CreateGroup(opts)
	return group, err
}

// 辅助方法

// mapGitLabAccessLevel 映射GitLab权限级别到教育角色
func (s *GitLabService) mapGitLabAccessLevel(level gitlab.AccessLevelValue) models.EducationRole {
	switch level {
	case gitlab.GuestPermissions:
		return models.EduRoleGuest
	case gitlab.ReporterPermissions:
		return models.EduRoleStudent
	case gitlab.DeveloperPermissions:
		return models.EduRoleAssistant
	case gitlab.MaintainerPermissions:
		return models.EduRoleTeacher
	case gitlab.OwnerPermissions:
		return models.EduRoleAdmin
	default:
		return models.EduRoleGuest
	}
}

// hasPermissionForAction 检查角色是否有执行指定动作的权限
func (s *GitLabService) hasPermissionForAction(role models.EducationRole, action string) bool {
	switch action {
	case "read":
		return role >= models.EduRoleGuest
	case "create_issue":
		return role >= models.EduRoleStudent
	case "edit_wiki":
		return role >= models.EduRoleAssistant
	case "push_code":
		return role >= models.EduRoleAssistant
	case "manage_project":
		return role >= models.EduRoleTeacher
	case "delete_project":
		return role >= models.EduRoleAdmin
	default:
		return false
	}
}

// EducationDashboard 教育仪表板数据结构
type EducationDashboard struct {
	Groups         []*gitlab.Group        `json:"groups"`
	Projects       []*gitlab.Project      `json:"projects"`
	AssignedIssues []*gitlab.Issue        `json:"assigned_issues"`
	AssignedMRs    []*gitlab.MergeRequest `json:"assigned_mrs"`
}
