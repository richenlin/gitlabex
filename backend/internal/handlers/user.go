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

	dashboard, err := h.userService.GetUserDashboard(currentUser.ID)
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
	profile := h.userService.ToProfile(user)

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
		profiles[i] = h.userService.ToProfile(&user)
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  profiles,
		"total": len(profiles),
	})
}

// GetUsers 获取用户列表（管理员权限）
func (h *UserHandler) GetUsers(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	users, total, err := h.userService.GetAllUsers(page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "获取用户列表失败",
			"details": err.Error(),
		})
		return
	}

	// 转换为用户资料
	profiles := make([]*services.UserProfile, len(users))
	for i, userWithRole := range users {
		profiles[i] = h.userService.ToProfile(userWithRole.User)
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  profiles,
		"total": total,
	})
}

// SearchUsers 搜索用户
func (h *UserHandler) SearchUsers(c *gin.Context) {
	keyword := c.Query("keyword")
	roleStr := c.Query("role")
	activeStr := c.Query("active")

	var role *int
	if roleStr != "" {
		if r, err := strconv.Atoi(roleStr); err == nil {
			role = &r
		}
	}

	var active *bool
	if activeStr != "" {
		if a, err := strconv.ParseBool(activeStr); err == nil {
			active = &a
		}
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	users, total, err := h.userService.SearchUsers(keyword, role, active, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "搜索用户失败",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  users,
		"total": total,
	})
}

// UpdateUser 更新用户信息
func (h *UserHandler) UpdateUser(c *gin.Context) {
	userID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "无效的用户ID",
		})
		return
	}

	user, err := h.userService.GetUserByID(uint(userID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "用户不存在",
		})
		return
	}

	var updates map[string]interface{}
	if err := c.ShouldBindJSON(&updates); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "请求参数错误",
			"details": err.Error(),
		})
		return
	}

	if err := h.userService.UpdateUserFields(user.ID, updates); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "更新用户信息失败",
			"details": err.Error(),
		})
		return
	}

	// 获取更新后的用户信息
	updatedUser, err := h.userService.GetUserByID(uint(userID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "获取更新后的用户信息失败",
		})
		return
	}

	profile := h.userService.ToProfile(updatedUser)

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
