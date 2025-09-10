package handlers

import (
	"gitlabex/internal/services"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"gitlabex/internal/models"
)

// NotificationHandler 通知处理器
type NotificationHandler struct {
	notificationService *services.NotificationService
	userService         *services.UserService
}

// NewNotificationHandler 创建通知处理器
func NewNotificationHandler(notificationService *services.NotificationService, userService *services.UserService) *NotificationHandler {
	return &NotificationHandler{
		notificationService: notificationService,
		userService:         userService,
	}
}

// GetNotifications 获取用户通知列表
func (h *NotificationHandler) GetNotifications(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户未登录"})
		return
	}

	// 获取分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	offset := (page - 1) * limit

	notifications, total, err := h.notificationService.GetUserNotifications(
		userID.(uuid.UUID),
		limit,
		offset,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取通知失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"notifications": notifications,
		"pagination": gin.H{
			"page":  page,
			"limit": limit,
			"total": total,
			"pages": (total + int64(limit) - 1) / int64(limit),
		},
	})
}

// GetUnreadCount 获取未读通知数量
func (h *NotificationHandler) GetUnreadCount(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户未登录"})
		return
	}

	count, err := h.notificationService.GetUnreadNotificationsCount(userID.(uuid.UUID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取未读通知数量失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"count": count})
}

// MarkAsRead 标记通知为已读
func (h *NotificationHandler) MarkAsRead(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户未登录"})
		return
	}

	notificationIDStr := c.Param("id")
	notificationID, err := uuid.Parse(notificationIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的通知ID"})
		return
	}

	if err := h.notificationService.MarkAsRead(notificationID, userID.(uuid.UUID)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "标记通知已读失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "已标记为已读"})
}

// MarkAllAsRead 标记所有通知为已读
func (h *NotificationHandler) MarkAllAsRead(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户未登录"})
		return
	}

	if err := h.notificationService.MarkAllAsRead(userID.(uuid.UUID)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "标记所有通知已读失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "所有通知已标记为已读"})
}

// CreateAnnouncement 创建公告
func (h *NotificationHandler) CreateAnnouncement(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户未登录"})
		return
	}

	var req struct {
		Title       string      `json:"title" binding:"required"`
		Content     string      `json:"content" binding:"required"`
		Priority    string      `json:"priority"`
		ValidFrom   *time.Time  `json:"valid_from"`
		ValidTo     *time.Time  `json:"valid_to"`
		TargetRoles []string    `json:"target_roles"`
		TargetUsers []uuid.UUID `json:"target_users"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 获取当前用户信息
	currentUser, err := h.userService.GetUserByID(userID.(uuid.UUID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取用户信息失败"})
		return
	}

	// 检查权限
	if currentUser.EduRole < models.EduRoleTeacher {
		c.JSON(http.StatusForbidden, gin.H{"error": "无权限创建公告"})
		return
	}

	// 如果没有指定目标用户，根据目标角色获取用户
	var targetUserIDs []uuid.UUID
	if len(req.TargetUsers) > 0 {
		targetUserIDs = req.TargetUsers
	} else {
		// 获取目标角色的用户
		var roles []models.UserRole
		for _, roleStr := range req.TargetRoles {
			roles = append(roles, models.UserRole(roleStr))
		}
		users, err := h.userService.GetUsersByRoles(roles)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "获取目标用户失败"})
			return
		}
		for _, user := range users {
			targetUserIDs = append(targetUserIDs, user.ID)
		}
	}

	announcement := &models.Announcement{
		Title:       req.Title,
		Content:     req.Content,
		AuthorID:    userID.(uuid.UUID),
		Priority:    req.Priority,
		ValidFrom:   time.Now(),
		ValidTo:     req.ValidTo,
		TargetRoles: req.TargetRoles,
	}

	if req.ValidFrom != nil {
		announcement.ValidFrom = *req.ValidFrom
	}

	if err := h.notificationService.CreateAnnouncement(announcement, targetUserIDs); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建公告失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "公告创建成功", "id": announcement.ID})
}

// GetAnnouncements 获取公告列表
func (h *NotificationHandler) GetAnnouncements(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	offset := (page - 1) * limit

	var announcements []models.Announcement
	var total int64

	err := h.notificationService.GetAnnouncementsWithPagination(&announcements, &total, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取公告失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"announcements": announcements,
		"pagination": gin.H{
			"page":  page,
			"limit": limit,
			"total": total,
			"pages": (total + int64(limit) - 1) / int64(limit),
		},
	})
}
