package handlers

import (
	"net/http"
	"strconv"

	"gitlabex/internal/services"

	"github.com/gin-gonic/gin"
)

// AssignmentHandler 作业管理处理器
type AssignmentHandler struct {
	assignmentService *services.AssignmentService
	userService       *services.UserService
}

// NewAssignmentHandler 创建作业管理处理器
func NewAssignmentHandler(assignmentService *services.AssignmentService, userService *services.UserService) *AssignmentHandler {
	return &AssignmentHandler{
		assignmentService: assignmentService,
		userService:       userService,
	}
}

// RegisterRoutes 注册作业管理路由
func (h *AssignmentHandler) RegisterRoutes(router *gin.RouterGroup) {
	assignments := router.Group("/assignments")
	{
		// 作业基本操作
		assignments.POST("", h.CreateAssignment)       // 创建作业（老师）
		assignments.GET("", h.ListAssignments)         // 获取作业列表
		assignments.GET("/:id", h.GetAssignment)       // 获取作业详情
		assignments.PUT("/:id", h.UpdateAssignment)    // 更新作业信息（老师）
		assignments.DELETE("/:id", h.DeleteAssignment) // 删除作业（老师）

		// 作业提交和评审
		assignments.POST("/:id/submit", h.SubmitAssignment)                   // 提交作业（学生）
		assignments.GET("/:id/submissions", h.GetSubmissions)                 // 获取作业提交列表
		assignments.GET("/submissions/:submission_id", h.GetSubmissionDetail) // 获取提交详情

		// 统计和分析
		assignments.GET("/:id/stats", h.GetAssignmentStats)    // 获取作业统计
		assignments.GET("/my-submissions", h.GetMySubmissions) // 获取我的提交记录
	}
}

// CreateAssignment 创建作业
func (h *AssignmentHandler) CreateAssignment(c *gin.Context) {
	// TODO: 从JWT获取用户ID，这里暂时使用测试用户
	teacherID := uint(1)

	var req services.CreateAssignmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "无效的请求数据",
			"details": err.Error(),
		})
		return
	}

	assignment, err := h.assignmentService.CreateAssignment(teacherID, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "创建作业失败",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "作业创建成功",
		"data":    assignment,
	})
}

// ListAssignments 获取作业列表
func (h *AssignmentHandler) ListAssignments(c *gin.Context) {
	// TODO: 从JWT获取用户信息
	userID := uint(1)
	userRole := 2 // 假设是老师角色

	// 检查是否按课题筛选
	projectIDStr := c.Query("project_id")
	if projectIDStr != "" {
		projectID, err := strconv.ParseUint(projectIDStr, 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "无效的课题ID",
			})
			return
		}

		assignments, err := h.assignmentService.GetAssignmentsByProject(uint(projectID))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "获取作业列表失败",
				"details": err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"data":  assignments,
			"total": len(assignments),
		})
		return
	}

	// 根据用户角色返回不同的作业列表
	if userRole == 1 { // 管理员
		page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
		pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

		assignments, total, err := h.assignmentService.GetAllAssignments(page, pageSize)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "获取作业列表失败",
				"details": err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"data":      assignments,
			"total":     total,
			"page":      page,
			"page_size": pageSize,
		})
	} else if userRole == 2 { // 老师
		assignments, err := h.assignmentService.GetAssignmentsByTeacher(userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "获取作业列表失败",
				"details": err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"data":  assignments,
			"total": len(assignments),
		})
	} else { // 学生
		assignments, err := h.assignmentService.GetAssignmentsByStudent(userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "获取作业列表失败",
				"details": err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"data":  assignments,
			"total": len(assignments),
		})
	}
}

// GetAssignment 获取作业详情
func (h *AssignmentHandler) GetAssignment(c *gin.Context) {
	assignmentID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "无效的作业ID",
		})
		return
	}

	assignment, err := h.assignmentService.GetAssignmentByID(uint(assignmentID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "作业不存在",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": assignment,
	})
}

// UpdateAssignment 更新作业信息
func (h *AssignmentHandler) UpdateAssignment(c *gin.Context) {
	assignmentID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "无效的作业ID",
		})
		return
	}

	var req services.UpdateAssignmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "无效的请求数据",
			"details": err.Error(),
		})
		return
	}

	assignment, err := h.assignmentService.UpdateAssignment(uint(assignmentID), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "更新作业失败",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "作业更新成功",
		"data":    assignment,
	})
}

// DeleteAssignment 删除作业
func (h *AssignmentHandler) DeleteAssignment(c *gin.Context) {
	assignmentID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "无效的作业ID",
		})
		return
	}

	if err := h.assignmentService.DeleteAssignment(uint(assignmentID)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "删除作业失败",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "作业删除成功",
	})
}

// SubmitAssignment 提交作业
func (h *AssignmentHandler) SubmitAssignment(c *gin.Context) {
	// TODO: 从JWT获取学生ID
	studentID := uint(2)

	assignmentID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "无效的作业ID",
		})
		return
	}

	var req services.SubmitAssignmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "无效的请求数据",
			"details": err.Error(),
		})
		return
	}

	submission, err := h.assignmentService.SubmitAssignment(studentID, uint(assignmentID), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "提交作业失败",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "作业提交成功",
		"data":    submission,
	})
}

// GetSubmissions 获取作业提交列表
func (h *AssignmentHandler) GetSubmissions(c *gin.Context) {
	assignmentID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "无效的作业ID",
		})
		return
	}

	submissions, err := h.assignmentService.GetSubmissionsByAssignment(uint(assignmentID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "获取提交列表失败",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  submissions,
		"total": len(submissions),
	})
}

// GetSubmissionDetail 获取提交详情
func (h *AssignmentHandler) GetSubmissionDetail(c *gin.Context) {
	submissionID, err := strconv.ParseUint(c.Param("submission_id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "无效的提交ID",
		})
		return
	}

	submission, err := h.assignmentService.GetSubmissionByID(uint(submissionID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "提交记录不存在",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": submission,
	})
}

// GetAssignmentStats 获取作业统计
func (h *AssignmentHandler) GetAssignmentStats(c *gin.Context) {
	assignmentID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "无效的作业ID",
		})
		return
	}

	stats, err := h.assignmentService.GetAssignmentStats(uint(assignmentID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "获取作业统计失败",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": stats,
	})
}

// GetMySubmissions 获取我的提交记录
func (h *AssignmentHandler) GetMySubmissions(c *gin.Context) {
	// TODO: 从JWT获取学生ID
	studentID := uint(2)

	submissions, err := h.assignmentService.GetSubmissionsByStudent(studentID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "获取提交记录失败",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  submissions,
		"total": len(submissions),
	})
}
