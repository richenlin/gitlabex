package services

import (
	"bytes"
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
	return &GitLabService{
		Config: cfg,
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// GitLabUser GitLab用户信息
type GitLabUser struct {
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
	Owner *GitLabUser `json:"owner,omitempty"`
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
	ID        string      `json:"id"`
	ShortID   string      `json:"short_id"`
	Title     string      `json:"title"`
	Author    *GitLabUser `json:"author"`
	Committer *GitLabUser `json:"committer"`
	Message   string      `json:"message"`
	CreatedAt time.Time   `json:"created_at"`
	WebURL    string      `json:"web_url"`
}

// GitLabIssue GitLab议题信息
type GitLabIssue struct {
	ID          int64       `json:"id"`
	IID         int64       `json:"iid"`
	Title       string      `json:"title"`
	Description string      `json:"description"`
	State       string      `json:"state"`
	Author      *GitLabUser `json:"author"`
	Assignee    *GitLabUser `json:"assignee"`
	Labels      []string    `json:"labels"`
	CreatedAt   string      `json:"created_at"`
	UpdatedAt   string      `json:"updated_at"`
	WebURL      string      `json:"web_url"`
}

// GitLabMergeRequest GitLab合并请求信息
type GitLabMergeRequest struct {
	ID           int64       `json:"id"`
	IID          int64       `json:"iid"`
	Title        string      `json:"title"`
	Description  string      `json:"description"`
	State        string      `json:"state"`
	SourceBranch string      `json:"source_branch"`
	TargetBranch string      `json:"target_branch"`
	Author       *GitLabUser `json:"author"`
	Assignee     *GitLabUser `json:"assignee"`
	CreatedAt    string      `json:"created_at"`
	UpdatedAt    string      `json:"updated_at"`
	WebURL       string      `json:"web_url"`
	MergedAt     *string     `json:"merged_at,omitempty"`
}

// CreateProjectRequest 创建项目请求
type CreateProjectRequest struct {
	Name                 string `json:"name"`
	Path                 string `json:"path,omitempty"`
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
	User       *GitLabUser     `json:"user"`
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
func (s *GitLabService) GetUser(accessToken string) (*GitLabUser, error) {
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

	var user GitLabUser
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

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitLab API error: %s", resp.Status)
	}

	var project GitLabProject
	if err := json.NewDecoder(resp.Body).Decode(&project); err != nil {
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
		return nil, fmt.Errorf("GitLab API error: %s", resp.Status)
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

// GetProjectIssues 获取项目Issues
func (s *GitLabService) GetProjectIssues(accessToken string, projectID int64) ([]GitLabIssue, error) {
	url := fmt.Sprintf("%s/api/v4/projects/%d/issues", s.Config.GitLabURL, projectID)

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
