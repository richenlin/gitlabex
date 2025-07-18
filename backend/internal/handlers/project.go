package handlers

import (
	"fmt"
	"net/http"
	"strconv"

	"gitlabex/internal/config"
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

	fmt.Printf("DEBUG: User role: %d, RoleAdmin: %d\n", currentUser.Role, config.RoleAdmin)

	switch currentUser.Role {
	case config.RoleAdmin:
		fmt.Printf("DEBUG: Admin case - getting all projects\n")
		// 管理员获取所有课题
		page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
		pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

		projects, total, err := h.projectService.GetAllProjects(page, pageSize)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to get projects",
				"details": err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"projects": projects,
			"total":    total,
			"page":     page,
			"pageSize": pageSize,
		})

	case config.RoleTeacher:
		fmt.Printf("DEBUG: Teacher case - getting teacher's projects\n")
		// 教师获取自己的课题
		projects, err := h.projectService.GetProjectsByTeacher(currentUser.ID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to get teacher projects",
				"details": err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"projects": projects,
		})

	case config.RoleStudent:
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
			"data":  projects,
			"total": len(projects),
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

// GetProjectWithGitLabStats 获取包含GitLab统计信息的课题详情
func (h *ProjectHandler) GetProjectWithGitLabStats(c *gin.Context) {
	projectID := c.GetUint("project_id")

	project, gitlabStats, err := h.projectService.GetProjectWithGitLabStats(projectID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get project with GitLab stats",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Success",
		"data": gin.H{
			"project":      project,
			"gitlab_stats": gitlabStats,
		},
	})
}

// GetProjectGitLabInfo 获取课题的GitLab信息
func (h *ProjectHandler) GetProjectGitLabInfo(c *gin.Context) {
	projectID := c.GetUint("project_id")

	gitlabInfo, err := h.projectService.GetProjectGitLabInfo(projectID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get GitLab info",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Success",
		"data":    gitlabInfo,
	})
}

// SubmitAssignmentToGitLab 提交作业到GitLab
func (h *ProjectHandler) SubmitAssignmentToGitLab(c *gin.Context) {
	user, exists := c.Get("current_user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Authentication required",
		})
		return
	}

	currentUser := user.(*models.User)
	projectID := c.GetUint("project_id")

	var req struct {
		AssignmentID uint              `json:"assignment_id" binding:"required"`
		Files        map[string]string `json:"files" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	submission, err := h.projectService.SubmitAssignmentToGitLab(projectID, currentUser.ID, req.AssignmentID, req.Files)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to submit assignment",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Assignment submitted successfully",
		"data":    submission,
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
		projects.POST("/join", permissionService.RequireRole(models.EducationRole(config.RoleStudent)), h.JoinProject)

		// 特定课题操作
		projectGroup := projects.Group("/:project_id")
		{
			// 获取课题详情（需要访问权限）
			projectGroup.GET("", permissionService.RequireProjectAccess(config.PermissionRead), h.GetProject)

			// 更新课题信息（需要管理权限）
			projectGroup.PUT("", permissionService.RequireProjectAccess(config.PermissionManage), h.UpdateProject)

			// 删除课题（需要管理权限）
			projectGroup.DELETE("", permissionService.RequireProjectAccess(config.PermissionManage), h.DeleteProject)

			// 移除课题成员（需要管理权限）
			projectGroup.DELETE("/members/:student_id", permissionService.RequireProjectAccess(config.PermissionManage), h.RemoveStudentFromProject)

			// 课题统计信息
			projectGroup.GET("/stats", permissionService.RequireProjectAccess(config.PermissionRead), h.GetProjectStats)

			// GitLab集成相关路由
			gitlab := projectGroup.Group("/gitlab")
			{
				gitlab.GET("/stats",
					permissionService.RequireRole(models.EducationRole(config.RoleStudent)), // 学生及以上权限
					h.GetProjectWithGitLabStats)

				gitlab.GET("/info",
					permissionService.RequireRole(models.EducationRole(config.RoleStudent)), // 学生及以上权限
					h.GetProjectGitLabInfo)

				gitlab.POST("/submit",
					permissionService.RequireRole(models.EducationRole(config.RoleStudent)),
					h.SubmitAssignmentToGitLab)
			}
		}
	}
}
