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
		"message":    "用户信息更新请前往GitLab",
		"gitlab_url": "/profile", // GitLab个人资料页面
		"redirect":   true,
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
		"users":   []interface{}{user},
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

// GetUserPersonalStats 获取用户个人统计信息
func (h *UserHandler) GetUserPersonalStats(c *gin.Context) {
	userID, exists := c.Get("gitlab_user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户未登录"})
		return
	}

	stats, err := h.userService.GetUserPersonalStats(userID.(int64))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "获取用户统计失败",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// GetSSHKeys 获取用户SSH密钥列表
func (h *UserHandler) GetSSHKeys(c *gin.Context) {
	accessToken, exists := c.Get("gitlab_access_token")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户未登录"})
		return
	}

	keys, err := h.userService.GetSSHKeys(accessToken.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "获取SSH密钥失败",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, keys)
}

// AddSSHKey 添加SSH密钥
func (h *UserHandler) AddSSHKey(c *gin.Context) {
	accessToken, exists := c.Get("gitlab_access_token")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户未登录"})
		return
	}

	var req struct {
		Title string `json:"title" binding:"required"`
		Key   string `json:"key" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误", "details": err.Error()})
		return
	}

	key, err := h.userService.AddSSHKey(accessToken.(string), req.Title, req.Key)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "添加SSH密钥失败",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, key)
}

// DeleteSSHKey 删除SSH密钥
func (h *UserHandler) DeleteSSHKey(c *gin.Context) {
	accessToken, exists := c.Get("gitlab_access_token")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户未登录"})
		return
	}

	keyIDStr := c.Param("id")
	keyID, err := strconv.Atoi(keyIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的密钥ID"})
		return
	}

	err = h.userService.DeleteSSHKey(accessToken.(string), keyID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "删除SSH密钥失败",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "SSH密钥删除成功"})
}

// ChangePassword 修改密码
func (h *UserHandler) ChangePassword(c *gin.Context) {
	accessToken, exists := c.Get("gitlab_access_token")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户未登录"})
		return
	}

	var req struct {
		CurrentPassword string `json:"currentPassword" binding:"required"`
		NewPassword     string `json:"newPassword" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误", "details": err.Error()})
		return
	}

	err := h.userService.ChangePassword(accessToken.(string), req.CurrentPassword, req.NewPassword)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "修改密码失败",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "密码修改成功"})
}

// GetNotifications 获取用户通知列表
func (h *UserHandler) GetNotifications(c *gin.Context) {
	accessToken, exists := c.Get("gitlab_access_token")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户未登录"})
		return
	}

	// 获取分页参数
	page := 1
	perPage := 20
	if pageStr := c.Query("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}
	if perPageStr := c.Query("per_page"); perPageStr != "" {
		if pp, err := strconv.Atoi(perPageStr); err == nil && pp > 0 && pp <= 100 {
			perPage = pp
		}
	}

	notifications, err := h.userService.GetNotifications(accessToken.(string), page, perPage)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "获取通知列表失败",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"notifications": notifications,
		"page":          page,
		"per_page":      perPage,
	})
}

// MarkNotificationAsRead 标记通知为已读
func (h *UserHandler) MarkNotificationAsRead(c *gin.Context) {
	accessToken, exists := c.Get("gitlab_access_token")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户未登录"})
		return
	}

	notificationID := c.Param("id")
	if notificationID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "通知ID不能为空"})
		return
	}

	err := h.userService.MarkNotificationAsRead(accessToken.(string), notificationID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "标记通知为已读失败",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "通知已标记为已读"})
}

// MarkAllNotificationsAsRead 标记所有通知为已读
func (h *UserHandler) MarkAllNotificationsAsRead(c *gin.Context) {
	accessToken, exists := c.Get("gitlab_access_token")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户未登录"})
		return
	}

	err := h.userService.MarkAllNotificationsAsRead(accessToken.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "标记所有通知为已读失败",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "所有通知已标记为已读"})
}
