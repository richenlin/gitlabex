package handlers

import (
	"gitlabex/internal/services"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// PermissionHandler 权限处理器
type PermissionHandler struct {
	gitlabService   *services.GitLabService
	researchService *services.ResearchService
}

// NewPermissionHandler 创建权限处理器
func NewPermissionHandler(gitlabService *services.GitLabService, researchService *services.ResearchService) *PermissionHandler {
	return &PermissionHandler{
		gitlabService:   gitlabService,
		researchService: researchService,
	}
}

// PermissionRequest 权限检查请求
type PermissionRequest struct {
	Action     string `json:"action" binding:"required"`   // 操作类型：create, read, update, delete
	Resource   string `json:"resource" binding:"required"` // 资源类型：project, topic, homework, document
	ResourceID string `json:"resource_id,omitempty"`       // 资源ID（可选）
}

// PermissionResponse 权限检查响应
type PermissionResponse struct {
	Allowed bool   `json:"allowed"`
	Reason  string `json:"reason,omitempty"`
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

	var req PermissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 根据资源类型和操作类型检查权限
	allowed, reason := h.checkResourcePermission(
		accessToken.(string),
		gitlabUserID.(int64),
		req.Resource,
		req.Action,
		req.ResourceID,
		c,
	)

	c.JSON(http.StatusOK, PermissionResponse{
		Allowed: allowed,
		Reason:  reason,
	})
}

// CheckProjectPermission 检查项目权限
func (h *PermissionHandler) CheckProjectPermission(c *gin.Context) {
	projectID := c.Param("id")
	action := c.Query("action") // create, read, update, delete, manage

	if action == "" {
		action = "read" // 默认为读取权限
	}

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

	allowed, reason := h.checkResourcePermission(
		accessToken.(string),
		gitlabUserID.(int64),
		"project",
		action,
		projectID,
		c,
	)

	c.JSON(http.StatusOK, PermissionResponse{
		Allowed: allowed,
		Reason:  reason,
	})
}

// GetUserPermissions 获取用户权限列表
func (h *PermissionHandler) GetUserPermissions(c *gin.Context) {
	gitlabUserID, exists := c.Get("gitlab_user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户未登录"})
		return
	}

	isAdmin, _ := c.Get("is_admin")
	userIsAdmin := isAdmin != nil && isAdmin.(bool)

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
				isOwner := project.CreatorID == gitlabUserID.(int64)
				permissions["can_edit_project"] = userIsAdmin || isOwner
				permissions["can_delete_project"] = userIsAdmin || isOwner
				permissions["can_manage_members"] = userIsAdmin || isOwner

				// 检查GitLab项目权限
				if project.GitLabProjectID != nil {
					accessToken, _ := c.Get("gitlab_access_token")
					if accessToken != nil {
						accessLevel, err := h.gitlabService.GetUserProjectAccessLevel(
							accessToken.(string),
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

	c.JSON(http.StatusOK, gin.H{
		"permissions": permissions,
		"user_id":     gitlabUserID,
		"is_admin":    userIsAdmin,
	})
}

// checkResourcePermission 检查资源权限的核心逻辑
func (h *PermissionHandler) checkResourcePermission(
	accessToken string,
	userID int64,
	resource string,
	action string,
	resourceID string,
	c *gin.Context,
) (bool, string) {
	// 获取用户是否为管理员
	isAdmin, _ := c.Get("is_admin")
	userIsAdmin := isAdmin != nil && isAdmin.(bool)

	// 管理员拥有所有权限
	if userIsAdmin {
		return true, "管理员权限"
	}

	switch resource {
	case "project":
		return h.checkProjectPermission(accessToken, userID, action, resourceID)
	case "topic":
		return h.checkTopicPermission(accessToken, userID, action, resourceID)
	case "homework":
		return h.checkHomeworkPermission(accessToken, userID, action, resourceID)
	case "document":
		return h.checkDocumentPermission(accessToken, userID, action, resourceID)
	default:
		return false, "未知的资源类型"
	}
}

// checkProjectPermission 检查项目权限
func (h *PermissionHandler) checkProjectPermission(accessToken string, userID int64, action string, resourceID string) (bool, string) {
	switch action {
	case "create":
		// 所有登录用户都可以创建项目（具体权限由GitLab控制）
		return true, "用户可以创建项目"

	case "read":
		if resourceID == "" {
			// 读取项目列表 - 所有人都可以
			return true, "可以查看项目列表"
		}

		// 读取特定项目
		projectUUID, err := uuid.Parse(resourceID)
		if err != nil {
			return false, "无效的项目ID"
		}

		project, err := h.researchService.GetResearchProjectByID(projectUUID)
		if err != nil {
			return false, "项目不存在"
		}

		// 公开项目所有人都可以查看
		if project.IsPublic {
			return true, "公开项目"
		}

		// 项目创建者可以查看
		if project.CreatorID == userID {
			return true, "项目创建者"
		}

		// 检查GitLab项目权限
		if project.GitLabProjectID != nil {
			err := h.gitlabService.ValidateRepositoryAccess(accessToken, *project.GitLabProjectID)
			if err == nil {
				return true, "GitLab项目成员"
			}
		}

		return false, "无权访问该项目"

	case "update", "delete", "manage":
		if resourceID == "" {
			return false, "需要指定项目ID"
		}

		projectUUID, err := uuid.Parse(resourceID)
		if err != nil {
			return false, "无效的项目ID"
		}

		project, err := h.researchService.GetResearchProjectByID(projectUUID)
		if err != nil {
			return false, "项目不存在"
		}

		// 项目创建者可以管理
		if project.CreatorID == userID {
			return true, "项目创建者"
		}

		// 检查GitLab项目权限（Maintainer级别以上）
		if project.GitLabProjectID != nil {
			accessLevel, err := h.gitlabService.GetUserProjectAccessLevel(accessToken, *project.GitLabProjectID)
			if err == nil && accessLevel >= 40 { // Maintainer level
				return true, "GitLab项目维护者"
			}
		}

		return false, "无权限管理该项目"

	default:
		return false, "未知的操作类型"
	}
}

// checkTopicPermission 检查话题权限
func (h *PermissionHandler) checkTopicPermission(accessToken string, userID int64, action string, resourceID string) (bool, string) {
	switch action {
	case "create":
		// 所有登录用户都可以创建话题
		return true, "用户可以创建话题"
	case "read":
		// 所有人都可以查看话题
		return true, "可以查看话题"
	case "update", "delete":
		// TODO: 实现话题的编辑和删除权限检查
		return true, "暂时允许编辑话题"
	default:
		return false, "未知的操作类型"
	}
}

// checkHomeworkPermission 检查作业权限
func (h *PermissionHandler) checkHomeworkPermission(accessToken string, userID int64, action string, resourceID string) (bool, string) {
	switch action {
	case "create":
		// 需要检查用户是否为项目的教师或管理员
		return true, "暂时允许创建作业"
	case "read":
		return true, "可以查看作业"
	case "submit":
		return true, "可以提交作业"
	case "grade":
		// 需要检查用户是否为教师
		return true, "暂时允许批改作业"
	default:
		return false, "未知的操作类型"
	}
}

// checkDocumentPermission 检查文档权限
func (h *PermissionHandler) checkDocumentPermission(accessToken string, userID int64, action string, resourceID string) (bool, string) {
	switch action {
	case "create", "upload":
		return true, "可以上传文档"
	case "read", "download":
		return true, "可以查看文档"
	case "update", "delete":
		return true, "暂时允许编辑文档"
	default:
		return false, "未知的操作类型"
	}
}
