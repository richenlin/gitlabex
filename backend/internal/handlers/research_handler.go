package handlers

import (
	"gitlabex/internal/models"
	"gitlabex/internal/services"
	"net/http"
	"strconv"
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

	// 检查权限 - 暂时允许所有登录用户创建项目
	// TODO: 可以根据GitLab用户信息或项目权限进行更细粒度的权限控制
	// isAdmin, exists := c.Get("is_admin")
	// if !exists || !isAdmin.(bool) {
	//     c.JSON(http.StatusForbidden, gin.H{"error": "无权限创建课题"})
	//     return
	// }

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

	createReq := &services.CreateProjectRequest{
		Name:                 projectName,
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

	// 检查权限 - 只有创建者和管理员可以更新
	isOwner, err := h.researchService.IsProjectOwnerByGitLabID(projectID, gitlabUserID.(int64))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":      "检查项目所有者失败",
			"details":    err.Error(),
			"user_id":    gitlabUserID,
			"project_id": projectID.String(),
		})
		return
	}

	// 检查权限：项目创建者或管理员可以更新
	isAdmin, _ := c.Get("is_admin")
	if !isOwner && (isAdmin == nil || !isAdmin.(bool)) {
		c.JSON(http.StatusForbidden, gin.H{"error": "无权限修改该课题"})
		return
	}

	updates := make(map[string]interface{})
	if req.Title != nil {
		updates["title"] = *req.Title
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

	// 删除GitLab项目 (DeleteProject方法需要实现)
	if project.GitLabProjectID != nil {
		// TODO: 实现GitLab项目删除
		// if err := h.gitlabService.DeleteProject(*project.GitLabProjectID); err != nil {
		//     c.JSON(http.StatusInternalServerError, gin.H{"error": "删除GitLab项目失败"})
		//     return
		// }
	}

	if err := h.researchService.DeleteResearchProject(projectID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除课题失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "课题删除成功"})
}

// GetMembers 获取课题成员列表
func (h *ResearchHandler) GetMembers(c *gin.Context) {
	// 注意：成员管理已移至GitLab，这里返回提示信息
	c.JSON(http.StatusOK, gin.H{
		"message": "成员管理已移至GitLab，请在GitLab项目中管理成员",
		"members": []interface{}{},
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
		UserID uuid.UUID `json:"user_id" binding:"required"`
		Role   string    `json:"role" binding:"required"`
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

	// 注意：成员管理已移至GitLab
	c.JSON(http.StatusNotImplemented, gin.H{
		"error": "成员管理已移至GitLab，请在GitLab项目中添加成员",
	})
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

	userIDToRemoveStr := c.Param("userId")
	userIDToRemove, err := uuid.Parse(userIDToRemoveStr)
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

	// 不能移除项目所有者
	isTargetOwner, err := h.researchService.IsProjectOwner(projectID, userIDToRemove)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "检查目标用户权限失败"})
		return
	}

	if isTargetOwner {
		c.JSON(http.StatusBadRequest, gin.H{"error": "不能移除项目所有者"})
		return
	}

	// 注意：成员管理已移至GitLab
	c.JSON(http.StatusNotImplemented, gin.H{
		"error": "成员管理已移至GitLab，请在GitLab项目中移除成员",
	})
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

	// TODO: 实现GetProjectIssues方法
	// issues, err := h.gitlabService.GetProjectIssues(*project.GitLabProjectID)
	// if err != nil {
	//     c.JSON(http.StatusInternalServerError, gin.H{"error": "获取Issues失败"})
	//     return
	// }

	// 临时返回空列表
	c.JSON(http.StatusOK, gin.H{"issues": []interface{}{}})
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

	// TODO: 实现GetIssue方法
	// issue, err := h.gitlabService.GetIssue(*project.GitLabProjectID, issueID)
	// if err != nil {
	//     c.JSON(http.StatusNotFound, gin.H{"error": "Issue不存在"})
	//     return
	// }

	// 临时返回模拟数据
	issue := map[string]interface{}{
		"id":    issueID,
		"title": "模拟Issue",
	}

	c.JSON(http.StatusOK, issue)
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
