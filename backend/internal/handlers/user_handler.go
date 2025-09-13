package handlers

import (
	"gitlabex/internal/services"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// UserHandler 用户处理器
type UserHandler struct {
	userService *services.UserService
}

// NewUserHandler 创建用户处理器
func NewUserHandler(userService *services.UserService) *UserHandler {
	return &UserHandler{
		userService: userService,
	}
}

// GetCurrentUser 获取当前用户信息
func (h *UserHandler) GetCurrentUser(c *gin.Context) {
	accessToken, exists := c.Get("gitlab_access_token")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户未登录"})
		return
	}

	user, err := h.userService.GetCurrentUser(accessToken.(string))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "获取用户信息失败", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, user)
}

// UpdateCurrentUser 更新当前用户信息 - 重定向到GitLab
func (h *UserHandler) UpdateCurrentUser(c *gin.Context) {
	// 用户信息更新应该在GitLab中进行
	c.JSON(http.StatusOK, gin.H{
		"message":     "用户信息更新请前往GitLab",
		"gitlab_url":  "/profile", // GitLab个人资料页面
		"redirect":    true,
	})
}

// GetUsers 获取用户列表
func (h *UserHandler) GetUsers(c *gin.Context) {
	accessToken, exists := c.Get("gitlab_access_token")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户未登录"})
		return
	}

	// 获取项目ID（如果提供）
	projectIDStr := c.Query("project_id")
	if projectIDStr != "" {
		projectID, err := strconv.ParseInt(projectIDStr, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "无效的项目ID"})
			return
		}
		
		// 获取项目成员列表
		users, err := h.userService.GetProjectMembers(accessToken.(string), projectID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "获取项目成员失败", "details": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"users": users})
		return
	}

	// 如果没有项目ID，返回当前用户信息（GitLab API不支持获取所有用户列表）
	user, err := h.userService.GetCurrentUser(accessToken.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取用户信息失败", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"users": []interface{}{user},
		"message": "GitLab API限制，只能获取当前用户或项目成员信息",
	})
}

// GetUserByID 根据GitLab用户ID获取用户信息
func (h *UserHandler) GetUserByID(c *gin.Context) {
	accessToken, exists := c.Get("gitlab_access_token")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户未登录"})
		return
	}

	userIDStr := c.Param("id")
	if userIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "用户ID不能为空"})
		return
	}

	// 如果请求的是当前用户信息，直接返回
	if userIDStr == "me" {
		user, err := h.userService.GetCurrentUser(accessToken.(string))
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "获取用户信息失败", "details": err.Error()})
			return
		}
		c.JSON(http.StatusOK, user)
		return
	}

	// 根据GitLab用户名获取用户信息
	user, err := h.userService.GetUserByUsername(accessToken.(string), userIDStr)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "用户不存在", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, user)
}
