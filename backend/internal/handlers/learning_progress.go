package handlers

import (
	"net/http"
	"strconv"

	"gitlabex/internal/services"

	"github.com/gin-gonic/gin"
)

type LearningProgressHandler struct {
	gitlabService *services.GitLabService
}

func NewLearningProgressHandler(gitlabService *services.GitLabService) *LearningProgressHandler {
	return &LearningProgressHandler{
		gitlabService: gitlabService,
	}
}

type LearningProgressData struct {
	Overview struct {
		TotalAssignments     int `json:"total_assignments"`
		CompletedAssignments int `json:"completed_assignments"`
		TotalProjects        int `json:"total_projects"`
		ActiveProjects       int `json:"active_projects"`
		TotalCommits         int `json:"total_commits"`
		TotalMergeRequests   int `json:"total_merge_requests"`
	} `json:"overview"`
	RecentActivity []struct {
		ID     int    `json:"id"`
		Type   string `json:"type"`
		Title  string `json:"title"`
		Time   string `json:"time"`
		Status string `json:"status"`
	} `json:"recent_activity"`
	ProgressChart []struct {
		Date      string `json:"date"`
		Completed int    `json:"completed"`
		Total     int    `json:"total"`
	} `json:"progress_chart"`
	Achievements []struct {
		Title       string `json:"title"`
		Description string `json:"description"`
		Icon        string `json:"icon"`
		Earned      bool   `json:"earned"`
	} `json:"achievements"`
}

// GetLearningProgress 获取学习进度数据
func (h *LearningProgressHandler) GetLearningProgress(c *gin.Context) {
	userIDStr := c.Param("user_id")
	_, err := strconv.Atoi(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	// 这里应该从GitLab API获取真实数据
	// 目前返回模拟数据
	progressData := LearningProgressData{
		Overview: struct {
			TotalAssignments     int `json:"total_assignments"`
			CompletedAssignments int `json:"completed_assignments"`
			TotalProjects        int `json:"total_projects"`
			ActiveProjects       int `json:"active_projects"`
			TotalCommits         int `json:"total_commits"`
			TotalMergeRequests   int `json:"total_merge_requests"`
		}{
			TotalAssignments:     12,
			CompletedAssignments: 8,
			TotalProjects:        3,
			ActiveProjects:       2,
			TotalCommits:         45,
			TotalMergeRequests:   12,
		},
		RecentActivity: []struct {
			ID     int    `json:"id"`
			Type   string `json:"type"`
			Title  string `json:"title"`
			Time   string `json:"time"`
			Status string `json:"status"`
		}{
			{
				ID:     1,
				Type:   "assignment",
				Title:  "完成了作业：数据结构实验",
				Time:   "2024-03-15 14:30:00",
				Status: "completed",
			},
			{
				ID:     2,
				Type:   "commit",
				Title:  "提交了代码：优化算法性能",
				Time:   "2024-03-15 10:15:00",
				Status: "success",
			},
			{
				ID:     3,
				Type:   "merge_request",
				Title:  "提交了合并请求：修复bug",
				Time:   "2024-03-14 16:45:00",
				Status: "pending",
			},
		},
		ProgressChart: []struct {
			Date      string `json:"date"`
			Completed int    `json:"completed"`
			Total     int    `json:"total"`
		}{
			{Date: "2024-03-01", Completed: 2, Total: 3},
			{Date: "2024-03-08", Completed: 4, Total: 6},
			{Date: "2024-03-15", Completed: 8, Total: 12},
		},
		Achievements: []struct {
			Title       string `json:"title"`
			Description string `json:"description"`
			Icon        string `json:"icon"`
			Earned      bool   `json:"earned"`
		}{
			{Title: "连续提交", Description: "连续7天提交代码", Icon: "Trophy", Earned: true},
			{Title: "作业达人", Description: "完成10个作业", Icon: "Star", Earned: true},
			{Title: "团队合作", Description: "参与3个项目", Icon: "User", Earned: false},
		},
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   progressData,
	})
}

// GetUserList 获取用户列表（用于选择学生）
func (h *LearningProgressHandler) GetUserList(c *gin.Context) {
	// 这里应该从GitLab API获取真实的用户数据
	// 目前返回模拟数据
	users := []gin.H{
		{"id": 1, "name": "张三", "username": "zhangsan", "avatar": ""},
		{"id": 2, "name": "李四", "username": "lisi", "avatar": ""},
		{"id": 3, "name": "王五", "username": "wangwu", "avatar": ""},
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   users,
	})
}

// RegisterRoutes 注册路由
func (h *LearningProgressHandler) RegisterRoutes(rg *gin.RouterGroup) {
	learningProgress := rg.Group("/learning-progress")
	{
		learningProgress.GET("/users", h.GetUserList)
		learningProgress.GET("/user/:user_id", h.GetLearningProgress)
	}
}
