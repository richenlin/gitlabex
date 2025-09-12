package middleware

import (
	"gitlabex/internal/services"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// PermissionMiddleware 权限检查中间件
type PermissionMiddleware struct {
	permissionService *services.PermissionService
}

// NewPermissionMiddleware 创建权限中间件
func NewPermissionMiddleware(permissionService *services.PermissionService) *PermissionMiddleware {
	return &PermissionMiddleware{
		permissionService: permissionService,
	}
}

// RequireProjectPermission 需要项目权限的中间件
func (m *PermissionMiddleware) RequireProjectPermission(permission services.ProjectPermission) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("userID")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "用户未登录"})
			c.Abort()
			return
		}

		// 从路径参数或查询参数获取项目ID
		var projectID uuid.UUID
		var err error

		if projectIDStr := c.Param("project_id"); projectIDStr != "" {
			projectID, err = uuid.Parse(projectIDStr)
		} else if projectIDStr := c.Param("id"); projectIDStr != "" {
			// 尝试从 id 参数获取项目ID（用于 /research-projects/:id 路由）
			projectID, err = uuid.Parse(projectIDStr)
		} else if projectIDStr := c.Query("project_id"); projectIDStr != "" {
			projectID, err = uuid.Parse(projectIDStr)
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"error": "项目ID不能为空"})
			c.Abort()
			return
		}

		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "无效的项目ID"})
			c.Abort()
			return
		}

		// 检查权限
		hasPermission, err := m.permissionService.CheckProjectPermission(
			userID.(uuid.UUID),
			projectID,
			permission,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "检查权限失败"})
			c.Abort()
			return
		}

		if !hasPermission {
			c.JSON(http.StatusForbidden, gin.H{"error": "无权限访问"})
			c.Abort()
			return
		}

		// 将项目ID保存到上下文中
		c.Set("projectID", projectID)
		c.Next()
	}
}

// RequireDocumentPermission 需要文档权限的中间件
func (m *PermissionMiddleware) RequireDocumentPermission(permission services.ProjectPermission) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("userID")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "用户未登录"})
			c.Abort()
			return
		}

		documentIDStr := c.Param("id")
		documentID, err := uuid.Parse(documentIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "无效的文档ID"})
			c.Abort()
			return
		}

		// 检查权限
		hasPermission, err := m.permissionService.CheckDocumentPermission(
			userID.(uuid.UUID),
			documentID,
			permission,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "检查权限失败"})
			c.Abort()
			return
		}

		if !hasPermission {
			c.JSON(http.StatusForbidden, gin.H{"error": "无权限访问"})
			c.Abort()
			return
		}

		// 将文档ID保存到上下文中
		c.Set("documentID", documentID)
		c.Next()
	}
}

// RequireTopicPermission 需要话题权限的中间件
func (m *PermissionMiddleware) RequireTopicPermission(permission services.ProjectPermission) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("userID")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "用户未登录"})
			c.Abort()
			return
		}

		topicIDStr := c.Param("id")
		topicID, err := uuid.Parse(topicIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "无效的话题ID"})
			c.Abort()
			return
		}

		// 检查权限
		hasPermission, err := m.permissionService.CheckTopicPermission(
			userID.(uuid.UUID),
			topicID,
			permission,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "检查权限失败"})
			c.Abort()
			return
		}

		if !hasPermission {
			c.JSON(http.StatusForbidden, gin.H{"error": "无权限访问"})
			c.Abort()
			return
		}

		// 将话题ID保存到上下文中
		c.Set("topicID", topicID)
		c.Next()
	}
}

// RequireHomeworkPermission 需要作业权限的中间件
func (m *PermissionMiddleware) RequireHomeworkPermission(permission services.ProjectPermission) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("userID")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "用户未登录"})
			c.Abort()
			return
		}

		homeworkIDStr := c.Param("id")
		homeworkID, err := uuid.Parse(homeworkIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "无效的作业ID"})
			c.Abort()
			return
		}

		// 检查权限
		hasPermission, err := m.permissionService.CheckHomeworkPermission(
			userID.(uuid.UUID),
			homeworkID,
			permission,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "检查权限失败"})
			c.Abort()
			return
		}

		if !hasPermission {
			c.JSON(http.StatusForbidden, gin.H{"error": "无权限访问"})
			c.Abort()
			return
		}

		// 将作业ID保存到上下文中
		c.Set("homeworkID", homeworkID)
		c.Next()
	}
}

// RequireAnyProjectRole 需要任意项目角色的中间件
func (m *PermissionMiddleware) RequireAnyProjectRole() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("userID")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "用户未登录"})
			c.Abort()
			return
		}

		projectIDStr := c.Param("project_id")
		if projectIDStr == "" {
			projectIDStr = c.Query("project_id")
		}
		if projectIDStr == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "项目ID不能为空"})
			c.Abort()
			return
		}

		projectID, err := uuid.Parse(projectIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "无效的项目ID"})
			c.Abort()
			return
		}

		// 检查用户是否是项目成员或管理员
		role, err := m.permissionService.GetUserProjectRole(userID.(uuid.UUID), projectID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "检查权限失败"})
			c.Abort()
			return
		}

		c.Set("projectID", projectID)
		c.Set("userRole", role)
		c.Next()
	}
}
