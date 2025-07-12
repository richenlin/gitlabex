package handlers

import (
	"net/http"
	"strconv"

	"gitlabex/internal/services"

	"github.com/gin-gonic/gin"
)

type NotificationHandler struct {
	gitlabService *services.GitLabService
}

func NewNotificationHandler(gitlabService *services.GitLabService) *NotificationHandler {
	return &NotificationHandler{
		gitlabService: gitlabService,
	}
}

type Notification struct {
	ID        int    `json:"id"`
	Type      string `json:"type"`
	Title     string `json:"title"`
	Content   string `json:"content"`
	CreatedAt string `json:"created_at"`
	Read      bool   `json:"read"`
	Priority  string `json:"priority"`
	Category  string `json:"category"`
	Actions   []struct {
		Label  string `json:"label"`
		Action string `json:"action"`
		Type   string `json:"type"`
	} `json:"actions,omitempty"`
}

// GetNotifications 获取通知列表
func (h *NotificationHandler) GetNotifications(c *gin.Context) {
	// 获取查询参数
	filterType := c.Query("type")
	filterRead := c.Query("read")

	// 这里应该从数据库获取真实数据
	// 目前返回模拟数据
	notifications := []Notification{
		{
			ID:        1,
			Type:      "assignment",
			Title:     "作业提醒：数据结构实验",
			Content:   "数据结构实验作业将于明天 23:59 截止，请及时提交。",
			CreatedAt: "2024-03-15 14:30:00",
			Read:      false,
			Priority:  "high",
			Category:  "作业",
			Actions: []struct {
				Label  string `json:"label"`
				Action string `json:"action"`
				Type   string `json:"type"`
			}{
				{Label: "查看作业", Action: "view-assignment", Type: "primary"},
				{Label: "提交作业", Action: "submit-assignment", Type: "success"},
			},
		},
		{
			ID:        2,
			Type:      "project",
			Title:     "项目更新：Web开发项目",
			Content:   "项目 \"Web开发项目\" 有新的提交，请查看最新进展。",
			CreatedAt: "2024-03-15 10:15:00",
			Read:      true,
			Priority:  "medium",
			Category:  "项目",
			Actions: []struct {
				Label  string `json:"label"`
				Action string `json:"action"`
				Type   string `json:"type"`
			}{
				{Label: "查看项目", Action: "view-project", Type: "primary"},
			},
		},
		{
			ID:        3,
			Type:      "system",
			Title:     "系统维护通知",
			Content:   "系统将于今晚 22:00-24:00 进行维护，期间可能影响服务使用。",
			CreatedAt: "2024-03-14 16:45:00",
			Read:      false,
			Priority:  "medium",
			Category:  "系统",
		},
		{
			ID:        4,
			Type:      "reminder",
			Title:     "课程提醒",
			Content:   "明天上午 9:00 有 \"算法分析\" 课程，请准时参加。",
			CreatedAt: "2024-03-14 12:00:00",
			Read:      true,
			Priority:  "low",
			Category:  "提醒",
		},
		{
			ID:        5,
			Type:      "warning",
			Title:     "作业逾期警告",
			Content:   "您有 2 个作业已逾期，请尽快联系教师处理。",
			CreatedAt: "2024-03-13 09:30:00",
			Read:      false,
			Priority:  "high",
			Category:  "警告",
			Actions: []struct {
				Label  string `json:"label"`
				Action string `json:"action"`
				Type   string `json:"type"`
			}{
				{Label: "查看逾期作业", Action: "view-overdue", Type: "warning"},
			},
		},
	}

	// 应用筛选条件
	var filteredNotifications []Notification
	for _, notification := range notifications {
		if filterType != "" && notification.Type != filterType {
			continue
		}
		if filterRead != "" {
			if filterRead == "read" && !notification.Read {
				continue
			}
			if filterRead == "unread" && notification.Read {
				continue
			}
		}
		filteredNotifications = append(filteredNotifications, notification)
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   filteredNotifications,
	})
}

// MarkAsRead 标记通知为已读
func (h *NotificationHandler) MarkAsRead(c *gin.Context) {
	notificationIDStr := c.Param("id")
	_, err := strconv.Atoi(notificationIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid notification ID"})
		return
	}

	// 这里应该更新数据库
	// 目前返回成功响应
	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "通知已标记为已读",
	})
}

// MarkAllAsRead 标记所有通知为已读
func (h *NotificationHandler) MarkAllAsRead(c *gin.Context) {
	// 这里应该更新数据库
	// 目前返回成功响应
	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "所有通知已标记为已读",
	})
}

// DeleteNotification 删除通知
func (h *NotificationHandler) DeleteNotification(c *gin.Context) {
	notificationIDStr := c.Param("id")
	_, err := strconv.Atoi(notificationIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid notification ID"})
		return
	}

	// 这里应该从数据库删除
	// 目前返回成功响应
	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "通知已删除",
	})
}

// DeleteNotifications 批量删除通知
func (h *NotificationHandler) DeleteNotifications(c *gin.Context) {
	var request struct {
		IDs []int `json:"ids"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 这里应该从数据库批量删除
	// 目前返回成功响应
	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "通知已批量删除",
	})
}

// RegisterRoutes 注册路由
func (h *NotificationHandler) RegisterRoutes(rg *gin.RouterGroup) {
	notifications := rg.Group("/notifications")
	{
		notifications.GET("", h.GetNotifications)
		notifications.PUT("/:id/read", h.MarkAsRead)
		notifications.PUT("/read-all", h.MarkAllAsRead)
		notifications.DELETE("/:id", h.DeleteNotification)
		notifications.DELETE("", h.DeleteNotifications)
	}
}
