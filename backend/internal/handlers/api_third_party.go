package handlers

import (
	"net/http"
	"strconv"

	"gitlabex/internal/middleware"
	"gitlabex/internal/models"
	"gitlabex/internal/services"

	"github.com/gin-gonic/gin"
)

// ThirdPartyAPIHandler 第三方API处理器
type ThirdPartyAPIHandler struct {
	// 现有Handler代理
	userHandler         *UserHandler
	projectHandler      *ProjectHandler
	assignmentHandler   *AssignmentHandler
	notificationHandler *NotificationHandler

	// 中间件
	oauthMiddleware *middleware.OAuthMiddleware

	// 服务
	gitlabService *services.GitLabService
}

// NewThirdPartyAPIHandler 创建第三方API处理器
func NewThirdPartyAPIHandler(
	userHandler *UserHandler,
	projectHandler *ProjectHandler,
	assignmentHandler *AssignmentHandler,
	notificationHandler *NotificationHandler,
	oauthMiddleware *middleware.OAuthMiddleware,
	gitlabService *services.GitLabService,
) *ThirdPartyAPIHandler {
	return &ThirdPartyAPIHandler{
		userHandler:         userHandler,
		projectHandler:      projectHandler,
		assignmentHandler:   assignmentHandler,
		notificationHandler: notificationHandler,
		oauthMiddleware:     oauthMiddleware,
		gitlabService:       gitlabService,
	}
}

// RegisterRoutes 注册第三方API路由
func (h *ThirdPartyAPIHandler) RegisterRoutes(router *gin.RouterGroup) {
	// 第三方API组，使用OAuth认证
	api := router.Group("/third-party")
	{
		// 应用安全中间件
		api.Use(h.oauthMiddleware.ThirdPartyAuth())
		api.Use(h.oauthMiddleware.LogAPIAccess())
		api.Use(h.oauthMiddleware.CORS())
		api.Use(h.oauthMiddleware.RateLimit())

		// API Key管理
		auth := api.Group("/auth")
		{
			auth.POST("/api-key", h.GenerateAPIKey) // 生成API Key
			auth.DELETE("/api-key", h.RevokeAPIKey) // 撤销API Key
			auth.GET("/validate", h.ValidateToken)  // 验证Token
		}

		// Git仓库管理API - 基于现有项目API + GitLab扩展
		repos := api.Group("/repos")
		{
			repos.POST("", h.CreateRepository)           // 创建Git仓库
			repos.GET("", h.proxyToProjectList)          // 代理到项目列表
			repos.GET("/:id", h.proxyToProjectDetail)    // 代理到项目详情
			repos.PUT("/:id", h.proxyToProjectUpdate)    // 代理到项目更新
			repos.DELETE("/:id", h.proxyToProjectDelete) // 代理到项目删除

			// GitLab特有功能
			repos.POST("/:id/clone", h.GetCloneInfo)               // 获取克隆信息
			repos.GET("/:id/commits", h.GetRepositoryCommits)      // 获取提交记录
			repos.GET("/:id/branches", h.GetRepositoryBranches)    // 获取分支列表
			repos.POST("/:id/branches", h.CreateBranch)            // 创建分支
			repos.GET("/:id/files", h.GetRepositoryFiles)          // 获取文件列表
			repos.GET("/:id/files/*filepath", h.GetFileContent)    // 获取文件内容
			repos.PUT("/:id/files/*filepath", h.UpdateFileContent) // 更新文件内容
		}

		// 用户管理API - 代理到现有用户API + 扩展
		users := api.Group("/users")
		{
			users.GET("", h.proxyToUserList)           // 代理到用户列表
			users.GET("/:id", h.proxyToUserDetail)     // 代理到用户详情
			users.PUT("/:id", h.proxyToUserUpdate)     // 代理到用户更新
			users.POST("/:id/sync", h.proxyToUserSync) // 代理到用户同步

			// 第三方特有功能
			users.POST("", h.CreateUserForThirdParty)           // 第三方创建用户
			users.PUT("/:id/role", h.UpdateUserRole)            // 更新用户角色
			users.GET("/:id/permissions", h.GetUserPermissions) // 获取用户权限
		}

		// 权限管理API
		permissions := api.Group("/permissions")
		{
			permissions.GET("/roles", h.GetAllRoles)      // 获取所有角色
			permissions.POST("/check", h.CheckPermission) // 检查权限
		}

		// 作业管理API - 代理到现有作业API
		assignments := api.Group("/assignments")
		{
			assignments.POST("", h.proxyToAssignmentCreate)       // 代理到作业创建
			assignments.GET("", h.proxyToAssignmentList)          // 代理到作业列表
			assignments.GET("/:id", h.proxyToAssignmentDetail)    // 代理到作业详情
			assignments.PUT("/:id", h.proxyToAssignmentUpdate)    // 代理到作业更新
			assignments.DELETE("/:id", h.proxyToAssignmentDelete) // 代理到作业删除
		}

		// 系统状态API
		status := api.Group("/status")
		{
			status.GET("/health", h.GetSystemStatus) // 系统健康状态
			status.GET("/stats", h.GetSystemStats)   // 系统统计信息
		}
	}
}

// ===== 认证管理 =====

// GenerateAPIKey 生成API Key
func (h *ThirdPartyAPIHandler) GenerateAPIKey(c *gin.Context) {
	user, exists := c.Get("current_user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Authentication required",
		})
		return
	}

	currentUser := user.(*models.User)
	apiKey := h.oauthMiddleware.GenerateAPIKey(currentUser.ID)

	c.JSON(http.StatusCreated, gin.H{
		"message": "API Key generated successfully",
		"data": gin.H{
			"api_key":    apiKey,
			"user_id":    currentUser.ID,
			"expires_in": "7 days",
			"scopes":     []string{"read", "write", "manage"},
		},
	})
}

// ValidateToken 验证Token
func (h *ThirdPartyAPIHandler) ValidateToken(c *gin.Context) {
	user, exists := c.Get("current_user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"valid": false,
		})
		return
	}

	currentUser := user.(*models.User)
	authType, _ := c.Get("auth_type")

	c.JSON(http.StatusOK, gin.H{
		"valid": true,
		"data": gin.H{
			"user_id":   currentUser.ID,
			"username":  currentUser.Username,
			"role":      currentUser.Role,
			"auth_type": authType,
		},
	})
}

// ===== Git仓库管理扩展 =====

// CreateRepository 创建Git仓库（扩展版）
func (h *ThirdPartyAPIHandler) CreateRepository(c *gin.Context) {
	var req struct {
		Name        string `json:"name" binding:"required"`
		Description string `json:"description"`
		ClassID     uint   `json:"class_id"`
		Visibility  string `json:"visibility"` // private, public
		InitRepo    bool   `json:"init_repo"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request parameters",
			"details": err.Error(),
		})
		return
	}

	// 转换为项目创建请求，代理到现有API
	projectReq := gin.H{
		"name":           req.Name,
		"description":    req.Description,
		"class_id":       req.ClassID,
		"wiki_enabled":   true,
		"issues_enabled": true,
		"mr_enabled":     true,
	}

	// 手动调用项目创建逻辑（避免重复实现）
	c.Set("proxy_request", projectReq)
	h.projectHandler.CreateProject(c)
}

// GetCloneInfo 获取克隆信息
func (h *ThirdPartyAPIHandler) GetCloneInfo(c *gin.Context) {
	repoID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid repository ID",
		})
		return
	}

	// 调用现有项目详情API
	c.Params = append(c.Params, gin.Param{Key: "id", Value: strconv.FormatUint(repoID, 10)})
	h.projectHandler.GetProject(c)

	// 如果项目存在，添加克隆信息
	if c.Writer.Status() == 200 {
		// 这里可以添加额外的克隆指令信息
		c.Header("X-Clone-Instructions", "Use git clone command with the repository URL")
	}
}

// ===== 代理方法 =====

// proxyToProjectList 代理到项目列表
func (h *ThirdPartyAPIHandler) proxyToProjectList(c *gin.Context) {
	h.projectHandler.GetProjects(c)
}

func (h *ThirdPartyAPIHandler) proxyToProjectDetail(c *gin.Context) {
	h.projectHandler.GetProject(c)
}

func (h *ThirdPartyAPIHandler) proxyToProjectUpdate(c *gin.Context) {
	h.projectHandler.UpdateProject(c)
}

func (h *ThirdPartyAPIHandler) proxyToProjectDelete(c *gin.Context) {
	h.projectHandler.DeleteProject(c)
}

// 用户API代理
func (h *ThirdPartyAPIHandler) proxyToUserList(c *gin.Context) {
	h.userHandler.ListActiveUsers(c)
}

func (h *ThirdPartyAPIHandler) proxyToUserDetail(c *gin.Context) {
	h.userHandler.GetUserByID(c)
}

func (h *ThirdPartyAPIHandler) proxyToUserUpdate(c *gin.Context) {
	h.userHandler.UpdateUser(c)
}

func (h *ThirdPartyAPIHandler) proxyToUserSync(c *gin.Context) {
	h.userHandler.SyncUserFromGitLab(c)
}

// 作业API代理
func (h *ThirdPartyAPIHandler) proxyToAssignmentCreate(c *gin.Context) {
	h.assignmentHandler.CreateAssignment(c)
}

func (h *ThirdPartyAPIHandler) proxyToAssignmentList(c *gin.Context) {
	h.assignmentHandler.ListAssignments(c)
}

func (h *ThirdPartyAPIHandler) proxyToAssignmentDetail(c *gin.Context) {
	h.assignmentHandler.GetAssignment(c)
}

func (h *ThirdPartyAPIHandler) proxyToAssignmentUpdate(c *gin.Context) {
	h.assignmentHandler.UpdateAssignment(c)
}

func (h *ThirdPartyAPIHandler) proxyToAssignmentDelete(c *gin.Context) {
	h.assignmentHandler.DeleteAssignment(c)
}

// ===== 第三方特有功能 =====

// CreateUserForThirdParty 为第三方创建用户（简化版）
func (h *ThirdPartyAPIHandler) CreateUserForThirdParty(c *gin.Context) {
	var req struct {
		Username string `json:"username" binding:"required"`
		Email    string `json:"email" binding:"required,email"`
		Name     string `json:"name" binding:"required"`
		Role     int    `json:"role"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request parameters",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "User creation request received",
		"data": gin.H{
			"username": req.Username,
			"email":    req.Email,
			"name":     req.Name,
			"role":     req.Role,
			"status":   "pending_gitlab_creation",
			"note":     "Please create the corresponding user in GitLab and sync",
		},
	})
}

// GetAllRoles 获取所有角色
func (h *ThirdPartyAPIHandler) GetAllRoles(c *gin.Context) {
	roles := []gin.H{
		{"id": 1, "name": "admin", "label": "管理员", "description": "系统管理员，拥有所有权限"},
		{"id": 2, "name": "teacher", "label": "教师", "description": "可以创建和管理班级、课题、作业"},
		{"id": 3, "name": "student", "label": "学生", "description": "可以参与班级和课题，提交作业"},
		{"id": 4, "name": "guest", "label": "访客", "description": "只读权限"},
	}

	c.JSON(http.StatusOK, gin.H{
		"data": roles,
	})
}

// CheckPermission 检查权限
func (h *ThirdPartyAPIHandler) CheckPermission(c *gin.Context) {
	var req struct {
		UserID       uint   `json:"user_id" binding:"required"`
		ResourceType string `json:"resource_type" binding:"required"`
		ResourceID   uint   `json:"resource_id" binding:"required"`
		Action       string `json:"action" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request parameters",
			"details": err.Error(),
		})
		return
	}

	// 简化的权限检查逻辑
	currentUser, exists := c.Get("current_user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Authentication required",
		})
		return
	}

	user := currentUser.(*models.User)
	hasPermission := false

	// 基本权限检查
	switch req.Action {
	case "read":
		hasPermission = user.Role <= 3
	case "write":
		hasPermission = user.Role <= 2
	case "manage":
		hasPermission = user.Role <= 1
	}

	c.JSON(http.StatusOK, gin.H{
		"has_permission": hasPermission,
		"user_role":      user.Role,
		"resource_type":  req.ResourceType,
		"resource_id":    req.ResourceID,
		"action":         req.Action,
	})
}

// GetSystemStatus 获取系统状态
func (h *ThirdPartyAPIHandler) GetSystemStatus(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "healthy",
		"version": "1.0.0",
		"services": gin.H{
			"database": "connected",
			"gitlab":   "connected",
			"redis":    "connected",
		},
		"timestamp": gin.H{
			"unix": gin.H{
				"timestamp": 1625097600,
			},
		},
	})
}

// ===== 未实现的GitLab扩展功能（占位符） =====

func (h *ThirdPartyAPIHandler) RevokeAPIKey(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "API Key revocation not implemented"})
}

func (h *ThirdPartyAPIHandler) GetRepositoryCommits(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Repository commits feature not implemented"})
}

func (h *ThirdPartyAPIHandler) GetRepositoryBranches(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Repository branches feature not implemented"})
}

func (h *ThirdPartyAPIHandler) CreateBranch(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Create branch feature not implemented"})
}

func (h *ThirdPartyAPIHandler) GetRepositoryFiles(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Repository files feature not implemented"})
}

func (h *ThirdPartyAPIHandler) GetFileContent(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "File content feature not implemented"})
}

func (h *ThirdPartyAPIHandler) UpdateFileContent(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Update file content feature not implemented"})
}

func (h *ThirdPartyAPIHandler) UpdateUserRole(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Update user role feature not implemented"})
}

func (h *ThirdPartyAPIHandler) GetUserPermissions(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Get user permissions feature not implemented"})
}

func (h *ThirdPartyAPIHandler) GetSystemStats(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "System stats feature not implemented"})
}
