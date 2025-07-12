package handlers

import (
	"net/http"
	"strconv"

	"gitlabex/internal/services"

	"github.com/gin-gonic/gin"
)

// NotificationHandler 通知管理处理器
type NotificationHandler struct {
	notificationService *services.NotificationService
	userService         *services.UserService
}

// NewNotificationHandler 创建通知管理处理器
func NewNotificationHandler(notificationService *services.NotificationService, userService *services.UserService) *NotificationHandler {
	return &NotificationHandler{
		notificationService: notificationService,
		userService:         userService,
	}
}

// RegisterRoutes 注册通知管理路由
func (h *NotificationHandler) RegisterRoutes(router *gin.RouterGroup) {
	notifications := router.Group("/notifications")
	{
		notifications.GET("", h.GetNotifications)                   // 获取通知列表
		notifications.GET("/unread", h.GetUnreadNotifications)      // 获取未读通知
		notifications.GET("/count", h.GetUnreadCount)               // 获取未读数量
		notifications.GET("/stats", h.GetNotificationStats)         // 获取通知统计
		notifications.PUT("/:id/read", h.MarkAsRead)                // 标记通知已读
		notifications.PUT("/read-all", h.MarkAllAsRead)             // 标记全部已读
		notifications.DELETE("/:id", h.DeleteNotification)          // 删除通知
		notifications.DELETE("/all", h.DeleteAllNotifications)      // 删除全部通知
		notifications.POST("", h.CreateNotification)                // 创建通知（管理员）
		notifications.GET("/types/:type", h.GetNotificationsByType) // 按类型获取通知
	}
}

// GetNotifications 获取通知列表
func (h *NotificationHandler) GetNotifications(c *gin.Context) {
	// TODO: 从JWT获取用户ID
	userID := uint(1)

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	notifications, total, err := h.notificationService.GetNotificationsByUser(userID, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "获取通知列表失败",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":      notifications,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// GetUnreadNotifications 获取未读通知
func (h *NotificationHandler) GetUnreadNotifications(c *gin.Context) {
	// TODO: 从JWT获取用户ID
	userID := uint(1)

	notifications, err := h.notificationService.GetUnreadNotifications(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "获取未读通知失败",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  notifications,
		"total": len(notifications),
	})
}

// GetUnreadCount 获取未读通知数量
func (h *NotificationHandler) GetUnreadCount(c *gin.Context) {
	// TODO: 从JWT获取用户ID
	userID := uint(1)

	count, err := h.notificationService.GetUnreadCount(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "获取未读数量失败",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"unread_count": count,
	})
}

// GetNotificationStats 获取通知统计信息
func (h *NotificationHandler) GetNotificationStats(c *gin.Context) {
	// TODO: 从JWT获取用户ID
	userID := uint(1)

	stats, err := h.notificationService.GetNotificationStats(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "获取通知统计失败",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": stats,
	})
}

// MarkAsRead 标记通知为已读
func (h *NotificationHandler) MarkAsRead(c *gin.Context) {
	// TODO: 从JWT获取用户ID
	userID := uint(1)

	notificationID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "无效的通知ID",
		})
		return
	}

	if err := h.notificationService.MarkAsRead(uint(notificationID), userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "标记已读失败",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "通知已标记为已读",
	})
}

// MarkAllAsRead 标记所有通知为已读
func (h *NotificationHandler) MarkAllAsRead(c *gin.Context) {
	// TODO: 从JWT获取用户ID
	userID := uint(1)

	if err := h.notificationService.MarkAllAsRead(userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "标记全部已读失败",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "所有通知已标记为已读",
	})
}

// DeleteNotification 删除通知
func (h *NotificationHandler) DeleteNotification(c *gin.Context) {
	// TODO: 从JWT获取用户ID
	userID := uint(1)

	notificationID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "无效的通知ID",
		})
		return
	}

	if err := h.notificationService.DeleteNotification(uint(notificationID), userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "删除通知失败",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "通知已删除",
	})
}

// DeleteAllNotifications 删除所有通知
func (h *NotificationHandler) DeleteAllNotifications(c *gin.Context) {
	// TODO: 从JWT获取用户ID
	userID := uint(1)

	if err := h.notificationService.DeleteAllNotifications(userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "删除全部通知失败",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "所有通知已删除",
	})
}

// CreateNotification 创建通知（管理员功能）
func (h *NotificationHandler) CreateNotification(c *gin.Context) {
	var req services.CreateNotificationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "无效的请求数据",
			"details": err.Error(),
		})
		return
	}

	notification, err := h.notificationService.CreateNotificationFromRequest(&req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "创建通知失败",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "通知创建成功",
		"data":    notification,
	})
}

// GetNotificationsByType 按类型获取通知
func (h *NotificationHandler) GetNotificationsByType(c *gin.Context) {
	// TODO: 从JWT获取用户ID
	userID := uint(1)

	notificationType := c.Param("type")
	if notificationType == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "通知类型不能为空",
		})
		return
	}

	notifications, err := h.notificationService.GetNotificationsByType(userID, notificationType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "获取通知失败",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  notifications,
		"total": len(notifications),
		"type":  notificationType,
	})
}
