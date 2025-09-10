package middleware

import (
	"gitlabex/internal/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

// RequireRole 要求特定角色
func RequireRole(minRole models.UserRole) gin.HandlerFunc {
	return func(c *gin.Context) {
		_, exists := c.Get("userID")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "用户未登录"})
			c.Abort()
			return
		}

		// 这里应该从数据库获取用户角色
		// 简化实现：直接从上下文中获取角色
		role, exists := c.Get("role")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "无法获取用户角色"})
			c.Abort()
			return
		}

		userRole := role.(models.UserRole)
		if userRole != minRole {
			c.JSON(http.StatusForbidden, gin.H{"error": "权限不足"})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireAdmin 要求管理员权限
func RequireAdmin() gin.HandlerFunc {
	return RequireRole(models.RoleAdmin)
}

// RequireTeacher 要求教师权限
func RequireTeacher() gin.HandlerFunc {
	return RequireRole(models.RoleTeacher)
}

// RequireStudent 要求学生权限
func RequireStudent() gin.HandlerFunc {
	return RequireRole(models.RoleStudent)
}