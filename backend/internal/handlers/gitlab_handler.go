package handlers

import (
	"gitlabex/internal/services"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

// GitLabHandler GitLab API处理器
type GitLabHandler struct {
	gitlabService *services.GitLabService
	userService   *services.UserService
}

// NewGitLabHandler 创建GitLab处理器
func NewGitLabHandler(gitlabService *services.GitLabService, userService *services.UserService) *GitLabHandler {
	return &GitLabHandler{
		gitlabService: gitlabService,
		userService:   userService,
	}
}

// GetCurrentUser 获取当前用户信息
func (h *GitLabHandler) GetCurrentUser(c *gin.Context) {
	accessToken := c.GetHeader("X-GitLab-Token")
	if accessToken == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "GitLab access token required"})
		return
	}

	user, err := h.gitlabService.GetUser(accessToken)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"user": user})
}

// GetProjects 获取用户的项目列表
func (h *GitLabHandler) GetProjects(c *gin.Context) {
	accessToken := c.GetHeader("X-GitLab-Token")
	if accessToken == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "GitLab access token required"})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "20"))

	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}

	projects, err := h.gitlabService.GetProjects(accessToken, page, perPage)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"projects": projects})
}

// GetProject 获取特定项目信息
func (h *GitLabHandler) GetProject(c *gin.Context) {
	accessToken := c.GetHeader("X-GitLab-Token")
	if accessToken == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "GitLab access token required"})
		return
	}

	projectID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	project, err := h.gitlabService.GetProject(accessToken, projectID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"project": project})
}

// CreateProject 创建新项目
func (h *GitLabHandler) CreateProject(c *gin.Context) {
	accessToken := c.GetHeader("X-GitLab-Token")
	if accessToken == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "GitLab access token required"})
		return
	}

	var req services.CreateProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Project name is required"})
		return
	}

	project, err := h.gitlabService.CreateProject(accessToken, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"project": project})
}

// GetBranches 获取项目分支列表
func (h *GitLabHandler) GetBranches(c *gin.Context) {
	accessToken := c.GetHeader("X-GitLab-Token")
	if accessToken == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "GitLab access token required"})
		return
	}

	projectID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	branches, err := h.gitlabService.GetBranches(accessToken, projectID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"branches": branches})
}

// CreateBranch 创建新分支
func (h *GitLabHandler) CreateBranch(c *gin.Context) {
	accessToken := c.GetHeader("X-GitLab-Token")
	if accessToken == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "GitLab access token required"})
		return
	}

	projectID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	var req services.CreateBranchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Branch == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Branch name is required"})
		return
	}

	branch, err := h.gitlabService.CreateBranch(accessToken, projectID, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"branch": branch})
}

// GetFileContent 获取文件内容
func (h *GitLabHandler) GetFileContent(c *gin.Context) {
	accessToken := c.GetHeader("X-GitLab-Token")
	if accessToken == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "GitLab access token required"})
		return
	}

	projectID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	filePath := c.Param("path")
	ref := c.DefaultQuery("ref", "main")

	file, err := h.gitlabService.GetFileContent(accessToken, projectID, filePath, ref)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"file": file})
}

// CreateFile 创建新文件
func (h *GitLabHandler) CreateFile(c *gin.Context) {
	accessToken := c.GetHeader("X-GitLab-Token")
	if accessToken == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "GitLab access token required"})
		return
	}

	projectID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	filePath := c.Param("path")
	var req services.CreateFileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.CommitMessage == "" {
		req.CommitMessage = "Create " + filePath
	}

	err = h.gitlabService.CreateFile(accessToken, projectID, filePath, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "File created successfully"})
}

// UpdateFile 更新文件内容
func (h *GitLabHandler) UpdateFile(c *gin.Context) {
	accessToken := c.GetHeader("X-GitLab-Token")
	if accessToken == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "GitLab access token required"})
		return
	}

	projectID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	filePath := c.Param("path")
	var req services.CreateFileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.CommitMessage == "" {
		req.CommitMessage = "Update " + filePath
	}

	err = h.gitlabService.UpdateFile(accessToken, projectID, filePath, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "File updated successfully"})
}

// GetCommits 获取提交历史
func (h *GitLabHandler) GetCommits(c *gin.Context) {
	accessToken := c.GetHeader("X-GitLab-Token")
	if accessToken == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "GitLab access token required"})
		return
	}

	projectID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	branch := c.DefaultQuery("branch", "main")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if limit < 1 || limit > 100 {
		limit = 20
	}

	commits, err := h.gitlabService.GetCommits(accessToken, projectID, branch, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"commits": commits})
}

// GetIssues 获取项目议题
func (h *GitLabHandler) GetIssues(c *gin.Context) {
	accessToken := c.GetHeader("X-GitLab-Token")
	if accessToken == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "GitLab access token required"})
		return
	}

	projectID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	state := c.DefaultQuery("state", "opened")
	var labels []string
	if labelStr := c.Query("labels"); labelStr != "" {
		labels = strings.Split(labelStr, ",")
	}

	issues, err := h.gitlabService.GetIssues(accessToken, projectID, state, labels)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"issues": issues})
}

// CreateIssue 创建新议题
func (h *GitLabHandler) CreateIssue(c *gin.Context) {
	accessToken := c.GetHeader("X-GitLab-Token")
	if accessToken == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "GitLab access token required"})
		return
	}

	projectID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	var req struct {
		Title       string   `json:"title" binding:"required"`
		Description string   `json:"description"`
		Labels      []string `json:"labels"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	issue, err := h.gitlabService.CreateIssue(accessToken, projectID, req.Title, req.Description, req.Labels, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"issue": issue})
}

// GetMergeRequests 获取合并请求
func (h *GitLabHandler) GetMergeRequests(c *gin.Context) {
	accessToken := c.GetHeader("X-GitLab-Token")
	if accessToken == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "GitLab access token required"})
		return
	}

	projectID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	state := c.DefaultQuery("state", "opened")

	mrs, err := h.gitlabService.GetMergeRequests(accessToken, projectID, state)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"merge_requests": mrs})
}

// GetRepositoryTree 获取仓库文件树
func (h *GitLabHandler) GetRepositoryTree(c *gin.Context) {
	accessToken := c.GetHeader("X-GitLab-Token")
	if accessToken == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "GitLab access token required"})
		return
	}

	projectID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	path := c.DefaultQuery("path", "")
	ref := c.DefaultQuery("ref", "main")

	tree, err := h.gitlabService.GetRepositoryTree(accessToken, projectID, path, ref)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"tree": tree})
}

// SearchFiles 搜索文件
func (h *GitLabHandler) SearchFiles(c *gin.Context) {
	accessToken := c.GetHeader("X-GitLab-Token")
	if accessToken == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "GitLab access token required"})
		return
	}

	projectID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	search := c.Query("search")
	if search == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Search query is required"})
		return
	}

	results, err := h.gitlabService.SearchFiles(accessToken, projectID, search)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"results": results})
}

// ValidateRepositoryAccess 验证仓库访问权限
func (h *GitLabHandler) ValidateRepositoryAccess(c *gin.Context) {
	accessToken := c.GetHeader("X-GitLab-Token")
	if accessToken == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "GitLab access token required"})
		return
	}

	projectID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	err = h.gitlabService.ValidateRepositoryAccess(accessToken, projectID)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "Repository access denied"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"valid": true})
}
