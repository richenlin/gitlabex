package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"gitlabex/internal/models"
	"gitlabex/internal/services"
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
	// 从上下文中获取用户信息（由中间件设置）
	user, exists := c.Get("current_user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "未登录",
		})
		return
	}

	currentUser := user.(*models.User)

	// 获取用户详细信息
	profile, err := h.userService.GetUserProfile(currentUser.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "获取用户信息失败",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": profile,
	})
}

// GetUserDashboard 获取用户仪表板
func (h *UserHandler) GetUserDashboard(c *gin.Context) {
	user, exists := c.Get("current_user")
	var currentUser *models.User
	var err error

	if !exists {
		// 如果没有认证中间件，使用测试用户
		currentUser, err = h.userService.GetCurrentUser()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "获取用户信息失败",
				"details": err.Error(),
			})
			return
		}
	} else {
		currentUser = user.(*models.User)
	}

	dashboard, err := h.userService.GetUserDashboard(currentUser)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "获取仪表板数据失败",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": dashboard,
	})
}

// GetUserByID 根据ID获取用户信息
func (h *UserHandler) GetUserByID(c *gin.Context) {
	idParam := c.Param("id")
	userID, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "无效的用户ID",
		})
		return
	}

	user, err := h.userService.GetUserByID(uint(userID))
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "用户不存在",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "获取用户信息失败",
			"details": err.Error(),
		})
		return
	}

	// 获取默认角色
	role := user.GetDefaultEducationRole()
	profile := h.userService.ToProfile(user, role)

	c.JSON(http.StatusOK, gin.H{
		"data": profile,
	})
}

// ListActiveUsers 获取活跃用户列表
func (h *UserHandler) ListActiveUsers(c *gin.Context) {
	users, err := h.userService.ListActiveUsers()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "获取用户列表失败",
			"details": err.Error(),
		})
		return
	}

	// 转换为用户资料列表
	profiles := make([]*services.UserProfile, len(users))
	for i, user := range users {
		role := user.GetDefaultEducationRole()
		profiles[i] = h.userService.ToProfile(user, role)
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  profiles,
		"total": len(profiles),
	})
}

// UpdateUser 更新用户信息
func (h *UserHandler) UpdateUser(c *gin.Context) {
	currentUser, exists := c.Get("current_user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "未登录",
		})
		return
	}

	user := currentUser.(*models.User)

	var updateData struct {
		Name   string `json:"name" binding:"omitempty,min=1,max=255"`
		Avatar string `json:"avatar" binding:"omitempty,url"`
	}

	if err := c.ShouldBindJSON(&updateData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "无效的请求数据",
			"details": err.Error(),
		})
		return
	}

	// 构建更新字段
	updates := make(map[string]interface{})
	if updateData.Name != "" {
		updates["name"] = updateData.Name
	}
	if updateData.Avatar != "" {
		updates["avatar"] = updateData.Avatar
	}

	if len(updates) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "没有要更新的字段",
		})
		return
	}

	err := h.userService.UpdateUserFields(user, updates)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "更新用户信息失败",
			"details": err.Error(),
		})
		return
	}

	// 重新获取更新后的用户信息
	updatedUser, err := h.userService.GetUserByID(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "获取更新后的用户信息失败",
		})
		return
	}

	role := updatedUser.GetDefaultEducationRole()
	profile := h.userService.ToProfile(updatedUser, role)

	c.JSON(http.StatusOK, gin.H{
		"message": "用户信息更新成功",
		"data":    profile,
	})
}

// SyncUserFromGitLab 手动同步用户信息（管理员功能）
func (h *UserHandler) SyncUserFromGitLab(c *gin.Context) {
	// 这个接口需要管理员权限，暂时跳过权限检查
	// TODO: 添加权限检查中间件

	idParam := c.Param("gitlab_id")
	gitlabID, err := strconv.Atoi(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "无效的GitLab用户ID",
		})
		return
	}

	// 这里需要GitLab API支持，暂时返回占位符响应
	c.JSON(http.StatusOK, gin.H{
		"message":   "用户同步功能待实现",
		"gitlab_id": gitlabID,
	})
}

// HealthCheck 健康检查
func (h *UserHandler) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "ok",
		"service": "user-service",
		"timestamp": gin.H{
			"unix": gin.H{
				"timestamp": 1625097600,
			},
		},
	})
}
