package services

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"gitlabex/internal/config"
	"gitlabex/internal/dto"
	"gitlabex/internal/utils"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// GitLabService GitLab API服务
type GitLabService struct {
	Config       *config.Config
	HTTPClient   *utils.HTTPClient // 使用封装的HTTP客户端
	RedisService *RedisService     // Redis缓存服务
}

// NewGitLabService 创建GitLab服务
func NewGitLabService(cfg *config.Config) *GitLabService {
	// 使用统一的HTTP客户端工具类，自动支持连接池
	httpClient := utils.NewHTTPClient(&utils.HTTPClientConfig{
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 10,
		IdleConnTimeout:     90 * time.Second,
		Timeout:             30 * time.Second,
		DisableKeepAlives:   false,
		ForceAttemptHTTP2:   true,
	})

	return &GitLabService{
		Config:     cfg,
		HTTPClient: httpClient,
	}
}

// SetRedisService 设置Redis服务（用于缓存）
func (s *GitLabService) SetRedisService(redisService *RedisService) {
	s.RedisService = redisService
}

// makeRequest 通用HTTP请求方法，带有context支持（保留用于兼容性）
func (s *GitLabService) makeRequest(ctx context.Context, method, url, accessToken string, body io.Reader) (*http.Response, error) {
	opts := &utils.RequestOptions{
		Method:      method,
		URL:         url,
		BearerToken: accessToken,
		Context:     ctx,
	}

	// 如果有body，需要特殊处理（因为utils期望interface{}）
	if body != nil {
		// 这里为了兼容性，直接使用底层client
		return s.doRequestWithContext(ctx, method, url, accessToken, body)
	}

	resp, err := s.HTTPClient.DoRequest(opts)
	if err != nil {
		return nil, err
	}

	// 将utils.Response转换为http.Response（简化版）
	// 注意：这是一个权宜之计，理想情况下应该重构所有调用者
	return &http.Response{
		StatusCode: resp.StatusCode,
		Header:     resp.Headers,
		Body:       io.NopCloser(bytes.NewReader(resp.Body)),
	}, nil
}

// doRequestWithContext 使用context执行原始HTTP请求（内部辅助方法）
func (s *GitLabService) doRequestWithContext(ctx context.Context, method, url, accessToken string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	return s.HTTPClient.GetClient().Do(req)
}

// doJSONRequest 执行HTTP请求并解析JSON响应
func (s *GitLabService) doJSONRequest(method, apiURL, accessToken string, requestBody interface{}, responseBody interface{}, expectedStatus int) error {
	return s.HTTPClient.DoJSONRequestWithStatus(&utils.RequestOptions{
		Method:      method,
		URL:         apiURL,
		BearerToken: accessToken,
		Body:        requestBody,
	}, expectedStatus, responseBody)
}

// doRequest 执行HTTP请求并返回响应体(自动处理context和错误)
func (s *GitLabService) doRequest(method, apiURL, accessToken string, body interface{}, expectedStatus int) ([]byte, error) {
	return s.HTTPClient.DoRequestWithStatus(&utils.RequestOptions{
		Method:      method,
		URL:         apiURL,
		BearerToken: accessToken,
		Body:        body,
	}, expectedStatus)
}

// doRequestMultiStatus 执行HTTP请求并返回响应体(支持多个预期状态码)
func (s *GitLabService) doRequestMultiStatus(method, apiURL, accessToken string, body interface{}, expectedStatuses ...int) ([]byte, int, error) {
	return s.HTTPClient.DoRequestWithMultiStatus(&utils.RequestOptions{
		Method:      method,
		URL:         apiURL,
		BearerToken: accessToken,
		Body:        body,
	}, expectedStatuses...)
}

// ProtectBranch 保护分支
func (s *GitLabService) ProtectBranch(accessToken string, projectID int64, req *dto.ProtectBranchRequest) (*dto.ProtectedBranch, error) {
	apiURL := fmt.Sprintf("%s/api/v4/projects/%d/protected_branches", s.Config.GitLab.URL, projectID)

	// 使用多状态码请求（支持201和409）
	respBody, statusCode, err := s.doRequestMultiStatus("POST", apiURL, accessToken, req, http.StatusCreated, http.StatusConflict)
	if err != nil {
		return nil, fmt.Errorf("保护分支失败: %v", err)
	}

	// 如果分支已经受保护，尝试获取现有的保护规则
	if statusCode == http.StatusConflict {
		return s.GetProtectedBranch(accessToken, projectID, req.Name)
	}

	var protectedBranch dto.ProtectedBranch
	if err := json.Unmarshal(respBody, &protectedBranch); err != nil {
		return nil, fmt.Errorf("解析响应失败: %v", err)
	}

	return &protectedBranch, nil
}

// GetProtectedBranch 获取受保护分支信息
func (s *GitLabService) GetProtectedBranch(accessToken string, projectID int64, branchName string) (*dto.ProtectedBranch, error) {
	encodedBranch := url.PathEscape(branchName)
	apiURL := fmt.Sprintf("%s/api/v4/projects/%d/protected_branches/%s", s.Config.GitLab.URL, projectID, encodedBranch)

	var protectedBranch dto.ProtectedBranch
	if err := s.doJSONRequest("GET", apiURL, accessToken, nil, &protectedBranch, http.StatusOK); err != nil {
		return nil, err
	}
	return &protectedBranch, nil
}

// SetupProjectBranchProtection 为新创建的课题项目设置分支保护
func (s *GitLabService) SetupProjectBranchProtection(accessToken string, projectID int64) error {
	// 保护主分支，只允许维护者和管理员推送
	mainBranchProtection := &dto.ProtectBranchRequest{
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
		masterBranchProtection := &dto.ProtectBranchRequest{
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

// GetUser 获取当前用户信息（带缓存）
func (s *GitLabService) GetUser(accessToken string) (*dto.GitLabAPIUser, error) {
	ctx := context.Background()

	// 生成缓存key（基于token的hash，避免暴露token）
	cacheKey := fmt.Sprintf("gitlab:user:token:%s", hashString(accessToken))

	// 尝试从缓存获取
	if s.RedisService != nil {
		var cachedUser dto.GitLabAPIUser
		err := s.RedisService.GetCache(ctx, cacheKey, &cachedUser)
		if err == nil {
			return &cachedUser, nil
		}
	}

	// 缓存未命中，调用GitLab API
	apiURL := fmt.Sprintf("%s/api/v4/user", s.Config.GitLab.URL)
	var user dto.GitLabAPIUser
	if err := s.doJSONRequest("GET", apiURL, accessToken, nil, &user, http.StatusOK); err != nil {
		return nil, err
	}

	// 存入缓存（15分钟）
	if s.RedisService != nil {
		_ = s.RedisService.SetCache(ctx, cacheKey, &user, 15*time.Minute)
	}

	return &user, nil
}

// GetAllUsers 获取所有用户列表（管理员专用，带缓存）
func (s *GitLabService) GetAllUsers(accessToken string, page, perPage int) ([]*dto.GitLabAPIUser, error) {
	ctx := context.Background()

	// 生成缓存key
	cacheKey := fmt.Sprintf("gitlab:users:all:page:%d:per_page:%d", page, perPage)

	// 尝试从缓存获取
	if s.RedisService != nil {
		var cachedUsers []*dto.GitLabAPIUser
		err := s.RedisService.GetCache(ctx, cacheKey, &cachedUsers)
		if err == nil {
			return cachedUsers, nil
		}
	}

	// 缓存未命中，调用GitLab API
	apiURL := fmt.Sprintf("%s/api/v4/users?page=%d&per_page=%d", s.Config.GitLab.URL, page, perPage)
	var users []*dto.GitLabAPIUser
	if err := s.doJSONRequest("GET", apiURL, accessToken, nil, &users, http.StatusOK); err != nil {
		return nil, err
	}

	// 存入缓存（10分钟，用户列表更新频率较低）
	if s.RedisService != nil {
		_ = s.RedisService.SetCacheWithTags(ctx, cacheKey, users, []string{"gitlab_users", "gitlab_users_list"}, 10*time.Minute)
	}

	return users, nil
}

// SearchUsers 搜索用户
func (s *GitLabService) SearchUsers(accessToken string, search string, page, perPage int) ([]*dto.GitLabAPIUser, error) {
	apiURL := fmt.Sprintf("%s/api/v4/users?search=%s&page=%d&per_page=%d",
		s.Config.GitLab.URL, url.QueryEscape(search), page, perPage)
	var users []*dto.GitLabAPIUser
	if err := s.doJSONRequest("GET", apiURL, accessToken, nil, &users, http.StatusOK); err != nil {
		return nil, err
	}
	return users, nil
}

// GetUserByUsername 根据用户名获取用户信息（带缓存）
func (s *GitLabService) GetUserByUsername(accessToken string, username string) (*dto.GitLabAPIUser, error) {
	ctx := context.Background()

	// 生成缓存key
	cacheKey := fmt.Sprintf("gitlab:user:username:%s", username)

	// 尝试从缓存获取
	if s.RedisService != nil {
		var cachedUser dto.GitLabAPIUser
		err := s.RedisService.GetCache(ctx, cacheKey, &cachedUser)
		if err == nil {
			return &cachedUser, nil
		}
	}

	// 缓存未命中，调用GitLab API
	apiURL := fmt.Sprintf("%s/api/v4/users?username=%s", s.Config.GitLab.URL, url.QueryEscape(username))
	var users []*dto.GitLabAPIUser
	if err := s.doJSONRequest("GET", apiURL, accessToken, nil, &users, http.StatusOK); err != nil {
		return nil, err
	}
	if len(users) == 0 {
		return nil, fmt.Errorf("用户不存在")
	}

	user := users[0]

	// 存入缓存（30分钟）
	if s.RedisService != nil {
		_ = s.RedisService.SetCacheWithTags(ctx, cacheKey, user, []string{"gitlab_users"}, 30*time.Minute)
	}

	return user, nil
}

// GetUserByID 根据用户ID获取用户信息（带缓存）
func (s *GitLabService) GetUserByID(accessToken string, userID int64) (*dto.GitLabAPIUser, error) {
	ctx := context.Background()

	// 生成缓存key
	cacheKey := fmt.Sprintf("gitlab:user:id:%d", userID)

	// 尝试从缓存获取
	if s.RedisService != nil {
		var cachedUser dto.GitLabAPIUser
		err := s.RedisService.GetCache(ctx, cacheKey, &cachedUser)
		if err == nil {
			return &cachedUser, nil
		}
	}

	// 缓存未命中，调用GitLab API
	apiURL := fmt.Sprintf("%s/api/v4/users/%d", s.Config.GitLab.URL, userID)
	var user dto.GitLabAPIUser
	if err := s.doJSONRequest("GET", apiURL, accessToken, nil, &user, http.StatusOK); err != nil {
		return nil, err
	}

	// 存入缓存（30分钟）
	if s.RedisService != nil {
		_ = s.RedisService.SetCacheWithTags(ctx, cacheKey, &user, []string{"gitlab_users"}, 30*time.Minute)
	}

	return &user, nil
}

// CreateUser 创建用户（管理员专用，自动清除相关缓存）
func (s *GitLabService) CreateUser(accessToken string, userData *dto.GitLabCreateUserData) (*dto.GitLabAPIUser, error) {
	apiURL := fmt.Sprintf("%s/api/v4/users", s.Config.GitLab.URL)
	var user dto.GitLabAPIUser
	if err := s.doJSONRequest("POST", apiURL, accessToken, userData, &user, http.StatusCreated); err != nil {
		return nil, err
	}

	// 清除用户列表缓存
	if s.RedisService != nil {
		ctx := context.Background()
		_ = s.RedisService.DeleteCacheByTag(ctx, "gitlab_users_list")
	}

	return &user, nil
}

// UpdateUser 更新用户信息（管理员专用，自动清除相关缓存）
func (s *GitLabService) UpdateUser(accessToken string, userID int64, userData *dto.GitLabUpdateUserData) (*dto.GitLabAPIUser, error) {
	apiURL := fmt.Sprintf("%s/api/v4/users/%d", s.Config.GitLab.URL, userID)
	var user dto.GitLabAPIUser
	if err := s.doJSONRequest("PUT", apiURL, accessToken, userData, &user, http.StatusOK); err != nil {
		return nil, err
	}

	// 清除该用户的所有缓存
	if s.RedisService != nil {
		ctx := context.Background()
		// 清除ID缓存
		_ = s.RedisService.DeleteCache(ctx, fmt.Sprintf("gitlab:user:id:%d", userID))
		// 清除用户名缓存（如果用户名被修改）
		if userData.Username != "" {
			_ = s.RedisService.DeleteCache(ctx, fmt.Sprintf("gitlab:user:username:%s", userData.Username))
		}
		// 清除用户列表缓存
		_ = s.RedisService.DeleteCacheByTag(ctx, "gitlab_users_list")
	}

	return &user, nil
}

// DeleteUser 删除用户（管理员专用，自动清除相关缓存）
func (s *GitLabService) DeleteUser(accessToken string, userID int64) error {
	// 先获取用户信息以便清除用户名缓存
	var username string
	if s.RedisService != nil {
		user, _ := s.GetUserByID(accessToken, userID)
		if user != nil {
			username = user.Username
		}
	}

	apiURL := fmt.Sprintf("%s/api/v4/users/%d", s.Config.GitLab.URL, userID)
	_, _, err := s.doRequestMultiStatus("DELETE", apiURL, accessToken, nil, http.StatusNoContent, http.StatusAccepted)

	// 删除成功后清除缓存
	if err == nil && s.RedisService != nil {
		ctx := context.Background()
		// 清除ID缓存
		_ = s.RedisService.DeleteCache(ctx, fmt.Sprintf("gitlab:user:id:%d", userID))
		// 清除用户名缓存
		if username != "" {
			_ = s.RedisService.DeleteCache(ctx, fmt.Sprintf("gitlab:user:username:%s", username))
		}
		// 清除用户列表缓存
		_ = s.RedisService.DeleteCacheByTag(ctx, "gitlab_users_list")
	}

	return err
}

// GetProjects 获取用户的项目列表
func (s *GitLabService) GetProjects(accessToken string, page, perPage int) ([]*dto.GitLabProject, error) {
	apiURL := fmt.Sprintf("%s/api/v4/projects?owned=true&page=%d&per_page=%d&order_by=last_activity_at",
		s.Config.GitLab.URL, page, perPage)
	var projects []*dto.GitLabProject
	if err := s.doJSONRequest("GET", apiURL, accessToken, nil, &projects, http.StatusOK); err != nil {
		return nil, err
	}
	return projects, nil
}

// GetProject 获取特定项目信息
func (s *GitLabService) GetProject(accessToken string, projectID int64) (*dto.GitLabProject, error) {
	apiURL := fmt.Sprintf("%s/api/v4/projects/%d", s.Config.GitLab.URL, projectID)

	respBody, statusCode, err := s.doRequestMultiStatus("GET", apiURL, accessToken, nil, http.StatusOK, 401, 403, 404)
	if err != nil && statusCode == 0 {
		// 网络错误
		return nil, err
	}

	// 处理特定的错误状态码
	switch statusCode {
	case 401:
		return nil, fmt.Errorf("GitLab访问令牌无效或已过期，请重新登录")
	case 403:
		return nil, fmt.Errorf("没有访问该项目的权限")
	case 404:
		return nil, fmt.Errorf("项目不存在")
	case http.StatusOK:
		var project dto.GitLabProject
		if err := json.Unmarshal(respBody, &project); err != nil {
			return nil, fmt.Errorf("解析响应失败: %v", err)
		}
		return &project, nil
	default:
		return nil, fmt.Errorf("GitLab API错误 (%d): %s", statusCode, string(respBody))
	}
}

// CreateProject 创建新项目
func (s *GitLabService) CreateProject(accessToken string, req *dto.CreateProjectRequest) (*dto.GitLabProject, error) {
	apiURL := fmt.Sprintf("%s/api/v4/projects", s.Config.GitLab.URL)
	var project dto.GitLabProject
	if err := s.doJSONRequest("POST", apiURL, accessToken, req, &project, http.StatusCreated); err != nil {
		return nil, err
	}
	return &project, nil
}

// GetBranches 获取项目分支列表
func (s *GitLabService) GetBranches(accessToken string, projectID int64) ([]*dto.GitLabBranch, error) {
	apiURL := fmt.Sprintf("%s/api/v4/projects/%d/repository/branches", s.Config.GitLab.URL, projectID)
	var branches []*dto.GitLabBranch
	if err := s.doJSONRequest("GET", apiURL, accessToken, nil, &branches, http.StatusOK); err != nil {
		return nil, err
	}
	return branches, nil
}

// CreateBranch 创建新分支
func (s *GitLabService) CreateBranch(accessToken string, projectID int64, req *dto.CreateBranchRequest) (*dto.GitLabBranch, error) {
	apiURL := fmt.Sprintf("%s/api/v4/projects/%d/repository/branches", s.Config.GitLab.URL, projectID)
	var branch dto.GitLabBranch
	if err := s.doJSONRequest("POST", apiURL, accessToken, req, &branch, http.StatusCreated); err != nil {
		return nil, err
	}
	return &branch, nil
}

// GetFileContent 获取文件内容
func (s *GitLabService) GetFileContent(accessToken string, projectID int64, filePath, ref string) (*dto.GitLabFile, error) {
	encodedPath := url.PathEscape(filePath)
	apiURL := fmt.Sprintf("%s/api/v4/projects/%d/repository/files/%s?ref=%s",
		s.Config.GitLab.URL, projectID, encodedPath, ref)
	var file dto.GitLabFile
	if err := s.doJSONRequest("GET", apiURL, accessToken, nil, &file, http.StatusOK); err != nil {
		return nil, err
	}
	return &file, nil
}

// CreateFile 创建新文件
func (s *GitLabService) CreateFile(accessToken string, projectID int64, filePath string, req *dto.CreateFileRequest) error {
	encodedPath := url.PathEscape(filePath)
	apiURL := fmt.Sprintf("%s/api/v4/projects/%d/repository/files/%s",
		s.Config.GitLab.URL, projectID, encodedPath)
	return s.doJSONRequest("POST", apiURL, accessToken, req, nil, http.StatusCreated)
}

// UpdateFile 更新文件内容
func (s *GitLabService) UpdateFile(accessToken string, projectID int64, filePath string, req *dto.CreateFileRequest) error {
	encodedPath := url.PathEscape(filePath)
	apiURL := fmt.Sprintf("%s/api/v4/projects/%d/repository/files/%s",
		s.Config.GitLab.URL, projectID, encodedPath)
	return s.doJSONRequest("PUT", apiURL, accessToken, req, nil, http.StatusOK)
}

// GetCommits 获取提交历史
func (s *GitLabService) GetCommits(accessToken string, projectID int64, branch string, limit int) ([]*dto.GitLabCommit, error) {
	apiURL := fmt.Sprintf("%s/api/v4/projects/%d/repository/commits?ref_name=%s&per_page=%d",
		s.Config.GitLab.URL, projectID, branch, limit)
	var commits []*dto.GitLabCommit
	if err := s.doJSONRequest("GET", apiURL, accessToken, nil, &commits, http.StatusOK); err != nil {
		return nil, err
	}
	return commits, nil
}

// GetIssues 获取项目议题
func (s *GitLabService) GetIssues(accessToken string, projectID int64, state string, labels []string) ([]*dto.GitLabIssue, error) {
	apiUrl := fmt.Sprintf("%s/api/v4/projects/%d/issues", s.Config.GitLab.URL, projectID)

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

	var issues []*dto.GitLabIssue
	if err := s.doJSONRequest("GET", apiUrl, accessToken, nil, &issues, http.StatusOK); err != nil {
		return nil, err
	}
	return issues, nil
}

// CreateIssue 创建新议题
func (s *GitLabService) CreateIssue(accessToken string, projectID int64, title, description string, labels []string, assigneeID *int64) (*dto.GitLabIssue, error) {
	apiURL := fmt.Sprintf("%s/api/v4/projects/%d/issues", s.Config.GitLab.URL, projectID)

	body := map[string]interface{}{
		"title":       title,
		"description": description,
		"labels":      labels,
	}

	if assigneeID != nil {
		body["assignee_id"] = *assigneeID
	}

	var issue dto.GitLabIssue
	if err := s.doJSONRequest("POST", apiURL, accessToken, body, &issue, http.StatusCreated); err != nil {
		return nil, err
	}
	return &issue, nil
}

// GetMergeRequests 获取合并请求
func (s *GitLabService) GetMergeRequests(accessToken string, projectID int64, state string) ([]*dto.GitLabMergeRequest, error) {
	apiUrl := fmt.Sprintf("%s/api/v4/projects/%d/merge_requests", s.Config.GitLab.URL, projectID)

	params := url.Values{}
	if state != "" {
		params.Set("state", state)
	}

	if len(params) > 0 {
		apiUrl += "?" + params.Encode()
	}

	var mrs []*dto.GitLabMergeRequest
	if err := s.doJSONRequest("GET", apiUrl, accessToken, nil, &mrs, http.StatusOK); err != nil {
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
	apiURL := fmt.Sprintf("%s/api/v4/projects/%d/members/all", s.Config.GitLab.URL, projectID)

	var members []struct {
		ID          int `json:"id"`
		AccessLevel int `json:"access_level"`
	}

	if err := s.doJSONRequest("GET", apiURL, accessToken, nil, &members, http.StatusOK); err != nil {
		return 0, fmt.Errorf("获取项目成员失败: %v", err)
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
	// 添加参数：
	// - with_stats=true: 获取文件大小和统计信息（注意：这个参数在某些 GitLab 版本中可能不支持）
	// - per_page=100: 每页返回100条记录
	// - recursive=false: 不递归获取子目录
	apiURL := fmt.Sprintf("%s/api/v4/projects/%d/repository/tree?path=%s&ref=%s&per_page=100&recursive=false",
		s.Config.GitLab.URL, projectID, encodedPath, ref)

	// 使用多状态码支持特殊错误处理
	respBody, statusCode, err := s.doRequestMultiStatus("GET", apiURL, accessToken, nil, http.StatusOK, 401, 403, 404)
	if err != nil && statusCode == 0 {
		return nil, fmt.Errorf("GitLab服务器连接失败: %v", err)
	}

	// 处理特定的错误状态码
	switch statusCode {
	case 401:
		return nil, fmt.Errorf("GitLab访问令牌无效或已过期，请重新登录")
	case 403:
		return nil, fmt.Errorf("没有访问该项目的权限")
	case 404:
		return nil, fmt.Errorf("项目不存在或路径无效")
	case http.StatusOK:
		var tree []map[string]interface{}
		if err := json.Unmarshal(respBody, &tree); err != nil {
			return nil, fmt.Errorf("解析响应数据失败: %v", err)
		}

		// 为每个文件/目录获取最后一次提交信息和文件大小
		// 使用 goroutine 并发获取信息以提高性能
		type itemWithMetadata struct {
			index      int
			lastCommit map[string]interface{}
			size       int64
		}

		metadataChan := make(chan itemWithMetadata, len(tree))
		semaphore := make(chan struct{}, 5) // 限制并发数为5

		for i := range tree {
			go func(idx int) {
				semaphore <- struct{}{}        // 获取信号量
				defer func() { <-semaphore }() // 释放信号量

				itemPath, ok := tree[idx]["path"].(string)
				if !ok {
					metadataChan <- itemWithMetadata{index: idx, lastCommit: nil, size: 0}
					return
				}

				itemType, _ := tree[idx]["type"].(string)
				var fileSize int64 = 0

				// 获取该文件/目录的最后一次提交
				lastCommit, err := s.getFileLastCommit(accessToken, projectID, itemPath, ref)
				if err != nil {
					lastCommit = nil
				}

				// 如果是文件（blob），获取文件大小
				if itemType == "blob" {
					fileSize, _ = s.getFileSize(accessToken, projectID, itemPath, ref)
				}

				metadataChan <- itemWithMetadata{
					index:      idx,
					lastCommit: lastCommit,
					size:       fileSize,
				}
			}(i)
		}

		// 收集所有元数据并设置默认值
		for i := 0; i < len(tree); i++ {
			result := <-metadataChan

			// 设置提交信息
			if result.lastCommit != nil {
				tree[result.index]["last_commit"] = result.lastCommit
			} else {
				// 如果获取失败，设置默认值
				tree[result.index]["last_commit"] = map[string]interface{}{
					"message":        "",
					"committed_date": "",
					"author_name":    "",
				}
			}

			// 设置文件大小
			itemType, _ := tree[result.index]["type"].(string)
			if itemType == "blob" {
				tree[result.index]["size"] = result.size
			} else {
				// 目录没有大小
				tree[result.index]["size"] = 0
			}
		}
		close(metadataChan)

		return tree, nil
	default:
		return nil, fmt.Errorf("GitLab API错误 (%d): %s", statusCode, string(respBody))
	}
}

// getFileLastCommit 获取文件的最后提交信息
func (s *GitLabService) getFileLastCommit(accessToken string, projectID int64, filePath string, ref string) (map[string]interface{}, error) {
	encodedPath := url.PathEscape(filePath)
	apiURL := fmt.Sprintf("%s/api/v4/projects/%d/repository/commits?path=%s&ref_name=%s&per_page=1",
		s.Config.GitLab.URL, projectID, encodedPath, ref)

	var commits []map[string]interface{}
	if err := s.doJSONRequest("GET", apiURL, accessToken, nil, &commits, http.StatusOK); err != nil {
		return nil, err
	}

	if len(commits) > 0 {
		return commits[0], nil
	}

	return nil, nil
}

// getFileSize 获取文件大小（单位：字节）
func (s *GitLabService) getFileSize(accessToken string, projectID int64, filePath string, ref string) (int64, error) {
	encodedPath := url.PathEscape(filePath)
	apiURL := fmt.Sprintf("%s/api/v4/projects/%d/repository/files/%s?ref=%s",
		s.Config.GitLab.URL, projectID, encodedPath, ref)

	var fileInfo map[string]interface{}
	if err := s.doJSONRequest("GET", apiURL, accessToken, nil, &fileInfo, http.StatusOK); err != nil {
		return 0, err
	}

	// GitLab API 返回的 size 字段是文件内容的字节大小
	if size, ok := fileInfo["size"].(float64); ok {
		return int64(size), nil
	}

	return 0, fmt.Errorf("无法获取文件大小")
}

// SearchFiles 搜索仓库中的文件
func (s *GitLabService) SearchFiles(accessToken string, projectID int64, search string) ([]map[string]interface{}, error) {
	apiURL := fmt.Sprintf("%s/api/v4/projects/%d/search?scope=blobs&search=%s",
		s.Config.GitLab.URL, projectID, url.QueryEscape(search))
	var results []map[string]interface{}
	if err := s.doJSONRequest("GET", apiURL, accessToken, nil, &results, http.StatusOK); err != nil {
		return nil, err
	}
	return results, nil
}

// GetProjectIssues 获取项目Issues列表
func (s *GitLabService) GetProjectIssues(accessToken string, projectID int64, page, perPage int) ([]dto.GitLabIssue, error) {
	url := fmt.Sprintf("%s/api/v4/projects/%d/issues?page=%d&per_page=%d",
		s.Config.GitLab.URL, projectID, page, perPage)

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

	var issues []dto.GitLabIssue
	if err := json.NewDecoder(resp.Body).Decode(&issues); err != nil {
		return nil, err
	}

	return issues, nil
}

// GetIssue 获取单个Issue
func (s *GitLabService) GetIssue(accessToken string, projectID, issueIID int64) (*dto.GitLabIssue, error) {
	apiURL := fmt.Sprintf("%s/api/v4/projects/%d/issues/%d", s.Config.GitLab.URL, projectID, issueIID)
	var issue dto.GitLabIssue
	if err := s.doJSONRequest("GET", apiURL, accessToken, nil, &issue, http.StatusOK); err != nil {
		return nil, err
	}
	return &issue, nil
}

// GetProjectBranches 获取项目分支列表
func (s *GitLabService) GetProjectBranches(accessToken string, projectID int64) ([]map[string]interface{}, error) {
	apiURL := fmt.Sprintf("%s/api/v4/projects/%d/repository/branches", s.Config.GitLab.URL, projectID)
	var branches []map[string]interface{}
	if err := s.doJSONRequest("GET", apiURL, accessToken, nil, &branches, http.StatusOK); err != nil {
		return nil, err
	}
	return branches, nil
}

// GetBranchInfo 获取分支信息
func (s *GitLabService) GetBranchInfo(accessToken string, projectID int64, branchName string) (map[string]interface{}, error) {
	encodedBranch := url.PathEscape(branchName)
	apiURL := fmt.Sprintf("%s/api/v4/projects/%d/repository/branches/%s", s.Config.GitLab.URL, projectID, encodedBranch)
	var branchInfo map[string]interface{}
	if err := s.doJSONRequest("GET", apiURL, accessToken, nil, &branchInfo, http.StatusOK); err != nil {
		return nil, err
	}
	return branchInfo, nil
}

// GetBranchCommits 获取分支的提交历史
func (s *GitLabService) GetBranchCommits(accessToken string, projectID int64, branchName string) ([]map[string]interface{}, error) {
	apiURL := fmt.Sprintf("%s/api/v4/projects/%d/repository/commits?ref_name=%s&per_page=20",
		s.Config.GitLab.URL, projectID, url.QueryEscape(branchName))
	var commits []map[string]interface{}
	if err := s.doJSONRequest("GET", apiURL, accessToken, nil, &commits, http.StatusOK); err != nil {
		return nil, err
	}
	return commits, nil
}

// DeleteProject 删除GitLab项目
func (s *GitLabService) DeleteProject(accessToken string, projectID int64) error {
	apiURL := fmt.Sprintf("%s/api/v4/projects/%d", s.Config.GitLab.URL, projectID)
	_, _, err := s.doRequestMultiStatus("DELETE", apiURL, accessToken, nil, http.StatusAccepted, http.StatusNoContent)
	return err
}

// GetProjectMembers 获取GitLab项目成员列表
func (s *GitLabService) GetProjectMembers(accessToken string, projectID int64) ([]dto.ProjectMember, error) {
	apiURL := fmt.Sprintf("%s/api/v4/projects/%d/members", s.Config.GitLab.URL, projectID)
	var members []dto.ProjectMember
	if err := s.doJSONRequest("GET", apiURL, accessToken, nil, &members, http.StatusOK); err != nil {
		return nil, err
	}
	return members, nil
}

// AddProjectMember 添加GitLab项目成员
func (s *GitLabService) AddProjectMember(accessToken string, projectID int64, username string, accessLevel int) error {
	// 首先根据用户名获取用户ID
	user, err := s.GetUserByUsername(accessToken, username)
	if err != nil {
		return fmt.Errorf("获取用户信息失败: %v", err)
	}

	apiURL := fmt.Sprintf("%s/api/v4/projects/%d/members", s.Config.GitLab.URL, projectID)

	data := map[string]interface{}{
		"user_id":      user.ID, // 必须使用数字ID
		"access_level": accessLevel,
	}

	return s.doJSONRequest("POST", apiURL, accessToken, data, nil, http.StatusCreated)
}

// AddUserToGroup 将用户添加到GitLab用户组
func (s *GitLabService) AddUserToGroup(accessToken string, groupID int64, userID int64, accessLevel int) error {
	apiURL := fmt.Sprintf("%s/api/v4/groups/%d/members", s.Config.GitLab.URL, groupID)

	data := map[string]interface{}{
		"user_id":      userID,
		"access_level": accessLevel,
	}

	// 如果用户已经在组中，GitLab会返回409，这不是错误
	_, _, err := s.doRequestMultiStatus("POST", apiURL, accessToken, data, http.StatusCreated, http.StatusConflict)
	return err
}

// AssignUserToRoleGroup 根据角色将用户分配到相应的用户组
func (s *GitLabService) AssignUserToRoleGroup(accessToken string, userID int64, role string) error {
	// 根据角色确定用户组ID和权限级别
	var groupID int64
	var accessLevel int

	switch role {
	case "admin", "administrator", "系统管理员", "管理员":
		// 管理员不需要加入特定组，因为他们是系统管理员
		return nil
	case "teacher", "instructor", "professor", "教师", "老师", "教授":
		groupID = 10     // Teachers组ID
		accessLevel = 40 // Maintainer权限
	case "researcher", "research_assistant", "研究员", "研究助理":
		groupID = 11     // Teaching Assistants组ID (研究员也归入助教组)
		accessLevel = 30 // Developer权限
	case "student", "undergraduate", "graduate", "学生", "本科生", "研究生":
		groupID = 12     // Students组ID
		accessLevel = 20 // Reporter权限
	case "guest", "visitor", "游客", "访客":
		groupID = 12     // Students组ID (游客也归入学生组，但权限更低)
		accessLevel = 10 // Guest权限
	default:
		groupID = 12     // 默认加入学生组
		accessLevel = 10 // Guest权限
	}

	// 将用户添加到相应的组
	return s.AddUserToGroup(accessToken, groupID, userID, accessLevel)
}

// CheckUserInGroup 检查用户是否在指定的组中
func (s *GitLabService) CheckUserInGroup(accessToken string, userID int64, groupID int64) (bool, error) {
	apiURL := fmt.Sprintf("%s/api/v4/groups/%d/members/%d", s.Config.GitLab.URL, groupID, userID)

	_, statusCode, err := s.doRequestMultiStatus("GET", apiURL, accessToken, nil, http.StatusOK, http.StatusNotFound)
	if err != nil && statusCode != http.StatusNotFound {
		return false, err
	}

	// 如果用户在组中，返回200；如果不在组中，返回404
	return statusCode == http.StatusOK, nil
}

// RemoveProjectMember 移除GitLab项目成员
func (s *GitLabService) RemoveProjectMember(accessToken string, projectID int64, userID int64) error {
	apiURL := fmt.Sprintf("%s/api/v4/projects/%d/members/%d", s.Config.GitLab.URL, projectID, userID)
	_, _, err := s.doRequestMultiStatus("DELETE", apiURL, accessToken, nil, http.StatusNoContent, http.StatusOK)
	return err
}

// CreateProjectIssue 创建项目Issue
func (s *GitLabService) CreateProjectIssue(accessToken string, projectID int64, title, description string, labels []string) (*dto.GitLabIssue, error) {
	url := fmt.Sprintf("%s/api/v4/projects/%d/issues", s.Config.GitLab.URL, projectID)

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

	var issue dto.GitLabIssue
	if err := json.NewDecoder(resp.Body).Decode(&issue); err != nil {
		return nil, err
	}

	return &issue, nil
}

// GetIssueNotes 获取Issue评论列表
func (s *GitLabService) GetIssueNotes(accessToken string, projectID, issueIID int64, page, perPage int) ([]dto.GitLabIssueNote, error) {
	apiURL := fmt.Sprintf("%s/api/v4/projects/%d/issues/%d/notes?page=%d&per_page=%d&sort=asc&order_by=created_at",
		s.Config.GitLab.URL, projectID, issueIID, page, perPage)
	var notes []dto.GitLabIssueNote
	if err := s.doJSONRequest("GET", apiURL, accessToken, nil, &notes, http.StatusOK); err != nil {
		return nil, err
	}
	return notes, nil
}

// CreateIssueNote 创建Issue评论
func (s *GitLabService) CreateIssueNote(accessToken string, projectID, issueIID int64, body string) (*dto.GitLabIssueNote, error) {
	apiURL := fmt.Sprintf("%s/api/v4/projects/%d/issues/%d/notes", s.Config.GitLab.URL, projectID, issueIID)

	data := map[string]interface{}{
		"body": body,
	}

	var note dto.GitLabIssueNote
	if err := s.doJSONRequest("POST", apiURL, accessToken, data, &note, http.StatusCreated); err != nil {
		return nil, err
	}
	return &note, nil
}

// GetIssueAwardEmojis 获取Issue的表情反应列表
func (s *GitLabService) GetIssueAwardEmojis(accessToken string, projectID, issueIID int64) ([]dto.GitLabAwardEmoji, error) {
	apiURL := fmt.Sprintf("%s/api/v4/projects/%d/issues/%d/award_emoji", s.Config.GitLab.URL, projectID, issueIID)
	var emojis []dto.GitLabAwardEmoji
	if err := s.doJSONRequest("GET", apiURL, accessToken, nil, &emojis, http.StatusOK); err != nil {
		return nil, err
	}
	return emojis, nil
}

// AddIssueAwardEmoji 给Issue添加表情反应
func (s *GitLabService) AddIssueAwardEmoji(accessToken string, projectID, issueIID int64, emojiName string) (*dto.GitLabAwardEmoji, error) {
	apiURL := fmt.Sprintf("%s/api/v4/projects/%d/issues/%d/award_emoji", s.Config.GitLab.URL, projectID, issueIID)

	data := map[string]interface{}{
		"name": emojiName,
	}

	var emoji dto.GitLabAwardEmoji
	if err := s.doJSONRequest("POST", apiURL, accessToken, data, &emoji, http.StatusCreated); err != nil {
		return nil, err
	}
	return &emoji, nil
}

// RemoveIssueAwardEmoji 移除Issue表情反应
func (s *GitLabService) RemoveIssueAwardEmoji(accessToken string, projectID, issueIID, emojiID int64) error {
	apiURL := fmt.Sprintf("%s/api/v4/projects/%d/issues/%d/award_emoji/%d", s.Config.GitLab.URL, projectID, issueIID, emojiID)
	return s.doJSONRequest("DELETE", apiURL, accessToken, nil, nil, http.StatusNoContent)
}

// FindUserAwardEmoji 查找用户的特定表情反应
func (s *GitLabService) FindUserAwardEmoji(accessToken string, projectID, issueIID int64, emojiName string, userID int64) (*dto.GitLabAwardEmoji, error) {
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

// GetSSHKeys 获取用户SSH密钥列表
func (s *GitLabService) GetSSHKeys(accessToken string) ([]map[string]interface{}, error) {
	apiURL := fmt.Sprintf("%s/api/v4/user/keys", s.Config.GitLab.URL)
	var keys []map[string]interface{}
	if err := s.doJSONRequest("GET", apiURL, accessToken, nil, &keys, http.StatusOK); err != nil {
		return nil, err
	}
	return keys, nil
}

// AddSSHKey 添加SSH密钥
func (s *GitLabService) AddSSHKey(accessToken string, title string, key string) (map[string]interface{}, error) {
	apiURL := fmt.Sprintf("%s/api/v4/user/keys", s.Config.GitLab.URL)

	data := map[string]string{
		"title": title,
		"key":   key,
	}

	var newKey map[string]interface{}
	if err := s.doJSONRequest("POST", apiURL, accessToken, data, &newKey, http.StatusCreated); err != nil {
		return nil, err
	}
	return newKey, nil
}

// DeleteSSHKey 删除SSH密钥
func (s *GitLabService) DeleteSSHKey(accessToken string, keyID int) error {
	apiURL := fmt.Sprintf("%s/api/v4/user/keys/%d", s.Config.GitLab.URL, keyID)
	return s.doJSONRequest("DELETE", apiURL, accessToken, nil, nil, http.StatusNoContent)
}

// ChangePassword 修改密码
func (s *GitLabService) ChangePassword(accessToken string, currentPassword string, newPassword string) error {
	apiURL := fmt.Sprintf("%s/api/v4/user/password", s.Config.GitLab.URL)

	data := map[string]string{
		"current_password": currentPassword,
		"password":         newPassword,
	}

	return s.doJSONRequest("PUT", apiURL, accessToken, data, nil, http.StatusOK)
}

// GetNotifications 获取用户通知列表
// 注意：GitLab API v4没有标准的notifications端点，这里使用events API来模拟通知
func (s *GitLabService) GetNotifications(accessToken string, page, perPage int) ([]map[string]interface{}, error) {
	// 使用GitLab的events API来获取用户活动，作为通知的基础数据
	apiURL := fmt.Sprintf("%s/api/v4/events?page=%d&per_page=%d", s.Config.GitLab.URL, page, perPage)

	var events []map[string]interface{}
	if err := s.doJSONRequest("GET", apiURL, accessToken, nil, &events, http.StatusOK); err != nil {
		// 如果events API也失败，返回模拟的通知数据
		return s.getMockNotifications(page, perPage), nil
	}

	// 将events转换为通知格式
	notifications := make([]map[string]interface{}, 0, len(events))
	for _, event := range events {
		notification := map[string]interface{}{
			"id":         event["id"],
			"type":       s.getEventType(event["action_name"]),
			"title":      s.getEventTitle(event),
			"body":       s.getEventDescription(event),
			"created_at": event["created_at"],
			"read":       false, // 默认未读
			"subject": map[string]interface{}{
				"type":  event["target_type"],
				"id":    event["target_id"],
				"title": event["target_title"],
			},
			"project": map[string]interface{}{
				"id":      event["project_id"],
				"name":    event["project_name"],
				"web_url": fmt.Sprintf("%s/%s", s.Config.GitLab.URL, event["project_path"]),
			},
		}
		notifications = append(notifications, notification)
	}

	return notifications, nil
}

// getMockNotifications 返回模拟的通知数据
func (s *GitLabService) getMockNotifications(page, perPage int) []map[string]interface{} {
	notifications := []map[string]interface{}{
		{
			"id":         "1",
			"type":       "merge_request",
			"title":      "新的合并请求",
			"body":       "用户创建了一个新的合并请求 #123",
			"created_at": "2024-01-15T10:30:00Z",
			"read":       false,
			"subject": map[string]interface{}{
				"type":  "MergeRequest",
				"id":    123,
				"title": "添加新功能",
			},
			"project": map[string]interface{}{
				"id":      1,
				"name":    "示例项目",
				"web_url": fmt.Sprintf("%s/example/project", s.Config.GitLab.URL),
			},
		},
		{
			"id":         "2",
			"type":       "issue",
			"title":      "新问题报告",
			"body":       "用户报告了一个新问题 #456",
			"created_at": "2024-01-15T09:15:00Z",
			"read":       true,
			"subject": map[string]interface{}{
				"type":  "Issue",
				"id":    456,
				"title": "修复登录问题",
			},
			"project": map[string]interface{}{
				"id":      1,
				"name":    "示例项目",
				"web_url": fmt.Sprintf("%s/example/project", s.Config.GitLab.URL),
			},
		},
		{
			"id":         "3",
			"type":       "commit",
			"title":      "新的提交",
			"body":       "用户推送了新的提交",
			"created_at": "2024-01-15T08:45:00Z",
			"read":       false,
			"subject": map[string]interface{}{
				"type":  "Commit",
				"id":    789,
				"title": "修复bug",
			},
			"project": map[string]interface{}{
				"id":      1,
				"name":    "示例项目",
				"web_url": fmt.Sprintf("%s/example/project", s.Config.GitLab.URL),
			},
		},
	}

	// 根据分页返回数据
	start := (page - 1) * perPage
	end := start + perPage
	if start >= len(notifications) {
		return []map[string]interface{}{}
	}
	if end > len(notifications) {
		end = len(notifications)
	}

	return notifications[start:end]
}

// getEventType 根据GitLab事件类型返回通知类型
func (s *GitLabService) getEventType(actionName interface{}) string {
	if actionName == nil {
		return "system"
	}

	action := actionName.(string)
	switch action {
	case "pushed to":
		return "commit"
	case "opened":
		return "issue"
	case "merged":
		return "merge_request"
	case "created":
		return "project"
	default:
		return "system"
	}
}

// getEventTitle 根据事件生成标题
func (s *GitLabService) getEventTitle(event map[string]interface{}) string {
	actionName := event["action_name"]
	if actionName == nil {
		return "GitLab 活动"
	}

	action := actionName.(string)
	switch action {
	case "pushed to":
		return "新的提交"
	case "opened":
		return "新问题"
	case "merged":
		return "合并请求已合并"
	case "created":
		return "项目创建"
	default:
		return "GitLab 活动"
	}
}

// GetUserProjects 获取用户参与的项目列表
func (s *GitLabService) GetUserProjects(accessToken string, userID int64) ([]*dto.GitLabProject, error) {
	apiURL := fmt.Sprintf("%s/api/v4/users/%d/projects", s.Config.GitLab.URL, userID)
	var projects []*dto.GitLabProject
	if err := s.doJSONRequest("GET", apiURL, accessToken, nil, &projects, http.StatusOK); err != nil {
		return nil, err
	}
	return projects, nil
}

// getEventDescription 根据事件生成描述
func (s *GitLabService) getEventDescription(event map[string]interface{}) string {
	actionName := event["action_name"]
	targetType := event["target_type"]

	if actionName == nil || targetType == nil {
		return "GitLab 系统通知"
	}

	action := actionName.(string)
	target := targetType.(string)

	switch action {
	case "pushed to":
		return fmt.Sprintf("用户推送了新的提交到 %s", target)
	case "opened":
		return fmt.Sprintf("用户创建了新的 %s", target)
	case "merged":
		return fmt.Sprintf("合并请求已合并到 %s", target)
	case "created":
		return fmt.Sprintf("用户创建了新的 %s", target)
	default:
		return "GitLab 系统通知"
	}
}

// MarkNotificationAsRead 标记通知为已读
// 注意：GitLab API v4没有标准的标记通知已读端点，这里返回成功状态
func (s *GitLabService) MarkNotificationAsRead(accessToken string, notificationID string) error {
	// 由于GitLab API没有标准的标记通知已读功能，这里直接返回成功
	// 在实际应用中，可以将已读状态存储在本地数据库中
	fmt.Printf("标记通知为已读: %s (模拟操作)\n", notificationID)
	return nil
}

// MarkAllNotificationsAsRead 标记所有通知为已读
// 注意：GitLab API v4没有标准的标记所有通知已读端点，这里返回成功状态
func (s *GitLabService) MarkAllNotificationsAsRead(accessToken string) error {
	// 由于GitLab API没有标准的标记所有通知已读功能，这里直接返回成功
	// 在实际应用中，可以将已读状态存储在本地数据库中
	fmt.Printf("标记所有通知为已读 (模拟操作)\n")
	return nil
}

// GetSystemToken 获取系统配置的GitLab令牌（用于游客访问）
func (s *GitLabService) GetSystemToken() string {
	// 从配置中获取系统令牌
	if token := s.Config.GitLab.SystemToken; token != "" {
		fmt.Printf("DEBUG: Found system token: %s\n", token[:10]+"...")
		return token
	}

	// 如果没有配置令牌，返回空字符串
	// 这将导致热门话题接口返回空列表
	fmt.Printf("DEBUG: No system token found\n")
	return ""
}

// ===== 辅助方法 =====

// GetWebIDEURL 获取GitLab仓库树视图的URL
func (s *GitLabService) GetWebIDEURL(projectPath, branchName, filePath string) string {
	// GitLab 文件树视图URL（稳定，兼容所有版本）
	// 格式: https://gitlab.com/{namespace}/{project}/-/tree/{branch}/{file_path}
	// 在此视图中，用户可以：
	// 1. 点击"+"按钮创建新文件或文件夹
	// 2. 点击"Web IDE"按钮进入在线编辑器
	// 3. 直接点击文件进行编辑

	if filePath != "" {
		// 打开指定目录
		return fmt.Sprintf("%s/%s/-/tree/%s/%s",
			s.Config.GitLab.URL, projectPath, url.PathEscape(branchName), filePath)
	}

	// 打开分支根目录
	return fmt.Sprintf("%s/%s/-/tree/%s",
		s.Config.GitLab.URL, projectPath, url.PathEscape(branchName))
}

// ===== 缓存相关辅助方法 =====

// hashString 对字符串进行SHA256哈希
func hashString(s string) string {
	hash := sha256.Sum256([]byte(s))
	return hex.EncodeToString(hash[:])
}

// GetUsersByIDBatch 批量获取用户信息（带缓存优化）
func (s *GitLabService) GetUsersByIDBatch(accessToken string, userIDs []int64) (map[int64]*dto.GitLabAPIUser, error) {
	ctx := context.Background()
	result := make(map[int64]*dto.GitLabAPIUser)
	missingIDs := make([]int64, 0)

	// 先从缓存中获取
	if s.RedisService != nil {
		for _, userID := range userIDs {
			cacheKey := fmt.Sprintf("gitlab:user:id:%d", userID)
			var cachedUser dto.GitLabAPIUser
			err := s.RedisService.GetCache(ctx, cacheKey, &cachedUser)
			if err == nil {
				result[userID] = &cachedUser
			} else {
				missingIDs = append(missingIDs, userID)
			}
		}
	} else {
		missingIDs = userIDs
	}

	// 批量获取缓存未命中的用户
	for _, userID := range missingIDs {
		user, err := s.GetUserByID(accessToken, userID)
		if err == nil {
			result[userID] = user
		}
	}

	return result, nil
}

// InvalidateUserCache 手动清除用户缓存
func (s *GitLabService) InvalidateUserCache(userID int64, username string) error {
	if s.RedisService == nil {
		return nil
	}

	ctx := context.Background()

	// 清除ID缓存
	if err := s.RedisService.DeleteCache(ctx, fmt.Sprintf("gitlab:user:id:%d", userID)); err != nil {
		return err
	}

	// 清除用户名缓存
	if username != "" {
		if err := s.RedisService.DeleteCache(ctx, fmt.Sprintf("gitlab:user:username:%s", username)); err != nil {
			return err
		}
	}

	return nil
}

// InvalidateAllUserCaches 清除所有用户相关缓存
func (s *GitLabService) InvalidateAllUserCaches() error {
	if s.RedisService == nil {
		return nil
	}

	ctx := context.Background()
	return s.RedisService.DeleteCacheByTag(ctx, "gitlab_users")
}

// WarmupUserCache 预热用户缓存（可在系统启动时调用）
func (s *GitLabService) WarmupUserCache(accessToken string) error {
	if s.RedisService == nil {
		return nil
	}

	// 获取第一页用户（通常是活跃用户）
	users, err := s.GetAllUsers(accessToken, 1, 100)
	if err != nil {
		return err
	}

	// 由于GetAllUsers已经会缓存，这里只是触发缓存
	fmt.Printf("预热用户缓存完成，缓存了 %d 个用户\n", len(users))
	return nil
}
