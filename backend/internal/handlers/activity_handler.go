package handlers

import (
	"gitlabex/internal/services"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// ActivityHandler 活动处理器
type ActivityHandler struct {
	activityService *services.ActivityService
}

// NewActivityHandler 创建活动处理器
func NewActivityHandler(activityService *services.ActivityService) *ActivityHandler {
	return &ActivityHandler{
		activityService: activityService,
	}
}

// GetRecentActivities 获取最近活动
func (h *ActivityHandler) GetRecentActivities(c *gin.Context) {
	// 获取限制参数
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	if limit < 1 || limit > 50 {
		limit = 10
	}

	activities, err := h.activityService.GetRecentActivities(limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "获取最近活动失败",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "获取最近活动成功",
		"data":    activities,
		"count":   len(activities),
	})
}

// GetUserActivities 获取用户活动
func (h *ActivityHandler) GetUserActivities(c *gin.Context) {
	// 获取用户ID
	userIDParam := c.Param("userID")
	var userID int64
	var err error

	if userIDParam == "me" {
		// 获取当前用户ID
		currentUserID, exists := c.Get("gitlab_user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "用户未登录"})
			return
		}
		userID = currentUserID.(int64)
	} else {
		// 解析指定用户ID（GitLab用户ID是int64）
		userID, err = strconv.ParseInt(userIDParam, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "无效的用户ID"})
			return
		}
	}

	// 获取限制参数
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	if limit < 1 || limit > 50 {
		limit = 10
	}

	activities, err := h.activityService.GetUserActivities(userID, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "获取用户活动失败",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "获取用户活动成功",
		"data":    activities,
		"count":   len(activities),
	})
}
