package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"gitlabex/internal/config"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// GitLabService GitLab API服务
type GitLabService struct {
	Config     *config.Config
	HTTPClient *http.Client
}

// NewGitLabService 创建GitLab服务
func NewGitLabService(cfg *config.Config) *GitLabService {
	// 配置HTTP传输层以优化连接复用和资源管理
	transport := &http.Transport{
		MaxIdleConns:        100,              // 最大空闲连接数
		MaxIdleConnsPerHost: 10,               // 每个主机的最大空闲连接数
		IdleConnTimeout:     90 * time.Second, // 空闲连接超时时间
		DisableKeepAlives:   false,            // 启用Keep-Alive
		ForceAttemptHTTP2:   true,             // 强制尝试HTTP/2
	}

	return &GitLabService{
		Config: cfg,
		HTTPClient: &http.Client{
			Timeout:   30 * time.Second,
			Transport: transport,
		},
	}
}

// makeRequest 通用HTTP请求方法，带有context支持
func (s *GitLabService) makeRequest(ctx context.Context, method, url, accessToken string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	return s.HTTPClient.Do(req)
}

// ProtectBranchRequest 保护分支请求
type ProtectBranchRequest struct {
	Name                      string `json:"name"`
	PushAccessLevel           int    `json:"push_access_level"`      // 30=Developer, 40=Maintainer, 60=Admin
	MergeAccessLevel          int    `json:"merge_access_level"`     // 30=Developer, 40=Maintainer, 60=Admin
	UnprotectAccessLevel      int    `json:"unprotect_access_level"` // 40=Maintainer, 60=Admin
	AllowForcePush            bool   `json:"allow_force_push"`
	CodeOwnerApprovalRequired bool   `json:"code_owner_approval_required"`
}

// ProtectedBranch 受保护的分支信息
type ProtectedBranch struct {
	ID                        int64         `json:"id"`
	Name                      string        `json:"name"`
	PushAccessLevels          []AccessLevel `json:"push_access_levels"`
	MergeAccessLevels         []AccessLevel `json:"merge_access_levels"`
	UnprotectAccessLevels     []AccessLevel `json:"unprotect_access_levels"`
	AllowForcePush            bool          `json:"allow_force_push"`
	CodeOwnerApprovalRequired bool          `json:"code_owner_approval_required"`
}

type AccessLevel struct {
	AccessLevel            int64  `json:"access_level"`
	AccessLevelDescription string `json:"access_level_description"`
}

// ProtectBranch 保护分支
func (s *GitLabService) ProtectBranch(accessToken string, projectID int64, req *ProtectBranchRequest) (*ProtectedBranch, error) {
	url := fmt.Sprintf("%s/api/v4/projects/%d/protected_branches", s.Config.GitLabURL, projectID)

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("序列化请求失败: %v", err)
	}

	httpReq, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %v", err)
	}

	httpReq.Header.Set("Authorization", "Bearer "+accessToken)
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := s.HTTPClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("发送请求失败: %v", err)
	}
	defer resp.Body.Close()

	// GitLab在分支已经受保护时会返回409状态码
	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusConflict {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("保护分支失败，状态码: %d, 响应: %s", resp.StatusCode, string(bodyBytes))
	}

	// 如果分支已经受保护，尝试获取现有的保护规则
	if resp.StatusCode == http.StatusConflict {
		return s.GetProtectedBranch(accessToken, projectID, req.Name)
	}

	var protectedBranch ProtectedBranch
	if err := json.NewDecoder(resp.Body).Decode(&protectedBranch); err != nil {
		return nil, fmt.Errorf("解析响应失败: %v", err)
	}

	return &protectedBranch, nil
}

// GetProtectedBranch 获取受保护分支信息
func (s *GitLabService) GetProtectedBranch(accessToken string, projectID int64, branchName string) (*ProtectedBranch, error) {
	encodedBranch := url.PathEscape(branchName)
	apiUrl := fmt.Sprintf("%s/api/v4/projects/%d/protected_branches/%s", s.Config.GitLabURL, projectID, encodedBranch)

	req, err := http.NewRequest("GET", apiUrl, nil)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %v", err)
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := s.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("发送请求失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("获取受保护分支失败，状态码: %d", resp.StatusCode)
	}

	var protectedBranch ProtectedBranch
	if err := json.NewDecoder(resp.Body).Decode(&protectedBranch); err != nil {
		return nil, fmt.Errorf("解析响应失败: %v", err)
	}

	return &protectedBranch, nil
}

// SetupProjectBranchProtection 为新创建的课题项目设置分支保护
func (s *GitLabService) SetupProjectBranchProtection(accessToken string, projectID int64) error {
	// 保护主分支，只允许维护者和管理员推送
	mainBranchProtection := &ProtectBranchRequest{
		Name:                      "main",
		PushAccessLevel:           40, // 40 = Maintainer level
		MergeAccessLevel:          40, // 40 = Maintainer level
		UnprotectAccessLevel:      40, // 40 = Maintainer level
		AllowForcePush:            false,
		CodeOwnerApprovalRequired: false,
	}

	_, err := s.ProtectBranch(accessToken, projectID, mainBranchProtection)
	if err != nil {
		// 如果main分支不存在，尝试保护master分支
		masterBranchProtection := &ProtectBranchRequest{
			Name:                      "master",
			PushAccessLevel:           40, // 40 = Maintainer level
			MergeAccessLevel:          40, // 40 = Maintainer level
			UnprotectAccessLevel:      40, // 40 = Maintainer level
			AllowForcePush:            false,
			CodeOwnerApprovalRequired: false,
		}

		_, masterErr := s.ProtectBranch(accessToken, projectID, masterBranchProtection)
		if masterErr != nil {
			return fmt.Errorf("保护主分支失败，main分支错误: %v, master分支错误: %v", err, masterErr)
		}
	}

	return nil
}

// GitLabUser GitLab用户信息 - 使用models包中的定义
// 这里保留一个简化的结构用于API响应解析
type GitLabAPIUser struct {
	ID       int64  `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Name     string `json:"name"`
	Avatar   string `json:"avatar_url"`
}

// GitLabProject GitLab项目信息
type GitLabProject struct {
	ID                int64  `json:"id"`
	Name              string `json:"name"`
	Path              string `json:"path"`
	PathWithNamespace string `json:"path_with_namespace"`
	WebURL            string `json:"web_url"`
	SSHURL            string `json:"ssh_url_to_repo"`
	HTTPURL           string `json:"http_url_to_repo"`
	Description       string `json:"description"`
	DefaultBranch     string `json:"default_branch"`
	Visibility        string `json:"visibility"`
	CreatedAt         string `json:"created_at"`
	LastActivityAt    string `json:"last_activity_at"`
	Namespace         struct {
		ID   int64  `json:"id"`
		Name string `json:"name"`
		Path string `json:"path"`
	} `json:"namespace"`
	Owner *GitLabAPIUser `json:"owner,omitempty"`
}

// GitLabFile GitLab文件信息
type GitLabFile struct {
	FileName      string `json:"file_name"`
	FilePath      string `json:"file_path"`
	Size          int    `json:"size"`
	Encoding      string `json:"encoding"`
	Content       string `json:"content"`
	ContentSHA256 string `json:"content_sha256"`
	Ref           string `json:"ref"`
	BlobID        string `json:"blob_id"`
	CommitID      string `json:"commit_id"`
	LastCommitID  string `json:"last_commit_id"`
}

// GitLabBranch GitLab分支信息
type GitLabBranch struct {
	Name   string `json:"name"`
	Commit struct {
		ID        string `json:"id"`
		ShortID   string `json:"short_id"`
		Title     string `json:"title"`
		Author    string `json:"author_name"`
		Message   string `json:"message"`
		CreatedAt string `json:"created_at"`
	} `json:"commit"`
	Protected bool `json:"protected"`
	Merged    bool `json:"merged"`
}

// GitLabCommit GitLab提交信息
type GitLabCommit struct {
	ID        string         `json:"id"`
	ShortID   string         `json:"short_id"`
	Title     string         `json:"title"`
	Author    *GitLabAPIUser `json:"author"`
	Committer *GitLabAPIUser `json:"committer"`
	Message   string         `json:"message"`
	CreatedAt time.Time      `json:"created_at"`
	WebURL    string         `json:"web_url"`
}

// GitLabIssue GitLab议题信息
type GitLabIssue struct {
	ID          int64    `json:"id"`
	IID         int64    `json:"iid"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	State       string   `json:"state"`
	CreatedAt   string   `json:"created_at"`
	UpdatedAt   string   `json:"updated_at"`
	Labels      []string `json:"labels"`
	Author      struct {
		ID        int64  `json:"id"`
		Username  string `json:"username"`
		Name      string `json:"name"`
		AvatarURL string `json:"avatar_url"`
	} `json:"author"`
	Upvotes        int    `json:"upvotes"`
	Downvotes      int    `json:"downvotes"`
	UserNotesCount int    `json:"user_notes_count"`
	WebURL         string `json:"web_url"`
}

// GitLabIssueNote GitLab Issue评论结构
type GitLabIssueNote struct {
	ID        int64  `json:"id"`
	Body      string `json:"body"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
	Author    struct {
		ID        int64  `json:"id"`
		Username  string `json:"username"`
		Name      string `json:"name"`
		AvatarURL string `json:"avatar_url"`
	} `json:"author"`
}

// GitLabAwardEmoji GitLab Award Emoji结构
type GitLabAwardEmoji struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
	User struct {
		ID        int64  `json:"id"`
		Username  string `json:"username"`
		Name      string `json:"name"`
		AvatarURL string `json:"avatar_url"`
	} `json:"user"`
}

// GitLabMergeRequest GitLab合并请求信息
type GitLabMergeRequest struct {
	ID           int64          `json:"id"`
	IID          int64          `json:"iid"`
	Title        string         `json:"title"`
	Description  string         `json:"description"`
	State        string         `json:"state"`
	SourceBranch string         `json:"source_branch"`
	TargetBranch string         `json:"target_branch"`
	Author       *GitLabAPIUser `json:"author"`
	Assignee     *GitLabAPIUser `json:"assignee"`
	CreatedAt    string         `json:"created_at"`
	UpdatedAt    string         `json:"updated_at"`
	WebURL       string         `json:"web_url"`
	MergedAt     *string        `json:"merged_at,omitempty"`
}

// CreateProjectRequest 创建项目请求
type CreateProjectRequest struct {
	Name                 string `json:"name"`
	Path                 string `json:"path"` // 必需字段，项目路径
	Description          string `json:"description,omitempty"`
	Visibility           string `json:"visibility,omitempty"` // private, internal, public
	InitializeWithReadme bool   `json:"initialize_with_readme,omitempty"`
	DefaultBranch        string `json:"default_branch,omitempty"`
}

// CreateBranchRequest 创建分支请求
type CreateBranchRequest struct {
	Branch string `json:"branch"`
	Ref    string `json:"ref"`
}

// CreateFileRequest 创建文件请求
type CreateFileRequest struct {
	Branch        string `json:"branch"`
	Content       string `json:"content"`
	CommitMessage string `json:"commit_message"`
	AuthorEmail   string `json:"author_email,omitempty"`
	AuthorName    string `json:"author_name,omitempty"`
}

// GitLabWebhookPayload GitLab webhook负载
type GitLabWebhookPayload struct {
	ObjectKind string          `json:"object_kind"`
	EventName  string          `json:"event_name"`
	User       *GitLabAPIUser  `json:"user"`
	Project    *GitLabProject  `json:"project"`
	Commits    []*GitLabCommit `json:"commits"`
	Ref        string          `json:"ref"`
	Before     string          `json:"before"`
	After      string          `json:"after"`
	Repository struct {
		Name        string `json:"name"`
		URL         string `json:"url"`
		Description string `json:"description"`
		Homepage    string `json:"homepage"`
	} `json:"repository"`
}

// GetUser 获取当前用户信息
func (s *GitLabService) GetUser(accessToken string) (*GitLabAPIUser, error) {
	url := fmt.Sprintf("%s/api/v4/user", s.Config.GitLabURL)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitLab API error: %s", resp.Status)
	}

	var user GitLabAPIUser
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return nil, err
	}

	return &user, nil
}

// GetProjects 获取用户的项目列表
func (s *GitLabService) GetProjects(accessToken string, page, perPage int) ([]*GitLabProject, error) {
	url := fmt.Sprintf("%s/api/v4/projects?owned=true&page=%d&per_page=%d&order_by=last_activity_at",
		s.Config.GitLabURL, page, perPage)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitLab API error: %s", resp.Status)
	}

	var projects []*GitLabProject
	if err := json.NewDecoder(resp.Body).Decode(&projects); err != nil {
		return nil, err
	}

	return projects, nil
}

// GetProject 获取特定项目信息
func (s *GitLabService) GetProject(accessToken string, projectID int64) (*GitLabProject, error) {
	url := fmt.Sprintf("%s/api/v4/projects/%d", s.Config.GitLabURL, projectID)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// 读取响应体用于错误诊断
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %v", err)
	}

	if resp.StatusCode == 401 {
		return nil, fmt.Errorf("GitLab访问令牌无效或已过期，请重新登录")
	} else if resp.StatusCode == 403 {
		return nil, fmt.Errorf("没有访问该项目的权限")
	} else if resp.StatusCode == 404 {
		return nil, fmt.Errorf("项目不存在")
	} else if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitLab API错误 (%d): %s", resp.StatusCode, string(body))
	}

	var project GitLabProject
	if err := json.Unmarshal(body, &project); err != nil {
		return nil, err
	}

	return &project, nil
}

// CreateProject 创建新项目
func (s *GitLabService) CreateProject(accessToken string, req *CreateProjectRequest) (*GitLabProject, error) {
	url := fmt.Sprintf("%s/api/v4/projects", s.Config.GitLabURL)

	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}

	httpReq.Header.Set("Authorization", "Bearer "+accessToken)
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := s.HTTPClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("GitLab API error: %s, body: %s", resp.Status, string(bodyBytes))
	}

	var project GitLabProject
	if err := json.NewDecoder(resp.Body).Decode(&project); err != nil {
		return nil, err
	}

	return &project, nil
}

// GetBranches 获取项目分支列表
func (s *GitLabService) GetBranches(accessToken string, projectID int64) ([]*GitLabBranch, error) {
	url := fmt.Sprintf("%s/api/v4/projects/%d/repository/branches", s.Config.GitLabURL, projectID)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitLab API error: %s", resp.Status)
	}

	var branches []*GitLabBranch
	if err := json.NewDecoder(resp.Body).Decode(&branches); err != nil {
		return nil, err
	}

	return branches, nil
}

// CreateBranch 创建新分支
func (s *GitLabService) CreateBranch(accessToken string, projectID int64, req *CreateBranchRequest) (*GitLabBranch, error) {
	url := fmt.Sprintf("%s/api/v4/projects/%d/repository/branches", s.Config.GitLabURL, projectID)

	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}

	httpReq.Header.Set("Authorization", "Bearer "+accessToken)
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := s.HTTPClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("GitLab API error: %s", resp.Status)
	}

	var branch GitLabBranch
	if err := json.NewDecoder(resp.Body).Decode(&branch); err != nil {
		return nil, err
	}

	return &branch, nil
}

// GetFileContent 获取文件内容
func (s *GitLabService) GetFileContent(accessToken string, projectID int64, filePath, ref string) (*GitLabFile, error) {
	encodedPath := url.PathEscape(filePath)
	url := fmt.Sprintf("%s/api/v4/projects/%d/repository/files/%s?ref=%s",
		s.Config.GitLabURL, projectID, encodedPath, ref)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitLab API error: %s", resp.Status)
	}

	var file GitLabFile
	if err := json.NewDecoder(resp.Body).Decode(&file); err != nil {
		return nil, err
	}

	return &file, nil
}

// CreateFile 创建新文件
func (s *GitLabService) CreateFile(accessToken string, projectID int64, filePath string, req *CreateFileRequest) error {
	encodedPath := url.PathEscape(filePath)
	url := fmt.Sprintf("%s/api/v4/projects/%d/repository/files/%s",
		s.Config.GitLabURL, projectID, encodedPath)

	body, err := json.Marshal(req)
	if err != nil {
		return err
	}

	httpReq, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		return err
	}

	httpReq.Header.Set("Authorization", "Bearer "+accessToken)
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := s.HTTPClient.Do(httpReq)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("GitLab API error: %s", resp.Status)
	}

	return nil
}

// UpdateFile 更新文件内容
func (s *GitLabService) UpdateFile(accessToken string, projectID int64, filePath string, req *CreateFileRequest) error {
	encodedPath := url.PathEscape(filePath)
	url := fmt.Sprintf("%s/api/v4/projects/%d/repository/files/%s",
		s.Config.GitLabURL, projectID, encodedPath)

	body, err := json.Marshal(req)
	if err != nil {
		return err
	}

	httpReq, err := http.NewRequest("PUT", url, bytes.NewBuffer(body))
	if err != nil {
		return err
	}

	httpReq.Header.Set("Authorization", "Bearer "+accessToken)
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := s.HTTPClient.Do(httpReq)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("GitLab API error: %s", resp.Status)
	}

	return nil
}

// GetCommits 获取提交历史
func (s *GitLabService) GetCommits(accessToken string, projectID int64, branch string, limit int) ([]*GitLabCommit, error) {
	url := fmt.Sprintf("%s/api/v4/projects/%d/repository/commits?ref_name=%s&per_page=%d",
		s.Config.GitLabURL, projectID, branch, limit)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitLab API error: %s", resp.Status)
	}

	var commits []*GitLabCommit
	if err := json.NewDecoder(resp.Body).Decode(&commits); err != nil {
		return nil, err
	}

	return commits, nil
}

// GetIssues 获取项目议题
func (s *GitLabService) GetIssues(accessToken string, projectID int64, state string, labels []string) ([]*GitLabIssue, error) {
	apiUrl := fmt.Sprintf("%s/api/v4/projects/%d/issues", s.Config.GitLabURL, projectID)

	params := url.Values{}
	if state != "" {
		params.Set("state", state)
	}
	if len(labels) > 0 {
		params.Set("labels", strings.Join(labels, ","))
	}

	if len(params) > 0 {
		apiUrl += "?" + params.Encode()
	}

	req, err := http.NewRequest("GET", apiUrl, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitLab API error: %s", resp.Status)
	}

	var issues []*GitLabIssue
	if err := json.NewDecoder(resp.Body).Decode(&issues); err != nil {
		return nil, err
	}

	return issues, nil
}

// CreateIssue 创建新议题
func (s *GitLabService) CreateIssue(accessToken string, projectID int64, title, description string, labels []string, assigneeID *int64) (*GitLabIssue, error) {
	url := fmt.Sprintf("%s/api/v4/projects/%d/issues", s.Config.GitLabURL, projectID)

	body := map[string]interface{}{
		"title":       title,
		"description": description,
		"labels":      labels,
	}

	if assigneeID != nil {
		body["assignee_id"] = *assigneeID
	}

	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, err
	}

	httpReq.Header.Set("Authorization", "Bearer "+accessToken)
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := s.HTTPClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("GitLab API error: %s", resp.Status)
	}

	var issue GitLabIssue
	if err := json.NewDecoder(resp.Body).Decode(&issue); err != nil {
		return nil, err
	}

	return &issue, nil
}

// GetMergeRequests 获取合并请求
func (s *GitLabService) GetMergeRequests(accessToken string, projectID int64, state string) ([]*GitLabMergeRequest, error) {
	apiUrl := fmt.Sprintf("%s/api/v4/projects/%d/merge_requests", s.Config.GitLabURL, projectID)

	params := url.Values{}
	if state != "" {
		params.Set("state", state)
	}

	if len(params) > 0 {
		apiUrl += "?" + params.Encode()
	}

	req, err := http.NewRequest("GET", apiUrl, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitLab API error: %s", resp.Status)
	}

	var mrs []*GitLabMergeRequest
	if err := json.NewDecoder(resp.Body).Decode(&mrs); err != nil {
		return nil, err
	}

	return mrs, nil
}

// ValidateRepositoryAccess 验证仓库访问权限
func (s *GitLabService) ValidateRepositoryAccess(accessToken string, projectID int64) error {
	_, err := s.GetProject(accessToken, projectID)
	return err
}

// GetUserProjectAccessLevel 获取用户在项目中的访问级别
func (s *GitLabService) GetUserProjectAccessLevel(accessToken string, projectID int64) (int, error) {
	// 首先获取当前用户信息
	user, err := s.GetUser(accessToken)
	if err != nil {
		return 0, err
	}

	// 获取项目成员信息
	apiUrl := fmt.Sprintf("%s/api/v4/projects/%d/members/all", s.Config.GitLabURL, projectID)

	req, err := http.NewRequest("GET", apiUrl, nil)
	if err != nil {
		return 0, err
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return 0, fmt.Errorf("项目不存在或用户无权限")
	}

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("获取项目成员失败: %s", resp.Status)
	}

	var members []struct {
		ID          int `json:"id"`
		AccessLevel int `json:"access_level"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&members); err != nil {
		return 0, err
	}

	// 查找当前用户的访问级别
	for _, member := range members {
		if member.ID == int(user.ID) {
			return member.AccessLevel, nil
		}
	}

	// 如果用户不是项目成员，检查项目是否为公开项目
	project, err := s.GetProject(accessToken, projectID)
	if err != nil {
		return 0, err
	}

	// 如果是公开项目，返回Guest级别权限
	if project.Visibility == "public" {
		return 10, nil // Guest level
	}

	return 0, fmt.Errorf("用户不是项目成员且项目非公开")
}

// GetRepositoryTree 获取仓库文件树
func (s *GitLabService) GetRepositoryTree(accessToken string, projectID int64, path string, ref string) ([]map[string]interface{}, error) {
	encodedPath := url.PathEscape(path)
	apiUrl := fmt.Sprintf("%s/api/v4/projects/%d/repository/tree?path=%s&ref=%s",
		s.Config.GitLabURL, projectID, encodedPath, ref)

	req, err := http.NewRequest("GET", apiUrl, nil)
	if err != nil {
		return nil, fmt.Errorf("创建HTTP请求失败: %v", err)
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("GitLab服务器连接失败: %v", err)
	}
	defer resp.Body.Close()

	// 读取响应体用于错误诊断
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %v", err)
	}

	if resp.StatusCode == 401 {
		return nil, fmt.Errorf("GitLab访问令牌无效或已过期，请重新登录")
	} else if resp.StatusCode == 403 {
		return nil, fmt.Errorf("没有访问该项目的权限")
	} else if resp.StatusCode == 404 {
		return nil, fmt.Errorf("项目不存在或路径无效")
	} else if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitLab API错误 (%d): %s", resp.StatusCode, string(body))
	}

	var tree []map[string]interface{}
	if err := json.Unmarshal(body, &tree); err != nil {
		return nil, fmt.Errorf("解析响应数据失败: %v", err)
	}

	// 为每个文件设置默认值，避免获取提交信息时的额外延迟
	for i := range tree {
		// 设置默认值
		tree[i]["last_commit_message"] = ""
		tree[i]["last_commit_date"] = ""
		tree[i]["last_update"] = ""
		tree[i]["last_commit_author"] = ""
	}

	return tree, nil
}

// getFileLastCommit 获取文件的最后提交信息
func (s *GitLabService) getFileLastCommit(accessToken string, projectID int64, filePath string, ref string) (map[string]interface{}, error) {
	encodedPath := url.PathEscape(filePath)
	apiUrl := fmt.Sprintf("%s/api/v4/projects/%d/repository/commits?path=%s&ref_name=%s&per_page=1",
		s.Config.GitLabURL, projectID, encodedPath, ref)

	req, err := http.NewRequest("GET", apiUrl, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitLab API error: %s", resp.Status)
	}

	var commits []map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&commits); err != nil {
		return nil, err
	}

	if len(commits) > 0 {
		return commits[0], nil
	}

	return nil, nil
}

// SearchFiles 搜索仓库中的文件
func (s *GitLabService) SearchFiles(accessToken string, projectID int64, search string) ([]map[string]interface{}, error) {
	url := fmt.Sprintf("%s/api/v4/projects/%d/search?scope=blobs&search=%s",
		s.Config.GitLabURL, projectID, url.QueryEscape(search))

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitLab API error: %s", resp.Status)
	}

	var results []map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&results); err != nil {
		return nil, err
	}

	return results, nil
}

// GetProjectIssues 获取项目Issues列表
func (s *GitLabService) GetProjectIssues(accessToken string, projectID int64, page, perPage int) ([]GitLabIssue, error) {
	url := fmt.Sprintf("%s/api/v4/projects/%d/issues?page=%d&per_page=%d",
		s.Config.GitLabURL, projectID, page, perPage)

	// 使用context控制请求超时，防止资源泄漏
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	resp, err := s.makeRequest(ctx, "GET", url, accessToken, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("获取Issues失败: %s - %s", resp.Status, string(body))
	}

	var issues []GitLabIssue
	if err := json.NewDecoder(resp.Body).Decode(&issues); err != nil {
		return nil, err
	}

	return issues, nil
}

// GetIssue 获取单个Issue
func (s *GitLabService) GetIssue(accessToken string, projectID, issueIID int64) (*GitLabIssue, error) {
	url := fmt.Sprintf("%s/api/v4/projects/%d/issues/%d", s.Config.GitLabURL, projectID, issueIID)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitLab API error: %s", resp.Status)
	}

	var issue GitLabIssue
	if err := json.NewDecoder(resp.Body).Decode(&issue); err != nil {
		return nil, err
	}

	return &issue, nil
}

// GetProjectBranches 获取项目分支列表
func (s *GitLabService) GetProjectBranches(accessToken string, projectID int64) ([]map[string]interface{}, error) {
	apiUrl := fmt.Sprintf("%s/api/v4/projects/%d/repository/branches", s.Config.GitLabURL, projectID)

	req, err := http.NewRequest("GET", apiUrl, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitLab API error: %s", resp.Status)
	}

	var branches []map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&branches); err != nil {
		return nil, err
	}

	return branches, nil
}

// GetBranchInfo 获取分支信息
func (s *GitLabService) GetBranchInfo(accessToken string, projectID int64, branchName string) (map[string]interface{}, error) {
	encodedBranch := url.PathEscape(branchName)
	apiUrl := fmt.Sprintf("%s/api/v4/projects/%d/repository/branches/%s", s.Config.GitLabURL, projectID, encodedBranch)

	req, err := http.NewRequest("GET", apiUrl, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitLab API error: %s", resp.Status)
	}

	var branchInfo map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&branchInfo); err != nil {
		return nil, err
	}

	return branchInfo, nil
}

// GetBranchCommits 获取分支的提交历史
func (s *GitLabService) GetBranchCommits(accessToken string, projectID int64, branchName string) ([]map[string]interface{}, error) {
	apiUrl := fmt.Sprintf("%s/api/v4/projects/%d/repository/commits?ref_name=%s&per_page=20",
		s.Config.GitLabURL, projectID, url.QueryEscape(branchName))

	req, err := http.NewRequest("GET", apiUrl, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitLab API error: %s", resp.Status)
	}

	var commits []map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&commits); err != nil {
		return nil, err
	}

	return commits, nil
}

// DeleteProject 删除GitLab项目
func (s *GitLabService) DeleteProject(accessToken string, projectID int64) error {
	url := fmt.Sprintf("%s/api/v4/projects/%d", s.Config.GitLabURL, projectID)

	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := s.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// GitLab删除项目成功返回202 Accepted
	if resp.StatusCode != http.StatusAccepted && resp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("删除GitLab项目失败: %s - %s", resp.Status, string(body))
	}

	return nil
}

// ProjectMember GitLab项目成员结构
type ProjectMember struct {
	ID          int64  `json:"id"`
	Username    string `json:"username"`
	Name        string `json:"name"`
	Email       string `json:"email"`
	AvatarURL   string `json:"avatar_url"`
	AccessLevel int    `json:"access_level"`
	State       string `json:"state"`
}

// GetProjectMembers 获取GitLab项目成员列表
func (s *GitLabService) GetProjectMembers(accessToken string, projectID int64) ([]ProjectMember, error) {
	url := fmt.Sprintf("%s/api/v4/projects/%d/members", s.Config.GitLabURL, projectID)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := s.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("获取项目成员失败: %s - %s", resp.Status, string(body))
	}

	var members []ProjectMember
	if err := json.NewDecoder(resp.Body).Decode(&members); err != nil {
		return nil, err
	}

	return members, nil
}

// AddProjectMember 添加GitLab项目成员
func (s *GitLabService) AddProjectMember(accessToken string, projectID int64, username string, accessLevel int) error {
	url := fmt.Sprintf("%s/api/v4/projects/%d/members", s.Config.GitLabURL, projectID)

	data := map[string]interface{}{
		"user_id":      username, // 可以是用户名或用户ID
		"access_level": accessLevel,
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("添加项目成员失败: %s - %s", resp.Status, string(body))
	}

	return nil
}

// RemoveProjectMember 移除GitLab项目成员
func (s *GitLabService) RemoveProjectMember(accessToken string, projectID int64, userID int64) error {
	url := fmt.Sprintf("%s/api/v4/projects/%d/members/%d", s.Config.GitLabURL, projectID, userID)

	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := s.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("移除项目成员失败: %s - %s", resp.Status, string(body))
	}

	return nil
}

// GetUserByID 根据用户ID获取GitLab用户信息
func (s *GitLabService) GetUserByID(accessToken string, userID int64) (*GitLabAPIUser, error) {
	url := fmt.Sprintf("%s/api/v4/users/%d", s.Config.GitLabURL, userID)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := s.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("获取用户信息失败: %s - %s", resp.Status, string(body))
	}

	var user GitLabAPIUser
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return nil, err
	}

	return &user, nil
}

// CreateProjectIssue 创建项目Issue
func (s *GitLabService) CreateProjectIssue(accessToken string, projectID int64, title, description string, labels []string) (*GitLabIssue, error) {
	url := fmt.Sprintf("%s/api/v4/projects/%d/issues", s.Config.GitLabURL, projectID)

	data := map[string]interface{}{
		"title":       title,
		"description": description,
	}
	if len(labels) > 0 {
		data["labels"] = strings.Join(labels, ",")
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	// 使用context控制请求超时
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	resp, err := s.makeRequest(ctx, "POST", url, accessToken, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("创建Issue失败: %s - %s", resp.Status, string(body))
	}

	var issue GitLabIssue
	if err := json.NewDecoder(resp.Body).Decode(&issue); err != nil {
		return nil, err
	}

	return &issue, nil
}

// GetIssueNotes 获取Issue评论列表
func (s *GitLabService) GetIssueNotes(accessToken string, projectID, issueIID int64, page, perPage int) ([]GitLabIssueNote, error) {
	url := fmt.Sprintf("%s/api/v4/projects/%d/issues/%d/notes?page=%d&per_page=%d&sort=asc&order_by=created_at",
		s.Config.GitLabURL, projectID, issueIID, page, perPage)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := s.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("获取评论失败: %s - %s", resp.Status, string(body))
	}

	var notes []GitLabIssueNote
	if err := json.NewDecoder(resp.Body).Decode(&notes); err != nil {
		return nil, err
	}

	return notes, nil
}

// CreateIssueNote 创建Issue评论
func (s *GitLabService) CreateIssueNote(accessToken string, projectID, issueIID int64, body string) (*GitLabIssueNote, error) {
	url := fmt.Sprintf("%s/api/v4/projects/%d/issues/%d/notes", s.Config.GitLabURL, projectID, issueIID)

	data := map[string]interface{}{
		"body": body,
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("创建评论失败: %s - %s", resp.Status, string(body))
	}

	var note GitLabIssueNote
	if err := json.NewDecoder(resp.Body).Decode(&note); err != nil {
		return nil, err
	}

	return &note, nil
}

// GetIssueAwardEmojis 获取Issue的表情反应列表
func (s *GitLabService) GetIssueAwardEmojis(accessToken string, projectID, issueIID int64) ([]GitLabAwardEmoji, error) {
	url := fmt.Sprintf("%s/api/v4/projects/%d/issues/%d/award_emoji", s.Config.GitLabURL, projectID, issueIID)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := s.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("获取表情反应失败: %s - %s", resp.Status, string(body))
	}

	var emojis []GitLabAwardEmoji
	if err := json.NewDecoder(resp.Body).Decode(&emojis); err != nil {
		return nil, err
	}

	return emojis, nil
}

// AddIssueAwardEmoji 给Issue添加表情反应
func (s *GitLabService) AddIssueAwardEmoji(accessToken string, projectID, issueIID int64, emojiName string) (*GitLabAwardEmoji, error) {
	url := fmt.Sprintf("%s/api/v4/projects/%d/issues/%d/award_emoji", s.Config.GitLabURL, projectID, issueIID)

	data := map[string]interface{}{
		"name": emojiName,
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("添加表情反应失败: %s - %s", resp.Status, string(body))
	}

	var emoji GitLabAwardEmoji
	if err := json.NewDecoder(resp.Body).Decode(&emoji); err != nil {
		return nil, err
	}

	return &emoji, nil
}

// RemoveIssueAwardEmoji 移除Issue表情反应
func (s *GitLabService) RemoveIssueAwardEmoji(accessToken string, projectID, issueIID, emojiID int64) error {
	url := fmt.Sprintf("%s/api/v4/projects/%d/issues/%d/award_emoji/%d", s.Config.GitLabURL, projectID, issueIID, emojiID)

	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := s.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("移除表情反应失败: %s - %s", resp.Status, string(body))
	}

	return nil
}

// FindUserAwardEmoji 查找用户的特定表情反应
func (s *GitLabService) FindUserAwardEmoji(accessToken string, projectID, issueIID int64, emojiName string, userID int64) (*GitLabAwardEmoji, error) {
	emojis, err := s.GetIssueAwardEmojis(accessToken, projectID, issueIID)
	if err != nil {
		return nil, err
	}

	for _, emoji := range emojis {
		if emoji.Name == emojiName && emoji.User.ID == userID {
			return &emoji, nil
		}
	}

	return nil, nil // 未找到
}
