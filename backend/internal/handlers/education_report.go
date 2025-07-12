package handlers

import (
	"net/http"

	"gitlabex/internal/services"

	"github.com/gin-gonic/gin"
)

type EducationReportHandler struct {
	gitlabService *services.GitLabService
}

func NewEducationReportHandler(gitlabService *services.GitLabService) *EducationReportHandler {
	return &EducationReportHandler{
		gitlabService: gitlabService,
	}
}

type ReportData struct {
	Overview struct {
		TotalStudents     int     `json:"total_students"`
		TotalAssignments  int     `json:"total_assignments"`
		TotalProjects     int     `json:"total_projects"`
		AverageCompletion float64 `json:"average_completion"`
	} `json:"overview"`
	ClassActivity []struct {
		ClassName      string  `json:"class_name"`
		ActiveStudents int     `json:"active_students"`
		TotalStudents  int     `json:"total_students"`
		CompletionRate float64 `json:"completion_rate"`
	} `json:"class_activity"`
	AssignmentStats []struct {
		Name      string `json:"name"`
		Submitted int    `json:"submitted"`
		Total     int    `json:"total"`
		OnTime    int    `json:"on_time"`
		Late      int    `json:"late"`
	} `json:"assignment_stats"`
	ProgressTrends []struct {
		Date          string `json:"date"`
		Completion    int    `json:"completion"`
		Participation int    `json:"participation"`
	} `json:"progress_trends"`
	TopPerformers []struct {
		ID             int     `json:"id"`
		Name           string  `json:"name"`
		CompletionRate float64 `json:"completion_rate"`
		Points         int     `json:"points"`
	} `json:"top_performers"`
}

// GetEducationReports 获取教育报表数据
func (h *EducationReportHandler) GetEducationReports(c *gin.Context) {
	// 获取查询参数
	timeRange := c.Query("time_range")
	className := c.Query("class")

	// 这里应该根据参数从GitLab API获取真实数据
	// 目前返回模拟数据
	reportData := ReportData{
		Overview: struct {
			TotalStudents     int     `json:"total_students"`
			TotalAssignments  int     `json:"total_assignments"`
			TotalProjects     int     `json:"total_projects"`
			AverageCompletion float64 `json:"average_completion"`
		}{
			TotalStudents:     125,
			TotalAssignments:  45,
			TotalProjects:     8,
			AverageCompletion: 87.5,
		},
		ClassActivity: []struct {
			ClassName      string  `json:"class_name"`
			ActiveStudents int     `json:"active_students"`
			TotalStudents  int     `json:"total_students"`
			CompletionRate float64 `json:"completion_rate"`
		}{
			{ClassName: "计算机科学1班", ActiveStudents: 28, TotalStudents: 30, CompletionRate: 93.3},
			{ClassName: "计算机科学2班", ActiveStudents: 25, TotalStudents: 32, CompletionRate: 78.1},
			{ClassName: "软件工程1班", ActiveStudents: 31, TotalStudents: 33, CompletionRate: 93.9},
			{ClassName: "软件工程2班", ActiveStudents: 27, TotalStudents: 30, CompletionRate: 90.0},
		},
		AssignmentStats: []struct {
			Name      string `json:"name"`
			Submitted int    `json:"submitted"`
			Total     int    `json:"total"`
			OnTime    int    `json:"on_time"`
			Late      int    `json:"late"`
		}{
			{Name: "数据结构实验", Submitted: 28, Total: 30, OnTime: 25, Late: 3},
			{Name: "算法分析作业", Submitted: 32, Total: 33, OnTime: 30, Late: 2},
			{Name: "Web开发项目", Submitted: 25, Total: 30, OnTime: 20, Late: 5},
			{Name: "数据库设计", Submitted: 29, Total: 32, OnTime: 26, Late: 3},
		},
		ProgressTrends: []struct {
			Date          string `json:"date"`
			Completion    int    `json:"completion"`
			Participation int    `json:"participation"`
		}{
			{Date: "2024-03-01", Completion: 75, Participation: 82},
			{Date: "2024-03-08", Completion: 82, Participation: 88},
			{Date: "2024-03-15", Completion: 87, Participation: 91},
			{Date: "2024-03-22", Completion: 89, Participation: 93},
		},
		TopPerformers: []struct {
			ID             int     `json:"id"`
			Name           string  `json:"name"`
			CompletionRate float64 `json:"completion_rate"`
			Points         int     `json:"points"`
		}{
			{ID: 1, Name: "张三", CompletionRate: 98.5, Points: 245},
			{ID: 2, Name: "李四", CompletionRate: 95.2, Points: 238},
			{ID: 3, Name: "王五", CompletionRate: 92.8, Points: 232},
			{ID: 4, Name: "赵六", CompletionRate: 90.5, Points: 226},
			{ID: 5, Name: "钱七", CompletionRate: 88.9, Points: 222},
		},
	}

	// 根据参数调整数据（模拟筛选逻辑）
	if timeRange == "week" {
		// 调整为周数据
		reportData.Overview.TotalStudents = 30
		reportData.Overview.TotalAssignments = 8
	} else if timeRange == "month" {
		// 调整为月数据
		reportData.Overview.TotalStudents = 125
		reportData.Overview.TotalAssignments = 45
	}

	if className != "" && className != "all" {
		// 根据班级筛选数据
		var filteredClassActivity []struct {
			ClassName      string  `json:"class_name"`
			ActiveStudents int     `json:"active_students"`
			TotalStudents  int     `json:"total_students"`
			CompletionRate float64 `json:"completion_rate"`
		}
		for _, class := range reportData.ClassActivity {
			if class.ClassName == className {
				filteredClassActivity = append(filteredClassActivity, class)
			}
		}
		reportData.ClassActivity = filteredClassActivity
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   reportData,
	})
}

// ExportReport 导出报表
func (h *EducationReportHandler) ExportReport(c *gin.Context) {
	// 获取查询参数
	format := c.Query("format")
	timeRange := c.Query("time_range")
	className := c.Query("class")

	// 这里应该实现真实的导出逻辑
	// 目前返回模拟响应
	response := gin.H{
		"status":  "success",
		"message": "报表导出成功",
		"data": gin.H{
			"format":       format,
			"time_range":   timeRange,
			"class":        className,
			"download_url": "/api/reports/download/" + format,
		},
	}

	c.JSON(http.StatusOK, response)
}

// GetClassList 获取班级列表
func (h *EducationReportHandler) GetClassList(c *gin.Context) {
	// 这里应该从GitLab Groups API获取真实数据
	// 目前返回模拟数据
	classes := []gin.H{
		{"value": "all", "label": "所有班级"},
		{"value": "计算机科学1班", "label": "计算机科学1班"},
		{"value": "计算机科学2班", "label": "计算机科学2班"},
		{"value": "软件工程1班", "label": "软件工程1班"},
		{"value": "软件工程2班", "label": "软件工程2班"},
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   classes,
	})
}

// RegisterRoutes 注册路由
func (h *EducationReportHandler) RegisterRoutes(rg *gin.RouterGroup) {
	reports := rg.Group("/education-reports")
	{
		reports.GET("", h.GetEducationReports)
		reports.GET("/classes", h.GetClassList)
		reports.POST("/export", h.ExportReport)
	}
}
