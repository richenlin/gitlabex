package handlers

import (
	"net/http"
	"strconv"

	"gitlabex/internal/models"
	"gitlabex/internal/services"

	"github.com/gin-gonic/gin"
)

// ProjectHandler 课题处理器
type ProjectHandler struct {
	projectService    *services.ProjectService
	permissionService *services.PermissionService
}

// NewProjectHandler 创建课题处理器
func NewProjectHandler(projectService *services.ProjectService, permissionService *services.PermissionService) *ProjectHandler {
	return &ProjectHandler{
		projectService:    projectService,
		permissionService: permissionService,
	}
}

// CreateProject 创建课题
func (h *ProjectHandler) CreateProject(c *gin.Context) {
	user, exists := c.Get("current_user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Authentication required",
		})
		return
	}

	currentUser := user.(*models.User)

	var req services.CreateProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	project, err := h.projectService.CreateProject(currentUser.ID, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to create project",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Project created successfully",
		"data":    project,
	})
}

// GetProject 获取课题详情
func (h *ProjectHandler) GetProject(c *gin.Context) {
	projectID := c.GetUint("project_id")

	project, err := h.projectService.GetProjectByID(projectID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "Project not found",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Success",
		"data":    project,
	})
}

// GetProjects 获取课题列表
func (h *ProjectHandler) GetProjects(c *gin.Context) {
	user, exists := c.Get("current_user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Authentication required",
		})
		return
	}

	currentUser := user.(*models.User)

	// 检查是否按班级筛选
	classIDStr := c.Query("class_id")
	if classIDStr != "" {
		classID, err := strconv.ParseUint(classIDStr, 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid class ID",
			})
			return
		}

		// 获取班级的课题列表
		projects, err := h.projectService.GetProjectsByClass(uint(classID))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to get projects",
				"details": err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "Success",
			"data":    projects,
		})
		return
	}

	switch currentUser.Role {
	case services.RoleAdmin:
		// 管理员获取所有课题
		page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
		pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

		if page < 1 {
			page = 1
		}
		if pageSize < 1 || pageSize > 100 {
			pageSize = 20
		}

		projects, total, err := h.projectService.GetAllProjects(page, pageSize)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to get projects",
				"details": err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "Success",
			"data": gin.H{
				"projects":  projects,
				"total":     total,
				"page":      page,
				"page_size": pageSize,
			},
		})

	case services.RoleTeacher:
		// 老师获取自己创建的课题
		projects, err := h.projectService.GetProjectsByTeacher(currentUser.ID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to get projects",
				"details": err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "Success",
			"data":    projects,
		})

	case services.RoleStudent:
		// 学生获取自己参加的课题
		projects, err := h.projectService.GetProjectsByStudent(currentUser.ID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to get projects",
				"details": err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "Success",
			"data":    projects,
		})

	default:
		c.JSON(http.StatusForbidden, gin.H{
			"error": "Insufficient permissions",
		})
	}
}

// UpdateProject 更新课题
func (h *ProjectHandler) UpdateProject(c *gin.Context) {
	projectID := c.GetUint("project_id")

	var req services.UpdateProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	project, err := h.projectService.UpdateProject(projectID, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to update project",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Project updated successfully",
		"data":    project,
	})
}

// DeleteProject 删除课题
func (h *ProjectHandler) DeleteProject(c *gin.Context) {
	projectID := c.GetUint("project_id")

	if err := h.projectService.DeleteProject(projectID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to delete project",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Project deleted successfully",
	})
}

// JoinProject 学生加入课题
func (h *ProjectHandler) JoinProject(c *gin.Context) {
	user, exists := c.Get("current_user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Authentication required",
		})
		return
	}

	currentUser := user.(*models.User)

	var req services.JoinProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	project, err := h.projectService.JoinProject(currentUser.ID, req.Code)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Failed to join project",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Successfully joined project",
		"data":    project,
	})
}

// AddStudentToProject 添加学生到课题
func (h *ProjectHandler) AddStudentToProject(c *gin.Context) {
	projectID := c.GetUint("project_id")

	studentIDStr := c.Param("student_id")
	studentID, err := strconv.ParseUint(studentIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid student ID",
		})
		return
	}

	// 获取角色参数（可选）
	role := c.DefaultQuery("role", "member")

	if err := h.projectService.AddStudentToProject(projectID, uint(studentID), role); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to add student to project",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Student added to project successfully",
	})
}

// RemoveStudentFromProject 从课题移除学生
func (h *ProjectHandler) RemoveStudentFromProject(c *gin.Context) {
	projectID := c.GetUint("project_id")

	studentIDStr := c.Param("student_id")
	studentID, err := strconv.ParseUint(studentIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid student ID",
		})
		return
	}

	if err := h.projectService.RemoveStudentFromProject(projectID, uint(studentID)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to remove student from project",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Student removed from project successfully",
	})
}

// UpdateStudentRole 更新学生角色
func (h *ProjectHandler) UpdateStudentRole(c *gin.Context) {
	projectID := c.GetUint("project_id")

	studentIDStr := c.Param("student_id")
	studentID, err := strconv.ParseUint(studentIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid student ID",
		})
		return
	}

	var req struct {
		Role string `json:"role" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	if err := h.projectService.UpdateStudentRole(projectID, uint(studentID), req.Role); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to update student role",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Student role updated successfully",
	})
}

// GetProjectMembers 获取课题成员
func (h *ProjectHandler) GetProjectMembers(c *gin.Context) {
	projectID := c.GetUint("project_id")

	members, err := h.projectService.GetProjectMembers(projectID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get project members",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Success",
		"data":    members,
	})
}

// GetProjectStats 获取课题统计
func (h *ProjectHandler) GetProjectStats(c *gin.Context) {
	projectID := c.GetUint("project_id")

	stats, err := h.projectService.GetProjectStats(projectID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get project stats",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Success",
		"data":    stats,
	})
}

// RegisterRoutes 注册课题相关路由
func (h *ProjectHandler) RegisterRoutes(router *gin.RouterGroup, permissionService *services.PermissionService) {
	projects := router.Group("/projects")
	{
		// 基础认证
		projects.Use(permissionService.RequireAuth())

		// 课题列表和创建
		projects.GET("", h.GetProjects)
		projects.POST("", permissionService.RequireTeacher(), h.CreateProject)

		// 学生加入课题
		projects.POST("/join", permissionService.RequireRole(services.RoleStudent), h.JoinProject)

		// 特定课题操作
		projectGroup := projects.Group("/:id")
		{
			// 获取课题详情（需要访问权限）
			projectGroup.GET("", permissionService.RequireProjectAccess(services.PermissionRead), h.GetProject)

			// 更新和删除课题（需要管理权限）
			projectGroup.PUT("", permissionService.RequireProjectAccess(services.PermissionManage), h.UpdateProject)
			projectGroup.DELETE("", permissionService.RequireProjectAccess(services.PermissionManage), h.DeleteProject)

			// 成员管理
			members := projectGroup.Group("/members")
			{
				members.GET("", permissionService.RequireProjectAccess(services.PermissionRead), h.GetProjectMembers)
				members.POST("/:student_id", permissionService.RequireProjectAccess(services.PermissionManage), h.AddStudentToProject)
				members.PUT("/:student_id/role", permissionService.RequireProjectAccess(services.PermissionManage), h.UpdateStudentRole)
				members.DELETE("/:student_id", permissionService.RequireProjectAccess(services.PermissionManage), h.RemoveStudentFromProject)
			}

			// 统计信息
			projectGroup.GET("/stats", permissionService.RequireProjectAccess(services.PermissionRead), h.GetProjectStats)
		}
	}
}
