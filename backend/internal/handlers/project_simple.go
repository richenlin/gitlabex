package handlers

import (
	"net/http"
	"strconv"

	"gitlabex/internal/services"

	"github.com/gin-gonic/gin"
)

// ProjectSimpleHandler 简化的课题管理处理器
type ProjectSimpleHandler struct {
	projectService *services.ProjectService
	userService    *services.UserService
}

// NewProjectSimpleHandler 创建简化的课题管理处理器
func NewProjectSimpleHandler(projectService *services.ProjectService, userService *services.UserService) *ProjectSimpleHandler {
	return &ProjectSimpleHandler{
		projectService: projectService,
		userService:    userService,
	}
}

// RegisterRoutes 注册课题管理路由
func (h *ProjectSimpleHandler) RegisterRoutes(router *gin.RouterGroup) {
	projects := router.Group("/projects")
	{
		// 课题基本操作
		projects.POST("", h.CreateProject)       // 创建课题（老师）
		projects.GET("", h.ListProjects)         // 获取课题列表
		projects.GET("/:id", h.GetProject)       // 获取课题详情
		projects.PUT("/:id", h.UpdateProject)    // 更新课题信息（老师）
		projects.DELETE("/:id", h.DeleteProject) // 删除课题（老师）

		// 课题成员管理
		projects.POST("/join", h.JoinProject)                    // 学生加入课题
		projects.POST("/:id/members", h.AddMember)               // 添加成员
		projects.DELETE("/:id/members/:user_id", h.RemoveMember) // 移除成员
		projects.GET("/:id/members", h.GetMembers)               // 获取成员列表

		// 课题统计和GitLab集成
		projects.GET("/:id/stats", h.GetProjectStats) // 获取课题统计
		projects.GET("/:id/gitlab", h.GetGitLabInfo)  // 获取GitLab信息
	}
}

// CreateProject 创建课题
func (h *ProjectSimpleHandler) CreateProject(c *gin.Context) {
	// TODO: 从JWT获取用户ID，这里暂时使用测试用户
	teacherID := uint(1)

	var req services.CreateProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "无效的请求数据",
			"details": err.Error(),
		})
		return
	}

	project, err := h.projectService.CreateProject(teacherID, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "创建课题失败",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "课题创建成功",
		"data":    project,
	})
}

// ListProjects 获取课题列表
func (h *ProjectSimpleHandler) ListProjects(c *gin.Context) {
	// TODO: 从JWT获取用户信息
	userID := uint(1)
	userRole := 2 // 假设是老师角色

	// 检查是否按班级筛选
	classIDStr := c.Query("class_id")
	if classIDStr != "" {
		classID, err := strconv.ParseUint(classIDStr, 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "无效的班级ID",
			})
			return
		}

		projects, err := h.projectService.GetProjectsByClass(uint(classID))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "获取课题列表失败",
				"details": err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"data":  projects,
			"total": len(projects),
		})
		return
	}

	// 根据用户角色返回不同的课题列表
	if userRole == 1 { // 管理员
		page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
		pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

		projects, total, err := h.projectService.GetAllProjects(page, pageSize)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "获取课题列表失败",
				"details": err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"data":      projects,
			"total":     total,
			"page":      page,
			"page_size": pageSize,
		})
	} else if userRole == 2 { // 老师
		projects, err := h.projectService.GetProjectsByTeacher(userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "获取课题列表失败",
				"details": err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"data":  projects,
			"total": len(projects),
		})
	} else { // 学生
		projects, err := h.projectService.GetProjectsByStudent(userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "获取课题列表失败",
				"details": err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"data":  projects,
			"total": len(projects),
		})
	}
}

// GetProject 获取课题详情
func (h *ProjectSimpleHandler) GetProject(c *gin.Context) {
	projectID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "无效的课题ID",
		})
		return
	}

	project, err := h.projectService.GetProjectByID(uint(projectID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "课题不存在",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": project,
	})
}

// UpdateProject 更新课题信息
func (h *ProjectSimpleHandler) UpdateProject(c *gin.Context) {
	projectID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "无效的课题ID",
		})
		return
	}

	var req services.UpdateProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "无效的请求数据",
			"details": err.Error(),
		})
		return
	}

	project, err := h.projectService.UpdateProject(uint(projectID), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "更新课题失败",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "课题更新成功",
		"data":    project,
	})
}

// DeleteProject 删除课题
func (h *ProjectSimpleHandler) DeleteProject(c *gin.Context) {
	projectID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "无效的课题ID",
		})
		return
	}

	if err := h.projectService.DeleteProject(uint(projectID)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "删除课题失败",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "课题删除成功",
	})
}

// JoinProject 学生加入课题
func (h *ProjectSimpleHandler) JoinProject(c *gin.Context) {
	// TODO: 从JWT获取学生ID
	studentID := uint(2)

	var req services.JoinProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "无效的请求数据",
			"details": err.Error(),
		})
		return
	}

	project, err := h.projectService.JoinProject(studentID, req.Code)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "加入课题失败",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "成功加入课题",
		"data":    project,
	})
}

// AddMember 添加课题成员
func (h *ProjectSimpleHandler) AddMember(c *gin.Context) {
	projectID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "无效的课题ID",
		})
		return
	}

	var req struct {
		StudentID uint   `json:"student_id" binding:"required"`
		Role      string `json:"role"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "无效的请求数据",
			"details": err.Error(),
		})
		return
	}

	if req.Role == "" {
		req.Role = "member"
	}

	if err := h.projectService.AddStudentToProject(uint(projectID), req.StudentID, req.Role); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "添加成员失败",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "成员添加成功",
	})
}

// RemoveMember 移除课题成员
func (h *ProjectSimpleHandler) RemoveMember(c *gin.Context) {
	projectID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "无效的课题ID",
		})
		return
	}

	userID, err := strconv.ParseUint(c.Param("user_id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "无效的用户ID",
		})
		return
	}

	if err := h.projectService.RemoveStudentFromProject(uint(projectID), uint(userID)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "移除成员失败",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "成员移除成功",
	})
}

// GetMembers 获取课题成员列表
func (h *ProjectSimpleHandler) GetMembers(c *gin.Context) {
	projectID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "无效的课题ID",
		})
		return
	}

	members, err := h.projectService.GetProjectMembers(uint(projectID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "获取成员列表失败",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  members,
		"total": len(members),
	})
}

// GetProjectStats 获取课题统计信息
func (h *ProjectSimpleHandler) GetProjectStats(c *gin.Context) {
	projectID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "无效的课题ID",
		})
		return
	}

	stats, err := h.projectService.GetProjectStats(uint(projectID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "获取课题统计失败",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": stats,
	})
}

// GetGitLabInfo 获取GitLab信息
func (h *ProjectSimpleHandler) GetGitLabInfo(c *gin.Context) {
	projectID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "无效的课题ID",
		})
		return
	}

	gitlabInfo, err := h.projectService.GetProjectGitLabInfo(uint(projectID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "获取GitLab信息失败",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": gitlabInfo,
	})
}
