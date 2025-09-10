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
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户未登录"})
		return
	}

	// 获取分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	isPublic := c.Query("is_public") == "true"
	includePrivate := c.Query("include_private") != "false"

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	offset := (page - 1) * limit

	// 获取当前用户信息
	currentUser, err := h.userService.GetUserByID(userID.(uuid.UUID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取用户信息失败"})
		return
	}

	var projects []models.ResearchProject
	var total int64

	// 根据用户角色和查询参数获取项目
	if currentUser.EduRole >= models.EduRoleTeacher {
		// 教师和管理员可以看到所有项目
		projects, total, err = h.researchService.GetAllProjects(limit, offset, isPublic, includePrivate)
	} else {
		// 学生只能看到公开项目和自己参与的项目
		projects, total, err = h.researchService.GetUserAccessibleProjects(userID.(uuid.UUID), limit, offset)
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
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户未登录"})
		return
	}

	var req struct {
		Title       string     `json:"title" binding:"required"`
		Description string     `json:"description"`
		GitLabURL   string     `json:"gitlab_url"`
		IsPublic    bool       `json:"is_public"`
		StartDate   time.Time  `json:"start_date"`
		EndDate     *time.Time `json:"end_date"`
		Tags        []string   `json:"tags"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 获取当前用户信息
	currentUser, err := h.userService.GetUserByID(userID.(uuid.UUID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取用户信息失败"})
		return
	}

	// 检查权限 - 只有教师和管理员可以创建项目
	if currentUser.EduRole < models.EduRoleTeacher {
		c.JSON(http.StatusForbidden, gin.H{"error": "无权限创建课题"})
		return
	}

	// 获取用户的访问令牌 (这里需要从用户信息中获取)
	accessToken := currentUser.AccessToken // 假设从当前用户获取
	if accessToken == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "缺少GitLab访问令牌"})
		return
	}

	// 创建GitLab项目
	visibility := "private"
	if req.IsPublic {
		visibility = "public"
	}

	createReq := &services.CreateProjectRequest{
		Name:                 req.Title,
		Description:          req.Description,
		Visibility:           visibility,
		InitializeWithReadme: true,
		DefaultBranch:        "main",
	}

	gitlabProject, err := h.gitlabService.CreateProject(accessToken, createReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建GitLab项目失败: " + err.Error()})
		return
	}

	project := &models.ResearchProject{
		Name:            req.Title,
		Description:     req.Description,
		GitLabURL:       gitlabProject.WebURL,
		GitLabProjectID: &gitlabProject.ID,
		CreatorID:       userID.(uuid.UUID),
		IsPublic:        req.IsPublic,
		StartDate:       req.StartDate,
		EndDate:         req.EndDate,
		Status:          "active",
	}

	if err := h.researchService.CreateResearchProject(project); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建课题失败"})
		return
	}

	// 自动添加创建者为项目管理员
	if err := h.researchService.AddProjectMember(project.ID, userID.(uuid.UUID), models.ProjectRoleOwner); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "添加项目成员失败"})
		return
	}

	c.JSON(http.StatusCreated, project)
}

// GetResearchProjectByID 根据ID获取研究课题详情
func (h *ResearchHandler) GetResearchProjectByID(c *gin.Context) {
	userID, exists := c.Get("userID")
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

	project, err := h.researchService.GetResearchProjectByID(projectID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "课题不存在"})
		return
	}

	// 检查访问权限
	currentUser, err := h.userService.GetUserByID(userID.(uuid.UUID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取用户信息失败"})
		return
	}

	// 教师和管理员可以访问所有项目
	// 学生只能访问公开项目或自己参与的项目
	if currentUser.EduRole < models.EduRoleTeacher && !project.IsPublic {
		isMember, err := h.researchService.IsProjectMember(projectID, userID.(uuid.UUID))
		if err != nil || !isMember {
			c.JSON(http.StatusForbidden, gin.H{"error": "无权限访问该课题"})
			return
		}
	}

	c.JSON(http.StatusOK, project)
}

// UpdateResearchProject 更新研究课题信息
func (h *ResearchHandler) UpdateResearchProject(c *gin.Context) {
	userID, exists := c.Get("userID")
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
	isOwner, err := h.researchService.IsProjectOwner(projectID, userID.(uuid.UUID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "检查权限失败"})
		return
	}

	currentUser, err := h.userService.GetUserByID(userID.(uuid.UUID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取用户信息失败"})
		return
	}

	if !isOwner && currentUser.EduRole < models.EduRoleAdmin {
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
	userID, exists := c.Get("userID")
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
	isOwner, err := h.researchService.IsProjectOwner(projectID, userID.(uuid.UUID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "检查权限失败"})
		return
	}

	currentUser, err := h.userService.GetUserByID(userID.(uuid.UUID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取用户信息失败"})
		return
	}

	if !isOwner && currentUser.EduRole < models.EduRoleAdmin {
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
	projectIDStr := c.Param("id")
	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的课题ID"})
		return
	}

	members, err := h.researchService.GetProjectMembers(projectID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取成员列表失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"members": members})
}

// AddMember 添加课题成员
func (h *ResearchHandler) AddMember(c *gin.Context) {
	userID, exists := c.Get("userID")
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
	isOwner, err := h.researchService.IsProjectOwner(projectID, userID.(uuid.UUID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "检查权限失败"})
		return
	}

	currentUser, err := h.userService.GetUserByID(userID.(uuid.UUID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取用户信息失败"})
		return
	}

	if !isOwner && currentUser.EduRole < models.EduRoleAdmin {
		c.JSON(http.StatusForbidden, gin.H{"error": "无权限添加成员"})
		return
	}

	// 映射角色到项目角色
	var projectRole models.ProjectRole
	switch req.Role {
	case "owner":
		projectRole = models.ProjectRoleOwner
	case "maintainer":
		projectRole = models.ProjectRoleMaintainer
	case "developer":
		projectRole = models.ProjectRoleDeveloper
	case "reporter":
		projectRole = models.ProjectRoleReporter
	default:
		projectRole = models.ProjectRoleReporter
	}

	if err := h.researchService.AddProjectMember(projectID, req.UserID, projectRole); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "添加成员失败"})
		return
	}

	// 同步到GitLab项目权限 (AddProjectMember方法需要实现)
	// TODO: 实现GitLab项目成员添加
	// if err := h.gitlabService.AddProjectMember(projectID, req.UserID, projectRole); err != nil {
	//     c.JSON(http.StatusInternalServerError, gin.H{"error": "同步GitLab权限失败"})
	//     return
	// }

	c.JSON(http.StatusOK, gin.H{"message": "成员添加成功"})
}

// RemoveMember 移除课题成员
func (h *ResearchHandler) RemoveMember(c *gin.Context) {
	userID, exists := c.Get("userID")
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
	isOwner, err := h.researchService.IsProjectOwner(projectID, userID.(uuid.UUID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "检查权限失败"})
		return
	}

	currentUser, err := h.userService.GetUserByID(userID.(uuid.UUID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取用户信息失败"})
		return
	}

	if !isOwner && currentUser.EduRole < models.EduRoleAdmin {
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

	if err := h.researchService.RemoveProjectMember(projectID, userIDToRemove); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "移除成员失败"})
		return
	}

	// 同步到GitLab (RemoveProjectMember方法需要实现)
	// TODO: 实现GitLab项目成员移除
	// if err := h.gitlabService.RemoveProjectMember(projectID, userIDToRemove); err != nil {
	//     c.JSON(http.StatusInternalServerError, gin.H{"error": "同步GitLab权限失败"})
	//     return
	// }

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
	userID, exists := c.Get("userID")
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

	// 检查权限
	isMember, err := h.researchService.IsProjectMember(projectID, userID.(uuid.UUID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "检查权限失败"})
		return
	}

	if !isMember {
		c.JSON(http.StatusForbidden, gin.H{"error": "无权限创建Issue"})
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

	// 获取当前用户信息
	currentUser, err := h.userService.GetUserByID(userID.(uuid.UUID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取用户信息失败"})
		return
	}

	// 获取用户的访问令牌
	accessToken := currentUser.AccessToken
	if accessToken == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "缺少GitLab访问令牌"})
		return
	}

	// 创建GitLab Issue
	issue, err := h.gitlabService.CreateIssue(accessToken, *project.GitLabProjectID, req.Title, req.Description, req.Labels, req.AssigneeID)
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
	userID, exists := c.Get("userID")
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

	// 检查权限
	isMember, err := h.researchService.IsProjectMember(projectID, userID.(uuid.UUID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "检查权限失败"})
		return
	}

	if !isMember {
		c.JSON(http.StatusForbidden, gin.H{"error": "无权限创建讨论"})
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
	userID, exists := c.Get("userID")
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
	isOwner, err := h.researchService.IsProjectOwner(projectID, userID.(uuid.UUID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "检查权限失败"})
		return
	}

	currentUser, err := h.userService.GetUserByID(userID.(uuid.UUID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取用户信息失败"})
		return
	}

	if !isOwner && currentUser.EduRole < models.EduRoleTeacher {
		c.JSON(http.StatusForbidden, gin.H{"error": "无权限创建作业"})
		return
	}

	homework := &models.Homework{
		ProjectID:   projectID,
		Title:       req.Title,
		Description: req.Description,
		DueDate:     req.DueDate,
		MaxGrade:    req.MaxGrade,
		CreatorID:   userID.(uuid.UUID),
		Status:      "active",
	}

	if err := h.researchService.CreateProjectHomework(homework); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建作业失败"})
		return
	}

	c.JSON(http.StatusCreated, homework)
}
