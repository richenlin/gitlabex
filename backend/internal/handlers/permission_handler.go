package handlers

import (
	"gitlabex/internal/models"
	"gitlabex/internal/services"
	"gitlabex/internal/types"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// PermissionHandler 权限处理器
type PermissionHandler struct {
	gitlabService   *services.GitLabService
	researchService *services.ResearchService
	topicService    *services.TopicService
}

// NewPermissionHandler 创建权限处理器
func NewPermissionHandler(gitlabService *services.GitLabService, researchService *services.ResearchService, topicService *services.TopicService) *PermissionHandler {
	return &PermissionHandler{
		gitlabService:   gitlabService,
		researchService: researchService,
		topicService:    topicService,
	}
}

// PermissionResponse 权限检查响应
type PermissionResponse struct {
	Allowed bool   `json:"allowed"`
	Reason  string `json:"reason,omitempty"`
}

// ProjectPermissionResponse 项目权限检查响应（扩展版，兼容旧版本）
type ProjectPermissionResponse struct {
	Allowed     bool            `json:"allowed"`          // 兼容旧版本
	Permissions map[string]bool `json:"permissions"`      // 详细权限映射
	Roles       []string        `json:"roles"`            // 用户角色列表
	AccessLevel int             `json:"access_level"`     // GitLab访问级别
	Reason      string          `json:"reason,omitempty"` // 权限检查说明
}

// CheckPermission 统一权限检查接口
func (h *PermissionHandler) CheckPermission(c *gin.Context) {
	// 获取用户信息
	gitlabUserID, exists := c.Get("gitlab_user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户未登录"})
		return
	}

	accessToken, exists := c.Get("gitlab_access_token")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "缺少访问令牌"})
		return
	}

	var req types.PermissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 安全地获取访问令牌和用户ID
	token, ok := accessToken.(string)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "无效的访问令牌"})
		return
	}

	_, ok = gitlabUserID.(int64)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "无效的用户ID"})
		return
	}

	// 构建权限上下文
	ctx := h.buildPermissionContext(c)

	// 根据资源类型和操作类型检查权限
	allowed, reason := h.checkResourcePermission(
		token,
		ctx,
		req.Resource,
		req.Action,
		req.ResourceID,
	)

	c.JSON(http.StatusOK, PermissionResponse{
		Allowed: allowed,
		Reason:  reason,
	})
}

// CheckProjectPermission 检查项目权限（兼容旧版本和新版本）
func (h *PermissionHandler) CheckProjectPermission(c *gin.Context) {
	projectID := c.Param("id")
	action := c.Query("action")               // 兼容旧版本的action参数
	detailed := c.Query("detailed") == "true" // 新参数，用于请求详细信息

	gitlabUserID, exists := c.Get("gitlab_user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户未登录"})
		return
	}

	accessToken, exists := c.Get("gitlab_access_token")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "缺少访问令牌"})
		return
	}

	// 安全地获取管理员标志
	isAdmin, _ := c.Get("is_admin")
	userIsAdmin := false
	if isAdmin != nil {
		if adminBool, ok := isAdmin.(bool); ok {
			userIsAdmin = adminBool
		}
	}

	// 安全地获取访问令牌和用户ID
	token, ok := accessToken.(string)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "无效的访问令牌"})
		return
	}

	userID, ok := gitlabUserID.(int64)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "无效的用户ID"})
		return
	}

	// 如果是旧版本调用（有action参数且不要求详细信息），返回简单格式
	if action != "" && !detailed {
		ctx := h.buildPermissionContext(c)
		allowed, reason := h.checkResourcePermission(
			token,
			ctx,
			models.ResourceProject,
			action,
			projectID,
		)

		c.JSON(http.StatusOK, PermissionResponse{
			Allowed: allowed,
			Reason:  reason,
		})
		return
	}

	// 解析项目ID
	projectUUID, err := uuid.Parse(projectID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的项目ID"})
		return
	}

	// 获取项目信息
	project, err := h.researchService.GetResearchProjectByID(projectUUID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "项目不存在"})
		return
	}

	// 初始化权限和角色
	permissions := map[string]bool{
		"read":            false,
		"edit":            false,
		"manage":          false,
		"create_topic":    false,
		"create_homework": false,
		"grade_homework":  false,
	}
	roles := []string{}
	accessLevel := 0

	// 管理员拥有所有权限
	if userIsAdmin {
		permissions["read"] = true
		permissions["edit"] = true
		permissions["manage"] = true
		permissions["create_topic"] = true
		permissions["create_homework"] = true
		permissions["grade_homework"] = true
		roles = append(roles, "admin")
		accessLevel = 50 // Owner level
	} else {
		// 检查项目创建者权限（创建者通常是教师）
		if project.CreatorID == userID {
			permissions["read"] = true
			permissions["edit"] = true
			permissions["manage"] = true
			permissions["create_topic"] = true
			permissions["create_homework"] = true
			permissions["grade_homework"] = true
			roles = append(roles, "owner", "teacher")
			accessLevel = 50
		} else if project.IsPublic {
			// 公开项目的基本权限（游客级别）
			permissions["read"] = true
			roles = append(roles, "guest")
			accessLevel = 10
		}

		// 检查GitLab项目权限
		if project.GitLabProjectID != nil {
			gitlabAccessLevel, err := h.gitlabService.GetUserProjectAccessLevel(token, *project.GitLabProjectID)
			if err == nil {
				accessLevel = gitlabAccessLevel

				// 根据GitLab访问级别设置权限和角色
				// 10: Guest - 游客
				if gitlabAccessLevel >= 10 {
					permissions["read"] = true
					if !contains(roles, "guest") {
						roles = append(roles, "guest")
					}
				}
				// 20: Reporter - 普通用户/学生
				if gitlabAccessLevel >= 20 {
					permissions["create_topic"] = true
					if !contains(roles, "reporter") {
						roles = append(roles, "reporter", "student")
					}
				}
				// 30: Developer - 研究员
				if gitlabAccessLevel >= 30 {
					permissions["grade_homework"] = true // 研究员可以批改作业
					if !contains(roles, "developer") {
						roles = append(roles, "developer", "researcher")
					}
				}
				// 40: Maintainer - 教师
				if gitlabAccessLevel >= 40 {
					permissions["edit"] = true
					permissions["manage"] = true
					permissions["create_homework"] = true
					permissions["grade_homework"] = true
					if !contains(roles, "maintainer") {
						roles = append(roles, "maintainer", "teacher")
					}
				}
				// 50: Owner - 管理员/项目所有者
				if gitlabAccessLevel >= 50 {
					permissions["edit"] = true
					permissions["manage"] = true
					permissions["create_homework"] = true
					permissions["grade_homework"] = true
					if !contains(roles, "owner") {
						roles = append(roles, "owner", "teacher")
					}
				}
			}
		}
	}

	// 计算总体权限状态（用于兼容性）
	allowed := permissions["read"] || permissions["edit"] || permissions["manage"]

	c.JSON(http.StatusOK, ProjectPermissionResponse{
		Allowed:     allowed,
		Permissions: permissions,
		Roles:       roles,
		AccessLevel: accessLevel,
		Reason:      "权限检查完成",
	})
}

// contains 辅助函数，检查字符串切片是否包含指定字符串
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// GetUserPermissions 获取用户权限列表
func (h *PermissionHandler) GetUserPermissions(c *gin.Context) {
	gitlabUserID, exists := c.Get("gitlab_user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户未登录"})
		return
	}

	// 安全地获取管理员标志
	isAdmin, _ := c.Get("is_admin")
	userIsAdmin := false
	if isAdmin != nil {
		if adminBool, ok := isAdmin.(bool); ok {
			userIsAdmin = adminBool
		}
	}

	// 安全地获取用户ID
	userID, ok := gitlabUserID.(int64)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "无效的用户ID"})
		return
	}

	permissions := map[string]bool{
		"can_create_project":  userIsAdmin || true, // 暂时允许所有用户创建项目
		"can_manage_users":    userIsAdmin,
		"can_view_admin":      userIsAdmin,
		"can_create_topic":    true, // 所有登录用户都可以创建话题
		"can_upload_document": true, // 所有登录用户都可以上传文档
	}

	// 如果提供了项目ID，检查项目级别的权限
	projectID := c.Query("project_id")
	if projectID != "" {
		projectUUID, err := uuid.Parse(projectID)
		if err == nil {
			project, err := h.researchService.GetResearchProjectByID(projectUUID)
			if err == nil {
				// 检查项目权限
				isOwner := project.CreatorID == userID
				permissions["can_edit_project"] = userIsAdmin || isOwner
				permissions["can_delete_project"] = userIsAdmin || isOwner
				permissions["can_manage_members"] = userIsAdmin || isOwner

				// 检查GitLab项目权限
				if project.GitLabProjectID != nil {
					accessToken, _ := c.Get("gitlab_access_token")
					if accessToken != nil {
						if token, ok := accessToken.(string); ok {
							accessLevel, err := h.gitlabService.GetUserProjectAccessLevel(
								token,
								*project.GitLabProjectID,
							)
							if err == nil {
								permissions["can_push_code"] = accessLevel >= 30 // Developer level
								permissions["can_create_merge_request"] = accessLevel >= 30
								permissions["can_manage_issues"] = accessLevel >= 20 // Reporter level
							}
						}
					}
				}
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"permissions": permissions,
		"user_id":     userID,
		"is_admin":    userIsAdmin,
	})
}

// buildPermissionContext 构建权限上下文
func (h *PermissionHandler) buildPermissionContext(c *gin.Context) *models.PermissionContext {
	ctx := &models.PermissionContext{}

	// 获取用户ID
	if gitlabUserID, exists := c.Get("gitlab_user_id"); exists {
		if userID, ok := gitlabUserID.(int64); ok {
			ctx.UserID = userID
		}
	}

	// 获取管理员标志
	if isAdmin, exists := c.Get("is_admin"); exists {
		if adminBool, ok := isAdmin.(bool); ok {
			ctx.IsAdmin = adminBool
		}
	}

	// 获取访问级别
	if accessLevel, exists := c.Get("user_access_level"); exists {
		if level, ok := accessLevel.(int); ok {
			ctx.AccessLevel = level
			ctx.Role = models.ParseGitLabRole(level)
		}
	}

	// 获取项目ID
	if projectID, exists := c.Get("project_id"); exists {
		if pid, ok := projectID.(int64); ok {
			ctx.ProjectID = &pid
		}
	}

	return ctx
}

// checkResourcePermission 检查资源权限
func (h *PermissionHandler) checkResourcePermission(
	accessToken string,
	ctx *models.PermissionContext,
	resource string,
	action string,
	resourceID string,
) (bool, string) {
	// 管理员拥有所有权限
	if ctx.IsAdmin {
		return true, "管理员拥有所有权限"
	}

	switch resource {
	case models.ResourceProject:
		return h.checkProjectPermission(accessToken, ctx, action, resourceID)
	case models.ResourceTopic:
		return h.checkTopicPermission(accessToken, ctx, action, resourceID)
	case models.ResourceHomework:
		return h.checkHomeworkPermission(accessToken, ctx, action, resourceID)
	case models.ResourceDocument:
		return h.checkDocumentPermission(accessToken, ctx, action, resourceID)
	default:
		return false, "未知的资源类型"
	}
}

// checkProjectPermission 检查课题权限（按照新需求）
// 需求：
// 1. 管理员拥有一切权限
// 2. 教师可以新建、编辑、删除课题（包括课题包含的话题、作业）
func (h *PermissionHandler) checkProjectPermission(accessToken string, ctx *models.PermissionContext, action string, resourceID string) (bool, string) {
	switch action {
	case models.ActionCreate:
		// 1. 管理员可以创建课题 - 已在上层检查
		// 2. 教师可以创建课题
		// 由于创建时还没有具体项目，暂时允许所有登录用户创建
		// 实际应该检查用户在GitLab中的全局角色
		return true, "允许创建课题"

	case models.ActionRead:
		if resourceID == "" {
			// 所有人都可以查看课题列表（包括游客）
			return true, "可以查看课题列表"
		}

		// 读取特定课题
		projectUUID, err := uuid.Parse(resourceID)
		if err != nil {
			return false, "无效的课题ID"
		}

		project, err := h.researchService.GetResearchProjectByID(projectUUID)
		if err != nil {
			return false, "课题不存在"
		}

		// 公开课题所有人都可以查看
		if project.IsPublic {
			return true, "公开课题"
		}

		// 课题创建者可以查看
		if project.CreatorID == ctx.UserID {
			return true, "课题创建者"
		}

		// 检查GitLab项目权限（至少Guest级别）
		if project.GitLabProjectID != nil {
			accessLevel, err := h.gitlabService.GetUserProjectAccessLevel(accessToken, *project.GitLabProjectID)
			if err == nil && accessLevel >= 10 {
				return true, "课题成员"
			}
		}

		return false, "无权访问该课题"

	case models.ActionUpdate, models.ActionDelete:
		if resourceID == "" {
			return false, "需要指定课题ID"
		}

		projectUUID, err := uuid.Parse(resourceID)
		if err != nil {
			return false, "无效的课题ID"
		}

		project, err := h.researchService.GetResearchProjectByID(projectUUID)
		if err != nil {
			return false, "课题不存在"
		}

		// 2. 教师可以编辑/删除课题
		// 检查是否为课题创建者或GitLab项目的Maintainer
		if project.CreatorID == ctx.UserID {
			return true, "课题创建者（教师）"
		}

		if project.GitLabProjectID != nil {
			accessLevel, err := h.gitlabService.GetUserProjectAccessLevel(accessToken, *project.GitLabProjectID)
			if err == nil && accessLevel >= 40 { // Maintainer/教师级别
				return true, "课题教师"
			}
		}

		return false, "只有管理员和教师可以编辑/删除课题"

	default:
		return false, "未知的操作类型"
	}
}

// checkTopicPermission 检查话题权限（按照新需求）
// 需求：
// 1. 管理员拥有一切权限
// 2. 教师可以新建、编辑、删除课题包含的话题
// 3. 研究员可以编辑所属课题的话题，可以发表话题，点赞、评论
// 4. 普通用户可以发表话题，点赞、评论
func (h *PermissionHandler) checkTopicPermission(accessToken string, ctx *models.PermissionContext, action string, resourceID string) (bool, string) {
	switch action {
	case models.ActionCreate:
		// 所有登录用户都可以创建话题
		return true, "可以创建话题"

	case models.ActionRead:
		// 所有人都可以查看话题（包括游客）
		return true, "可以查看话题"

	case models.ActionUpdate:
		if resourceID == "" {
			return false, "需要指定话题ID"
		}

		topicUUID, err := uuid.Parse(resourceID)
		if err != nil {
			return false, "无效的话题ID"
		}

		topic, err := h.topicService.GetTopicByID(topicUUID)
		if err != nil {
			return false, "话题不存在"
		}

		// 2. 教师可以编辑课题包含的话题
		if topic.ProjectID != nil {
			project, err := h.researchService.GetResearchProjectByID(*topic.ProjectID)
			if err == nil {
				// 检查是否为课题创建者（教师）
				if project.CreatorID == ctx.UserID {
					return true, "课题教师"
				}

				// 检查GitLab项目权限
				if project.GitLabProjectID != nil {
					accessLevel, err := h.gitlabService.GetUserProjectAccessLevel(accessToken, *project.GitLabProjectID)
					if err == nil && accessLevel >= 40 { // Maintainer/教师级别
						return true, "课题教师"
					}
					// 3. 研究员可以编辑所属课题的话题（只能编辑自己的）
					if err == nil && accessLevel >= 30 { // Developer/研究员级别
						if topic.AuthorID == ctx.UserID {
							return true, "话题作者（研究员）"
						}
						return false, "研究员只能编辑自己的话题"
					}
				}
			}
		}

		// 话题作者可以编辑自己的话题
		if topic.AuthorID == ctx.UserID {
			return true, "话题作者"
		}

		return false, "无权编辑该话题"

	case models.ActionDelete:
		if resourceID == "" {
			return false, "需要指定话题ID"
		}

		topicUUID, err := uuid.Parse(resourceID)
		if err != nil {
			return false, "无效的话题ID"
		}

		topic, err := h.topicService.GetTopicByID(topicUUID)
		if err != nil {
			return false, "话题不存在"
		}

		// 2. 教师可以删除课题包含的话题
		if topic.ProjectID != nil {
			project, err := h.researchService.GetResearchProjectByID(*topic.ProjectID)
			if err == nil {
				if project.CreatorID == ctx.UserID {
					return true, "课题教师"
				}

				if project.GitLabProjectID != nil {
					accessLevel, err := h.gitlabService.GetUserProjectAccessLevel(accessToken, *project.GitLabProjectID)
					if err == nil && accessLevel >= 40 { // Maintainer/教师级别
						return true, "课题教师"
					}
				}
			}
		}

		// 话题作者可以删除自己的话题
		if topic.AuthorID == ctx.UserID {
			return true, "话题作者"
		}

		return false, "只有管理员、教师和话题作者可以删除话题"

	case models.ActionLike, models.ActionComment:
		// 所有登录用户都可以点赞和评论
		return true, "可以点赞和评论"

	default:
		return false, "未知的操作类型"
	}
}

// checkHomeworkPermission 检查作业权限（按照新需求）
// 需求：
// 1. 管理员拥有一切权限
// 2. 教师可以批改课题作业
// 3. 研究员可以批改课题作业
// 4. 普通用户可以提交作业
func (h *PermissionHandler) checkHomeworkPermission(accessToken string, ctx *models.PermissionContext, action string, resourceID string) (bool, string) {
	switch action {
	case models.ActionCreate:
		// 1. 管理员可以创建作业 - 已在上层检查
		// 2. 教师可以创建作业
		// 暂时允许所有登录用户创建，实际应该在handler中检查课题权限
		return true, "允许创建作业"

	case models.ActionRead:
		// 所有登录用户都可以查看作业
		return true, "可以查看作业"

	case models.ActionUpdate, models.ActionDelete:
		// 1. 管理员可以编辑/删除作业 - 已在上层检查
		// 2. 教师可以编辑/删除课题的作业
		// 需要在handler中检查作业所属课题的权限
		return true, "允许编辑/删除作业"

	case models.ActionGrade:
		// 1. 管理员可以批改作业 - 已在上层检查
		// 2. 教师可以批改课题作业
		// 3. 研究员可以批改课题作业
		// 需要在handler中检查课题权限（Developer级别以上）
		return true, "允许批改作业"

	case models.ActionSubmit:
		// 4. 普通用户可以提交作业
		// 所有登录用户都可以提交作业
		return true, "可以提交作业"

	default:
		return false, "未知的操作类型"
	}
}

// checkDocumentPermission 检查文档权限（按照新需求）
// 需求：
// 1. 管理员拥有一切权限
// 2. 教师可以新建、编辑、同步文档，可以审核文档
// 3. 研究员可以新建、编辑自己的文档
// 4. 普通用户可以新建、编辑自己的文档
func (h *PermissionHandler) checkDocumentPermission(accessToken string, ctx *models.PermissionContext, action string, resourceID string) (bool, string) {
	switch action {
	case models.ActionCreate:
		// 所有登录用户都可以创建文档
		return true, "可以创建文档"

	case models.ActionRead:
		// 所有人都可以查看文档
		return true, "可以查看文档"

	case models.ActionUpdate:
		// 需要在handler中检查文档所有者
		// 1. 管理员可以编辑任何文档 - 已在上层检查
		// 2. 教师可以直接编辑文档
		// 3. 研究员可以编辑自己的文档
		// 4. 普通用户可以编辑自己的文档
		return true, "允许编辑文档"

	case models.ActionDelete:
		// 1. 管理员可以删除任何文档 - 已在上层检查
		// 2. 教师可以删除文档
		// 3. 文档所有者可以删除自己的文档
		return true, "允许删除文档"

	case models.ActionSync:
		// 1. 管理员可以同步文档 - 已在上层检查
		// 2. 教师可以同步文档
		return true, "允许同步文档"

	case models.ActionApprove:
		// 1. 管理员可以审核文档 - 已在上层检查
		// 2. 教师可以审核文档
		return true, "允许审核文档"

	default:
		return false, "未知的操作类型"
	}
}
