package handlers

import (
	"net/http"
	"strconv"

	"gitlabex/internal/models"
	"gitlabex/internal/services"

	"github.com/gin-gonic/gin"
)

// ClassHandler 班级处理器
type ClassHandler struct {
	classService      *services.ClassService
	permissionService *services.PermissionService
}

// NewClassHandler 创建班级处理器
func NewClassHandler(classService *services.ClassService, permissionService *services.PermissionService) *ClassHandler {
	return &ClassHandler{
		classService:      classService,
		permissionService: permissionService,
	}
}

// CreateClass 创建班级
func (h *ClassHandler) CreateClass(c *gin.Context) {
	user, exists := c.Get("current_user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Authentication required",
		})
		return
	}

	currentUser := user.(*models.User)

	var req services.CreateClassRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	class, err := h.classService.CreateClass(currentUser.ID, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to create class",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Class created successfully",
		"data":    class,
	})
}

// GetClass 获取班级详情
func (h *ClassHandler) GetClass(c *gin.Context) {
	classID := c.GetUint("class_id")

	class, err := h.classService.GetClassByID(classID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "Class not found",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Success",
		"data":    class,
	})
}

// GetClasses 获取班级列表
func (h *ClassHandler) GetClasses(c *gin.Context) {
	user, exists := c.Get("current_user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Authentication required",
		})
		return
	}

	currentUser := user.(*models.User)

	switch currentUser.Role {
	case services.RoleAdmin:
		// 管理员获取所有班级
		page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
		pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

		if page < 1 {
			page = 1
		}
		if pageSize < 1 || pageSize > 100 {
			pageSize = 20
		}

		classes, total, err := h.classService.GetAllClasses(page, pageSize)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to get classes",
				"details": err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "Success",
			"data": gin.H{
				"classes":   classes,
				"total":     total,
				"page":      page,
				"page_size": pageSize,
			},
		})

	case services.RoleTeacher:
		// 老师获取自己创建的班级
		classes, err := h.classService.GetClassesByTeacher(currentUser.ID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to get classes",
				"details": err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "Success",
			"data":    classes,
		})

	case services.RoleStudent:
		// 学生获取自己加入的班级
		classes, err := h.classService.GetClassesByStudent(currentUser.ID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to get classes",
				"details": err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "Success",
			"data":    classes,
		})

	default:
		c.JSON(http.StatusForbidden, gin.H{
			"error": "Insufficient permissions",
		})
	}
}

// UpdateClass 更新班级
func (h *ClassHandler) UpdateClass(c *gin.Context) {
	classID := c.GetUint("class_id")

	var req services.UpdateClassRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	class, err := h.classService.UpdateClass(classID, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to update class",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Class updated successfully",
		"data":    class,
	})
}

// DeleteClass 删除班级
func (h *ClassHandler) DeleteClass(c *gin.Context) {
	classID := c.GetUint("class_id")

	if err := h.classService.DeleteClass(classID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to delete class",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Class deleted successfully",
	})
}

// JoinClass 学生加入班级
func (h *ClassHandler) JoinClass(c *gin.Context) {
	user, exists := c.Get("current_user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Authentication required",
		})
		return
	}

	currentUser := user.(*models.User)

	var req services.JoinClassRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	class, err := h.classService.JoinClass(currentUser.ID, req.Code)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Failed to join class",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Successfully joined class",
		"data":    class,
	})
}

// AddStudentToClass 添加学生到班级
func (h *ClassHandler) AddStudentToClass(c *gin.Context) {
	classID := c.GetUint("class_id")

	studentIDStr := c.Param("student_id")
	studentID, err := strconv.ParseUint(studentIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid student ID",
		})
		return
	}

	if err := h.classService.AddStudentToClass(classID, uint(studentID)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to add student to class",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Student added to class successfully",
	})
}

// RemoveStudentFromClass 从班级移除学生
func (h *ClassHandler) RemoveStudentFromClass(c *gin.Context) {
	classID := c.GetUint("class_id")

	studentIDStr := c.Param("student_id")
	studentID, err := strconv.ParseUint(studentIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid student ID",
		})
		return
	}

	if err := h.classService.RemoveStudentFromClass(classID, uint(studentID)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to remove student from class",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Student removed from class successfully",
	})
}

// GetClassMembers 获取班级成员
func (h *ClassHandler) GetClassMembers(c *gin.Context) {
	classID := c.GetUint("class_id")

	members, err := h.classService.GetClassMembers(classID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get class members",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Success",
		"data":    members,
	})
}

// GetClassStats 获取班级统计
func (h *ClassHandler) GetClassStats(c *gin.Context) {
	classID := c.GetUint("class_id")

	stats, err := h.classService.GetClassStats(classID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get class stats",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Success",
		"data":    stats,
	})
}

// RegisterRoutes 注册班级相关路由
func (h *ClassHandler) RegisterRoutes(router *gin.RouterGroup, permissionService *services.PermissionService) {
	classes := router.Group("/classes")
	{
		// 基础认证
		classes.Use(permissionService.RequireAuth())

		// 班级列表和创建
		classes.GET("", h.GetClasses)
		classes.POST("", permissionService.RequireTeacher(), h.CreateClass)

		// 学生加入班级
		classes.POST("/join", permissionService.RequireRole(services.RoleStudent), h.JoinClass)

		// 特定班级操作
		classGroup := classes.Group("/:id")
		{
			// 获取班级详情（需要访问权限）
			classGroup.GET("", permissionService.RequireClassAccess(services.PermissionRead), h.GetClass)

			// 更新和删除班级（需要管理权限）
			classGroup.PUT("", permissionService.RequireClassAccess(services.PermissionManage), h.UpdateClass)
			classGroup.DELETE("", permissionService.RequireClassAccess(services.PermissionManage), h.DeleteClass)

			// 成员管理
			members := classGroup.Group("/members")
			{
				members.GET("", permissionService.RequireClassAccess(services.PermissionRead), h.GetClassMembers)
				members.POST("/:student_id", permissionService.RequireClassAccess(services.PermissionManage), h.AddStudentToClass)
				members.DELETE("/:student_id", permissionService.RequireClassAccess(services.PermissionManage), h.RemoveStudentFromClass)
			}

			// 统计信息
			classGroup.GET("/stats", permissionService.RequireClassAccess(services.PermissionRead), h.GetClassStats)
		}
	}
}
