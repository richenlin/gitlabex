package middleware

import (
	"gitlabex/internal/models"
	"gitlabex/internal/services"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// RequireProjectAccess 要求项目访问权限
func RequireProjectAccess(gitlabService *services.GitLabService, minRole models.GitLabRole) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 检查用户是否已登录
		_, exists := c.Get("gitlab_access_token")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "用户未登录"})
			c.Abort()
			return
		}

		// 获取项目ID（从路径参数或查询参数）
		projectIDStr := c.Param("project_id")
		if projectIDStr == "" {
			projectIDStr = c.Query("project_id")
		}

		if projectIDStr == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "缺少项目ID"})
			c.Abort()
			return
		}

		projectID, err := strconv.ParseInt(projectIDStr, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "无效的项目ID"})
			c.Abort()
			return
		}

		// 实现GitLab项目权限检查
		accessToken, exists := c.Get("gitlab_access_token")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "缺少访问令牌"})
			c.Abort()
			return
		}

		// 获取用户在项目中的访问级别
		accessLevel, err := gitlabService.GetUserProjectAccessLevel(accessToken.(string), projectID)
		if err != nil {
			c.JSON(http.StatusForbidden, gin.H{"error": "无权限访问该项目", "details": err.Error()})
			c.Abort()
			return
		}

		// 检查是否满足最小权限要求
		minAccessLevel := getMinAccessLevel(minRole)
		if accessLevel < minAccessLevel {
			c.JSON(http.StatusForbidden, gin.H{
				"error":             "权限不足",
				"required_role":     minRole.GetRoleString(),
				"user_access_level": accessLevel,
			})
			c.Abort()
			return
		}

		c.Set("project_id", projectID)
		c.Set("user_access_level", accessLevel)
		c.Next()
	}
}

// RequireGitLabAdmin 要求GitLab管理员权限
func RequireGitLabAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		isAdmin, exists := c.Get("is_admin")
		if !exists || !isAdmin.(bool) {
			c.JSON(http.StatusForbidden, gin.H{"error": "需要管理员权限"})
			c.Abort()
			return
		}
		c.Next()
	}
}

// RequireProjectOwner 要求项目所有者权限
func RequireProjectOwner(gitlabService *services.GitLabService) gin.HandlerFunc {
	return RequireProjectAccess(gitlabService, models.GitLabOwner)
}

// RequireProjectMaintainer 要求项目维护者权限（教师）
func RequireProjectMaintainer(gitlabService *services.GitLabService) gin.HandlerFunc {
	return RequireProjectAccess(gitlabService, models.GitLabMaintainer)
}

// RequireProjectDeveloper 要求项目开发者权限（研究员）
func RequireProjectDeveloper(gitlabService *services.GitLabService) gin.HandlerFunc {
	return RequireProjectAccess(gitlabService, models.GitLabDeveloper)
}

// RequireProjectReporter 要求项目报告者权限（学生）
func RequireProjectReporter(gitlabService *services.GitLabService) gin.HandlerFunc {
	return RequireProjectAccess(gitlabService, models.GitLabReporter)
}

// getMinAccessLevel 根据GitLab角色获取最小访问级别
func getMinAccessLevel(role models.GitLabRole) int {
	switch role {
	case models.GitLabGuest:
		return 10
	case models.GitLabReporter:
		return 20
	case models.GitLabDeveloper:
		return 30
	case models.GitLabMaintainer:
		return 40
	case models.GitLabOwner:
		return 50
	default:
		return 10 // 默认为Guest级别
	}
}
