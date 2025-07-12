package handlers

import (
	"net/http"
	"strconv"

	"gitlabex/internal/services"

	"github.com/gin-gonic/gin"
)

// ClassHandler 班级管理处理器
type ClassHandler struct {
	classService *services.ClassService
	userService  *services.UserService
}

// NewClassHandler 创建班级管理处理器
func NewClassHandler(classService *services.ClassService, userService *services.UserService) *ClassHandler {
	return &ClassHandler{
		classService: classService,
		userService:  userService,
	}
}

// RegisterRoutes 注册班级管理路由
func (h *ClassHandler) RegisterRoutes(router *gin.RouterGroup) {
	classes := router.Group("/classes")
	{
		// 班级基本操作
		classes.POST("", h.CreateClass)       // 创建班级
		classes.GET("", h.ListClasses)        // 获取班级列表
		classes.GET("/:id", h.GetClass)       // 获取班级详情
		classes.PUT("/:id", h.UpdateClass)    // 更新班级信息
		classes.DELETE("/:id", h.DeleteClass) // 删除班级

		// 班级成员管理
		classes.POST("/join", h.JoinClass)                      // 学生加入班级
		classes.POST("/:id/members", h.AddMember)               // 添加成员
		classes.DELETE("/:id/members/:user_id", h.RemoveMember) // 移除成员
		classes.GET("/:id/members", h.GetMembers)               // 获取成员列表

		// 班级统计和同步
		classes.GET("/:id/stats", h.GetClassStats) // 获取班级统计
		classes.POST("/:id/sync", h.SyncToGitLab)  // 同步到GitLab
	}
}

// CreateClass 创建班级
func (h *ClassHandler) CreateClass(c *gin.Context) {
	// TODO: 从JWT获取用户ID，这里暂时使用测试用户
	teacherID := uint(1)

	var req services.CreateClassRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "无效的请求数据",
			"details": err.Error(),
		})
		return
	}

	class, err := h.classService.CreateClass(teacherID, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "创建班级失败",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "班级创建成功",
		"data":    class,
	})
}

// ListClasses 获取班级列表
func (h *ClassHandler) ListClasses(c *gin.Context) {
	// TODO: 根据用户角色返回不同的班级列表
	// 这里简单返回所有班级
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	classes, total, err := h.classService.GetAllClasses(page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "获取班级列表失败",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":      classes,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// GetClass 获取班级详情
func (h *ClassHandler) GetClass(c *gin.Context) {
	classID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "无效的班级ID",
		})
		return
	}

	class, err := h.classService.GetClassByID(uint(classID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "班级不存在",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": class,
	})
}

// UpdateClass 更新班级信息
func (h *ClassHandler) UpdateClass(c *gin.Context) {
	classID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "无效的班级ID",
		})
		return
	}

	var req services.UpdateClassRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "无效的请求数据",
			"details": err.Error(),
		})
		return
	}

	class, err := h.classService.UpdateClass(uint(classID), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "更新班级失败",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "班级更新成功",
		"data":    class,
	})
}

// DeleteClass 删除班级
func (h *ClassHandler) DeleteClass(c *gin.Context) {
	classID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "无效的班级ID",
		})
		return
	}

	if err := h.classService.DeleteClass(uint(classID)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "删除班级失败",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "班级删除成功",
	})
}

// JoinClass 学生加入班级
func (h *ClassHandler) JoinClass(c *gin.Context) {
	// TODO: 从JWT获取学生ID
	studentID := uint(2)

	var req services.JoinClassRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "无效的请求数据",
			"details": err.Error(),
		})
		return
	}

	class, err := h.classService.JoinClass(studentID, req.Code)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "加入班级失败",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "成功加入班级",
		"data":    class,
	})
}

// AddMember 添加班级成员
func (h *ClassHandler) AddMember(c *gin.Context) {
	classID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "无效的班级ID",
		})
		return
	}

	var req struct {
		StudentID uint `json:"student_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "无效的请求数据",
			"details": err.Error(),
		})
		return
	}

	if err := h.classService.AddStudentToClass(uint(classID), req.StudentID); err != nil {
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

// RemoveMember 移除班级成员
func (h *ClassHandler) RemoveMember(c *gin.Context) {
	classID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "无效的班级ID",
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

	if err := h.classService.RemoveStudentFromClass(uint(classID), uint(userID)); err != nil {
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

// GetMembers 获取班级成员列表
func (h *ClassHandler) GetMembers(c *gin.Context) {
	classID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "无效的班级ID",
		})
		return
	}

	members, err := h.classService.GetClassMembers(uint(classID))
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

// GetClassStats 获取班级统计信息
func (h *ClassHandler) GetClassStats(c *gin.Context) {
	classID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "无效的班级ID",
		})
		return
	}

	stats, err := h.classService.GetClassStats(uint(classID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "获取班级统计失败",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": stats,
	})
}

// SyncToGitLab 同步班级到GitLab
func (h *ClassHandler) SyncToGitLab(c *gin.Context) {
	classID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "无效的班级ID",
		})
		return
	}

	if err := h.classService.SyncClassToGitLab(uint(classID)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "同步到GitLab失败",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "同步到GitLab成功",
	})
}
