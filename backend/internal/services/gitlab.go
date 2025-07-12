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
	// 检查GitLab客户端是否有有效的token
	if s.config.GitLab.Token == "" || s.config.GitLab.Token == "your-gitlab-token" {
		// 返回默认的空仪表板，而不是错误
		return &EducationDashboard{
			Groups:         []*gitlab.Group{},
			Projects:       []*gitlab.Project{},
			AssignedIssues: []*gitlab.Issue{},
			AssignedMRs:    []*gitlab.MergeRequest{},
		}, nil
	}

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

	// 并发获取分配的MRs
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

	// 如果所有请求都失败了，返回空数据而不是错误
	if len(errs) == 4 {
		return &EducationDashboard{
			Groups:         []*gitlab.Group{},
			Projects:       []*gitlab.Project{},
			AssignedIssues: []*gitlab.Issue{},
			AssignedMRs:    []*gitlab.MergeRequest{},
		}, nil
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

// GetUserProjects 获取用户项目列表
func (s *GitLabService) GetUserProjects(userID int) ([]*gitlab.Project, error) {
	projects, _, err := s.client.Projects.ListUserProjects(userID, &gitlab.ListProjectsOptions{
		MinAccessLevel: gitlab.AccessLevel(gitlab.ReporterPermissions),
	})
	return projects, err
}

// GetWikiPages 获取Wiki页面列表
func (s *GitLabService) GetWikiPages(projectID int) ([]*gitlab.Wiki, error) {
	// GitLab API 中可能没有 ListWikiPages，暂时返回空列表
	// TODO: 实现正确的 Wiki 页面列表获取逻辑
	return []*gitlab.Wiki{}, nil
}

// UpdateWikiPage 更新Wiki页面
func (s *GitLabService) UpdateWikiPage(projectID int, slug, title, content string) (*gitlab.Wiki, error) {
	format := gitlab.WikiFormatValue("markdown")
	wiki, _, err := s.client.Wikis.EditWikiPage(projectID, slug, &gitlab.EditWikiPageOptions{
		Title:   gitlab.String(title),
		Content: gitlab.String(content),
		Format:  &format,
	})
	return wiki, err
}

// CheckWikiEditPermission 检查Wiki编辑权限
func (s *GitLabService) CheckWikiEditPermission(userID int, projectID int) (bool, error) {
	// 获取用户在项目中的权限
	member, _, err := s.client.ProjectMembers.GetProjectMember(projectID, userID)
	if err != nil {
		return false, err
	}

	// Wiki编辑需要Developer及以上权限
	return member.AccessLevel >= gitlab.DeveloperPermissions, nil
}

// === 新增的课题仓库管理功能 ===

// CreateProjectRepository 创建课题仓库
func (s *GitLabService) CreateProjectRepository(name, description, readmeContent string, teacherID int) (*gitlab.Project, error) {
	// 创建项目选项
	opts := &gitlab.CreateProjectOptions{
		Name:                             gitlab.String(name),
		Description:                      gitlab.String(description),
		Visibility:                       gitlab.Visibility(gitlab.PrivateVisibility),
		InitializeWithReadme:             gitlab.Bool(true),
		WikiEnabled:                      gitlab.Bool(true),
		IssuesEnabled:                    gitlab.Bool(true),
		MergeRequestsEnabled:             gitlab.Bool(true),
		JobsEnabled:                      gitlab.Bool(false),
		SnippetsEnabled:                  gitlab.Bool(true),
		ContainerRegistryEnabled:         gitlab.Bool(false),
		SharedRunnersEnabled:             gitlab.Bool(false),
		MergeMethod:                      gitlab.MergeMethod("merge"),
		OnlyAllowMergeIfPipelineSucceeds: gitlab.Bool(false),
		OnlyAllowMergeIfAllDiscussionsAreResolved: gitlab.Bool(false),
		RemoveSourceBranchAfterMerge:              gitlab.Bool(false),
		RequestAccessEnabled:                      gitlab.Bool(false),
		PrintingMergeRequestLinkEnabled:           gitlab.Bool(true),
		AutoDevopsEnabled:                         gitlab.Bool(false),
		ApprovalsBeforeMerge:                      gitlab.Int(0),
	}

	// 创建项目
	project, _, err := s.client.Projects.CreateProject(opts)
	if err != nil {
		return nil, fmt.Errorf("failed to create GitLab project: %w", err)
	}

	// 等待项目创建完成，然后更新README
	if readmeContent != "" {
		// 等待几秒让项目完全创建
		time.Sleep(2 * time.Second)

		// 更新README文件
		_, _, err = s.client.RepositoryFiles.UpdateFile(project.ID, "README.md", &gitlab.UpdateFileOptions{
			Branch:        gitlab.String("main"),
			Content:       gitlab.String(readmeContent),
			CommitMessage: gitlab.String("Update README with course description"),
		})
		if err != nil {
			// 如果更新README失败，我们记录错误但不失败整个创建过程
			fmt.Printf("Warning: Failed to update README: %v\n", err)
		}
	}

	// 设置项目成员权限 - 教师为Owner
	if teacherID > 0 {
		_, _, err = s.client.ProjectMembers.AddProjectMember(project.ID, &gitlab.AddProjectMemberOptions{
			UserID:      gitlab.Int(teacherID),
			AccessLevel: gitlab.AccessLevel(gitlab.OwnerPermissions),
		})
		if err != nil {
			fmt.Printf("Warning: Failed to add teacher as owner: %v\n", err)
		}
	}

	return project, nil
}

// CreateStudentBranch 为学生创建个人分支
func (s *GitLabService) CreateStudentBranch(projectID int, studentID int, branchName string) (*gitlab.Branch, error) {
	// 创建分支
	branch, _, err := s.client.Branches.CreateBranch(projectID, &gitlab.CreateBranchOptions{
		Branch: gitlab.String(branchName),
		Ref:    gitlab.String("main"),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create student branch: %w", err)
	}

	// 创建学生专用目录
	studentDir := fmt.Sprintf("students/%s", branchName)
	initialContent := fmt.Sprintf("# %s 的作业目录\n\n这是 %s 的个人作业目录，请在此目录下提交作业。\n", branchName, branchName)

	_, _, err = s.client.RepositoryFiles.CreateFile(projectID, fmt.Sprintf("%s/README.md", studentDir), &gitlab.CreateFileOptions{
		Branch:        gitlab.String(branchName),
		Content:       gitlab.String(initialContent),
		CommitMessage: gitlab.String(fmt.Sprintf("Initialize student directory for %s", branchName)),
	})
	if err != nil {
		fmt.Printf("Warning: Failed to create student directory: %v\n", err)
	}

	return branch, nil
}

// AddStudentToProject 将学生添加到项目
func (s *GitLabService) AddStudentToProject(projectID int, studentID int, accessLevel gitlab.AccessLevelValue) error {
	// 添加项目成员
	_, _, err := s.client.ProjectMembers.AddProjectMember(projectID, &gitlab.AddProjectMemberOptions{
		UserID:      gitlab.Int(studentID),
		AccessLevel: gitlab.AccessLevel(accessLevel),
	})
	if err != nil {
		return fmt.Errorf("failed to add student to project: %w", err)
	}

	return nil
}

// SubmitAssignment 学生提交作业
func (s *GitLabService) SubmitAssignment(projectID int, branchName string, files map[string]string, commitMessage string) (string, error) {
	var actions []*gitlab.CommitActionOptions

	// 为每个文件创建commit action
	for filePath, content := range files {
		actions = append(actions, &gitlab.CommitActionOptions{
			Action:   gitlab.FileAction(gitlab.FileCreate), // 或者 gitlab.FileUpdate
			FilePath: gitlab.String(filePath),
			Content:  gitlab.String(content),
		})
	}

	// 创建commit
	commit, _, err := s.client.Commits.CreateCommit(projectID, &gitlab.CreateCommitOptions{
		Branch:        gitlab.String(branchName),
		CommitMessage: gitlab.String(commitMessage),
		Actions:       actions,
	})
	if err != nil {
		return "", fmt.Errorf("failed to submit assignment: %w", err)
	}

	return commit.ID, nil
}

// CreateMergeRequestForAssignment 为作业创建合并请求
func (s *GitLabService) CreateMergeRequestForAssignment(projectID int, sourceBranch, title, description string, assigneeID int) (*gitlab.MergeRequest, error) {
	mr, _, err := s.client.MergeRequests.CreateMergeRequest(projectID, &gitlab.CreateMergeRequestOptions{
		Title:        gitlab.String(title),
		Description:  gitlab.String(description),
		SourceBranch: gitlab.String(sourceBranch),
		TargetBranch: gitlab.String("main"),
		AssigneeIDs:  &[]int{assigneeID},
		Labels:       &gitlab.LabelOptions{"homework", "assignment"},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create merge request: %w", err)
	}

	return mr, nil
}

// GetProjectStatistics 获取项目统计信息
func (s *GitLabService) GetProjectStatistics(projectID int) (*ProjectStatistics, error) {
	// 并发获取各种统计信息
	var (
		branches []string
		issues   []*gitlab.Issue
		mrs      []*gitlab.MergeRequest
		commits  []*gitlab.Commit
		wg       sync.WaitGroup
		mu       sync.Mutex
	)

	wg.Add(4)

	// 获取分支
	go func() {
		defer wg.Done()
		if branchList, _, err := s.client.Branches.ListBranches(projectID, &gitlab.ListBranchesOptions{}); err == nil {
			mu.Lock()
			for _, branch := range branchList {
				branches = append(branches, branch.Name)
			}
			mu.Unlock()
		}
	}()

	// 获取Issues
	go func() {
		defer wg.Done()
		if issueList, _, err := s.client.Issues.ListProjectIssues(projectID, &gitlab.ListProjectIssuesOptions{}); err == nil {
			mu.Lock()
			issues = issueList
			mu.Unlock()
		}
	}()

	// 获取MRs
	go func() {
		defer wg.Done()
		if mrList, _, err := s.client.MergeRequests.ListProjectMergeRequests(projectID, &gitlab.ListProjectMergeRequestsOptions{}); err == nil {
			mu.Lock()
			mrs = mrList
			mu.Unlock()
		}
	}()

	// 获取Commits
	go func() {
		defer wg.Done()
		if commitList, _, err := s.client.Commits.ListCommits(projectID, &gitlab.ListCommitsOptions{}); err == nil {
			mu.Lock()
			commits = commitList
			mu.Unlock()
		}
	}()

	wg.Wait()

	// 计算统计信息
	stats := &ProjectStatistics{
		ProjectID:          projectID,
		TotalCommits:       len(commits),
		TotalIssues:        len(issues),
		TotalMergeRequests: len(mrs),
		ActiveBranches:     len(branches),
		OpenIssues:         0,
		OpenMergeRequests:  0,
		WikiPages:          0, // Wiki页面统计需要单独实现
	}

	// 计算开放的Issues和MRs
	for _, issue := range issues {
		if issue.State == "opened" {
			stats.OpenIssues++
		}
	}
	for _, mr := range mrs {
		if mr.State == "opened" {
			stats.OpenMergeRequests++
		}
	}

	return stats, nil
}

// GetBranchCommits 获取分支的提交历史
func (s *GitLabService) GetBranchCommits(projectID int, branchName string, limit int) ([]*gitlab.Commit, error) {
	opts := &gitlab.ListCommitsOptions{
		RefName: gitlab.String(branchName),
		ListOptions: gitlab.ListOptions{
			PerPage: limit,
		},
	}

	commits, _, err := s.client.Commits.ListCommits(projectID, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to get branch commits: %w", err)
	}

	return commits, nil
}

// GetDiscussions 获取项目讨论（通过Issues实现）
func (s *GitLabService) GetDiscussions(projectID int) ([]*gitlab.Issue, error) {
	// 使用Issues来实现讨论功能
	issues, _, err := s.client.Issues.ListProjectIssues(projectID, &gitlab.ListProjectIssuesOptions{
		Labels: &gitlab.LabelOptions{"discussion"},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get discussions: %w", err)
	}

	return issues, nil
}

// CreateDiscussion 创建讨论（通过Issues实现）
func (s *GitLabService) CreateDiscussion(projectID int, title, content string) (*gitlab.Issue, error) {
	labelOptions := gitlab.LabelOptions{"discussion"}
	issue, _, err := s.client.Issues.CreateIssue(projectID, &gitlab.CreateIssueOptions{
		Title:       gitlab.String(title),
		Description: gitlab.String(content),
		Labels:      &labelOptions,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create discussion: %w", err)
	}

	return issue, nil
}

// ProjectStatistics 项目统计信息
type ProjectStatistics struct {
	ProjectID          int `json:"project_id"`
	TotalCommits       int `json:"total_commits"`
	TotalIssues        int `json:"total_issues"`
	OpenIssues         int `json:"open_issues"`
	TotalMergeRequests int `json:"total_merge_requests"`
	OpenMergeRequests  int `json:"open_merge_requests"`
	WikiPages          int `json:"wiki_pages"`
	ActiveBranches     int `json:"active_branches"`
}

// EducationDashboard 教育仪表板数据结构
type EducationDashboard struct {
	Groups         []*gitlab.Group        `json:"groups"`
	Projects       []*gitlab.Project      `json:"projects"`
	AssignedIssues []*gitlab.Issue        `json:"assigned_issues"`
	AssignedMRs    []*gitlab.MergeRequest `json:"assigned_mrs"`
}
