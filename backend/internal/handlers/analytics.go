package handlers

import (
	"net/http"
	"strconv"
	"time"

	"gitlabex/internal/services"

	"github.com/gin-gonic/gin"
)

type AnalyticsHandler struct {
	analyticsService *services.AnalyticsService
	userService      *services.UserService
}

func NewAnalyticsHandler(analyticsService *services.AnalyticsService, userService *services.UserService) *AnalyticsHandler {
	return &AnalyticsHandler{
		analyticsService: analyticsService,
		userService:      userService,
	}
}

// GetAnalyticsOverview 获取分析概览数据
func (h *AnalyticsHandler) GetAnalyticsOverview(c *gin.Context) {
	// 获取当前用户ID
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}

	// 获取用户信息
	user, err := h.userService.GetUserByID(userID.(uint))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "用户不存在"})
		return
	}

	// 根据用户角色获取不同的概览数据
	var stats interface{}
	switch user.Role {
	case 1: // 管理员
		stats, err = h.analyticsService.GetAdminOverview(user.ID)
	case 2: // 教师
		stats, err = h.analyticsService.GetTeacherOverview(user.ID)
	case 3: // 学生
		stats, err = h.analyticsService.GetStudentOverview(user.ID)
	default:
		c.JSON(http.StatusForbidden, gin.H{"error": "权限不足"})
		return
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取概览数据失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   stats,
	})
}

// GetProjectStats 获取课题统计数据
func (h *AnalyticsHandler) GetProjectStats(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}

	user, err := h.userService.GetUserByID(userID.(uint))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "用户不存在"})
		return
	}

	stats, err := h.analyticsService.GetProjectStats(user.ID, user.Role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取课题统计失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   stats,
	})
}

// GetStudentStats 获取学生统计数据
func (h *AnalyticsHandler) GetStudentStats(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}

	user, err := h.userService.GetUserByID(userID.(uint))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "用户不存在"})
		return
	}

	// 仅教师和管理员可以查看学生统计
	if user.Role != 2 && user.Role != 1 { // 2: 教师, 1: 管理员
		c.JSON(http.StatusForbidden, gin.H{"error": "权限不足"})
		return
	}

	stats, err := h.analyticsService.GetStudentStats(user.ID, user.Role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取学生统计失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   stats,
	})
}

// GetAssignmentStats 获取作业统计数据
func (h *AnalyticsHandler) GetAssignmentStats(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}

	user, err := h.userService.GetUserByID(userID.(uint))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "用户不存在"})
		return
	}

	stats, err := h.analyticsService.GetAssignmentStats(user.ID, user.Role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取作业统计失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   stats,
	})
}

// GetSubmissionTrend 获取提交趋势数据
func (h *AnalyticsHandler) GetSubmissionTrend(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}

	// 获取时间范围参数
	startDateStr := c.Query("start_date")
	endDateStr := c.Query("end_date")

	var startDate, endDate time.Time
	var err error

	if startDateStr != "" {
		startDate, err = time.Parse("2006-01-02", startDateStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "开始日期格式错误"})
			return
		}
	} else {
		startDate = time.Now().AddDate(0, 0, -30) // 默认30天前
	}

	if endDateStr != "" {
		endDate, err = time.Parse("2006-01-02", endDateStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "结束日期格式错误"})
			return
		}
	} else {
		endDate = time.Now() // 默认当前时间
	}

	user, err := h.userService.GetUserByID(userID.(uint))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "用户不存在"})
		return
	}

	trendData, err := h.analyticsService.GetSubmissionTrend(user.ID, user.Role, startDate, endDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取提交趋势失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   trendData,
	})
}

// GetProjectDistribution 获取课题分布数据
func (h *AnalyticsHandler) GetProjectDistribution(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}

	user, err := h.userService.GetUserByID(userID.(uint))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "用户不存在"})
		return
	}

	distributionData, err := h.analyticsService.GetProjectDistribution(user.ID, user.Role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取课题分布失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   distributionData,
	})
}

// GetGradeDistribution 获取成绩分布数据
func (h *AnalyticsHandler) GetGradeDistribution(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}

	user, err := h.userService.GetUserByID(userID.(uint))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "用户不存在"})
		return
	}

	gradeData, err := h.analyticsService.GetGradeDistribution(user.ID, user.Role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取成绩分布失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   gradeData,
	})
}

// GetActivityStats 获取活跃度统计数据
func (h *AnalyticsHandler) GetActivityStats(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}

	user, err := h.userService.GetUserByID(userID.(uint))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "用户不存在"})
		return
	}

	activityData, err := h.analyticsService.GetActivityStats(user.ID, user.Role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取活跃度统计失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   activityData,
	})
}

// GetDashboardStats 获取仪表盘统计数据
func (h *AnalyticsHandler) GetDashboardStats(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}

	user, err := h.userService.GetUserByID(userID.(uint))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "用户不存在"})
		return
	}

	dashboardStats, err := h.analyticsService.GetDashboardStats(user.ID, user.Role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取仪表盘统计失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   dashboardStats,
	})
}

// GetRecentActivities 获取最近活动数据
func (h *AnalyticsHandler) GetRecentActivities(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}

	// 获取限制参数
	limitStr := c.DefaultQuery("limit", "10")
	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		limit = 10
	}

	user, err := h.userService.GetUserByID(userID.(uint))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "用户不存在"})
		return
	}

	activities, err := h.analyticsService.GetRecentActivities(user.ID, user.Role, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取最近活动失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   activities,
	})
}
