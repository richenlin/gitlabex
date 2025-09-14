package handlers

import (
	"fmt"
	"gitlabex/internal/models"
	"gitlabex/internal/services"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// ResearchHandler 研究课题处理器
type ResearchHandler struct {
	researchService *services.ResearchService
	userService     *services.UserService
	gitlabService   *services.GitLabService
}

// NewResearchHandler 创建研究课题处理器
func NewResearchHandler(researchService *services.ResearchService, userService *services.UserService, gitlabService *services.GitLabService) *ResearchHandler {
	return &ResearchHandler{
		researchService: researchService,
		userService:     userService,
		gitlabService:   gitlabService,
	}
}

// generateProjectPath 生成符合GitLab要求的项目路径
// GitLab路径要求：只能包含字母、数字、'_', '-' 和 '.'，不能以'-'开头，不能以'.git'或'.atom'结尾
func generateProjectPath(name string) string {
	// 转换为小写
	path := strings.ToLower(name)

	// 替换空格和其他特殊字符为连字符
	reg := regexp.MustCompile(`[^a-z0-9._-]`)
	path = reg.ReplaceAllString(path, "-")

	// 移除连续的特殊字符
	reg = regexp.MustCompile(`[-._]{2,}`)
	path = reg.ReplaceAllString(path, "-")

	// 确保不以'-'开头
	path = strings.TrimPrefix(path, "-")

	// 确保不以'.git'或'.atom'结尾
	path = strings.TrimSuffix(path, ".git")
	path = strings.TrimSuffix(path, ".atom")

	// 确保不以特殊字符开头或结尾
	path = strings.Trim(path, "-._")

	// 如果路径为空或过短，使用默认值
	if len(path) < 2 {
		path = "research-project"
	}

	return path
}

// hasProjectCreationPermission 检查用户是否有创建项目的权限
func (h *ResearchHandler) hasProjectCreationPermission(accessToken string) bool {
	// 通过GitLab API检查用户权限
	// 检查用户是否可以获取自己的用户信息，如果可以获取说明token有效且用户有基本权限
	user, err := h.gitlabService.GetUser(accessToken)
	if err != nil {
		return false
	}

	// 检查用户是否有效
	// 如果能够成功获取用户信息，说明token有效且用户有基本权限
	// 在GitLab中，正常用户都应该能够在自己的命名空间下创建项目
	return user.ID > 0 // 有效的用户ID表示用户有基本权限
}

// GetResearchProjects 获取研究课题列表
func (h *ResearchHandler) GetResearchProjects(c *gin.Context) {
	// 检查是否为游客模式
	isGuest, _ := c.Get("is_guest")
	gitlabUserID, _ := c.Get("gitlab_user_id")

	// 获取分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	isPublic := c.Query("is_public") == "true"
	includePrivate := c.Query("include_private") != "false"
	ownerIDStr := c.Query("ownerId")

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	offset := (page - 1) * limit

	var projects []models.ResearchProject
	var total int64
	var err error

	// 如果是游客模式，只返回公开项目
	if isGuest == true || gitlabUserID == nil {
		projects, total, err = h.researchService.GetAllProjects(limit, offset, true, false)
	} else {
		gitlabUID := gitlabUserID.(int64)

		// 如果指定了ownerId参数，则按创建者过滤
		if ownerIDStr != "" {
			ownerGitLabID, parseErr := strconv.ParseInt(ownerIDStr, 10, 64)
			if parseErr != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "无效的创建者ID"})
				return
			}

			// 根据创建者GitLab ID获取项目
			projects, total, err = h.researchService.GetUserProjectsByGitLabID(ownerGitLabID, limit, offset)
		} else {
			// 检查用户权限 - 管理员可以看到所有项目，普通用户看到公开项目和自己创建的项目
			isAdmin, _ := c.Get("is_admin")
			if isAdmin != nil && isAdmin.(bool) {
				// 管理员可以看到所有项目
				projects, total, err = h.researchService.GetAllProjects(limit, offset, isPublic, includePrivate)
			} else {
				// 普通用户只能看到公开项目和自己创建的项目
				projects, total, err = h.researchService.GetUserAccessibleProjectsByGitLabID(gitlabUID, limit, offset)
			}
		}
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取课题失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"projects": projects,
		"pagination": gin.H{
			"page":  page,
			"limit": limit,
			"total": total,
			"pages": (total + int64(limit) - 1) / int64(limit),
		},
	})
}

// CreateResearchProject 创建研究课题
func (h *ResearchHandler) CreateResearchProject(c *gin.Context) {
	// 从上下文获取GitLab访问令牌和用户信息
	accessToken, exists := c.Get("gitlab_access_token")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户未登录"})
		return
	}

	gitlabUserID, exists := c.Get("gitlab_user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "无法获取用户信息"})
		return
	}

	var req struct {
		Name        string     `json:"name" binding:"required"`
		Title       string     `json:"title"`
		Description string     `json:"description"`
		GitLabURL   string     `json:"gitlab_url"`
		IsPublic    bool       `json:"is_public"`
		Visibility  string     `json:"visibility"`
		StartDate   *time.Time `json:"start_date"`
		EndDate     *time.Time `json:"end_date"`
		Tags        []string   `json:"tags"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 检查权限 - 根据GitLab权限机制进行权限控制
	// 1. 系统管理员可以创建项目
	// 2. 普通用户需要有足够的GitLab权限才能创建项目
	isAdmin, _ := c.Get("is_admin")

	if isAdmin != nil && isAdmin.(bool) {
		// 系统管理员，允许创建
	} else {
		// 普通用户，检查GitLab权限
		// 获取用户的GitLab信息来验证是否有创建项目的权限
		accessToken := c.MustGet("gitlab_access_token").(string)

		// 检查用户是否有创建项目的权限（通过尝试获取用户的命名空间信息）
		if !h.hasProjectCreationPermission(accessToken) {
			c.JSON(http.StatusForbidden, gin.H{
				"error":   "权限不足：您没有创建研究项目的权限",
				"message": "请联系管理员或确保您在GitLab中有足够的权限",
			})
			return
		}
	}

	// 处理项目名称和可见性
	projectName := req.Name
	if req.Title != "" {
		projectName = req.Title
	}

	// 处理可见性设置
	isPublic := req.IsPublic
	if req.Visibility != "" {
		isPublic = req.Visibility == "public"
	}

	// 创建GitLab项目
	visibility := "private"
	if isPublic {
		visibility = "public"
	}

	// 生成项目路径（GitLab要求的path字段）
	projectPath := generateProjectPath(projectName)

	createReq := &services.CreateProjectRequest{
		Name:                 projectName,
		Path:                 projectPath,
		Description:          req.Description,
		Visibility:           visibility,
		InitializeWithReadme: true,
		DefaultBranch:        "main",
	}

	gitlabProject, err := h.gitlabService.CreateProject(accessToken.(string), createReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建GitLab项目失败: " + err.Error()})
		return
	}

	// 设置分支保护，防止学生随意修改主分支
	if err := h.gitlabService.SetupProjectBranchProtection(accessToken.(string), gitlabProject.ID); err != nil {
		// 分支保护失败不应该阻止项目创建，只记录错误
		fmt.Printf("警告：为项目 %d 设置分支保护失败: %v\n", gitlabProject.ID, err)
	}

	// 处理开始日期，如果没有提供则使用当前时间
	startDate := time.Now()
	if req.StartDate != nil {
		startDate = *req.StartDate
	}

	project := &models.ResearchProject{
		Name:            projectName,
		Description:     req.Description,
		GitLabURL:       gitlabProject.WebURL,
		GitLabProjectID: &gitlabProject.ID,
		CreatorID:       gitlabUserID.(int64),
		IsPublic:        isPublic,
		StartDate:       startDate,
		EndDate:         req.EndDate,
		Status:          "active",
	}

	if err := h.researchService.CreateResearchProject(project); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建课题失败"})
		return
	}

	// 注意：项目创建者（CreatorID）本身就是Owner，不需要添加到ProjectMember表中
	// ProjectMember表只用于存储由创建者添加的其他成员

	c.JSON(http.StatusCreated, project)
}

// GetResearchProjectByID 根据ID获取研究课题详情
func (h *ResearchHandler) GetResearchProjectByID(c *gin.Context) {
	// 从上下文获取GitLab用户信息
	gitlabUserID, exists := c.Get("gitlab_user_id")
	accessToken, hasToken := c.Get("gitlab_access_token")

	// 允许游客查看公开项目
	isGuest := !exists || !hasToken

	projectIDStr := c.Param("id")
	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的课题ID"})
		return
	}

	project, err := h.researchService.GetResearchProjectByID(projectID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "课题不存在"})
		return
	}

	// 检查访问权限
	if !isGuest {
		gitlabUID := gitlabUserID.(int64)

		// 使用GitLab权限检查（如果项目关联了GitLab）
		if project.GitLabProjectID != nil {
			// 对于关联GitLab的项目，使用GitLab权限检查
			hasPermission := false

			// 检查是否是管理员
			isAdmin, _ := c.Get("is_admin")
			if isAdmin != nil && isAdmin.(bool) {
				hasPermission = true
			} else if project.IsPublic {
				// 公开项目所有人都可以查看
				hasPermission = true
			} else if project.CreatorID == gitlabUID {
				// 项目创建者可以访问
				hasPermission = true
			} else {
				// 使用GitLab API检查项目权限
				err := h.gitlabService.ValidateRepositoryAccess(accessToken.(string), *project.GitLabProjectID)
				hasPermission = (err == nil)
			}

			if !hasPermission {
				c.JSON(http.StatusForbidden, gin.H{"error": "无权限访问该课题"})
				return
			}
		} else {
			// 对于未关联GitLab的项目，使用简化权限检查
			isAdmin, _ := c.Get("is_admin")
			if !(isAdmin != nil && isAdmin.(bool)) && !project.IsPublic && project.CreatorID != gitlabUID {
				c.JSON(http.StatusForbidden, gin.H{"error": "无权限访问该课题"})
				return
			}
		}
	} else {
		// 游客只能查看公开项目
		if !project.IsPublic {
			c.JSON(http.StatusForbidden, gin.H{"error": "无权限访问该课题"})
			return
		}
	}

	c.JSON(http.StatusOK, project)
}

// UpdateResearchProject 更新研究课题信息
func (h *ResearchHandler) UpdateResearchProject(c *gin.Context) {
	// 从上下文获取GitLab用户信息
	gitlabUserID, exists := c.Get("gitlab_user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户未登录"})
		return
	}

	projectIDStr := c.Param("id")
	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的课题ID"})
		return
	}

	var req struct {
		Title       *string    `json:"title"`
		Description *string    `json:"description"`
		IsPublic    *bool      `json:"is_public"`
		StartDate   *time.Time `json:"start_date"`
		EndDate     *time.Time `json:"end_date"`
		Tags        []string   `json:"tags"`
		Status      *string    `json:"status"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 获取项目信息以检查权限
	project, err := h.researchService.GetResearchProjectByID(projectID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "课题不存在"})
		return
	}

	// 检查权限 - 管理员、项目创建者可以更新
	// 注意：为了简化权限管理，我们暂时只允许项目创建者和管理员编辑项目
	// 教师权限可以通过GitLab项目成员管理来实现
	isAdmin, _ := c.Get("is_admin")
	isOwner := project.CreatorID == gitlabUserID.(int64)

	// 权限检查：项目创建者或管理员可以更新
	if !isOwner && (isAdmin == nil || !isAdmin.(bool)) {
		c.JSON(http.StatusForbidden, gin.H{"error": "无权限修改该课题"})
		return
	}

	updates := make(map[string]interface{})
	if req.Title != nil {
		updates["name"] = *req.Title
	}
	if req.Description != nil {
		updates["description"] = *req.Description
	}
	if req.IsPublic != nil {
		updates["is_public"] = *req.IsPublic
	}
	if req.StartDate != nil {
		updates["start_date"] = *req.StartDate
	}
	if req.EndDate != nil {
		updates["end_date"] = *req.EndDate
	}
	if req.Tags != nil {
		updates["tags"] = req.Tags
	}
	if req.Status != nil {
		updates["status"] = *req.Status
	}

	if err := h.researchService.UpdateResearchProject(projectID, updates); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新课题失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "课题更新成功"})
}

// DeleteResearchProject 删除研究课题
func (h *ResearchHandler) DeleteResearchProject(c *gin.Context) {
	gitlabUserID, exists := c.Get("gitlab_user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户未登录"})
		return
	}

	projectIDStr := c.Param("id")
	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的课题ID"})
		return
	}

	// 检查权限
	isOwner, err := h.researchService.IsProjectOwnerByGitLabID(projectID, gitlabUserID.(int64))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "检查权限失败"})
		return
	}

	// 检查权限：项目创建者或管理员可以删除
	isAdmin, _ := c.Get("is_admin")
	if !isOwner && (isAdmin == nil || !isAdmin.(bool)) {
		c.JSON(http.StatusForbidden, gin.H{"error": "无权限删除该课题"})
		return
	}

	// 获取项目信息
	project, err := h.researchService.GetResearchProjectByID(projectID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "课题不存在"})
		return
	}

	// 删除GitLab项目
	if project.GitLabProjectID != nil {
		accessToken, exists := c.Get("gitlab_access_token")
		if exists {
			if err := h.gitlabService.DeleteProject(accessToken.(string), *project.GitLabProjectID); err != nil {
				// 记录错误但不阻止删除，因为数据库中的项目记录仍需要删除
				fmt.Printf("警告: 删除GitLab项目失败: %v\n", err)
			}
		}
	}

	if err := h.researchService.DeleteResearchProject(projectID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除课题失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "课题删除成功"})
}

// GetMembers 获取课题成员列表
func (h *ResearchHandler) GetMembers(c *gin.Context) {
	projectIDStr := c.Param("id")
	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的课题ID"})
		return
	}

	// 获取项目信息
	project, err := h.researchService.GetResearchProjectByID(projectID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "课题不存在"})
		return
	}

	// 如果没有关联GitLab项目，返回空成员列表
	if project.GitLabProjectID == nil {
		c.JSON(http.StatusOK, gin.H{
			"message": "该课题未关联GitLab项目",
			"members": []interface{}{},
		})
		return
	}

	// 获取GitLab访问令牌
	accessToken, exists := c.Get("gitlab_access_token")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "缺少GitLab访问令牌"})
		return
	}

	// 通过GitLab API获取项目成员
	members, err := h.gitlabService.GetProjectMembers(accessToken.(string), *project.GitLabProjectID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "获取项目成员失败",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "获取成员列表成功",
		"members": members,
		"count":   len(members),
	})
}

// AddMember 添加课题成员
func (h *ResearchHandler) AddMember(c *gin.Context) {
	gitlabUserID, exists := c.Get("gitlab_user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户未登录"})
		return
	}

	projectIDStr := c.Param("id")
	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的课题ID"})
		return
	}

	var req struct {
		Username    string `json:"username" binding:"required"`     // GitLab用户名
		AccessLevel int    `json:"access_level" binding:"required"` // GitLab访问级别
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 检查权限 - 只有项目所有者和管理员可以添加成员
	isOwner, err := h.researchService.IsProjectOwnerByGitLabID(projectID, gitlabUserID.(int64))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "检查权限失败"})
		return
	}

	// 检查权限：项目创建者或管理员可以操作
	isAdmin, _ := c.Get("is_admin")
	if !isOwner && (isAdmin == nil || !isAdmin.(bool)) {
		c.JSON(http.StatusForbidden, gin.H{"error": "无权限添加成员"})
		return
	}

	// 获取项目信息
	project, err := h.researchService.GetResearchProjectByID(projectID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "课题不存在"})
		return
	}

	// 检查是否关联了GitLab项目
	if project.GitLabProjectID == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "该课题未关联GitLab项目"})
		return
	}

	// 获取GitLab访问令牌
	accessToken, exists := c.Get("gitlab_access_token")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "缺少GitLab访问令牌"})
		return
	}

	// 通过GitLab API添加项目成员
	err = h.gitlabService.AddProjectMember(accessToken.(string), *project.GitLabProjectID, req.Username, req.AccessLevel)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "添加项目成员失败",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "成员添加成功"})
}

// RemoveMember 移除课题成员
func (h *ResearchHandler) RemoveMember(c *gin.Context) {
	gitlabUserID, exists := c.Get("gitlab_user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户未登录"})
		return
	}

	projectIDStr := c.Param("id")
	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的课题ID"})
		return
	}

	// 获取要移除的用户ID
	userIDStr := c.Param("userId")
	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的用户ID"})
		return
	}

	// 检查权限
	isOwner, err := h.researchService.IsProjectOwnerByGitLabID(projectID, gitlabUserID.(int64))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "检查权限失败"})
		return
	}

	// 检查权限：项目创建者或管理员可以操作
	isAdmin, _ := c.Get("is_admin")
	if !isOwner && (isAdmin == nil || !isAdmin.(bool)) {
		c.JSON(http.StatusForbidden, gin.H{"error": "无权限移除成员"})
		return
	}

	// 获取项目信息
	project, err := h.researchService.GetResearchProjectByID(projectID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "课题不存在"})
		return
	}

	// 检查是否关联了GitLab项目
	if project.GitLabProjectID == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "该课题未关联GitLab项目"})
		return
	}

	// 获取GitLab访问令牌
	accessToken, exists := c.Get("gitlab_access_token")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "缺少GitLab访问令牌"})
		return
	}

	// 通过GitLab API移除项目成员
	err = h.gitlabService.RemoveProjectMember(accessToken.(string), *project.GitLabProjectID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "移除项目成员失败",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "成员移除成功"})
}

// GetIssues 获取课题相关Issues
func (h *ResearchHandler) GetIssues(c *gin.Context) {
	projectIDStr := c.Param("id")
	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的课题ID"})
		return
	}

	// 获取课题信息
	project, err := h.researchService.GetResearchProjectByID(projectID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "课题不存在"})
		return
	}

	// 检查GitLab项目ID是否存在
	if project.GitLabProjectID == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "项目未关联GitLab"})
		return
	}

	// 获取访问令牌
	accessToken, exists := c.Get("gitlab_access_token")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "缺少GitLab访问令牌"})
		return
	}

	// 通过GitLab API获取项目Issues（获取第一页，每页20个）
	issues, err := h.gitlabService.GetProjectIssues(accessToken.(string), *project.GitLabProjectID, 1, 20)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取Issues失败", "details": err.Error()})
		return
	}

	// 转换为map格式以便前端处理
	issuesList := make([]map[string]interface{}, len(issues))
	for i, issue := range issues {
		issuesList[i] = map[string]interface{}{
			"id":          issue.ID,
			"iid":         issue.IID,
			"title":       issue.Title,
			"description": issue.Description,
			"state":       issue.State,
			"created_at":  issue.CreatedAt,
			"updated_at":  issue.UpdatedAt,
			"author": map[string]interface{}{
				"id":       issue.Author.ID,
				"name":     issue.Author.Name,
				"username": issue.Author.Username,
			},
		}
	}

	c.JSON(http.StatusOK, gin.H{"issues": issuesList})
}

// CreateIssue 创建新Issue
func (h *ResearchHandler) CreateIssue(c *gin.Context) {
	_, exists := c.Get("gitlab_user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户未登录"})
		return
	}

	projectIDStr := c.Param("id")
	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的课题ID"})
		return
	}

	var req struct {
		Title       string   `json:"title" binding:"required"`
		Description string   `json:"description"`
		Labels      []string `json:"labels"`
		AssigneeID  *int64   `json:"assignee_id"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 获取课题信息以检查GitLab关联
	project, err := h.researchService.GetResearchProjectByID(projectID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "课题不存在"})
		return
	}

	// 检查项目是否关联GitLab
	if project.GitLabProjectID == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "课题未关联GitLab项目"})
		return
	}

	// 获取GitLab访问令牌
	accessToken, exists := c.Get("gitlab_access_token")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "缺少GitLab访问令牌"})
		return
	}

	// 权限检查将由GitLab API在创建Issue时进行
	// 如果用户没有权限，GitLab API会返回错误

	// 创建GitLab Issue
	issue, err := h.gitlabService.CreateIssue(accessToken.(string), *project.GitLabProjectID, req.Title, req.Description, req.Labels, req.AssigneeID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建Issue失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, issue)
}

// GetIssue 获取单个Issue详情
func (h *ResearchHandler) GetIssue(c *gin.Context) {
	projectIDStr := c.Param("id")
	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的课题ID"})
		return
	}

	issueIDStr := c.Param("issueId")
	issueID, err := strconv.ParseInt(issueIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的Issue ID"})
		return
	}

	// 获取课题信息
	project, err := h.researchService.GetResearchProjectByID(projectID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "课题不存在"})
		return
	}

	// 检查项目是否关联GitLab
	if project.GitLabProjectID == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "课题未关联GitLab项目"})
		return
	}

	// 获取访问令牌
	accessToken, exists := c.Get("gitlab_access_token")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "缺少GitLab访问令牌"})
		return
	}

	// 通过GitLab API获取Issue详情
	issue, err := h.gitlabService.GetIssue(accessToken.(string), *project.GitLabProjectID, issueID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Issue不存在", "details": err.Error()})
		return
	}

	// 转换为map格式以便前端处理
	issueData := map[string]interface{}{
		"id":          issue.ID,
		"iid":         issue.IID,
		"title":       issue.Title,
		"description": issue.Description,
		"state":       issue.State,
		"created_at":  issue.CreatedAt,
		"updated_at":  issue.UpdatedAt,
		"author": map[string]interface{}{
			"id":       issue.Author.ID,
			"name":     issue.Author.Name,
			"username": issue.Author.Username,
		},
	}

	c.JSON(http.StatusOK, issueData)
}

// GetDiscussions 获取Issue的讨论
func (h *ResearchHandler) GetDiscussions(c *gin.Context) {
	projectIDStr := c.Param("id")
	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的课题ID"})
		return
	}

	issueIDStr := c.Param("issueId")
	issueID, err := strconv.ParseInt(issueIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的Issue ID"})
		return
	}

	// 获取课题信息
	project, err := h.researchService.GetResearchProjectByID(projectID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "课题不存在"})
		return
	}

	// 检查项目是否关联GitLab
	if project.GitLabProjectID == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "课题未关联GitLab项目"})
		return
	}

	// TODO: 实现GetIssueDiscussions方法
	// discussions, err := h.gitlabService.GetIssueDiscussions(*project.GitLabProjectID, issueID)
	_ = issueID // 临时忽略未使用警告
	// if err != nil {
	//     c.JSON(http.StatusInternalServerError, gin.H{"error": "获取讨论失败"})
	//     return
	// }

	// 临时返回空讨论列表
	c.JSON(http.StatusOK, gin.H{"discussions": []interface{}{}})
}

// CreateDiscussion 创建新讨论
func (h *ResearchHandler) CreateDiscussion(c *gin.Context) {
	_, exists := c.Get("gitlab_user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户未登录"})
		return
	}

	projectIDStr := c.Param("id")
	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的课题ID"})
		return
	}

	issueIDStr := c.Param("issueId")
	issueID, err := strconv.ParseInt(issueIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的Issue ID"})
		return
	}

	var req struct {
		Body string `json:"body" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 注意：权限检查已简化，具体权限由GitLab控制

	// 获取课题信息
	project, err := h.researchService.GetResearchProjectByID(projectID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "课题不存在"})
		return
	}

	// 检查项目是否关联GitLab
	if project.GitLabProjectID == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "课题未关联GitLab项目"})
		return
	}

	// TODO: 实现CreateIssueDiscussion方法
	// discussion, err := h.gitlabService.CreateIssueDiscussion(*project.GitLabProjectID, issueID, req.Body)
	_ = issueID // 临时忽略未使用警告
	// if err != nil {
	//     c.JSON(http.StatusInternalServerError, gin.H{"error": "创建讨论失败"})
	//     return
	// }

	// 临时返回成功响应
	c.JSON(http.StatusCreated, gin.H{"message": "讨论创建成功（模拟）"})
}

// GetHomework 获取课题相关作业
func (h *ResearchHandler) GetHomework(c *gin.Context) {
	projectIDStr := c.Param("id")
	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的课题ID"})
		return
	}

	homeworks, err := h.researchService.GetProjectHomework(projectID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取作业失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"homeworks": homeworks})
}

// CreateHomework 创建课题作业
func (h *ResearchHandler) CreateHomework(c *gin.Context) {
	gitlabUserID, exists := c.Get("gitlab_user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户未登录"})
		return
	}

	projectIDStr := c.Param("id")
	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的课题ID"})
		return
	}

	var req struct {
		Title       string     `json:"title" binding:"required"`
		Description string     `json:"description"`
		DueDate     *time.Time `json:"due_date"`
		MaxGrade    int        `json:"max_grade"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 检查权限
	isOwner, err := h.researchService.IsProjectOwnerByGitLabID(projectID, gitlabUserID.(int64))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "检查权限失败"})
		return
	}

	// 用户信息已从GitLab获取，无需本地查询

	// 检查权限：项目创建者或管理员可以创建作业
	isAdmin, _ := c.Get("is_admin")
	if !isOwner && (isAdmin == nil || !isAdmin.(bool)) {
		c.JSON(http.StatusForbidden, gin.H{"error": "无权限创建作业"})
		return
	}

	homework := &models.Homework{
		ProjectID:   projectID,
		Title:       req.Title,
		Description: req.Description,
		DueDate:     req.DueDate,
		MaxGrade:    req.MaxGrade,
		CreatorID:   gitlabUserID.(int64),
		Status:      "active",
	}

	if err := h.researchService.CreateProjectHomework(homework); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建作业失败"})
		return
	}

	c.JSON(http.StatusCreated, homework)
}

// GetHotProjects 获取热门项目
func (h *ResearchHandler) GetHotProjects(c *gin.Context) {
	// 获取限制参数
	limitStr := c.DefaultQuery("limit", "10")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 || limit > 50 {
		limit = 10
	}

	// 获取热门项目
	projects, err := h.researchService.GetHotProjects(limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取热门项目失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"projects": projects,
		"count":    len(projects),
	})
}

// GetGitLabIDEURL 获取GitLab IDE URL
func (h *ResearchHandler) GetGitLabIDEURL(c *gin.Context) {
	projectID := c.Param("id")
	filePath := c.Query("file")
	branch := c.DefaultQuery("branch", "main")

	if filePath == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "文件路径不能为空"})
		return
	}

	// 获取项目信息
	var project models.ResearchProject
	if err := h.researchService.DB.First(&project, projectID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "项目不存在"})
		return
	}

	if project.GitLabProjectID == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "项目未关联GitLab项目"})
		return
	}

	// 构建GitLab IDE URL
	gitlabURL := h.researchService.Config.GitLabURL
	ideURL := fmt.Sprintf("%s/-/ide/project/root/%d/edit/%s/-/%s",
		gitlabURL, *project.GitLabProjectID, branch, filePath)

	c.JSON(http.StatusOK, gin.H{
		"ide_url":    ideURL,
		"gitlab_url": gitlabURL,
		"project_id": *project.GitLabProjectID,
		"file_path":  filePath,
		"branch":     branch,
	})
}
